package bootstrap_test

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	. "github.com/Kong/kuma/pkg/xds/bootstrap"

	bootstrap_config "github.com/Kong/kuma/pkg/config/xds/bootstrap"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
	"github.com/Kong/kuma/pkg/xds/bootstrap/types"
)

var _ = Describe("bootstrapGenerator", func() {

	var resManager core_manager.ResourceManager

	BeforeEach(func() {
		resManager = core_manager.NewResourceManager(memory.NewStore())
	})

	BeforeEach(func() {
		// given
		dataplane := mesh.DataplaneResource{
			Spec: mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "8.8.8.8",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Port:        443,
							ServicePort: 8443,
							Tags: map[string]string{
								"service": "backend",
							},
						},
					},
				},
			},
		}

		// when
		meshRes := mesh.MeshResource{
			Spec: mesh_proto.Mesh{
				Tracing: &mesh_proto.Tracing{
					Backends: []*mesh_proto.TracingBackend{
						{
							Name: "zipkin-us",
							Type: mesh_proto.TracingZipkinType,
							Conf: util_proto.MustToStruct(&mesh_proto.ZipkinTracingBackendConfig{
								Url: "http://zipkin.us:9090/v2/spans",
							}),
						},
					},
				},
			},
		}
		err := resManager.Create(context.Background(), &meshRes, store.CreateByKey("mesh", "mesh"))
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		err = resManager.Create(context.Background(), &dataplane, store.CreateByKey("name.namespace", "mesh"))
		// then
		Expect(err).ToNot(HaveOccurred())
	})

	type testCase struct {
		config             func() *bootstrap_config.BootstrapParamsConfig
		request            types.BootstrapRequest
		expectedConfigFile string
	}
	DescribeTable("should generate bootstrap configuration",
		func(given testCase) {
			// setup
			generator := NewDefaultBootstrapGenerator(resManager, given.config(), "")

			// when
			bootstrapConfig, err := generator.Generate(context.Background(), given.request)
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := util_proto.ToYAML(bootstrapConfig)
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			expected, err := ioutil.ReadFile(filepath.Join("testdata", given.expectedConfigFile))
			// then
			Expect(err).ToNot(HaveOccurred())

			// expect
			Expect(actual).To(MatchYAML(expected))
		},
		Entry("default config with minimal request", testCase{
			config: func() *bootstrap_config.BootstrapParamsConfig {
				cfg := bootstrap_config.DefaultBootstrapParamsConfig()
				cfg.XdsHost = "127.0.0.1"
				cfg.XdsPort = 5678
				return cfg
			},
			request: types.BootstrapRequest{
				Mesh: "mesh",
				Name: "name.namespace",
			},
			expectedConfigFile: "generator.default-config-minimal-request.golden.yaml",
		}),
		Entry("default config", testCase{
			config: func() *bootstrap_config.BootstrapParamsConfig {
				cfg := bootstrap_config.DefaultBootstrapParamsConfig()
				cfg.XdsHost = "127.0.0.1"
				cfg.XdsPort = 5678
				return cfg
			},
			request: types.BootstrapRequest{
				Mesh:               "mesh",
				Name:               "name.namespace",
				AdminPort:          1234,
				DataplaneTokenPath: "/tmp/token",
			},
			expectedConfigFile: "generator.default-config.golden.yaml",
		}),
		Entry("custom config with minimal request", testCase{
			config: func() *bootstrap_config.BootstrapParamsConfig {
				return &bootstrap_config.BootstrapParamsConfig{
					AdminAddress:       "192.168.0.1", // by default, Envoy Admin interface should listen on loopback address
					AdminPort:          9902,          // by default, turn off Admin interface of Envoy
					AdminAccessLogPath: "/var/log",
					XdsHost:            "kuma-control-plane.internal",
					XdsPort:            15678,
					XdsConnectTimeout:  2 * time.Second,
				}
			},
			request: types.BootstrapRequest{
				Mesh: "mesh",
				Name: "name.namespace",
			},
			expectedConfigFile: "generator.custom-config-minimal-request.golden.yaml",
		}),
		Entry("custom config", testCase{
			config: func() *bootstrap_config.BootstrapParamsConfig {
				return &bootstrap_config.BootstrapParamsConfig{
					AdminAddress:       "192.168.0.1", // by default, Envoy Admin interface should listen on loopback address
					AdminPort:          9902,          // by default, turn off Admin interface of Envoy
					AdminAccessLogPath: "/var/log",
					XdsHost:            "kuma-control-plane.internal",
					XdsPort:            15678,
					XdsConnectTimeout:  2 * time.Second,
				}
			},
			request: types.BootstrapRequest{
				Mesh:               "mesh",
				Name:               "name.namespace",
				AdminPort:          1234,
				DataplaneTokenPath: "/tmp/token",
			},
			expectedConfigFile: "generator.custom-config.golden.yaml",
		}),
	)

	It("should generate bootstrap configuration with zipkin tracing", func() {
		// setup
		trafficTrace := mesh.TrafficTraceResource{
			Spec: mesh_proto.TrafficTrace{
				Selectors: []*mesh_proto.Selector{
					{
						Match: map[string]string{
							"service": "backend",
						},
					},
				},
				Conf: &mesh_proto.TrafficTrace_Conf{
					Backend: "zipkin-us",
				},
			},
		}
		err := resManager.Create(context.Background(), &trafficTrace, store.CreateByKey("tt", "mesh"))
		// then
		Expect(err).ToNot(HaveOccurred())

		// given
		params := bootstrap_config.DefaultBootstrapParamsConfig()
		params.XdsHost = "127.0.0.1"
		params.XdsPort = 5678

		generator := NewDefaultBootstrapGenerator(resManager, params, "")
		request := types.BootstrapRequest{
			Mesh: "mesh",
			Name: "name.namespace",
		}

		// when
		bootstrapConfig, err := generator.Generate(context.Background(), request)
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		actual, err := util_proto.ToYAML(bootstrapConfig)
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		expected, err := ioutil.ReadFile(filepath.Join("testdata", "bootstrap.tracing.yaml"))
		// then
		Expect(err).ToNot(HaveOccurred())

		// expect
		Expect(actual).To(MatchYAML(expected))
	})

	It("should fail bootstrap configuration due to conflicting port in inbound", func() {
		// setup
		dataplane := mesh.DataplaneResource{
			Spec: mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "8.8.8.8",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Address:     "127.0.0.1",
							Port:        9901,
							ServicePort: 8443,
							Tags: map[string]string{
								"service": "backend",
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
		params := bootstrap_config.DefaultBootstrapParamsConfig()
		params.XdsHost = "127.0.0.1"
		params.XdsPort = 5678

		generator := NewDefaultBootstrapGenerator(resManager, params, "")
		request := types.BootstrapRequest{
			Mesh:      "mesh",
			Name:      "name-1.namespace",
			AdminPort: 9901,
		}

		// when
		_, err = generator.Generate(context.Background(), request)
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
		_, err = generator.Generate(context.Background(), request)
		// then
		Expect(err).To(HaveOccurred())
		// and
		Expect(err.Error()).To(Equal("Resource precondition failed: Port 9901 requested as both admin and inbound port."))

	})

	It("should fail bootstrap configuration due to conflicting port in outbound", func() {
		// setup
		dataplane := mesh.DataplaneResource{
			Spec: mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "8.8.8.8",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Port:        443,
							ServicePort: 8443,
							Tags: map[string]string{
								"service": "backend",
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
		params := bootstrap_config.DefaultBootstrapParamsConfig()
		params.XdsHost = "127.0.0.1"
		params.XdsPort = 5678

		generator := NewDefaultBootstrapGenerator(resManager, params, "")
		request := types.BootstrapRequest{
			Mesh:      "mesh",
			Name:      "name-3.namespace",
			AdminPort: 9901,
		}

		// when
		_, err = generator.Generate(context.Background(), request)
		// then
		Expect(err).To(HaveOccurred())
		// and
		Expect(err.Error()).To(Equal("Resource precondition failed: Port 9901 requested as both admin and outbound port."))

	})
})
