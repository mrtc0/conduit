package strategy

import "time"

var DefaultClock = &defaultClock{}

type Clock interface {
	Now() time.Time
	NewTicker(d time.Duration) *time.Ticker
}

type defaultClock struct{}

func (c *defaultClock) Now() time.Time {
	return time.Now()
}

func (c *defaultClock) NewTicker(d time.Duration) *time.Ticker {
	return time.NewTicker(d)
}
