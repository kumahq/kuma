package component

import (
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-retry"
)

type resilientComponent struct {
	log             logr.Logger
	component       Component
	backoffBaseTime time.Duration
	backoffMaxTime  time.Duration
}

func NewResilientComponent(log logr.Logger, component Component, backoffBaseTime, backoffMaxTime time.Duration) Component {
	return &resilientComponent{
		log:             log,
		component:       component,
		backoffBaseTime: backoffBaseTime,
		backoffMaxTime:  backoffMaxTime,
	}
}

func (r *resilientComponent) Start(stop <-chan struct{}) error {
	r.log.Info("starting resilient component ...")
	backoff := r.newBackoff()
	for generationID := uint64(1); ; generationID++ {
		lastStart := time.Now()
		errCh := make(chan error, 1)
		go func(errCh chan<- error) {
			defer close(errCh)
			// recover from a panic
			defer func() {
				if e := recover(); e != nil {
					if err, ok := e.(error); ok {
						errCh <- errors.WithStack(err)
					} else {
						errCh <- errors.Errorf("%v", e)
					}
				}
			}()

			errCh <- r.component.Start(stop)
		}(errCh)
		select {
		case <-stop:
			r.log.Info("done")
			return nil
		case err := <-errCh:
			if err != nil {
				r.log.WithValues("generationID", generationID).Error(err, "component terminated with an error")
			}
		}
		if time.Since(lastStart) > r.backoffMaxTime {
			// reset backoff so in a case of event with the following steps
			// 1) Component is unhealthy until max backoff is reached
			// 2) Component is healthy for at least backoff max time
			// 3) Component is unhealthy
			// We start backoff from the beginning
			backoff = r.newBackoff()
		}
		dur, _ := backoff.Next()
		<-time.After(dur)
	}
}

func (r *resilientComponent) newBackoff() retry.Backoff {
	return retry.WithJitter(r.backoffBaseTime, retry.WithCappedDuration(r.backoffMaxTime, retry.NewExponential(r.backoffBaseTime)))
}

func (r *resilientComponent) NeedLeaderElection() bool {
	return r.component.NeedLeaderElection()
}
