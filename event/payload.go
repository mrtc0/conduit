package event

// Payload represents an collection of Event ready to be sent to sink.
type Payload[T any] struct {
	Metadata *Metadata

	// The encoded content of the event.
	JSONEncodedContent []byte
}

func NewPayload[T any](metadata *Metadata, jsonEncodedContent []byte) *Payload[T] {
	return &Payload[T]{
		Metadata:           metadata,
		JSONEncodedContent: jsonEncodedContent,
	}
}
