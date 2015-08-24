package ess

import (
	"bytes"
	"fmt"
)

// CommandResult represents the result of the application handling a
// command.
type CommandResult struct {
	aggregateId string
	err         error
}

// Error returns any error encountered or caused by processing the
// command.  If nil is returned, it is safe to assume that the
// application processed the command successfully.
func (self *CommandResult) Error() error {
	return self.err
}

// AggregateId returns the id of the command's receiver.
func (self *CommandResult) AggregateId() string {
	return self.aggregateId
}

// NewErrorResult wraps err in a CommandResult.
func NewErrorResult(err error) *CommandResult {
	return &CommandResult{
		err: err,
	}
}

// NewSuccessResult returns a CommandResult that marks a success for
// receiver.
func NewSuccessResult(receiver Aggregate) *CommandResult {
	return &CommandResult{
		aggregateId: receiver.Id(),
	}
}

// CommandDefinition is used for defining the commands accepted by the
// application.  Essentially it is a dynamically built definition of
// messages the system accepts.
//
// New command instances are created from an existing command
// definition.
//
// The following is a complete example of defining a command your
// application accepts:
//
//	SignUp = ess.NewCommandDefinition("sign-up").
// 		Id("username", ess.Id()).
// 		Field("email", ess.EmailAddress()).
// 		Field("password", ess.Password()).
// 		Target(UserFromCommand)
//
//	func UserFromCommand(command *Command) Aggregate {
//		return NewUser(command.Get("username").String())
//	}
//
// New command instances can then be created from this definition:
//
//     signUp := SignUp.NewCommand().Set("username", "admin") /* ... */
//
// A convenience method is provided to set parameters based on a HTTP
// request object:
//
//     signUp := SignUp.FromForm(req)
type CommandDefinition struct {
	Name   string           // name of the command, e.g. "sign-up"
	Fields map[string]Value // map of parameter name to accepted type

	// TargetFunc constructs a new instance of the receiver of the
	// command.  This callback exists to avoid the use of
	// reflection.
	TargetFunc func(*Command) Aggregate

	// IdField is the name of the parameter which identifies the
	// command receiver, defaults to "id"
	IdField string
}

// NewCommandDefinition creates a new command definition using name as
// the name for the command.
func NewCommandDefinition(name string) *CommandDefinition {
	return &CommandDefinition{
		Name:    name,
		Fields:  map[string]Value{},
		IdField: "id",
	}
}

// Id sets the name and type of the field that is considered identify
// the command's receiver.
//
// The default is to use a field named "id" of type "Identifier".
func (self *CommandDefinition) Id(name string, value Value) *CommandDefinition {
	self.IdField = name
	return self.Field(name, value)
}

// Field defines a field with the given name and type.  Use this
// method to define the different parameters of a command.
func (self *CommandDefinition) Field(name string, value Value) *CommandDefinition {
	self.Fields[name] = value
	return self
}

// Target sets the function to create a new receiver of the right type
// for this command to constructor.
//
// The constructor's task is to return an aggregate instance with the
// appropriate ID extracted from the command passed to the
// constructor.
//
// Example:
//
// 	func UserFromCommand(command *Command) Aggregate {
// 		return NewUser(command.Get("username").String())
// 	}
func (self *CommandDefinition) Target(constructor func(*Command) Aggregate) *CommandDefinition {
	self.TargetFunc = constructor
	return self
}

// NewCommand constructs a new instance of a command, according to
// this command definition.
func (self *CommandDefinition) NewCommand() *Command {
	cmd := &Command{
		Name: self.Name,
		Fields: map[string]Value{
			self.IdField: Id(),
		},
		IdField:      self.IdField,
		errors:       NewValidationError(),
		receiverFunc: self.TargetFunc,
	}

	for field, val := range self.Fields {
		cmd.Fields[field] = val.Copy()
	}

	return cmd
}

// FromForm is a convenience method to create a new command instance
// and populate immediately from form.
func (self *CommandDefinition) FromForm(form Form) *Command {
	command := self.NewCommand()
	return command.FromForm(form)
}

// Command represents a message sent to your application with the
// intention to change application state.
//
// Commands are named in the imperative, e.g. "sign-up" or
// "place-order".  A command is targetted at a single receiver, the so
// called aggregate.
//
// The fields of a command are of type Value to provide a uniform
// interface for sanitizing inputs.
type Command struct {
	Name    string
	Fields  map[string]Value
	IdField string

	errors       *ValidationError
	receiver     Aggregate
	receiverFunc func(*Command) Aggregate
}

// AggregateId returns the id of the command's receiver, according to
// the command's IdField.  If the field is not present, it returns the
// empty string.
func (self *Command) AggregateId() string {
	val := self.Get(self.IdField)
	if val != nil {
		return val.String()
	} else {
		return ""
	}
}

// err adds an error to the list of errors for field
func (self *Command) err(field string, err error) {
	self.errors.Add(field, err.Error())
}

// Get returns the field identified by name or nil if the field does
// not exist.
func (self *Command) Get(name string) Value {
	return self.Fields[name]
}

// Receiver returns an instance of the command's receiver, possibly
// creating the instance.
func (self *Command) Receiver() Aggregate {
	if self.receiver == nil {
		self.receiver = self.receiverFunc(self)
	}

	return self.receiver
}

// Set sets the value for the field identified by name.  Setting a
// value using this method parses the string given in value according
// to the field's type and remembers any errors encountered.
//
// Use this method to "fill in" the parameters of a command.
func (self *Command) Set(name string, value string) *Command {
	target, found := self.Fields[name]
	if found {
		err := target.UnmarshalText([]byte(value))
		if err != nil {
			self.err(name, err)
		}
	}

	return self
}

// FromForm sets all of the command's fields with the values found in
// form.
func (self *Command) FromForm(form Form) *Command {
	for field, value := range self.Fields {
		text := form.FormValue(field)
		if err := value.UnmarshalText([]byte(text)); err != nil {
			self.err(field, err)
		}
	}

	return self
}

// Acknowledge marks the command as having been received by the
// system.
//
// Calling Acknowledge on a command sets the field "now" to the time
// provided by clock.
//
// This is useful for recording the time of actions in published
// events.
func (self *Command) Acknowledge(clock Clock) {
	now := clock.Now()
	self.Fields["now"] = &Time{now}
}

// Execute passes this command to its receiver, merging any errors
// returned into the errors encountered during parameter processing.
func (self *Command) Execute() error {
	err := self.receiver.HandleCommand(self)

	if !self.errors.Ok() {
		return self.errors.Merge(err).Return()
	}

	return err
}

// String returns a multiline representation of the command.
//
// The information contained in the returned string is enough to
// reconstruct the command.
func (self *Command) String() string {
	out := bytes.NewBufferString(self.Name + "\n")

	for field, value := range self.Fields {
		fmt.Fprintf(out, "param %s: ", field)
		fmt.Fprintf(out, "%q", value)
		fmt.Fprintf(out, "\n")
	}

	return out.String()
}
