package pipeline

import (
	"context"

	"github.com/mrtc0/conduit/event"
	"github.com/mrtc0/conduit/processor/rule"
)

var _ Provider[any] = (*provider[any])(nil)

type Provider[T any] interface {
	Start()
	Stop() error
	Flush(ctx context.Context) error
	PipelineInput() chan *event.Event[T]
}

type provider[T any] struct {
	pipeline *Pipeline[T]
}

func NewProvider[T any](
	processingRules []rule.Rule[T],
	strategyOption *StrategyOption,
	sinkInput chan<- *event.Payload[T],
) *provider[T] {
	return &provider[T]{
		pipeline: NewPipeline(processingRules, strategyOption, sinkInput),
	}
}

func (p *provider[T]) Start() {
	p.pipeline.Start()
}

func (p *provider[T]) Stop() error {
	return p.pipeline.Stop()
}

func (p *provider[T]) Flush(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		p.pipeline.Flush(ctx)
	}

	return nil
}

func (p *provider[T]) PipelineInput() chan *event.Event[T] {
	return p.pipeline.Input()
}
