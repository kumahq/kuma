package server_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	token_server "github.com/Kong/kuma/pkg/config/token-server"
	"github.com/Kong/kuma/pkg/core/xds"
	"github.com/Kong/kuma/pkg/sds/auth"
	"github.com/Kong/kuma/pkg/test"
	"github.com/Kong/kuma/pkg/tokens/builtin/issuer"
	"github.com/Kong/kuma/pkg/tokens/builtin/server"
	"github.com/Kong/kuma/pkg/tokens/builtin/server/types"
	http2 "github.com/Kong/kuma/pkg/util/http"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
)

type staticTokenIssuer struct {
	resp string
}

var _ issuer.DataplaneTokenIssuer = &staticTokenIssuer{}

func (s *staticTokenIssuer) Generate(proxyId xds.ProxyId) (auth.Credential, error) {
	return auth.Credential(s.resp), nil
}

func (s *staticTokenIssuer) Validate(credential auth.Credential) (xds.ProxyId, error) {
	return xds.ProxyId{}, errors.New("not implemented")
}

var _ = Describe("Dataplane Token Server", func() {

	var port int
	var publicPort int
	const credentials = "test"

	httpsClient := func(name string) *http.Client {
		httpClient := &http.Client{}
		err := http2.ConfigureTls(
			httpClient,
			filepath.Join("testdata", "server-cert.pem"),
			filepath.Join("testdata", fmt.Sprintf("%s-cert.pem", name)),
			filepath.Join("testdata", fmt.Sprintf("%s-key.pem", name)),
		)
		Expect(err).ToNot(HaveOccurred())
		return httpClient
	}

	BeforeEach(func() {
		p, err := test.GetFreePort()
		Expect(err).ToNot(HaveOccurred())
		port = p
		p, err = test.GetFreePort()
		Expect(err).ToNot(HaveOccurred())
		publicPort = p
		srv := server.DataplaneTokenServer{
			Issuer: &staticTokenIssuer{credentials},
			Config: &token_server.DataplaneTokenServerConfig{
				Local: &token_server.LocalDataplaneTokenServerConfig{
					Port: uint32(port),
				},
				Public: &token_server.PublicDataplaneTokenServerConfig{
					Port:            uint32(publicPort),
					Interface:       "localhost",
					TlsCertFile:     filepath.Join("testdata", "server-cert.pem"),
					TlsKeyFile:      filepath.Join("testdata", "server-key.pem"),
					ClientCertFiles: []string{filepath.Join("testdata", "authorized-client-cert.pem")},
				},
			},
		}

		ch := make(chan struct{})
		errCh := make(chan error)
		go func() {
			defer GinkgoRecover()
			errCh <- srv.Start(ch)
		}()

		// wait for the http server to be started
		Eventually(func() error {
			req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%d/tokens", port), nil)
			Expect(err).ToNot(HaveOccurred())
			_, err = http.DefaultClient.Do(req)
			return err
		}, "5s", "100ms").ShouldNot(HaveOccurred())

		// wait for the https server to be started
		Eventually(func() error {
			req, err := http.NewRequest("POST", fmt.Sprintf("https://localhost:%d/tokens", publicPort), nil)
			Expect(err).ToNot(HaveOccurred())
			_, err = httpsClient("authorized-client").Do(req)
			return err
		}, "5s", "100ms").ShouldNot(HaveOccurred())
	})

	type testCase struct {
		clientFn func() *http.Client
		url      string
	}
	DescribeTable("should respond with generated token",
		func(given testCase) {
			// given
			idReq := types.DataplaneTokenRequest{
				Mesh: "defualt",
				Name: "dp-1",
			}
			reqBytes, err := json.Marshal(idReq)
			Expect(err).ToNot(HaveOccurred())

			// when
			req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/tokens", port), bytes.NewReader(reqBytes))
			Expect(err).ToNot(HaveOccurred())
			resp, err := http.DefaultClient.Do(req)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(200))

			// when
			respBody, err := ioutil.ReadAll(resp.Body)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(string(respBody)).To(Equal(credentials))
		},
		Entry("using http server", testCase{
			clientFn: func() *http.Client {
				return http.DefaultClient
			},
			url: fmt.Sprintf("http://localhost:%d/tokens", port),
		}),
		Entry("using https server and authorized client", testCase{
			clientFn: func() *http.Client {
				return httpsClient("authorized-client")
			},
			url: fmt.Sprintf("https://localhost:%d/tokens", publicPort),
		}),
	)

	DescribeTable("should return bad request on invalid json",
		func(json string) {
			// when
			req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/tokens", port), strings.NewReader(json))

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			resp, err := http.DefaultClient.Do(req)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(400))
		},
		Entry("json does not contain name", `{"mesh": "default"}`),
		Entry("json does not contain mesh", `{"name": "default"}`),
		Entry("not valid json", `not-valid-json`),
	)

	It("should not let unauthorized clients generate a token", func() {
		// given
		idReq := types.DataplaneTokenRequest{
			Mesh: "defualt",
			Name: "dp-1",
		}
		reqBytes, err := json.Marshal(idReq)
		Expect(err).ToNot(HaveOccurred())

		// when
		req, err := http.NewRequest("GET", fmt.Sprintf("https://localhost:%d/tokens", publicPort), bytes.NewReader(reqBytes))
		Expect(err).ToNot(HaveOccurred())
		_, err = httpsClient("unauthorized-client").Do(req)

		// then
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(HaveSuffix("tls: bad certificate"))
	})
})
