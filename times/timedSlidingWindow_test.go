package times

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testClock struct {
	now time.Duration
}

func (tc *testClock) Now() time.Time {
	return time.Unix(0, int64(tc.now.Round(time.Nanosecond)))
}

func (tc *testClock) adv(d time.Duration) {
	tc.now += d
}

func TestTimedSlidingWindowI64(t *testing.T) {
	tc := &testClock{}
	sw := NewTimedSlidingWindowI64(5*time.Second, 1*time.Second, tc)
	test := func(adv time.Duration, add int64, expected []int64) {
		tc.adv(adv)
		sw.Add(add)
		assert.Equal(t, expected, sw.buckets)
		total := int64(0)
		for i := 0; i < len(expected); i++ {
			total += expected[i]
		}
		assert.Equal(t, total, sw.Total())
	}
	// [  0   1   2   3   4]
	//    5
	test(0, 5, []int64{5, 0, 0, 0, 0})

	// [  0   1   2   3   4]
	//   11
	test(0, 6, []int64{11, 0, 0, 0, 0})

	// [  0   1   2   3   4]
	//   11       4
	test(2*time.Second, 4, []int64{11, 0, 4, 0, 0})

	// [  0   1   2   3   4]
	//        2   4
	test(4*time.Second, 2, []int64{0, 2, 4, 0, 0})

	// [  0   1   2   3   4]
	//    9
	test(10*time.Second, 9, []int64{9, 0, 0, 0, 0})

	sw.Reset()
	test(10*time.Second, 7, []int64{7, 0, 0, 0, 0})

	assert.PanicsWithValue(t, "window must be a non-zero multiple of non-zero bucket", func() {
		NewTimedSlidingWindowI64(0, time.Second)
	})

	assert.PanicsWithValue(t, "window must be a non-zero multiple of non-zero bucket", func() {
		NewTimedSlidingWindowI64(time.Minute, 0, tc)
	})
}

const (
	tswBenchSeed          = int64(1234)
	tswBenchWindow        = time.Minute
	tswBenchBucket        = time.Second
	tswBenchAddCount      = 100000
	tswBenchAddRange      = 1000
	tswBenchClockAdvRange = 2 * time.Minute
)

func BenchmarkTimedSlidingWindowI64(b *testing.B) {
	rand.Seed(tswBenchSeed)
	tc := &testClock{}
	sw := NewTimedSlidingWindowI64(tswBenchWindow, tswBenchBucket, tc)
	for i := 0; i < b.N; i++ {
		sw.Reset()
		for j := 0; j < tswBenchAddCount; j++ {
			tc.adv(time.Duration(rand.Int63() % int64(tswBenchClockAdvRange)))
			add := rand.Int63() % tswBenchAddRange
			sw.Add(add)
			sw.Total()
		}
	}
}
