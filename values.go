package ess

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// String is an implementation of Value for handling string
// parameters.
type String struct {
	original  string
	sanitized string
	sanitizer func(string) string
}

// TrimmedString constructs a string value which removes initial and
// trailing whitespace from the initial string.
func TrimmedString() *String {
	return &String{
		sanitizer: strings.TrimSpace,
	}
}

// StringValue constructs a string which returns str when calling its
// String method.
func StringValue(str string) *String {
	return &String{
		sanitized: str,
		sanitizer: func(s string) string { return s },
	}
}

// UnmarshalText accepts data as the string's content and applies and
// internal sanitization function to data.
func (self *String) UnmarshalText(data []byte) error {
	self.original = string(data)
	self.sanitized = self.sanitizer(self.original)
	return nil
}

func (self *String) String() string {
	return self.sanitized
}

func (self *String) Copy() Value {
	return &String{
		sanitized: self.sanitized,
		original:  self.original,
		sanitizer: self.sanitizer,
	}
}

// Time is an implementation of Value for handling timestamps.  It
// works with timestamps formatted according to time.RFC3339Nano.
type Time struct {
	time.Time
}

func (self Time) String() string {
	data, _ := self.Time.MarshalText()
	return string(data)
}

func (self Time) Copy() Value {
	return &Time{self.Time}
}

var (
	identifierRegexp = regexp.MustCompile(`^[-a-z0-9]+$`)

	// ErrMalformedIdentifier is returned when parsing an
	// identifier fails.
	ErrMalformedIdentifier = errors.New(`malformed_identifier`)

	// ErrEmpty is returned when a non-empty input string is
	// expected.
	ErrEmpty = errors.New("empty")
)

// Identifier is a value for handling parameters that serve as
// identifiers.  It accepts any string consisting only of dashes,
// lowercase letters and digits.
//
// The empty string is not a valid identifier.
type Identifier struct {
	id string
}

// Id returns a new empty identifier.
func Id() *Identifier {
	return &Identifier{}
}

// UnmarshalText returns ErrMalformedIdentifier identifier is data is
// not a valid identifier.
func (self *Identifier) UnmarshalText(data []byte) error {
	id := strings.TrimSpace(string(data))
	if !identifierRegexp.MatchString(id) {
		return ErrMalformedIdentifier
	}

	self.id = id
	return nil
}

func (self *Identifier) String() string {
	return self.id
}

func (self *Identifier) Copy() Value {
	return &Identifier{id: self.id}
}

// Email is an implementation of value for handling email addresses.
// It parses email addresses according to RFC 5322, e.g. "Barry Gibbs
// <bg@example.com>".
type Email struct {
	address *mail.Address
}

func (self *Email) UnmarshalText(data []byte) error {
	address, err := mail.ParseAddress(string(data))
	if err != nil {
		return err
	}

	self.address = address
	return nil
}

func (self *Email) String() string {
	if self.address != nil {
		return self.address.Address
	}

	return ""
}

func (self *Email) Copy() Value {
	return &Email{address: self.address}
}

// EmailAddress returns a new, empty email value.
func EmailAddress() *Email { return &Email{} }

// BcryptedPassword is an implementation for securely handling
// password parameters.  It uses the bcrypt algorithm for hashing
// passwords.
type BcryptedPassword struct {
	plain []byte
	bytes []byte
}

// UnmarshalText generates a password from data using bcrypt.  It
// returns ErrEmpty is data is empty.
func (self *BcryptedPassword) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		return ErrEmpty
	}
	bytes, err := bcrypt.GenerateFromPassword(data, bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	self.plain = append(self.plain, data...)
	self.bytes = bytes
	return nil
}

// Copy copies the password.  The copy does not contain the password's
// plain text anymore.
func (self *BcryptedPassword) Copy() Value { return &BcryptedPassword{bytes: self.bytes} }

// String returns the hashed password as a string.
func (self *BcryptedPassword) String() string { return string(self.bytes) }

// Matches returns true if this password matches hashedPassword.
func (self *BcryptedPassword) Matches(hashedPassword string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), self.plain) == nil
}

// Password returns a new, empty BcryptedPassword.
func Password() *BcryptedPassword { return &BcryptedPassword{} }
