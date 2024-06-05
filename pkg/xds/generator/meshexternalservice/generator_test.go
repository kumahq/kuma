package meshexternalservice_test

import (
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/datasource"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	rest_v1alpha1 "github.com/kumahq/kuma/pkg/core/resources/model/rest/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/xds"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/generator/meshexternalservice"
)

func parseResource(bytes []byte, resource core_model.Resource) {
	Expect(core_model.FromYAML(bytes, resource.GetSpec())).To(Succeed())
	resMeta := rest_v1alpha1.ResourceMeta{}
	err := yaml.Unmarshal(bytes, &resMeta)
	Expect(err).ToNot(HaveOccurred())
	resource.SetMeta(&model.ResourceMeta{
		Mesh: resMeta.Mesh,
		Name: resMeta.Name,
	})
}

var _ = Describe("MeshExternalServiceGenerator", func() {
	generator := meshexternalservice.Generator{}

	type testCase struct {
		file string
	}

	DescribeTable("should generate envoy config",
		func(given testCase) {
			// given

			// dataplane
			dataplane := core_mesh.NewDataplaneResource()
			bytes, err := os.ReadFile(filepath.Join("testdata", "dataplane.input.yaml"))
			Expect(err).ToNot(HaveOccurred())
			parseResource(bytes, dataplane)

			// mesh
			mesh := core_mesh.NewMeshResource()
			mesh.SetMeta(&model.ResourceMeta{Name: "default"})

			// MeshExternalService
			meshExternalService := v1alpha1.NewMeshExternalServiceResource()
			bytes, err = os.ReadFile(filepath.Join("testdata", given.file+".input.yaml"))
			Expect(err).ToNot(HaveOccurred())
			parseResource(bytes, meshExternalService)

			// loader
			secrets := []*system.SecretResource{
				{
					Meta: &model.ResourceMeta{
						Mesh: "default",
						Name: "123",
					},
					Spec: &system_proto.Secret{
						Data: util_proto.Bytes([]byte("abc")),
					},
				},
			}
			dataSourceLoader := datasource.NewStaticLoader(secrets)

			ctx := xds_context.Context{
				ControlPlane: nil,
				Mesh: xds_context.MeshContext{
					DataSourceLoader: dataSourceLoader,
					Resource:         mesh,
					Resources: xds_context.Resources{
						MeshLocalResources: map[core_model.ResourceType]core_model.ResourceList{
							v1alpha1.MeshExternalServiceType: &v1alpha1.MeshExternalServiceResourceList{
								Items: []*v1alpha1.MeshExternalServiceResource{meshExternalService},
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
			resources, err := generator.Generate(context.Background(), nil, ctx, proxy)

			// then
			Expect(err).ToNot(HaveOccurred())

			resp, err := resources.List().ToDeltaDiscoveryResponse()
			Expect(err).ToNot(HaveOccurred())
			actual, err := util_proto.ToYAML(resp)
			Expect(err).ToNot(HaveOccurred())

			// and output matches golden files
			Expect(actual).To(MatchGoldenYAML(filepath.Join("testdata", given.file+".golden.yaml")))
		},
		Entry("for a sample MeshExternalService", testCase{
			file: "01.sample",
		}),
		Entry("for mode: SkipAll", testCase{
			file: "02.skip-all",
		}),
		Entry("for match: TCP", testCase{
			file: "03.tcp",
		}),
		Entry("for match: gRPC", testCase{
			file: "04.grpc",
		}),
	)
})
