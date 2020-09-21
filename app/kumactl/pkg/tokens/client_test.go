package tokens_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/app/kumactl/pkg/tokens"
	admin_server "github.com/kumahq/kuma/pkg/admin-server"
	admin_server_config "github.com/kumahq/kuma/pkg/config/admin-server"
	config_kumactl "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	"github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/sds/auth"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
	"github.com/kumahq/kuma/pkg/tokens/builtin/server"
)

type staticTokenIssuer struct {
}

var _ issuer.DataplaneTokenIssuer = &staticTokenIssuer{}

func (s *staticTokenIssuer) Generate(identity issuer.DataplaneIdentity) (auth.Credential, error) {
	return auth.Credential(fmt.Sprintf("token-for-%s-%s", identity.Name, identity.Mesh)), nil
}

func (s *staticTokenIssuer) Validate(credential auth.Credential) (issuer.DataplaneIdentity, error) {
	return issuer.DataplaneIdentity{}, errors.New("not implemented")
}

var _ = Describe("Tokens Client", func() {

	var port int
	var publicPort int

	BeforeEach(func() {
		p, err := test.GetFreePort()
		Expect(err).ToNot(HaveOccurred())
		port = p
		p, err = test.GetFreePort()
		Expect(err).ToNot(HaveOccurred())
		publicPort = p

		adminCfg := admin_server_config.AdminServerConfig{
			Apis: &admin_server_config.AdminServerApisConfig{
				DataplaneToken: &admin_server_config.DataplaneTokenApiConfig{
					Enabled: true,
				},
			},
			Local: &admin_server_config.LocalAdminServerConfig{
				Port: uint32(port),
			},
			Public: &admin_server_config.PublicAdminServerConfig{
				Enabled:        true,
				Port:           uint32(publicPort),
				Interface:      "localhost",
				TlsCertFile:    filepath.Join("..", "..", "..", "..", "pkg", "admin-server", "testdata", "server-cert.pem"),
				TlsKeyFile:     filepath.Join("..", "..", "..", "..", "pkg", "admin-server", "testdata", "server-key.pem"),
				ClientCertsDir: filepath.Join("..", "..", "..", "..", "pkg", "admin-server", "testdata", "authorized-clients"),
			},
		}
		metrics, err := metrics.NewMetrics("Standalone")
		Expect(err).ToNot(HaveOccurred())
		srv := admin_server.NewAdminServer(adminCfg, metrics, server.NewWebservice(&staticTokenIssuer{}))

		ch := make(chan struct{})
		errCh := make(chan error)
		go func() {
			defer GinkgoRecover()
			errCh <- srv.Start(ch)
		}()
	})

	type testCase struct {
		url    func() string
		config *config_kumactl.Context_AdminApiCredentials
	}
	DescribeTable("should return a token",
		func(given testCase) {
			// given
			client, err := tokens.NewDataplaneTokenClient(given.url(), given.config)
			Expect(err).ToNot(HaveOccurred())

			// wait for server
			Eventually(func() error {
				_, err := client.Generate("example", "default", nil)
				return err
			}, "5s", "100ms").ShouldNot(HaveOccurred())

			// when
			token, err := client.Generate("example", "default", nil)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(token).To(Equal("token-for-example-default"))
		},
		Entry("with http server", testCase{
			url: func() string {
				return fmt.Sprintf("http://localhost:%d", port)
			},
			config: nil,
		}),
		Entry("with https server configuration", testCase{
			url: func() string {
				return fmt.Sprintf("https://localhost:%d", publicPort)
			},
			config: &config_kumactl.Context_AdminApiCredentials{
				ClientCert: filepath.Join("..", "..", "..", "..", "pkg", "admin-server", "testdata", "authorized-client-cert.pem"),
				ClientKey:  filepath.Join("..", "..", "..", "..", "pkg", "admin-server", "testdata", "authorized-client-key.pem"),
			},
		}),
	)

	It("should return an error when status code is different than 200", func() {
		// given
		mux := http.NewServeMux()
		server := httptest.NewServer(mux)
		defer server.Close()
		mux.HandleFunc("/tokens", func(writer http.ResponseWriter, req *http.Request) {
			writer.WriteHeader(500)
		})
		client, err := tokens.NewDataplaneTokenClient(server.URL, nil)
		Expect(err).ToNot(HaveOccurred())

		// when
		_, err = client.Generate("example", "default", nil)

		// then
		Expect(err).To(MatchError("unexpected status code 500. Expected 200"))
	})
})
