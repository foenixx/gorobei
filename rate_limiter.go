package main

import (
	"math"
	"time"
)

// RateLimiter implements telegram message limits:
//  In particular chat:
//  - 1 message per second
//  - 20 messages per minute
type (
	RateLimiter struct {
		maxRpm    float64
		maxRps    float64
		calls     int
		firstCall time.Time
		delayFunc func(calls float64) float64
		clock 	Clock
	}

	Clock interface {
		Now() time.Time
	}

	RealClock struct {

	}
)

func (r *RealClock) Now() time.Time {
	return time.Now()
}

const (
	MaxRatePerMinute = 20.0
	MaxRatePerSecond = 1.0
)

func NewLimiter() *RateLimiter {
	l := &RateLimiter{
		maxRpm:    MaxRatePerMinute,
		maxRps:    MaxRatePerSecond,
		clock: &RealClock{},
	}
	// https://www.wolframalpha.com/input/?i=plot+e%5E%28x%2F%2819%2F5%29%29-1+from+x%3D0+to+19
	// scale f=(e^x-1), x=[0..5] to the x=[0..maxRpm-1] which is e^(x/(maxRpm-1)/5)-1
	//
	l.delayFunc = func(x float64) float64 {
		return math.Pow(math.E, x/((l.maxRpm-1)/5.0)) - 1
	}
	return l
}

func (r *RateLimiter) Init() *RateLimiter {
	r.firstCall = r.clock.Now()
	r.calls = 0
	return r
}

// delayByCallNo calculates the delay (in seconds) from the number of current call within a last minute
func (r *RateLimiter) delayByCallNo(call int) float64 {
	var cur, prev float64
	cur = r.delayFunc(float64(call))
	if call > 0 {
		prev = r.delayFunc(float64(call) - 1)
	} else {
		prev = 0
	}
	// 1.0/maxRps - minimal delayByStep between messages
	// 60 - maxRpm/maxRps - "extra" delayByStep which we spread between messages according to proportion set by delayFunc
	return 1.0/r.maxRps + (cur-prev)/r.delayFunc(r.maxRpm-1)*(60-r.maxRpm/r.maxRps)
}

func (r *RateLimiter) nexDelay() time.Duration {

	now := r.clock.Now()
	elapsed := now.Sub(r.firstCall)
	if elapsed >= 1*time.Minute {
		// a minute or more passed, starting over
		r.firstCall = now
		r.calls = 0
	}

	delay := time.Duration(r.delayByCallNo(r.calls) * float64(time.Second))
	r.calls += 1
	return delay
}

func (r *RateLimiter) TikTak() {
	// waiting for the next call window
	<-time.After(r.nexDelay())
}
