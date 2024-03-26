package api_server_test

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	api_server "github.com/kumahq/kuma/pkg/api-server"
	config "github.com/kumahq/kuma/pkg/config/api-server"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/certs"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/tls"
	http2 "github.com/kumahq/kuma/pkg/util/http"
)

var _ = Describe("Auth test", func() {
	var httpsClient *http.Client
	var httpsClientWithoutCerts *http.Client
	var httpPort uint32
	var httpsPort uint32
	stop := func() {}
	var externalIP string

	BeforeEach(func() {
		externalIP = getExternalIP()
		Expect(externalIP).ToNot(BeEmpty())
		certPath, keyPath := createCertsForIP(externalIP)
		var apiServer *api_server.ApiServer
		apiServer, _, stop = StartApiServer(NewTestApiServerConfigurer().WithConfigMutator(func(cfg *config.ApiServerConfig) {
			cfg.HTTPS.TlsCertFile = certPath
			cfg.HTTPS.TlsKeyFile = keyPath
			cfg.Authn.Type = certs.PluginName
			cfg.Auth.ClientCertsDir = filepath.Join("..", "..", "test", "certs", "client")
		}))

		cfg := apiServer.Config()
		httpsPort = cfg.HTTPS.Port
		httpPort = cfg.HTTP.Port

		// configure https clients
		httpsClient = &http.Client{}
		Expect(http2.ConfigureMTLS(
			httpsClient,
			certPath,
			filepath.Join("..", "..", "test", "certs", "client", "client.pem"),
			filepath.Join("..", "..", "test", "certs", "client", "client.key"),
		)).To(Succeed())

		httpsClientWithoutCerts = &http.Client{}
		Expect(http2.ConfigureMTLS(httpsClientWithoutCerts, certPath, "", "")).To(Succeed())

		// wait for both http and https server
		Eventually(func(g Gomega) {
			resp, err := httpsClient.Get(fmt.Sprintf("https://localhost:%d/secrets", httpsPort))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp).To(HaveHTTPStatus(200))
			resp, err = http.Get(fmt.Sprintf("http://localhost:%d/secrets", httpPort))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp).To(HaveHTTPStatus(200))
		}, "5s", "100ms").Should(Succeed())
	})

	AfterEach(func() {
		stop()
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
		resp, err := httpsClient.Get(fmt.Sprintf("https://%s/secrets", net.JoinHostPort(externalIP, strconv.Itoa(int(httpsPort)))))

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(200))
	})

	It("should be block an access to admin endpoints from other machine using HTTP", func() {
		// when
		resp, err := http.Get(fmt.Sprintf("http://%s/secrets", net.JoinHostPort(externalIP, strconv.Itoa(int(httpPort)))))

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(403))
		bytes, err := io.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Body.Close()).To(Succeed())
		Expect(bytes).To(matchers.MatchGoldenJSON(path.Join("testdata", "auth-admin-non-localhost.golden.json")))
	})

	It("should be block an access to admin endpoints from other machine using HTTPS without proper client certs", func() {
		// when
		resp, err := httpsClientWithoutCerts.Get(fmt.Sprintf("https://%s/secrets", net.JoinHostPort(externalIP, strconv.Itoa(int(httpsPort)))))

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(403))
		bytes, err := io.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Body.Close()).To(Succeed())
		Expect(bytes).To(matchers.MatchGoldenJSON(path.Join("testdata", "auth-admin-https-bad-creds.golden.json")))
	})
})

// we need to autogenerate cert dynamically for the external IP so the HTTPS client can validate san
func createCertsForIP(ip string) (string, string) {
	keyPair, err := tls.NewSelfSignedCert(tls.ServerCertType, tls.DefaultKeyType, "localhost", ip)
	Expect(err).ToNot(HaveOccurred())
	dir, err := os.MkdirTemp("", "temp-certs")
	Expect(err).ToNot(HaveOccurred())
	certPath := dir + "/cert.pem"
	keyPath := dir + "/cert.key"
	err = os.WriteFile(certPath, keyPair.CertPEM, 0o600)
	Expect(err).ToNot(HaveOccurred())
	err = os.WriteFile(keyPath, keyPair.KeyPEM, 0o600)
	Expect(err).ToNot(HaveOccurred())
	return certPath, keyPath
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
