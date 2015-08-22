package ess

type EventsInMemory struct {
	events []*Event
}

func NewEventsInMemory() *EventsInMemory {
	return &EventsInMemory{
		events: []*Event{},
	}
}

func (self *EventsInMemory) Store(events []*Event) error {
	self.events = append(self.events, events...)
	return nil
}

func (self *EventsInMemory) Replay(streamId string, receiver EventHandler) error {
	for _, event := range self.events {
		if streamId == "*" || streamId == event.StreamId {
			receiver.HandleEvent(event)
		}
	}
	return nil
}

func (self *EventsInMemory) PublishEvent(event *Event) EventPublisher {
	self.events = append(self.events, event)
	return self
}

func (self *EventsInMemory) Events() []*Event {
	return self.events
}
