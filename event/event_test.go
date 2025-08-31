package event_test

import (
	"testing"
	"time"

	"github.com/mrtc0/conduit/event"
	"github.com/mrtc0/conduit/testutils"
	"github.com/stretchr/testify/assert"
)

func TestEvent_NewEvent(t *testing.T) {
	t.Parallel()

	content := testutils.DummyEvent{ID: "123", Name: "Test Event"}

	testCases := map[string]struct {
		rawEvent *event.RawEvent[testutils.DummyEvent]
		want     *event.Event[testutils.DummyEvent]
	}{
		"with no metadata": {
			rawEvent: event.NewRawEvent(content, nil),
			want: &event.Event[testutils.DummyEvent]{
				Metadata: event.Metadata{
					Tags: event.Tags{},
				},
			},
		},
		"with tags": {
			rawEvent: event.NewRawEvent(
				content,
				&event.Metadata{
					Tags: event.Tags{"tag1": "value1", "tag2": "value2"},
				},
			),
			want: &event.Event[testutils.DummyEvent]{
				Metadata: event.Metadata{
					Tags: event.Tags{"tag1": "value1", "tag2": "value2"},
				},
			},
		},
		"with ingested time": {
			rawEvent: event.NewRawEvent(
				content,
				&event.Metadata{
					IngestionTime: time.Date(2025, 7, 18, 13, 0, 0, 0, time.UTC),
				},
			),
			want: &event.Event[testutils.DummyEvent]{
				Metadata: event.Metadata{
					Tags:          event.Tags{},
					IngestionTime: time.Date(2025, 7, 18, 13, 0, 0, 0, time.UTC),
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.want.SetContent(content)

			evt := event.NewEvent(tc.rawEvent)

			assert.Equal(t, tc.want.Tags, evt.Tags)
			assert.Equal(t, tc.want.IngestionTime, evt.IngestionTime)
			assert.Equal(t, tc.want.Content(), evt.Content())
		})
	}
}

func TestEvent_MarshalJSON(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		event *testutils.DummyEvent
		want  string
	}{
		"without user identity": {
			event: &testutils.DummyEvent{
				ID:   "123",
				Name: "Test Event",
			},
			want: `{"id":"123","name":"Test Event"}`,
		},
		"with user identity": {
			event: &testutils.DummyEvent{
				ID:   "123",
				Name: "User Event",
				UserIdentity: &testutils.UserIdentity{
					ID:       "abc",
					Username: "alice",
				},
			},
			want: `{"id":"123","name":"User Event","user_identity":{"id":"abc","username":"alice"}}`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			rawEvent := event.NewRawEvent(tc.event, nil)

			evt := event.NewEvent(rawEvent)
			data, err := evt.MarshalJSON()
			assert.NoError(t, err)
			assert.JSONEq(t, tc.want, string(data))
		})
	}
}

func TestEvent_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		data []byte
		want testutils.DummyEvent
	}{
		"without user identity": {
			data: []byte(`{"id":"123","name":"Test Event"}`),
			want: testutils.DummyEvent{
				ID:   "123",
				Name: "Test Event",
			},
		},
		"with user identity": {
			data: []byte(
				`{"id":"123","name":"User Event","user_identity":{"id":"abc","username":"alice"}}`,
			),
			want: testutils.DummyEvent{
				ID:   "123",
				Name: "User Event",
				UserIdentity: &testutils.UserIdentity{
					ID:       "abc",
					Username: "alice",
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			evt := &event.Event[testutils.DummyEvent]{}
			err := evt.UnmarshalJSON(tc.data)
			assert.NoError(t, err)

			assert.Equal(t, tc.want, evt.Content())
		})
	}
}
