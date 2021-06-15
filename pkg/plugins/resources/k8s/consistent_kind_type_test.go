package k8s

import (
	. "github.com/onsi/ginkgo"
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

			Expect(obj.GetObjectKind().GroupVersionKind().Kind).To(Equal(string(res.GetType())))
		}
	})

	It("Kind for lists is the same as ResourceType", func() {
		types := core_registry.Global()
		k8sTypes := k8s_registry.Global()

		for _, typ := range types.ListTypes() {
			if IgnoredTypes[typ] {
				continue
			}

			res, err := types.NewObject(typ)
			Expect(err).ToNot(HaveOccurred())
			obj, err := k8sTypes.NewObject(res.GetSpec())
			Expect(err).ToNot(HaveOccurred())

			Expect(obj.GetObjectKind().GroupVersionKind().Kind).To(Equal(string(res.GetType())))
		}
	})
})
