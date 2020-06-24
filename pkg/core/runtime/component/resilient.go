package component

import (
	"time"

	k8s_manager "sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
)

const (
	backoffTime = 5 * time.Second
)

type ComponentFactory func(log logr.Logger) k8s_manager.Runnable

type resilientComponent struct {
	log     logr.Logger
	factory ComponentFactory
}

func NewResilientComponent(log logr.Logger, factory ComponentFactory) Component {
	return &resilientComponent{
		log:     log,
		factory: factory,
	}
}

func (r *resilientComponent) Start(stop <-chan struct{}) error {
	r.log.Info("starting resilient component ...")
	for generationID := uint64(1); ; generationID++ {
		errCh := make(chan error, 1)
		go func(errCh chan<- error) {
			defer close(errCh)
			// recover from a panic
			defer func() {
				if e := recover(); e != nil {
					if err, ok := e.(error); ok {
						errCh <- err
					} else {
						errCh <- errors.Errorf("%v", e)
					}
				}
			}()

			comp := r.factory(r.log.WithValues("generationID", generationID))
			errCh <- comp.Start(stop)
		}(errCh)
		select {
		case <-stop:
			r.log.Info("done")
			break
		case err := <-errCh:
			if err != nil {
				r.log.WithValues("generationID", generationID).Error(err, "component terminated with an error")
			}
		}
		<-time.After(backoffTime)
	}
}

func (r *resilientComponent) NeedLeaderElection() bool {
	return false
}
