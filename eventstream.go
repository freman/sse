package sse

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// EventStream is your handy dandy all in one http.Handler and Emitter
type EventStream struct {
	bus chan event

	clients map[chan<- event]struct{}

	register   chan chan<- event
	unregister chan chan<- event

	onConnect func(Emitter)
}

// New EventStream will construct a stream and start a goroutine to forward
// braodcast events in the background
func New(opts ...NewOption) *EventStream {
	es := &EventStream{
		bus:        make(chan event),
		clients:    make(map[chan<- event]struct{}),
		register:   make(chan chan<- event, 5),
		unregister: make(chan chan<- event, 5),
	}

	for _, opt := range opts {
		opt(es)
	}

	go es.run()

	return es
}

// Emit an event
func (e *EventStream) Emit(body interface{}, opts ...EmitOption) {
	e.bus <- buildEvent(body, opts)
}

// ServeHTTP implements the http.Handler interface, requests will be handled if possible.
// It's this point where your onConnect callback will fire if you set it up.
func (e *EventStream) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)

	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	eventChan := make(chan event, 2)
	e.register <- eventChan
	defer func() {
		e.unregister <- eventChan
	}()

	if e.onConnect != nil {
		go func() {
			e.onConnect(Single(eventChan))
		}()
	}

	enc := json.NewEncoder(w)
	for {
		select {
		case <-r.Context().Done():
			return

		case event, ok := <-eventChan:
			if !ok {
				return
			}

			if event.Name != "" {
				if _, err := fmt.Fprintf(w, "event: %s\n", event.Name); err != nil {
					log.Println(err)
					return
				}
			}

			if _, err := fmt.Fprint(w, "data: "); err != nil {
				log.Println(err)
				return
			}

			if err := enc.Encode(event.Body); err != nil {
				log.Println(err)
				return
			}

			if _, err := fmt.Fprint(w, "\n\n"); err != nil {
				log.Println(err)
				return
			}

			flusher.Flush()
		}
	}
}

// Len returns the number of connected clients
func (e *EventStream) Len() int {
	return len(e.clients)
}

func (e *EventStream) run() {
	for {
		select {
		case client := <-e.register:
			e.clients[client] = struct{}{}
		case client := <-e.unregister:
			delete(e.clients, client)
			close(client)
		case event := <-e.bus:
			for client := range e.clients {
				select {
				case client <- event:
				default:
					e.unregister <- client
				}
			}
		}
	}
}
