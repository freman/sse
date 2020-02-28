package sse

// Emitters Emit events with Optional arguments
type Emitter interface {
	Emit(body interface{}, opts ...EmitOption)
}

// Single is a single event emitter, not a broadcaster like EventStream
// used to handle on* callbacks
type Single chan<- event

// Emit an event
func (c Single) Emit(body interface{}, opts ...EmitOption) {
	c <- buildEvent(body, opts)
}

// type checks
var _ Emitter = (*Single)(nil)
var _ Emitter = (*EventStream)(nil)
