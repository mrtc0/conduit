package sink

import (
	"fmt"
	"os"

	"github.com/mrtc0/conduit/event"
)

var _ Sink[any] = (*FileWriterSink[any])(nil)

type FileWriterSink[T any] struct {
	file *os.File
}

func (s FileWriterSink[T]) Write(payload *event.Payload[T]) error {
	if _, err := s.file.Write(payload.JSONEncodedContent); err != nil {
		return fmt.Errorf("failed to write to file sink: %w", err)
	}

	return nil
}

func (s FileWriterSink[T]) Close() error {
	if err := s.file.Sync(); err != nil {
		return fmt.Errorf("failed to sync file sink: %w", err)
	}
	return s.file.Close()
}

func NewFileSink[T any](filepath string) (Sink[T], error) {
	file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600) //#nosec G304
	if err != nil {
		return nil, err
	}

	return &FileWriterSink[T]{
		file: file,
	}, nil
}
