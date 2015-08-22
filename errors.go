package ess

import (
	"bytes"
	"fmt"
)

type ValidationError struct {
	Errors map[string][]string `json:"error"`
}

func NewValidationError() *ValidationError {
	return &ValidationError{
		Errors: map[string][]string{},
	}
}

func (self *ValidationError) Ok() bool { return len(self.Errors) == 0 }

func (self *ValidationError) Add(field string, desc string) *ValidationError {
	self.Errors[field] = append(self.Errors[field], desc)
	return self
}

func (self *ValidationError) Return() error {
	if len(self.Errors) == 0 {
		return nil
	} else {
		return self
	}
}

func (self *ValidationError) Error() string {
	out := new(bytes.Buffer)
	for field, errors := range self.Errors {
		fmt.Fprintf(out, "%s: ", field)
		for i, desc := range errors {
			fmt.Fprintf(out, "%s", desc)
			if i < len(errors)-1 {
				fmt.Fprintf(out, ", ")
			}
		}
		fmt.Fprintf(out, "; ")
	}
	return out.String()
}
