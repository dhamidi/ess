package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/dhamidi/ess"
)

var (
	WritePost = ess.NewCommandDefinition("write-post").
			Field("id", ess.Id()).
			Field("title", ess.TrimmedString()).
			Field("body", ess.TrimmedString()).
			Field("username", ess.Id()).
			Target(PostFromCommand)

	EditPost = ess.NewCommandDefinition("edit-post").
			Field("id", ess.Id()).
			Field("title", ess.TrimmedString()).
			Field("body", ess.TrimmedString()).
			Field("reason", ess.TrimmedString()).
			Field("username", ess.Id()).
			Target(PostFromCommand)

	SignUp = ess.NewCommandDefinition("sign-up").
		Id("username", ess.Id()).
		Field("email", ess.EmailAddress()).
		Field("password", ess.Password()).
		Target(UserFromCommand)

	LogIn = ess.NewCommandDefinition("login").
		Id("username", ess.Id()).
		Field("password", ess.Password()).
		Field("session", ess.TrimmedString()).
		Target(UserFromCommand)

	LogOut = ess.NewCommandDefinition("logout").
		Id("username", ess.Id()).
		Field("session", ess.TrimmedString()).
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
	case "GET":
		ShowSignupForm(w)
	case "POST":
		params := SignUp.FromForm(req)
		result = self.app.Send(params)
		if err := result.Error(); err != nil {
			ShowSignupFormErrors(w, params, err)
		} else {
			http.Redirect(w, req, "/sessions", http.StatusSeeOther)
		}
	default:
		MethodNotSupported(w)
	}
}

type SessionsResource struct {
	app         *ess.Application
	allSessions SessionStore
}

func (self *SessionsResource) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	action := strings.TrimPrefix(req.URL.Path, "/sessions/")

	switch req.Method {
	case "GET":
		ShowLoginForm(w, req)
	case "POST":
		req.ParseForm()
		if action == "logout" {
			self.Logout(w, req)
		} else {
			self.Login(w, req)
		}

	default:
		MethodNotSupported(w)
	}

}

func (self *SessionsResource) Logout(w http.ResponseWriter, req *http.Request) {
	currentUser := loadCurrentUser(req, self.allSessions)
	if currentUser != nil {
		req.Form["session"] = []string{currentUser.SessionId}
		req.Form["username"] = []string{currentUser.Username}
		params := LogOut.FromForm(req)
		self.app.Send(params)
	}
	http.Redirect(w, req, "/", http.StatusSeeOther)
}

func (self *SessionsResource) Login(w http.ResponseWriter, req *http.Request) {
	sessionId := GenerateSessionId()
	req.Form["session"] = []string{sessionId}
	params := LogIn.FromForm(req)
	result := self.app.Send(params)
	if err := result.Error(); err != nil {
		ShowLoginFormError(w, params, err)
	} else {
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    sessionId,
			Expires:  time.Now().Add(24 * time.Hour),
			Path:     "/",
			Domain:   req.URL.Host,
			HttpOnly: true,
		})
		returnTo := "/"
		if returnPath := req.FormValue("return"); returnPath != "" {
			returnTo = returnPath
		}
		http.Redirect(w, req, returnTo, http.StatusSeeOther)
	}
}

type PostsResource struct {
	app         *ess.Application
	allSessions *AllSessionsInMemory
}

func (self *PostsResource) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	currentUser := loadCurrentUser(req, self.allSessions)
	if currentUser == nil {
		RequireLogin(w, req)
		return
	}

	switch req.Method {
	case "GET":
		ShowPostForm(w)
	case "POST":
		req.ParseForm()
		req.Form["username"] = []string{currentUser.Username}
		params := WritePost.FromForm(req)
		result := self.app.Send(params)
		if err := result.Error(); err != nil {
			ShowPostFormError(w, params, err)
		} else {
			http.Redirect(w, req, "/posts/"+params.Get("id").String(), http.StatusSeeOther)
		}
	}
}

func MethodNotSupported(w http.ResponseWriter) {
	http.Error(w, "Method Not Supported", http.StatusMethodNotAllowed)
}

func NotFound(w http.ResponseWriter) {
	http.Error(w, "404 Not Found", http.StatusNotFound)
}

func RequireLogin(w http.ResponseWriter, req *http.Request) {

	returnTo := &url.URL{
		Path: "/sessions",
		RawQuery: url.Values{
			"return": []string{req.URL.Path},
		}.Encode(),
	}
	http.Redirect(w, req, returnTo.String(), http.StatusSeeOther)
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
	app         *ess.Application
	allPosts    *AllPostsInMemory
	allSessions SessionStore
}

func (self *PostResource) ServeHTTP(w http.ResponseWriter, req *http.Request) {
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
			NotFound(w)
		} else {
			ShowPost(w, post)
		}
	case "edit":
		self.handleEdits(w, req, postId)
	}
}

func (self *PostResource) handleEdits(w http.ResponseWriter, req *http.Request, postId string) {
	currentUser := loadCurrentUser(req, self.allSessions)
	if currentUser == nil {
		RequireLogin(w, req)
		return
	}

	post, err := self.allPosts.ById(postId)
	if err != nil {
		NotFound(w)
		return
	}

	switch req.Method {
	case "GET":
		params := EditPost.NewCommand().
			Set("username", currentUser.Username).
			Set("title", post.Title).
			Set("body", post.Body).
			Set("id", postId)

		ShowEditPostForm(w, params)
	case "POST":
		params := EditPost.FromForm(req).Set("id", postId)
		result := self.app.Send(params)
		if err := result.Error(); err != nil {
			ShowEditPostFormError(w, params, err)
		} else {
			http.Redirect(w, req, post.Path, http.StatusSeeOther)
		}
	default:
		MethodNotSupported(w)
	}
}

type IndexResource struct {
	app         *ess.Application
	allPosts    *AllPostsInMemory
	allSessions *AllSessionsInMemory
}

func (self *IndexResource) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		currentUser := loadCurrentUser(req, self.allSessions)
		allPosts, _ := self.allPosts.Recent()
		ShowAllPostsIndex(w, currentUser, allPosts)
	default:
		MethodNotSupported(w)
	}
}

type SessionStore interface {
	ById(id string) (*ProjectedUser, error)
}

func loadCurrentUser(req *http.Request, sessions SessionStore) *ProjectedUser {
	sessionCookie, err := req.Cookie("session")
	if err != nil {
		return nil
	}

	user, err := sessions.ById(sessionCookie.Value)
	if err != nil {
		return nil
	}
	return user
}

func main() {
	logger := log.New(os.Stderr, "blog ", 0)
	store, err := ess.NewEventsOnDisk("events.json", ess.SystemClock)
	if err != nil {
		logger.Fatal(err)
	}

	allPostsInMemory := NewAllPostsInMemory()
	allSessionsInMemory := NewAllSessionsInMemory()
	application := ess.NewApplication("blog").
		WithLogger(logger).
		WithStore(store).
		WithProjection("all-posts", allPostsInMemory).
		WithProjection("all-sessions", allSessionsInMemory)

	if err := application.Init(); err != nil {
		logger.Fatal(err)
	}

	http.Handle("/sessions", &SessionsResource{app: application, allSessions: allSessionsInMemory})
	http.Handle("/sessions/", &SessionsResource{app: application, allSessions: allSessionsInMemory})
	http.Handle("/signups", &SignupsResource{app: application})
	http.Handle("/posts/", &PostResource{app: application, allPosts: allPostsInMemory, allSessions: allSessionsInMemory})
	http.Handle("/posts", &PostsResource{app: application, allSessions: allSessionsInMemory})
	http.Handle("/", &IndexResource{app: application, allPosts: allPostsInMemory, allSessions: allSessionsInMemory})

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
