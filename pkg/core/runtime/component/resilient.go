package component

import (
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-retry"
)

const (
	resilientLogSampleAfterAttempt = 5
	resilientLogSampleWindow       = 10 * time.Second
)

type resilientComponent struct {
	log             logr.Logger
	component       Component
	backoffBaseTime time.Duration
	backoffMaxTime  time.Duration
}

func NewResilientComponent(log logr.Logger, component Component, backoffBaseTime time.Duration, backoffMaxTime time.Duration) Component {
	return &resilientComponent{
		log:             log,
		component:       component,
		backoffBaseTime: backoffBaseTime,
		backoffMaxTime:  backoffMaxTime,
	}
}

func (r *resilientComponent) Start(stop <-chan struct{}) error {
	r.log.Info("starting resilient component")
	backoff := r.newBackoff()
	var (
		lastError      error
		lastSuccess    = time.Now()
		nextSampledAt  time.Time
		suppressedLogs uint64
	)
	for generationID := uint64(1); ; generationID++ {
		lastStart := time.Now()
		errCh := make(chan error, 1)
		go func(errCh chan<- error) {
			defer close(errCh)
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
				lastError = err
				r.log.WithValues("generationID", generationID).Error(err, "component terminated with an error")
			} else {
				// Clean exit: don't carry a prior failure into the next restart log.
				lastError = nil
			}
		}
		if time.Since(lastStart) > r.backoffMaxTime {
			// Stable run ended: reset backoff, success baseline, and sampling
			// state so a new outage starts fresh.
			backoff = r.newBackoff()
			lastSuccess = time.Now()
			nextSampledAt = time.Time{}
			suppressedLogs = 0
		}
		dur, _ := backoff.Next()
		if generationID <= resilientLogSampleAfterAttempt || !time.Now().Before(nextSampledAt) {
			fields := []any{
				"attempt", generationID,
				"sinceLastSuccess", time.Since(lastSuccess),
				"nextBackoff", dur,
			}
			if lastError != nil {
				fields = append(fields, "lastError", lastError.Error())
			}
			if suppressedLogs > 0 {
				fields = append(fields, "suppressed", suppressedLogs)
				suppressedLogs = 0
			}
			r.log.Info("scheduling component restart", fields...)
			if generationID > resilientLogSampleAfterAttempt {
				nextSampledAt = time.Now().Add(resilientLogSampleWindow)
			}
		} else {
			suppressedLogs++
		}
		<-time.After(dur)
	}
}

func (r *resilientComponent) newBackoff() retry.Backoff {
	return retry.WithJitter(r.backoffBaseTime, retry.WithCappedDuration(r.backoffMaxTime, retry.NewExponential(r.backoffBaseTime)))
}

func (r *resilientComponent) NeedLeaderElection() bool {
	return r.component.NeedLeaderElection()
}
