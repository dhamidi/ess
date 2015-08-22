package ess

import (
	"bytes"
	"fmt"
)

type CommandResult struct {
	aggregateId string
	err         error
}

func (self *CommandResult) Error() error {
	return self.err
}

func (self *CommandResult) AggregateId() string {
	return self.aggregateId
}

func NewErrorResult(err error) *CommandResult {
	return &CommandResult{
		err: err,
	}
}

func NewSuccessResult(receiver Aggregate) *CommandResult {
	return &CommandResult{
		aggregateId: receiver.Id(),
	}
}

type CommandDefinition struct {
	Name       string
	Fields     map[string]Value
	TargetFunc func(*Command) Aggregate
	IdField    string
}

func NewCommandDefinition(name string) *CommandDefinition {
	return &CommandDefinition{
		Name:    name,
		Fields:  map[string]Value{},
		IdField: "id",
	}
}

func (self *CommandDefinition) Id(name string, value Value) *CommandDefinition {
	self.IdField = name
	return self.Field(name, value)
}

func (self *CommandDefinition) Field(name string, value Value) *CommandDefinition {
	self.Fields[name] = value
	return self
}

func (self *CommandDefinition) Target(constructor func(*Command) Aggregate) *CommandDefinition {
	self.TargetFunc = constructor
	return self
}

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

func (self *CommandDefinition) FromForm(form Form) *Command {
	command := self.NewCommand()
	return command.FromForm(form)
}

type Command struct {
	Name    string
	Fields  map[string]Value
	IdField string

	errors       *ValidationError
	receiver     Aggregate
	receiverFunc func(*Command) Aggregate
}

func (self *Command) AggregateId() string {
	val := self.Get(self.IdField)
	if val != nil {
		return val.String()
	} else {
		return ""
	}
}

func (self *Command) err(field string, err error) {
	self.errors.Add(field, err.Error())
}

func (self *Command) Get(name string) Value {
	return self.Fields[name]
}

func (self *Command) Receiver() Aggregate {
	if self.receiver == nil {
		self.receiver = self.receiverFunc(self)
	}

	return self.receiver
}

func (self *Command) FromForm(form Form) *Command {
	for field, value := range self.Fields {
		text := form.FormValue(field)
		if err := value.UnmarshalText([]byte(text)); err != nil {
			self.err(field, err)
		}
	}

	return self
}

func (self *Command) Acknowledge(clock Clock) {
	now := clock.Now()
	self.Fields["now"] = &Time{now}
	if self.AggregateId() == "" {
		self.Fields[self.IdField] = StringValue(fmt.Sprintf("%d", now.UnixNano()))
	}
}

func (self *Command) Execute() error {
	if !self.errors.Ok() {
		return self.errors.Return()
	}
	return self.receiver.HandleCommand(self)
}

func (self *Command) String() string {
	out := bytes.NewBufferString(self.Name + "\n")

	for field, value := range self.Fields {
		fmt.Fprintf(out, "param %s: ", field)
		fmt.Fprintf(out, "%q", value)
		fmt.Fprintf(out, "\n")
	}

	return out.String()
}
