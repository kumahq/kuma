package v3

import (
	"bytes"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_auth "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
)

func CreateCaSecret(secret *core_xds.CaSecret, name string) *envoy_auth.Secret {
	return &envoy_auth.Secret{
		Name: name,
		Type: &envoy_auth.Secret_ValidationContext{
			ValidationContext: &envoy_auth.CertificateValidationContext{
				TrustedCa: &envoy_core.DataSource{
					Specifier: &envoy_core.DataSource_InlineBytes{
						InlineBytes: bytes.Join(secret.PemCerts, []byte("\n")),
					},
				},
			},
		},
	}
}
