package adapter_test

import (
	"testing"
	"time"

	"github.com/mrtc0/conduit/adapter"
	"github.com/mrtc0/conduit/event"
	"github.com/mrtc0/conduit/source"
	"github.com/mrtc0/conduit/testutils"
	"github.com/stretchr/testify/assert"
)

func TestEventAdapter_run(t *testing.T) {
	t.Parallel()

	source := &source.EventSource[testutils.DummyEvent]{
		InputChannel: make(chan *event.RawEvent[testutils.DummyEvent]),
	}
	pipelineInput := make(chan *event.Event[testutils.DummyEvent])

	eventAdapter := adapter.NewEventAdapter(source, pipelineInput)
	eventAdapter.SetTimeNowFunc(func() time.Time {
		return time.Date(2025, 1, 2, 3, 0, 0, 0, time.UTC)
	})

	eventAdapter.Start()

	defer func() {
		close(source.InputChannel)
		eventAdapter.WaitClose()
	}()

	testCases := map[string]struct {
		rawEvent *event.RawEvent[testutils.DummyEvent]
		want     *event.Metadata
	}{
		"when IngestionTime is not set in Metadata": {
			rawEvent: event.NewRawEvent(
				testutils.DummyEvent{ID: "123", Name: "Test Event"},
				nil,
			),
			want: &event.Metadata{
				Tags:          event.Tags{},
				IngestionTime: time.Date(2025, 1, 2, 3, 0, 0, 0, time.UTC),
			},
		},
		"when IngestionTime is set in Metadata": {
			rawEvent: event.NewRawEvent(
				testutils.DummyEvent{ID: "456", Name: "Another Event"},
				&event.Metadata{
					IngestionTime: time.Date(2025, 7, 18, 13, 0, 0, 0, time.UTC),
				},
			),
			want: &event.Metadata{
				Tags:          event.Tags{},
				IngestionTime: time.Date(2025, 7, 18, 13, 0, 0, 0, time.UTC),
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			source.InputChannel <- tc.rawEvent

			evt := <-pipelineInput

			assert.Equal(t, tc.rawEvent.Content, evt.Content())
			assert.Equal(t, tc.want.IngestionTime, evt.IngestionTime)
			assert.Equal(t, tc.want.Tags, evt.Tags)
		})
	}
}
