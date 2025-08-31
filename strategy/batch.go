package strategy

import (
	"context"
	"fmt"
	"time"

	"github.com/mrtc0/conduit/event"
	"github.com/mrtc0/conduit/log"
)

type payloadBuffer[T any] struct {
	payloads []*event.Payload[T]
	// sizeLimit is the maximum byte size of the payload buffer.
	sizeLimit   int
	currentSize int
}

func newPayloadBuffer[T any](sizeLimit int) *payloadBuffer[T] {
	return &payloadBuffer[T]{
		payloads:    make([]*event.Payload[T], 0),
		sizeLimit:   sizeLimit,
		currentSize: 0,
	}
}

func (pb *payloadBuffer[T]) add(payload *event.Payload[T]) bool {
	contentSize := len(payload.JSONEncodedContent)

	if pb.reachLimit(contentSize) {
		return false
	}

	pb.payloads = append(pb.payloads, payload)
	pb.currentSize += len(payload.JSONEncodedContent)

	return true
}

func (pb *payloadBuffer[T]) payload() *event.Payload[T] {
	if len(pb.payloads) == 0 {
		return nil
	}

	payload := make([]byte, 0)

	for _, p := range pb.payloads {
		payload = append(payload, p.JSONEncodedContent...)
	}

	return event.NewPayload[T](&event.Metadata{}, payload)
}

func (pb payloadBuffer[T]) reachLimit(nextMessageContentSize int) bool {
	return pb.currentSize+nextMessageContentSize > pb.sizeLimit
}

func (pb *payloadBuffer[T]) clear() {
	pb.payloads = []*event.Payload[T]{}
	pb.currentSize = 0
}

type batchStrategy[T any] struct {
	inputChan  chan *event.Event[T]
	outputChan chan<- *event.Payload[T]

	forceFlush     chan struct{}
	forceFlushDone chan struct{}

	buffer       *payloadBuffer[T]
	waitDuration time.Duration

	clock Clock

	quit chan struct{}
}

type BatchStrategyOptionsFunc[T any] func(*batchStrategy[T])

func WithClock[T any](clock Clock) BatchStrategyOptionsFunc[T] {
	return func(b *batchStrategy[T]) {
		b.clock = clock
	}
}

func NewBatchStrategy[T any](
	inputChan chan *event.Event[T],
	outputChan chan<- *event.Payload[T],
	waitTimeDuration time.Duration,
	bufferLimitBytes int,
	opts ...BatchStrategyOptionsFunc[T],
) SendingStrategy[T] {
	s := &batchStrategy[T]{
		inputChan:      inputChan,
		outputChan:     outputChan,
		forceFlush:     make(chan struct{}, 1),
		forceFlushDone: make(chan struct{}, 1),
		buffer:         newPayloadBuffer[T](bufferLimitBytes),
		waitDuration:   waitTimeDuration,
		clock:          DefaultClock,
		quit:           make(chan struct{}, 1),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (b *batchStrategy[T]) Start() {
	go func() {
		flushTicker := b.clock.NewTicker(b.waitDuration)
		defer func() {
			flushTicker.Stop()
			close(b.forceFlushDone)
			close(b.forceFlush)
			close(b.quit)
		}()

		for {
			select {
			case evt, ok := <-b.inputChan:
				if !ok {
					return
				}
				b.processMessage(evt)
			case <-flushTicker.C:
				b.flush()
			case <-b.forceFlush:
				b.flush()
				// non-blocking send to avoid blocking if no one is waiting
				select {
				case b.forceFlushDone <- struct{}{}:
				default:
				}
			}
		}
	}()
}

func (b *batchStrategy[T]) WaitStop() {
	<-b.quit
}

func (b *batchStrategy[T]) Flush(ctx context.Context) {
	b.forceFlush <- struct{}{}

	for {
		select {
		case <-ctx.Done():
			return
		case <-b.forceFlushDone:
			return
		}
	}
}

func (b *batchStrategy[T]) processMessage(evt *event.Event[T]) {
	encodedContent, err := evt.MarshalJSON()
	if err != nil {
		log.Error(fmt.Sprintf("failed to marshal event content: %v", err))
		return
	}

	payload := event.NewPayload[T](&evt.Metadata, encodedContent)

	if added := b.buffer.add(payload); !added {
		b.flush()

		if added := b.buffer.add(payload); !added {
			log.Warn("Payload size exceeds buffer limit, dropping message")
		}
	}
}

func (b *batchStrategy[T]) flush() {
	payload := b.buffer.payload()
	if payload == nil {
		return
	}

	b.buffer.clear()
	b.outputChan <- payload
}
