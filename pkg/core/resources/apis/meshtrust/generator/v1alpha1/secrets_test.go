package v1alpha1_test

import (
	"fmt"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/kri"
	meshtrust_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshtrust/api/v1alpha1"
	generator "github.com/kumahq/kuma/pkg/core/resources/apis/meshtrust/generator/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	xds_builders "github.com/kumahq/kuma/pkg/test/xds/builders"
	util_yaml "github.com/kumahq/kuma/pkg/util/yaml"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
)

var _ = Describe("MeshTrust Secret Generator", func() {
	type testCase struct {
		caseName         string
		workloadIdentity *core_xds.WorkloadIdentity
		trustDomains     map[string][]string
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
			resources, err := util_yaml.GetResourcesToYaml(resourceSet, envoy_resource.SecretType)
			Expect(err).ToNot(HaveOccurred())
			Expect(resources).To(matchers.MatchGoldenYAML(fmt.Sprintf("testdata/%s.secrets.golden.yaml", given.caseName)))
		},
		Entry("with-multiple-trust-domains", testCase{
			caseName: "secrets-multiple-trust-domains",
			workloadIdentity: &core_xds.WorkloadIdentity{
				KRI:            kri.Identifier{ResourceType: meshtrust_api.MeshTrustType, Mesh: "default", Name: "identity"},
				ManagementMode: core_xds.KumaManagementMode,
			},
			trustDomains: map[string][]string{
				"domain-1": {"123", "456"},
				"domain-2": {"789"},
			},
		}),
		Entry("with-multiple-trust-domains-and-default-name", testCase{
			caseName: "secrets-multiple-trust-domains-default-name",
			workloadIdentity: &core_xds.WorkloadIdentity{
				KRI:            kri.Identifier{ResourceType: meshtrust_api.MeshTrustType, Mesh: "default", Name: "identity"},
				ManagementMode: core_xds.KumaManagementMode,
			},
			trustDomains: map[string][]string{
				"domain-1": {"123", "456"},
				"domain-2": {"789"},
			},
		}),
		Entry("no workload identity and trusts", testCase{
			caseName: "no-secrets",
		}),
		Entry("no secrets for externally managed", testCase{
			caseName: "no-secrets-externally-managed",
			workloadIdentity: &core_xds.WorkloadIdentity{
				KRI:            kri.Identifier{ResourceType: meshtrust_api.MeshTrustType, Mesh: "default", Name: "identity"},
				ManagementMode: core_xds.ExternalManagementMode,
			},
			trustDomains: map[string][]string{
				"domain-1": {"123", "456"},
				"domain-2": {"789"},
			},
		}),
	)
})
