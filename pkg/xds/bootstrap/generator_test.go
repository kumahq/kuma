package bootstrap_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	xds_config "github.com/kumahq/kuma/pkg/config/xds"
	bootstrap_config "github.com/kumahq/kuma/pkg/config/xds/bootstrap"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	. "github.com/kumahq/kuma/pkg/xds/bootstrap"
	"github.com/kumahq/kuma/pkg/xds/bootstrap/types"
)

var defaultVersion = types.Version{
	KumaDp: types.KumaDpVersion{
		Version:   "0.0.1",
		GitTag:    "v0.0.1",
		GitCommit: "91ce236824a9d875601679aa80c63783fb0e8725",
		BuildDate: "2019-08-07T11:26:06Z",
	},
	Envoy: types.EnvoyVersion{
		Build:   "hash/1.15.0/RELEASE",
		Version: "1.15.0",
	},
}

var _ = Describe("bootstrapGenerator", func() {

	var resManager core_manager.ResourceManager

	core.TempDir = func() string {
		return "/tmp"
	}

	BeforeEach(func() {
		resManager = core_manager.NewResourceManager(memory.NewStore())
		core.Now = func() time.Time {
			now, _ := time.Parse(time.RFC3339, "2018-07-17T16:05:36.995+00:00")
			return now
		}
	})

	defaultDataplane := func() *core_mesh.DataplaneResource {
		return &core_mesh.DataplaneResource{
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "8.8.8.8",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Port:        443,
							ServicePort: 8443,
							Tags: map[string]string{
								"kuma.io/service": "backend",
							},
						},
					},
					Admin: &mesh_proto.EnvoyAdmin{},
				},
			},
		}
	}

	BeforeEach(func() {
		// when
		err := resManager.Create(context.Background(), &core_mesh.MeshResource{
			Spec: &mesh_proto.Mesh{
				Metrics: &mesh_proto.Metrics{
					EnabledBackend: "prometheus-1",
					Backends: []*mesh_proto.MetricsBackend{
						{
							Name: "prometheus-1",
							Type: mesh_proto.MetricsPrometheusType,
						},
					},
				},
			},
		}, store.CreateByKey("mesh", model.NoMesh))
		// then
		Expect(err).ToNot(HaveOccurred())
	})

	type testCase struct {
		serverConfig       *bootstrap_config.BootstrapServerConfig
		proxyConfig        *xds_config.Proxy
		dataplane          func() *core_mesh.DataplaneResource
		dpAuthEnabled      bool
		useTokenPath       bool
		request            types.BootstrapRequest
		expectedConfigFile string
		hdsEnabled         bool
	}
	DescribeTable("should generate bootstrap configuration",
		func(given testCase) {
			// setup
			err := resManager.Create(context.Background(), given.dataplane(), store.CreateByKey("name.namespace", "mesh"))
			Expect(err).ToNot(HaveOccurred())

			proxyConfig := xds_config.DefaultProxyConfig()
			if given.proxyConfig != nil {
				proxyConfig = *given.proxyConfig
			}

			generator, err := NewDefaultBootstrapGenerator(resManager, given.serverConfig, proxyConfig, filepath.Join("..", "..", "..", "test", "certs", "server-cert.pem"), given.dpAuthEnabled, given.useTokenPath, given.hdsEnabled, 0, false)
			Expect(err).ToNot(HaveOccurred())

			// when
			bootstrapConfig, _, err := generator.Generate(context.Background(), given.request)

			// then
			Expect(err).ToNot(HaveOccurred())

			// and config is as expected
			actual, err := util_proto.ToYAML(bootstrapConfig)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchGoldenYAML(filepath.Join("testdata", given.expectedConfigFile)))
		},
		Entry("default config with minimal request", testCase{
			dpAuthEnabled: false,
			serverConfig: func() *bootstrap_config.BootstrapServerConfig {
				cfg := bootstrap_config.DefaultBootstrapServerConfig()
				cfg.Params.XdsHost = "localhost"
				cfg.Params.XdsPort = 5678
				return cfg
			}(),
			dataplane: defaultDataplane,
			request: types.BootstrapRequest{
				Mesh:    "mesh",
				Name:    "name.namespace",
				Version: defaultVersion,
			},
			expectedConfigFile: "generator.default-config-minimal-request.golden.yaml",
			hdsEnabled:         true,
		}),
		Entry("default config", testCase{
			dpAuthEnabled: true,
			serverConfig: func() *bootstrap_config.BootstrapServerConfig {
				cfg := bootstrap_config.DefaultBootstrapServerConfig()
				cfg.Params.XdsHost = "localhost"
				cfg.Params.XdsPort = 5678
				return cfg
			}(),
			dataplane: func() *core_mesh.DataplaneResource {
				dp := defaultDataplane()
				dp.Spec.Networking.Admin.Port = 1234
				return dp
			},
			request: types.BootstrapRequest{
				Mesh:           "mesh",
				Name:           "name.namespace",
				DataplaneToken: "token",
				Version:        defaultVersion,
				DNSPort:        53001,
				EmptyDNSPort:   53002,
			},
			expectedConfigFile: "generator.default-config.golden.yaml",
			hdsEnabled:         true,
		}),
		Entry("custom config with minimal request", testCase{
			dpAuthEnabled: false,
			serverConfig: func() *bootstrap_config.BootstrapServerConfig {
				return &bootstrap_config.BootstrapServerConfig{
					Params: &bootstrap_config.BootstrapParamsConfig{
						AdminAddress:       "192.168.0.1", // by default, Envoy Admin interface should listen on loopback address
						AdminAccessLogPath: "/var/log",
						XdsHost:            "localhost",
						XdsPort:            15678,
						XdsConnectTimeout:  2 * time.Second,
					},
				}
			}(),
			dataplane: func() *core_mesh.DataplaneResource {
				dp := defaultDataplane()
				dp.Spec.Networking.Admin.Port = 9902
				return dp
			},
			request: types.BootstrapRequest{
				Mesh:    "mesh",
				Name:    "name.namespace",
				Version: defaultVersion,
			},
			expectedConfigFile: "generator.custom-config-minimal-request.golden.yaml",
			hdsEnabled:         true,
		}),
		Entry("custom config", testCase{
			dpAuthEnabled: true,
			serverConfig: func() *bootstrap_config.BootstrapServerConfig {
				return &bootstrap_config.BootstrapServerConfig{
					Params: &bootstrap_config.BootstrapParamsConfig{
						AdminAddress:       "192.168.0.1", // by default, Envoy Admin interface should listen on loopback address
						AdminPort:          9902,          // by default, turn off Admin interface of Envoy
						AdminAccessLogPath: os.DevNull,
						XdsHost:            "localhost",
						XdsPort:            15678,
						XdsConnectTimeout:  2 * time.Second,
					},
				}
			}(),
			dataplane: defaultDataplane,
			request: types.BootstrapRequest{
				Mesh:            "mesh",
				Name:            "name.namespace",
				DataplaneToken:  "token",
				OperatingSystem: "windows",
				DynamicMetadata: map[string]string{
					"test": "value",
				},
				DataplaneResource: `
{
  "type": "Dataplane",
  "mesh": "mesh",
  "name": "name.namespace",
  "creationTime": "1970-01-01T00:00:00Z",
  "modificationTime": "1970-01-01T00:00:00Z",
  "networking": {
    "address": "127.0.0.1",
    "inbound": [
      {
        "port": 22022,
        "servicePort": 8443,
        "tags": {
          "kuma.io/protocol": "http2",
          "kuma.io/service": "backend"
        }
      },
    ],
    "admin": {
      "port": 1234
    }
  }
}`,
				Version: defaultVersion,
			},
			expectedConfigFile: "generator.custom-config.golden.yaml",
			hdsEnabled:         true,
		}),
		Entry("default config, kubernetes", testCase{
			dpAuthEnabled: true,
			serverConfig: func() *bootstrap_config.BootstrapServerConfig {
				cfg := bootstrap_config.DefaultBootstrapServerConfig()
				cfg.Params.XdsHost = "localhost"
				cfg.Params.XdsPort = 5678
				return cfg
			}(),
			dataplane: func() *core_mesh.DataplaneResource {
				dp := defaultDataplane()
				dp.Spec.Networking.Admin.Port = 1234
				return dp
			},
			request: types.BootstrapRequest{
				Mesh:           "mesh",
				Name:           "name.namespace",
				DataplaneToken: "token",
				Version:        defaultVersion,
			},
			expectedConfigFile: "generator.default-config.kubernetes.golden.yaml",
			hdsEnabled:         false,
		}),
		Entry("default config, kubernetes with IPv6", testCase{
			dpAuthEnabled: true,
			serverConfig: func() *bootstrap_config.BootstrapServerConfig {
				cfg := bootstrap_config.DefaultBootstrapServerConfig()
				cfg.Params.XdsHost = "fd00:a123::1"
				cfg.Params.XdsPort = 5678
				return cfg
			}(),
			dataplane: func() *core_mesh.DataplaneResource {
				dp := defaultDataplane()
				dp.Spec.Networking.Admin.Port = 1234
				return dp
			},
			request: types.BootstrapRequest{
				Mesh:           "mesh",
				Name:           "name.namespace",
				DataplaneToken: "token",
				Version:        defaultVersion,
			},
			expectedConfigFile: "generator.default-config.kubernetes.ipv6.golden.yaml",
			hdsEnabled:         false,
		}),
		Entry("default config, kubernetes with application metrics", testCase{
			dpAuthEnabled: true,
			serverConfig: func() *bootstrap_config.BootstrapServerConfig {
				cfg := bootstrap_config.DefaultBootstrapServerConfig()
				cfg.Params.XdsHost = "localhost"
				cfg.Params.XdsPort = 5678
				return cfg
			}(),
			dataplane: func() *core_mesh.DataplaneResource {
				dp := defaultDataplane()
				dp.Spec.Networking.Admin.Port = 1234
				dp.Spec.Metrics = &mesh_proto.MetricsBackend{
					Type: mesh_proto.MetricsPrometheusType,
					Conf: util_proto.MustToStruct(&mesh_proto.PrometheusMetricsBackendConfig{
						Aggregate: []*mesh_proto.PrometheusAggregateMetricsConfig{
							{
								Name: "app1",
								Port: 123,
								Path: "/stats",
							},
						},
					}),
				}
				return dp
			},
			request: types.BootstrapRequest{
				Mesh:           "mesh",
				Name:           "name.namespace",
				DataplaneToken: "token",
				Version:        defaultVersion,
			},
			expectedConfigFile: "generator.metrics-config.kubernetes.golden.yaml",
			hdsEnabled:         false,
		}),
		Entry("backwards compatibility, adminPort both in bootstrapRequest and in DPP resource", testCase{ // https://github.com/kumahq/kuma/issues/4002
			dpAuthEnabled: true,
			serverConfig: func() *bootstrap_config.BootstrapServerConfig {
				cfg := bootstrap_config.DefaultBootstrapServerConfig()
				cfg.Params.XdsHost = "localhost"
				cfg.Params.XdsPort = 5678
				return cfg
			}(),
			dataplane: func() *core_mesh.DataplaneResource {
				dp := defaultDataplane()
				dp.Spec.Networking.Admin.Port = 1234
				return dp
			},
			request: types.BootstrapRequest{
				Mesh:           "mesh",
				Name:           "name.namespace",
				DataplaneToken: "token",
				Version:        defaultVersion,
				DNSPort:        53001,
				EmptyDNSPort:   53002,
			},
			expectedConfigFile: "generator.default-config.golden.yaml",
			hdsEnabled:         true,
		}),
		Entry("default config with useTokenPath", testCase{
			dpAuthEnabled: true,
			serverConfig: func() *bootstrap_config.BootstrapServerConfig {
				cfg := bootstrap_config.DefaultBootstrapServerConfig()
				cfg.Params.XdsHost = "localhost"
				cfg.Params.XdsPort = 5678
				return cfg
			}(),
			dataplane: func() *core_mesh.DataplaneResource {
				dp := defaultDataplane()
				dp.Spec.Networking.Admin.Port = 1234
				return dp
			},
			request: types.BootstrapRequest{
				Mesh:               "mesh",
				Name:               "name.namespace",
				DataplaneToken:     "token",
				Version:            defaultVersion,
				DNSPort:            53001,
				EmptyDNSPort:       53002,
				DataplaneTokenPath: "/path/to/file",
			},
			expectedConfigFile: "generator.default-config-token-path.golden.yaml",
			hdsEnabled:         true,
			useTokenPath:       true,
		}),
		Entry("gateway settings", testCase{
			dpAuthEnabled: true,
			serverConfig: func() *bootstrap_config.BootstrapServerConfig {
				cfg := bootstrap_config.DefaultBootstrapServerConfig()
				cfg.Params.XdsHost = "localhost"
				cfg.Params.XdsPort = 5678
				return cfg
			}(),
			proxyConfig: func() *xds_config.Proxy {
				cfg := xds_config.DefaultProxyConfig()
				cfg.Gateway.GlobalDownstreamMaxConnections = 35678
				return &cfg
			}(),
			dataplane: func() *core_mesh.DataplaneResource {
				return &core_mesh.DataplaneResource{
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Address: "8.8.8.8",
							Gateway: &mesh_proto.Dataplane_Networking_Gateway{
								Type: mesh_proto.Dataplane_Networking_Gateway_BUILTIN,
								Tags: map[string]string{
									mesh_proto.ServiceTag: "gateway",
								},
							},
							Admin: &mesh_proto.EnvoyAdmin{},
						},
					},
				}
			},
			request: types.BootstrapRequest{
				Mesh:               "mesh",
				Name:               "name.namespace",
				DataplaneToken:     "token",
				Version:            defaultVersion,
				DNSPort:            53001,
				EmptyDNSPort:       53002,
				DataplaneTokenPath: "/path/to/file",
			},
			expectedConfigFile: "generator.gateway.golden.yaml",
			hdsEnabled:         true,
			useTokenPath:       true,
		}),
	)

	type errTestCase struct {
		request  types.BootstrapRequest
		expected string
	}
	DescribeTable("should fail bootstrap",
		func(given errTestCase) {
			// given
			err := resManager.Create(context.Background(), defaultDataplane(), store.CreateByKey("name.namespace", "mesh"))
			Expect(err).ToNot(HaveOccurred())

			cfg := bootstrap_config.DefaultBootstrapServerConfig()
			proxyCfg := xds_config.DefaultProxyConfig()

			generator, err := NewDefaultBootstrapGenerator(resManager, cfg, proxyCfg, filepath.Join("..", "..", "..", "test", "certs", "server-cert.pem"), false, false, true, 9901, false)
			Expect(err).ToNot(HaveOccurred())

			// when
			_, _, err = generator.Generate(context.Background(), given.request)
			// then
			Expect(err).To(HaveOccurred())
			// and
			fmt.Println(err.Error())
			Expect(err.Error()).To(Equal(given.expected))
		},
		Entry("due to invalid hostname", errTestCase{
			request: types.BootstrapRequest{
				Mesh: "mesh",
				Name: "name-3.namespace",
				Host: "kuma.internal",
			},
			expected: `A data plane proxy is trying to connect to the control plane using "kuma.internal" address, but the certificate in the control plane has the following SANs ["fd00:a123::1" "localhost"]. Either change the --cp-address in kuma-dp to one of those or execute the following steps:
1) Generate a new certificate with the address you are trying to use. It is recommended to use trusted Certificate Authority, but you can also generate self-signed certificates using 'kumactl generate tls-certificate --type=server --cp-hostname=kuma.internal'
2) Set KUMA_GENERAL_TLS_CERT_FILE and KUMA_GENERAL_TLS_KEY_FILE or the equivalent in Kuma CP config file to the new certificate.
3) Restart the control plane to read the new certificate and start kuma-dp.`,
		}),
		Entry("when CaCert is not a CA and EnvoyGRPC is used", errTestCase{
			request: types.BootstrapRequest{
				Host:           "localhost",
				Mesh:           "mesh",
				Name:           "name.namespace",
				DataplaneToken: "token",
				CaCert: `
-----BEGIN CERTIFICATE-----
MIIDdzCCAl+gAwIBAgIJAPHcHHoejP+XMA0GCSqGSIb3DQEBCwUAMD4xCzAJBgNV
BAYTAlBMMQ8wDQYDVQQIDAZXYXJzYXcxDzANBgNVBAcMBldhcnNhdzENMAsGA1UE
CgwES29uZzAeFw0yMTAzMzAwOTEwMTFaFw0yMzA3MDMwOTEwMTFaMC8xCzAJBgNV
BAYTAlBMMQ8wDQYDVQQIDAZXYXJzYXcxDzANBgNVBAcMBldhcnNhdzCCASIwDQYJ
KoZIhvcNAQEBBQADggEPADCCAQoCggEBAMOxLuQiRSsDKb+E/iliMN0ME1ENxx6v
S362cmyL6pCS6HdsJnOiCeAiiezRdotf7pD87DkwLrAI2v6IOEueXmXu/pRkZZdj
GFdYOJ0j28Qg79VfhLZPGZrATowUkmNfWFuX7gyjButP5+M6yMEm8piKkMgYtj8H
13Jj5GBazYojBdVkdC7VCRjwiF3oudDC+I0f5RFwqrU89zfLf8fIYn0waioUZKT9
W48oVmRw2SqYFf5O+T+EY3mcSWRNrzweZX7YdFvHFJLSglkmn7275cdwqle68iZn
xbVn7MW5nlp5W0ONAFLB3JJ7TRee2o8P9CkiuqG+ppmMPQq5zWPGuxUCAwEAAaOB
hjCBgzBYBgNVHSMEUTBPoUKkQDA+MQswCQYDVQQGEwJQTDEPMA0GA1UECAwGV2Fy
c2F3MQ8wDQYDVQQHDAZXYXJzYXcxDTALBgNVBAoMBEtvbmeCCQDpKl9mxhgHFzAJ
BgNVHRMEAjAAMAsGA1UdDwQEAwIE8DAPBgNVHREECDAGhwTAqAANMA0GCSqGSIb3
DQEBCwUAA4IBAQCHU5JyuMwayeVBVSOnGw8A9ugrGfyHy4nN+vK+IjkyPaDynyob
i1mXzK1JDn2koHqRRlSGQGy/eJdHRPxUj8+VzyIbCVqpiiOYxC2tQUQ5BhVGC08u
oCZcflyypSej2QVYtj83ty8ty1EFSdO8v23oPhzVSjc+SkF5c+q326piXf+a5wWh
uAxW1XJnTaqAFhGR9c0zRCrbz86yQTsdFAm1UVMMucnZjNpWL4pHLJC6FCiOO17q
w/vjIriD0mGwwccxbojmEHq4rO4ZrjQNmwvOgxoL2dTm/L9Smr6RXmIgu/0Pnrlq
7RLK1pnDttr4brFafbIvWIBvshe2hoCT6jBW
-----END CERTIFICATE-----
`,
			},
			expected: `A data plane proxy is trying to verify the control plane using the certificate which is not a certificate authority (basic constraint 'CA' is set to 'false').
Provide CA that was used to sign a certificate used in the control plane by using 'kuma-dp run --ca-cert-file=file' or via KUMA_CONTROL_PLANE_CA_CERT_FILE`,
		}),
	)

	It("should override configuration from Mesh", func() {
		// given
		err := resManager.Create(context.Background(), &core_mesh.MeshResource{
			Spec: &mesh_proto.Mesh{
				Metrics: &mesh_proto.Metrics{
					EnabledBackend: "prometheus-1",
					Backends: []*mesh_proto.MetricsBackend{
						{
							Name: "prometheus-1",
							Type: mesh_proto.MetricsPrometheusType,
							Conf: util_proto.MustToStruct(&mesh_proto.PrometheusMetricsBackendConfig{
								Aggregate: []*mesh_proto.PrometheusAggregateMetricsConfig{
									{
										Name: "opa",
										Port: 123,
										Path: "/mesh/config",
									},
									{
										Name: "dp-disabled",
										Port: 999,
										Path: "/stats/default",
									},
								},
							}),
						},
					},
				},
			},
		}, store.CreateByKey("metrics", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		// and
		dataplane := &core_mesh.DataplaneResource{
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "8.8.8.8",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Port:        443,
							ServicePort: 8443,
							Tags: map[string]string{
								"kuma.io/service": "backend",
							},
						},
					},
					TransparentProxying: &mesh_proto.Dataplane_Networking_TransparentProxying{
						RedirectPortInbound:  12345,
						RedirectPortOutbound: 12346,
					},
					Admin: &mesh_proto.EnvoyAdmin{},
				},
				Metrics: &mesh_proto.MetricsBackend{
					Type: mesh_proto.MetricsPrometheusType,
					Conf: util_proto.MustToStruct(&mesh_proto.PrometheusMetricsBackendConfig{
						Aggregate: []*mesh_proto.PrometheusAggregateMetricsConfig{
							{
								Name:    "dp-disabled",
								Enabled: util_proto.Bool(false),
							},
							{
								Name: "app",
								Port: 12,
								Path: "/dp/override",
							},
						},
					}),
				},
			},
		}

		config := func() *bootstrap_config.BootstrapServerConfig {
			cfg := bootstrap_config.DefaultBootstrapServerConfig()
			cfg.Params.XdsHost = "localhost"
			cfg.Params.XdsPort = 5678
			return cfg
		}
		proxyCfg := xds_config.DefaultProxyConfig()

		err = resManager.Create(context.Background(), dataplane, store.CreateByKey("name.namespace", "metrics"))
		Expect(err).ToNot(HaveOccurred())

		generator, err := NewDefaultBootstrapGenerator(resManager, config(), proxyCfg, filepath.Join("..", "..", "..", "test", "certs", "server-cert.pem"), true, false, false, 0, false)
		Expect(err).ToNot(HaveOccurred())

		// when
		bootstrapConfig, configParam, err := generator.Generate(context.Background(), types.BootstrapRequest{
			Mesh:           "metrics",
			Name:           "name.namespace",
			DataplaneToken: "token",
			Version:        defaultVersion,
		})

		// then
		Expect(err).ToNot(HaveOccurred())

		// and config is as expected
		_, err = util_proto.ToYAML(bootstrapConfig)
		Expect(err).ToNot(HaveOccurred())
		Expect(configParam.NetworkingConfig.IsUsingTransparentProxy).To(BeTrue())
		Expect(configParam.AggregateMetricsConfig).To(ContainElements([]AggregateMetricsConfig{
			{
				Address: "8.8.8.8",
				Name:    "opa",
				Path:    "/mesh/config",
				Port:    123,
			},
			{
				Address: "8.8.8.8",
				Name:    "app",
				Path:    "/dp/override",
				Port:    12,
			},
		}))
	})

	It("should take configuration from Mesh when service do not define", func() {
		// given
		err := resManager.Create(context.Background(), &core_mesh.MeshResource{
			Spec: &mesh_proto.Mesh{
				Metrics: &mesh_proto.Metrics{
					EnabledBackend: "prometheus-1",
					Backends: []*mesh_proto.MetricsBackend{
						{
							Name: "prometheus-1",
							Type: mesh_proto.MetricsPrometheusType,
							Conf: util_proto.MustToStruct(&mesh_proto.PrometheusMetricsBackendConfig{
								Aggregate: []*mesh_proto.PrometheusAggregateMetricsConfig{
									{
										Name: "opa",
										Port: 123,
										Path: "/mesh/opa",
									},
									{
										Name: "app",
										Port: 999,
										Path: "/mesh/app",
									},
								},
							}),
						},
					},
				},
			},
		}, store.CreateByKey("metrics", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		// and
		dataplane := &core_mesh.DataplaneResource{
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "8.8.8.8",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Port:        443,
							ServicePort: 8443,
							Tags: map[string]string{
								"kuma.io/service": "backend",
							},
						},
					},
					Admin: &mesh_proto.EnvoyAdmin{},
				},
			},
		}

		config := func() *bootstrap_config.BootstrapServerConfig {
			cfg := bootstrap_config.DefaultBootstrapServerConfig()
			cfg.Params.XdsHost = "localhost"
			cfg.Params.XdsPort = 5678
			return cfg
		}
		proxyCfg := xds_config.DefaultProxyConfig()

		err = resManager.Create(context.Background(), dataplane, store.CreateByKey("name.namespace", "metrics"))
		Expect(err).ToNot(HaveOccurred())

		generator, err := NewDefaultBootstrapGenerator(resManager, config(), proxyCfg, filepath.Join("..", "..", "..", "test", "certs", "server-cert.pem"), true, false, false, 0, false)
		Expect(err).ToNot(HaveOccurred())

		// when
		bootstrapConfig, configParam, err := generator.Generate(context.Background(), types.BootstrapRequest{
			Mesh:           "metrics",
			Name:           "name.namespace",
			DataplaneToken: "token",
			Version:        defaultVersion,
		})

		// then
		Expect(err).ToNot(HaveOccurred())

		// and config is as expected
		_, err = util_proto.ToYAML(bootstrapConfig)
		Expect(err).ToNot(HaveOccurred())
		Expect(configParam.NetworkingConfig.IsUsingTransparentProxy).To(BeFalse())
		Expect(configParam.AggregateMetricsConfig).To(Equal([]AggregateMetricsConfig{
			{
				Address: "8.8.8.8",
				Name:    "opa",
				Path:    "/mesh/opa",
				Port:    123,
			},
			{
				Address: "8.8.8.8",
				Name:    "app",
				Path:    "/mesh/app",
				Port:    999,
			},
		}))
	})
})
