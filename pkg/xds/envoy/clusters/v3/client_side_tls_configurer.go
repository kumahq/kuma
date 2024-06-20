package clusters

import (
	"github.com/asaskevich/govalidator"
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"google.golang.org/protobuf/types/known/structpb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/util/proto"
	envoy_metadata "github.com/kumahq/kuma/pkg/xds/envoy/metadata/v3"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
	envoy_tls "github.com/kumahq/kuma/pkg/xds/envoy/tls/v3"
)

type ClientSideTLSConfigurer struct {
	Endpoints           []xds.Endpoint
	SystemCaPath        string
	UseCommonTlsContext bool // used to handle MeshExternalService
}

var _ ClusterConfigurer = &ClientSideTLSConfigurer{}

func (c *ClientSideTLSConfigurer) Configure(cluster *envoy_cluster.Cluster) error {
	if c.UseCommonTlsContext && len(c.Endpoints) > 0 {
		ep := c.Endpoints[0]
		if ep.ExternalService != nil && ep.ExternalService.TLSEnabled {
			tsm, err := c.createTransportSocketMatch(&ep, false)
			if err != nil {
				return err
			}
			cluster.TransportSocketMatches = append(cluster.TransportSocketMatches, tsm)
		}
	} else {
		for i, ep := range c.Endpoints {
			if ep.ExternalService != nil && ep.ExternalService.TLSEnabled {
				tsm, err := c.createTransportSocketMatch(&c.Endpoints[i], true)
				if err != nil {
					return err
				}
				cluster.TransportSocketMatches = append(cluster.TransportSocketMatches, tsm)
			}
		}
	}

	return nil
}

func (c *ClientSideTLSConfigurer) createTransportSocketMatch(ep *xds.Endpoint, withMatch bool) (*envoy_cluster.Cluster_TransportSocketMatch, error) {
	sni := ep.ExternalService.ServerName
	if ep.ExternalService.ServerName == "" && govalidator.IsDNSName(ep.Target) {
		// SNI can only be a hostname, not IP
		sni = ep.Target
	}

	tlsContext, err := envoy_tls.UpstreamTlsContextOutsideMesh(
		c.SystemCaPath,
		ep.ExternalService.CaCert,
		ep.ExternalService.ClientCert,
		ep.ExternalService.ClientKey,
		ep.ExternalService.AllowRenegotiation,
		ep.ExternalService.SkipHostnameVerification,
		ep.ExternalService.FallbackToSystemCa,
		ep.Target,
		sni,
		ep.ExternalService.SANs,
		ep.ExternalService.MinTlsVersion,
		ep.ExternalService.MaxTlsVersion,
	)
	if err != nil {
		return nil, err
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

	tsm := envoy_cluster.Cluster_TransportSocketMatch{
		Name:            ep.Target,
		TransportSocket: transportSocket,
	}

	if withMatch {
		tsm.Match = &structpb.Struct{
			Fields: envoy_metadata.MetadataFields(tags.Tags(ep.Tags).WithoutTags(mesh_proto.ServiceTag)),
		}
	}

	return &tsm, nil
}
