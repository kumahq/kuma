package envoy_test

import (
	"context"
	"encoding/json"
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

	. "github.com/kumahq/kuma/app/kuma-dp/pkg/dataplane/envoy"
	kuma_dp "github.com/kumahq/kuma/pkg/config/app/kuma-dp"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest/unversioned"
	rest_v1alpha1 "github.com/kumahq/kuma/pkg/core/resources/model/rest/v1alpha1"
	kuma_version "github.com/kumahq/kuma/pkg/version"
	"github.com/kumahq/kuma/pkg/xds/bootstrap/types"
)

var _ = Describe("Remote Bootstrap", func() {
	type testCase struct {
		config                   kuma_dp.Config
		dataplane                rest.Resource
		dynamicMetadata          map[string]string
		expectedBootstrapRequest string
		sidecarConfiguration     *types.KumaSidecarConfiguration
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
			Expect(req.Header.Get("accept")).To(Equal("application/json"))

			bootstrap, err := os.ReadFile(filepath.Join("testdata", "remote-bootstrap-config.golden.yaml"))
			Expect(err).ToNot(HaveOccurred())
			response := &types.BootstrapResponse{
				Bootstrap: bootstrap,
			}
			responseBytes, err := json.Marshal(response)
			Expect(err).ToNot(HaveOccurred())
			_, err = writer.Write(responseBytes)
			Expect(err).ToNot(HaveOccurred())
		})
		port, err := strconv.Atoi(strings.Split(server.Listener.Addr().String(), ":")[1])
		Expect(err).ToNot(HaveOccurred())

		// and
		generator := NewRemoteBootstrapGenerator(http.DefaultClient, "linux", []string{})

		// when
		params := BootstrapParams{
			Dataplane: given.dataplane,
			EnvoyVersion: EnvoyVersion{
				Build:   "hash/1.15.0/RELEASE",
				Version: "1.15.0",
			},
			DynamicMetadata: given.dynamicMetadata,
		}
		bootstrap, kumaSidecar, err := generator(context.Background(), fmt.Sprintf("http://localhost:%d", port), given.config, params)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(bootstrap).ToNot(BeNil())
		Expect(kumaSidecar).ToNot(BeNil())
	},
		Entry("should support port range with exactly 1 port",
			func() testCase {
				cfg := kuma_dp.DefaultConfig()
				cfg.Dataplane.Mesh = "demo"
				cfg.Dataplane.Name = "sample"
				cfg.DataplaneRuntime.Token = "token"

				return testCase{
					config:               cfg,
					sidecarConfiguration: &types.KumaSidecarConfiguration{},
					dataplane: &unversioned.Resource{
						Meta: rest_v1alpha1.ResourceMeta{
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
						  "build": "hash/1.15.0/RELEASE",
						  "kumaDpCompatible": false
						}
					  },
					  "caCert": "",
					  "dynamicMetadata": {
					    "test": "value"
					  },
					  "operatingSystem": "linux",
					  "features": [],
					  "resources": {
					    "maxHeapSizeBytes": 0
				      }
					}`,
				}
			}()),

		Entry("should read token from the file and include file path",
			func() testCase {
				cfg := kuma_dp.DefaultConfig()
				cfg.Dataplane.Mesh = "demo"
				cfg.Dataplane.Name = "sample"
				cfg.DataplaneRuntime.TokenPath = "testdata/token"

				return testCase{
					config:               cfg,
					sidecarConfiguration: &types.KumaSidecarConfiguration{},
					dataplane: &unversioned.Resource{
						Meta: rest_v1alpha1.ResourceMeta{
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
					  "dataplaneToken": "token1234",
                      "dataplaneTokenPath": "testdata/token",
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
                          "build": "hash/1.15.0/RELEASE",
                          "kumaDpCompatible": false
                        }
                      },
                      "caCert": "",
                      "dynamicMetadata": null,
                      "operatingSystem": "linux",
                      "features": [],
					  "resources": {
					    "maxHeapSizeBytes": 0
				      }
                    }`,
				}
			}()),

		Entry("should support port range with multiple ports (choose the lowest port)",
			func() testCase {
				cfg := kuma_dp.DefaultConfig()
				cfg.Dataplane.Mesh = "demo"
				cfg.Dataplane.Name = "sample"
				cfg.DataplaneRuntime.Token = "token"

				return testCase{
					config:               cfg,
					sidecarConfiguration: &types.KumaSidecarConfiguration{},
					dataplane: &unversioned.Resource{
						Meta: rest_v1alpha1.ResourceMeta{
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
                          "build": "hash/1.15.0/RELEASE",
                          "kumaDpCompatible": false
                        }
                      },
                      "caCert": "",
                      "dynamicMetadata": null,
                      "operatingSystem": "linux",
                      "features": [],
					  "resources": {
					    "maxHeapSizeBytes": 0
				      }
                    }`,
				}
			}()),
		Entry("should support empty port range",
			func() testCase {
				cfg := kuma_dp.DefaultConfig()
				cfg.Dataplane.Mesh = "demo"
				cfg.Dataplane.Name = "sample"
				cfg.DataplaneRuntime.Token = "token"

				return testCase{
					config:               cfg,
					sidecarConfiguration: &types.KumaSidecarConfiguration{},
					dataplane: &unversioned.Resource{
						Meta: rest_v1alpha1.ResourceMeta{
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
                          "build": "hash/1.15.0/RELEASE",
                          "kumaDpCompatible": false
                        }
                      },
                      "caCert": "",
					  "dynamicMetadata": null,
					  "operatingSystem": "linux",
					  "features": [],
					  "resources": {
					    "maxHeapSizeBytes": 0
				      }
                    }`,
				}
			}()),
	)

	It("should get configuration of kuma sidecar", func() {
		// given
		mux := http.NewServeMux()
		server := httptest.NewServer(mux)
		defer server.Close()
		mux.HandleFunc("/bootstrap", func(writer http.ResponseWriter, req *http.Request) {
			defer GinkgoRecover()
			Expect(req.Header.Get("accept")).To(Equal("application/json"))

			bootstrap, err := os.ReadFile(filepath.Join("testdata", "remote-bootstrap-config.golden.yaml"))
			Expect(err).ToNot(HaveOccurred())
			response := &types.BootstrapResponse{
				Bootstrap: bootstrap,
				KumaSidecarConfiguration: types.KumaSidecarConfiguration{
					Metrics: types.MetricsConfiguration{
						Aggregate: []types.Aggregate{
							{
								Address: "127.0.0.1",
								Name:    "my-app",
								Port:    123,
								Path:    "/stats",
							},
							{
								Address: "1.2.3.4",
								Name:    "my-app-2",
								Port:    12345,
								Path:    "/stats/2",
							},
						},
					},
				},
			}
			responseBytes, err := json.Marshal(response)
			Expect(err).ToNot(HaveOccurred())
			_, err = writer.Write(responseBytes)
			Expect(err).ToNot(HaveOccurred())
		})
		port, err := strconv.Atoi(strings.Split(server.Listener.Addr().String(), ":")[1])
		Expect(err).ToNot(HaveOccurred())

		// and
		generator := NewRemoteBootstrapGenerator(http.DefaultClient, "linux", []string{})

		// when
		cfg := kuma_dp.DefaultConfig()
		params := BootstrapParams{
			Dataplane: &unversioned.Resource{
				Meta: rest_v1alpha1.ResourceMeta{
					Type: "Dataplane",
					Mesh: "demo",
					Name: "sample",
				},
			},
		}

		bootstrap, kumaSidecarConfiguration, err := generator(context.Background(), fmt.Sprintf("http://localhost:%d", port), cfg, params)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(bootstrap).ToNot(BeNil())
		Expect(len(kumaSidecarConfiguration.Metrics.Aggregate)).To(Equal(2))
		Expect(kumaSidecarConfiguration.Metrics.Aggregate).To(ContainElements(types.Aggregate{
			Address: "127.0.0.1",
			Name:    "my-app",
			Port:    123,
			Path:    "/stats",
		}, types.Aggregate{
			Address: "1.2.3.4",
			Name:    "my-app-2",
			Port:    12345,
			Path:    "/stats/2",
		}))

	})

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
				bootstrap, err := os.ReadFile(filepath.Join("testdata", "remote-bootstrap-config.golden.yaml"))
				Expect(err).ToNot(HaveOccurred())
				response := &types.BootstrapResponse{
					Bootstrap: bootstrap,
				}
				responseBytes, err := json.Marshal(response)
				Expect(err).ToNot(HaveOccurred())
				_, err = writer.Write(responseBytes)
				Expect(err).ToNot(HaveOccurred())
			}
		})
		port, err := strconv.Atoi(strings.Split(server.Listener.Addr().String(), ":")[1])
		Expect(err).ToNot(HaveOccurred())

		// and
		generator := NewRemoteBootstrapGenerator(http.DefaultClient, "linux", []string{})

		// when
		cfg := kuma_dp.DefaultConfig()
		cfg.ControlPlane.Retry.Backoff = 10 * time.Millisecond
		params := BootstrapParams{
			Dataplane: &unversioned.Resource{
				Meta: rest_v1alpha1.ResourceMeta{
					Type: "Dataplane",
					Mesh: "default",
					Name: "dp-1",
				},
			},
		}
		_, _, err = generator(context.Background(), fmt.Sprintf("http://localhost:%d", port), cfg, params)

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
		generator := NewRemoteBootstrapGenerator(http.DefaultClient, "linux", []string{})

		// when
		config := kuma_dp.DefaultConfig()
		config.ControlPlane.Retry.Backoff = 10 * time.Millisecond
		config.ControlPlane.Retry.MaxDuration = 100 * time.Millisecond
		params := BootstrapParams{
			Dataplane: &unversioned.Resource{
				Meta: rest_v1alpha1.ResourceMeta{
					Type: "Dataplane",
					Mesh: "default",
					Name: "dp-1",
				},
			},
		}
		_, _, err = generator(context.Background(), fmt.Sprintf("http://localhost:%d", port), config, params)

		// then
		Expect(err).To(MatchError("Dataplane entity not found. If you are running on Universal please create a Dataplane entity on kuma-cp before starting kuma-dp or pass it to kuma-dp run --dataplane-file=/file. If you are running on Kubernetes, please check the kuma-cp logs to determine why the Dataplane entity could not be created by the automatic sidecar injection."))
	})
})
