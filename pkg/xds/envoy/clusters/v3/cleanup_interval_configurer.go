package clusters

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	"google.golang.org/protobuf/types/known/durationpb"
)

type CleanupIntervalConfigurer struct {
	Interval int64
}

var _ ClusterConfigurer = &CleanupIntervalConfigurer{}

func (config *CleanupIntervalConfigurer) Configure(c *envoy_cluster.Cluster) error {
	c.CleanupInterval = &durationpb.Duration{Seconds: config.Interval}
	return nil
}
