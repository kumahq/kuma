package clusters

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"google.golang.org/protobuf/types/known/structpb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/util/proto"
	envoy_metadata "github.com/kumahq/kuma/pkg/xds/envoy/metadata/v3"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
	"github.com/kumahq/kuma/pkg/xds/envoy/tls"
	envoy_tls "github.com/kumahq/kuma/pkg/xds/envoy/tls/v3"
)

type ClientSideMTLSConfigurer struct {
	SecretsTracker        core_xds.SecretsTracker
	UpstreamMesh          *core_mesh.MeshResource
	UpstreamService       string
	LocalMesh             *core_mesh.MeshResource
	Tags                  []tags.Tags
	SNI                   string
	UpstreamTLSReady      bool
	VerifyIdentities      []string
	UnifiedResourceNaming bool
}

var _ ClusterConfigurer = &ClientSideMTLSConfigurer{}

func (c *ClientSideMTLSConfigurer) Configure(cluster *envoy_cluster.Cluster) error {
	if !c.UpstreamMesh.MTLSEnabled() || !c.LocalMesh.MTLSEnabled() {
		return nil
	}
	if c.UpstreamMesh.GetEnabledCertificateAuthorityBackend().Mode == mesh_proto.CertificateAuthorityBackend_PERMISSIVE &&
		!c.UpstreamTLSReady {
		return nil
	}

	meshName := c.UpstreamMesh.GetMeta().GetName()
	// there might be a situation when there are multiple sam tags passed here for example two outbound listeners with the same tags, therefore we need to distinguish between them.
	distinctTags := tags.DistinctTags(c.Tags)
	switch {
	case len(distinctTags) == 0:
		transportSocket, err := c.createTransportSocket(c.SNI)
		if err != nil {
			return err
		}
		cluster.TransportSocket = transportSocket
	case len(distinctTags) == 1:
		sni := tls.SNIFromTags(c.Tags[0].WithTags("mesh", meshName))
		transportSocket, err := c.createTransportSocket(sni)
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
					Fields: envoy_metadata.MetadataFields(tags.WithoutTags(mesh_proto.ServiceTag)),
				},
				TransportSocket: transportSocket,
			})
		}
	}
	return nil
}

func (c *ClientSideMTLSConfigurer) createTransportSocket(sni string) (*envoy_core.TransportSocket, error) {
	if !c.UpstreamMesh.MTLSEnabled() {
		return nil, nil
	}

	ca := c.SecretsTracker.RequestCa(c.UpstreamMesh.GetMeta().GetName())
	identity := c.SecretsTracker.RequestIdentityCert()

	var verifyIdentities []string
	if c.VerifyIdentities != nil {
		verifyIdentities = c.VerifyIdentities
	}
	tlsContext, err := envoy_tls.CreateUpstreamTlsContext(identity, ca, c.UpstreamService, sni, verifyIdentities, c.UnifiedResourceNaming)
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
