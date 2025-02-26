package watchdog

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/util/channels"
)

type SimpleWatchdog struct {
	NewTicker     func() *time.Ticker
	OnTick        func(context.Context) error
	OnError       func(error)
	OnStop        func()
	hasTickedChan chan struct{}
}

func (w *SimpleWatchdog) WithTickCheck() *SimpleWatchdog {
	w.hasTickedChan = make(chan struct{})
	return w
}

// WaitForFirstTick return whether this has ticked at least once. This is an optional feature and is opt-in by using `WithTickCheck` after creation
func (w *SimpleWatchdog) HasTicked(blocking bool) bool {
	if w.hasTickedChan == nil {
		panic("Calling HasTicked() before watchdog was started this is not supposed to happen")
	}
	if blocking {
		<-w.hasTickedChan
		return true
	}
	select {
	case <-w.hasTickedChan:
		return true
	default:
		return false
	}
}

func (w *SimpleWatchdog) Start(ctx context.Context) {
	ticker := w.NewTicker()
	defer ticker.Stop()
	for {
		if err := w.onTick(ctx); err != nil {
			if !channels.IsClosed(ctx.Done()) && !errors.Is(err, context.Canceled) && w.OnError != nil {
				w.OnError(err)
			}
		}
		if w.hasTickedChan != nil && !channels.IsClosed(w.hasTickedChan) { // On the first tick we close the channel
			close(w.hasTickedChan)
		}
		select {
		case <-ctx.Done():
		case <-ticker.C:
		}
		// cases are non prioritized so we first check is the context is done
		select {
		case <-ctx.Done():
			if w.OnStop != nil {
				w.OnStop()
			}
			return
		default:
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
