package clusters

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_wellknown "github.com/envoyproxy/go-control-plane/pkg/wellknown"
	pstruct "github.com/golang/protobuf/ptypes/struct"

	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
)

func ClientSideTLS(endpoints []xds.Endpoint) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.Add(&clientSideTLSConfigurer{
			endpoints: endpoints,
		})
	})
}

type clientSideTLSConfigurer struct {
	endpoints []xds.Endpoint
}

func (c *clientSideTLSConfigurer) Configure(cluster *envoy_api.Cluster) error {
	for _, ep := range c.endpoints {
		if ep.ExternalService.TLSEnabled {
			ca, cert, key := externalServiceTlsCerts(ep.ExternalService)
			tlsContext, err := envoy.CreateUpstreamTlsContextNoMetadata(ca, cert, key, ep.Target)
			if err != nil {
				return err
			}

			pbst, err := proto.MarshalAnyDeterministic(tlsContext)
			if err != nil {
				return err
			}

			transportSocket := &envoy_core.TransportSocket{
				Name: envoy_wellknown.TransportSocketTls,
				ConfigType: &envoy_core.TransportSocket_TypedConfig{
					TypedConfig: pbst,
				},
			}

			cluster.TransportSocketMatches = append(cluster.TransportSocketMatches, &envoy_api.Cluster_TransportSocketMatch{
				Name: ep.Target,
				Match: &pstruct.Struct{
					Fields: envoy.MetadataFields(ep.Tags),
				},
				TransportSocket: transportSocket,
			})
		}
	}

	return nil
}

func externalServiceTlsCerts(es *xds.ExternalService) (ca, cert, key *envoy_core.DataSource) {
	if es.CaCert != nil {
		ca = &envoy_core.DataSource{
			Specifier: &envoy_core.DataSource_InlineBytes{
				InlineBytes: es.CaCert,
			},
		}
	}
	if es.ClientCert != nil {
		cert = &envoy_core.DataSource{
			Specifier: &envoy_core.DataSource_InlineBytes{
				InlineBytes: es.ClientCert,
			},
		}
	}
	if es.ClientKey != nil {
		key = &envoy_core.DataSource{
			Specifier: &envoy_core.DataSource_InlineBytes{
				InlineBytes: es.ClientKey,
			},
		}
	}
	return
}
