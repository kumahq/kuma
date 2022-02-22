package envoy

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kuma_dp "github.com/kumahq/kuma/pkg/config/app/kuma-dp"
	config_types "github.com/kumahq/kuma/pkg/config/types"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	kuma_version "github.com/kumahq/kuma/pkg/version"
)

var _ = Describe("Remote Bootstrap", func() {

	type testCase struct {
		config                   kuma_dp.Config
		dataplane                *rest.Resource
		dynamicMetadata          map[string]string
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
			body, err := io.ReadAll(req.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(body).To(MatchJSON(given.expectedBootstrapRequest))

			response, err := os.ReadFile(filepath.Join("testdata", "remote-bootstrap-config.golden.yaml"))
			Expect(err).ToNot(HaveOccurred())
			_, err = writer.Write(response)
			Expect(err).ToNot(HaveOccurred())
		})
		port, err := strconv.Atoi(strings.Split(server.Listener.Addr().String(), ":")[1])
		Expect(err).ToNot(HaveOccurred())

		// and
		generator := NewRemoteBootstrapGenerator(http.DefaultClient)

		// when
		params := BootstrapParams{
			Dataplane: given.dataplane,
			EnvoyVersion: EnvoyVersion{
				Build:   "hash/1.15.0/RELEASE",
				Version: "1.15.0",
			},
			DynamicMetadata: given.dynamicMetadata,
		}
		_, config, err := generator(fmt.Sprintf("http://localhost:%d", port), given.config, params)

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
				cfg.DataplaneRuntime.Token = "token"

				return testCase{
					config: cfg,
					dataplane: &rest.Resource{
						Meta: rest.ResourceMeta{
							Type: "Dataplane",
							Mesh: "demo",
							Name: "sample",
						},
					},
					dynamicMetadata: map[string]string{
						"test": "value",
					},
					expectedBootstrapRequest: `
					{
					  "mesh": "demo",
					  "name": "sample",
					  "proxyType": "dataplane",
					  "adminPort": 4321,
					  "dataplaneToken": "token",
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
					  },
					  "caCert": "",
					  "dynamicMetadata": {
					    "test": "value"
					  },
                      "bootstrapVersion": "3"
					}`,
				}
			}()),

		Entry("should support port range with multiple ports (choose the lowest port)",
			func() testCase {
				cfg := kuma_dp.DefaultConfig()
				cfg.Dataplane.Mesh = "demo"
				cfg.Dataplane.Name = "sample"
				cfg.Dataplane.AdminPort = config_types.MustPortRange(4321, 8765) // port range
				cfg.DataplaneRuntime.Token = "token"

				return testCase{
					config: cfg,
					dataplane: &rest.Resource{
						Meta: rest.ResourceMeta{
							Type: "Dataplane",
							Mesh: "demo",
							Name: "sample",
						},
					},
					expectedBootstrapRequest: `
                    {
                      "mesh": "demo",
                      "name": "sample",
                      "proxyType": "dataplane",
                      "adminPort": 4321,
                      "dataplaneToken": "token",
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
                      },
                      "caCert": "",
                      "dynamicMetadata": null,
                      "bootstrapVersion": "3"
                    }`,
				}
			}()),
		Entry("should support empty port range",
			func() testCase {
				cfg := kuma_dp.DefaultConfig()
				cfg.Dataplane.Mesh = "demo"
				cfg.Dataplane.Name = "sample"
				cfg.Dataplane.AdminPort = config_types.PortRange{} // empty port range
				cfg.DataplaneRuntime.Token = "token"

				return testCase{
					config: cfg,
					dataplane: &rest.Resource{
						Meta: rest.ResourceMeta{
							Type: "Dataplane",
							Mesh: "demo",
							Name: "sample",
						},
					},
					expectedBootstrapRequest: `
                    {
                      "mesh": "demo",
                      "name": "sample",
                      "proxyType": "dataplane",
                      "dataplaneToken": "token",
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
                      },
                      "caCert": "",
					  "dynamicMetadata": null,
                      "bootstrapVersion": "3"
                    }`,
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
				response, err := os.ReadFile(filepath.Join("testdata", "remote-bootstrap-config.golden.yaml"))
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
		params := BootstrapParams{
			Dataplane: &rest.Resource{
				Meta: rest.ResourceMeta{
					Type: "Dataplane",
					Mesh: "default",
					Name: "dp-1",
				},
			},
		}
		_, _, err = generator(fmt.Sprintf("http://localhost:%d", port), cfg, params)

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
		params := BootstrapParams{
			Dataplane: &rest.Resource{
				Meta: rest.ResourceMeta{
					Type: "Dataplane",
					Mesh: "default",
					Name: "dp-1",
				},
			},
		}
		_, _, err = generator(fmt.Sprintf("http://localhost:%d", port), config, params)

		// then
		Expect(err).To(MatchError("Dataplane entity not found. If you are running on Universal please create a Dataplane entity on kuma-cp before starting kuma-dp or pass it to kuma-dp run --dataplane-file=/file. If you are running on Kubernetes, please check the kuma-cp logs to determine why the Dataplane entity could not be created by the automatic sidecar injection."))
	})
})
