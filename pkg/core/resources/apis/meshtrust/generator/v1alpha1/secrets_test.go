package v1alpha1_test

import (
	"fmt"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/kri"
	meshidentity_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	meshtrust_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshtrust/api/v1alpha1"
	generator "github.com/kumahq/kuma/pkg/core/resources/apis/meshtrust/generator/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	xds_builders "github.com/kumahq/kuma/pkg/test/xds/builders"
	"github.com/kumahq/kuma/pkg/util/pointer"
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

var _ = Describe("MeshTrust Secret Generator", func() {
	type testCase struct {
		caseName         string
		workloadIdentity *core_xds.WorkloadIdentity
		trustDomains     map[string][]*meshtrust_api.MeshTrust
	}
	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// given
			context := *xds_builders.Context().
				WithMeshBuilder(samples.MeshDefaultBuilder()).
				Build()
			context.Mesh.TrustsByTrustDomain = given.trustDomains
			resourceSet := core_xds.NewResourceSet()
			proxy := xds_builders.Proxy().
				WithWorkloadIdentity(given.workloadIdentity).
				WithApiVersion(envoy_common.APIV3).
				Build()

			plugin := generator.NewPlugin()

			// when
			Expect(plugin.Generate(resourceSet, context, proxy)).To(Succeed())

			// then
			Expect(getResource(resourceSet, envoy_resource.SecretType)).
				To(matchers.MatchGoldenYAML(fmt.Sprintf("testdata/%s.secrets.golden.yaml", given.caseName)))
		},
		Entry("with-multiple-trust-domains", testCase{
			caseName: "secrets-multiple-trust-domains",
			workloadIdentity: &core_xds.WorkloadIdentity{
				KRI:        kri.Identifier{ResourceType: meshtrust_api.MeshTrustType, Mesh: "default", Name: "identity"},
				ManageType: core_xds.KumaManagedType,
			},
			trustDomains: map[string][]*meshtrust_api.MeshTrust{
				"domain-1": {
					{
						TrustDomain: "domain-1",
						Origin: &meshtrust_api.Origin{
							KRI: pointer.To(kri.Identifier{ResourceType: meshidentity_api.MeshIdentityType, Name: "domain-1"}.String()),
						},
						CABundles: []meshtrust_api.CABundle{
							{
								Type: meshtrust_api.PemCABundleType,
								PEM: &meshtrust_api.PEM{
									Value: "123",
								},
							},
							{
								Type: meshtrust_api.PemCABundleType,
								PEM: &meshtrust_api.PEM{
									Value: "456",
								},
							},
						},
					},
				},
				"domain-2": {
					{
						TrustDomain: "domain-2",
						CABundles: []meshtrust_api.CABundle{
							{
								Type: meshtrust_api.PemCABundleType,
								PEM: &meshtrust_api.PEM{
									Value: "789",
								},
							},
						},
					},
				},
			},
		}),
		Entry("with-multiple-trust-domains-and-default-name", testCase{
			caseName: "secrets-multiple-trust-domains-default-name",
			workloadIdentity: &core_xds.WorkloadIdentity{
				KRI:        kri.Identifier{ResourceType: meshtrust_api.MeshTrustType, Mesh: "default", Name: "identity"},
				ManageType: core_xds.KumaManagedType,
			},
			trustDomains: map[string][]*meshtrust_api.MeshTrust{
				"domain-1": {
					{
						TrustDomain: "domain-1",
						Origin: &meshtrust_api.Origin{
							KRI: pointer.To(kri.Identifier{ResourceType: meshidentity_api.MeshIdentityType, Name: "domain-1"}.String()),
						},
						CABundles: []meshtrust_api.CABundle{
							{
								Type: meshtrust_api.PemCABundleType,
								PEM: &meshtrust_api.PEM{
									Value: "123",
								},
							},
							{
								Type: meshtrust_api.PemCABundleType,
								PEM: &meshtrust_api.PEM{
									Value: "456",
								},
							},
						},
					},
				},
				"domain-2": {
					{
						TrustDomain: "domain-2",
						CABundles: []meshtrust_api.CABundle{
							{
								Type: meshtrust_api.PemCABundleType,
								PEM: &meshtrust_api.PEM{
									Value: "789",
								},
							},
						},
					},
				},
			},
		}),
		Entry("no workload identity and trusts", testCase{
			caseName: "no-secrets",
		}),
		Entry("no secrets for externally managed", testCase{
			caseName: "no-secrets-externally-managed",
			workloadIdentity: &core_xds.WorkloadIdentity{
				KRI:        kri.Identifier{ResourceType: meshtrust_api.MeshTrustType, Mesh: "default", Name: "identity"},
				ManageType: core_xds.ExternalManagedType,
			},
			trustDomains: map[string][]*meshtrust_api.MeshTrust{
				"domain-1": {
					{
						TrustDomain: "domain-1",
						Origin: &meshtrust_api.Origin{
							KRI: pointer.To(kri.Identifier{ResourceType: meshidentity_api.MeshIdentityType, Name: "domain-1"}.String()),
						},
						CABundles: []meshtrust_api.CABundle{
							{
								Type: meshtrust_api.PemCABundleType,
								PEM: &meshtrust_api.PEM{
									Value: "123",
								},
							},
							{
								Type: meshtrust_api.PemCABundleType,
								PEM: &meshtrust_api.PEM{
									Value: "456",
								},
							},
						},
					},
				},
				"domain-2": {
					{
						TrustDomain: "domain-2",
						CABundles: []meshtrust_api.CABundle{
							{
								Type: meshtrust_api.PemCABundleType,
								PEM: &meshtrust_api.PEM{
									Value: "789",
								},
							},
						},
					},
				},
			},
		}),
	)
})
