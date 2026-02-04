package watchdog

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

type SimpleWatchdog struct {
<<<<<<< HEAD
	NewTicker func() *time.Ticker
	OnTick    func(context.Context) error
	OnError   func(error)
	OnStop    func()
=======
	NewTicker     func() *time.Ticker
	OnTick        func(context.Context) error
	OnError       func(error)
	OnStop        func()
	hasTickedChan chan struct{}
	// StreamCtx is an optional context tied to the gRPC stream.
	// When set, the watchdog will skip ticks if the stream is closed,
	// preventing race conditions between stream closure and snapshot updates.
	StreamCtx context.Context
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
>>>>>>> 42c3b352ba (fix(xds): prevent panic on send to closed channel during stream closure (#15511))
}

func (w *SimpleWatchdog) Start(ctx context.Context) {
	ticker := w.NewTicker()
	defer ticker.Stop()

	for {
		// If stream context is set and closed, skip tick and wait for main context
		if w.StreamCtx != nil {
			select {
			case <-w.StreamCtx.Done():
				<-ctx.Done()
				if w.OnStop != nil {
					w.OnStop()
				}
				return
			default:
			}
		}

		if err := w.onTick(ctx); err != nil {
			if !errors.Is(err, context.Canceled) && w.OnError != nil {
				w.OnError(err)
			}
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
