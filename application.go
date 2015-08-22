package ess

import (
	"log"
	"os"
)

type Application struct {
	name   string
	clock  Clock
	store  EventStore
	logger *log.Logger
}

func NewApplication(name string) *Application {
	return &Application{
		name:   name,
		logger: log.New(os.Stderr, name+" ", log.LstdFlags),
		store:  NewEventsInMemory(),
		clock:  SystemClock,
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

func (self *Application) Init() error {
	return nil
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
