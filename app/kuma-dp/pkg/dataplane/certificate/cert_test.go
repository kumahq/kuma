package certificate

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func newCertPEM(cn string) []byte {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	Expect(err).ToNot(HaveOccurred())
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: cn},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	Expect(err).ToNot(HaveOccurred())
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

func bundleCommonNames(path string) []string {
	data, err := os.ReadFile(path)
	Expect(err).ToNot(HaveOccurred())
	var names []string
	for block, rest := pem.Decode(data); block != nil; block, rest = pem.Decode(rest) {
		cert, err := x509.ParseCertificate(block.Bytes)
		Expect(err).ToNot(HaveOccurred())
		names = append(names, cert.Subject.CommonName)
	}
	return names
}

var _ = Describe("findOsCa", func() {
	var root, workDir string

	BeforeEach(func() {
		root = GinkgoT().TempDir()
		workDir = GinkgoT().TempDir()
	})

	writeFile := func(path string, data []byte) {
		Expect(os.MkdirAll(filepath.Dir(path), 0o755)).To(Succeed())
		Expect(os.WriteFile(path, data, 0o600)).To(Succeed())
	}

	It("prefers a CA bundle file", func() {
		file := filepath.Join(root, "ca-certificates.crt")
		writeFile(file, newCertPEM("bundle"))
		writeFile(filepath.Join(root, "certs", "a.pem"), newCertPEM("a"))

		Expect(findOsCa([]string{filepath.Join(root, "missing.crt"), file}, []string{filepath.Join(root, "certs")}, workDir)).To(Equal(file))
		Expect(filepath.Join(workDir, BundleFileName)).ToNot(BeAnExistingFile())
	})

	It("does not take a directory for a bundle file", func() {
		dir := filepath.Join(root, "certs")
		writeFile(filepath.Join(dir, "a.pem"), newCertPEM("a"))

		path := findOsCa([]string{dir}, []string{dir}, workDir)
		Expect(path).To(Equal(filepath.Join(workDir, BundleFileName)))
	})

	It("bundles the certificates of the trust store directories", func() {
		a := newCertPEM("a")
		writeFile(filepath.Join(root, "certs", "a.pem"), a)
		writeFile(filepath.Join(root, "certs", "b.crt"), append(newCertPEM("b"), newCertPEM("c")...))
		writeFile(filepath.Join(root, "certs", "README"), []byte("not a certificate"))
		writeFile(filepath.Join(root, "certs", "key.pem"), pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte("x")}))
		writeFile(filepath.Join(root, "certs", "nested", "d.pem"), newCertPEM("d"))
		writeFile(filepath.Join(root, "other", "a-copy.pem"), a)
		writeFile(filepath.Join(root, "other", "e.pem"), newCertPEM("e"))

		path := findOsCa(nil, []string{filepath.Join(root, "certs"), filepath.Join(root, "missing"), "", filepath.Join(root, "other")}, workDir)

		Expect(path).To(Equal(filepath.Join(workDir, BundleFileName)))
		Expect(bundleCommonNames(path)).To(ConsistOf("a", "b", "c", "e"))
	})

	It("returns nothing when no certificate is found", func() {
		writeFile(filepath.Join(root, "certs", "README"), []byte("not a certificate"))

		Expect(findOsCa([]string{filepath.Join(root, "missing.crt")}, []string{filepath.Join(root, "certs")}, workDir)).To(BeEmpty())
		Expect(filepath.Join(workDir, BundleFileName)).ToNot(BeAnExistingFile())
	})

	It("returns nothing when there is no work directory for the bundle", func() {
		writeFile(filepath.Join(root, "certs", "a.pem"), newCertPEM("a"))

		Expect(findOsCa(nil, []string{filepath.Join(root, "certs")}, "")).To(BeEmpty())
	})
})

var _ = Describe("GetOsCaFilePath", func() {
	setEnv := func(key, value string) {
		old, had := os.LookupEnv(key)
		Expect(os.Setenv(key, value)).To(Succeed())
		DeferCleanup(func() {
			if had {
				Expect(os.Setenv(key, old)).To(Succeed())
			} else {
				Expect(os.Unsetenv(key)).To(Succeed())
			}
		})
	}

	It("honors SSL_CERT_FILE", func() {
		file := filepath.Join(GinkgoT().TempDir(), "custom.pem")
		Expect(os.WriteFile(file, newCertPEM("custom"), 0o600)).To(Succeed())
		setEnv("SSL_CERT_FILE", file)

		Expect(GetOsCaFilePath(GinkgoT().TempDir())).To(Equal(file))
	})

	It("honors SSL_CERT_DIR with several directories", func() {
		root := GinkgoT().TempDir()
		for _, name := range []string{"one", "two"} {
			Expect(os.MkdirAll(filepath.Join(root, name), 0o755)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(root, name, name+".pem"), newCertPEM(name), 0o600)).To(Succeed())
		}
		setEnv("SSL_CERT_FILE", filepath.Join(root, "missing.pem"))
		setEnv("SSL_CERT_DIR", filepath.Join(root, "one")+":"+filepath.Join(root, "two"))
		workDir := GinkgoT().TempDir()

		path := GetOsCaFilePath(workDir)

		Expect(path).To(Equal(filepath.Join(workDir, BundleFileName)))
		Expect(bundleCommonNames(path)).To(ConsistOf("one", "two"))
	})
})
