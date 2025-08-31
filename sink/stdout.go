package sink

import "os"

func NewStdoutSink[T any]() Sink[T] {
	return NewWriterSink[T](os.Stdout)
}
