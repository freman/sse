package sse

// EmitOption for being passed to an Emitter's Emit function mostly for altering
// the behavior or structure of the event
type EmitOption func(*event)

// WithName is an EmitOption that sets the name for a given event to make it possible
// to make use of the addEventHandler on the EventSource in javascript
//
// Emit("data", WithName("borg"))
//
// source.addEventListener("borg", function(ev) {
//     console.log("Oh no, the borg have ", ev.data)
// })
func WithName(name string) EmitOption {
	return func(e *event) {
		e.Name = name
	}
}

// NewOption for being passed to the New() func
type NewOption func(*EventStream)

// WithOnConnect allows you to pass an OnConnect function that will be passed
// an Emitter when a new client connects.
func WithOnConnect(fn func(Emitter)) NewOption {
	return func(e *EventStream) {
		e.onConnect = fn
	}
}

func buildEvent(body interface{}, opts []EmitOption) event {
	ev := event{
		Body: body,
	}

	for _, opt := range opts {
		opt(&ev)
	}

	return ev
}
