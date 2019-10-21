package k8s_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/plugins/resources/k8s"
	k8s_registry "github.com/Kong/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
	sample_k8s "github.com/Kong/kuma/pkg/plugins/resources/k8s/native/test/api/sample/v1alpha1"
	sample_proto "github.com/Kong/kuma/pkg/test/apis/sample/v1alpha1"
	test_store "github.com/Kong/kuma/pkg/test/store"
)

var _ = Describe("KubernetesStore template", func() {

	test_store.ExecuteStoreTests(func() store.ResourceStore {
		kubeTypes := k8s_registry.NewTypeRegistry()
		Expect(kubeTypes.RegisterObjectType(&sample_proto.TrafficRoute{}, &sample_k8s.TrafficRoute{})).To(Succeed())
		Expect(kubeTypes.RegisterListType(&sample_proto.TrafficRoute{}, &sample_k8s.TrafficRouteList{})).To(Succeed())

		ks := &k8s.KubernetesStore{
			Client: k8sClient,
			Converter: &k8s.SimpleConverter{
				KubeFactory: &k8s.SimpleKubeFactory{
					KubeTypes: kubeTypes,
				},
			},
		}
		s := store.NewStrictResourceStore(ks)
		return s
	})
})
