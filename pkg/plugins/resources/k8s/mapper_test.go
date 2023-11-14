package k8s_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	rest_v1alpha1 "github.com/kumahq/kuma/pkg/core/resources/model/rest/v1alpha1"
	core_mtp "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	k8s_mtp "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/k8s/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	k8s_registry "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
)

var _ = Describe("KubernetesStore template", func() {
	Context("on namespaced entity", func() {
		kubeTypes := k8s_registry.NewTypeRegistry()
		Expect(kubeTypes.RegisterObjectType(&core_mtp.MeshTrafficPermission{}, &k8s_mtp.MeshTrafficPermission{})).To(Succeed())
		Expect(kubeTypes.RegisterObjectType(&mesh_proto.Mesh{}, &mesh_k8s.Mesh{})).To(Succeed())

		It("works passing no namespace on inference", func() {
			in := core_mtp.NewMeshTrafficPermissionResource()
			in.SetMeta(&rest_v1alpha1.ResourceMeta{Name: "foo", Mesh: "default"})

			mapper := k8s.NewInferenceMapper("core-ns", &k8s.SimpleKubeFactory{KubeTypes: kubeTypes})
			res, err := mapper(in, "")

			Expect(err).ToNot(HaveOccurred())
			Expect(res.GetMesh()).To(Equal("default"))
			Expect(res.GetName()).To(Equal("foo"))
			Expect(res.GetNamespace()).To(Equal("core-ns"))
		})

		It("works passing namespace on inference", func() {
			in := core_mtp.NewMeshTrafficPermissionResource()
			in.SetMeta(&rest_v1alpha1.ResourceMeta{Name: "foo", Mesh: "default"})

			mapper := k8s.NewInferenceMapper("core-ns", &k8s.SimpleKubeFactory{KubeTypes: kubeTypes})
			res, err := mapper(in, "my-ns")

			Expect(err).ToNot(HaveOccurred())
			Expect(res.GetMesh()).To(Equal("default"))
			Expect(res.GetName()).To(Equal("foo"))
			Expect(res.GetNamespace()).To(Equal("my-ns"))
		})

		It("works passing no namespace on k8s store", func() {
			in := core_mtp.NewMeshTrafficPermissionResource()
			in.SetMeta(&rest_v1alpha1.ResourceMeta{Name: "foo", Mesh: "default"})
			in.SetMeta(&k8s.KubernetesMetaAdapter{ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "a-namespace", Labels: map[string]string{"hello": "world"}}, Mesh: "default"})

			mapper := k8s.NewKubernetesMapper(&k8s.SimpleKubeFactory{KubeTypes: kubeTypes})
			res, err := mapper(in, "")

			Expect(err).ToNot(HaveOccurred())
			Expect(res.GetMesh()).To(Equal("default"))
			Expect(res.GetName()).To(Equal("foo"))
			Expect(res.GetNamespace()).To(Equal("a-namespace"))
			Expect(res.GetLabels()).To(Equal(map[string]string{"hello": "world"}))
		})

		It("works passing namespace on k8s store", func() {
			in := core_mtp.NewMeshTrafficPermissionResource()
			in.SetMeta(&k8s.KubernetesMetaAdapter{ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "a-namespace", Labels: map[string]string{"hello": "world"}}, Mesh: "default"})

			mapper := k8s.NewKubernetesMapper(&k8s.SimpleKubeFactory{KubeTypes: kubeTypes})
			res, err := mapper(in, "my-ns")

			Expect(err).ToNot(HaveOccurred())
			Expect(res.GetMesh()).To(Equal("default"))
			Expect(res.GetName()).To(Equal("foo"))
			Expect(res.GetNamespace()).To(Equal("my-ns"))
			Expect(res.GetLabels()).To(Equal(map[string]string{"hello": "world"}))
		})
	})
	Context("on cluster scoped entity", func() {
		kubeTypes := k8s_registry.NewTypeRegistry()
		Expect(kubeTypes.RegisterObjectType(&mesh_proto.Mesh{}, &mesh_k8s.Mesh{})).To(Succeed())

		It("works with inference", func() {
			in := core_mesh.NewMeshResource()
			in.SetMeta(&rest_v1alpha1.ResourceMeta{Name: "foo", Mesh: "default"})

			mapper := k8s.NewInferenceMapper("core-ns", &k8s.SimpleKubeFactory{KubeTypes: kubeTypes})
			res, err := mapper(in, "my-ns")

			Expect(err).ToNot(HaveOccurred())
			Expect(res.GetMesh()).To(Equal("default"))
			Expect(res.GetName()).To(Equal("foo"))
		})

		It("works with kubernetes", func() {
			in := core_mesh.NewMeshResource()
			in.SetMeta(&k8s.KubernetesMetaAdapter{ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "a-namespace", Labels: map[string]string{"hello": "world"}}, Mesh: "default"})

			mapper := k8s.NewKubernetesMapper(&k8s.SimpleKubeFactory{KubeTypes: kubeTypes})
			res, err := mapper(in, "my-ns")

			Expect(err).ToNot(HaveOccurred())
			Expect(res.GetMesh()).To(Equal("default"))
			Expect(res.GetName()).To(Equal("foo"))
			Expect(res.GetLabels()).To(Equal(map[string]string{"hello": "world"}))
		})
	})
})
