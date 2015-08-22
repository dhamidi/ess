package main

import (
	"crypto/rand"
	"fmt"

	"github.com/dhamidi/ess"
)

func GenerateSessionId() string {
	id := make([]byte, 16)
	_, err := rand.Read(id)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", id)
}

type ProjectedUser struct {
	Username  string
	SessionId string
}

type AllSessionsInMemory struct {
	sessions map[string]*ProjectedUser
}

func NewAllSessionsInMemory() *AllSessionsInMemory {
	return &AllSessionsInMemory{
		sessions: map[string]*ProjectedUser{},
	}
}

func (self *AllSessionsInMemory) HandleEvent(event *ess.Event) {
	switch event.Name {
	case "user.logged-in":
		if session := event.Payload["session"]; session != nil {
			user := &ProjectedUser{
				Username:  event.StreamId,
				SessionId: session.(string),
			}
			self.sessions[session.(string)] = user
		}
	case "user.logged-out":
		if session := event.Payload["session"]; session != nil {
			delete(self.sessions, session.(string))
		}
	}
}

func (self *AllSessionsInMemory) ById(id string) (*ProjectedUser, error) {
	user, found := self.sessions[id]

	if found {
		return user, nil
	}

	return nil, ErrNotFound
}
