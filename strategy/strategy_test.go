package strategy_test

import (
	"testing"

	"github.com/mrtc0/conduit/event"
	"github.com/mrtc0/conduit/strategy"
	"github.com/mrtc0/conduit/testutils"
	"github.com/stretchr/testify/assert"
)

func TestStreamStrategy(t *testing.T) {
	t.Parallel()

	inputChan := make(chan *event.Event[testutils.DummyEvent])
	outputChan := make(chan *event.Payload[testutils.DummyEvent])

	strategy := strategy.NewStreamStrategy(inputChan, outputChan)
	strategy.Start()
	defer func() {
		close(inputChan)
		strategy.WaitStop()
	}()

	done := make(chan struct{})

	go func() {
		for payload := range outputChan {
			assert.Equal(
				t,
				payload.JSONEncodedContent,
				[]byte("{\"id\":\"123\",\"name\":\"Test Event\"}\n"),
			)
			done <- struct{}{}
		}
	}()

	inputChan <- event.NewEvent(
		&event.RawEvent[testutils.DummyEvent]{
			Content: testutils.DummyEvent{ID: "123", Name: "Test Event"},
		},
	)

	<-done
}
