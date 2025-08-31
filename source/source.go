package source

import "github.com/mrtc0/conduit/event"

type EventSource[T any] struct {
	InputChannel chan *event.RawEvent[T]
}
