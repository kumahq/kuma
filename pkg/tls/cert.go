package tls

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net"
	"time"

	util_rsa "github.com/kumahq/kuma/pkg/util/rsa"
)

var DefaultValidityPeriod = 10 * 365 * 24 * time.Hour

type CertType string

const (
	ServerCertType CertType = "server"
	ClientCertType CertType = "client"
)

type KeyType func() (crypto.Signer, error)

var ECDSAKeyType KeyType = func() (crypto.Signer, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

var RSAKeyType KeyType = func() (crypto.Signer, error) {
	return util_rsa.GenerateKey(util_rsa.DefaultKeySize)
}

var DefaultKeyType = RSAKeyType

func NewSelfSignedCert(commonName string, certType CertType, keyType KeyType, hosts ...string) (KeyPair, error) {
	key, err := keyType()
	if err != nil {
		return KeyPair{}, fmt.Errorf("failed to generate TLS key: %w", err)
	}

	certBytes, err := generateCert(key, commonName, certType, hosts...)
	if err != nil {
		return KeyPair{}, err
	}

	keyBytes, err := pemEncodeKey(key)
	if err != nil {
		return KeyPair{}, err
	}

	return KeyPair{
		CertPEM: certBytes,
		KeyPEM:  keyBytes,
	}, nil
}

func generateCert(signer crypto.Signer, commonName string, certType CertType, hosts ...string) ([]byte, error) {
	csr, err := newCert(commonName, certType, hosts...)
	if err != nil {
		return nil, err
	}
	certDerBytes, err := x509.CreateCertificate(rand.Reader, &csr, &csr, signer.Public(), signer)
	if err != nil {
		return nil, fmt.Errorf("failed to generate TLS certificate: %w", err)
	}
	return pemEncodeCert(certDerBytes)
}

func newCert(commonName string, certType CertType, hosts ...string) (x509.Certificate, error) {
	notBefore := time.Now()
	notAfter := notBefore.Add(DefaultValidityPeriod)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return x509.Certificate{}, fmt.Errorf("failed to generate serial number: %w", err)
	}
	csr := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: commonName,
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{},
		BasicConstraintsValid: true,
	}
	switch certType {
	case ServerCertType:
		csr.ExtKeyUsage = append(csr.ExtKeyUsage, x509.ExtKeyUsageServerAuth)
	case ClientCertType:
		csr.ExtKeyUsage = append(csr.ExtKeyUsage, x509.ExtKeyUsageClientAuth)
	default:
		return x509.Certificate{}, fmt.Errorf("invalid certificate type %q, expected either %q or %q",
			certType, ServerCertType, ClientCertType)
	}
	for _, host := range hosts {
		if ip := net.ParseIP(host); ip != nil {
			csr.IPAddresses = append(csr.IPAddresses, ip)
		} else {
			csr.DNSNames = append(csr.DNSNames, host)
		}
	}
	return csr, nil
}
