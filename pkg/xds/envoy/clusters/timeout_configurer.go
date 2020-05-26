package clusters

import (
	"time"

	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/golang/protobuf/ptypes"
)

const defaultConnectTimeout = 5 * time.Second

func ConnectTimeout(timeout time.Duration) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.Add(&timeoutConfigurer{
			connectTimeout: timeout,
		})
	})
}

type timeoutConfigurer struct {
	connectTimeout time.Duration
}

func (t *timeoutConfigurer) Configure(cluster *envoy_api.Cluster) error {
	if t.connectTimeout.Nanoseconds() == 0 {
		t.connectTimeout = defaultConnectTimeout
	}
	cluster.ConnectTimeout = ptypes.DurationProto(t.connectTimeout)
	return nil
}
