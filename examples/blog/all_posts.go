package main

import (
	"errors"
	"time"

	"github.com/dhamidi/ess"
)

var (
	ErrNotFound = errors.New("not_found")
)

type ProjectedPost struct {
	Id        string    `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	WrittenAt time.Time `json:"writtenAt"`
}

func NewProjectedPostFromEvent(event *ess.Event) *ProjectedPost {
	post := &ProjectedPost{Id: event.StreamId}
	post.Update(event)
	return post
}

func (self *ProjectedPost) Update(event *ess.Event) *ProjectedPost {
	self.Title = event.Payload["title"].(string)
	self.Body = event.Payload["body"].(string)

	if event.Name == "post.written" {
		self.WrittenAt = event.OccurredOn
	}

	return self
}

type AllPostsInMemory struct {
	byId map[string]*ProjectedPost
}

func NewAllPostsInMemory() *AllPostsInMemory {
	return &AllPostsInMemory{
		byId: map[string]*ProjectedPost{},
	}
}

func (self *AllPostsInMemory) HandleEvent(event *ess.Event) {
	switch event.Name {
	case "post.edited":
		self.byId[event.StreamId].Update(event)
	case "post.written":
		post := NewProjectedPostFromEvent(event)
		self.byId[event.StreamId] = post
	}
}

func (self *AllPostsInMemory) ById(id string) (*ProjectedPost, error) {
	post, found := self.byId[id]
	if !found {
		return nil, ErrNotFound
	}
	return post, nil
}
