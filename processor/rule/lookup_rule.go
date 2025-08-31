package rule

import (
	"fmt"

	"github.com/mrtc0/conduit/event"
	"github.com/mrtc0/conduit/log"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

var (
	_ Rule[any] = (*LookupRule[any])(nil)
)

type LookupTableEntry map[string]string

type LookupTable map[string]LookupTableEntry

type LookupRule[T any] struct {
	Table  LookupTable
	Source string
	Target string
}

// NewLookupRule creates a new LookupRule with the specified lookup table, source field, and target field.
// The lookup rule maps values from the source field to the target field using the provided lookup table.
func NewLookupRule[T any](table LookupTable, source, target string) *LookupRule[T] {
	return &LookupRule[T]{
		Table:  table,
		Source: source,
		Target: target,
	}
}

func (r *LookupRule[T]) RuleType() RuleType {
	return TypeTransform
}

func (r *LookupRule[T]) Apply(evt *event.Event[T]) Result[T] {
	var err error

	data, err := evt.MarshalJSON()
	if err != nil {
		return TransformResult[T]{
			Event: evt,
		}
	}

	result := gjson.Get(string(data), r.Source)

	entry, exists := r.Table[result.String()]
	if !exists {
		return TransformResult[T]{
			Event: evt,
		}
	}

	enrichedMessage := string(data)
	for key, value := range entry {
		targetAttribute := r.Target + "." + key
		enrichedMessage, err = sjson.Set(enrichedMessage, targetAttribute, value)
		if err != nil {
			log.Warn(
				fmt.Sprintf("LookupRule failed setting attribute %s: %v", targetAttribute, err),
			)

			continue
		}
	}

	if err := evt.UnmarshalJSON([]byte(enrichedMessage)); err != nil {
		log.Warn(fmt.Sprintf("LookupRule failed unmarshalling enriched message: %v", err))
	}

	return TransformResult[T]{
		Event: evt,
	}
}
