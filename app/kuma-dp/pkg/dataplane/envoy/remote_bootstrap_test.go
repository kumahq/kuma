package envoy

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"

	kuma_dp "github.com/Kong/kuma/pkg/config/app/kuma-dp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Remote Bootstrap", func() {

	It("should generate bootstrap configuration", func() {
		// given
		mux := http.NewServeMux()
		server := httptest.NewServer(mux)
		mux.HandleFunc("/bootstrap", func(writer http.ResponseWriter, req *http.Request) {
			defer GinkgoRecover()
			body, err := ioutil.ReadAll(req.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(body).To(MatchJSON(`
			{
				"mesh": "demo",
				"name": "sample",
				"adminPort": 4321,
				"dataplaneTokenPath": "/tmp/token"
			}
			`))

			response, err := ioutil.ReadFile(filepath.Join("testdata", "remote-bootstrap-config.golden.yaml"))
			Expect(err).ToNot(HaveOccurred())
			_, err = writer.Write(response)
			Expect(err).ToNot(HaveOccurred())
		})
		port, err := strconv.Atoi(strings.Split(server.Listener.Addr().String(), ":")[1])
		Expect(err).ToNot(HaveOccurred())

		// and
		generator := NewRemoteBootstrapGenerator(http.DefaultClient)

		cfg := kuma_dp.DefaultConfig()
		cfg.Dataplane.Mesh = "demo"
		cfg.Dataplane.Name = "sample"
		cfg.Dataplane.AdminPort = 4321
		cfg.DataplaneRuntime.TokenPath = "/tmp/token"
		cfg.ControlPlane.BootstrapServer.URL = fmt.Sprintf("http://localhost:%d", port)

		// when
		config, err := generator(cfg)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(config).ToNot(BeNil())
	})
})
