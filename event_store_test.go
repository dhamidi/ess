package ess

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

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
