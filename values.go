package ess

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type String struct {
	original  string
	sanitized string
	sanitizer func(string) string
}

func TrimmedString() *String {
	return &String{
		sanitizer: strings.TrimSpace,
	}
}

func StringValue(str string) *String {
	return &String{
		sanitized: str,
		sanitizer: func(s string) string { return s },
	}
}

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
	identifierRegexp       = regexp.MustCompile(`^[-a-z0-9]+$`)
	ErrMalformedIdentifier = errors.New(`malformed_identifier`)
)

type Identifier struct {
	id string
}

func Id() *Identifier {
	return &Identifier{}
}

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

func EmailAddress() *Email { return &Email{} }

type BcryptedPassword struct {
	plain []byte
	bytes []byte
}

func (self *BcryptedPassword) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		return errors.New("empty")
	}
	bytes, err := bcrypt.GenerateFromPassword(data, bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	self.plain = append(self.plain, data...)
	self.bytes = bytes
	return nil
}

func (self *BcryptedPassword) Copy() Value    { return &BcryptedPassword{bytes: self.bytes} }
func (self *BcryptedPassword) String() string { return string(self.bytes) }

func (self *BcryptedPassword) Matches(hashedPassword string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), self.plain) == nil
}

func Password() *BcryptedPassword { return &BcryptedPassword{} }
