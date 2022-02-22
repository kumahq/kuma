package k8s_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	k8s_registry "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
	sample_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/test/api/sample/v1alpha1"
	sample_proto "github.com/kumahq/kuma/pkg/test/apis/sample/v1alpha1"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	sample_core "github.com/kumahq/kuma/pkg/test/resources/apis/sample"
	util_k8s "github.com/kumahq/kuma/pkg/util/k8s"
)

var _ = Describe("KubernetesStore", func() {

	var ks *k8s.KubernetesStore
	var s store.ResourceStore
	var ns string // each test should run in a dedicated k8s namespace
	var coreName string
	const name = "demo"
	const mesh = "default"

	var backend = struct {
		ParseYAML       func(yaml string) client.Object
		Create          func(obj client.Object)
		Get             func(obj client.Object, ns, name string)
		AssertNotExists func(obj client.Object, ns, name string)
		Delete          func(obj client.Object)
	}{
		ParseYAML: func(text string) client.Object {
			// setup
			decoder := serializer.NewCodecFactory(k8sClientScheme).UniversalDeserializer()
			// when
			obj, _, err := decoder.Decode([]byte(text), nil, nil)
			// then
			Expect(err).ToNot(HaveOccurred())
			return obj.(client.Object)
		},
		Create: func(obj client.Object) {
			// when
			err := k8sClient.Create(context.Background(), obj)
			// then
			Expect(err).ToNot(HaveOccurred())
		},
		Get: func(obj client.Object, ns, name string) {
			// when
			err := k8sClient.Get(context.Background(), client.ObjectKey{Namespace: ns, Name: name}, obj)
			// then
			Expect(err).ToNot(HaveOccurred())
		},
		AssertNotExists: func(obj client.Object, ns, name string) {
			// when
			err := k8sClient.Get(context.Background(), client.ObjectKey{Namespace: ns, Name: name}, obj)
			// then
			Expect(apierrs.IsNotFound(err)).To(BeTrue())
		},
		Delete: func(obj client.Object) {
			// when
			err := k8sClient.Delete(context.Background(), obj)
			// then
			Expect(err).ToNot(HaveOccurred())
		},
	}

	BeforeEach(func() {
		kubeTypes := k8s_registry.NewTypeRegistry()
		Expect(kubeTypes.RegisterObjectType(&sample_proto.TrafficRoute{}, &sample_k8s.SampleTrafficRoute{})).To(Succeed())
		Expect(kubeTypes.RegisterObjectType(&mesh_proto.Mesh{}, &mesh_k8s.Mesh{})).To(Succeed())
		Expect(kubeTypes.RegisterListType(&sample_proto.TrafficRoute{}, &sample_k8s.SampleTrafficRouteList{})).To(Succeed())
		Expect(kubeTypes.RegisterListType(&mesh_proto.Mesh{}, &mesh_k8s.MeshList{})).To(Succeed())

		ks = &k8s.KubernetesStore{
			Client: k8sClient,
			Converter: &k8s.SimpleConverter{
				KubeFactory: &k8s.SimpleKubeFactory{
					KubeTypes: kubeTypes,
				},
			},
			Scheme: k8sClientScheme,
		}
		s = store.NewStrictResourceStore(store.NewPaginationStore(ks))
		ns = core.NewUUID()

		err := k8sClient.Create(context.Background(), &kube_core.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: ns},
		})
		Expect(err).ToNot(HaveOccurred())

		coreName = util_k8s.K8sNamespacedNameToCoreName(name, ns)
	})

	AfterEach(func() {
		err := k8sClient.DeleteAllOf(context.Background(), &sample_k8s.SampleTrafficRoute{}, client.InNamespace(ns))
		Expect(err).ToNot(HaveOccurred())
		err = k8sClient.DeleteAllOf(context.Background(), &mesh_k8s.Mesh{}, client.InNamespace(ns))
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("Create()", func() {
		It("should create a new resource", func() {
			// given
			tr := &sample_core.TrafficRouteResource{
				Spec: &sample_proto.TrafficRoute{
					Path: "/example",
				},
			}
			expected := backend.ParseYAML(`
            apiVersion: sample.test.kuma.io/v1alpha1
            kind: SampleTrafficRoute
            mesh: default
            spec:
              path: /example
`).(*sample_k8s.SampleTrafficRoute)

			// when
			err := s.Create(context.Background(), tr, store.CreateByKey(coreName, mesh))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(tr.Meta.GetName()).To(Equal(coreName))
			Expect(tr.Meta.GetMesh()).To(Equal(mesh))
			Expect(tr.Meta.GetVersion()).ToNot(Equal(""))

			// when
			actual := sample_k8s.SampleTrafficRoute{}
			backend.Get(&actual, ns, name)

			// then
			Expect(actual.Spec).To(Equal(expected.Spec))
			// and
			Expect(actual.ObjectMeta.ResourceVersion).To(Equal(tr.Meta.GetVersion()))
		})

		It("should create a new Mesh", func() {
			// given
			mesh := core_mesh.MeshResource{
				Spec: &mesh_proto.Mesh{
					Mtls: &mesh_proto.Mesh_Mtls{
						EnabledBackend: "builtin-1",
						Backends: []*mesh_proto.CertificateAuthorityBackend{
							{
								Name: "builtin-1",
								Type: "builtin",
							},
						},
					},
				},
			}

			expected := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: kuma.io/v1alpha1
            kind: Mesh
            metadata:
              name: %s
            spec:
              mtls:
                enabledBackend: builtin-1
                backends:
                - name: builtin-1
                  type: builtin
`, name)).(*mesh_k8s.Mesh)

			// when
			err := s.Create(context.Background(), &mesh, store.CreateByKey(name, name))

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actual := mesh_k8s.Mesh{}
			backend.Get(&actual, ns, name)

			// then
			Expect(actual.Spec).To(Equal(expected.Spec))
			// and
			Expect(actual.ObjectMeta.ResourceVersion).To(Equal(mesh.Meta.GetVersion()))
		})

		It("should not create a duplicate resource", func() {
			// setup
			backend.AssertNotExists(&sample_k8s.SampleTrafficRoute{}, ns, name)

			// when
			err := s.Create(context.Background(), sample_core.NewTrafficRouteResource(), store.CreateByKey(coreName, mesh))

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			err = s.Create(context.Background(), sample_core.NewTrafficRouteResource(), store.CreateByKey(coreName, mesh))

			// then
			Expect(err).To(MatchError(store.ErrorResourceAlreadyExists(sample_core.TrafficRouteType, coreName, mesh)))
		})

		It("should set owner reference", func() {
			// setup
			mesh := core_mesh.NewMeshResource()
			err := s.Create(context.Background(), mesh, store.CreateByKey("mesh", core_model.NoMesh))
			Expect(err).ToNot(HaveOccurred())

			err = k8sClient.Get(context.Background(), client.ObjectKey{Namespace: ns, Name: "mesh"}, &mesh_k8s.Mesh{})
			Expect(err).ToNot(HaveOccurred())

			tr := sample_core.TrafficRouteResource{
				Spec: &sample_proto.TrafficRoute{
					Path: "/example",
				},
			}
			// when
			err = s.Create(context.Background(), &tr, store.CreateByKey(coreName, "mesh"), store.CreateWithOwner(mesh))
			Expect(err).ToNot(HaveOccurred())

			// then
			obj := sample_k8s.SampleTrafficRoute{}
			err = k8sClient.Get(context.Background(), client.ObjectKey{Namespace: ns, Name: name}, &obj)
			Expect(err).ToNot(HaveOccurred())
			owners := obj.GetOwnerReferences()
			Expect(owners).To(HaveLen(1))
			Expect(owners[0].Name).To(Equal("mesh"))
			Expect(owners[0].Kind).To(Equal("Mesh"))
		})
	})

	Describe("Update()", func() {
		It("should update an existing resource", func() {
			// setup
			initial := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: sample.test.kuma.io/v1alpha1
            kind: SampleTrafficRoute
            mesh: default
            metadata:
              namespace: %s
              name: %s
            spec:
              path: /example
`, ns, name))
			backend.Create(initial)
			// and
			expected := backend.ParseYAML(`
            apiVersion: sample.test.kuma.io/v1alpha1
            kind: SampleTrafficRoute
            mesh: default
            spec:
              path: /another
`).(*sample_k8s.SampleTrafficRoute)

			// given
			tr := sample_core.NewTrafficRouteResource()

			// when
			err := s.Get(context.Background(), tr, store.GetByKey(coreName, mesh))
			// then
			Expect(err).ToNot(HaveOccurred())
			version := tr.Meta.GetVersion()

			// when
			tr.Spec.Path = "/another"
			err = s.Update(context.Background(), tr)

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(tr.Meta.GetVersion()).ToNot(Equal(""))
			Expect(tr.Meta.GetVersion()).ToNot(Equal(version))

			// when
			actual := sample_k8s.SampleTrafficRoute{}
			backend.Get(&actual, ns, name)

			// then
			Expect(actual.Spec).To(Equal(expected.Spec))
			// and
			Expect(actual.ObjectMeta.ResourceVersion).To(Equal(tr.Meta.GetVersion()))
		})

		It("should create a new Mesh", func() {
			// setup
			initial := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: kuma.io/v1alpha1
            kind: Mesh
            metadata:
              name: %s
            spec:
              mtls: {}
`, name))
			backend.Create(initial)

			// and
			expected := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: kuma.io/v1alpha1
            kind: Mesh
            metadata:
              name: %s
            spec:
              mtls:
                enabledBackend: builtin
                backends:
                - name: builtin
                  type: builtin
`, name)).(*mesh_k8s.Mesh)

			// given
			mesh := core_mesh.NewMeshResource()

			// when
			err := s.Get(context.Background(), mesh, store.GetByKey(name, core_model.NoMesh))

			// then
			Expect(err).ToNot(HaveOccurred())
			version := mesh.Meta.GetVersion()

			// when
			mesh.Spec.Mtls.EnabledBackend = "builtin"
			mesh.Spec.Mtls.Backends = []*mesh_proto.CertificateAuthorityBackend{
				{
					Name: "builtin",
					Type: "builtin",
				},
			}
			err = s.Update(context.Background(), mesh)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(mesh.Meta.GetVersion()).ToNot(Equal(""))
			Expect(mesh.Meta.GetVersion()).ToNot(Equal(version))

			// when
			actual := mesh_k8s.Mesh{}
			backend.Get(&actual, ns, name)

			// then
			Expect(actual.Spec).To(Equal(expected.Spec))
			// and
			Expect(actual.ObjectMeta.ResourceVersion).To(Equal(mesh.Meta.GetVersion()))
		})

		It("should return an error if resource is not found", func() {
			// setup
			initial := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: sample.test.kuma.io/v1alpha1
            kind: SampleTrafficRoute
            mesh: default
            metadata:
              namespace: %s
              name: %s
`, ns, name))
			backend.Create(initial)

			// given
			tr := sample_core.NewTrafficRouteResource()

			// when
			err := s.Get(context.Background(), tr, store.GetByKey(coreName, mesh))
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			backend.Delete(initial)
			// and
			err = s.Update(context.Background(), tr)

			// then
			Expect(err).To(MatchError(store.ErrorResourceConflict(sample_core.TrafficRouteType, coreName, mesh)))
		})

		It("should return an error if resource has changed", func() {
			// setup
			initial := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: sample.test.kuma.io/v1alpha1
            kind: SampleTrafficRoute
            mesh: default
            metadata:
              namespace: %s
              name: %s
`, ns, name))
			backend.Create(initial)

			// given
			tr1 := sample_core.NewTrafficRouteResource()

			// when
			err := s.Get(context.Background(), tr1, store.GetByKey(coreName, mesh))
			// then
			Expect(err).ToNot(HaveOccurred())

			// given
			tr2 := sample_core.NewTrafficRouteResource()

			// when
			err = s.Get(context.Background(), tr2, store.GetByKey(coreName, mesh))
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			tr1.Spec.Path = "/example"
			err = s.Update(context.Background(), tr1)
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			tr2.Spec.Path = "/another"
			err = s.Update(context.Background(), tr2)
			// then
			Expect(err).To(MatchError(store.ErrorResourceConflict(sample_core.TrafficRouteType, coreName, mesh)))
		})
	})

	Describe("Get()", func() {
		It("should return an error if resource is not found", func() {
			// setup
			backend.AssertNotExists(&sample_k8s.SampleTrafficRoute{}, ns, name)

			// when
			err := s.Get(context.Background(), sample_core.NewTrafficRouteResource(), store.GetByKey(coreName, mesh))

			// then
			Expect(err).To(MatchError(store.ErrorResourceNotFound(sample_core.TrafficRouteType, coreName, mesh)))
		})

		It("should return an existing resource", func() {
			// setup
			expected := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: sample.test.kuma.io/v1alpha1
            kind: SampleTrafficRoute
            mesh: default
            metadata:
              namespace: %s
              name: %s
            spec:
              path: /example
`, ns, name))
			backend.Create(expected)

			// given
			actual := sample_core.NewTrafficRouteResource()

			// when
			err := s.Get(context.Background(), actual, store.GetByKey(coreName, mesh))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual.Meta.GetName()).To(Equal(coreName))
			// and
			Expect(actual.Meta.GetNameExtensions()).To(Equal(core_model.ResourceNameExtensions{
				"k8s.kuma.io/namespace": ns,
				"k8s.kuma.io/name":      name,
			}))
			// and
			Expect(actual.Spec.Path).To(Equal("/example"))
		})

		It("should return Mesh", func() {
			// setup
			expected := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: kuma.io/v1alpha1
            kind: Mesh
            metadata:
              name: %s
            spec:
              mtls:
                enabledBackend: builtin
                backends:
                - name: builtin
                  type: builtin
`, name))
			backend.Create(expected)

			// given
			actual := core_mesh.NewMeshResource()

			// when
			err := s.Get(context.Background(), actual, store.GetByKey(name, core_model.NoMesh))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual.Meta.GetName()).To(Equal(name))
			// and
			Expect(actual.Meta.GetNameExtensions()).To(Equal(core_model.ResourceNameExtensions{
				"k8s.kuma.io/namespace": "",
				"k8s.kuma.io/name":      name,
			}))
			// and
			Expect(actual.Spec).To(MatchProto(&mesh_proto.Mesh{
				Mtls: &mesh_proto.Mesh_Mtls{
					EnabledBackend: "builtin",
					Backends: []*mesh_proto.CertificateAuthorityBackend{
						{
							Name: "builtin",
							Type: "builtin",
						},
					},
				},
			}))
		})
	})

	Describe("Delete()", func() {
		It("should return en error if resource is not found", func() {
			// setup
			backend.AssertNotExists(&sample_k8s.SampleTrafficRoute{}, ns, name)
			resource := sample_core.NewTrafficRouteResource()

			// when
			err := s.Delete(context.Background(), resource, store.DeleteByKey(coreName, mesh))

			// then
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(store.ErrorResourceNotFound(resource.Descriptor().Name, coreName, mesh)))
		})

		It("should delete an existing resource", func() {
			// setup
			initial := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: sample.test.kuma.io/v1alpha1
            kind: SampleTrafficRoute
            mesh: default
            metadata:
              namespace: %s
              name: %s
`, ns, name))
			backend.Create(initial)

			// when
			err := s.Delete(context.Background(), sample_core.NewTrafficRouteResource(), store.DeleteByKey(coreName, mesh))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			backend.AssertNotExists(&sample_k8s.SampleTrafficRoute{}, ns, name)
		})

		It("should delete Mesh", func() {
			// setup
			initial := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: kuma.io/v1alpha1
            kind: Mesh
            metadata:
              name: %s
`, name))
			backend.Create(initial)

			// when
			err := s.Delete(context.Background(), core_mesh.NewMeshResource(), store.DeleteByKey(name, core_model.NoMesh))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			backend.AssertNotExists(&mesh_k8s.Mesh{}, ns, "demo")
		})
	})

	Describe("List()", func() {
		It("should return an empty list if there are no matching resources", func() {
			// given
			trl := &sample_core.TrafficRouteResourceList{}

			// when
			err := s.List(context.Background(), trl, store.ListByMesh(mesh))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(trl.Items).To(HaveLen(0))
		})

		It("should return a list of matching resource", func() {
			// setup
			demoMesh := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: kuma.io/v1alpha1
            kind: Mesh
            metadata:
              name: %s
`, "demo"))
			backend.Create(demoMesh)

			// and
			one := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: sample.test.kuma.io/v1alpha1
            kind: SampleTrafficRoute
            mesh: default
            metadata:
              namespace: %s
              name: %s
            spec:
              path: /example
`, ns, "one"))
			backend.Create(one)
			// and
			two := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: sample.test.kuma.io/v1alpha1
            kind: SampleTrafficRoute
            mesh: default
            metadata:
              namespace: %s
              name: %s
            spec:
              path: /another
`, ns, "two"))
			backend.Create(two)
			// and
			three := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: sample.test.kuma.io/v1alpha1
            kind: SampleTrafficRoute
            mesh: demo
            metadata:
              namespace: %s
              name: %s
            spec:
              path: /third
`, ns, "three"))
			backend.Create(three)

			By("listing resources from default mesh")
			// given
			trl := &sample_core.TrafficRouteResourceList{}

			// when
			err := s.List(context.Background(), trl, store.ListByMesh(mesh))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(trl.Pagination.Total).To(Equal(uint32(2)))
			// and
			Expect(trl.Items).To(HaveLen(2))

			// when
			actualResources := map[string]*sample_core.TrafficRouteResource{
				trl.Items[0].Meta.GetName(): trl.Items[0],
				trl.Items[1].Meta.GetName(): trl.Items[1],
			}
			// then
			actualResourceOne := actualResources[fmt.Sprintf("one.%s", ns)]
			Expect(actualResourceOne.Spec.Path).To(Equal("/example"))
			Expect(actualResourceOne.Meta.GetNameExtensions()).To(Equal(core_model.ResourceNameExtensions{
				"k8s.kuma.io/namespace": ns,
				"k8s.kuma.io/name":      "one",
			}))

			// and
			actualResourceTwo := actualResources[fmt.Sprintf("two.%s", ns)]
			Expect(actualResourceTwo.Spec.Path).To(Equal("/another"))
			Expect(actualResourceTwo.Meta.GetNameExtensions()).To(Equal(core_model.ResourceNameExtensions{
				"k8s.kuma.io/namespace": ns,
				"k8s.kuma.io/name":      "two",
			}))

			By("listing resources from demo mesh")

			// given
			trl = &sample_core.TrafficRouteResourceList{}

			// when
			err = s.List(context.Background(), trl, store.ListByMesh(name))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(trl.Pagination.Total).To(Equal(uint32(1)))
			// and
			Expect(trl.Items).To(HaveLen(1))

			// when
			actualResources = map[string]*sample_core.TrafficRouteResource{
				trl.Items[0].Meta.GetName(): trl.Items[0],
			}
			// then
			actualResourceThree := actualResources[fmt.Sprintf("three.%s", ns)]
			Expect(actualResourceThree.Spec.Path).To(Equal("/third"))
			Expect(actualResourceThree.Meta.GetNameExtensions()).To(Equal(core_model.ResourceNameExtensions{
				"k8s.kuma.io/namespace": ns,
				"k8s.kuma.io/name":      "three",
			}))

			// when
			err = s.Delete(context.Background(), core_mesh.NewMeshResource(), store.DeleteByKey("demo", core_model.NoMesh))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			backend.AssertNotExists(&mesh_k8s.Mesh{}, ns, "demo")

		})

		It("should return a list of matching Meshes", func() {
			// setup
			one := backend.ParseYAML(`
            apiVersion: kuma.io/v1alpha1
            kind: Mesh
            metadata:
              name: demo-1
`)
			backend.Create(one)
			// and
			two := backend.ParseYAML(`
            apiVersion: kuma.io/v1alpha1
            kind: Mesh
            metadata:
              name: demo-2
`)
			backend.Create(two)

			// given
			meshes := &core_mesh.MeshResourceList{}

			// when
			err := s.List(context.Background(), meshes)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(meshes.Items).To(HaveLen(2))

			// when
			actualResources := map[string]*core_mesh.MeshResource{
				meshes.Items[0].Meta.GetName(): meshes.Items[0],
				meshes.Items[1].Meta.GetName(): meshes.Items[1],
			}
			// then
			Expect(actualResources).To(HaveKey("demo-1"))
			Expect(actualResources["demo-1"].Meta.GetNameExtensions()).To(Equal(core_model.ResourceNameExtensions{
				"k8s.kuma.io/namespace": "",
				"k8s.kuma.io/name":      "demo-1",
			}))
			// and
			Expect(actualResources).To(HaveKey("demo-2"))
			Expect(actualResources["demo-2"].Meta.GetNameExtensions()).To(Equal(core_model.ResourceNameExtensions{
				"k8s.kuma.io/namespace": "",
				"k8s.kuma.io/name":      "demo-2",
			}))
		})
	})
})
