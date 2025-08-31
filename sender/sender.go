package sender

import (
	"context"

	"github.com/mrtc0/conduit/event"
	"github.com/mrtc0/conduit/sink"
)

var (
	// defaultQueueSize is the default size of the queue for the sender.
	defaultQueueSize = 100
)

type Sender[T any] struct {
	sink     sink.Sink[T]
	resultCh chan *sink.Result[T]
	queue    chan *event.Payload[T]

	done chan struct{}
}

func NewSender[T any](sink sink.Sink[T], resultCh chan *sink.Result[T]) *Sender[T] {
	queue := make(chan *event.Payload[T], defaultQueueSize)

	return &Sender[T]{
		sink:     sink,
		resultCh: resultCh,
		queue:    queue,

		done: make(chan struct{}),
	}
}

func (s *Sender[T]) Start() {
	go s.run()
}

func (s *Sender[T]) Stop() error {
	close(s.queue)
	<-s.done

	if err := s.sink.Close(); err != nil {
		return err
	}

	return nil
}

func (s *Sender[T]) Flush(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if len(s.queue) == 0 {
				return nil
			}

			payload := <-s.queue
			s.process(payload)
		}
	}
}

func (s *Sender[T]) In() chan<- *event.Payload[T] {
	return s.queue
}

func (s *Sender[T]) run() {
	defer func() {
		close(s.done)
	}()

	for payload := range s.queue {
		s.process(payload)
	}
}

func (s *Sender[T]) process(payload *event.Payload[T]) {
	err := s.sink.Write(payload)

	if s.resultCh != nil {
		s.resultCh <- &sink.Result[T]{
			Payload: payload,
			Err:     err,
		}
	}
}
