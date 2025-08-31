package event

import "time"

type Tags map[string]string

type Metadata struct {
	Tags          Tags
	IngestionTime time.Time
}
