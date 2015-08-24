package ess

import "testing"

// EventStoreTest encapsulates the tests for the EventStore interface.
// Any compliant implementation of an EventStore should pass these
// tests.
//
// This type is public so that implementations of an EventStore
// outside of this package can be tested.
type EventStoreTest struct {
	// SetUp is responsible for creating a new EventStore
	// instance.  It is called before each test.
	SetUp func(t *testing.T) EventStore

	// TearDown is responsible for doing any cleanup work.  It is
	// called at the end of each test.
	TearDown func()
}

// NewEventStoreTest returns a new test suite using setup as the test
// setup function.  TearDown is set to do nothing.
func NewEventStoreTest(setup func(t *testing.T) EventStore) *EventStoreTest {
	return &EventStoreTest{
		SetUp:    setup,
		TearDown: func() {},
	}
}

// Run runs all tests.
func (self *EventStoreTest) Run(t *testing.T) {
	self.testStoredEventsCanBeReplayedByStreamId(t)
	self.testStoredEventsCanBeReplayedOverAllStreams(t)
}

func (self *EventStoreTest) testStoredEventsCanBeReplayedByStreamId(t *testing.T) {
	store := self.SetUp(t)
	t.Logf("testStoredEventsCanBeReplayedByStreamId %T", store)
	defer self.TearDown()

	subject := newTestAggregate("id")
	other := newTestAggregate("other")

	history := []*Event{
		NewEvent("test.run-1").For(subject).Add("param", "value"),
		NewEvent("test.run-1").For(other).Add("param", "other"),
		NewEvent("test.run-2").For(subject).Add("param", "new-value"),
	}

	if err := store.Store(history); err != nil {
		t.Fatal(err)
	}

	seen := []string{}
	if err := store.Replay(subject.Id(), EventHandlerFunc(func(event *Event) {
		seen = append(seen, event.Name)
	})); err != nil {
		t.Fatal(err)
	}

	if got, want := len(seen), 2; got != want {
		t.Fatalf(`len(seen) = %v; want %v`, got, want)
	}

	if got, want := seen[0], history[0].Name; got != want {
		t.Errorf(`seen[0] = %v; want %v`, got, want)
	}

	if got, want := seen[1], history[2].Name; got != want {
		t.Errorf(`seen[1] = %v; want %v`, got, want)
	}

}

func (self *EventStoreTest) testStoredEventsCanBeReplayedOverAllStreams(t *testing.T) {
	store := self.SetUp(t)
	t.Logf("testStoredEventsCanBeReplayedOverAllStreams %T", store)
	defer self.TearDown()

	subject := newTestAggregate("id")
	other := newTestAggregate("other")

	history := []*Event{
		NewEvent("test.run-1").For(subject).Add("param", "value"),
		NewEvent("test.run-1").For(other).Add("param", "other"),
		NewEvent("test.run-2").For(subject).Add("param", "new-value"),
	}

	if err := store.Store(history); err != nil {
		t.Fatal(err)
	}

	seen := []string{}
	if err := store.Replay("*", EventHandlerFunc(func(event *Event) {
		seen = append(seen, event.Name)
	})); err != nil {
		t.Fatal(err)
	}

	if got, want := len(seen), 3; got != want {
		t.Fatalf(`len(seen) = %v; want %v`, got, want)
	}

	if got, want := seen[0], history[0].Name; got != want {
		t.Errorf(`seen[0] = %v; want %v`, got, want)
	}

	if got, want := seen[1], history[1].Name; got != want {
		t.Errorf(`seen[1] = %v; want %v`, got, want)
	}

	if got, want := seen[2], history[2].Name; got != want {
		t.Errorf(`seen[2] = %v; want %v`, got, want)
	}
}
