package conduit

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/mrtc0/conduit/adapter"
	"github.com/mrtc0/conduit/event"
	"github.com/mrtc0/conduit/pipeline"
	"github.com/mrtc0/conduit/processor/rule"
	"github.com/mrtc0/conduit/sender"
	"github.com/mrtc0/conduit/sink"
	"github.com/mrtc0/conduit/source"
	"github.com/mrtc0/conduit/strategy"
)

var (
	// DefaultFlushTimeout is the default timeout for flushing the pipeline and sender.
	DefaultFlushTimeout = 30 * time.Second
)

type Conduit[T any] struct {
	inputChannel chan *event.RawEvent[T]

	adapter          *adapter.EventAdapter[T]
	pipelineProvider pipeline.Provider[T]
	sender           *sender.Sender[T]

	stopped bool

	mu sync.Mutex
}

type Config[T any] struct {
	// ProcessingRules is a list of processing rules to be applied to the messages.
	// These rules will be executed in the order they are defined.
	ProcessingRules []rule.Rule[T]
	// Sink is the sink where the processed messages will be sent.
	Sink sink.Sink[T]
	// Result is a channel where the results of the sending operation will be sent.
	Result chan *sink.Result[T]
	// SendingStrategy defines the strategy for sending messages.
	// If not specified, the StreamStrategy will be used.
	SendingStrategy SendingStrategy
}

type SendingStrategy struct {
	// Type is the type of the sending strategy.
	Type strategy.StrategyType

	// BufferLimitBytes specifies the maximum size of the buffer in bytes for the BatchStrategy.
	// When the buffer size exceeds this limit, it will be flushed immediately.
	BufferLimitBytes int

	// FlushInterval specifies the interval at which the BatchStrategy buffer will be flushed.
	// Even if the buffer is not full, it will be flushed after this interval.
	FlushInterval time.Duration
}

// New creates a new Conduit instance with the provided configuration.
func New[T any](config Config[T]) *Conduit[T] {
	strategyOpt := &pipeline.StrategyOption{
		StrategyType:  config.SendingStrategy.Type,
		BufferLimit:   config.SendingStrategy.BufferLimitBytes,
		FlushInterval: config.SendingStrategy.FlushInterval,
	}

	inputChannel := make(chan *event.RawEvent[T])

	sinkSender := sender.NewSender(config.Sink, config.Result)
	pp := pipeline.NewProvider(config.ProcessingRules, strategyOpt, sinkSender.In())

	source := &source.EventSource[T]{InputChannel: inputChannel}
	adapter := adapter.NewEventAdapter(source, pp.PipelineInput())

	return &Conduit[T]{
		inputChannel:     inputChannel,
		pipelineProvider: pp,
		adapter:          adapter,
		sender:           sinkSender,
		stopped:          true,
	}
}

// Start is starting to receive messages.
func (c *Conduit[T]) Start() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.sender.Start()
	c.pipelineProvider.Start()
	c.adapter.Start()

	c.stopped = false
}

// Write sends a raw message to the Conduit for processing.
func (c *Conduit[T]) Write(rawEvt *event.RawEvent[T]) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.stopped {
		return errors.New("conduit is stopped, cannot write messages")
	}

	c.inputChannel <- rawEvt

	return nil
}

// Stop stops the Conduit and flushes any remaining messages in the pipeline and sender.
func (c *Conduit[T]) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.stopped {
		return errors.New("conduit is already stopped")
	}

	close(c.inputChannel)

	c.adapter.WaitClose()

	ctx, cancel := context.WithTimeout(context.Background(), DefaultFlushTimeout)
	defer cancel()

	if err := c.pipelineProvider.Flush(ctx); err != nil {
		return fmt.Errorf("failed to flush pipeline: %w", err)
	}
	if err := c.sender.Flush(ctx); err != nil {
		return fmt.Errorf("failed to flush sender: %w", err)
	}

	if err := c.pipelineProvider.Stop(); err != nil {
		return fmt.Errorf("failed to stop pipeline provider: %w", err)
	}
	if err := c.sender.Stop(); err != nil {
		return fmt.Errorf("failed to stop sender: %w", err)
	}

	c.stopped = true
	return nil
}
