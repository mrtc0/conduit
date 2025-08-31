package conduit_test

import (
	"bytes"
	"testing"

	"github.com/mrtc0/conduit"
	"github.com/mrtc0/conduit/event"
	"github.com/mrtc0/conduit/processor/rule"
	"github.com/mrtc0/conduit/sink"
	"github.com/mrtc0/conduit/testutils"
	"github.com/stretchr/testify/assert"
)

func TestConduit_Write(t *testing.T) {
	t.Parallel()

	rules := []rule.Rule[testutils.DummyEvent]{}

	buf := &bytes.Buffer{}
	bufSink := sink.NewWriterSink[testutils.DummyEvent](buf)

	c := conduit.New(conduit.Config[testutils.DummyEvent]{
		ProcessingRules: rules,
		Sink:            bufSink,
	})
	c.Start()

	err := c.Write(event.NewRawEvent(testutils.DummyEvent{
		ID:   "123",
		Name: "Test Event",
	}, nil))
	assert.NoError(t, err)

	assert.NoError(t, c.Stop())

	assert.Equal(t, "{\"id\":\"123\",\"name\":\"Test Event\"}\n", buf.String())
}
