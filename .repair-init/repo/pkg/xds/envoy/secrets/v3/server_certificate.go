package v3

import (
	"bytes"
	"encoding/pem"
	"io"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_auth "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
)

// NewServerCertificateSecret populates a new Envoy TLS certificate
// secret containing the given key and chain of certificates.
func NewServerCertificateSecret(key *pem.Block, certificates []*pem.Block) *envoy_auth.Secret {
	mustEncode := func(out io.Writer, b *pem.Block) {
		if err := pem.Encode(out, b); err != nil {
			panic(err.Error())
		}
	}

	keyBytes := &bytes.Buffer{}
	certificateBytes := &bytes.Buffer{}

	mustEncode(keyBytes, key)

	for _, c := range certificates {
		mustEncode(certificateBytes, c)
		certificateBytes.WriteString("\n")
	}

	return &envoy_auth.Secret{
		Type: &envoy_auth.Secret_TlsCertificate{
			TlsCertificate: &envoy_auth.TlsCertificate{
				CertificateChain: &envoy_core.DataSource{
					Specifier: &envoy_core.DataSource_InlineString{
						InlineString: certificateBytes.String(),
					},
				},
				PrivateKey: &envoy_core.DataSource{
					Specifier: &envoy_core.DataSource_InlineString{
						InlineString: keyBytes.String(),
					},
				},
			},
		},
	}
}
