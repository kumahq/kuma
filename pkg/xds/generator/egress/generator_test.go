package egress_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/permissions"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	rest_types "github.com/kumahq/kuma/pkg/core/resources/model/rest"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/generator/egress"
	generator_secrets "github.com/kumahq/kuma/pkg/xds/generator/secrets"
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
			// given
			var zoneEgress *core_mesh.ZoneEgressResource
			var zoneIngresses []*core_mesh.ZoneIngressResource
			var trafficPermissions []*core_mesh.TrafficPermissionResource

			meshResourcesMap := map[string]*core_xds.MeshResources{}

			resourcePath := filepath.Join(
				"testdata", "input",
				given.fileWithResourcesName,
			)

			resourceBytes, err := os.ReadFile(resourcePath)
			Expect(err).ToNot(HaveOccurred())

			rawResources := strings.Split(string(resourceBytes), "---")
			for _, rawResource := range rawResources {
				bytes := []byte(rawResource)

				res, err := rest_types.YAML.UnmarshalCore(bytes)
				Expect(err).To(BeNil())

				meshName := res.GetMeta().GetMesh()

				switch res.Descriptor().Name {
				case core_mesh.ZoneEgressType:
					Expect(zoneEgress).To(BeNil(), "there can be only one zone egress in resources")
					zoneEgress = res.(*core_mesh.ZoneEgressResource)
				case core_mesh.ZoneIngressType:
					zoneIngresses = append(zoneIngresses, res.(*core_mesh.ZoneIngressResource))
				case core_mesh.TrafficPermissionType:
					trafficPermissions = append(trafficPermissions, res.(*core_mesh.TrafficPermissionResource))
				case core_mesh.MeshType:
					meshName := res.GetMeta().GetName()

					if _, ok := meshResourcesMap[meshName]; !ok {
						meshResourcesMap[meshName] = &core_xds.MeshResources{}
					}

					meshResourcesMap[meshName].Mesh = res.(*core_mesh.MeshResource)
				case core_mesh.ExternalServiceType:
					if _, ok := meshResourcesMap[meshName]; !ok {
						meshResourcesMap[meshName] = &core_xds.MeshResources{}
					}

					meshResourcesMap[meshName].ExternalServices = append(
						meshResourcesMap[meshName].ExternalServices,
						res.(*core_mesh.ExternalServiceResource),
					)
				case core_mesh.TrafficRouteType:
					if _, ok := meshResourcesMap[meshName]; !ok {
						meshResourcesMap[meshName] = &core_xds.MeshResources{}
					}

					meshResourcesMap[meshName].TrafficRoutes = append(
						meshResourcesMap[meshName].TrafficRoutes,
						res.(*core_mesh.TrafficRouteResource),
					)
				}
			}

			Expect(zoneEgress).NotTo(BeNil(), "without zone egress in resources we cannot run test")

			loader := fakeLoader{}

			for _, meshResources := range meshResourcesMap {
				meshResources.EndpointMap = xds_topology.BuildRemoteEndpointMap(
					context.Background(),
					meshResources.Mesh,
					zoneName,
					zoneIngresses,
					meshResources.ExternalServices,
					&loader,
				)

				meshResources.ExternalServicePermissionMap = permissions.BuildExternalServicesPermissionsMapForZoneEgress(
					meshResources.ExternalServices,
					trafficPermissions,
				)
			}

			gen := egress.Generator{
				ZoneEgressGenerators: []egress.ZoneEgressGenerator{
					&egress.InternalServicesGenerator{},
					&egress.ExternalServicesGenerator{},
				},
				SecretGenerator: &generator_secrets.Generator{},
			}

			var meshResourcesList []*core_xds.MeshResources
			for _, meshResources := range meshResourcesMap {
				meshResourcesList = append(meshResourcesList, meshResources)
			}

			proxy := &core_xds.Proxy{
				Id:         *core_xds.BuildProxyId("default", "egress"),
				APIVersion: envoy_common.APIV3,
				ZoneEgressProxy: &core_xds.ZoneEgressProxy{
					ZoneEgressResource: zoneEgress,
					ZoneIngresses:      zoneIngresses,
					MeshResourcesList:  meshResourcesList,
				},
			}
			ctx := xds_context.Context{
				ControlPlane: &xds_context.ControlPlaneContext{
					Zone:    zoneName,
					Secrets: &xds.TestSecrets{},
				},
			}

			// when
			rs, err := gen.Generate(ctx, proxy)

			// then
			Expect(err).ToNot(HaveOccurred())

			resp, err := rs.List().ToDeltaDiscoveryResponse()
			Expect(err).ToNot(HaveOccurred())

			actual, err := util_proto.ToYAML(resp)
			Expect(err).ToNot(HaveOccurred())

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
		Entry("05. custom trafficpermission, mixed internal and external services", testCase{
			fileWithResourcesName: "05.mixed-services-with-custom-trafficpermissions.yaml",
			expected:              "05.mixed-services-with-custom-trafficpermissions.golden.yaml",
		}),
		Entry("06. mixed-services-with-external-in-other-zone", testCase{
			fileWithResourcesName: "06.mixed-services-with-external-in-other-zone.yaml",
			expected:              "06.mixed-services-with-external-in-other-zone.golden.yaml",
		}),
	)
})
