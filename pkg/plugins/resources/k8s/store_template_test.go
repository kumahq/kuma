package k8s_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	meshservice_k8s "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/k8s/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	k8s_registry "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
	test_store "github.com/kumahq/kuma/pkg/test/store"
)

var _ = Describe("KubernetesStore template", func() {
	test_store.ExecuteStoreTests(func() store.ResourceStore {
		kubeTypes := k8s_registry.NewTypeRegistry()

		Expect(kubeTypes.RegisterObjectType(&mesh_proto.TrafficRoute{}, &mesh_k8s.TrafficRoute{})).To(Succeed())
		Expect(kubeTypes.RegisterObjectType(&mesh_proto.Mesh{}, &mesh_k8s.Mesh{})).To(Succeed())
		Expect(kubeTypes.RegisterObjectType(&meshservice_api.MeshService{}, &meshservice_k8s.MeshService{})).To(Succeed())
		Expect(kubeTypes.RegisterListType(&mesh_proto.TrafficRoute{}, &mesh_k8s.TrafficRouteList{})).To(Succeed())
		Expect(kubeTypes.RegisterListType(&mesh_proto.Mesh{}, &mesh_k8s.MeshList{})).To(Succeed())
		Expect(kubeTypes.RegisterListType(&meshservice_api.MeshService{}, &meshservice_k8s.MeshServiceList{})).To(Succeed())

		return &k8s.KubernetesStore{
			Client: k8sClient,
			Converter: &k8s.SimpleConverter{
				KubeFactory: &k8s.SimpleKubeFactory{
					KubeTypes: kubeTypes,
				},
			},
			Scheme: k8sClientScheme,
		}
	}, "kubernetes")
})
