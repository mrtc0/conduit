package pipeline_test

import (
	"context"
	"testing"

	"github.com/mrtc0/conduit/pipeline"
	"github.com/mrtc0/conduit/processor/rule"
	"github.com/mrtc0/conduit/sender"
	"github.com/mrtc0/conduit/sink"
	"github.com/mrtc0/conduit/strategy"
	"github.com/mrtc0/conduit/testutils"
)

func TestPipeline(t *testing.T) {
	t.Parallel()

	stdoutSink := sink.NewStdoutSink[testutils.DummyEvent]()
	sinkSender := sender.NewSender(stdoutSink, nil)
	sinkSender.Start()

	p := pipeline.NewPipeline(
		[]rule.Rule[testutils.DummyEvent]{},
		&pipeline.StrategyOption{StrategyType: strategy.Stream},
		sinkSender.In(),
	)

	p.Start()
	defer func() {
		if err := p.Stop(); err != nil {
			t.Errorf("Failed to stop pipeline: %v", err)
		}
	}()

	p.Flush(context.Background())
}
