package issuer

import (
	"crypto"
	"crypto/x509"
	"math/big"
	"net/url"
	"time"

	"github.com/spiffe/go-spiffe/spiffe"
)

func NewWorkloadTemplate(spiffeID string, trustDomain string, publicKey crypto.PublicKey, notBefore, notAfter time.Time, serialNumber *big.Int) (*x509.Certificate, error) {
	uri, err := spiffe.ParseID(spiffeID, spiffe.AllowTrustDomainWorkload(trustDomain))
	if err != nil {
		return nil, err
	}
	return &x509.Certificate{
		SerialNumber: serialNumber,
		// Subject is deliberately left empty
		URIs:      []*url.URL{uri},
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
