package component

import (
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-retry"
)

const (
	backoffBaseTime = 5 * time.Second
	backoffMaxTime  = 1 * time.Minute
)

type resilientComponent struct {
	log       logr.Logger
	component Component
}

func NewResilientComponent(log logr.Logger, component Component) Component {
	return &resilientComponent{
		log:       log,
		component: component,
	}
}

func (r *resilientComponent) Start(stop <-chan struct{}) error {
	r.log.Info("starting resilient component ...")
	backoff := retry.WithCappedDuration(backoffMaxTime, retry.NewExponential(backoffBaseTime))
	for generationID := uint64(1); ; generationID++ {
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
		dur, _ := backoff.Next()
		<-time.After(dur)
	}
}

func (r *resilientComponent) NeedLeaderElection() bool {
	return r.component.NeedLeaderElection()
}
