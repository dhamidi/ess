package ess

import "time"

var (
	// SystemClock provides access to time.Now through the Clock
	// interface.
	SystemClock = &SystemTime{}
)

// SystemTime wraps time.Now to implement the Clock interface.
type SystemTime struct{}

func (self *SystemTime) Now() time.Time {
	return time.Now()
}

// StaticClock implements the Clock interface by returning a static
// time.  Its intended use is in test cases.
type StaticClock struct {
	time.Time
}

func (self *StaticClock) Now() time.Time {
	return self.Time
}
