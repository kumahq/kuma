package issuer

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net/url"
	"time"

	"github.com/pkg/errors"
	"github.com/spiffe/spire/pkg/common/x509util"

	util_tls "github.com/Kong/kuma/pkg/tls"
)

const (
	DefaultRsaBits                    = 2048
	DefaultAllowedClockSkew           = 10 * time.Second
	DefaultCACertValidityPeriod       = 10 * 365 * 24 * time.Hour
	DefaultWorkloadCertValidityPeriod = 90 * 24 * time.Hour
)

func NewRootCA(mesh string) (*util_tls.KeyPair, error) {
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

	template, err := NewCATemplate(spiffeID.String(), trustDomain, subject, signer.Public(), notBefore, notAfter, big.NewInt(0))
	if err != nil {
		return nil, err
	}
	return x509.CreateCertificate(rand.Reader, template, template, signer.Public(), signer)
}

func NewWorkloadCert(ca util_tls.KeyPair, mesh string, workload string) (*util_tls.KeyPair, error) {
	caPrivateKey, caCert, err := loadKeyPair(ca)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load CA key pair")
	}

	workloadKey, err := rsa.GenerateKey(rand.Reader, DefaultRsaBits)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate a private key")
	}
	workloadCert, err := newWorkloadCert(caPrivateKey, caCert, mesh, workload, workloadKey.Public())
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate X509 certificate")
	}
	return util_tls.ToKeyPair(workloadKey, workloadCert)
}

func newWorkloadCert(signer crypto.PrivateKey, parent *x509.Certificate, trustDomain string, workload string, publicKey crypto.PublicKey) ([]byte, error) {
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
