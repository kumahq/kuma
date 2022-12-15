package test

import (
	"sync"
	"time"
)

type Clock struct {
	now time.Time
	sync.Mutex
}

func NewClock(t time.Time) *Clock {
	return &Clock{now: t}
}

func (c *Clock) Now() time.Time {
	c.Lock()
	defer c.Unlock()
	return c.now
}

func (c *Clock) Add(duration time.Duration) {
	c.Lock()
	c.now = c.now.Add(duration)
	c.Unlock()
}
