package tls_test

import (
	"os"
	"path/filepath"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	util_tls "github.com/kumahq/kuma/v3/pkg/tls"
)

var _ = Describe("WatchKeyPair", func() {
	var certFile, keyFile string
	var stop chan struct{}

	writeKeyPair := func(host string) {
		pair, err := util_tls.NewSelfSignedCert(util_tls.ServerCertType, util_tls.ECDSAKeyType, host)
		Expect(err).ToNot(HaveOccurred())
		Expect(os.WriteFile(certFile, pair.CertPEM, 0o600)).To(Succeed())
		Expect(os.WriteFile(keyFile, pair.KeyPEM, 0o600)).To(Succeed())
	}

	BeforeEach(func() {
		dir := GinkgoT().TempDir()
		certFile = filepath.Join(dir, "tls.crt")
		keyFile = filepath.Join(dir, "tls.key")
		stop = make(chan struct{})
		DeferCleanup(func() {
			close(stop)
		})
	})

	It("should serve the certificate loaded on startup", func() {
		// given
		writeKeyPair("kuma.io")

		// when
		getCertificate, err := util_tls.WatchKeyPair(certFile, keyFile, stop, logr.Discard())

		// then
		Expect(err).ToNot(HaveOccurred())
		cert, err := getCertificate(nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(cert.Leaf.DNSNames).To(ConsistOf("kuma.io"))
	})

	It("should pick up a certificate rotated on disk", func() {
		// given
		writeKeyPair("kuma.io")
		getCertificate, err := util_tls.WatchKeyPair(certFile, keyFile, stop, logr.Discard())
		Expect(err).ToNot(HaveOccurred())

		// when
		writeKeyPair("rotated.kuma.io")

		// then
		Eventually(func(g Gomega) {
			cert, err := getCertificate(nil)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(cert.Leaf.DNSNames).To(ConsistOf("rotated.kuma.io"))
		}, "30s", "100ms").Should(Succeed())
	})

	It("should keep the previous certificate when the new one cannot be loaded", func() {
		// given
		writeKeyPair("kuma.io")
		getCertificate, err := util_tls.WatchKeyPair(certFile, keyFile, stop, logr.Discard())
		Expect(err).ToNot(HaveOccurred())

		// when a rotation is caught half-written
		Expect(os.WriteFile(certFile, []byte("-----BEGIN CERTIFICATE-----\n"), 0o600)).To(Succeed())

		// then
		Consistently(func(g Gomega) {
			cert, err := getCertificate(nil)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(cert.Leaf.DNSNames).To(ConsistOf("kuma.io"))
		}, "1s", "100ms").Should(Succeed())

		// and when the rotation completes
		writeKeyPair("rotated.kuma.io")

		// then
		Eventually(func(g Gomega) {
			cert, err := getCertificate(nil)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(cert.Leaf.DNSNames).To(ConsistOf("rotated.kuma.io"))
		}, "30s", "100ms").Should(Succeed())
	})

	It("should fail when the certificate cannot be loaded on startup", func() {
		// when
		_, err := util_tls.WatchKeyPair(certFile, keyFile, stop, logr.Discard())

		// then
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to load TLS certificate"))
	})
})
