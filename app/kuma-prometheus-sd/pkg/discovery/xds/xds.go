package xds

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/discovery/targetgroup"

	"github.com/kumahq/kuma/app/kuma-prometheus-sd/pkg/discovery/xds/common"
	v1 "github.com/kumahq/kuma/app/kuma-prometheus-sd/pkg/discovery/xds/v1"
)

type discoverer struct {
	factory common.DiscovererFactory
	log     logr.Logger
	config  common.DiscoveryConfig
}

func NewDiscoverer(config common.DiscoveryConfig, log logr.Logger) (discovery.Discoverer, error) {
	var factory common.DiscovererFactory
	switch config.ApiVersion {
	case common.V1:
		factory = v1.NewFactory()
	default:
		return nil, fmt.Errorf("invalid MADS apiVersion %s", config.ApiVersion)
	}

	return &discoverer{
		factory: factory,
		log:     log,
		config:  config,
	}, nil
}

// Run implements discovery.Discoverer interface.
func (d *discoverer) Run(ctx context.Context, ch chan<- []*targetgroup.Group) {
	// notice that Prometheus discovery.Discoverer abstraction doesn't allow failures,
	// so we must ensure that xDS client is up-and-running all the time.
	for streamID := uint64(1); ; streamID++ {
		logger := d.log.WithValues("streamID", streamID)
		errCh := make(chan error, 1)
		go func(errCh chan<- error) {
			defer close(errCh)
			// recover from a panic
			defer func() {
				if e := recover(); e != nil {
					if err, ok := e.(error); ok {
						errCh <- err
					} else {
						errCh <- fmt.Errorf("%v", e)
					}
				}
			}()
			stream := d.factory.CreateDiscoverer(d.config, logger)
			errCh <- stream.Run(ctx, ch)
		}(errCh)
		select {
		case <-ctx.Done():
			logger.Info("done")
			break
		case err := <-errCh:
			if err != nil {
				logger.Error(err, "xDS stream terminated with an error")
			}
		}
	}
}
