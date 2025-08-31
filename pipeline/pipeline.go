package pipeline

import (
	"context"
	"time"

	"github.com/mrtc0/conduit/event"
	"github.com/mrtc0/conduit/processor"
	"github.com/mrtc0/conduit/processor/rule"
	"github.com/mrtc0/conduit/strategy"
)

type StrategyOption struct {
	StrategyType  strategy.StrategyType
	BufferLimit   int
	FlushInterval time.Duration
}

type Pipeline[T any] struct {
	input         chan *event.Event[T]
	strategyInput chan *event.Event[T]

	processor *processor.Processor[T]
	strategy  strategy.SendingStrategy[T]
}

func NewPipeline[T any](
	processingRules []rule.Rule[T],
	strategyOption *StrategyOption,
	sinkInput chan<- *event.Payload[T],
) *Pipeline[T] {
	input := make(chan *event.Event[T])
	strategyInput := make(chan *event.Event[T])

	processor := processor.NewProcessor(processingRules, input, strategyInput)
	strategy := newStrategy(strategyInput, sinkInput, strategyOption)

	return &Pipeline[T]{
		processor:     processor,
		strategy:      strategy,
		input:         input,
		strategyInput: strategyInput,
	}
}

func (p *Pipeline[T]) Input() chan *event.Event[T] {
	return p.input
}

func (p *Pipeline[T]) Start() {
	p.strategy.Start()
	p.processor.Start()
}

func (p *Pipeline[T]) Stop() error {
	close(p.input)
	p.processor.WaitStop()

	close(p.strategyInput)
	p.strategy.WaitStop()

	return nil
}

func (p *Pipeline[T]) Flush(ctx context.Context) {
	p.processor.Flush(ctx)
	p.strategy.Flush(ctx)
}

func newStrategy[T any](
	inputChan strategy.InputChannel[T],
	outputChan chan<- *event.Payload[T],
	opt *StrategyOption,
) strategy.SendingStrategy[T] {
	switch opt.StrategyType {
	case strategy.Batch:
		return strategy.NewBatchStrategy(inputChan, outputChan, opt.FlushInterval, opt.BufferLimit)
	default: // Default to Stream strategy if no specific type is provided
		return strategy.NewStreamStrategy(inputChan, outputChan)
	}
}
