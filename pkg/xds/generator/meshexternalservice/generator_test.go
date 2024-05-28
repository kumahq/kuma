package meshexternalservice_test

import (
    "context"
    "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
    "github.com/kumahq/kuma/pkg/xds/generator/meshexternalservice"
    "os"
    "path/filepath"

    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    "sigs.k8s.io/yaml"

    core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
    core_model "github.com/kumahq/kuma/pkg/core/resources/model"
    rest_v1alpha1 "github.com/kumahq/kuma/pkg/core/resources/model/rest/v1alpha1"
    "github.com/kumahq/kuma/pkg/core/xds"
    . "github.com/kumahq/kuma/pkg/test/matchers"
    "github.com/kumahq/kuma/pkg/test/resources/model"
    util_proto "github.com/kumahq/kuma/pkg/util/proto"
    xds_context "github.com/kumahq/kuma/pkg/xds/context"
    envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
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
        dataplaneFile           string
        meshExternalServiceFile string
        expected                string
    }

    DescribeTable("should generate envoy config",
        func(given testCase) {
            // given

            // dataplane
            dataplane := core_mesh.NewDataplaneResource()
            bytes, err := os.ReadFile(filepath.Join("testdata", given.dataplaneFile))
            Expect(err).ToNot(HaveOccurred())
            parseResource(bytes, dataplane)

            // MeshExternalService
            meshExternalService := v1alpha1.NewMeshExternalServiceResource()
            bytes, err = os.ReadFile(filepath.Join("testdata", given.meshExternalServiceFile))
            Expect(err).ToNot(HaveOccurred())
            parseResource(bytes, meshExternalService)

            mesh := core_mesh.NewMeshResource()
            mesh.SetMeta(&model.ResourceMeta{Name: "default"})

            ctx := xds_context.Context{
                ControlPlane: nil,
                Mesh: xds_context.MeshContext{
                    Resource: mesh,
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
            Expect(actual).To(MatchGoldenYAML(filepath.Join("testdata", given.expected)))
        },
        Entry("should not generate resources when transparent proxy is off", testCase{
            dataplaneFile:           "01.dataplane.input.yaml",
            meshExternalServiceFile: "01.meshexternalservice.input.yaml",
            expected:                "01.envoy-config.golden.yaml",
        }),
    )
})
