package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/dhamidi/ess"
)

var SignUp = ess.NewCommandDefinition("sign-up").
	Id("username", ess.Id()).
	Field("name", ess.TrimmedString()).
	Field("email", ess.EmailAddress()).
	Field("password", ess.Password()).
	Target(NewUserFromCommand)

func NewUserFromCommand(command *ess.Command) ess.Aggregate {
	return NewUser(command.Get("username").String())
}

type User struct {
	id       string
	events   ess.EventPublisher
	signedUp bool
}

func NewUser(username string) *User {
	return &User{
		id: username,
	}
}

func (self *User) PublishWith(publisher ess.EventPublisher) ess.Aggregate {
	self.events = publisher
	return self
}

func (self *User) Id() string {
	return self.id
}

func (self *User) HandleCommand(command *ess.Command) error {
	switch command.Name {
	case "sign-up":
		return self.SignUp(
			command.Get("name").String(),
			command.Get("email").String(),
			command.Get("password").String(),
		)
	}

	return nil
}

func (self *User) SignUp(name, email, password string) error {
	err := ess.NewValidationError()

	if self.signedUp {
		err.Add("username", "not_unique")
	}

	if password == "" {
		err.Add("password", "empty")
	}

	if email == "" {
		err.Add("email", "empty")
	}

	if err.Ok() {
		self.events.PublishEvent(
			ess.NewEvent("user.signed-up").
				For(self).
				Add("password", password).
				Add("email", email).
				Add("name", name),
		)
	}

	return err.Return()
}

func (self *User) HandleEvent(event *ess.Event) {
	switch event.Name {
	case "user.signed-up":
		self.signedUp = true
	}
}

func main() {
	app := ess.NewApplication("user-signup-example")

	http.HandleFunc("/", ShowSignupForm)
	http.Handle("/signups", HandleSignup(app))
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}

func ShowSignupForm(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<html><head><title>Signup</title></head>
	<body>
	<h1>Sign up</h1>
	<form id="signup" action="/signups" method="POST">
	  <p>
	    <label>Username:</label>
	    <input type="text" required name="username">
	  </p>
	  <p>
	    <label>Your name:</label>
	    <input type="text" required name="name">
	  </p>
	  <p>
	    <label>Your email:</label>
	    <input type="email" required name="email">
	  </p>
	  <p>
	    <label>Password:</label>
	    <input type="password" required name="password">
	  </p>
	  <p>
	    <button type="submit">Sign up</button>
	  </p>
	</form>
	`)
}

func HandleSignup(app *ess.Application) http.Handler {
	handler := func(w http.ResponseWriter, req *http.Request) {
		command := SignUp.FromForm(req)
		result := app.Send(command)

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if err := result.Error(); err != nil {
			fmt.Fprintf(w, "Errors: %s\n", err)
		} else {
			fmt.Fprintf(w, "Signed up successfully.\n")
		}
	}
	return http.HandlerFunc(handler)
}
