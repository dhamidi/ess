package ess

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

type EventStoreTest struct {
	SetUp    func(t *testing.T) EventStore
	TearDown func()
}

func NewEventStoreTest(setup func(t *testing.T) EventStore) *EventStoreTest {
	return &EventStoreTest{
		SetUp:    setup,
		TearDown: func() {},
	}
}

func (self *EventStoreTest) Run(t *testing.T) {
	self.testStoredEventsCanBeReplayedByStreamId(t)
	self.testStoredEventsCanBeReplayedOverAllStreams(t)
}

func (self *EventStoreTest) testStoredEventsCanBeReplayedByStreamId(t *testing.T) {
	store := self.SetUp(t)
	t.Logf("testStoredEventsCanBeReplayedByStreamId %T", store)
	defer self.TearDown()

	subject := NewTestAggregate("id")
	other := NewTestAggregate("other")

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

	subject := NewTestAggregate("id")
	other := NewTestAggregate("other")

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

func TestEventsInMemory_EventStoreBehavior(t *testing.T) {
	setup := func(t *testing.T) EventStore { return NewEventsInMemory() }
	suite := NewEventStoreTest(setup)
	suite.Run(t)
}

func TestEventsOnDisk_EventStoreBehavior(t *testing.T) {
	filename := filepath.Join(os.TempDir(), fmt.Sprintf("events-%s.json", os.Getpid()))
	teardown := func() {
		os.Remove(filename)
	}
	setup := func(t *testing.T) EventStore {
		store, err := NewEventsOnDisk(filename, SystemClock)
		if err != nil {
			t.Fatalf("EventsOnDisk setup [filename=%q]: %s", filename, err)
		}
		return store
	}

	suite := NewEventStoreTest(setup)
	suite.TearDown = teardown

	suite.Run(t)
}
