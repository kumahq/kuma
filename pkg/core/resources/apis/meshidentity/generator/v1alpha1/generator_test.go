package v1alpha1_test

import (
	"fmt"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_auth "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	generator "github.com/kumahq/kuma/pkg/core/resources/apis/meshidentity/generator/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	xds_builders "github.com/kumahq/kuma/pkg/test/xds/builders"
<<<<<<< HEAD
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
)

func getResource(
	resourceSet *core_xds.ResourceSet,
	typ envoy_resource.Type,
) []byte {
	resources, err := resourceSet.ListOf(typ).ToDeltaDiscoveryResponse()
	Expect(err).ToNot(HaveOccurred())
	actual, err := util_proto.ToYAML(resources)
	Expect(err).ToNot(HaveOccurred())

	return actual
}

=======
	util_yaml "github.com/kumahq/kuma/pkg/util/yaml"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
)

>>>>>>> master
var _ = Describe("MeshIdentity Generator", func() {
	type testCase struct {
		caseName            string
		additionalResources *core_xds.ResourceSet
	}
	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// given
			context := *xds_builders.Context().
				WithMeshBuilder(samples.MeshDefaultBuilder()).
				Build()
			resourceSet := core_xds.NewResourceSet()
			identity := &core_xds.WorkloadIdentity{
				AdditionalResources: given.additionalResources,
			}
			proxy := xds_builders.Proxy().
				WithWorkloadIdentity(identity).
				WithApiVersion(envoy_common.APIV3).
				Build()

			plugin := generator.NewPlugin()

			// when
			Expect(plugin.Generate(resourceSet, context, proxy)).To(Succeed())

			// then
			resource, err := util_yaml.GetResourcesToYaml(resourceSet, envoy_resource.SecretType)
			Expect(err).ToNot(HaveOccurred())
			Expect(resource).To(matchers.MatchGoldenYAML(fmt.Sprintf("testdata/%s.secrets.golden.yaml", given.caseName)))
		},
		Entry("with-resources", testCase{
			caseName: "with-resources",
			additionalResources: core_xds.NewResourceSet().Add(
				&core_xds.Resource{
					Name:   "test-secret",
					Origin: "OriginAdditionalResource",
					Resource: &envoy_auth.Secret{
						Name: "test-secret",
						Type: &envoy_auth.Secret_ValidationContext{
							ValidationContext: &envoy_auth.CertificateValidationContext{
								TrustedCa: &envoy_core.DataSource{
									Specifier: &envoy_core.DataSource_EnvironmentVariable{
										EnvironmentVariable: "MY_ENV",
									},
								},
							},
						},
					},
				}),
		}),
		Entry("without-resources", testCase{
			caseName:            "without-resources",
			additionalResources: core_xds.NewResourceSet(),
		}))
})
