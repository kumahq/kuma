package envoy_admin

import (
	"github.com/kumahq/kuma/test/framework/envoy_admin/clusters"
	"github.com/kumahq/kuma/test/framework/envoy_admin/stats"
)

type Tunnel interface {
	GetStats(name string) (*stats.Stats, error)
	GetClusters() (*clusters.Clusters, error)
	ResetCounters() error
}
