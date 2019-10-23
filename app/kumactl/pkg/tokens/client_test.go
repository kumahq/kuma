package tokens_test

import (
	"fmt"
	"github.com/Kong/kuma/app/kumactl/pkg/tokens"
	config_kumactl "github.com/Kong/kuma/pkg/config/app/kumactl/v1alpha1"
	token_server "github.com/Kong/kuma/pkg/config/token-server"
	"github.com/Kong/kuma/pkg/core/xds"
	"github.com/Kong/kuma/pkg/sds/auth"
	"github.com/Kong/kuma/pkg/test"
	"github.com/Kong/kuma/pkg/tokens/builtin/issuer"
	"github.com/Kong/kuma/pkg/tokens/builtin/server"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
)

type staticTokenIssuer struct {
}

var _ issuer.DataplaneTokenIssuer = &staticTokenIssuer{}

func (s *staticTokenIssuer) Generate(proxyId xds.ProxyId) (auth.Credential, error) {
	return auth.Credential(fmt.Sprintf("token-for-%s-%s", proxyId.Name, proxyId.Mesh)), nil
}

func (s *staticTokenIssuer) Validate(credential auth.Credential) (xds.ProxyId, error) {
	return xds.ProxyId{}, errors.New("not implemented")
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

		srv := server.DataplaneTokenServer{
			Config: &token_server.DataplaneTokenServerConfig{
				Local: &token_server.LocalDataplaneTokenServerConfig{
					Port: uint32(port),
				},
				Public: &token_server.PublicDataplaneTokenServerConfig{
					Port:            uint32(publicPort),
					Interface:       "localhost",
					TlsCertFile:     filepath.Join("..", "..", "..", "..", "pkg", "tokens", "builtin", "server", "testdata", "server-cert.pem"),
					TlsKeyFile:      filepath.Join("..", "..", "..", "..", "pkg", "tokens", "builtin", "server", "testdata", "server-key.pem"),
					ClientCertFiles: []string{filepath.Join("..", "..", "..", "..", "pkg", "tokens", "builtin", "server", "testdata", "authorized-client-cert.pem")},
				},
			},
			Issuer: &staticTokenIssuer{},
		}

		ch := make(chan struct{})
		errCh := make(chan error)
		go func() {
			defer GinkgoRecover()
			errCh <- srv.Start(ch)
		}()
	})

	type testCase struct {
		url    func() string
		config *config_kumactl.DataplaneToken
	}
	DescribeTable("should return a token",
		func(given testCase) {
			// given
			client, err := tokens.NewDataplaneTokenClient(given.url(), given.config)
			Expect(err).ToNot(HaveOccurred())

			// wait for server
			Eventually(func() error {
				_, err := client.Generate("example", "default")
				return err
			}, "5s", "100ms").ShouldNot(HaveOccurred())

			// when
			token, err := client.Generate("example", "default")

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
			config: &config_kumactl.DataplaneToken{
				ServerCert: filepath.Join("..", "..", "..", "..", "pkg", "tokens", "builtin", "server", "testdata", "server-cert.pem"),
				ClientCert: filepath.Join("..", "..", "..", "..", "pkg", "tokens", "builtin", "server", "testdata", "authorized-client-cert.pem"),
				ClientKey:  filepath.Join("..", "..", "..", "..", "pkg", "tokens", "builtin", "server", "testdata", "authorized-client-key.pem"),
			},
		}),
	)

	It("should return an error when status code is different than 200", func() {
		// given
		mux := http.NewServeMux()
		server := httptest.NewServer(mux)
		mux.HandleFunc("/tokens", func(writer http.ResponseWriter, req *http.Request) {
			writer.WriteHeader(500)
		})
		client, err := tokens.NewDataplaneTokenClient(server.URL, nil)
		Expect(err).ToNot(HaveOccurred())

		// when
		_, err = client.Generate("example", "default")

		// then
		Expect(err).To(MatchError("unexpected status code 500. Expected 200"))
	})
})
