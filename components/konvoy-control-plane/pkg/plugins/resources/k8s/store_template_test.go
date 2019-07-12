package k8s_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/k8s"
	k8s_registry "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/k8s/native/pkg/registry"
	sample_k8s "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/k8s/native/test/api/sample/v1alpha1"
	sample_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/test/apis/sample/v1alpha1"
)

var _ = Describe("KubernetesStore template", func() {

	store.ExecuteStoreTests(func() store.ResourceStore {
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
