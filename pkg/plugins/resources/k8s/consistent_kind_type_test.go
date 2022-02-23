package k8s

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_registry "github.com/kumahq/kuma/pkg/core/resources/registry"
	k8s_registry "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
)

// Those types are not mapped directly to Kubernetes Resource
var IgnoredTypes = map[model.ResourceType]bool{
	system.SecretType:           true,
	system.GlobalSecretType:     true,
	system.ConfigType:           true,
	mesh.ZoneIngressInsightType: true, // uses DataplaneInsight under the hood
	mesh.ZoneEgressInsightType:  true, // uses DataplaneInsight under the hood
	mesh.MeshGatewayType:        true, // Gateway is only in Universal ATM.
	mesh.MeshGatewayRouteType:   true, // GatewayRoute is only in Universal ATM.
}

var _ = Describe("Consistent Kind Types", func() {
	It("Kind for objects is the same as ResourceType", func() {
		types := core_registry.Global()
		k8sTypes := k8s_registry.Global()

		for _, typ := range types.ObjectTypes() {
			if IgnoredTypes[typ] {
				continue
			}

			res, err := types.NewObject(typ)
			Expect(err).ToNot(HaveOccurred())
			obj, err := k8sTypes.NewObject(res.GetSpec())
			Expect(err).ToNot(HaveOccurred())

			Expect(obj.GetObjectKind().GroupVersionKind().Kind).To(Equal(string(res.Descriptor().Name)))
		}
	})

	It("Kind for lists is the same as ResourceType", func() {
		types := core_registry.Global()
		k8sTypes := k8s_registry.Global()

		for _, desc := range types.ObjectDescriptors() {
			if IgnoredTypes[desc.Name] {
				continue
			}

			res, err := types.NewObject(desc.Name)
			Expect(err).ToNot(HaveOccurred())
			obj, err := k8sTypes.NewObject(res.GetSpec())
			Expect(err).ToNot(HaveOccurred())

			Expect(obj.GetObjectKind().GroupVersionKind().Kind).To(Equal(string(res.Descriptor().Name)))
		}
	})
})
