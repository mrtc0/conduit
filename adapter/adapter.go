package adapter

import (
	"time"

	"github.com/mrtc0/conduit/event"
	"github.com/mrtc0/conduit/source"
)

type EventAdapter[T any] struct {
	source        *source.EventSource[T]
	pipelineInput chan *event.Event[T]

	quit        chan struct{}
	timeNowFunc func() time.Time
}

func NewEventAdapter[T any](
	source *source.EventSource[T],
	pipelineInput chan *event.Event[T],
) *EventAdapter[T] {
	return &EventAdapter[T]{
		source:        source,
		pipelineInput: pipelineInput,
		quit:          make(chan struct{}),
		timeNowFunc:   time.Now,
	}
}

func (a *EventAdapter[T]) Start() {
	go a.run()
}

func (a *EventAdapter[T]) WaitClose() {
	<-a.quit
}

func (a *EventAdapter[T]) run() {
	defer func() {
		close(a.quit)
	}()

	for rawEvt := range a.source.InputChannel {
		evt := event.NewEvent(rawEvt)

		if evt.IngestionTime.IsZero() {
			evt.IngestionTime = a.timeNowFunc()
		}

		a.pipelineInput <- evt
	}
}

func (a *EventAdapter[T]) SetTimeNowFunc(fn func() time.Time) {
	a.timeNowFunc = fn
}
