package processor

import (
	"context"

	"github.com/mrtc0/conduit/event"
	"github.com/mrtc0/conduit/processor/rule"
	"github.com/mrtc0/conduit/strategy"
)

type Processor[T any] struct {
	rules      []rule.Rule[T]
	inputChan  chan *event.Event[T]
	outputChan chan *event.Event[T]

	quit chan struct{}
}

func NewProcessor[T any](
	rules []rule.Rule[T],
	inputChan chan *event.Event[T],
	outputChan strategy.InputChannel[T],
) *Processor[T] {
	return &Processor[T]{
		rules:      rules,
		inputChan:  inputChan,
		outputChan: outputChan,
		quit:       make(chan struct{}),
	}
}

func (p *Processor[T]) Start() {
	go p.run()
}

func (p *Processor[T]) WaitStop() {
	<-p.quit
}

func (p *Processor[T]) Flush(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if len(p.inputChan) == 0 {
				return
			}
			evt := <-p.inputChan
			p.processMessage(evt)
			return
		}
	}
}

func (p *Processor[T]) run() {
	defer func() {
		close(p.quit)
	}()

	for evt := range p.inputChan {
		p.processMessage(evt)
	}
}

func (p *Processor[T]) processMessage(evt *event.Event[T]) {
	if passed := p.ApplyRules(evt); passed {
		p.outputChan <- evt
	}
}

func (p *Processor[T]) ApplyRules(evt *event.Event[T]) bool {
	for _, r := range p.rules {
		result := r.Apply(evt)

		if result.TypeOf() == rule.TypeFilter {
			// cast to FilterRuleResult
			filterResult, ok := result.(rule.FilterResult[T])
			if !ok {
				// If the result is not a FilterRuleResult, we cannot process it as a filter rule
				continue
			}
			if filterResult.Drop {
				return false
			}
		}

		if result.TypeOf() == rule.TypeTransform {
			// cast to TransformRuleResult
			transformResult, ok := result.(rule.TransformResult[T])
			if !ok {
				// If the result is not a TransformRuleResult, we cannot process it as a transform rule
				continue
			}
			if transformResult.Event != nil {
				evt = transformResult.Event // Update the event with the transformed one
			}
		}
	}

	return true // Message passed all rules
}
