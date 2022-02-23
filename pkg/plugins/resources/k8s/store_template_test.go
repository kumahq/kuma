package k8s_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	k8s_registry "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
	sample_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/test/api/sample/v1alpha1"
	sample_proto "github.com/kumahq/kuma/pkg/test/apis/sample/v1alpha1"
	test_store "github.com/kumahq/kuma/pkg/test/store"
)

var _ = Describe("KubernetesStore template", func() {

	test_store.ExecuteStoreTests(func() store.ResourceStore {
		kubeTypes := k8s_registry.NewTypeRegistry()

		Expect(kubeTypes.RegisterObjectType(&sample_proto.TrafficRoute{}, &sample_k8s.SampleTrafficRoute{})).To(Succeed())
		Expect(kubeTypes.RegisterObjectType(&mesh_proto.Mesh{}, &mesh_k8s.Mesh{})).To(Succeed())
		Expect(kubeTypes.RegisterListType(&sample_proto.TrafficRoute{}, &sample_k8s.SampleTrafficRouteList{})).To(Succeed())
		Expect(kubeTypes.RegisterListType(&mesh_proto.Mesh{}, &mesh_k8s.MeshList{})).To(Succeed())

		return &k8s.KubernetesStore{
			Client: k8sClient,
			Converter: &k8s.SimpleConverter{
				KubeFactory: &k8s.SimpleKubeFactory{
					KubeTypes: kubeTypes,
				},
			},
			Scheme: k8sClientScheme,
		}
	})
})
