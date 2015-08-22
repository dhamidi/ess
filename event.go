package ess

import "time"

type Event struct {
	Id          string
	StreamId    string
	Name        string
	OccurredOn  time.Time
	PersistedAt time.Time
	Payload     map[string]interface{}
}

func NewEvent(name string) *Event {
	return &Event{
		Name:    name,
		Payload: map[string]interface{}{},
	}
}

func (self *Event) For(source Aggregate) *Event {
	self.StreamId = source.Id()
	return self
}

func (self *Event) Add(name string, value interface{}) *Event {
	self.Payload[name] = value
	return self
}

func (self *Event) Occur(clock Clock) *Event {
	self.OccurredOn = clock.Now()
	return self
}

func (self *Event) Persist(clock Clock) *Event {
	self.PersistedAt = clock.Now()
	return self
}
