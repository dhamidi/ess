package ess

import "time"

// Event represents a state change that has occurred.  Events are
// named in the past tense, e.g. "user.signed-up".
type Event struct {
	// Id is the unique identifier of this event.
	Id string

	// StreamId is the id of the aggregate that emitted this
	// event.
	StreamId string

	// Name names the type of the event.
	Name string

	// OccurredOn is the time at which the application has seen
	// the event.
	OccurredOn time.Time

	// PersistedAt is the time at which the event has been written
	// to persistent storage.
	PersistedAt time.Time

	// Payload is additional data that needed to be recorded with
	// the event in order to reconstruct state.
	Payload map[string]interface{}
}

// NewEvent creates a new, empty event of type name.
func NewEvent(name string) *Event {
	return &Event{
		Name:    name,
		Payload: map[string]interface{}{},
	}
}

// For marks the event as being emitted by source.
func (self *Event) For(source Aggregate) *Event {
	self.StreamId = source.Id()
	return self
}

// Add sets the payload for the field name to value.
func (self *Event) Add(name string, value interface{}) *Event {
	self.Payload[name] = value
	return self
}

// Occur marks the occurrence time of the event according to clock.
func (self *Event) Occur(clock Clock) *Event {
	self.OccurredOn = clock.Now()
	return self
}

// Persist marks the time of persisting the event according to clock.
func (self *Event) Persist(clock Clock) *Event {
	self.PersistedAt = clock.Now()
	return self
}
