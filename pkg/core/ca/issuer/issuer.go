package issuer

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"net/url"
	"time"

	"github.com/Kong/kuma/pkg/core"

	"github.com/pkg/errors"
	"github.com/spiffe/spire/pkg/common/x509util"

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

func NewWorkloadCert(ca util_tls.KeyPair, mesh string, workload string, certOpts ...CertOptsFn) (*util_tls.KeyPair, error) {
	caPrivateKey, caCert, err := loadKeyPair(ca)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load CA key pair")
	}

	workloadKey, err := rsa.GenerateKey(rand.Reader, DefaultRsaBits)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate a private key")
	}
	workloadCert, err := newWorkloadCert(caPrivateKey, caCert, mesh, workload, workloadKey.Public(), certOpts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate X509 certificate")
	}
	return util_tls.ToKeyPair(workloadKey, workloadCert)
}

func newWorkloadCert(signer crypto.PrivateKey, parent *x509.Certificate, trustDomain string, workload string, publicKey crypto.PublicKey, certOpts ...CertOptsFn) ([]byte, error) {
	spiffeID := &url.URL{
		Scheme: "spiffe",
		Host:   trustDomain,
		Path:   workload,
	}

	now := time.Now()
	notBefore := now.Add(-DefaultAllowedClockSkew)
	notAfter := now.Add(DefaultWorkloadCertValidityPeriod)

	serialNumber, err := x509util.NewSerialNumber()
	if err != nil {
		return nil, err
	}

	template, err := NewWorkloadTemplate(spiffeID.String(), trustDomain, publicKey, notBefore, notAfter, serialNumber)
	if err != nil {
		return nil, err
	}

	for _, opt := range certOpts {
		opt(template)
	}

	return x509.CreateCertificate(rand.Reader, template, parent, publicKey, signer)
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
