/*

This package provides tools for implementing an application using
event sourcing.  For a detailed explanation of event sourcing, see
Martin Fowler's article on the topic:
http://martinfowler.com/eaaDev/EventNarrative.html#EventSourcing.

Short summary of event sourcing

Event sourcing works by capturing every state change to the system in
the form of events and logging these events to usually persistent
storage.  The current state of the system can be computed by going
through the event history.

Queries are served from separate models targeted at specific use
cases (projections).  These models are updated by listening to events as they are
appended to the event history.  Multiple such models can exist next to
each other: a search index, a relational database, static files.

Event sourcing as implemented by this package

The following sketch gives an overview about the dataflow and
components in applications built with this package:


	Performing writes:
		Client --->
			Command --->
			Application --->
			Domain Object --->
			Event History ---> Projections

	Performing reads:
		Client <--- Projections


A client can be anything that is capable of sending commands to the
application instance over any transport mechanism.  In most cases the
client will be a web browser sending commands over HTTP in the form of
POST requests.  Other likely clients include a command line interface
for performing administrative tasks.

A command is a message sent to the application with the intent of
changing application state.  Commands map directly to actions intended
by the user, e.g. "sign-up" or "login".  Every command sent to the
application is targeted a specific domain object.  The "login" command
for example would be targeted at a specific user, the one who is
logging in.

The application processes a command by first identifying and
instantiating the command's receiver.  It then replays any historic
events for the receiver, in order for the receiver to reconstruct any
necessary internal state.  If the reconstructing the receiver's
current state caused no problem, the command is passed to the receiver
to handle.

The domain object is where your business logic lives.  Domain objects
act as command receivers and enforce business rules.  If the domain
object accepts a command, it emits events to make note of this fact.
Otherwise the command is rejected and an error is reported to the
client.

An event captures a fact with regards to the application state,
e.g. that a user logged in or signed up.  Events carry all the
necessary data for reconstructing state and building read models.  For
example, a "user signed up" event most likely contains the user's
name, a cryptographically hashed version of the password she provided
and her contact email address.

Events are persisted in an append-only log, the event history.  If
persisting the events succeeds, any interested parties are notified
about this.

Projections process events as they happen and use them to build some
sort of state.  This state can be stored anywhere and can take any
form, because the application does not depend on this projected state.
Projected state can be thrown away and rebuilt from history if
necessary, because all important information has been captured in the
event history already.

Tutorial

Building an application using event sourcing works best by working
from the outside in.  Start with the user story and think about which
information the user will have to enter to fulfill his goal.  This
example looks at building user sign up for your application.  We
imagine the form to look somewhat like this:


			Sign up for $COOL_PRODUCT

	Username:       [__________________]
	Your Name:	[__________________]
	Your Email:	[__________________]
	Password:	[__________________]

			[ Sign up ]

This translates directly to a command, "sign-up", capturing the data
of the form:

	var SignUp = ess.NewCommandDefinition("sign-up").
			Id("username", ess.Id()).
			Field("name", ess.TrimmedString()).
			Field("email", ess.EmailAddress()).
			Field("password", ess.Password()).
			Target(NewUserFromCommand)

The above definition mirrors the data captured in the form.  This
package already provides types for parsing email addresses and
handling password input parameters, so we use those to capture the
user's mail address and password.

In order to ensure uniqueness, we let the user choose a username for
our platform.  The chosen username will serve as the user's id.  We
could use the user's email addresses instead but then we'd get into
trouble once the user changes her contact email address.

The function NewUserFromCommand is responsible for creating a new user
object from a command matching above structure:

	func NewUserFromCommand(command *ess.Command) Aggregate {
		return NewUser(command.Get("username").String())
	}

The user object is responsible for ensuring our business rules about
users:

- the username needs to be unique

- an email address needs to be provided

- a password needs to be present

This is an empty implementation of our user object, providing all the
necessary methods for acting as an Aggregate.

	type User struct {
		id     string
		events ess.EventPublisher
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
		// ...
		return nil
	}

	func (self *User) HandleEvent(event *ess.Event) {
		// ...
	}


When a user submits the form, a new instance of this object will be
created by the NewUserFromCommand function and the object's
HandleCommand method will be called.  So let's add the signup logic
there:

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

	        // How to check for username uniqueness?

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

Checking for the presence of the password and email fields is pretty
straightforward, but how can we ensure that every username is unique?
We can arrive at an answer by looking at what "unique username" means:
it means that no user must have signed up with the same username
already. The phrase "signed up" is key here -- it points us to an
event.  In other words: trying to sign up a user who has already
signed up should fail.

Let's add a field to our user object and note the fact that the user
has signed up in the user's HandleEvent method:

	type User struct {
		// existing fields omitted
		signedUp bool
	}

	func (self *User) HandleEvent(event *ess.Event) {
		switch event.Name {
		case "user.signed-up":
			self.signedUp = true
		}
	}

Now we can check this field in the SignUp method and return an error:

	func (self *User) SignUp(name, email, password string) error {
	        err := ess.NewValidationError()

	        if self.signedUp {
			err.Add("username", "not_unique")
		}

		// ...
	}

That's it!  Adding new commands follows the same process.

Let's hook everything up to see the whole example in action:

	import (
		"fmt"
		"log"
		"net/http"

		"github.com/dhamidi/ess"
	)

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
*/
package ess
