package k8s

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Kong/kuma/pkg/core/resources/apis/system"
	core_registry "github.com/Kong/kuma/pkg/core/resources/registry"
	k8s_registry "github.com/Kong/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
)

var _ = Describe("Consistent Kind Types", func() {
	It("Kind for objects is the same as ResourceType", func() {
		types := core_registry.Global()
		k8sTypes := k8s_registry.Global()

		for _, typ := range types.ObjectTypes() {
			if typ == system.SecretType {
				continue // ignore Secret since we cannot map it directly to Secret K8S resource
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
			if typ == system.SecretType {
				continue // ignore Secret since we cannot map it directly to Secret K8S resource
			}

			res, err := types.NewObject(typ)
			Expect(err).ToNot(HaveOccurred())
			obj, err := k8sTypes.NewObject(res.GetSpec())
			Expect(err).ToNot(HaveOccurred())

			Expect(obj.GetObjectKind().GroupVersionKind().Kind).To(Equal(string(res.GetType())))
		}
	})
})
