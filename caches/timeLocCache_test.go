package caches

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTimeLocation(t *testing.T) {
	TimeLocationCache = NewLoadingCache()
	assert.Equal(t, 0, len(TimeLocationCache.DumpForTest()))
	// failure case
	expr, err :=  GetTimeLocation("unknown")
	assert.Error(t, err)
	assert.Equal(t, "unknown time zone unknown", err.Error())
	assert.Nil(t, expr)
	assert.Equal(t, 0, len(TimeLocationCache.DumpForTest()))
	// success case
	expr, err = GetTimeLocation("America/New_York")
	assert.NoError(t, err)
	assert.NotNil(t, expr)
	assert.Equal(t, 1, len(TimeLocationCache.DumpForTest()))
	// repeat success case shouldn't case any cache growth
	expr, err = GetTimeLocation("America/New_York")
	assert.NoError(t, err)
	assert.NotNil(t, expr)
	assert.Equal(t, 1, len(TimeLocationCache.DumpForTest()))
}
