package maths

import "math"

// MaxInt returns the bigger value of the two input ints.
func MaxInt(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// MinInt returns the smaller value of the two input ints.
func MinInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// AbsInt returns the absolute value of an int value.
func AbsInt(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

// MaxIntValue is the max value for type int.
// https://groups.google.com/forum/#!msg/golang-nuts/a9PitPAHSSU/ziQw1-QHw3EJ
const MaxIntValue = int(^uint(0) >> 1)

// MaxI64 returns the bigger value of the two input int64s.
func MaxI64(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}

// MinI64 returns the smaller value of the two input int64s.
func MinI64(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

// AbsI64 returns the absolute value of an int64 value.
func AbsI64(a int64) int64 {
	if a < 0 {
		return -a
	}
	return a
}

// MaxI64Value is the max value for type int64.
const MaxI64Value = math.MaxInt64
