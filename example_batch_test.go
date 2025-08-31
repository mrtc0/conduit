package conduit_test

import (
	"fmt"
	"log"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	cloudevent "github.com/cloudevents/sdk-go/v2/event"

	"github.com/mrtc0/conduit"
	"github.com/mrtc0/conduit/event"
	"github.com/mrtc0/conduit/processor/rule"
	"github.com/mrtc0/conduit/sink"
	"github.com/mrtc0/conduit/strategy"
)

// Example_batchStrategy demonstrates how to use the Batch sending strategy
// The received messages are buffered and sent in bulk when the buffer reaches a certain size or after a certain time has passed.
// In this example, 10 messages are received, and since the buffer is flushed every 5 seconds, the Sink will receive the messages in two batches.
func Example_batchStrategy() {
	eventsGenerator := func(eventType string, msg string) cloudevent.Event {
		evt := cloudevents.NewEvent()
		evt.SetSource("example/uri")
		evt.SetType(eventType)
		err := evt.SetData(cloudevents.ApplicationJSON, map[string]string{"message": msg})
		if err != nil {
			panic(fmt.Sprintf("failed to set data: %v", err))
		}

		return evt
	}

	rules := []rule.Rule[cloudevent.Event]{
		rule.NewRule(
			"filtered-out-ping-events",
			"Filter out ping events",
			rule.TypeFilter,
			func(evt *event.Event[cloudevent.Event]) rule.Result[cloudevent.Event] {
				// Example processing logic
				content := evt.Content()
				if content.Type() == "example.event.ping" {
					return rule.FilterResult[cloudevent.Event]{
						Drop:   true,
						Reason: "not ping event",
					}
				}

				return rule.FilterResult[cloudevent.Event]{
					Drop: false,
				}
			},
		),
	}

	c := conduit.New(conduit.Config[cloudevent.Event]{
		ProcessingRules: rules,
		Sink:            sink.NewStdoutSink[cloudevent.Event](),
		SendingStrategy: conduit.SendingStrategy{
			Type:             strategy.Batch,
			BufferLimitBytes: 512,
			FlushInterval:    time.Duration(5 * time.Second),
		},
	})

	c.Start()

	quit := make(chan struct{}, 1)

	go func() {
		count := 1

		for range 20 {
			var evt *event.RawEvent[cloudevent.Event]

			if count%2 == 0 {
				evt = event.NewRawEvent(
					eventsGenerator(
						"example.event.ping",
						fmt.Sprintf("test message %d", count)),
					nil,
				)
			} else {
				evt = event.NewRawEvent(eventsGenerator("example.event.pong", fmt.Sprintf("test message %d", count)), nil)
			}

			if err := c.Write(evt); err != nil {
				log.Printf("Error writing message: %v", err)
			}

			time.Sleep(500 * time.Millisecond)
			count++
		}

		quit <- struct{}{}
	}()

	<-quit
	close(quit)

	if err := c.Stop(); err != nil {
		log.Printf("Error stopping conduit: %v", err)
	}

	// Output:
	// {"specversion":"1.0","id":"","source":"example/uri","type":"example.event.pong","datacontenttype":"application/json","data":{"message":"test message 1"}}
	// {"specversion":"1.0","id":"","source":"example/uri","type":"example.event.pong","datacontenttype":"application/json","data":{"message":"test message 3"}}
	// {"specversion":"1.0","id":"","source":"example/uri","type":"example.event.pong","datacontenttype":"application/json","data":{"message":"test message 5"}}
	// {"specversion":"1.0","id":"","source":"example/uri","type":"example.event.pong","datacontenttype":"application/json","data":{"message":"test message 7"}}
	// {"specversion":"1.0","id":"","source":"example/uri","type":"example.event.pong","datacontenttype":"application/json","data":{"message":"test message 9"}}
	// {"specversion":"1.0","id":"","source":"example/uri","type":"example.event.pong","datacontenttype":"application/json","data":{"message":"test message 11"}}
	// {"specversion":"1.0","id":"","source":"example/uri","type":"example.event.pong","datacontenttype":"application/json","data":{"message":"test message 13"}}
	// {"specversion":"1.0","id":"","source":"example/uri","type":"example.event.pong","datacontenttype":"application/json","data":{"message":"test message 15"}}
	// {"specversion":"1.0","id":"","source":"example/uri","type":"example.event.pong","datacontenttype":"application/json","data":{"message":"test message 17"}}
	// {"specversion":"1.0","id":"","source":"example/uri","type":"example.event.pong","datacontenttype":"application/json","data":{"message":"test message 19"}}
}
