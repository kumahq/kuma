package envoy

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	kuma_version "github.com/kumahq/kuma/pkg/version"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/resources/model/rest"

	kuma_dp "github.com/kumahq/kuma/pkg/config/app/kuma-dp"
	config_types "github.com/kumahq/kuma/pkg/config/types"
)

var _ = Describe("Remote Bootstrap", func() {

	type testCase struct {
		config                   kuma_dp.Config
		dataplane                *rest.Resource
		expectedBootstrapRequest string
	}

	BeforeEach(func() {
		kuma_version.Build = kuma_version.BuildInfo{
			Version:   "0.0.1",
			GitTag:    "v0.0.1",
			GitCommit: "91ce236824a9d875601679aa80c63783fb0e8725",
			BuildDate: "2019-08-07T11:26:06Z",
		}
	})

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
		config, err := generator(fmt.Sprintf("http://localhost:%d", port), given.config, given.dataplane, EnvoyVersion{
			Build:   "hash/1.15.0/RELEASE",
			Version: "1.15.0",
		})

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
					dataplane: &rest.Resource{
						Meta: rest.ResourceMeta{
							Type: "Dataplane",
							Mesh: "demo",
							Name: "sample",
						},
					},
					expectedBootstrapRequest: flatJson(`
					{
					  "mesh": "demo",
					  "name": "sample",
					  "adminPort": 4321,
					  "dataplaneTokenPath": "/tmp/token",
					  "dataplaneResource": "{\"type\":\"Dataplane\",\"mesh\":\"demo\",\"name\":\"sample\",\"creationTime\":\"0001-01-01T00:00:00Z\",\"modificationTime\":\"0001-01-01T00:00:00Z\"}",
					  "version": {
						"kumaDp": {
						  "version": "0.0.1",
						  "gitTag": "v0.0.1",
						  "gitCommit": "91ce236824a9d875601679aa80c63783fb0e8725",
						  "buildDate": "2019-08-07T11:26:06Z"
						},
						"envoy": {
						  "version": "1.15.0",
						  "build": "hash/1.15.0/RELEASE"
						}
					  }
					}`),
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
					dataplane: &rest.Resource{
						Meta: rest.ResourceMeta{
							Type: "Dataplane",
							Mesh: "demo",
							Name: "sample",
						},
					},
					expectedBootstrapRequest: flatJson(`
                    {
                      "mesh": "demo",
                      "name": "sample",
                      "adminPort": 4321,
                      "dataplaneTokenPath": "/tmp/token",
                      "dataplaneResource": "{\"type\":\"Dataplane\",\"mesh\":\"demo\",\"name\":\"sample\",\"creationTime\":\"0001-01-01T00:00:00Z\",\"modificationTime\":\"0001-01-01T00:00:00Z\"}",
					  "version": {
						"kumaDp": {
						  "version": "0.0.1",
						  "gitTag": "v0.0.1",
						  "gitCommit": "91ce236824a9d875601679aa80c63783fb0e8725",
						  "buildDate": "2019-08-07T11:26:06Z"
						},
						"envoy": {
						  "version": "1.15.0",
						  "build": "hash/1.15.0/RELEASE"
						}
					  }
                    }`),
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
					dataplane: &rest.Resource{
						Meta: rest.ResourceMeta{
							Type: "Dataplane",
							Mesh: "demo",
							Name: "sample",
						},
					},
					expectedBootstrapRequest: flatJson(`
                    {
                      "mesh": "demo",
                      "name": "sample",
                      "dataplaneTokenPath": "/tmp/token",
                      "dataplaneResource": "{\"type\":\"Dataplane\",\"mesh\":\"demo\",\"name\":\"sample\",\"creationTime\":\"0001-01-01T00:00:00Z\",\"modificationTime\":\"0001-01-01T00:00:00Z\"}",
					  "version": {
						"kumaDp": {
						  "version": "0.0.1",
						  "gitTag": "v0.0.1",
						  "gitCommit": "91ce236824a9d875601679aa80c63783fb0e8725",
						  "buildDate": "2019-08-07T11:26:06Z"
						},
						"envoy": {
						  "version": "1.15.0",
						  "build": "hash/1.15.0/RELEASE"
						}
					  }
                    }`),
				}
			}()),
	)

	It("should retry when DP is not found", func() {
		// given
		mux := http.NewServeMux()
		server := httptest.NewServer(mux)
		defer server.Close()
		i := 0
		mux.HandleFunc("/bootstrap", func(writer http.ResponseWriter, req *http.Request) {
			defer GinkgoRecover()
			if i < 2 {
				writer.WriteHeader(404)
				i++
			} else {
				response, err := ioutil.ReadFile(filepath.Join("testdata", "remote-bootstrap-config.golden.yaml"))
				Expect(err).ToNot(HaveOccurred())
				_, err = writer.Write(response)
				Expect(err).ToNot(HaveOccurred())
			}
		})
		port, err := strconv.Atoi(strings.Split(server.Listener.Addr().String(), ":")[1])
		Expect(err).ToNot(HaveOccurred())

		// and
		generator := NewRemoteBootstrapGenerator(http.DefaultClient)

		// when
		cfg := kuma_dp.DefaultConfig()
		cfg.ControlPlane.Retry.Backoff = 10 * time.Millisecond
		_, err = generator(fmt.Sprintf("http://localhost:%d", port), cfg, &rest.Resource{
			Meta: rest.ResourceMeta{
				Type: "Dataplane",
				Mesh: "default",
				Name: "dp-1",
			},
		}, EnvoyVersion{})

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(cfg).ToNot(BeNil())
	})

	It("should return error when DP is not found", func() {
		// given
		mux := http.NewServeMux()
		server := httptest.NewServer(mux)
		defer server.Close()
		mux.HandleFunc("/bootstrap", func(writer http.ResponseWriter, req *http.Request) {
			defer GinkgoRecover()
			writer.WriteHeader(404)
		})
		port, err := strconv.Atoi(strings.Split(server.Listener.Addr().String(), ":")[1])
		Expect(err).ToNot(HaveOccurred())

		// and
		generator := NewRemoteBootstrapGenerator(http.DefaultClient)

		// when
		config := kuma_dp.DefaultConfig()
		config.ControlPlane.Retry.Backoff = 10 * time.Millisecond
		config.ControlPlane.Retry.MaxDuration = 100 * time.Millisecond
		_, err = generator(fmt.Sprintf("http://localhost:%d", port), config, &rest.Resource{
			Meta: rest.ResourceMeta{Mesh: "default", Name: "dp-1"},
		}, EnvoyVersion{})

		// then
		Expect(err).To(MatchError("retryable: Dataplane entity not found. If you are running on Universal please create a Dataplane entity on kuma-cp before starting kuma-dp. If you are running on Kubernetes, please check the kuma-cp logs to determine why the Dataplane entity could not be created by the automatic sidecar injection."))
	})
})

func flatJson(s string) string {
	noNewLines := strings.Trim(s, "\n")
	noSpaces := strings.Trim(noNewLines, " ")
	return noSpaces
}
