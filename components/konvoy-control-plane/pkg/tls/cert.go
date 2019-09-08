package tls

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"time"

	"github.com/pkg/errors"
)

var (
	DefaultRsaBits        = 2048
	DefaultValidityPeriod = 10 * 365 * 24 * time.Hour
)

type KeyPair struct {
	CertPEM []byte
	KeyPEM  []byte
}

func NewSelfSignedCert(commonName string, hosts ...string) (KeyPair, error) {
	key, err := rsa.GenerateKey(rand.Reader, DefaultRsaBits)
	if err != nil {
		return KeyPair{}, errors.Wrap(err, "failed to generate TLS key")
	}

	certBytes, err := generateCert(key, commonName, hosts...)
	if err != nil {
		return KeyPair{}, err
	}

	keyBytes, err := marshalKey(key)
	if err != nil {
		return KeyPair{}, err
	}

	return KeyPair{
		CertPEM: certBytes,
		KeyPEM:  keyBytes,
	}, nil
}

func generateCert(signer crypto.Signer, commonName string, hosts ...string) ([]byte, error) {
	csr, err := newCert(commonName, hosts...)
	if err != nil {
		return nil, err
	}
	certDerBytes, err := x509.CreateCertificate(rand.Reader, &csr, &csr, signer.Public(), signer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate TLS certificate")
	}
	var certBuf bytes.Buffer
	if err := pem.Encode(&certBuf, &pem.Block{Type: "CERTIFICATE", Bytes: certDerBytes}); err != nil {
		return nil, errors.Wrap(err, "failed to PEM encode TLS certificate")
	}
	return certBuf.Bytes(), nil
}

func newCert(commonName string, hosts ...string) (x509.Certificate, error) {
	notBefore := time.Now()
	notAfter := notBefore.Add(DefaultValidityPeriod)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return x509.Certificate{}, errors.Wrap(err, "failed to generate serial number")
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
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
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

func marshalKey(priv interface{}) ([]byte, error) {
	var block *pem.Block
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		block = &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}
	default:
		return nil, errors.Errorf("unsupported private key type %T", priv)
	}
	var keyBuf bytes.Buffer
	if err := pem.Encode(&keyBuf, block); err != nil {
		return nil, err
	}
	return keyBuf.Bytes(), nil
}
