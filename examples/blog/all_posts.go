package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dhamidi/ess"
)

var (
	ErrNotFound = errors.New("not_found")
)

type ProjectedPost struct {
	Id    string `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`

	Paragraphs []string `json:"paragraphs"`

	Path      string    `json:"path"`
	WrittenAt time.Time `json:"writtenAt"`

	Author string `json:"author"`
}

func NewProjectedPostFromEvent(event *ess.Event) *ProjectedPost {
	post := &ProjectedPost{Id: event.StreamId}
	post.Update(event)
	return post
}

func (self *ProjectedPost) Update(event *ess.Event) *ProjectedPost {
	self.Title = event.Payload["title"].(string)
	self.Body = event.Payload["body"].(string)
	self.Paragraphs = strings.Split(
		strings.NewReplacer("\r\n", "\n").Replace(self.Body),
		"\n",
	)
	self.Path = fmt.Sprintf("/posts/%s", self.Id)

	if author := event.Payload["author"]; author != nil {
		self.Author = author.(string)
	} else {
		self.Author = "anonymous"
	}

	if event.Name == "post.written" {
		self.WrittenAt = event.OccurredOn
	}

	return self
}

type AllPostsInMemory struct {
	byId   map[string]*ProjectedPost
	recent []*ProjectedPost
}

func NewAllPostsInMemory() *AllPostsInMemory {
	return &AllPostsInMemory{
		byId:   map[string]*ProjectedPost{},
		recent: []*ProjectedPost{},
	}
}

func (self *AllPostsInMemory) HandleEvent(event *ess.Event) {
	switch event.Name {
	case "post.edited":
		self.byId[event.StreamId].Update(event)
	case "post.written":
		post := NewProjectedPostFromEvent(event)
		self.byId[event.StreamId] = post
		self.recent = append([]*ProjectedPost{post}, self.recent...)
	}
}

func (self *AllPostsInMemory) ById(id string) (*ProjectedPost, error) {
	post, found := self.byId[id]
	if !found {
		return nil, ErrNotFound
	}
	return post, nil
}

func (self *AllPostsInMemory) Recent() ([]*ProjectedPost, error) {
	return self.recent, nil
}
