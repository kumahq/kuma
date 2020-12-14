package clusters

import (
	"time"

	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	"github.com/golang/protobuf/ptypes"
)

const defaultConnectTimeout = 5 * time.Second

type TimeoutConfigurer struct {
	connectTimeout time.Duration
}

var _ ClusterConfigurer = &TimeoutConfigurer{}

func (t *TimeoutConfigurer) Configure(cluster *envoy_cluster.Cluster) error {
	if t.connectTimeout.Nanoseconds() == 0 {
		t.connectTimeout = defaultConnectTimeout
	}
	cluster.ConnectTimeout = ptypes.DurationProto(t.connectTimeout)
	return nil
}
