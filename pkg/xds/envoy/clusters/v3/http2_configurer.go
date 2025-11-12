package clusters

import (
    "strings"

    envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
    envoy_http "github.com/envoyproxy/go-control-plane/envoy/extensions/upstreams/http/v3"
    util_proto "github.com/kumahq/kuma/v2/pkg/util/proto"
)

const (
    DefaultInitialStreamWindowSize     = 1024 * 1024
    DefaultInitialConnectionWindowSize = 1024 * 1024
)

type Http2Configurer struct {
    EdgeProxyWindowSizes bool
}

func (c *Http2Configurer) Configure(cluster *envoy_cluster.Cluster) error {
    opts := envoy_http.HttpProtocolOptions{}

    // ✅ Enable HTTP/2 for internal proxy-to-proxy clusters
    // even when the service protocol is HTTP/1.1
    if cluster != nil && !strings.Contains(cluster.Name, "external") {
        if cluster.Http2ProtocolOptions == nil {
            cluster.Http2ProtocolOptions = &envoy_http.Http2ProtocolOptions{
                InitialStreamWindowSize:     util_proto.UInt32(DefaultInitialStreamWindowSize),
                InitialConnectionWindowSize: util_proto.UInt32(DefaultInitialConnectionWindowSize),
            }
        }
    }

    // These are from Envoy's best practices for edge proxy clusters
    // https://www.envoyproxy.io/docs/envoy/latest/configuration/best_practices/edge
    if opts.Http2ProtocolOptions == nil {
        opts.Http2ProtocolOptions = &envoy_http.Http2ProtocolOptions{
            InitialStreamWindowSize:     util_proto.UInt32(DefaultInitialStreamWindowSize),
            InitialConnectionWindowSize: util_proto.UInt32(DefaultInitialConnectionWindowSize),
        }
    }

    return nil
}
