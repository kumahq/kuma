package xds

import (
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	"github.com/kumahq/kuma/v3/pkg/xds/envoy"
	xds_clusters "github.com/kumahq/kuma/v3/pkg/xds/envoy/clusters"
)

// Cluster describes the passthrough cluster a filter chain forwards to. Hostname and
// Port are set for a domain match, so the sidecar resolves the destination itself
// instead of following the original destination: the client picks both the address it
// dials and the SNI or Host it sends, so an ORIGINAL_DST cluster lets it reach any
// address with an allowed domain.
type Cluster struct {
	Protocol core_meta.Protocol
	Hostname string
	Port     uint32
}

func CreateCluster(apiVersion core_xds.APIVersion, name string, cluster Cluster, hasIPv6 bool) (envoy.NamedResource, error) {
	clusterBuilder := xds_clusters.NewClusterBuilder(apiVersion, name).
		Configure(xds_clusters.DefaultTimeout())
	if cluster.Hostname != "" {
		clusterBuilder.Configure(xds_clusters.ProvidedEndpointCluster(hasIPv6, core_xds.Endpoint{
			Target: cluster.Hostname,
			Port:   cluster.Port,
		}))
	} else {
		clusterBuilder.Configure(xds_clusters.PassThroughCluster())
	}
	switch cluster.Protocol {
	case core_meta.ProtocolGRPC, core_meta.ProtocolHTTP2:
		clusterBuilder.Configure(xds_clusters.Http2())
	}
	return clusterBuilder.Build()
}

// pinnedHostname returns the hostname the cluster resolves on its own, empty when the
// match has no destination to pin: a wildcard domain has no address to resolve, and
// without a port there is nothing to connect to. A domain without a usable port is
// dropped before it gets here, the check keeps a cluster off port 0 either way.
func pinnedHostname(matchType MatchType, value string, port uint32) string {
	if matchType != Domain || port == 0 {
		return ""
	}
	return value
}
