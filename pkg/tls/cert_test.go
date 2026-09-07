package tls_test

import (
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	util_tls "github.com/kumahq/kuma/v3/pkg/tls"
)

func parseCert(pair util_tls.KeyPair) *x509.Certificate {
	GinkgoHelper()
	block, _ := pem.Decode(pair.CertPEM)
	Expect(block).ToNot(BeNil())
	cert, err := x509.ParseCertificate(block.Bytes)
	Expect(err).ToNot(HaveOccurred())
	return cert
}

func parseSigner(pair util_tls.KeyPair) crypto.Signer {
	GinkgoHelper()
	block, _ := pem.Decode(pair.KeyPEM)
	Expect(block).ToNot(BeNil())
	key, err := util_tls.ParsePrivateKey(block.Bytes)
	Expect(err).ToNot(HaveOccurred())
	signer, ok := key.(crypto.Signer)
	Expect(ok).To(BeTrue())
	return signer
}

var _ = Describe("NewSelfSignedCert", func() {
	DescribeTable("should generate a usable certificate",
		func(certType util_tls.CertType, keyType util_tls.KeyType, expectedUsage x509.ExtKeyUsage) {
			// when
			pair, err := util_tls.NewSelfSignedCert(certType, keyType, "localhost", "127.0.0.1")

			// then
			Expect(err).ToNot(HaveOccurred())
			_, err = tls.X509KeyPair(pair.CertPEM, pair.KeyPEM)
			Expect(err).ToNot(HaveOccurred())

			cert := parseCert(pair)
			Expect(cert.ExtKeyUsage).To(ConsistOf(expectedUsage))
			Expect(cert.DNSNames).To(ConsistOf("localhost"))
			Expect(cert.IPAddresses).To(HaveLen(1))
			Expect(cert.IPAddresses[0].Equal(net.ParseIP("127.0.0.1"))).To(BeTrue())
			// self-signed certs are their own CA
			Expect(cert.IsCA).To(BeTrue())
			Expect(cert.KeyUsage & x509.KeyUsageCertSign).ToNot(BeZero())
			Expect(cert.CheckSignatureFrom(cert)).To(Succeed())
		},
		Entry("server cert with an RSA key", util_tls.ServerCertType, util_tls.RSAKeyType, x509.ExtKeyUsageServerAuth),
		Entry("client cert with an RSA key", util_tls.ClientCertType, util_tls.RSAKeyType, x509.ExtKeyUsageClientAuth),
		Entry("server cert with an ECDSA key", util_tls.ServerCertType, util_tls.ECDSAKeyType, x509.ExtKeyUsageServerAuth),
	)

	It("should reject an unknown certificate type", func() {
		// when
		_, err := util_tls.NewSelfSignedCert("unknown", util_tls.DefaultKeyType)

		// then
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(`invalid certificate type "unknown"`))
	})
})

var _ = Describe("NewCert", func() {
	It("should generate a certificate signed by the given CA", func() {
		ca, err := util_tls.GenerateCA(util_tls.DefaultKeyType, pkix.Name{CommonName: "test-ca"})
		Expect(err).ToNot(HaveOccurred())
		caCert := parseCert(*ca)

		// when
		pair, err := util_tls.NewCert(*caCert, parseSigner(*ca), util_tls.ServerCertType, util_tls.DefaultKeyType, "backend.mesh")

		// then
		Expect(err).ToNot(HaveOccurred())
		_, err = tls.X509KeyPair(pair.CertPEM, pair.KeyPEM)
		Expect(err).ToNot(HaveOccurred())

		cert := parseCert(pair)
		Expect(cert.IsCA).To(BeFalse())
		Expect(cert.DNSNames).To(ConsistOf("backend.mesh"))
		Expect(cert.CheckSignatureFrom(caCert)).To(Succeed())
	})
})

var _ = Describe("GenerateCA", func() {
	It("should generate a CA that can sign", func() {
		// when
		pair, err := util_tls.GenerateCA(util_tls.DefaultKeyType, pkix.Name{CommonName: "test-ca"})

		// then
		Expect(err).ToNot(HaveOccurred())
		_, err = tls.X509KeyPair(pair.CertPEM, pair.KeyPEM)
		Expect(err).ToNot(HaveOccurred())

		cert := parseCert(*pair)
		Expect(cert.IsCA).To(BeTrue())
		Expect(cert.Subject.CommonName).To(Equal("test-ca"))
		Expect(cert.KeyUsage & x509.KeyUsageCertSign).ToNot(BeZero())
		Expect(cert.KeyUsage & x509.KeyUsageCRLSign).ToNot(BeZero())
	})
})
