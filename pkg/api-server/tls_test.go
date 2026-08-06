package api_server_test

import (
	"crypto/tls"
	"fmt"
	"math/big"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	config "github.com/kumahq/kuma/v3/pkg/config/api-server"
	util_tls "github.com/kumahq/kuma/v3/pkg/tls"
)

var _ = Describe("HTTPS server certificate", func() {
	var certPath, keyPath string
	var httpsPort uint32

	// writeCert stores a new self-signed certificate in the files the server
	// was started with.
	writeCert := func() {
		keyPair, err := util_tls.NewSelfSignedCert(util_tls.ServerCertType, util_tls.ECDSAKeyType, "localhost")
		Expect(err).ToNot(HaveOccurred())
		Expect(os.WriteFile(certPath, keyPair.CertPEM, 0o600)).To(Succeed())
		Expect(os.WriteFile(keyPath, keyPair.KeyPEM, 0o600)).To(Succeed())
	}

	servedCertSerial := func(g Gomega) *big.Int {
		conn, err := tls.Dial("tcp", fmt.Sprintf("localhost:%d", httpsPort), &tls.Config{
			InsecureSkipVerify: true, // #nosec G402 -- the test asserts on the served certificate itself
			MinVersion:         tls.VersionTLS12,
		})
		g.Expect(err).ToNot(HaveOccurred())
		defer func() {
			g.Expect(conn.Close()).To(Succeed())
		}()
		return conn.ConnectionState().PeerCertificates[0].SerialNumber
	}

	BeforeEach(func() {
		dir := GinkgoT().TempDir()
		certPath = filepath.Join(dir, "tls.crt")
		keyPath = filepath.Join(dir, "tls.key")
		writeCert()

		cfg := NewTestApiServerConfigurer().WithConfigMutator(func(cfg *config.ApiServerConfig) {
			cfg.HTTPS.TlsCertFile = certPath
			cfg.HTTPS.TlsKeyFile = keyPath
		})
		apiServer, _, stop := StartApiServer(cfg)
		DeferCleanup(stop)
		httpsPort = apiServer.Config().HTTPS.Port
	})

	It("should serve a certificate rotated on disk without a restart", func() {
		// given
		var initialSerial *big.Int
		Eventually(func(g Gomega) {
			initialSerial = servedCertSerial(g)
		}, "5s", "100ms").Should(Succeed())

		// when
		writeCert()

		// then
		Eventually(func(g Gomega) {
			g.Expect(servedCertSerial(g)).ToNot(Equal(initialSerial))
		}, "30s", "100ms").Should(Succeed())
	})
})
