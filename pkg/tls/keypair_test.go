package tls_test

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	util_tls "github.com/kumahq/kuma/v3/pkg/tls"
)

func selfSigned(signer crypto.Signer) []byte {
	GinkgoHelper()
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "test"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, signer.Public(), signer)
	Expect(err).ToNot(HaveOccurred())
	return der
}

var _ = Describe("ToKeyPair", func() {
	// The encoding side used to support fewer key types than ParsePrivateKey accepted,
	// which let a CA be loaded and signed with but not written back out.
	DescribeTable("should round-trip every key type ParsePrivateKey accepts",
		func(newSigner func() crypto.Signer, expectedBlockType string, expectedKey any) {
			signer := newSigner()

			// when
			pair, err := util_tls.ToKeyPair(signer, selfSigned(signer))

			// then the PEM is well formed
			Expect(err).ToNot(HaveOccurred())
			certBlock, rest := pem.Decode(pair.CertPEM)
			Expect(certBlock).ToNot(BeNil())
			Expect(certBlock.Type).To(Equal("CERTIFICATE"))
			Expect(rest).To(BeEmpty())

			keyBlock, rest := pem.Decode(pair.KeyPEM)
			Expect(keyBlock).ToNot(BeNil())
			Expect(keyBlock.Type).To(Equal(expectedBlockType))
			Expect(rest).To(BeEmpty())

			// and it parses back to the same key type
			parsed, err := util_tls.ParsePrivateKey(keyBlock.Bytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(parsed).To(BeAssignableToTypeOf(expectedKey))

			// and the pair loads as a TLS certificate
			loaded, err := tls.X509KeyPair(pair.CertPEM, pair.KeyPEM)
			Expect(err).ToNot(HaveOccurred())
			Expect(loaded.Certificate).To(HaveLen(1))
		},
		Entry("RSA", func() crypto.Signer {
			key, err := rsa.GenerateKey(rand.Reader, 2048)
			Expect(err).ToNot(HaveOccurred())
			return key
		}, "RSA PRIVATE KEY", &rsa.PrivateKey{}),
		Entry("ECDSA P-256", func() crypto.Signer {
			key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			Expect(err).ToNot(HaveOccurred())
			return key
		}, "EC PRIVATE KEY", &ecdsa.PrivateKey{}),
		Entry("ECDSA P-384", func() crypto.Signer {
			key, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
			Expect(err).ToNot(HaveOccurred())
			return key
		}, "EC PRIVATE KEY", &ecdsa.PrivateKey{}),
		// ED25519 has no SEC1/PKCS1 encoding, so it has to go out as PKCS8
		Entry("ED25519", func() crypto.Signer {
			_, key, err := ed25519.GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())
			return key
		}, "PRIVATE KEY", ed25519.PrivateKey{}),
	)

	It("should fail on a key type it can't encode", func() {
		type unsupportedKey struct{}

		// when
		_, err := util_tls.ToKeyPair(unsupportedKey{}, []byte{})

		// then
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unsupported private key type"))
	})
})

var _ = Describe("ParsePrivateKey", func() {
	It("should fail on data that isn't a private key", func() {
		// when
		_, err := util_tls.ParsePrivateKey([]byte("not a key"))

		// then
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to parse private key"))
	})

	It("should parse a PKCS8 encoded RSA key", func() {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		Expect(err).ToNot(HaveOccurred())
		der, err := x509.MarshalPKCS8PrivateKey(key)
		Expect(err).ToNot(HaveOccurred())

		// when
		parsed, err := util_tls.ParsePrivateKey(der)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(parsed).To(BeAssignableToTypeOf(&rsa.PrivateKey{}))
	})
})
