package bootstrap_test

import (
	"context"
	"path/filepath"
	"time"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	. "github.com/kumahq/kuma/pkg/xds/bootstrap"

	bootstrap_config "github.com/kumahq/kuma/pkg/config/xds/bootstrap"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/bootstrap/types"
)

var _ = Describe("bootstrapGenerator", func() {

	var resManager core_manager.ResourceManager

	defaultVersion := types.Version{
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

	BeforeEach(func() {
		resManager = core_manager.NewResourceManager(memory.NewStore())
		core.Now = func() time.Time {
			now, _ := time.Parse(time.RFC3339, "2018-07-17T16:05:36.995+00:00")
			return now
		}
	})

	BeforeEach(func() {
		// given
		dataplane := mesh.DataplaneResource{
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
				},
			},
		}

		// when
		err := resManager.Create(context.Background(), mesh.NewMeshResource(), store.CreateByKey("mesh", model.NoMesh))
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		err = resManager.Create(context.Background(), &dataplane, store.CreateByKey("name.namespace", "mesh"))
		// then
		Expect(err).ToNot(HaveOccurred())
	})

	type testCase struct {
		config                   func() *bootstrap_config.BootstrapServerConfig
		dpAuthEnabled            bool
		request                  types.BootstrapRequest
		expectedConfigFile       string
		expectedBootstrapVersion types.BootstrapVersion
		hdsEnabled               bool
	}
	DescribeTable("should generate bootstrap configuration",
		func(given testCase) {
			// setup
			generator, err := NewDefaultBootstrapGenerator(resManager, given.config(), filepath.Join("..", "..", "..", "test", "certs", "server-cert.pem"), given.dpAuthEnabled, given.hdsEnabled)
			Expect(err).ToNot(HaveOccurred())

			// when
			bootstrapConfig, version, err := generator.Generate(context.Background(), given.request)

			// then
			Expect(err).ToNot(HaveOccurred())

			// and config is as expected
			actual, err := util_proto.ToYAML(bootstrapConfig)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchGoldenYAML(filepath.Join("testdata", given.expectedConfigFile)))

			// and
			Expect(version).To(Equal(given.expectedBootstrapVersion))
		},
		Entry("default config with minimal request", testCase{
			dpAuthEnabled: false,
			config: func() *bootstrap_config.BootstrapServerConfig {
				cfg := bootstrap_config.DefaultBootstrapServerConfig()
				cfg.Params.XdsHost = "localhost"
				cfg.Params.XdsPort = 5678
				cfg.APIVersion = envoy_common.APIV3
				return cfg
			},
			request: types.BootstrapRequest{
				Mesh:    "mesh",
				Name:    "name.namespace",
				Version: defaultVersion,
			},
			expectedConfigFile:       "generator.default-config-minimal-request.golden.yaml",
			expectedBootstrapVersion: types.BootstrapV3,
			hdsEnabled:               true,
		}),
		Entry("default config", testCase{
			dpAuthEnabled: true,
			config: func() *bootstrap_config.BootstrapServerConfig {
				cfg := bootstrap_config.DefaultBootstrapServerConfig()
				cfg.Params.XdsHost = "localhost"
				cfg.Params.XdsPort = 5678
				cfg.APIVersion = envoy_common.APIV3
				return cfg
			},
			request: types.BootstrapRequest{
				Mesh:           "mesh",
				Name:           "name.namespace",
				AdminPort:      1234,
				DataplaneToken: "token",
				Version:        defaultVersion,
				DNSPort:        53001,
				EmptyDNSPort:   53002,
			},
			expectedConfigFile:       "generator.default-config.golden.yaml",
			expectedBootstrapVersion: types.BootstrapV3,
			hdsEnabled:               true,
		}),
		Entry("custom config with minimal request", testCase{
			dpAuthEnabled: false,
			config: func() *bootstrap_config.BootstrapServerConfig {
				return &bootstrap_config.BootstrapServerConfig{
					Params: &bootstrap_config.BootstrapParamsConfig{
						AdminAddress:       "192.168.0.1", // by default, Envoy Admin interface should listen on loopback address
						AdminPort:          9902,          // by default, turn off Admin interface of Envoy
						AdminAccessLogPath: "/var/log",
						XdsHost:            "localhost",
						XdsPort:            15678,
						XdsConnectTimeout:  2 * time.Second,
					},
					APIVersion: envoy_common.APIV3,
				}
			},
			request: types.BootstrapRequest{
				Mesh:             "mesh",
				Name:             "name.namespace",
				Version:          defaultVersion,
				BootstrapVersion: types.BootstrapV3,
			},
			expectedConfigFile:       "generator.custom-config-minimal-request.golden.yaml",
			expectedBootstrapVersion: types.BootstrapV3,
			hdsEnabled:               true,
		}),
		Entry("custom config", testCase{
			dpAuthEnabled: true,
			config: func() *bootstrap_config.BootstrapServerConfig {
				return &bootstrap_config.BootstrapServerConfig{
					Params: &bootstrap_config.BootstrapParamsConfig{
						AdminAddress:       "192.168.0.1", // by default, Envoy Admin interface should listen on loopback address
						AdminPort:          9902,          // by default, turn off Admin interface of Envoy
						AdminAccessLogPath: "/var/log",
						XdsHost:            "localhost",
						XdsPort:            15678,
						XdsConnectTimeout:  2 * time.Second,
					},
					APIVersion: envoy_common.APIV3,
				}
			},
			request: types.BootstrapRequest{
				Mesh:               "mesh",
				Name:               "name.namespace",
				AdminPort:          1234,
				DataplaneTokenPath: "/tmp/token",
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
    ]
  }
}`,
				Version: defaultVersion,
			},
			expectedConfigFile:       "generator.custom-config.golden.yaml",
			expectedBootstrapVersion: types.BootstrapV3,
			hdsEnabled:               true,
		}),
		Entry("default config, kubernetes", testCase{
			dpAuthEnabled: true,
			config: func() *bootstrap_config.BootstrapServerConfig {
				cfg := bootstrap_config.DefaultBootstrapServerConfig()
				cfg.Params.XdsHost = "localhost"
				cfg.Params.XdsPort = 5678
				cfg.APIVersion = envoy_common.APIV3
				return cfg
			},
			request: types.BootstrapRequest{
				Mesh:               "mesh",
				Name:               "name.namespace",
				AdminPort:          1234,
				DataplaneTokenPath: "/tmp/token",
				Version:            defaultVersion,
			},
			expectedConfigFile:       "generator.default-config.kubernetes.golden.yaml",
			expectedBootstrapVersion: types.BootstrapV3,
			hdsEnabled:               false,
		}),
		Entry("default config, kubernetes with IPv6", testCase{
			dpAuthEnabled: true,
			config: func() *bootstrap_config.BootstrapServerConfig {
				cfg := bootstrap_config.DefaultBootstrapServerConfig()
				cfg.Params.XdsHost = "fd00:a123::1"
				cfg.Params.XdsPort = 5678
				cfg.APIVersion = envoy_common.APIV3
				return cfg
			},
			request: types.BootstrapRequest{
				Mesh:               "mesh",
				Name:               "name.namespace",
				AdminPort:          1234,
				DataplaneTokenPath: "/tmp/token",
				Version:            defaultVersion,
			},
			expectedConfigFile:       "generator.default-config.kubernetes.ipv6.golden.yaml",
			expectedBootstrapVersion: types.BootstrapV3,
			hdsEnabled:               false,
		}),
	)

	It("should fail bootstrap configuration due to conflicting port in inbound", func() {
		// setup
		dataplane := mesh.DataplaneResource{
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "8.8.8.8",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Address:     "127.0.0.1",
							Port:        9901,
							ServicePort: 8443,
							Tags: map[string]string{
								"kuma.io/service": "backend",
							},
						},
					},
					Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
						{
							Address: "1.1.1.1",
							Port:    9000,
							Service: "redis",
						},
					},
				},
			},
		}
		// when
		err := resManager.Create(context.Background(), &dataplane, store.CreateByKey("name-1.namespace", "mesh"))
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		dataplane.Spec.Networking.Address = "127.0.0.1"
		dataplane.Spec.Networking.Inbound[0].Address = ""
		err = resManager.Create(context.Background(), &dataplane, store.CreateByKey("name-2.namespace", "mesh"))
		// then
		Expect(err).ToNot(HaveOccurred())

		// given
		cfg := bootstrap_config.DefaultBootstrapServerConfig()
		cfg.Params.XdsHost = "localhost"
		cfg.Params.XdsPort = 5678

		generator, err := NewDefaultBootstrapGenerator(resManager, cfg, filepath.Join("..", "..", "..", "test", "certs", "server-cert.pem"), false, true)
		Expect(err).ToNot(HaveOccurred())
		request := types.BootstrapRequest{
			Mesh:      "mesh",
			Name:      "name-1.namespace",
			AdminPort: 9901,
		}

		// when
		_, _, err = generator.Generate(context.Background(), request)
		// then
		Expect(err).To(HaveOccurred())
		// and
		Expect(err.Error()).To(Equal("Resource precondition failed: Port 9901 requested as both admin and inbound port."))

		request = types.BootstrapRequest{
			Mesh:      "mesh",
			Name:      "name-2.namespace",
			AdminPort: 9901,
		}

		// when
		_, _, err = generator.Generate(context.Background(), request)
		// then
		Expect(err).To(HaveOccurred())
		// and
		Expect(err.Error()).To(Equal("Resource precondition failed: Port 9901 requested as both admin and inbound port."))

	})

	It("should fail bootstrap configuration due to conflicting port in outbound", func() {
		// setup
		dataplane := mesh.DataplaneResource{
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
					Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
						{
							Address: "127.0.0.1",
							Port:    9901,
							Service: "redis",
						},
					},
				},
			},
		}
		// when
		err := resManager.Create(context.Background(), &dataplane, store.CreateByKey("name-3.namespace", "mesh"))
		// then
		Expect(err).ToNot(HaveOccurred())

		// given
		cfg := bootstrap_config.DefaultBootstrapServerConfig()
		cfg.Params.XdsHost = "localhost"
		cfg.Params.XdsPort = 5678

		generator, err := NewDefaultBootstrapGenerator(resManager, cfg, filepath.Join("..", "..", "..", "test", "certs", "server-cert.pem"), false, true)
		Expect(err).ToNot(HaveOccurred())
		request := types.BootstrapRequest{
			Mesh:      "mesh",
			Name:      "name-3.namespace",
			AdminPort: 9901,
		}

		// when
		_, _, err = generator.Generate(context.Background(), request)
		// then
		Expect(err).To(HaveOccurred())
		// and
		Expect(err.Error()).To(Equal("Resource precondition failed: Port 9901 requested as both admin and outbound port."))
	})

	type errTestCase struct {
		request  types.BootstrapRequest
		expected string
	}
	DescribeTable("should fail bootstrap",
		func(given errTestCase) {
			// given
			cfg := bootstrap_config.DefaultBootstrapServerConfig()

			generator, err := NewDefaultBootstrapGenerator(resManager, cfg, filepath.Join("..", "..", "..", "test", "certs", "server-cert.pem"), false, true)
			Expect(err).ToNot(HaveOccurred())

			// when
			_, _, err = generator.Generate(context.Background(), given.request)
			// then
			Expect(err).To(HaveOccurred())
			// and
			Expect(err.Error()).To(Equal(given.expected))
		},
		Entry("due to invalid hostname", errTestCase{
			request: types.BootstrapRequest{
				Mesh:      "mesh",
				Name:      "name-3.namespace",
				AdminPort: 9901,
				Host:      "kuma.internal",
			},
			expected: `A data plane proxy is trying to connect to the control plane using "kuma.internal" address, but the certificate in the control plane has the following SANs ["fd00:a123::1" "localhost"]. Either change the --cp-address in kuma-dp to one of those or execute the following steps:
1) Generate a new certificate with the address you are trying to use. It is recommended to use trusted Certificate Authority, but you can also generate self-signed certificates using 'kumactl generate tls-certificate --type=server --cp-hostname=kuma.internal'
2) Set KUMA_GENERAL_TLS_CERT_FILE and KUMA_GENERAL_TLS_KEY_FILE or the equivalent in Kuma CP config file to the new certificate.
3) Restart the control plane to read the new certificate and start kuma-dp.`,
		}),
		Entry("due to invalid bootstrap version", errTestCase{
			request: types.BootstrapRequest{
				Host:             "localhost",
				Mesh:             "mesh",
				Name:             "name.namespace",
				AdminPort:        9901,
				BootstrapVersion: "5",
			},
			expected: `Invalid BootstrapVersion. Available values are: "3"`,
		}),
		Entry("when both dataplane and dataplane token path are defined", errTestCase{
			request: types.BootstrapRequest{
				Host:               "localhost",
				Mesh:               "mesh",
				Name:               "name.namespace",
				AdminPort:          9901,
				DataplaneTokenPath: "/tmp",
				DataplaneToken:     "token",
			},
			expected: `dataplaneToken: only one of dataplaneToken and dataplaneTokenField can be defined`,
		}),
		Entry("when CaCert is not a CA and EnvoyGRPC is used", errTestCase{
			request: types.BootstrapRequest{
				Host:           "localhost",
				Mesh:           "mesh",
				Name:           "name.namespace",
				AdminPort:      9901,
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
})
