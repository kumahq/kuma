package v3

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	envoy_type_matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/tls"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy/names"
	xds_tls "github.com/kumahq/kuma/pkg/xds/envoy/tls"
)

// CreateDownstreamTlsContext creates DownstreamTlsContext for incoming connections
// It verifies that incoming connection has TLS certificate signed by Mesh CA with URI SAN of prefix spiffe://{mesh_name}/
// It secures inbound listener with certificate of "identity_cert" that will be received from the SDS (it contains URI SANs of all inbounds).
func CreateDownstreamTlsContext(mesh *core_mesh.MeshResource) (*envoy_tls.DownstreamTlsContext, error) {
	if !mesh.MTLSEnabled() {
		return nil, nil
	}
	validationSANMatcher := MeshSpiffeIDPrefixMatcher(mesh.Meta.GetName())
	commonTlsContext, err := createCommonTlsContext(mesh, validationSANMatcher)
	if err != nil {
		return nil, err
	}
	return &envoy_tls.DownstreamTlsContext{
		CommonTlsContext:         commonTlsContext,
		RequireClientCertificate: util_proto.Bool(true),
	}, nil
}

// CreateUpstreamTlsContext creates UpstreamTlsContext for outgoing connections
// It verifies that the upstream server has TLS certificate signed by Mesh CA with URI SAN of spiffe://{mesh_name}/{upstream_service}
// The downstream client exposes for the upstream server cert with multiple URI SANs, which means that if DP has inbound with services "web" and "web-api" and communicates with "backend"
// the upstream server ("backend") will see that DP with TLS certificate of URIs of "web" and "web-api".
// There is no way to correlate incoming request to "web" or "web-api" with outgoing request to "backend" to expose only one URI SAN.
//
// Pass "*" for upstreamService to validate that upstream service is a service that is part of the mesh (but not specific one)
func CreateUpstreamTlsContext(mesh *core_mesh.MeshResource, upstreamService string, sni string) (*envoy_tls.UpstreamTlsContext, error) {
	if !mesh.MTLSEnabled() {
		return nil, nil
	}
	var validationSANMatcher *envoy_type_matcher.StringMatcher
	meshName := mesh.Meta.GetName()
	if upstreamService == "*" {
		validationSANMatcher = MeshSpiffeIDPrefixMatcher(meshName)
	} else {
		validationSANMatcher = ServiceSpiffeIDMatcher(meshName, upstreamService)
	}
	commonTlsContext, err := createCommonTlsContext(mesh, validationSANMatcher)
	if err != nil {
		return nil, err
	}
	commonTlsContext.AlpnProtocols = xds_tls.KumaALPNProtocols
	return &envoy_tls.UpstreamTlsContext{
		CommonTlsContext: commonTlsContext,
		Sni:              sni,
	}, nil
}

func createCommonTlsContext(mesh *core_mesh.MeshResource, validationSANMatcher *envoy_type_matcher.StringMatcher) (*envoy_tls.CommonTlsContext, error) {
	meshName := mesh.GetMeta().GetName()
	meshCaSecret := NewSecretConfigSource(names.GetSecretName(xds_tls.MeshCaResource, "secret", meshName))
	identitySecret := NewSecretConfigSource(names.GetSecretName(xds_tls.IdentityCertResource, "secret", meshName))

	return &envoy_tls.CommonTlsContext{
		ValidationContextType: &envoy_tls.CommonTlsContext_CombinedValidationContext{
			CombinedValidationContext: &envoy_tls.CommonTlsContext_CombinedCertificateValidationContext{
				DefaultValidationContext: &envoy_tls.CertificateValidationContext{
					MatchSubjectAltNames: []*envoy_type_matcher.StringMatcher{validationSANMatcher},
				},
				ValidationContextSdsSecretConfig: meshCaSecret,
			},
		},
		TlsCertificateSdsSecretConfigs: []*envoy_tls.SdsSecretConfig{
			identitySecret,
		},
	}, nil
}

func NewSecretConfigSource(secretName string) *envoy_tls.SdsSecretConfig {
	return &envoy_tls.SdsSecretConfig{
		Name: secretName,
		SdsConfig: &envoy_core.ConfigSource{
			ResourceApiVersion:    envoy_core.ApiVersion_V3,
			ConfigSourceSpecifier: &envoy_core.ConfigSource_Ads{},
		},
	}
}

func UpstreamTlsContextOutsideMesh(ca, cert, key []byte, allowRenegotiation bool, hostname string, sni string) (*envoy_tls.UpstreamTlsContext, error) {
	tlsContext := &envoy_tls.UpstreamTlsContext{
		AllowRenegotiation: allowRenegotiation,
		Sni:                sni,
	}
	if cert != nil && key != nil {
		tlsContext.CommonTlsContext = &envoy_tls.CommonTlsContext{
			TlsCertificates: []*envoy_tls.TlsCertificate{
				{
					CertificateChain: dataSourceFromBytes(cert),
					PrivateKey:       dataSourceFromBytes(key),
				},
			},
		}
	}

	if ca != nil {
		if tlsContext.CommonTlsContext == nil {
			tlsContext.CommonTlsContext = &envoy_tls.CommonTlsContext{}
		}
		tlsContext.CommonTlsContext.ValidationContextType = &envoy_tls.CommonTlsContext_ValidationContext{
			ValidationContext: &envoy_tls.CertificateValidationContext{
				TrustedCa: dataSourceFromBytes(ca),
				MatchSubjectAltNames: []*envoy_type_matcher.StringMatcher{
					{
						MatchPattern: &envoy_type_matcher.StringMatcher_Exact{
							Exact: hostname,
						},
					},
				},
			},
		}
	}
	return tlsContext, nil
}

func dataSourceFromBytes(bytes []byte) *envoy_core.DataSource {
	return &envoy_core.DataSource{
		Specifier: &envoy_core.DataSource_InlineBytes{
			InlineBytes: bytes,
		},
	}
}

func MeshSpiffeIDPrefixMatcher(mesh string) *envoy_type_matcher.StringMatcher {
	return &envoy_type_matcher.StringMatcher{
		MatchPattern: &envoy_type_matcher.StringMatcher_Prefix{
			Prefix: xds_tls.MeshSpiffeIDPrefix(mesh),
		},
	}
}

func ServiceSpiffeIDMatcher(mesh string, service string) *envoy_type_matcher.StringMatcher {
	return &envoy_type_matcher.StringMatcher{
		MatchPattern: &envoy_type_matcher.StringMatcher_Exact{
			Exact: xds_tls.ServiceSpiffeID(mesh, service),
		},
	}
}

func KumaIDMatcher(tagName, tagValue string) *envoy_type_matcher.StringMatcher {
	return &envoy_type_matcher.StringMatcher{
		MatchPattern: &envoy_type_matcher.StringMatcher_Exact{
			Exact: xds_tls.KumaID(tagName, tagValue),
		},
	}
}

func StaticDownstreamTlsContext(keyPair *tls.KeyPair) *envoy_tls.DownstreamTlsContext {
	return &envoy_tls.DownstreamTlsContext{
		CommonTlsContext: &envoy_tls.CommonTlsContext{
			TlsCertificates: []*envoy_tls.TlsCertificate{
				{
					CertificateChain: &envoy_core.DataSource{
						Specifier: &envoy_core.DataSource_InlineBytes{
							InlineBytes: keyPair.CertPEM,
						},
					},
					PrivateKey: &envoy_core.DataSource{
						Specifier: &envoy_core.DataSource_InlineBytes{
							InlineBytes: keyPair.KeyPEM,
						},
					},
				},
			},
		},
	}
}
