package ca

import (
	"bytes"

	envoy_auth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"

	sds_provider "github.com/Kong/kuma/pkg/sds/provider"
)

type MeshCaSecret struct {
	PemCerts [][]byte
}

var _ sds_provider.Secret = &MeshCaSecret{}

func (s *MeshCaSecret) ToResource(name string) *envoy_auth.Secret {
	return &envoy_auth.Secret{
		Name: name,
		Type: &envoy_auth.Secret_ValidationContext{
			ValidationContext: &envoy_auth.CertificateValidationContext{
				TrustedCa: &envoy_core.DataSource{
					Specifier: &envoy_core.DataSource_InlineBytes{
						InlineBytes: bytes.Join(s.PemCerts, []byte("\n")),
					},
				},
			},
		},
	}
}
