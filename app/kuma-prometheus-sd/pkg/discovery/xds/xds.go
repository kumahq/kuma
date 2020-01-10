package xds

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"github.com/prometheus/prometheus/discovery/targetgroup"
)

type discoverer struct {
	log     logr.Logger
	config  DiscoveryConfig
	handler Handler
}

func NewDiscoverer(config DiscoveryConfig, log logr.Logger) (*discoverer, error) {
	return &discoverer{
		log:    log,
		config: config,
	}, nil
}

// Run implements discovery.Discoverer interface.
func (d *discoverer) Run(ctx context.Context, ch chan<- []*targetgroup.Group) {
	// notice that Prometheus discovery.Discoverer abstraction doesn't allow failures,
	// so we must ensure that xDS client is up-and-running all the time.
	for streamID := uint64(1); ; streamID++ {
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
			stream := stream{
				log:     d.log.WithValues("streamID", streamID),
				config:  d.config,
				handler: &d.handler,
			}
			errCh <- stream.Run(ctx, ch)
		}(errCh)
		select {
		case <-ctx.Done():
			d.log.Info("done")
			break
		case err := <-errCh:
			if err != nil {
				d.log.WithValues("streamID", streamID).Error(err, "xDS stream terminated with an error")
			}
		}
	}
}
