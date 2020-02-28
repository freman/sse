package sse

// TODO: retry, and id?
type event struct {
	Name string
	Body interface{}
}
