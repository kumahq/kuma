package envoy_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/app/kuma-dp/pkg/dataplane/envoy"
	kuma_dp "github.com/kumahq/kuma/pkg/config/app/kuma-dp"
	config_types "github.com/kumahq/kuma/pkg/config/types"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest/unversioned"
	rest_v1alpha1 "github.com/kumahq/kuma/pkg/core/resources/model/rest/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/matchers"
	tproxy_dp "github.com/kumahq/kuma/pkg/transparentproxy/config/dataplane"
	kuma_version "github.com/kumahq/kuma/pkg/version"
	"github.com/kumahq/kuma/pkg/xds/bootstrap/types"
)

var _ = Describe("Remote Bootstrap", func() {
	type testCase struct {
		optsBuilder                  optsBuilder
		metadata                     map[string]string
		expectedBootstrapRequestFile string
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
			Expect(body).To(matchers.MatchGoldenJSON("testdata", given.expectedBootstrapRequestFile))
			Expect(req.Header.Get("accept")).To(Equal("application/json"))

			bootstrap, err := os.ReadFile(filepath.Join("testdata", "remote-bootstrap-config.response.yaml"))
			Expect(err).ToNot(HaveOccurred())
			response := &types.BootstrapResponse{
				Bootstrap: bootstrap,
			}
			responseBytes, err := json.Marshal(response)
			Expect(err).ToNot(HaveOccurred())
			_, err = writer.Write(responseBytes)
			Expect(err).ToNot(HaveOccurred())
		})

		// and
		bootstrapClient := NewRemoteBootstrapClient("linux")

		// when
		bootstrap, kumaSidecar, err := bootstrapClient.Fetch(
			context.Background(),
			given.optsBuilder.cpURL(server.URL).build(),
			given.metadata,
		)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(bootstrap).ToNot(BeNil())
		Expect(kumaSidecar).ToNot(BeNil())
	},
		Entry("includes TLS metrics config and custom metadata",
			testCase{
				optsBuilder:                  newOptsBuilder().metrics("/tmp/cert.pem", "/tmp/key.pem"),
				metadata:                     map[string]string{"test": "value"},
				expectedBootstrapRequestFile: "bootstrap-request-metrics-metadata.golden.json",
			},
		),

		Entry("reads token from file and sets file path in request",
			testCase{
				optsBuilder:                  newOptsBuilder().tokenPath("testdata/token"),
				expectedBootstrapRequestFile: "bootstrap-request-token-path.golden.json",
			},
		),

		Entry("includes transparent proxy configuration when provided",
			testCase{
				optsBuilder: newOptsBuilder().tproxy(tproxy_dp.DataplaneConfig{
					IPFamilyMode: "ipv4",
					Redirect: tproxy_dp.DataplaneRedirect{
						Inbound: tproxy_dp.DatalpaneTrafficFlow{
							Enabled: false,
							Port:    1234,
						},
						Outbound: tproxy_dp.DatalpaneTrafficFlow{
							Enabled: false,
							Port:    2345,
						},
						DNS: tproxy_dp.DatalpaneTrafficFlow{
							Enabled: false,
							Port:    3456,
						},
					},
				}),
				expectedBootstrapRequestFile: "bootstrap-request-transparent-proxy.golden.json",
			},
		),
	)

	It("should get configuration of kuma sidecar", func() {
		// given
		mux := http.NewServeMux()
		server := httptest.NewServer(mux)
		defer server.Close()
		mux.HandleFunc("/bootstrap", func(writer http.ResponseWriter, req *http.Request) {
			defer GinkgoRecover()
			Expect(req.Header.Get("accept")).To(Equal("application/json"))

			bootstrap, err := os.ReadFile(filepath.Join("testdata", "remote-bootstrap-config.response.yaml"))
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

		// and
		bootstrapClient := NewRemoteBootstrapClient("linux")

		// when
		bootstrap, kumaSidecarConfiguration, err := bootstrapClient.Fetch(
			context.Background(),
			newOptsBuilder().token("").cpURL(server.URL).build(),
			nil,
		)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(bootstrap).ToNot(BeNil())
		Expect(kumaSidecarConfiguration.Metrics.Aggregate).To(HaveLen(2))
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
				bootstrap, err := os.ReadFile(filepath.Join("testdata", "remote-bootstrap-config.response.yaml"))
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

		// and
		bootstrapClient := NewRemoteBootstrapClient("linux")

		// when
		bootstrap, _, err := bootstrapClient.Fetch(
			context.Background(),
			newOptsBuilder().
				mesh("default").
				name("dp-1").
				token("").
				retryBackoff(10*time.Millisecond).
				cpURL(server.URL).
				build(),
			nil,
		)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(bootstrap).ToNot(BeNil())
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

		// and
		bootstrapClient := NewRemoteBootstrapClient("linux")

		// when
		_, _, err := bootstrapClient.Fetch(
			context.Background(),
			newOptsBuilder().
				mesh("default").
				name("dp-1").
				token("").
				retryBackoff(10*time.Millisecond).
				maxDuration(100*time.Millisecond).
				cpURL(server.URL).build(),
			nil,
		)

		// then
		Expect(err).To(MatchError("Dataplane entity not found. If you are running on Universal please create a Dataplane entity on kuma-cp before starting kuma-dp or pass it to kuma-dp run --dataplane-file=/file. If you are running on Kubernetes, please check the kuma-cp logs to determine why the Dataplane entity could not be created by the automatic sidecar injection."))
	})
})

type optsBuilder Opts

func (b optsBuilder) mesh(mesh string) optsBuilder {
	b.Config.Dataplane.Mesh = mesh
	return b
}

func (b optsBuilder) name(name string) optsBuilder {
	b.Config.Dataplane.Name = name
	return b
}

func (b optsBuilder) token(token string) optsBuilder {
	b.Config.DataplaneRuntime.Token = token
	return b
}

func (b optsBuilder) tokenPath(tokenPath string) optsBuilder {
	b.Config.DataplaneRuntime.Token = ""
	b.Config.DataplaneRuntime.TokenPath = tokenPath
	return b
}

func (b optsBuilder) metrics(certPath, keyPath string) optsBuilder {
	b.Config.DataplaneRuntime.Metrics.CertPath = certPath
	b.Config.DataplaneRuntime.Metrics.KeyPath = keyPath
	return b
}

func (b optsBuilder) retryBackoff(duration time.Duration) optsBuilder {
	b.Config.ControlPlane.Retry.Backoff = config_types.Duration{Duration: duration}
	return b
}

func (b optsBuilder) maxDuration(duration time.Duration) optsBuilder {
	b.Config.ControlPlane.Retry.MaxDuration = config_types.Duration{Duration: duration}
	return b
}

func (b optsBuilder) cpURL(cpURL string) optsBuilder {
	b.Config.ControlPlane.URL = cpURL
	return b
}

func (b optsBuilder) tproxy(cfg tproxy_dp.DataplaneConfig) optsBuilder {
	b.Config.DataplaneRuntime.TransparentProxy = &cfg
	return b
}

func (b optsBuilder) build() Opts {
	return Opts{
		Config: b.Config,
		Dataplane: &unversioned.Resource{
			Meta: rest_v1alpha1.ResourceMeta{
				Type: "Dataplane",
				Mesh: b.Config.Dataplane.Mesh,
				Name: b.Config.Dataplane.Name,
			},
		},
	}
}

func newOptsBuilder() optsBuilder {
	cfg := kuma_dp.DefaultConfig()
	cfg.Dataplane.Mesh = "demo"
	cfg.Dataplane.Name = "sample"
	cfg.DataplaneRuntime.Token = "token"
	cfg.DataplaneRuntime.BinaryPath = filepath.Join("testdata", "envoy-mock.exit-0.sh")
	cfg.DataplaneRuntime.SocketDir = "/tmp"
	return optsBuilder{Config: cfg}
}
