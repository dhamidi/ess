package ess

import (
	"bytes"
	"fmt"
)

// ValidationError captures errors about the values of a command's
// parameter or the state of a whole aggregate.
//
// Return errors of this type by the methods handling commands in your
// aggregates.
type ValidationError struct {
	Errors map[string][]string `json:"error"`
}

// NewValidationError returns a new, empty validation error.
func NewValidationError() *ValidationError {
	return &ValidationError{
		Errors: map[string][]string{},
	}
}

// Ok returns true if no errors have been recorded with this instance.
func (self *ValidationError) Ok() bool { return len(self.Errors) == 0 }

// Add records an error for field using desc as the error description.
func (self *ValidationError) Add(field string, desc string) *ValidationError {
	self.Errors[field] = append(self.Errors[field], desc)
	return self
}

// Merge records errors from err into this instance.
//
// If err is a ValidationError, all recorded errors for all fields
// from err are merged into this instance.
//
// Otherwise err's string representation is recorded in the field
// $all.
func (self *ValidationError) Merge(err error) *ValidationError {
	verr, ok := err.(*ValidationError)
	if !ok {
		return self.Add("$all", err.Error())
	}

	for field, errors := range verr.Errors {
		self.Errors[field] = append(self.Errors[field], errors...)
	}

	return self
}

// Return returns nil if no errors have been recorded with this
// instance.  Otherwise this instance is returned.
//
// This method exists to avoid returning a typed nil accidentally.
//
// Example:
//
// 	func (obj *MyDomainObject) DoSomething(param string) error {
// 		err := NewValidationError()
// 		if param == "" {
//			err.Add("param", "empty")
//		}
//		return err.Return()
//	}
func (self *ValidationError) Return() error {
	if len(self.Errors) == 0 {
		return nil
	} else {
		return self
	}
}

// Error implements the error interface.
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
