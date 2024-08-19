package watchdog

import (
	"context"
	"time"

	"github.com/pkg/errors"
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
		// cancel is called at the end of the loop
		go func() {
			select {
			case <-stop:
				cancel()
			case <-ctx.Done():
			}
		}()

		select {
		case <-stop:
		default:
			if err := w.onTick(ctx); err != nil && !errors.Is(err, context.Canceled) {
				w.OnError(err)
			}
		}
		cancel()

		select {
		case <-ticker.C:
			continue
		case <-stop:
			if w.OnStop != nil {
				w.OnStop()
			}
			// cancel will be called by the above goroutine
			return
		}
	}
}

func (w *SimpleWatchdog) onTick(ctx context.Context) error {
	defer func() {
		if cause := recover(); cause != nil {
			if w.OnError != nil {
				var err error
				switch typ := cause.(type) {
				case error:
					err = errors.WithStack(typ)
				default:
					err = errors.Errorf("%v", cause)
				}
				w.OnError(err)
			}
		}
	}()
	return w.OnTick(ctx)
}
