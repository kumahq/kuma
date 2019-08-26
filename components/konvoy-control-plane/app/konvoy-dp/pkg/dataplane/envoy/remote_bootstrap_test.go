package envoy

import (
	konvoy_dp "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/konvoy-dp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
)

var _ = Describe("Remote Bootstrap", func() {

	It("should generate bootstrap configuration", func() {
		// given
		mux := http.NewServeMux()
		server := httptest.NewServer(mux)
		mux.HandleFunc("/bootstrap", func(writer http.ResponseWriter, req *http.Request) {
			body, err := ioutil.ReadAll(req.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(body).To(MatchJSON(`
			{
				"nodeId": "demo.sample"
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

		cfg := konvoy_dp.DefaultConfig()
		cfg.Dataplane.Id = "demo.sample"
		cfg.ControlPlane.BootstrapServer.Address = "localhost"
		cfg.ControlPlane.BootstrapServer.Port = uint32(port)

		// when
		config, err := generator(cfg)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(config).ToNot(BeNil())
	})
})
