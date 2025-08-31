# Conduit

Conduit is an event processing library written in Go. It applies processing rules such as filtering and transformation to received events and sends the processed events to specified output destinations.

## Features

- **Type Safe**: Provides type-safe event processing using Go generics
- **Flexible Processing Rules**: Supports both filtering and transformation
- **Multiple Sending Strategies**: Supports streaming and batch sending

## Installation

```bash
go get github.com/mrtc0/conduit
```

## Basic Usage

### Simple Streaming Processing

In this example, streaming CloudEvents data and outputting it to Standard Output.

```go
package main

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

func main() {
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
                // If the event type is "example.event.ping", drop it
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
```

### Batch Processing

You can also modify the `SendingStrategy` to buffer data before sending it to the Sink.

```go
c := conduit.New(conduit.Config[MyEvent]{
    ProcessingRules: rules,
    Sink:            sink.NewStdoutSink[MyEvent](),
    SendingStrategy: conduit.SendingStrategy{
        Type:             strategy.Batch,
        // If the buffer limit is reached or the flush interval elapses, send the buffered events
        BufferLimitBytes: 1024,
        FlushInterval:    5 * time.Second,
    },
})
```

## Processing Rules

The entered event can be filtered and transformed.

### Filtering Rules

Rules for filtering events can be defined as follows:

```go
rule.NewRule(
    "example-filter",
    "Filter description",
    rule.TypeFilter,
    func(evt *event.Event[T]) rule.Result[T] {
        // Implement filtering conditions here
        if evt.Content().Something {
             return rule.FilterResult[T]{Drop: true, Reason: "reason for filtering"}
        }

        return rule.FilterResult[T]{
            Drop: false,
        }
    },
)
```

### Transformation Rules

Rules for modifying events can be defined as follows:

```go
type T struct {
    Value string
}

rule.NewRule(
    "example-transform",
    "Transform description",
    rule.TypeTransform,
    func(evt *event.Event[T]) rule.Result[T] {
        evt.Content().Value = "new value"

        return rule.TransformResult[T]{Event: evt}
        // => {Value: "new value"}
    },
)
```

### Built-in Rules

Built-in rules provide common functionalities that can be reused across different event types.

#### Lookup Rule

You can use lookup rules to define mappings between event attributes and human-readable values stored in a table.
For example, you can enrich events with CMDB information.

```go
type MyEvent struct {
    RequestContenxt struct {
        TenantID string `json:"tenantID"`
    } `json:"requestContext"`
}

// tenantTable is a lookup table for tenant information, keyed by tenant ID.
tenantTable := rule.LookupTable{
    "1": rule.LookupTableEntry{
        "name": "Big Corp",
        "plan": "Premium",
    },
    "2": rule.LookupTableEntry{
        "name": "Small Corp",
        "plan": "Basic",
    },
}

// If MyEvent.RequestContext.TenantID exists in tenantTable,
// the corresponding information is added to MyEvent.RequestContext.TenantDetails
lookupRule := rule.NewLookupRule[MyEvent](tenantTable, "requestContext.tenantID", "requestContext.tenantDetails")
// => { requestContext: { TenantID: "1", TenantDetails: { Name: "Big Corp", Plan: "Premium" } } }
```

## Sinks

When the event is processed, it is sent to the Sink. It can be output to standard output, written to a file, or sent as an HTTP request.

### Standard Output

```go
sink.NewStdoutSink[MyEvent]()
```

### File Output

```go
fileSink, err := sink.NewFileSink[MyEvent]("/path/to/output.json")
```

### Custom Writer

By implementing the `Sink` interface, you can use your own custom Sink.

```go
type Sink[T any] interface {
	// Write sends the processed event to the sink.
	Write(payload *event.Payload[T]) error
	// Close closes the sink and cleans up resources.
	Close() error
}
```
