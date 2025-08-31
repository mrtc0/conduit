package event

import (
	"bytes"
	"encoding/json"
)

// Event represents an event processed in the pipeline.
type Event[T any] struct {
	Metadata

	content T
}

func NewEvent[T any](rawEvent *RawEvent[T]) *Event[T] {
	metadata := Metadata{
		Tags: Tags{},
	}

	if rawEvent.Metadata != nil {
		if rawEvent.Metadata.Tags != nil {
			metadata.Tags = rawEvent.Metadata.Tags
		}

		if !rawEvent.Metadata.IngestionTime.IsZero() {
			metadata.IngestionTime = rawEvent.Metadata.IngestionTime
		}
	}

	return &Event[T]{
		Metadata: metadata,
		content:  rawEvent.Content,
	}
}

func (e *Event[T]) Content() T {
	return e.content
}

func (e *Event[T]) SetContent(content T) {
	e.content = content
}

func (e *Event[T]) MarshalJSON() ([]byte, error) {
	b := &bytes.Buffer{}

	encoder := json.NewEncoder(b)
	encoder.SetEscapeHTML(true)
	if err := encoder.Encode(e.content); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (e *Event[T]) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &e.content); err != nil {
		return err
	}

	return nil
}
