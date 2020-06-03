package generator_test

import (
	"io/ioutil"
	"path/filepath"

	"github.com/ghodss/yaml"

	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/model/rest"
	"github.com/Kong/kuma/pkg/core/xds"
	"github.com/Kong/kuma/pkg/test/resources/model"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
	util_yaml "github.com/Kong/kuma/pkg/util/yaml"
	"github.com/Kong/kuma/pkg/xds/context"
	"github.com/Kong/kuma/pkg/xds/generator"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

func parseResource(bytes []byte, resource core_model.Resource) {
	Expect(util_proto.FromYAML(bytes, resource.GetSpec())).To(Succeed())
	resMeta := rest.ResourceMeta{}
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
		dataplaneFile   string
		dataplanesFile  string
		meshFile        string
		envoyConfigFile string
	}

	DescribeTable("should generate envoy config",
		func(given testCase) {
			// given

			// dataplane
			dataplane := &core_mesh.DataplaneResource{}
			bytes, err := ioutil.ReadFile(filepath.Join("testdata", "direct-access", given.dataplaneFile))
			Expect(err).ToNot(HaveOccurred())
			parseResource(bytes, dataplane)

			// all dataplanes
			var dataplanes []*core_mesh.DataplaneResource
			dpsBytes, err := ioutil.ReadFile(filepath.Join("testdata", "direct-access", given.dataplanesFile))
			Expect(err).ToNot(HaveOccurred())
			dpYamls := util_yaml.SplitYAML(string(dpsBytes))
			for _, dpYAML := range dpYamls {
				dataplane := &core_mesh.DataplaneResource{}
				parseResource([]byte(dpYAML), dataplane)
				dataplanes = append(dataplanes, dataplane)
			}

			// mesh
			mesh := &core_mesh.MeshResource{}
			bytes, err = ioutil.ReadFile(filepath.Join("testdata", "direct-access", given.meshFile))
			Expect(err).ToNot(HaveOccurred())
			parseResource(bytes, mesh)

			ctx := context.Context{
				ControlPlane: nil,
				Mesh: context.MeshContext{
					Resource: mesh,
					Dataplanes: &core_mesh.DataplaneResourceList{
						Items: dataplanes,
					},
				},
			}

			proxy := &xds.Proxy{
				Dataplane: dataplane,
			}

			// when
			resources, err := generator.Generate(ctx, proxy)

			// then
			Expect(err).ToNot(HaveOccurred())

			resp, err := xds.ResourceList(resources).ToDeltaDiscoveryResponse()
			Expect(err).ToNot(HaveOccurred())
			actual, err := util_proto.ToYAML(resp)
			Expect(err).ToNot(HaveOccurred())
			expected, err := ioutil.ReadFile(filepath.Join("testdata", "direct-access", given.envoyConfigFile))
			Expect(err).ToNot(HaveOccurred())

			Expect(actual).To(MatchYAML(expected))
		},
		Entry("should not generate resources when transparent proxy is off", testCase{
			dataplaneFile:   "01.dataplane.input.yaml",
			dataplanesFile:  "01.dataplanes.input.yaml",
			meshFile:        "01.mesh.input.yaml",
			envoyConfigFile: "01.envoy-config.golden.yaml",
		}),
		Entry("should not generate resources when there are no direct access services", testCase{
			dataplaneFile:   "02.dataplane.input.yaml",
			dataplanesFile:  "02.dataplanes.input.yaml",
			meshFile:        "02.mesh.input.yaml",
			envoyConfigFile: "02.envoy-config.golden.yaml",
		}),
		Entry("should generate direct access for all services except taken endpoints by outbound", testCase{
			dataplaneFile:   "03.dataplane.input.yaml",
			dataplanesFile:  "03.dataplanes.input.yaml",
			meshFile:        "03.mesh.input.yaml",
			envoyConfigFile: "03.envoy-config.golden.yaml",
		}),
		Entry("should generate direct access for given services", testCase{
			dataplaneFile:   "04.dataplane.input.yaml",
			dataplanesFile:  "04.dataplanes.input.yaml",
			meshFile:        "04.mesh.input.yaml",
			envoyConfigFile: "04.envoy-config.golden.yaml",
		}),
	)
})
