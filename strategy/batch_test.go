package strategy_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/mrtc0/conduit/event"
	"github.com/mrtc0/conduit/strategy"
	"github.com/mrtc0/conduit/testutils"
	"github.com/stretchr/testify/assert"
)

var _ strategy.Clock = &testutils.MockClock{}

func TestBatchStrategy_Start(t *testing.T) {
	t.Parallel()

	t.Run("should send messages in batches", func(t *testing.T) {
		t.Parallel()

		clock := testutils.NewMockClock()

		inputChan := make(chan *event.Event[testutils.DummyEvent])
		outputChan := make(chan *event.Payload[testutils.DummyEvent])

		strategy := strategy.NewBatchStrategy(
			inputChan,
			outputChan,
			time.Duration(10*time.Second),
			100,
			strategy.WithClock[testutils.DummyEvent](clock),
		)

		strategy.Start()
		defer func() {
			close(inputChan)
			strategy.WaitStop()
		}()

		got := ""
		restartChan := make(chan struct{})

		go func() {
			for payload := range outputChan {
				got += string(payload.JSONEncodedContent)
				restartChan <- struct{}{}
			}
		}()

		for id := range 3 {
			inputChan <- event.NewEvent(
				&event.RawEvent[testutils.DummyEvent]{
					Content: testutils.DummyEvent{ID: fmt.Sprintf("%d", id), Name: "Test Event"},
				},
			)
		}

		assert.Empty(t, got)
		clock.Add(10 * time.Second)
		<-restartChan

		want := `{"id":"0","name":"Test Event"}
{"id":"1","name":"Test Event"}
{"id":"2","name":"Test Event"}
`
		assert.Equal(t, want, got)
	})

	t.Run("If the buffer becomes full, it will be flushed immediately", func(t *testing.T) {
		t.Parallel()

		inputChan := make(chan *event.Event[testutils.DummyEvent])
		outputChan := make(chan *event.Payload[testutils.DummyEvent])

		bufferLimitBytes := 120

		strategy := strategy.NewBatchStrategy(
			inputChan,
			outputChan,
			time.Duration(60*time.Second),
			bufferLimitBytes,
		)

		strategy.Start()
		defer func() {
			close(inputChan)
			strategy.WaitStop()
		}()

		restartChen := make(chan struct{})

		got := make([]string, 0)
		go func() {
			for paylaod := range outputChan {
				got = append(got, string(paylaod.JSONEncodedContent))
				restartChen <- struct{}{}
			}
		}()

		inputChan <- event.NewEvent(
			&event.RawEvent[testutils.DummyEvent]{
				Content: testutils.DummyEvent{ID: "1", Name: testutils.RandomString(5)},
			},
		)
		assert.Equal(t, 0, len(got)) // buffer is not full yet

		inputChan <- event.NewEvent(
			&event.RawEvent[testutils.DummyEvent]{
				Content: testutils.DummyEvent{ID: "2", Name: testutils.RandomString(100)},
			},
		)
		<-restartChen
		assert.Equal(t, 1, len(got))
	})
}

func TestBatchStrategy_Flush(t *testing.T) {
	t.Parallel()

	inputChan := make(chan *event.Event[testutils.DummyEvent])
	outputChan := make(chan *event.Payload[testutils.DummyEvent])

	strategy := strategy.NewBatchStrategy(
		inputChan,
		outputChan,
		time.Duration(60*time.Second),
		1000,
	)

	strategy.Start()
	defer func() {
		close(inputChan)
		strategy.WaitStop()
	}()

	got := make([]string, 0)
	restartChan := make(chan struct{})
	go func() {
		for payload := range outputChan {
			got = append(got, string(payload.JSONEncodedContent))
			restartChan <- struct{}{}
		}
	}()

	names := []string{}
	for i := range 5 {
		names = append(names, testutils.RandomString(5))
		inputChan <- event.NewEvent(
			&event.RawEvent[testutils.DummyEvent]{
				Content: testutils.DummyEvent{ID: fmt.Sprintf("%d", i), Name: names[i]},
			},
		)
	}

	assert.Equal(t, 0, len(got)) // buffer is not full yet

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	strategy.Flush(ctx)
	<-restartChan

	assert.Equal(t, 1, len(got))

	for i, name := range names {
		want := fmt.Sprintf(`{"id":"%d","name":"%s"}`, i, name)
		assert.True(t, strings.Contains(got[0], want))
	}
}
