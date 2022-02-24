package times

import (
	"time"

	"github.com/jf-tech/go-corelib/maths"
)

/*
Until we move to golang generic, the interface{} based generic implementation is simply too
slow, compared with raw type (int, int64, etc) implementation. For example, we compared this
generic interface{} based implementation against an nearly identical but with direct int type
implementation, the benchmark is not even close: too many int<->interface{} conversion induced
heap escape:

BenchmarkTimedSlidingWindowIntRaw-8     	     100	  11409978 ns/op	     600 B/op	       3 allocs/op
BenchmarkTimedSlidingWindowIntIFace-8   	      31	  37520116 ns/op	11837979 B/op	 1479595 allocs/op

So the decision is to comment out the interface{} implementation for reference only.

type TimedSlidingWindowOp func(a, b interface{}) interface{}

type TimedSlidingWindowCfg struct {
	Clock             Clock
	Window, Bucket    time.Duration
	Adder, Subtracter TimedSlidingWindowOp
}

type TimedSlidingWindow struct {
	cfg        TimedSlidingWindowCfg
	n          int
	buckets    []interface{}
	start, end int
	startTime  time.Time
	total      interface{}
}

func (s *TimedSlidingWindow) Add(amount interface{}) {
	now := s.cfg.Clock.Now()
	idx := int(now.Sub(s.startTime) / s.cfg.Bucket)
	e2 := s.end
	if s.end < s.start {
		e2 += s.n
	}
	if s.start+idx-e2 < s.n {
		for i := e2 + 1; i <= s.start+idx; i++ {
			s.total = s.cfg.Subtracter(s.total, s.buckets[i%s.n])
			s.buckets[i%s.n] = nil
		}
		s.end = (s.start + idx) % s.n
		newStart := maths.MaxInt(s.start+idx-s.n+1, s.start)
		s.startTime = s.startTime.Add(time.Duration(newStart-s.start) * s.cfg.Bucket)
		s.start = newStart
		s.buckets[s.end] = s.cfg.Adder(s.buckets[s.end], amount)
		s.total = s.cfg.Adder(s.total, amount)
	} else {
		for i := 0; i < s.n; i++ {
			s.buckets[i] = nil
		}
		s.start, s.end = 0, 0
		s.buckets[0] = amount
		s.total = amount
		s.startTime = now
	}
}

func (s *TimedSlidingWindow) Total() interface{} {
	s.Add(nil)
	return s.total
}

func NewTimedSlidingWindow(cfg TimedSlidingWindowCfg) *TimedSlidingWindow {
	if cfg.Window == 0 || cfg.Window%cfg.Bucket != 0 {
		panic("time window must be non-zero multiple of bucket")
	}
	n := int(cfg.Window / cfg.Bucket)
	return &TimedSlidingWindow{
		cfg:       cfg,
		n:         n,
		buckets:   make([]interface{}, n),
		startTime: cfg.Clock.Now(),
	}
}
*/

type TimedSlidingWindowI64 struct {
	clock          Clock
	window, bucket time.Duration
	n              int
	buckets        []int64
	start, end     int
	startTime      time.Time
	total          int64
}

func (t *TimedSlidingWindowI64) Add(amount int64) {
	now := t.clock.Now()
	idx := int(now.Sub(t.startTime) / t.bucket)
	e2 := t.end
	if t.end < t.start {
		e2 += t.n
	}
	if t.start+idx-e2 < t.n {
		for i := e2 + 1; i <= t.start+idx; i++ {
			t.total -= t.buckets[i%t.n]
			t.buckets[i%t.n] = 0
		}
		t.end = (t.start + idx) % t.n
		newStart := maths.MaxInt(t.start+idx-t.n+1, t.start)
		t.startTime = t.startTime.Add(time.Duration(newStart-t.start) * t.bucket)
		t.start = newStart
		t.buckets[t.end] += amount
		t.total += amount
	} else {
		for i := 0; i < t.n; i++ {
			t.buckets[i] = 0
		}
		t.start, t.end = 0, 0
		t.buckets[0] = amount
		t.total = amount
		t.startTime = now
	}
}

func (t *TimedSlidingWindowI64) Total() int64 {
	t.Add(0)
	return t.total
}

func (t *TimedSlidingWindowI64) Reset() {
	for i := 0; i < t.n; i++ {
		t.buckets[i] = 0
	}
	t.start, t.end = 0, 0
	t.startTime = t.clock.Now()
	t.total = 0
}

func NewTimedSlidingWindowI64(window, bucket time.Duration, clock ...Clock) *TimedSlidingWindowI64 {
	if window == 0 || bucket == 0 || window%bucket != 0 {
		panic("window must be a non-zero multiple of non-zero bucket")
	}
	c := Clock(NewOSClock())
	if len(clock) > 0 {
		c = clock[0]
	}
	n := int(window / bucket)
	return &TimedSlidingWindowI64{
		clock:     c,
		window:    window,
		bucket:    bucket,
		n:         n,
		buckets:   make([]int64, n),
		startTime: c.Now(),
	}
}
