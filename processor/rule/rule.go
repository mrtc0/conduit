package rule

import (
	"github.com/mrtc0/conduit/event"
)

type RuleType int

const (
	TypeFilter RuleType = iota
	TypeTransform
)

type Result[T any] interface {
	TypeOf() RuleType
}

type FilterResult[T any] struct {
	// Drop indicates whether the event should be dropped (filtered out).
	Drop bool
	// Reason indicates the reason for filtering out the event.
	Reason string
}

func (r FilterResult[T]) TypeOf() RuleType {
	return TypeFilter
}

type TransformResult[T any] struct {
	Event *event.Event[T] // The transformed event
}

func (r TransformResult[T]) TypeOf() RuleType {
	return TypeTransform
}

// Rule defines the interface for processing rules that can be applied to events.
// Implementations of this interface must provide a method to apply the rule to a event
// and a method to identify the type of rule.
type Rule[T any] interface {
	// Apply applies the processing rule to the given event.
	// It returns the processed event and a boolean indicating whether the event should be kept.
	// If the boolean is false, the event is considered filtered out and should not be passed further.
	Apply(*event.Event[T]) Result[T]
	// RuleType returns the type of the processing rule.
	RuleType() RuleType
}

type rule[T any] struct {
	Name        string
	Description string
	Type        RuleType

	// fn is the function that processes the event.
	fn func(*event.Event[T]) Result[T]
}

func NewRule[T any](
	name, description string,
	ruleType RuleType,
	fn func(*event.Event[T]) Result[T],
) Rule[T] {
	return &rule[T]{
		Name:        name,
		Description: description,
		Type:        ruleType,
		fn:          fn,
	}
}

func (r *rule[T]) Apply(evt *event.Event[T]) Result[T] {
	if r.fn == nil {
		return FilterResult[T]{
			Drop:   false, // Default behavior if no processing function is defined
			Reason: "No processing function defined",
		}
	}

	return r.fn(evt)
}

func (r *rule[T]) RuleType() RuleType {
	return r.Type
}
