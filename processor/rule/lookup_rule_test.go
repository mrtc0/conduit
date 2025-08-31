package rule_test

import (
	"fmt"
	"testing"

	"github.com/mrtc0/conduit/event"
	"github.com/mrtc0/conduit/processor/rule"
	"github.com/stretchr/testify/assert"
)

func TestLookupRule(t *testing.T) {
	t.Parallel()

	type RequestCustomerDetails struct {
		Name string `json:"name,omitempty"`
		Plan string `json:"plan,omitempty"`
	}

	type RequestCustomer struct {
		ID string `json:"id"`
		// Details will be enriched by the lookup rule
		Details RequestCustomerDetails `json:"details,omitempty"`
	}

	type RequestEvent struct {
		Method   string          `json:"method"`
		Customer RequestCustomer `json:"customer"`
	}

	table := rule.LookupTable{
		"123": rule.LookupTableEntry{
			"name": "Big Company",
			"plan": "Premium",
		},
		"456": rule.LookupTableEntry{
			"name": "Small Business",
			"plan": "Basic",
		},
		"789": rule.LookupTableEntry{
			"name": "Individual",
			"plan": "Free",
		},
	}

	r := rule.NewLookupRule[RequestEvent](table, "customer.id", "customer.details")

	testCases := []struct {
		input  *event.Event[RequestEvent]
		expect RequestEvent
	}{
		{
			input: event.NewEvent(
				&event.RawEvent[RequestEvent]{
					Content: RequestEvent{Method: "GET", Customer: RequestCustomer{ID: "123"}},
				},
			),
			expect: RequestEvent{
				Method: "GET",
				Customer: RequestCustomer{
					ID:      "123",
					Details: RequestCustomerDetails{Name: "Big Company", Plan: "Premium"},
				},
			},
		},
		{
			input: event.NewEvent(
				&event.RawEvent[RequestEvent]{
					Content: RequestEvent{Method: "GET", Customer: RequestCustomer{ID: "456"}},
				},
			),
			expect: RequestEvent{
				Method: "GET",
				Customer: RequestCustomer{
					ID:      "456",
					Details: RequestCustomerDetails{Name: "Small Business", Plan: "Basic"},
				},
			},
		},
		{
			input: event.NewEvent(
				&event.RawEvent[RequestEvent]{
					Content: RequestEvent{Method: "GET", Customer: RequestCustomer{ID: "789"}},
				},
			),
			expect: RequestEvent{
				Method: "GET",
				Customer: RequestCustomer{
					ID:      "789",
					Details: RequestCustomerDetails{Name: "Individual", Plan: "Free"},
				},
			},
		},
		{
			input: event.NewEvent(
				&event.RawEvent[RequestEvent]{
					Content: RequestEvent{Method: "GET", Customer: RequestCustomer{ID: "999"}},
				},
			),
			expect: RequestEvent{
				Method:   "GET",
				Customer: RequestCustomer{ID: "999"},
			}, // No match in the lookup table, so no
		},
		{
			input: event.NewEvent(
				&event.RawEvent[RequestEvent]{
					Content: RequestEvent{Method: "GET", Customer: RequestCustomer{}},
				},
			),
			expect: RequestEvent{
				Method:   "GET",
				Customer: RequestCustomer{},
			}, // No id field, no enrichment
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()

			processedResult := r.Apply(tc.input)
			transformResult, ok := processedResult.(rule.TransformResult[RequestEvent])
			assert.True(t, ok)

			assert.Equal(t, tc.expect, transformResult.Event.Content())
		})
	}
}
