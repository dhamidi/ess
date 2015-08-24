package ess

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
)

// EventsOnDisk is a persistent, file-based implementation of an
// EventStore.
//
// Events are serialized as JSON and appended to a log file.  Storing
// and replaying events access the disk.  File handles are kept open
// no longer than necessary.
type EventsOnDisk struct {
	filename string
	clock    Clock
}

// NewEventsOnDisk returns an new instance appending events to file
// and using clock for marking events as persisted.
func NewEventsOnDisk(file string, clock Clock) (*EventsOnDisk, error) {
	return &EventsOnDisk{
		filename: filepath.Clean(file),
		clock:    clock,
	}, nil
}

// Store stores events by serializing them as JSON and appending them
// to the configured log file.  Intermediate directories are created.
func (self *EventsOnDisk) Store(events []*Event) error {
	os.MkdirAll(filepath.Dir(self.filename), 0700)
	out, err := os.OpenFile(self.filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer out.Close()

	enc := json.NewEncoder(out)
	for _, event := range events {
		event.Persist(self.clock)
		if err := enc.Encode(event); err != nil {
			return err
		}
	}

	return nil
}

// Replay replays all events matching streamId using receiver.
//
// Events are deserialized from the log file and then passed to
// receiver.
//
// Use "*" as the streamId to match all events.
func (self *EventsOnDisk) Replay(streamId string, receiver EventHandler) error {
	in, err := os.Open(self.filename)
	if err != nil {
		return err
	}

	dec := json.NewDecoder(in)
	for {
		event := Event{}
		err := dec.Decode(&event)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if streamId == "*" || streamId == event.StreamId {
			receiver.HandleEvent(&event)
		}
	}

	return nil
}
