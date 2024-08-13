package watchdog

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

type SimpleWatchdog struct {
	NewTicker func() *time.Ticker
	OnTick    func(context.Context) error
	OnError   func(error)
	OnStop    func()
}

func (w *SimpleWatchdog) Start(ctx context.Context) {
	ticker := w.NewTicker()
	defer ticker.Stop()

	for {
		if err := w.onTick(ctx); err != nil {
			if !errors.Is(err, context.Canceled) && w.OnError != nil {
				w.OnError(err)
			}
		}
		select {
		case <-ticker.C:
		case <-ctx.Done():
			if w.OnStop != nil {
				w.OnStop()
			}
			return
		}
	}
}

func (w *SimpleWatchdog) onTick(ctx context.Context) (err error) {
	defer func() {
		if cause := recover(); cause != nil {
			switch typ := cause.(type) {
			case error:
				err = errors.WithStack(typ)
			default:
				err = errors.Errorf("%v", cause)
			}
		}
	}()
	return w.OnTick(ctx)
}
