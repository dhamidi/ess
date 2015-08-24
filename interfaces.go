package ess

import (
	"encoding"
	"time"
)

// Aggregate defines the necessary methods for acting interacting with
// an event sourced application.
//
// Aggregates receives requests for state changes in the form of
// commands and emit events via an EventPublisher.  In order to
// reconstruct an aggregate's state, it needs to be able to handle
// events as well.
type Aggregate interface {
	// Id returns a string uniquely identifying the aggregate.
	// The aggregate's id is used for routing commands to it and
	// associating emitted events with this aggregate.
	Id() string

	// PublishWith configures the aggregate to emit events using
	// publisher.
	PublishWith(publisher EventPublisher) Aggregate

	CommandHandler
	EventHandler
}

// Clock is an interface for providing the current time.
//
// This interface exists to decouple objects that need access to the
// current time from the system time.  Mainly used for testing.
type Clock interface {
	Now() time.Time
}

// Value is defines the interface for converting text into Go values.
// A value is used for capturing, sanitizing and validating the
// parameters accepted by commands.
type Value interface {
	encoding.TextUnmarshaler

	// String should convert the value to a string representing
	// this value's data.
	String() string

	// Copy creates a new instance with the same internal state as
	// this value instance.
	Copy() Value
}

// EventPublisher defines the interface for publishing events in
// aggregates.
type EventPublisher interface {
	// PublishEvent queues event for publishing.
	PublishEvent(event *Event) EventPublisher
}

// CommandHandler defines the interface for handling commands.
type CommandHandler interface {
	// HandleCommand tries to process command.  If command cannot
	// be processed due to a validation of business rules, a
	// *ValidationError should be returned.  In that case, no
	// event should be emitted.
	HandleCommand(command *Command) error
}

// EventHandler defines the interface for processing events.
type EventHandler interface {
	HandleEvent(event *Event)
}

// EventHandlerFunc is a wrapper type to allow a function to fulfull
// the EventHandler interface by calling the function.
type EventHandlerFunc func(event *Event)

// HandleEvent implements the EventHandler interface.
func (self EventHandlerFunc) HandleEvent(event *Event) { self(event) }

// EventStore defines the necessary operations for persisting events
// and restoring application state from the log of persisted events.
type EventStore interface {
	// Store append events to the store in a manner that allows
	// them to be retrieved by Replay.  The returned error is
	// implementation defined.
	Store(events []*Event) error

	// Replay replays the events belonging to the stream
	// identified by streamId using receiver.
	//
	// Using "*" as the streamId select all events, regardless of
	// the event's actual stream id.
	//
	// Any error returned is implementation defined.
	Replay(streamId string, receiver EventHandler) error
}

// Form defines how to access form values.  This allows commands to
// fill in parameters automatically.
//
// This interface is modelled after net/http.Request so that http
// requests can be used where a Form is required.
type Form interface {
	// FromValue returns the string value associated with the form
	// field "field".
	FormValue(field string) string
}
