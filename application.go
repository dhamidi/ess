package ess

import (
	"log"
	"os"
)

// Application represents an event sourced application.
//
// Any interaction with an application happens by sending it commands.
//
// Commands are messages expressing user intent and lead to changes of
// the application state.  Every command is routed to an object
// handling the application's business logic, called an Aggregate.
//
// Processing a command by an aggregate either fails or produces
// events.  An event records a state change.  The application appends
// all events produced in this manner to an append-only log, the
// EventStore.
//
// Every time a command is processed, the object handling the command
// is passed all the previous events that it emitted, so that it can
// reconstruct any internal state necessary for it to function.
//
// If a command has been processed successfully and emitted events
// have been stored, all events are passed to the projections
// registered with the application.
//
// A projection accepts events and produces a secondary data model
// which is used for querying and represents the current application
// state.  Multiple such models can exist in parallel.  By using
// projections an application can maintain models that are optimized
// for serving a specific use case.  Examples range from regenerating
// static files over maintaining a normalized relational database to
// updating a search index.
//
// When the application starts the whole history is replayed through
// all projections.  This restricts projections to idempotent
// operations.
type Application struct {
	name        string
	clock       Clock
	store       EventStore
	logger      *log.Logger
	projections map[string]EventHandler
}

// NewApplication creates a new application instance with reasonable
// default settings.  Events are stored in memory only and
// informational messages are logged to standard error.
func NewApplication(name string) *Application {
	return &Application{
		name:        name,
		logger:      log.New(os.Stderr, name+" ", log.LstdFlags),
		store:       NewEventsInMemory(),
		clock:       SystemClock,
		projections: map[string]EventHandler{},
	}
}

// WithLogger sets the application's logger to logger.
func (self *Application) WithLogger(logger *log.Logger) *Application {
	self.logger = logger
	return self
}

// WithStore sets the application's event store to store.  Do not call
// this method after Init has been called.
func (self *Application) WithStore(store EventStore) *Application {
	self.store = store
	return self
}

// WithProjection registers projection with name at the application.
func (self *Application) WithProjection(name string, projection EventHandler) *Application {
	self.projections[name] = projection
	return self
}

// Project passes event to all of the application's projections.
func (self *Application) Project(event *Event) {
	for name, handler := range self.projections {
		self.logger.Printf("PROJECT %s TO %s", event.Name, name)
		handler.HandleEvent(event)
	}
}

// Init reconstructs application state from history.  Call this method
// once initially after configuring your application.
func (self *Application) Init() error {
	return self.store.Replay("*", EventHandlerFunc(self.Project))
}

// Send sends command to the application for processing.  Send is not
// thread safe.
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

	for _, event := range events {
		self.Project(event)
	}

	return NewSuccessResult(receiver)
}
