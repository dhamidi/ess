package ess

import (
	"log"
	"os"
)

type Application struct {
	name        string
	clock       Clock
	store       EventStore
	logger      *log.Logger
	projections map[string]EventHandler
}

func NewApplication(name string) *Application {
	return &Application{
		name:        name,
		logger:      log.New(os.Stderr, name+" ", log.LstdFlags),
		store:       NewEventsInMemory(),
		clock:       SystemClock,
		projections: map[string]EventHandler{},
	}
}

func (self *Application) WithLogger(logger *log.Logger) *Application {
	self.logger = logger
	return self
}

func (self *Application) WithStore(store EventStore) *Application {
	self.store = store
	return self
}

func (self *Application) WithProjection(name string, projection EventHandler) *Application {
	self.projections[name] = projection
	return self
}

func (self *Application) project(event *Event) {
	for _, handler := range self.projections {
		handler.HandleEvent(event)
	}
}

func (self *Application) Init() error {
	return self.store.Replay("*", EventHandlerFunc(self.project))
}

func (self *Application) Send(command *Command) *CommandResult {
	command.Acknowledge(self.clock)

	receiver := command.Receiver()

	if err := self.store.Replay(receiver.Id(), receiver); err != nil {
		return NewErrorResult(err)
	}

	transaction := NewEventsInMemory()
	receiver.PublishWith(transaction)

	self.logger.Printf("EXECUTE %s", command)
	if err := command.Execute(); err != nil {
		self.logger.Printf("DENY %s", err)
		return NewErrorResult(err)
	}

	events := transaction.Events()
	for _, event := range events {
		event.Occur(self.clock)
		self.logger.Printf("EVENT %s", event.Name)
	}
	if err := self.store.Store(events); err != nil {
		return NewErrorResult(err)
	}

	return NewSuccessResult(receiver)
}
