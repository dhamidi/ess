package ess

import (
	"encoding"
	"time"
)

type Aggregate interface {
	Id() string
	PublishWith(publisher EventPublisher) Aggregate
	CommandHandler
	EventHandler
}

type Clock interface {
	Now() time.Time
}

type Value interface {
	encoding.TextUnmarshaler
	String() string
	Copy() Value
}

type EventPublisher interface {
	PublishEvent(event *Event) EventPublisher
}

type CommandHandler interface {
	HandleCommand(command *Command) error
}

type EventHandler interface {
	HandleEvent(event *Event)
}

type EventStore interface {
	Store(events []*Event) error
	Replay(streamId string, receiver EventHandler) error
}

type Form interface {
	FormValue(field string) string
}
