package clusters

import (
	"time"

	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/golang/protobuf/ptypes"
)

const defaultConnectTimeout = 5 * time.Second

type TimeoutConfigurer struct {
	connectTimeout time.Duration
}

var _ ClusterConfigurer = &TimeoutConfigurer{}

func (t *TimeoutConfigurer) Configure(cluster *envoy_api.Cluster) error {
	if t.connectTimeout.Nanoseconds() == 0 {
		t.connectTimeout = defaultConnectTimeout
	}
	cluster.ConnectTimeout = ptypes.DurationProto(t.connectTimeout)
	return nil
}
