package builtin

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net/url"
	"time"

	"github.com/pkg/errors"
	"github.com/spiffe/go-spiffe/spiffe"

	core_ca "github.com/Kong/kuma/pkg/core/ca"
	util_tls "github.com/Kong/kuma/pkg/tls"
)

const (
	DefaultRsaBits              = 2048
	DefaultAllowedClockSkew     = 10 * time.Second
	DefaultCACertValidityPeriod = 10 * 365 * 24 * time.Hour
)

func newRootCa(mesh string) (*core_ca.KeyPair, error) {
	key, err := rsa.GenerateKey(rand.Reader, DefaultRsaBits)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate a private key")
	}
	cert, err := newCACert(key, mesh)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate X509 certificate")
	}
	return util_tls.ToKeyPair(key, cert)
}

func newCACert(signer crypto.Signer, trustDomain string) ([]byte, error) {
	spiffeID := &url.URL{
		Scheme: "spiffe",
		Host:   trustDomain,
	}
	subject := pkix.Name{
		Organization:       []string{"Kuma"},
		OrganizationalUnit: []string{"Mesh"},
		CommonName:         trustDomain,
	}
	now := time.Now()
	notBefore := now.Add(-DefaultAllowedClockSkew)
	notAfter := now.Add(DefaultCACertValidityPeriod)

	template, err := caTemplate(spiffeID.String(), trustDomain, subject, signer.Public(), notBefore, notAfter, big.NewInt(0))
	if err != nil {
		return nil, err
	}
	return x509.CreateCertificate(rand.Reader, template, template, signer.Public(), signer)
}

func caTemplate(spiffeID string, trustDomain string, subject pkix.Name, publicKey crypto.PublicKey, notBefore, notAfter time.Time, serialNumber *big.Int) (*x509.Certificate, error) {
	uri, err := spiffe.ParseID(spiffeID, spiffe.AllowTrustDomain(trustDomain))
	if err != nil {
		return nil, err
	}
	return &x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      subject,
		URIs:         []*url.URL{uri},
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		KeyUsage: x509.KeyUsageCertSign |
			x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		PublicKey:             publicKey,
	}, nil
}
