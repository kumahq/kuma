package passivehealth

import (
	"time"

	"github.com/kumahq/kuma/pkg/core"
)

type LastSeenSetter interface {
	SetLastSeen(t time.Time, period time.Duration)
}

type Checker struct {
	m      map[string]*delayedPassiveHC
	period time.Duration
	delta  time.Duration
}

func NewChecker(period, delta time.Duration) *Checker {
	return &Checker{
		m:      map[string]*delayedPassiveHC{},
		period: period,
		delta:  delta,
	}
}

func (hc *Checker) MarkAsAlive(key string, s LastSeenSetter) {
	if _, ok := hc.m[key]; !ok {
		hc.m[key] = &delayedPassiveHC{
			period:   hc.period,
			delta:    hc.delta,
			lastSeen: core.Now().Add(-2 * hc.period),
		}
	}
	hc.m[key].MarkAsAlive(s)
}

type delayedPassiveHC struct {
	lastSeen time.Time
	period   time.Duration
	delta    time.Duration
}

func (hc *delayedPassiveHC) MarkAsAlive(s LastSeenSetter) {
	now := core.Now()
	if hc.lastSeen.Add(hc.period).Before(now) {
		s.SetLastSeen(now, hc.period+hc.delta)
		hc.lastSeen = now
	} else {
		s.SetLastSeen(hc.lastSeen, hc.period+hc.delta)
	}
}
