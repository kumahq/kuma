package envoy

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	kuma_dp "github.com/Kong/kuma/pkg/config/app/kuma-dp"
	config_types "github.com/Kong/kuma/pkg/config/types"
)

var _ = Describe("Remote Bootstrap", func() {

	type testCase struct {
		config                   kuma_dp.Config
		expectedBootstrapRequest string
	}

	DescribeTable("should generate bootstrap configuration", func(given testCase) {
		// given
		mux := http.NewServeMux()
		server := httptest.NewServer(mux)
		defer server.Close()
		mux.HandleFunc("/bootstrap", func(writer http.ResponseWriter, req *http.Request) {
			defer GinkgoRecover()
			body, err := ioutil.ReadAll(req.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(body).To(MatchJSON(given.expectedBootstrapRequest))

			response, err := ioutil.ReadFile(filepath.Join("testdata", "remote-bootstrap-config.golden.yaml"))
			Expect(err).ToNot(HaveOccurred())
			_, err = writer.Write(response)
			Expect(err).ToNot(HaveOccurred())
		})
		port, err := strconv.Atoi(strings.Split(server.Listener.Addr().String(), ":")[1])
		Expect(err).ToNot(HaveOccurred())

		// and
		generator := NewRemoteBootstrapGenerator(http.DefaultClient)

		// when
		config, err := generator(fmt.Sprintf("http://localhost:%d", port), given.config)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(config).ToNot(BeNil())
	},
		Entry("should support port range with exactly 1 port",
			func() testCase {
				cfg := kuma_dp.DefaultConfig()
				cfg.Dataplane.Mesh = "demo"
				cfg.Dataplane.Name = "sample"
				cfg.Dataplane.AdminPort = config_types.MustExactPort(4321) // exact port
				cfg.DataplaneRuntime.TokenPath = "/tmp/token"

				return testCase{
					config: cfg,
					expectedBootstrapRequest: `
                    {
                      "mesh": "demo",
                      "name": "sample",
                      "adminPort": 4321,
                      "dataplaneTokenPath": "/tmp/token"
                    }
`,
				}
			}()),

		Entry("should support port range with multiple ports (choose the lowest port)",
			func() testCase {
				cfg := kuma_dp.DefaultConfig()
				cfg.Dataplane.Mesh = "demo"
				cfg.Dataplane.Name = "sample"
				cfg.Dataplane.AdminPort = config_types.MustPortRange(4321, 8765) // port range
				cfg.DataplaneRuntime.TokenPath = "/tmp/token"

				return testCase{
					config: cfg,
					expectedBootstrapRequest: `
                    {
                      "mesh": "demo",
                      "name": "sample",
                      "adminPort": 4321,
                      "dataplaneTokenPath": "/tmp/token"
                    }
`,
				}
			}()),
		Entry("should support empty port range",
			func() testCase {
				cfg := kuma_dp.DefaultConfig()
				cfg.Dataplane.Mesh = "demo"
				cfg.Dataplane.Name = "sample"
				cfg.Dataplane.AdminPort = config_types.PortRange{} // empty port range
				cfg.DataplaneRuntime.TokenPath = "/tmp/token"

				return testCase{
					config: cfg,
					expectedBootstrapRequest: `
                    {
                      "mesh": "demo",
                      "name": "sample",
                      "dataplaneTokenPath": "/tmp/token"
                    }
`,
				}
			}()),
	)
})
