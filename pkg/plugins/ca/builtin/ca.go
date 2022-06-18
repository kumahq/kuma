package builtin

import (
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net/url"
	"time"

	"github.com/spiffe/go-spiffe/spiffe"

	"github.com/kumahq/kuma/pkg/core"
	core_ca "github.com/kumahq/kuma/pkg/core/ca"
	util_tls "github.com/kumahq/kuma/pkg/tls"
	util_rsa "github.com/kumahq/kuma/pkg/util/rsa"
)

const (
	DefaultAllowedClockSkew     = 10 * time.Second
	DefaultCACertValidityPeriod = 10 * 365 * 24 * time.Hour
)

type certOptsFn = func(*x509.Certificate)

func withExpirationTime(expiration time.Duration) certOptsFn {
	return func(certificate *x509.Certificate) {
		now := core.Now()
		certificate.NotAfter = now.Add(expiration)
	}
}

func newRootCa(mesh string, rsaBits int, certOpts ...certOptsFn) (*core_ca.KeyPair, error) {
	if rsaBits == 0 {
		rsaBits = util_rsa.DefaultKeySize
	}
	key, err := util_rsa.GenerateKey(rsaBits)
	if err != nil {
		return nil, fmt.Errorf("failed to generate a private key: %w", err)
	}
	cert, err := newCACert(key, mesh, certOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to generate X509 certificate: %w", err)
	}
	return util_tls.ToKeyPair(key, cert)
}

func newCACert(signer crypto.Signer, trustDomain string, certOpts ...certOptsFn) ([]byte, error) {
	spiffeID := &url.URL{
		Scheme: "spiffe",
		Host:   trustDomain,
	}
	subject := pkix.Name{
		Organization:       []string{"Kuma"},
		OrganizationalUnit: []string{"Mesh"},
		CommonName:         trustDomain,
	}
	now := core.Now()
	notBefore := now.Add(-DefaultAllowedClockSkew)
	notAfter := now.Add(DefaultCACertValidityPeriod)

	template, err := caTemplate(spiffeID.String(), trustDomain, subject, signer.Public(), notBefore, notAfter, big.NewInt(0))
	if err != nil {
		return nil, err
	}

	for _, opt := range certOpts {
		opt(template)
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
