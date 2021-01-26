package clusters

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_upstream_http "github.com/envoyproxy/go-control-plane/envoy/extensions/upstreams/http/v3"
	"github.com/golang/protobuf/ptypes/any"

	"github.com/kumahq/kuma/pkg/util/proto"
)

type Http2Configurer struct {
}

var _ ClusterConfigurer = &Http2Configurer{}

func (p *Http2Configurer) Configure(c *envoy_cluster.Cluster) error {
	options := &envoy_upstream_http.HttpProtocolOptions{
		UpstreamProtocolOptions: &envoy_upstream_http.HttpProtocolOptions_ExplicitHttpConfig_{
			ExplicitHttpConfig: &envoy_upstream_http.HttpProtocolOptions_ExplicitHttpConfig{
				ProtocolConfig: &envoy_upstream_http.HttpProtocolOptions_ExplicitHttpConfig_Http2ProtocolOptions{
					Http2ProtocolOptions: &envoy_core.Http2ProtocolOptions{},
				},
			},
		},
	}

	pbst, err := proto.MarshalAnyDeterministic(options)
	if err != nil {
		return err
	}
	c.TypedExtensionProtocolOptions = map[string]*any.Any{
		"envoy.extensions.upstreams.http.v3.HttpProtocolOptions": pbst,
	}
	return nil
}
