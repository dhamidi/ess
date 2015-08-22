package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/dhamidi/ess"
)

var (
	WritePost = ess.NewCommandDefinition("write-post").
			Field("id", ess.Id()).
			Field("title", ess.TrimmedString()).
			Field("body", ess.TrimmedString()).
			Target(PostFromCommand)

	EditPost = ess.NewCommandDefinition("edit-post").
			Field("id", ess.Id()).
			Field("title", ess.TrimmedString()).
			Field("body", ess.TrimmedString()).
			Field("reason", ess.TrimmedString()).
			Target(PostFromCommand)
)

func PostFromCommand(params *ess.Command) ess.Aggregate {
	return NewPost(params.AggregateId())
}

type Post struct {
	events  ess.EventPublisher
	id      string
	written bool

	previousChecksum, checksum []byte
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

type PostsResource struct {
	app *ess.Application
}

func (self *PostsResource) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	result := (*ess.CommandResult)(nil)
	switch req.Method {
	case "POST":
		result = self.app.Send(WritePost.FromForm(req))
	}

	ShowResult(w, result)
}

func ShowResult(w http.ResponseWriter, result *ess.CommandResult) {
	if result == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	if err := result.Error(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err)
	} else {
		fmt.Fprintf(w, "{\"status\":\"ok\"}\n")
	}
}

func Show(w http.ResponseWriter, thing interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(thing)
}

type PostResource struct {
	app      *ess.Application
	allPosts *AllPostsInMemory
}

func (self *PostResource) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	result := (*ess.CommandResult)(nil)
	subpath := strings.TrimPrefix(req.URL.Path, "/posts/")
	fields := strings.Split(subpath, "/")
	postId := fields[0]
	action := ""
	if len(fields) > 1 {
		action = fields[1]
	}

	switch action {
	case "":
		post, err := self.allPosts.ById(postId)
		if err != nil {
			result = ess.NewErrorResult(err)
		} else {
			Show(w, post)
			return
		}
	case "edit":
		params := EditPost.FromForm(req)
		params.Fields["id"] = ess.StringValue(postId)
		result = self.app.Send(params)
	}

	ShowResult(w, result)
}

func main() {
	logger := log.New(os.Stderr, "blog ", 0)
	store, err := ess.NewEventsOnDisk("events.json", ess.SystemClock)
	if err != nil {
		logger.Fatal(err)
	}

	allPostsInMemory := NewAllPostsInMemory()
	application := ess.NewApplication("blog").
		WithLogger(logger).
		WithStore(store).
		WithProjection("all-posts", allPostsInMemory)

	if err := application.Init(); err != nil {
		logger.Fatal(err)
	}

	http.Handle("/posts/", &PostResource{app: application, allPosts: allPostsInMemory})
	http.Handle("/posts", &PostsResource{app: application})

	logger.Fatal(http.ListenAndServe(args(args(os.Args[1:]...), "localhost:6060"), nil))
}

func args(argv ...string) string {
	for _, arg := range argv {
		if arg != "" {
			return arg
		}
	}

	return ""
}
