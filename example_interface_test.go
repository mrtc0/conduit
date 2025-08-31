package conduit_test

import (
	"log"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	cloudevent "github.com/cloudevents/sdk-go/v2/event"

	"github.com/mrtc0/conduit"
	"github.com/mrtc0/conduit/event"
	"github.com/mrtc0/conduit/processor/rule"
	"github.com/mrtc0/conduit/sink"
)

// Example demonstrates a simple streaming event processing example.
func Example() {
	eventsGenerator := func(eventType string, msg string) cloudevent.Event {
		evt := cloudevents.NewEvent()
		evt.SetSource("example/uri")
		evt.SetType(eventType)
		if err := evt.SetData(cloudevents.ApplicationJSON, map[string]string{"message": msg}); err != nil {
			panic(err)
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
	})

	c.Start()

	events := []*event.RawEvent[cloudevent.Event]{
		event.NewRawEvent(eventsGenerator("example.event.ping", "test message 1"), nil),
		event.NewRawEvent(eventsGenerator("example.event.pong", "test message 2"), nil),
		event.NewRawEvent(eventsGenerator("example.event.ping", "test message 3"), nil),
		event.NewRawEvent(eventsGenerator("example.event.pong", "test message 4"), nil),
	}

	for _, evt := range events {
		if err := c.Write(evt); err != nil {
			log.Printf("Error writing message: %v", err)
		}
		time.Sleep(500 * time.Millisecond) // Simulate some delay between messages
	}

	if err := c.Stop(); err != nil {
		log.Printf("Error stopping conduit: %v", err)
	}

	// Output:
	// {"specversion":"1.0","id":"","source":"example/uri","type":"example.event.pong","datacontenttype":"application/json","data":{"message":"test message 2"}}
	// {"specversion":"1.0","id":"","source":"example/uri","type":"example.event.pong","datacontenttype":"application/json","data":{"message":"test message 4"}}
}
