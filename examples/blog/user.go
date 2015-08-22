package main

import "github.com/dhamidi/ess"

type User struct {
	id     string
	events ess.EventPublisher

	signedUp       bool
	password       string
	activeSessions map[string]bool
}

func NewUser(id string) *User {
	return &User{
		id:             id,
		activeSessions: map[string]bool{},
	}
}

func (self *User) HandleCommand(command *ess.Command) error {
	switch command.Name {
	case "sign-up":
		return self.SignUp(command)
	case "login":
		return self.Login(command.Get("session").String(), command.Get("password").(*ess.BcryptedPassword))
	case "logout":
		return self.Logout(command.Get("session").String())
	}
	return nil
}

func (self *User) SignUp(params *ess.Command) error {
	err := ess.NewValidationError()

	if self.signedUp {
		err.Add("username", "not_unique")
	}

	if err.Ok() {
		self.events.PublishEvent(
			ess.NewEvent("user.signed-up").
				For(self).
				Add("username", params.Get("username").String()).
				Add("password", params.Get("password").String()).
				Add("email", params.Get("email").String()),
		)
	}

	return err.Return()
}

func (self *User) Login(session string, password *ess.BcryptedPassword) error {
	err := ess.NewValidationError()

	if !self.signedUp {
		err.Add("user", "not_found")
	}

	if !password.Matches(self.password) {
		err.Add("password", "mismatch")
	}

	if err.Ok() {
		self.events.PublishEvent(
			ess.NewEvent("user.logged-in").
				For(self).
				Add("session", session),
		)
	}

	return err.Return()
}

func (self *User) Logout(session string) error {
	err := ess.NewValidationError()

	if !self.signedUp {
		err.Add("user", "not_found")
	}

	if session == "" {
		err.Add("session", "empty")
	}

	if !self.HasActiveSession(session) {
		err.Add("session", "expired")
	}

	if err.Ok() {
		self.events.PublishEvent(
			ess.NewEvent("user.logged-out").
				For(self).
				Add("session", session),
		)
	}

	return err.Return()
}

func (self *User) HasActiveSession(session string) bool {
	_, found := self.activeSessions[session]
	return found
}

func (self *User) HandleEvent(event *ess.Event) {
	switch event.Name {
	case "user.signed-up":
		self.signedUp = true
		self.password = event.Payload["password"].(string)
	case "user.logged-in":
		if session := event.Payload["session"]; session != nil {
			self.activeSessions[session.(string)] = true
		}
	}
}

func (self *User) Id() string {
	return self.id
}

func (self *User) PublishWith(events ess.EventPublisher) ess.Aggregate {
	self.events = events
	return self
}
