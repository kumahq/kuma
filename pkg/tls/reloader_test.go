package tls_test

import (
	"os"
	"path/filepath"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	util_tls "github.com/kumahq/kuma/v3/pkg/tls"
)

var _ = Describe("KeyPairReloader", func() {
	var certFile, keyFile string

	// writeKeyPair generates a new self-signed key pair and stores it in the
	// files watched by the reloader. Modification times are set explicitly
	// because a test can easily write twice within the resolution of the
	// filesystem clock.
	writeKeyPair := func(host string, modTime time.Time) util_tls.KeyPair {
		pair, err := util_tls.NewSelfSignedCert(util_tls.ServerCertType, util_tls.ECDSAKeyType, host)
		Expect(err).ToNot(HaveOccurred())
		Expect(os.WriteFile(certFile, pair.CertPEM, 0o600)).To(Succeed())
		Expect(os.WriteFile(keyFile, pair.KeyPEM, 0o600)).To(Succeed())
		Expect(os.Chtimes(certFile, modTime, modTime)).To(Succeed())
		Expect(os.Chtimes(keyFile, modTime, modTime)).To(Succeed())
		return pair
	}

	BeforeEach(func() {
		dir := GinkgoT().TempDir()
		certFile = filepath.Join(dir, "tls.crt")
		keyFile = filepath.Join(dir, "tls.key")
	})

	It("should serve the certificate loaded on startup", func() {
		// given
		writeKeyPair("kuma.io", time.Now())

		// when
		reloader, err := util_tls.NewKeyPairReloader(certFile, keyFile, logr.Discard())

		// then
		Expect(err).ToNot(HaveOccurred())
		cert, err := reloader.GetCertificate(nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(cert.Leaf.DNSNames).To(ConsistOf("kuma.io"))
	})

	It("should pick up a certificate rotated on disk", func() {
		// given
		writeKeyPair("kuma.io", time.Now())
		reloader, err := util_tls.NewKeyPairReloader(certFile, keyFile, logr.Discard())
		Expect(err).ToNot(HaveOccurred())

		// when
		writeKeyPair("rotated.kuma.io", time.Now().Add(time.Minute))

		// then
		cert, err := reloader.GetCertificate(nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(cert.Leaf.DNSNames).To(ConsistOf("rotated.kuma.io"))
	})

	It("should keep the previous certificate when the new one cannot be loaded", func() {
		// given
		writeKeyPair("kuma.io", time.Now())
		reloader, err := util_tls.NewKeyPairReloader(certFile, keyFile, logr.Discard())
		Expect(err).ToNot(HaveOccurred())

		// when a rotation is caught half-written
		Expect(os.WriteFile(certFile, []byte("-----BEGIN CERTIFICATE-----\n"), 0o600)).To(Succeed())

		// then
		cert, err := reloader.GetCertificate(nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(cert.Leaf.DNSNames).To(ConsistOf("kuma.io"))

		// and when the rotation completes
		writeKeyPair("rotated.kuma.io", time.Now().Add(time.Minute))

		// then
		cert, err = reloader.GetCertificate(nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(cert.Leaf.DNSNames).To(ConsistOf("rotated.kuma.io"))
	})

	It("should fail when the certificate cannot be loaded on startup", func() {
		// when
		_, err := util_tls.NewKeyPairReloader(certFile, keyFile, logr.Discard())

		// then
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to load TLS certificate"))
	})
})
