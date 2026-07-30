package matchers_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/core/destinationname"
	meshexternalservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/matchers"
	core_rules "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules"
	policies_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
)

var _ = Describe("EgressMatchedPolicies", func() {
	It("should build inbound rules for mesh-scoped MeshTrafficPermission on MeshExternalService egress", func() {
		mes := builders.MeshExternalService().WithMesh("mesh-1").WithName("es-1").Build()
		mtp := builders.MeshTrafficPermission().
			WithMesh("mesh-1").
			WithTargetRef(common_api.TargetRef{
				Kind: common_api.Mesh,
			}).
			AddRule(policies_api.Allow).
			Build()

		resources := xds_context.Resources{
			MeshLocalResources: map[core_model.ResourceType]core_model.ResourceList{
				meshexternalservice_api.MeshExternalServiceType: &meshexternalservice_api.MeshExternalServiceResourceList{
					Items: []*meshexternalservice_api.MeshExternalServiceResource{mes},
				},
				policies_api.MeshTrafficPermissionType: &policies_api.MeshTrafficPermissionResourceList{
					Items: []*policies_api.MeshTrafficPermissionResource{mtp},
				},
			},
		}

		policies, err := matchers.EgressMatchedPolicies(
			policies_api.MeshTrafficPermissionType,
			map[string]string{mesh_proto.ServiceTag: destinationname.MustResolve(false, mes, mes.Spec.Match)},
			resources,
		)
		Expect(err).ToNot(HaveOccurred())

		rules := policies.FromRules.InboundRules[core_rules.InboundListener{}]
		Expect(rules).To(HaveLen(1))
		Expect(rules[0].Conf.(policies_api.RuleConf).Allow).ToNot(BeNil())
	})
})
