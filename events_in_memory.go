package ess

// EventsInMemory is an in-memory implementation of an event store.
type EventsInMemory struct {
	events []*Event
}

// NewEventsInMemory creates a new instance of this event store
// holding no events initially.
func NewEventsInMemory() *EventsInMemory {
	return &EventsInMemory{
		events: []*Event{},
	}
}

// Store stores the given events in this event store.  It never
// returns an error.
func (self *EventsInMemory) Store(events []*Event) error {
	self.events = append(self.events, events...)
	return nil
}

// Replay handles all events with a matching stream id using receiver.
// It never returns an error.
//
// Use "*" as the stream id to match all events.
func (self *EventsInMemory) Replay(streamId string, receiver EventHandler) error {
	for _, event := range self.events {
		if streamId == "*" || streamId == event.StreamId {
			receiver.HandleEvent(event)
		}
	}
	return nil
}

// PublishEvent stores event in this instance.  This method is
// implemented to satisfy the EventPublisher interface.
//
// Using an EventsInMemory instance as an event publisher allows for
// capturing events across aggregates and facilitates testing.
func (self *EventsInMemory) PublishEvent(event *Event) EventPublisher {
	self.events = append(self.events, event)
	return self
}

// Events returns all events stored by this instance.
func (self *EventsInMemory) Events() []*Event {
	return self.events
}
