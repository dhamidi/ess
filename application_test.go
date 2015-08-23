package ess

import (
	"testing"
	"time"
)

var (
	TestCommand = NewCommandDefinition("test").
			Field("param", TrimmedString()).
			Target(NewTestAggregateFromCommand)

	TheTime = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
)

type TestAggregate struct {
	id     string
	events EventPublisher
	error  error

	onEvent   func(event *Event)
	onCommand func(*TestAggregate)
}

func NewTestAggregateFromCommand(command *Command) Aggregate {
	return NewTestAggregate(command.Get("id").String())
}

func NewTestAggregate(id string) *TestAggregate {
	return &TestAggregate{id: id}
}

func (self *TestAggregate) FailWith(err error) *TestAggregate {
	self.error = err
	return self
}

func (self *TestAggregate) Id() string {
	return self.id
}

func (self *TestAggregate) HandleEvent(e *Event) {
	if self.onEvent != nil {
		self.onEvent(e)
	}
}

func (self *TestAggregate) HandleCommand(command *Command) error {
	if self.onCommand != nil {
		self.onCommand(self)
	}
	return self.error
}

func (self *TestAggregate) PublishWith(publisher EventPublisher) Aggregate {
	self.events = publisher
	return self
}

func NewTestApp() *Application {
	app := NewApplication("test")
	app.clock = &StaticClock{TheTime}
	return app
}

func TestApplication_Send_acknowledgesCommand(t *testing.T) {
	app := NewTestApp()
	cmd := TestCommand.NewCommand()
	result := app.Send(cmd)

	if err := result.Error(); err != nil {
		t.Fatal(err)
	}

	if got, want := cmd.Get("now").(*Time).Time, TheTime; !got.Equal(want) {
		t.Errorf(`cmd.Get("now").(*Time).Time = %q; want %q`, got, want)
	}
}

func TestApplication_Send_replaysHistoryOnReceiver(t *testing.T) {
	app := NewTestApp()
	seen := 0
	other := NewTestAggregate("other")
	receiver := NewTestAggregate("test")
	receiver.onEvent = func(*Event) { seen++ }
	history := []*Event{
		NewEvent("test.run").For(other),
		NewEvent("test.run").For(receiver),
		NewEvent("test.run").For(receiver),
	}
	app.store.Store(history)
	cmd := TestCommand.NewCommand()
	cmd.receiver = receiver
	result := app.Send(cmd)

	if err := result.Error(); err != nil {
		t.Fatal(err)
	}

	if got, want := seen, len(history)-1; got != want {
		t.Errorf("seen = %d; want %d", got, want)
	}
}

func TestApplication_Send_returnsErrorIfExecutingCommandFails(t *testing.T) {
	cmd := TestCommand.NewCommand()
	failure := NewValidationError().Add("param", "invalid")
	cmd.receiver = NewTestAggregate("test").FailWith(failure.Return())
	app := NewTestApp()
	result := app.Send(cmd)

	if err := result.Error(); err != failure {
		t.Errorf("result.Error() = %q; want %q", err, failure)
	}
}

func TestApplication_Send_marksOccurrenceOnEvents(t *testing.T) {
	app := NewTestApp()
	cmd := TestCommand.NewCommand()
	receiver := NewTestAggregate("test")
	cmd.receiver = receiver
	event := NewEvent("test.run").For(cmd.receiver)
	receiver.onCommand = func(agg *TestAggregate) {
		agg.events.PublishEvent(event)
	}

	result := app.Send(cmd)
	if err := result.Error(); err != nil {
		t.Fatal(err)
	}

	if got, want := event.OccurredOn, TheTime; !got.Equal(want) {
		t.Errorf("event.OccurredOn = %q; want %q", got, want)
	}
}

func TestApplication_Send_storesEvents(t *testing.T) {
	transaction := NewEventsInMemory()
	app := NewTestApp().WithStore(transaction)
	cmd := TestCommand.NewCommand()
	receiver := NewTestAggregate("test")
	cmd.receiver = receiver
	event := NewEvent("test.run").For(cmd.receiver)
	receiver.onCommand = func(agg *TestAggregate) {
		agg.events.PublishEvent(event)
	}

	result := app.Send(cmd)
	if err := result.Error(); err != nil {
		t.Fatal(err)
	}

	if got, want := transaction.Events()[0], event; got != want {
		t.Errorf("transaction.Events()[0] = %q; want %q", got, want)
	}
}

func TestApplication_Send_projectsEvents(t *testing.T) {
	projected := 0
	app := NewTestApp().
		WithProjection("test", EventHandlerFunc(func(*Event) {
		projected++
	}))
	cmd := TestCommand.NewCommand()
	receiver := NewTestAggregate("test")
	cmd.receiver = receiver
	event := NewEvent("test.run").For(cmd.receiver)
	receiver.onCommand = func(agg *TestAggregate) {
		agg.events.PublishEvent(event)
	}

	result := app.Send(cmd)
	if err := result.Error(); err != nil {
		t.Fatal(err)
	}

	if got, want := projected, 1; got != want {
		t.Errorf("projected = %d; want %d", got, want)
	}
}
