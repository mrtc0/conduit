package strategy

import (
	"context"
	"fmt"

	"github.com/mrtc0/conduit/event"
	"github.com/mrtc0/conduit/log"
)

type StrategyType string

const (
	Stream StrategyType = "stream"
	Batch  StrategyType = "batch"
)

type InputChannel[T any] chan *event.Event[T]

// SendingStrategy is responsible for sending messages from input channels to output channels.
// Its responsibility is to convert messages into payloads and send them to the Sender's channel.
type SendingStrategy[T any] interface {
	// Start starts the sending strategy.
	Start()
	// WaitStop waits for the strategy to stop processing messages.
	WaitStop()
	// Flush flushes any remaining messages in the strategy.
	Flush(ctx context.Context)
}

// StreamStrategy is a concrete implementation of SendingStrategy that sends messages as they are received.
type StreamStrategy[T any] struct {
	inputChan  InputChannel[T]
	outputChan chan<- *event.Payload[T]
	done       chan struct{}
}

func NewStreamStrategy[T any](
	inputChan InputChannel[T],
	outputChan chan<- *event.Payload[T],
) SendingStrategy[T] {
	return &StreamStrategy[T]{
		inputChan:  inputChan,
		outputChan: outputChan,
		done:       make(chan struct{}),
	}
}

func (s *StreamStrategy[T]) Start() {
	go func() {
		for evt := range s.inputChan {
			s.processMessage(evt)
		}

		s.done <- struct{}{}
	}()
}

func (s *StreamStrategy[T]) WaitStop() {
	<-s.done
}

func (s *StreamStrategy[T]) Flush(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if len(s.inputChan) == 0 {
				return
			}
			evt, ok := <-s.inputChan
			if !ok {
				return
			}

			s.processMessage(evt)
			return
		}
	}
}

func (s *StreamStrategy[T]) processMessage(evt *event.Event[T]) {
	encodedContent, err := evt.MarshalJSON()
	if err != nil {
		log.Error(fmt.Sprintf("failed to marshal event content: %v", err))
		return
	}

	s.outputChan <- event.NewPayload[T](&evt.Metadata, encodedContent)
}
