package times

import "time"

type Clock interface {
	Now() time.Time
}

type osClock struct{}

func (*osClock) Now() time.Time {
	return time.Now()
}

func NewOSClock() *osClock {
	return &osClock{}
}
