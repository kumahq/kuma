package clusters

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"google.golang.org/protobuf/types/known/structpb"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/util/maps"
	"github.com/kumahq/kuma/v3/pkg/util/proto"
	envoy_metadata "github.com/kumahq/kuma/v3/pkg/xds/envoy/metadata/v3"
	"github.com/kumahq/kuma/v3/pkg/xds/envoy/tags"
)

type UpstreamTLSContextConfigure struct {
	Config *envoy_tls.UpstreamTlsContext
	// ZoneMatches maps a zone name to the UpstreamTlsContext that endpoints in
	// that zone must use. When non-empty, every entry is added as a
	// transport_socket_match keyed on the kuma.io/zone endpoint metadata, while
	// Config stays the cluster-wide default for endpoints that don't match any
	// zone (local zone and zones that share the default SNI format).
	ZoneMatches map[string]*envoy_tls.UpstreamTlsContext
}

var _ ClusterConfigurer = &UpstreamTLSContextConfigure{}

func (c *UpstreamTLSContextConfigure) Configure(cluster *envoy_cluster.Cluster) error {
	transportSocket, err := createTLSTransportSocket(c.Config)
	if err != nil {
		return err
	}
	cluster.TransportSocket = transportSocket

	for _, zone := range maps.SortedKeys(c.ZoneMatches) {
		ts, err := createTLSTransportSocket(c.ZoneMatches[zone])
		if err != nil {
			return err
		}
		cluster.TransportSocketMatches = append(cluster.TransportSocketMatches, &envoy_cluster.Cluster_TransportSocketMatch{
			Name: zone,
			Match: &structpb.Struct{
				Fields: envoy_metadata.MetadataFields(tags.Tags{mesh_proto.ZoneTag: zone}),
			},
			TransportSocket: ts,
		})
	}
	return nil
}

func createTLSTransportSocket(config *envoy_tls.UpstreamTlsContext) (*envoy_core.TransportSocket, error) {
	pbst, err := proto.MarshalAnyDeterministic(config)
	if err != nil {
		return nil, err
	}
	return &envoy_core.TransportSocket{
		Name: "envoy.transport_sockets.tls",
		ConfigType: &envoy_core.TransportSocket_TypedConfig{
			TypedConfig: pbst,
		},
	}, nil
}
