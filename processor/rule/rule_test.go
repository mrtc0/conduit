package rule_test

import (
	"testing"

	"github.com/mrtc0/conduit/event"
	"github.com/mrtc0/conduit/processor/rule"
	"github.com/mrtc0/conduit/testutils"
	"github.com/stretchr/testify/assert"
)

func TestProcessingRule_Apply(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		rule       rule.Rule[*testutils.DummyEvent]
		wantResult rule.Result[*testutils.DummyEvent]
	}{
		"filter rule with drop event": {
			rule: rule.NewRule(
				"drop-all",
				"Drop all events",
				rule.TypeFilter,
				func(e *event.Event[*testutils.DummyEvent]) rule.Result[*testutils.DummyEvent] {
					return rule.FilterResult[*testutils.DummyEvent]{
						Drop:   true,
						Reason: "Dropping all events for testing",
					}
				},
			),
			wantResult: rule.FilterResult[*testutils.DummyEvent]{
				Drop:   true,
				Reason: "Dropping all events for testing",
			},
		},
		"filter rule with allow event": {
			rule: rule.NewRule(
				"allow-all",
				"Allow all events",
				rule.TypeFilter,
				func(e *event.Event[*testutils.DummyEvent]) rule.Result[*testutils.DummyEvent] {
					return rule.FilterResult[*testutils.DummyEvent]{
						Drop:   false,
						Reason: "Allowing all events for testing",
					}
				},
			),
			wantResult: rule.FilterResult[*testutils.DummyEvent]{
				Drop:   false,
				Reason: "Allowing all events for testing",
			},
		},
		"transform rule": {
			rule: rule.NewRule(
				"transform-event",
				"Transform event",
				rule.TypeTransform,
				func(e *event.Event[*testutils.DummyEvent]) rule.Result[*testutils.DummyEvent] {
					e.Content().Name = "Transformed"

					return rule.TransformResult[*testutils.DummyEvent]{
						Event: e,
					}
				},
			),
			wantResult: rule.TransformResult[*testutils.DummyEvent]{
				Event: event.NewEvent(&event.RawEvent[*testutils.DummyEvent]{
					Content: &testutils.DummyEvent{ID: "123", Name: "Transformed"},
				}),
			},
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dummyEvent := &testutils.DummyEvent{ID: "123", Name: "Test Event"}

			evt := event.NewEvent(&event.RawEvent[*testutils.DummyEvent]{
				Content: dummyEvent,
			})

			result := tt.rule.Apply(evt)
			assert.Equal(t, tt.wantResult, result)
		})
	}
}
