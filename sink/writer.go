package sink

import (
	"fmt"
	"io"

	"github.com/mrtc0/conduit/event"
)

type WriterSink[T any] struct {
	Writer io.Writer
}

func NewWriterSink[T any](writer io.Writer) Sink[T] {
	return WriterSink[T]{
		Writer: writer,
	}
}

func (s WriterSink[T]) Write(payload *event.Payload[T]) error {
	if _, err := s.Writer.Write(payload.JSONEncodedContent); err != nil {
		return fmt.Errorf("failed to write payload to writer sink: %w", err)
	}

	return nil
}

func (s WriterSink[T]) Close() error {
	if closer, ok := s.Writer.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			return err
		}
	}

	return nil
}
