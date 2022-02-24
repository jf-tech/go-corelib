package times

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOSClock(t *testing.T) {
	c := NewOSClock()
	cnow := c.Now()
	osnow := time.Now()
	assert.True(t, cnow.Before(osnow) || cnow.Equal(osnow))
}
