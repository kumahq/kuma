package egress_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/generator/egress"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

type fakeLoader struct {
}

func (f *fakeLoader) Load(
	_ context.Context,
	_ string,
	source *system_proto.DataSource,
) ([]byte, error) {
	// In test resources we are currently using only strings
	return []byte(source.GetInlineString()), nil
}

var _ = Describe("EgressGenerator", func() {
	zoneName := "zone-1"

	type testCase struct {
		fileWithResourcesName string
		expected              string
	}

	DescribeTable("should generate Envoy xDS resources",
		func(given testCase) {
			var zoneEgress *core_mesh.ZoneEgressResource
			var meshes []*core_mesh.MeshResource
			var zoneIngresses []*core_mesh.ZoneIngressResource

			externalServiceMap := map[string][]*core_mesh.ExternalServiceResource{}
			trafficRouteMap := map[string][]*core_mesh.TrafficRouteResource{}
			meshEndpointMap := map[string]core_xds.EndpointMap{}

			resourcePath := filepath.Join(
				"testdata", "input",
				given.fileWithResourcesName,
			)

			resourceBytes, err := os.ReadFile(resourcePath)
			Expect(err).ToNot(HaveOccurred())

			resourceReader := bytes.NewReader(resourceBytes)
			yamlDecoder := yaml.NewDecoder(resourceReader)

			var parsedResource map[string]interface{}

			for yamlDecoder.Decode(&parsedResource) == nil {
				var mesh string
				kind := parsedResource["type"].(string)
				name := parsedResource["name"].(string)
				delete(parsedResource, "type")
				delete(parsedResource, "name")
				if m, ok := parsedResource["mesh"].(string); ok {
					mesh = m
					delete(parsedResource, "mesh")
				}

				specBytes, err := yaml.Marshal(parsedResource)
				Expect(err).To(BeNil())

				object, err := registry.Global().
					NewObject(core_model.ResourceType(kind))
				Expect(err).To(BeNil())

				meta := &test_model.ResourceMeta{
					Name: name,
					Mesh: mesh,
				}
				object.SetMeta(meta)

				Expect(util_proto.FromYAML(specBytes, object.GetSpec())).To(Succeed())

				switch object.Descriptor().Name {
				case core_mesh.ZoneEgressType:
					Expect(zoneEgress).To(BeNil(), "there can be only one zone egress in resources")
					zoneEgress = object.(*core_mesh.ZoneEgressResource)
				case core_mesh.MeshType:
					meshes = append(meshes, object.(*core_mesh.MeshResource))
				case core_mesh.ZoneIngressType:
					zoneIngresses = append(zoneIngresses, object.(*core_mesh.ZoneIngressResource))
				case core_mesh.ExternalServiceType:
					externalServiceMap[mesh] = append(
						externalServiceMap[mesh],
						object.(*core_mesh.ExternalServiceResource),
					)
				case core_mesh.TrafficRouteType:
					trafficRouteMap[mesh] = append(
						trafficRouteMap[mesh],
						object.(*core_mesh.TrafficRouteResource),
					)
				}
			}

			Expect(zoneEgress).NotTo(BeNil(), "without zone egress in resources we cannot run test")

			loader := fakeLoader{}

			for _, mesh := range meshes {
				meshName := mesh.GetMeta().GetName()

				meshEndpointMap[meshName] = xds_topology.BuildRemoteEndpointMap(
					mesh,
					zoneName,
					zoneIngresses,
					externalServiceMap[meshName],
					&loader,
				)
			}

			gen := egress.Generator{
				Generators: []egress.ZoneEgressGenerator{
					&egress.ListenerGenerator{},
					&egress.InternalServicesGenerator{},
					&egress.ExternalServicesGenerator{},
				},
			}

			proxy := &core_xds.Proxy{
				Id:         *core_xds.BuildProxyId("default", "egress"),
				APIVersion: envoy_common.APIV3,
				ZoneEgressProxy: &core_xds.ZoneEgressProxy{
					Zone:               zoneName,
					ZoneEgressResource: zoneEgress,
					DataSourceLoader:   &loader,
					ExternalServiceMap: externalServiceMap,
					MeshEndpointMap:    meshEndpointMap,
					Meshes:             meshes,
					TrafficRouteMap:    trafficRouteMap,
					ZoneIngresses:      zoneIngresses,
				},
			}

			// when
			rs, err := gen.Generate(xds_context.Context{}, proxy)
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			resp, err := rs.List().ToDeltaDiscoveryResponse()
			// then
			Expect(err).ToNot(HaveOccurred())
			// when
			actual, err := util_proto.ToYAML(resp)
			// then
			Expect(err).ToNot(HaveOccurred())

			// and output matches golden files
			Expect(actual).To(MatchGoldenYAML(filepath.Join("testdata", given.expected)))
		},
		Entry("01. default trafficroute, one externalservice", testCase{
			fileWithResourcesName: "01.externalservice-only.yaml",
			expected:              "01.externalservice-only.golden.yaml",
		}),
		Entry("02. default trafficroute, one service behind zoneingress", testCase{
			fileWithResourcesName: "02.internalservice-only.yaml",
			expected:              "02.internalservice-only.golden.yaml",
		}),
		Entry("03. default trafficroute, mixed internal and external services", testCase{
			fileWithResourcesName: "03.mixed-services.yaml",
			expected:              "03.mixed-services.golden.yaml",
		}),
		Entry("04. custom trafficroute, mixed internal and external services", testCase{
			fileWithResourcesName: "04.mixed-services-custom-trafficroute.yaml",
			expected:              "04.mixed-services-custom-trafficroute.golden.yaml",
		}),
	)
})
