package watchdog

import (
	"time"
)

type Watchdog interface {
	Start(stop <-chan struct{})
}

type SimpleWatchdog struct {
	NewTicker func() *time.Ticker
	OnTick    func() error
	OnError   func(error)
	OnStop    func()
}

func (w *SimpleWatchdog) Start(stop <-chan struct{}) {
	ticker := w.NewTicker()
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := w.OnTick(); err != nil {
				w.OnError(err)
			}
		case <-stop:
			if w.OnStop != nil {
				w.OnStop()
			}
			return
		}
	}
}
