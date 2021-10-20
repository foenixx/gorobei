package limiter

import (
	"gorobei/clock"
	"testing"
	"time"
)

func Test_delayFunc(t *testing.T) {
	//main.initLog(false)
	rl := NewLimiter().Init()
	for i := 0; i <= 19; i++ {
		t.Logf("Step %v: %v", i, rl.delayFunc(float64(i)))
	}
}

func Test_delayByCallNo(t *testing.T) {
	//main.initLog(false)
	rl := NewLimiter().Init()
	sum := 1.0
	for i := 0; i <= 19; i++ {
		d := rl.delayByCallNo(i)
		sum += d
		t.Logf("Call %v: %v", i, d)
	}
	t.Logf("Sum: %v", sum)
}

func Test_nextDelay(t *testing.T) {
	//main.initLog(false)
	rl := NewLimiter()
	tc := &clock.TestClock{Tm: time.Now()}
	rl.clock = tc
	rl.Init()
	for i := 0; i <= 100; i++ {
		d := rl.nexDelay()
		t.Logf("Delay %v (%v): %v", i, tc.Now().Format("15:04:05"), d)
		tc.Add(1 *time.Second)
	}

}

func Test_TikTak(t *testing.T) {
	//main.initLog(false)
	rl := NewLimiter().Init()

	for i := 0; i <= 10; i++ {
		t.Logf("Delay %v: %v", i, time.Now().Format("15:04:05"))
		rl.TikTak()
	}
}
