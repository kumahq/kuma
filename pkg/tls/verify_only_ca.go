package tls

import (
	"crypto/x509"
	"time"

	"github.com/pkg/errors"
)

func VerifyOnlyCA(caPool *x509.CertPool) func([][]byte, [][]*x509.Certificate) error {
	return func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
		// Code copy/pasted and adapted from
		// https://github.com/golang/go/blob/81555cb4f3521b53f9de4ce15f64b77cc9df61b9/src/crypto/tls/handshake_client.go#L327-L344, but adapted to skip the hostname verification.
		// See https://github.com/golang/go/issues/21971#issuecomment-412836078.

		// If this is the first handshake on a connection, process and
		// (optionally) verify the server's certificates.
		certs := make([]*x509.Certificate, len(rawCerts))
		for i, asn1Data := range rawCerts {
			cert, err := x509.ParseCertificate(asn1Data)
			if err != nil {
				return errors.Wrap(err, "tls: failed to parse certificate from server")
			}
			certs[i] = cert
		}

		opts := x509.VerifyOptions{
			Roots:         caPool,
			CurrentTime:   time.Now(),
			DNSName:       "", // <- skip hostname verification
			Intermediates: x509.NewCertPool(),
		}

		// Skip the first cert because it's the leaf. All others are intermediates.
		for _, cert := range certs[:1] {
			opts.Intermediates.AddCert(cert)
		}
		_, err := certs[0].Verify(opts)
		return err
	}
}
