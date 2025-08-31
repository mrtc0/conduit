package sender_test

import (
	"context"
	"testing"

	"github.com/mrtc0/conduit/event"
	"github.com/mrtc0/conduit/sender"
	"github.com/mrtc0/conduit/sink"
	"github.com/stretchr/testify/assert"
)

func TestSender_Start(t *testing.T) {
	t.Parallel()

	t.Run("written to the sink when an event is received", func(t *testing.T) {
		t.Parallel()

		mockSink := &MockSink{
			writeFunc: func(payload *event.Payload[string]) error {
				return nil
			},
		}

		s := sender.NewSender(mockSink, nil)
		s.Start()
		s.In() <- &event.Payload[string]{
			JSONEncodedContent: []byte("test event"),
		}

		assert.NoError(t, s.Flush(context.Background()))

		assert.Equal(t, 1, mockSink.CallCount)

		assert.NoError(t, s.Stop())
	})

	t.Run("if returned error from Sink, written to error channel", func(t *testing.T) {
		t.Parallel()

		payload := &event.Payload[string]{
			JSONEncodedContent: []byte("test event"),
		}

		mockSink := &MockSink{
			writeFunc: func(payload *event.Payload[string]) error {
				return assert.AnError
			},
		}

		resultCh := make(chan *sink.Result[string], 1)
		s := sender.NewSender(mockSink, resultCh)
		s.Start()

		s.In() <- payload

		assert.NoError(t, s.Flush(context.Background()))

		assert.Equal(t, 1, mockSink.CallCount)

		assert.Len(t, resultCh, 1)
		result := <-resultCh
		assert.Equal(t, assert.AnError, result.Err)
		assert.Equal(t, payload, result.Payload)

		assert.NoError(t, s.Stop())
	})
}

type MockSink struct {
	CallCount int
	writeFunc func(payload *event.Payload[string]) error
}

func (m *MockSink) Start() {}
func (m *MockSink) Write(payload *event.Payload[string]) error {
	m.CallCount++
	return m.writeFunc(payload)
}
func (m *MockSink) Close() error {
	return nil
}
