package generator_test

import (
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	rest_v1alpha1 "github.com/kumahq/kuma/pkg/core/resources/model/rest/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/xds"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	util_yaml "github.com/kumahq/kuma/pkg/util/yaml"
	"github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

func parseResource(bytes []byte, resource core_model.Resource) {
	Expect(util_proto.FromYAML(bytes, resource.GetSpec())).To(Succeed())
	resMeta := rest_v1alpha1.ResourceMeta{}
	err := yaml.Unmarshal(bytes, &resMeta)
	Expect(err).ToNot(HaveOccurred())
	resource.SetMeta(&model.ResourceMeta{
		Mesh: resMeta.Mesh,
		Name: resMeta.Name,
	})
}

var _ = Describe("DirectAccessProxyGenerator", func() {
	generator := generator.DirectAccessProxyGenerator{}

	type testCase struct {
		dataplaneFile  string
		dataplanesFile string
		meshFile       string
		expected       string
	}

	DescribeTable("should generate envoy config",
		func(given testCase) {
			// given

			// dataplane
			dataplane := core_mesh.NewDataplaneResource()
			bytes, err := os.ReadFile(filepath.Join("testdata", "direct-access", given.dataplaneFile))
			Expect(err).ToNot(HaveOccurred())
			parseResource(bytes, dataplane)

			// all dataplanes
			var dataplanes []*core_mesh.DataplaneResource
			dpsBytes, err := os.ReadFile(filepath.Join("testdata", "direct-access", given.dataplanesFile))
			Expect(err).ToNot(HaveOccurred())
			dpYamls := util_yaml.SplitYAML(string(dpsBytes))
			for _, dpYAML := range dpYamls {
				dataplane := core_mesh.NewDataplaneResource()
				parseResource([]byte(dpYAML), dataplane)
				dataplanes = append(dataplanes, dataplane)
			}

			// mesh
			mesh := core_mesh.NewMeshResource()
			bytes, err = os.ReadFile(filepath.Join("testdata", "direct-access", given.meshFile))
			Expect(err).ToNot(HaveOccurred())
			parseResource(bytes, mesh)

			ctx := context.Context{
				ControlPlane: nil,
				Mesh: context.MeshContext{
					Resource: mesh,
					Resources: context.Resources{
						MeshLocalResources: map[core_model.ResourceType]core_model.ResourceList{
							core_mesh.DataplaneType: &core_mesh.DataplaneResourceList{
								Items: dataplanes,
							},
						},
					},
				},
			}

			proxy := &xds.Proxy{
				Dataplane:  dataplane,
				APIVersion: envoy_common.APIV3,
			}

			// when
			resources, err := generator.Generate(ctx, proxy)

			// then
			Expect(err).ToNot(HaveOccurred())

			resp, err := resources.List().ToDeltaDiscoveryResponse()
			Expect(err).ToNot(HaveOccurred())
			actual, err := util_proto.ToYAML(resp)
			Expect(err).ToNot(HaveOccurred())

			// and output matches golden files
			Expect(actual).To(MatchGoldenYAML(filepath.Join("testdata", "direct-access", given.expected)))
		},
		Entry("should not generate resources when transparent proxy is off", testCase{
			dataplaneFile:  "01.dataplane.input.yaml",
			dataplanesFile: "01.dataplanes.input.yaml",
			meshFile:       "01.mesh.input.yaml",
			expected:       "01.envoy-config.golden.yaml",
		}),
		Entry("should not generate resources when there are no direct access services", testCase{
			dataplaneFile:  "02.dataplane.input.yaml",
			dataplanesFile: "02.dataplanes.input.yaml",
			meshFile:       "02.mesh.input.yaml",
			expected:       "02.envoy-config.golden.yaml",
		}),
		Entry("should generate direct access for all services except taken endpoints by outbound", testCase{
			dataplaneFile:  "03.dataplane.input.yaml",
			dataplanesFile: "03.dataplanes.input.yaml",
			meshFile:       "03.mesh.input.yaml",
			expected:       "03.envoy-config.golden.yaml",
		}),
		Entry("should generate direct access for given services", testCase{
			dataplaneFile:  "04.dataplane.input.yaml",
			dataplanesFile: "04.dataplanes.input.yaml",
			meshFile:       "04.mesh.input.yaml",
			expected:       "04.envoy-config.golden.yaml",
		}),
	)
})
