package tls_test

import (
	"context"
	"os"
	"path/filepath"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	util_tls "github.com/kumahq/kuma/v3/pkg/tls"
)

var _ = Describe("Watchers", func() {
	var dir, certFile, keyFile string
	var watchers *util_tls.Watchers

	newKeyPair := func(host string) *util_tls.KeyPair {
		pair, err := util_tls.NewSelfSignedCert(util_tls.ServerCertType, util_tls.ECDSAKeyType, host)
		Expect(err).ToNot(HaveOccurred())
		return &pair
	}

	writeKeyPair := func(host string) {
		pair := newKeyPair(host)
		Expect(os.WriteFile(certFile, pair.CertPEM, 0o600)).To(Succeed())
		Expect(os.WriteFile(keyFile, pair.KeyPEM, 0o600)).To(Succeed())
	}

	BeforeEach(func() {
		dir = GinkgoT().TempDir()
		certFile = filepath.Join(dir, "tls.crt")
		keyFile = filepath.Join(dir, "tls.key")
		ctx, cancel := context.WithCancel(context.Background())
		DeferCleanup(cancel)
		watchers = util_tls.NewWatchers(ctx, logr.Discard())
	})

	It("should serve the certificate loaded on startup", func() {
		// given
		writeKeyPair("kuma.io")

		// when
		watcher, err := watchers.Watch(certFile, keyFile)

		// then
		Expect(err).ToNot(HaveOccurred())
		cert, err := watcher.Certificate()
		Expect(err).ToNot(HaveOccurred())
		Expect(cert.Leaf.DNSNames).To(ConsistOf("kuma.io"))
	})

	It("should watch the same files once", func() {
		// given
		writeKeyPair("kuma.io")
		watcher, err := watchers.Watch(certFile, keyFile)
		Expect(err).ToNot(HaveOccurred())

		// when
		same, err := watchers.Watch(certFile, keyFile)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(same).To(BeIdenticalTo(watcher))
	})

	It("should watch a key pair per server that configures its own", func() {
		// given a server pointed at its own key pair
		writeKeyPair("kuma.io")
		otherCertFile := filepath.Join(dir, "other.crt")
		otherKeyFile := filepath.Join(dir, "other.key")
		other := newKeyPair("other.kuma.io")
		Expect(os.WriteFile(otherCertFile, other.CertPEM, 0o600)).To(Succeed())
		Expect(os.WriteFile(otherKeyFile, other.KeyPEM, 0o600)).To(Succeed())

		// when
		watcher, err := watchers.Watch(certFile, keyFile)
		Expect(err).ToNot(HaveOccurred())
		otherWatcher, err := watchers.Watch(otherCertFile, otherKeyFile)
		Expect(err).ToNot(HaveOccurred())

		// then each serves the key pair it was configured with
		Expect(otherWatcher).ToNot(BeIdenticalTo(watcher))
		cert, err := watcher.Certificate()
		Expect(err).ToNot(HaveOccurred())
		Expect(cert.Leaf.DNSNames).To(ConsistOf("kuma.io"))
		otherCert, err := otherWatcher.Certificate()
		Expect(err).ToNot(HaveOccurred())
		Expect(otherCert.Leaf.DNSNames).To(ConsistOf("other.kuma.io"))

		// and rotating one leaves the other alone
		Expect(os.WriteFile(otherCertFile, newKeyPair("rotated.kuma.io").CertPEM, 0o600)).To(Succeed())
		Consistently(func(g Gomega) {
			cert, err := watcher.Certificate()
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(cert.Leaf.DNSNames).To(ConsistOf("kuma.io"))
		}, "1s", "100ms").Should(Succeed())
	})

	It("should pick up a certificate rotated on disk", func() {
		// given
		writeKeyPair("kuma.io")
		watcher, err := watchers.Watch(certFile, keyFile)
		Expect(err).ToNot(HaveOccurred())

		// when
		writeKeyPair("rotated.kuma.io")

		// then
		Eventually(func(g Gomega) {
			cert, err := watcher.Certificate()
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(cert.Leaf.DNSNames).To(ConsistOf("rotated.kuma.io"))
		}, "30s", "100ms").Should(Succeed())
	})

	It("should pick up a certificate rotated by swapping a symlinked directory", func() {
		// given a mounted secret, which kubelet rotates by writing a new
		// directory and swapping the ..data symlink instead of rewriting files
		writeData := func(name string, pair *util_tls.KeyPair) {
			data := filepath.Join(dir, name)
			Expect(os.Mkdir(data, 0o700)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(data, "tls.crt"), pair.CertPEM, 0o600)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(data, "tls.key"), pair.KeyPEM, 0o600)).To(Succeed())
		}
		swapData := func(name string) {
			link := filepath.Join(dir, "..data.tmp")
			Expect(os.Symlink(filepath.Join(dir, name), link)).To(Succeed())
			Expect(os.Rename(link, filepath.Join(dir, "..data"))).To(Succeed())
		}
		writeData("..timestamp-1", newKeyPair("kuma.io"))
		swapData("..timestamp-1")
		certFile = filepath.Join(dir, "..data", "tls.crt")
		keyFile = filepath.Join(dir, "..data", "tls.key")

		watcher, err := watchers.Watch(certFile, keyFile)
		Expect(err).ToNot(HaveOccurred())

		// when
		writeData("..timestamp-2", newKeyPair("rotated.kuma.io"))
		swapData("..timestamp-2")

		// then
		Eventually(func(g Gomega) {
			cert, err := watcher.Certificate()
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(cert.Leaf.DNSNames).To(ConsistOf("rotated.kuma.io"))
		}, "30s", "100ms").Should(Succeed())
	})

	It("should keep the previous certificate when the new one cannot be parsed", func() {
		// given
		writeKeyPair("kuma.io")
		watcher, err := watchers.Watch(certFile, keyFile)
		Expect(err).ToNot(HaveOccurred())

		// when a rotation is caught with the certificate half-written
		Expect(os.WriteFile(certFile, []byte("-----BEGIN CERTIFICATE-----\n"), 0o600)).To(Succeed())

		// then
		Consistently(func(g Gomega) {
			cert, err := watcher.Certificate()
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(cert.Leaf.DNSNames).To(ConsistOf("kuma.io"))
		}, "1s", "100ms").Should(Succeed())

		// and when the rotation completes
		writeKeyPair("rotated.kuma.io")

		// then
		Eventually(func(g Gomega) {
			cert, err := watcher.Certificate()
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(cert.Leaf.DNSNames).To(ConsistOf("rotated.kuma.io"))
		}, "30s", "100ms").Should(Succeed())
	})

	It("should keep the previous certificate when the key does not match it", func() {
		// given
		writeKeyPair("kuma.io")
		watcher, err := watchers.Watch(certFile, keyFile)
		Expect(err).ToNot(HaveOccurred())

		// when the key of the rotation lands before its certificate
		rotated := newKeyPair("rotated.kuma.io")
		Expect(os.WriteFile(keyFile, rotated.KeyPEM, 0o600)).To(Succeed())

		// then
		Consistently(func(g Gomega) {
			cert, err := watcher.Certificate()
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(cert.Leaf.DNSNames).To(ConsistOf("kuma.io"))
		}, "1s", "100ms").Should(Succeed())

		// and when the certificate lands too
		Expect(os.WriteFile(certFile, rotated.CertPEM, 0o600)).To(Succeed())

		// then
		Eventually(func(g Gomega) {
			cert, err := watcher.Certificate()
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(cert.Leaf.DNSNames).To(ConsistOf("rotated.kuma.io"))
		}, "30s", "100ms").Should(Succeed())
	})

	It("should fail when the certificate cannot be loaded", func() {
		// when
		_, err := watchers.Watch(certFile, keyFile)

		// then
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to load TLS certificate"))
	})
})
