package testutils

import (
	"sync"
	"time"
)

type clockTimer struct {
	until time.Time
	fn    func()
}

type MockClock struct {
	mu  sync.Mutex
	now time.Time

	timer clockTimer
}

func NewMockClock() *MockClock {
	return &MockClock{
		now: time.Now(),
	}
}

func (c *MockClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *MockClock) NewTicker(d time.Duration) *time.Ticker {
	ch := make(chan time.Time, 1)

	var tick func(time.Time)

	tick = func(now time.Time) {
		next := now.Add(d)
		c.runAt(next, func() {
			defer tick(next)

			select {
			case ch <- next:
			default:
			}
		})
	}

	tick(c.Now())

	return &time.Ticker{C: ch}
}

func (c *MockClock) runAt(t time.Time, fn func()) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.timer = clockTimer{
		until: t,
		fn:    fn,
	}
}

func (c *MockClock) Add(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	newTime := c.now.Add(d)

	timer := c.timer
	if timer.until.After(newTime) {
		c.now = newTime
		return
	}

	c.now = timer.until
	c.mu.Unlock()
	timer.fn()
	// other goroutines may be waiting for the timer to run
	time.Sleep(1 * time.Millisecond)
	c.mu.Lock()

	c.now = newTime
}
