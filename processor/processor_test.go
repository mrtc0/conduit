package processor_test

import (
	"strings"
	"testing"

	"github.com/mrtc0/conduit/event"
	"github.com/mrtc0/conduit/processor"
	"github.com/mrtc0/conduit/processor/rule"
	"github.com/mrtc0/conduit/testutils"
	"github.com/stretchr/testify/assert"
)

func TestProcessor_ApplyRules(t *testing.T) {
	t.Parallel()

	type expect struct {
		applyResult  bool
		appliedEvent *testutils.DummyEvent
	}

	testCases := map[string]struct {
		rules  []rule.Rule[*testutils.DummyEvent]
		input  *event.Event[*testutils.DummyEvent]
		expect expect
	}{
		"no rules": {
			rules: []rule.Rule[*testutils.DummyEvent]{},
			input: event.NewEvent(
				&event.RawEvent[*testutils.DummyEvent]{
					Content: &testutils.DummyEvent{ID: "123", Name: "Test Event"},
				},
			),
			expect: expect{
				applyResult:  true,
				appliedEvent: &testutils.DummyEvent{ID: "123", Name: "Test Event"},
			},
		},
		"filtered out": {
			rules: []rule.Rule[*testutils.DummyEvent]{
				rule.NewRule(
					"filter-rule",
					"Filter out event containing 'filter'",
					rule.TypeFilter,
					func(evt *event.Event[*testutils.DummyEvent]) rule.Result[*testutils.DummyEvent] {
						return rule.FilterResult[*testutils.DummyEvent]{
							Drop:   true,
							Reason: "Event contains 'filter'",
						}
					},
				),
			},
			input: event.NewEvent(
				&event.RawEvent[*testutils.DummyEvent]{
					Content: &testutils.DummyEvent{ID: "123", Name: "Test Event"},
				},
			),
			expect: expect{
				applyResult:  false,
				appliedEvent: &testutils.DummyEvent{ID: "123", Name: "Test Event"},
			},
		},
		"multiple rules": {
			rules: []rule.Rule[*testutils.DummyEvent]{
				rule.NewRule(
					"id-snake-case-transform",
					"Transform ID value to snake_case",
					rule.TypeTransform,
					func(evt *event.Event[*testutils.DummyEvent]) rule.Result[*testutils.DummyEvent] {
						content := evt.Content()
						content.ID = strings.ReplaceAll(content.ID, "-", "_")
						return rule.TransformResult[*testutils.DummyEvent]{
							Event: evt,
						}
					},
				),
				rule.NewRule(
					"name-uppercase-transform",
					"Transform Name value to uppercase",
					rule.TypeTransform,
					func(evt *event.Event[*testutils.DummyEvent]) rule.Result[*testutils.DummyEvent] {
						content := evt.Content()
						content.Name = strings.ToUpper(content.Name)
						evt.SetContent(content)
						return rule.TransformResult[*testutils.DummyEvent]{
							Event: evt,
						}
					},
				),
			},
			input: event.NewEvent(
				&event.RawEvent[*testutils.DummyEvent]{
					Content: &testutils.DummyEvent{ID: "123-456", Name: "Test Event"},
				},
			),
			expect: expect{
				applyResult:  true,
				appliedEvent: &testutils.DummyEvent{ID: "123_456", Name: "TEST EVENT"},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			processor := processor.NewProcessor(tc.rules, nil, nil)
			assert.Equal(t, tc.expect.applyResult, processor.ApplyRules(tc.input))
			assert.Equal(t, tc.expect.appliedEvent, tc.input.Content())
		})
	}
}
