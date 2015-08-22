package ess

import "time"

var (
	SystemClock = &SystemTime{}
)

type SystemTime struct{}

func (self *SystemTime) Now() time.Time {
	return time.Now()
}
