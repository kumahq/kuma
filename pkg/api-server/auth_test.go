package api_server_test

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	config "github.com/kumahq/kuma/pkg/config/api-server"
	"github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/certs"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/tls"
	http2 "github.com/kumahq/kuma/pkg/util/http"
)

var _ = Describe("Auth test", func() {

	var httpsClient *http.Client
	var httpsClientWithoutCerts *http.Client
	var httpPort uint32
	var httpsPort uint32
	var stop chan struct{}
	var externalIP string

	BeforeEach(func() {
		externalIP = getExternalIP()
		Expect(externalIP).ToNot(BeEmpty())
		certPath, keyPath := createCertsForIP(externalIP)

		resourceStore := memory.NewStore()
		metrics, err := metrics.NewMetrics("Standalone")
		Expect(err).ToNot(HaveOccurred())

		cfg := config.DefaultApiServerConfig()
		cfg.HTTPS.TlsCertFile = certPath
		cfg.HTTPS.TlsKeyFile = keyPath
		cfg.Authn.Type = certs.PluginName
		cfg.Auth.ClientCertsDir = filepath.Join("..", "..", "test", "certs", "client")
		apiServer := createTestApiServer(resourceStore, cfg, true, metrics)
		httpsPort = cfg.HTTPS.Port
		httpPort = cfg.HTTP.Port

		// start the server
		stop = make(chan struct{})
		go func() {
			defer GinkgoRecover()
			err := apiServer.Start(stop)
			Expect(err).ToNot(HaveOccurred())
		}()

		// configure https clients
		httpsClient = &http.Client{}
		err = http2.ConfigureMTLS(
			httpsClient,
			certPath,
			filepath.Join("..", "..", "test", "certs", "client", "client.pem"),
			filepath.Join("..", "..", "test", "certs", "client", "client.key"),
		)
		Expect(err).ToNot(HaveOccurred())

		httpsClientWithoutCerts = &http.Client{}
		err = http2.ConfigureMTLS(httpsClientWithoutCerts, certPath, "", "")
		Expect(err).ToNot(HaveOccurred())

		// wait for both http and https server
		Eventually(func() bool {
			resp, err := httpsClient.Get(fmt.Sprintf("https://localhost:%d/secrets", httpsPort))
			if err != nil || resp.StatusCode != 200 {
				return false
			}
			resp, err = http.Get(fmt.Sprintf("http://localhost:%d/secrets", httpPort))
			if err != nil || resp.StatusCode != 200 {
				return false
			}
			return true
		}, "5s", "100ms").Should(BeTrue())
	})

	AfterEach(func() {
		close(stop)
	})

	It("should be able to access secrets on localhost using HTTP", func() {
		// when
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/secrets", httpPort))

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(200))
	})

	It("should be able to access admin endpoints using client certs and HTTPS", func() {
		// when
		resp, err := httpsClient.Get(fmt.Sprintf("https://%s:%d/secrets", externalIP, httpsPort))

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(200))
	})

	It("should be block an access to admin endpoints from other machine using HTTP", func() {
		// when
		resp, err := http.Get(fmt.Sprintf("http://%s:%d/secrets", externalIP, httpPort))

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(403))
		body, err := io.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Body.Close()).To(Succeed())
		Expect(string(body)).To(MatchJSON(`{"title": "Access Denied", "details": "user \"mesh-system:anonymous/mesh-system:unauthenticated\" cannot access the resource of type \"Secret\""}`))
	})

	It("should be block an access to admin endpoints from other machine using HTTPS without proper client certs", func() {
		// when
		resp, err := httpsClientWithoutCerts.Get(fmt.Sprintf("https://%s:%d/secrets", externalIP, httpsPort))

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(403))
		body, err := io.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Body.Close()).To(Succeed())
		Expect(string(body)).To(MatchJSON(`{"title": "Access Denied", "details": "user \"mesh-system:anonymous/mesh-system:unauthenticated\" cannot access the resource of type \"Secret\""}`))
	})
})

// we need to autogenerate cert dynamically for the external IP so the HTTPS client can validate san
func createCertsForIP(ip string) (certPath string, keyPath string) {
	keyPair, err := tls.NewSelfSignedCert("kuma", tls.ServerCertType, tls.DefaultKeyType, "localhost", ip)
	Expect(err).ToNot(HaveOccurred())
	dir, err := os.MkdirTemp("", "temp-certs")
	Expect(err).ToNot(HaveOccurred())
	certPath = dir + "/cert.pem"
	keyPath = dir + "/cert.key"
	err = os.WriteFile(certPath, keyPair.CertPEM, 0600)
	Expect(err).ToNot(HaveOccurred())
	err = os.WriteFile(keyPath, keyPair.KeyPEM, 0600)
	Expect(err).ToNot(HaveOccurred())
	return
}

// GetLocalIP returns the non loopback local IP of the host. It assumes there is another network interface on the machine aside of loopback
// We need to use this interface to simulate accessing the server from the remote machine
func getExternalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok &&
			!ipnet.IP.IsLoopback() &&
			ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}
	return ""
}
