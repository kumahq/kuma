package v2

import (
	"bytes"

	envoy_auth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/xds/envoy/tls"
)

func CreateCaSecret(secret *core_xds.CaSecret) *envoy_auth.Secret {
	return &envoy_auth.Secret{
		Name: tls.MeshCaResource,
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
