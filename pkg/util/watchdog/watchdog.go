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
	// StreamCtx is an optional context tied to the gRPC stream.
	// When set, the watchdog will skip ticks if the stream is closed,
	// preventing race conditions between stream closure and snapshot updates.
	StreamCtx context.Context
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
		case <-ticker.C:
			select {
			case <-stop:
			default:
				// If stream context is set and closed, skip tick and stop
				if w.StreamCtx != nil {
					select {
					case <-w.StreamCtx.Done():
						cancel()
						if w.OnStop != nil {
							w.OnStop()
						}
						return
					default:
					}
				}
				if err := w.onTick(ctx); err != nil && !errors.Is(err, context.Canceled) {
					w.OnError(err)
				}
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
