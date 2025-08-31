package event

// RawEvent represents a raw event before it is sent to the pipeline.
type RawEvent[T any] struct {
	Metadata *Metadata
	Content  T
}

// NewRawEvent creates a new RawEvent with the given content.
func NewRawEvent[T any](content T, metadata *Metadata) *RawEvent[T] {
	return &RawEvent[T]{
		Metadata: metadata,
		Content:  content,
	}
}
