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
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Interface: "8.8.8.8:443:8443",
							Tags: map[string]string{
								"service": "backend",
							},
						},
					},
				},
			},
		}

		// when
		err := resManager.Create(context.Background(), &mesh.MeshResource{}, store.CreateByKey("default", "mesh", "mesh"))
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		err = resManager.Create(context.Background(), &dataplane, store.CreateByKey("namespace", "name", "mesh"))
		// then
		Expect(err).ToNot(HaveOccurred())
	})

	type testCase struct {
		config             *bootstrap_config.BootstrapParamsConfig
		expectedConfigFile string
	}
	DescribeTable("should generate bootstrap configuration",
		func(given testCase) {
			// setup
			generator := NewDefaultBootstrapGenerator(resManager, given.config)

			// when
			bootstrapConfig, err := generator.Generate(context.Background(), types.BootstrapRequest{
				Mesh: "mesh",
				Name: "name.namespace",
			})
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
		Entry("default config", testCase{
			config:             bootstrap_config.DefaultBootstrapParamsConfig(),
			expectedConfigFile: "generator.default-config.golden.yaml",
		}),
		Entry("custom config", testCase{
			config: &bootstrap_config.BootstrapParamsConfig{
				AdminAddress:       "192.168.0.1", // by default, Envoy Admin interface should listen on loopback address
				AdminPort:          9902,          // by default, turn off Admin interface of Envoy
				AdminAccessLogPath: "/var/log",
				XdsHost:            "kuma-control-plane.internal",
				XdsPort:            15678,
				XdsConnectTimeout:  2 * time.Second,
			},
			expectedConfigFile: "generator.custom-config.golden.yaml",
		}),
	)
})
