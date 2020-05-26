package issuer

import (
	"crypto"
	"crypto/x509"
	"math/big"
	"net/url"
	"time"

	"github.com/spiffe/go-spiffe/spiffe"
)

func NewWorkloadTemplate(spiffeIDs []string, trustDomain string, publicKey crypto.PublicKey, notBefore, notAfter time.Time, serialNumber *big.Int) (*x509.Certificate, error) {
	var uris []*url.URL
	for _, spiffeID := range spiffeIDs {
		uri, err := spiffe.ParseID(spiffeID, spiffe.AllowTrustDomainWorkload(trustDomain))
		if err != nil {
			return nil, err
		}
		uris = append(uris, uri)
	}
	return &x509.Certificate{
		SerialNumber: serialNumber,
		// Subject is deliberately left empty
		URIs:      uris,
		NotBefore: notBefore,
		NotAfter:  notAfter,
		KeyUsage: x509.KeyUsageKeyEncipherment |
			x509.KeyUsageKeyAgreement |
			x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
		BasicConstraintsValid: true,
		PublicKey:             publicKey,
	}, nil
}
