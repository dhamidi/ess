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

	SignUp = ess.NewCommandDefinition("sign-up").
		Id("username", ess.Id()).
		Field("email", ess.EmailAddress()).
		Field("password", ess.Password()).
		Target(UserFromCommand)

	LogIn = ess.NewCommandDefinition("login").
		Id("username", ess.Id()).
		Field("password", ess.Password()).
		Target(UserFromCommand)
)

func PostFromCommand(params *ess.Command) ess.Aggregate {
	return NewPost(params.AggregateId())
}

func UserFromCommand(params *ess.Command) ess.Aggregate {
	return NewUser(params.Get("username").String())
}

type SignupsResource struct {
	app *ess.Application
}

func (self *SignupsResource) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	result := (*ess.CommandResult)(nil)
	switch req.Method {
	case "POST":
		result = self.app.Send(SignUp.FromForm(req))
	default:
		MethodNotSupported(w)
		return
	}

	ShowResult(w, result)
}

type SessionsResource struct {
	app *ess.Application
}

func (self *SessionsResource) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	result := (*ess.CommandResult)(nil)
	switch req.Method {
	case "POST":
		result = self.app.Send(LogIn.FromForm(req))
	default:
		MethodNotSupported(w)
		return
	}

	ShowResult(w, result)
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

func MethodNotSupported(w http.ResponseWriter) {
	http.Error(w, "Method Not Supported", http.StatusMethodNotAllowed)
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

	http.Handle("/sessions", &SessionsResource{app: application})
	http.Handle("/signups", &SignupsResource{app: application})
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
