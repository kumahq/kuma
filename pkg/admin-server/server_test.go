package admin_server_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/emicklei/go-restful"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	admin_server "github.com/Kong/kuma/pkg/admin-server"
	admin_server_config "github.com/Kong/kuma/pkg/config/admin-server"
	"github.com/Kong/kuma/pkg/test"
	util_http "github.com/Kong/kuma/pkg/util/http"
)

var _ = Describe("Admin Server", func() {

	httpsClient := func(name string) *http.Client {
		httpClient := &http.Client{}
		err := util_http.ConfigureTls(
			httpClient,
			filepath.Join("testdata", "server-cert.pem"),
			filepath.Join("testdata", fmt.Sprintf("%s-cert.pem", name)),
			filepath.Join("testdata", fmt.Sprintf("%s-key.pem", name)),
		)
		Expect(err).ToNot(HaveOccurred())
		return httpClient
	}

	var port int
	var publicPort int

	BeforeEach(func() {
		// setup server
		p, err := test.GetFreePort()
		Expect(err).ToNot(HaveOccurred())
		port = p
		p, err = test.GetFreePort()
		Expect(err).ToNot(HaveOccurred())
		publicPort = p

		cfg := admin_server_config.AdminServerConfig{
			Local: &admin_server_config.LocalAdminServerConfig{
				Port: uint32(port),
			},
			Public: &admin_server_config.PublicAdminServerConfig{
				Enabled:        true,
				Port:           uint32(publicPort),
				Interface:      "localhost",
				TlsCertFile:    filepath.Join("testdata", "server-cert.pem"),
				TlsKeyFile:     filepath.Join("testdata", "server-key.pem"),
				ClientCertsDir: filepath.Join("testdata", "authorized-clients"),
			},
		}

		srv := admin_server.NewAdminServer(cfg, pingWs())

		ch := make(chan struct{})
		go func() {
			defer GinkgoRecover()
			Expect(srv.Start(ch)).ToNot(HaveOccurred())
		}()

		// wait for server to start
		Eventually(func() error {
			req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/ping", port), nil)
			Expect(err).ToNot(HaveOccurred())
			_, err = http.DefaultClient.Do(req)
			return err
		}, "5s", "100ms").ShouldNot(HaveOccurred())

		// and should response on public port
		Eventually(func() error {
			req, err := http.NewRequest("GET", fmt.Sprintf("https://localhost:%d/ping", publicPort), nil)
			Expect(err).ToNot(HaveOccurred())
			_, err = httpsClient("authorized-client").Do(req)
			return err
		}, "5s", "100ms").ShouldNot(HaveOccurred())
	})

	It("should serve on http", func() {
		// given
		req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/ping", port), nil)
		Expect(err).ToNot(HaveOccurred())

		// when
		resp, err := http.DefaultClient.Do(req)

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		bytes, err := ioutil.ReadAll(resp.Body)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(string(bytes)).To(Equal("pong"))
	})

	It("should serve on https for authorized client", func() {
		// given
		req, err := http.NewRequest("GET", fmt.Sprintf("https://localhost:%d/ping", publicPort), nil)
		Expect(err).ToNot(HaveOccurred())

		// when
		resp, err := httpsClient("authorized-client").Do(req)

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		bytes, err := ioutil.ReadAll(resp.Body)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(string(bytes)).To(Equal("pong"))
	})

	It("should not let unauthorized to use an https server", func() {
		// given
		req, err := http.NewRequest("GET", fmt.Sprintf("https://localhost:%d/ping", publicPort), nil)
		Expect(err).ToNot(HaveOccurred())

		// when
		_, err = httpsClient("unauthorized-client").Do(req)

		// then
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(HaveSuffix("tls: bad certificate"))
	})
})

func pingWs() *restful.WebService {
	ws := new(restful.WebService).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	ws.Path("/").
		Route(ws.GET("/ping").To(func(request *restful.Request, response *restful.Response) {
			_, _ = response.Write([]byte(`pong`))
		}),
		)
	return ws
}
