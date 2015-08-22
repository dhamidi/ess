package ess

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
)

type EventsOnDisk struct {
	filename string
	clock    Clock
}

func NewEventsOnDisk(file string, clock Clock) (*EventsOnDisk, error) {
	return &EventsOnDisk{
		filename: filepath.Clean(file),
		clock:    clock,
	}, nil
}

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
