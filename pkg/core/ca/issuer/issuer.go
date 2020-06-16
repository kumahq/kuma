package issuer

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/url"
	"time"

	"github.com/pkg/errors"
	"github.com/spiffe/go-spiffe/spiffe"
	"github.com/spiffe/spire/pkg/common/x509util"

	"github.com/Kong/kuma/pkg/core"
	util_tls "github.com/Kong/kuma/pkg/tls"
)

const (
	DefaultRsaBits                    = 2048
	DefaultAllowedClockSkew           = 10 * time.Second
	DefaultWorkloadCertValidityPeriod = 24 * time.Hour
)

type CertOptsFn = func(*x509.Certificate)

func WithExpirationTime(expiration time.Duration) CertOptsFn {
	return func(certificate *x509.Certificate) {
		now := core.Now()
		certificate.NotAfter = now.Add(expiration)
	}
}

func NewWorkloadCert(ca util_tls.KeyPair, mesh string, services []string, certOpts ...CertOptsFn) (*util_tls.KeyPair, error) {
	caPrivateKey, caCert, err := loadKeyPair(ca)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load CA key pair")
	}

	workloadKey, err := rsa.GenerateKey(rand.Reader, DefaultRsaBits)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate a private key")
	}
	template, err := newWorkloadTemplate(mesh, services, workloadKey.Public(), certOpts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate X509 certificate template")
	}
	workloadCert, err := x509.CreateCertificate(rand.Reader, template, caCert, workloadKey.Public(), caPrivateKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate X509 certificate")
	}
	return util_tls.ToKeyPair(workloadKey, workloadCert)
}

func newWorkloadTemplate(trustDomain string, services []string, publicKey crypto.PublicKey, certOpts ...CertOptsFn) (*x509.Certificate, error) {
	var uris []*url.URL
	for _, service := range services {
		uri, err := spiffe.ParseID(fmt.Sprintf("spiffe://%s/%s", trustDomain, service), spiffe.AllowTrustDomainWorkload(trustDomain))
		if err != nil {
			return nil, err
		}
		uris = append(uris, uri)
	}

	now := time.Now()
	serialNumber, err := x509util.NewSerialNumber()
	if err != nil {
		return nil, err
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		// Subject is deliberately left empty
		URIs:      uris,
		NotBefore: now.Add(-DefaultAllowedClockSkew),
		NotAfter:  now.Add(DefaultWorkloadCertValidityPeriod),
		KeyUsage: x509.KeyUsageKeyEncipherment |
			x509.KeyUsageKeyAgreement |
			x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
		BasicConstraintsValid: true,
		PublicKey:             publicKey,
	}

	for _, opt := range certOpts {
		opt(template)
	}
	return template, nil
}

func loadKeyPair(pair util_tls.KeyPair) (crypto.PrivateKey, *x509.Certificate, error) {
	root, err := tls.X509KeyPair(pair.CertPEM, pair.KeyPEM)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to parse TLS key pair")
	}
	rootCert, err := x509.ParseCertificate(root.Certificate[0])
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to parse X509 certificate")
	}
	return root.PrivateKey, rootCert, nil
}
