package watchdog

import (
	"time"

	"github.com/pkg/errors"
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
			if err := w.onTick(); err != nil {
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

func (w *SimpleWatchdog) onTick() error {
	defer func() {
		if cause := recover(); cause != nil {
			if w.OnError != nil {
				var err error
				switch typ := cause.(type) {
				case error:
					err = typ
				default:
					err = errors.Errorf("%v", cause)
				}
				w.OnError(err)
			}
		}
	}()
	return w.OnTick()
}
