package issuer

import (
	"crypto"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"math/big"
	"net/url"
	"time"

	"github.com/spiffe/go-spiffe/spiffe"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	util_tls "github.com/kumahq/kuma/pkg/tls"
	util_rsa "github.com/kumahq/kuma/pkg/util/rsa"
)

const (
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

func NewWorkloadCert(ca util_tls.KeyPair, mesh string, tags mesh_proto.MultiValueTagSet, certOpts ...CertOptsFn) (*util_tls.KeyPair, error) {
	caPrivateKey, caCert, err := loadKeyPair(ca)
	if err != nil {
		return nil, fmt.Errorf("failed to load CA key pair: %w", err)
	}

	workloadKey, err := util_rsa.GenerateKey(util_rsa.DefaultKeySize)
	if err != nil {
		return nil, fmt.Errorf("failed to generate a private key: %w", err)
	}
	template, err := newWorkloadTemplate(mesh, tags, workloadKey.Public(), certOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to generate X509 certificate template: %w", err)
	}
	workloadCert, err := x509.CreateCertificate(rand.Reader, template, caCert, workloadKey.Public(), caPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate X509 certificate: %w", err)
	}
	return util_tls.ToKeyPair(workloadKey, workloadCert)
}

func newWorkloadTemplate(trustDomain string, tags mesh_proto.MultiValueTagSet, publicKey crypto.PublicKey, certOpts ...CertOptsFn) (*x509.Certificate, error) {
	var uris []*url.URL
	for _, service := range tags.Values(mesh_proto.ServiceTag) {
		uri, err := spiffe.ParseID(fmt.Sprintf("spiffe://%s/%s", trustDomain, service), spiffe.AllowTrustDomainWorkload(trustDomain))
		if err != nil {
			return nil, err
		}
		uris = append(uris, uri)
	}
	for _, tag := range tags.Keys() {
		for _, value := range tags.UniqueValues(tag) {
			uri := fmt.Sprintf("kuma://%s/%s", tag, value)
			u, err := url.Parse(uri)
			if err != nil {
				return nil, fmt.Errorf("invalid Kuma URI: %w", err)
			}
			uris = append(uris, u)
		}
	}

	now := time.Now()
	serialNumber, err := newSerialNumber()
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

var maxUint128, one *big.Int

func init() {
	one = big.NewInt(1)
	m := new(big.Int)
	m.Lsh(one, 128)
	maxUint128 = m.Sub(m, one)
}

func newSerialNumber() (*big.Int, error) {
	res, err := rand.Int(rand.Reader, maxUint128)
	if err != nil {
		return nil, fmt.Errorf("failed generation of serial number: %w", err)
	}
	// Because we generate in the range [0, maxUint128) and 0 is an invalid serial and maxUint128 is valid we add 1
	// to have a number in range [1, maxUint128] See: https://cabforum.org/2016/03/31/ballot-164/
	return res.Add(res, one), nil
}

func loadKeyPair(pair util_tls.KeyPair) (crypto.PrivateKey, *x509.Certificate, error) {
	root, err := tls.X509KeyPair(pair.CertPEM, pair.KeyPEM)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse TLS key pair: %w", err)
	}
	rootCert, err := x509.ParseCertificate(root.Certificate[0])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse X509 certificate: %w", err)
	}
	return root.PrivateKey, rootCert, nil
}
