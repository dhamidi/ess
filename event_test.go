package ess

import (
	"testing"
	"time"
)

func TestEvent_For_usesAggregateIdAsStreamId(t *testing.T) {
	aggregate := newTestAggregate("id")
	event := NewEvent("test.run").For(aggregate)

	if got, want := event.StreamId, aggregate.Id(); got != want {
		t.Errorf(`event.StreamId = %v; want %v`, got, want)
	}
}

func TestEvent_Add_addsFieldToPayload(t *testing.T) {
	event := NewEvent("test.run").
		Add("a", 1).
		Add("b", 2)

	if got, want := event.Payload["a"].(int), 1; got != want {
		t.Errorf(`event.Payload["a"].(int) = %v; want %v`, got, want)
	}

	if got, want := event.Payload["b"].(int), 2; got != want {
		t.Errorf(`event.Payload["b"].(int) = %v; want %v`, got, want)
	}

}

func TestEvent_Add_overwritesExistingValues(t *testing.T) {
	event := NewEvent("test.run").
		Add("a", 1).
		Add("a", 2)

	if got, want := event.Payload["a"], 2; got != want {
		t.Errorf(`event.Payload["a"] = %v; want %v`, got, want)
	}

}

func TestEvent_Occur_setsOccurredOnBasedOnClock(t *testing.T) {
	clock := &StaticClock{time.Now()}
	event := NewEvent("test.run").
		Occur(clock)

	if got, want := event.OccurredOn, clock.Time; !got.Equal(want) {
		t.Errorf(`event.OccurredOn = %v; want %v`, got, want)
	}
}

func TestEvent_Persist_setsPersistedAtBasedOnClock(t *testing.T) {
	clock := &StaticClock{time.Now()}
	event := NewEvent("test.run").
		Persist(clock)

	if got, want := event.PersistedAt, clock.Time; !got.Equal(want) {
		t.Errorf(`event.PersistedAt = %v; want %v`, got, want)
	}
}
