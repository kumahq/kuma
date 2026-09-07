package api_server_test

import (
	"net"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	config_api_server "github.com/kumahq/kuma/v3/pkg/config/api-server"
	"github.com/kumahq/kuma/v3/pkg/test"
)

var _ = Describe("API Server startup", func() {
	It("closes HTTP when HTTPS cannot start", func() {
		httpsListener, err := net.Listen("tcp", "127.0.0.1:0")
		Expect(err).ToNot(HaveOccurred())
		DeferCleanup(httpsListener.Close)
		httpsPort := httpsListener.Addr().(*net.TCPAddr).Port

		httpPort, err := test.GetFreePort()
		Expect(err).ToNot(HaveOccurred())
		configurer := NewTestApiServerConfigurer().WithConfigMutator(func(cfg *config_api_server.ApiServerConfig) {
			cfg.HTTP.Port = uint32(httpPort)
			cfg.HTTPS.Port = uint32(httpsPort)
		})
		apiServer, _, err := newTestApiServer(configurer)
		Expect(err).ToNot(HaveOccurred())

		err = apiServer.Start(make(chan struct{}))
		Expect(err).To(MatchError(ContainSubstring(httpsListener.Addr().String())))
		Expect(apiServer.Ready()).To(BeFalse())

		httpAddress := net.JoinHostPort("127.0.0.1", strconv.Itoa(httpPort))
		conn, err := net.DialTimeout("tcp", httpAddress, 100*time.Millisecond)
		if conn != nil {
			DeferCleanup(conn.Close)
		}
		Expect(err).To(HaveOccurred())
	})
})
