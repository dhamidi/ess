package main

import "github.com/dhamidi/ess"

type Post struct {
	events  ess.EventPublisher
	id      string
	written bool
}

func NewPost(id string) *Post {
	return &Post{id: id}
}

func (self *Post) Id() string { return self.id }
func (self *Post) PublishWith(publisher ess.EventPublisher) ess.Aggregate {
	self.events = publisher
	return self
}

func (self *Post) HandleEvent(event *ess.Event) {
	switch event.Name {
	case "post.written":
		self.written = true
	}
}

func (self *Post) HandleCommand(command *ess.Command) error {
	switch command.Name {
	case "write-post":
		return self.Write(command.Get("title").String(), command.Get("body").String())
	case "edit-post":
		return self.Edit(
			command.Get("title").String(),
			command.Get("body").String(),
			command.Get("reason").String(),
		)
	}

	return nil
}

func (self *Post) Edit(title, body, reason string) error {
	err := ess.NewValidationError()

	if !self.written {
		err.Add("post", "not_found")
	}

	if title == "" {
		err.Add("title", "empty")
	}

	if body == "" {
		err.Add("body", "empty")
	}

	if reason == "" {
		err.Add("reason", "empty")
	}

	if err.Ok() {
		self.events.PublishEvent(
			ess.NewEvent("post.edited").
				For(self).
				Add("title", title).
				Add("body", body).
				Add("reason", reason),
		)
	}

	return err.Return()
}

func (self *Post) Write(title, body string) error {
	err := ess.NewValidationError()

	if self.written {
		err.Add("post", "not_unique")
	}

	if title == "" {
		err.Add("title", "empty")
	}

	if body == "" {
		err.Add("body", "empty")
	}

	if err.Ok() {
		self.events.PublishEvent(
			ess.NewEvent("post.written").
				For(self).
				Add("title", title).
				Add("body", body),
		)
	}

	return err.Return()
}
