package tls_test

import (
	tls "crypto/tls"
	"crypto/x509"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	intercp_tls "github.com/kumahq/kuma/pkg/intercp/tls"
)

var _ = Describe("PKI", func() {
	var caCert tls.Certificate

	BeforeAll(func() {
		pair, err := intercp_tls.GenerateCA()
		Expect(err).ToNot(HaveOccurred())
		caCert, err = tls.X509KeyPair(pair.CertPEM, pair.KeyPEM)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should generate server cert", func() {
		// given
		ip := "192.168.0.1"

		// when
		cert, err := intercp_tls.GenerateServerCert(caCert, ip)

		// then
		Expect(err).ToNot(HaveOccurred())
		clientCert, err := x509.ParseCertificate(cert.Certificate[0])
		Expect(err).ToNot(HaveOccurred())
		Expect(clientCert.IPAddresses[0].String()).To(Equal(ip))
		Expect(clientCert.ExtKeyUsage).To(ContainElement(x509.ExtKeyUsageServerAuth))
	})

	It("should generate client cert", func() {
		// given
		ip := "192.168.0.1"

		// when
		cert, err := intercp_tls.GenerateClientCert(caCert, ip)

		// then
		Expect(err).ToNot(HaveOccurred())
		clientCert, err := x509.ParseCertificate(cert.Certificate[0])
		Expect(err).ToNot(HaveOccurred())
		Expect(clientCert.IPAddresses[0].String()).To(Equal(ip))
		Expect(clientCert.ExtKeyUsage).To(ContainElement(x509.ExtKeyUsageClientAuth))
	})
}, Ordered)
