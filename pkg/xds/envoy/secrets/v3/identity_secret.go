package v2

import (
	"bytes"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_auth "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/xds/envoy/tls"
)

func CreateIdentitySecret(secret *core_xds.IdentitySecret) *envoy_auth.Secret {
	return &envoy_auth.Secret{
		Name: tls.IdentityCertResource,
		Type: &envoy_auth.Secret_TlsCertificate{
			TlsCertificate: &envoy_auth.TlsCertificate{
				CertificateChain: &envoy_core.DataSource{
					Specifier: &envoy_core.DataSource_InlineBytes{
						InlineBytes: bytes.Join(secret.PemCerts, []byte("\n")),
					},
				},
				PrivateKey: &envoy_core.DataSource{
					Specifier: &envoy_core.DataSource_InlineBytes{
						InlineBytes: secret.PemKey,
					},
				},
			},
		},
	}
}
