package clock

import "time"

type (
	Clock interface {
		Now() time.Time
	}

	RealClock struct {
	}

	TestClock struct {
		Tm time.Time
	}
)

func (r *RealClock) Now() time.Time {
	return time.Now()
}

func (tc *TestClock) Now() time.Time {
	return tc.Tm
}

func (tc *TestClock) Add(d time.Duration) {
	tc.Tm = tc.Tm.Add(d)
}
