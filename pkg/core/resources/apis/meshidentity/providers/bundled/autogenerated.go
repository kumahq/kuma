package bundled

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net/url"
	"time"

	"github.com/pkg/errors"
	"github.com/spiffe/go-spiffe/v2/spiffeid"

	"github.com/kumahq/kuma/pkg/core"
	core_ca "github.com/kumahq/kuma/pkg/core/ca"
	util_tls "github.com/kumahq/kuma/pkg/tls"
)

const (
	defaultCACertValidityPeriod = 10 * 365 * 24 * time.Hour
	defaultKeySize              = 2048
)

func RootCAName(resourceName string) string {
	return fmt.Sprintf("%s-root-ca", resourceName)
}

func PrivateKeyName(resourceName string) string {
	return fmt.Sprintf("%s-private-key", resourceName)
}

// We are using RSA since Envoy not fully works with ED25519 or ecliptic P-384
func GenerateRootCA(trustDomain string) (*core_ca.KeyPair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, defaultKeySize)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate ecdsa key")
	}
	cert, err := newCACert(privateKey, trustDomain)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate X509 certificate")
	}
	return util_tls.ToKeyPair(privateKey, cert)
}

func newCACert(signer crypto.Signer, trustDomain string) ([]byte, error) {
	subject := pkix.Name{
		Organization:       []string{"Kuma"},
		OrganizationalUnit: []string{"Mesh"},
		CommonName:         trustDomain,
	}
	now := core.Now()
	notBefore := now.Add(-DefaultAllowedClockSkew)
	notAfter := now.Add(defaultCACertValidityPeriod)

	domain, err := spiffeid.TrustDomainFromString(trustDomain)
	if err != nil {
		return nil, err
	}
	uri, err := spiffeid.FromSegments(domain)
	if err != nil {
		return nil, err
	}
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, err
	}
	tmpl := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      subject,
		URIs:         []*url.URL{uri.URL()},
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		KeyUsage: x509.KeyUsageCertSign |
			x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		PublicKey:             signer.Public(),
	}
	return x509.CreateCertificate(rand.Reader, tmpl, tmpl, signer.Public(), signer)
}
