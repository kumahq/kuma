package clusters

import (
	"time"

	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/golang/protobuf/ptypes"
)

const defaultConnectTimeout = 5 * time.Second

func ConnectTimeout(timeout time.Duration) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.Add(&TimeoutConfigurer{
			ConnectTimeout: timeout,
		})
	})
}

type TimeoutConfigurer struct {
	ConnectTimeout time.Duration
}

func (t *TimeoutConfigurer) Configure(cluster *envoy_api.Cluster) error {
	if t.ConnectTimeout.Nanoseconds() == 0 {
		t.ConnectTimeout = defaultConnectTimeout
	}
	cluster.ConnectTimeout = ptypes.DurationProto(t.ConnectTimeout)
	return nil
}
