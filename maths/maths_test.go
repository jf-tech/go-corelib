package maths

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMinMaxInt(t *testing.T) {
	tests := []struct {
		name        string
		x           int
		y           int
		expectedMin int
		expectedMax int
	}{
		{
			name:        "x less than y",
			x:           1,
			y:           2,
			expectedMin: 1,
			expectedMax: 2,
		},
		{
			name:        "x greater than y",
			x:           2,
			y:           1,
			expectedMin: 1,
			expectedMax: 2,
		},
		{
			name:        "x equal to y",
			x:           2,
			y:           2,
			expectedMin: 2,
			expectedMax: 2,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedMin, MinInt(test.x, test.y))
			assert.Equal(t, test.expectedMax, MaxInt(test.x, test.y))
		})
	}
}

func TestAbsIntAndInt64(t *testing.T) {
	tests := []struct {
		name     string
		in       int
		expected int
	}{
		{
			name:     "in > 0",
			in:       1,
			expected: 1,
		},
		{
			name:     "in == 0",
			in:       0,
			expected: 0,
		},
		{
			name:     "in < 0",
			in:       -4,
			expected: 4,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, AbsInt(test.in))
			assert.Equal(t, int64(test.expected), AbsI64(int64(test.in)))
		})
	}
}

func TestMinMaxI64(t *testing.T) {
	tests := []struct {
		name        string
		x           int64
		y           int64
		expectedMin int64
		expectedMax int64
	}{
		{
			name:        "x less than y",
			x:           1,
			y:           2,
			expectedMin: 1,
			expectedMax: 2,
		},
		{
			name:        "x greater than y",
			x:           2,
			y:           1,
			expectedMin: 1,
			expectedMax: 2,
		},
		{
			name:        "x equal to y",
			x:           2,
			y:           2,
			expectedMin: 2,
			expectedMax: 2,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedMin, MinI64(test.x, test.y))
			assert.Equal(t, test.expectedMax, MaxI64(test.x, test.y))
		})
	}
}
