package main

import (
	"testing"

	"github.com/dhamidi/ess"
)

func TestUser_SignUp_failsIfUserHasSignedUpAlready(t *testing.T) {
	password := ess.Password()
	password.UnmarshalText([]byte("password"))

	history := []*ess.Event{
		ess.NewEvent("user.signed-up").
			Add("username", "test-user").
			Add("password", password.String()).
			Add("email", "test@example.com").
			Add("name", "John Doe"),
	}

	user := NewUser("username")
	for _, event := range history {
		user.HandleEvent(event)
	}

	err := user.SignUp("Jane Doe", "jane.doe@example.com", password.String())
	if err == nil {
		t.Fatal("Expected an error")
	}

	verr, ok := err.(*ess.ValidationError)
	if !ok {
		t.Fatalf("err.(type) = %T; want %T", err, verr)
	}

	if got, want := len(verr.Errors["username"]), 1; got != want {
		t.Errorf(`len(verr.Errors["username"]) = %v; want %v`, got, want)
	}

	if got, want := verr.Errors["username"][0], "not_unique"; got != want {
		t.Errorf(`verr.Errors["username"][0] = %v; want %v`, got, want)
	}
}
