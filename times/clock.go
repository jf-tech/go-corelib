package times

import "time"

// Clock tells the current time.
type Clock interface {
	Now() time.Time
}

type osClock struct{}

func (*osClock) Now() time.Time {
	return time.Now()
}

// NewOSClock returns a Clock interface implementation that uses time.Now.
func NewOSClock() *osClock {
	return &osClock{}
}
