package sink

import (
	"github.com/mrtc0/conduit/event"
)

type Sink[T any] interface {
	// Write sends the processed event to the sink.
	Write(payload *event.Payload[T]) error
	// Close closes the sink and cleans up resources.
	Close() error
}

type Result[T any] struct {
	Payload *event.Payload[T]
	Err     error
}
