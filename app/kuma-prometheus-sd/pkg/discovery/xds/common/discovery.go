package common

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/prometheus/prometheus/discovery/targetgroup"
)

// DiscovererE is like discovery.Discoverer but allows returning errors
type DiscovererE interface {
	Run(ctx context.Context, ch chan<- []*targetgroup.Group) error
}

// DiscovererFactory creates DiscovererE instances
type DiscovererFactory interface {
	CreateDiscoverer(config DiscoveryConfig, log logr.Logger) DiscovererE
}
