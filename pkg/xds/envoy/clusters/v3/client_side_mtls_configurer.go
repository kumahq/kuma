package clusters

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"google.golang.org/protobuf/types/known/structpb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_metadata "github.com/kumahq/kuma/pkg/xds/envoy/metadata/v3"
	"github.com/kumahq/kuma/pkg/xds/envoy/tls"
	envoy_tls "github.com/kumahq/kuma/pkg/xds/envoy/tls/v3"
)

type ClientSideMTLSConfigurer struct {
	Mesh             *core_mesh.MeshResource
	UpstreamService  string
	Tags             []envoy.Tags
	UpstreamTLSReady bool
}

var _ ClusterConfigurer = &ClientSideTLSConfigurer{}

func (c *ClientSideMTLSConfigurer) Configure(cluster *envoy_cluster.Cluster) error {
	if !c.Mesh.MTLSEnabled() {
		return nil
	}
	if c.Mesh.GetEnabledCertificateAuthorityBackend().Mode == mesh_proto.CertificateAuthorityBackend_PERMISSIVE &&
		!c.UpstreamTLSReady {
		return nil
	}

	meshName := c.Mesh.GetMeta().GetName()
	// there might be a situation when there are multiple sam tags passed here for example two outbound listeners with the same tags, therefore we need to distinguish between them.
	distinctTags := envoy.DistinctTags(c.Tags)
	switch {
	case len(distinctTags) == 0:
		transportSocket, err := c.createTransportSocket("")
		if err != nil {
			return err
		}
		cluster.TransportSocket = transportSocket
	case len(distinctTags) == 1:
		transportSocket, err := c.createTransportSocket(tls.SNIFromTags(c.Tags[0].WithTags("mesh", meshName)))
		if err != nil {
			return err
		}
		cluster.TransportSocket = transportSocket
	default:
		for _, tags := range distinctTags {
			sni := tls.SNIFromTags(tags.WithTags("mesh", meshName))
			transportSocket, err := c.createTransportSocket(sni)
			if err != nil {
				return err
			}
			cluster.TransportSocketMatches = append(cluster.TransportSocketMatches, &envoy_cluster.Cluster_TransportSocketMatch{
				Name: sni,
				Match: &structpb.Struct{
					Fields: envoy_metadata.MetadataFields(tags),
				},
				TransportSocket: transportSocket,
			})
		}
	}
	return nil
}

func (c *ClientSideMTLSConfigurer) createTransportSocket(sni string) (*envoy_core.TransportSocket, error) {
	tlsContext, err := envoy_tls.CreateUpstreamTlsContext(c.Mesh, c.UpstreamService, sni)
	if err != nil {
		return nil, err
	}
	if tlsContext == nil {
		return nil, nil
	}
	pbst, err := proto.MarshalAnyDeterministic(tlsContext)
	if err != nil {
		return nil, err
	}
	transportSocket := &envoy_core.TransportSocket{
		Name: "envoy.transport_sockets.tls",
		ConfigType: &envoy_core.TransportSocket_TypedConfig{
			TypedConfig: pbst,
		},
	}
	return transportSocket, nil
}
