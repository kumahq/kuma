package watchdog

import (
	"context"
	"time"
)

type Watchdog interface {
	Start(stop <-chan struct{})
}

type SimpleWatchdog struct {
	NewTicker func() *time.Ticker
	OnTick    func(context.Context) error
	OnError   func(error)
	OnStop    func()
}

func (w *SimpleWatchdog) Start(stop <-chan struct{}) {
	ticker := w.NewTicker()
	defer ticker.Stop()

	for {
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			<-stop
			cancel()
		}()
		select {
		case <-ticker.C:
<<<<<<< HEAD
			if err := w.OnTick(); err != nil {
				w.OnError(err)
=======
			select {
			case <-stop:
			default:
				if err := w.onTick(ctx); err != nil {
					w.OnError(err)
				}
>>>>>>> 8be55a569 (fix(kuma-cp): cancel OnTick when watchdog stopped (#7221))
			}
		case <-stop:
			if w.OnStop != nil {
				w.OnStop()
			}
			return
		}
	}
}
<<<<<<< HEAD
=======

func (w *SimpleWatchdog) onTick(ctx context.Context) error {
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
	return w.OnTick(ctx)
}
>>>>>>> 8be55a569 (fix(kuma-cp): cancel OnTick when watchdog stopped (#7221))
