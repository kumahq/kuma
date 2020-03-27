package provider

import (
	"bytes"

	envoy_auth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
)

type MeshCaSecret struct {
	PemCerts [][]byte
}

var _ Secret = &MeshCaSecret{}

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

type IdentityCertSecret struct {
	PemCerts [][]byte
	PemKey   []byte
}

var _ Secret = &IdentityCertSecret{}

func (s *IdentityCertSecret) ToResource(name string) *envoy_auth.Secret {
	return &envoy_auth.Secret{
		Name: name,
		Type: &envoy_auth.Secret_TlsCertificate{
			TlsCertificate: &envoy_auth.TlsCertificate{
				CertificateChain: &envoy_core.DataSource{
					Specifier: &envoy_core.DataSource_InlineBytes{
						InlineBytes: bytes.Join(s.PemCerts, []byte("\n")),
					},
				},
				PrivateKey: &envoy_core.DataSource{
					Specifier: &envoy_core.DataSource_InlineBytes{
						InlineBytes: s.PemKey,
					},
				},
			},
		},
	}
}
