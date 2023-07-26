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
<<<<<<< HEAD
=======
		ctx, cancel := context.WithCancel(context.Background())
		// cancel is called at the end of the loop
		go func() {
			select {
			case <-stop:
				cancel()
			case <-ctx.Done():
			}
		}()
>>>>>>> 1f97d444d (fix(kuma-cp): don't leak goroutine on every tick in SimpleWatchdog (#7348))
		select {
		case <-ticker.C:
			if err := w.OnTick(); err != nil {
				w.OnError(err)
			}
		case <-stop:
			if w.OnStop != nil {
				w.OnStop()
			}
			// cancel will be called by the above goroutine
			return
		}
		cancel()
	}
}
