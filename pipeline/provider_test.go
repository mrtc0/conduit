package pipeline_test

import (
	"context"
	"testing"
	"time"

	"github.com/mrtc0/conduit/pipeline"
	"github.com/mrtc0/conduit/processor/rule"
	"github.com/mrtc0/conduit/sender"
	"github.com/mrtc0/conduit/sink"
	"github.com/mrtc0/conduit/strategy"
	"github.com/mrtc0/conduit/testutils"
	"github.com/stretchr/testify/assert"
)

func TestProvider_PipelineInput(t *testing.T) {
	t.Parallel()

	stdoutSink := sink.NewStdoutSink[testutils.DummyEvent]()
	sinkSender := sender.NewSender(stdoutSink, nil)
	sinkSender.Start()

	provider := pipeline.NewProvider(
		[]rule.Rule[testutils.DummyEvent]{},
		&pipeline.StrategyOption{StrategyType: strategy.Stream},
		sinkSender.In(),
	)

	provider.Start()
	defer func() {
		err := provider.Stop()
		assert.NoError(t, err)
	}()

	assert.NotNil(t, provider.PipelineInput())
}

func TestProvider_Flush(t *testing.T) {
	t.Parallel()

	stdoutSink := sink.NewStdoutSink[testutils.DummyEvent]()
	sinkSender := sender.NewSender(stdoutSink, nil)
	sinkSender.Start()

	provider := pipeline.NewProvider(
		[]rule.Rule[testutils.DummyEvent]{},
		&pipeline.StrategyOption{StrategyType: strategy.Stream},
		sinkSender.In(),
	)

	provider.Start()
	defer func() {
		err := provider.Stop()
		assert.NoError(t, err)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := provider.Flush(ctx)
	assert.NoError(t, err, "Flush should not return an error")
}
