package k8s_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshtrace/api/v1alpha1"
	v1alpha1_k8s "github.com/kumahq/kuma/pkg/plugins/policies/meshtrace/k8s/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	k8s_registry "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	. "github.com/kumahq/kuma/pkg/test/matchers"
)

var _ = Describe("KubernetesStore", func() {
	var ks *k8s.KubernetesStore
	var s store.ResourceStore
	const name = "demo"
	const mesh = "default"

	backend := struct {
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
		Expect(kubeTypes.RegisterObjectType(&mesh_proto.TrafficRoute{}, &mesh_k8s.TrafficRoute{})).To(Succeed())
		Expect(kubeTypes.RegisterObjectType(&mesh_proto.Mesh{}, &mesh_k8s.Mesh{})).To(Succeed())
		Expect(kubeTypes.RegisterObjectType(&v1alpha1.MeshTrace{}, &v1alpha1_k8s.MeshTrace{})).To(Succeed())
		Expect(kubeTypes.RegisterListType(&mesh_proto.TrafficRoute{}, &mesh_k8s.TrafficRouteList{})).To(Succeed())
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
	})

	AfterEach(func() {
		err := k8sClient.DeleteAllOf(context.Background(), &mesh_k8s.TrafficRoute{})
		Expect(err).ToNot(HaveOccurred())
		err = k8sClient.DeleteAllOf(context.Background(), &mesh_k8s.Mesh{})
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("Create()", func() {
		It("should create a new resource", func() {
			// given
			tr := &core_mesh.TrafficRouteResource{
				Spec: &mesh_proto.TrafficRoute{
					Conf: &mesh_proto.TrafficRoute_Conf{
						Destination: map[string]string{
							"path": "/example",
						},
					},
				},
			}
			expected := backend.ParseYAML(`
            apiVersion: kuma.io/v1alpha1
            kind: TrafficRoute
            mesh: default
            spec:
              conf:
                destination:
                  path: /example
`).(*mesh_k8s.TrafficRoute)

			// when
			err := s.Create(context.Background(), tr, store.CreateByKey(name, mesh))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(tr.Meta.GetName()).To(Equal(name))
			Expect(tr.Meta.GetMesh()).To(Equal(mesh))
			Expect(tr.Meta.GetVersion()).ToNot(Equal(""))

			// when
			actual := mesh_k8s.TrafficRoute{}
			backend.Get(&actual, "", name)

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
			backend.Get(&actual, "", name)

			// then
			Expect(actual.Spec).To(Equal(expected.Spec))
			// and
			Expect(actual.ObjectMeta.ResourceVersion).To(Equal(mesh.Meta.GetVersion()))
		})

		It("should not create a duplicate resource", func() {
			// setup
			backend.AssertNotExists(&mesh_k8s.TrafficRoute{}, "", name)

			// when
			err := s.Create(context.Background(), core_mesh.NewTrafficRouteResource(), store.CreateByKey(name, mesh))

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			err = s.Create(context.Background(), core_mesh.NewTrafficRouteResource(), store.CreateByKey(name, mesh))

			// then
			Expect(err).To(MatchError(store.ErrorResourceAlreadyExists(core_mesh.TrafficRouteType, name, mesh)))
		})

		It("should set owner reference", func() {
			// setup
			mesh := core_mesh.NewMeshResource()
			err := s.Create(context.Background(), mesh, store.CreateByKey("mesh", core_model.NoMesh))
			Expect(err).ToNot(HaveOccurred())

			err = k8sClient.Get(context.Background(), client.ObjectKey{Name: "mesh"}, &mesh_k8s.Mesh{})
			Expect(err).ToNot(HaveOccurred())

			tr := core_mesh.TrafficRouteResource{
				Spec: &mesh_proto.TrafficRoute{
					Conf: &mesh_proto.TrafficRoute_Conf{
						Destination: map[string]string{
							"path": "/example",
						},
					},
				},
			}
			// when
			err = s.Create(context.Background(), &tr, store.CreateByKey(name, "mesh"), store.CreateWithOwner(mesh))
			Expect(err).ToNot(HaveOccurred())

			// then
			obj := mesh_k8s.TrafficRoute{}
			err = k8sClient.Get(context.Background(), client.ObjectKey{Name: name}, &obj)
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
            apiVersion: kuma.io/v1alpha1
            kind: TrafficRoute
            mesh: default
            metadata:
              name: %s
            spec:
              conf:
                destination:
                  path: /example
`, name))
			backend.Create(initial)
			// and
			expected := backend.ParseYAML(`
            apiVersion: kuma.io/v1alpha1
            kind: TrafficRoute
            mesh: default
            spec:
              conf:
                destination:
                  path: /another
`).(*mesh_k8s.TrafficRoute)

			// given
			tr := core_mesh.NewTrafficRouteResource()

			// when
			err := s.Get(context.Background(), tr, store.GetByKey(name, mesh))
			// then
			Expect(err).ToNot(HaveOccurred())
			version := tr.Meta.GetVersion()

			// when
			tr.Spec.Conf.Destination["path"] = "/another"
			err = s.Update(context.Background(), tr)

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(tr.Meta.GetVersion()).ToNot(Equal(""))
			Expect(tr.Meta.GetVersion()).ToNot(Equal(version))

			// when
			actual := mesh_k8s.TrafficRoute{}
			backend.Get(&actual, "", name)

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
			backend.Get(&actual, "", name)

			// then
			Expect(actual.Spec).To(Equal(expected.Spec))
			// and
			Expect(actual.ObjectMeta.ResourceVersion).To(Equal(mesh.Meta.GetVersion()))
		})

		It("should return an error if resource is not found", func() {
			// setup
			initial := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: kuma.io/v1alpha1
            kind: TrafficRoute
            mesh: default
            metadata:
              name: %s
`, name))
			backend.Create(initial)

			// given
			tr := core_mesh.NewTrafficRouteResource()

			// when
			err := s.Get(context.Background(), tr, store.GetByKey(name, mesh))
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			backend.Delete(initial)
			// and
			err = s.Update(context.Background(), tr)

			// then
			Expect(err).To(MatchError(store.ErrorResourceConflict(core_mesh.TrafficRouteType, name, mesh)))
		})

		It("should return an error if resource has changed", func() {
			// setup
			initial := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: kuma.io/v1alpha1
            kind: TrafficRoute
            mesh: default
            metadata:
              name: %s
            spec:
              conf: {}
`, name))
			backend.Create(initial)

			// given
			tr1 := core_mesh.NewTrafficRouteResource()

			// when
			err := s.Get(context.Background(), tr1, store.GetByKey(name, mesh))
			// then
			Expect(err).ToNot(HaveOccurred())

			// given
			tr2 := core_mesh.NewTrafficRouteResource()

			// when
			err = s.Get(context.Background(), tr2, store.GetByKey(name, mesh))
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			tr1.Spec.Conf.Destination = map[string]string{
				"path": "/example",
			}
			err = s.Update(context.Background(), tr1)
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			tr2.Spec.Conf.Destination = map[string]string{
				"path": "/another",
			}
			err = s.Update(context.Background(), tr2)
			// then
			Expect(err).To(MatchError(store.ErrorResourceConflict(core_mesh.TrafficRouteType, name, mesh)))
		})
	})

	Describe("Get()", func() {
		It("should return an error if resource is not found", func() {
			// setup
			backend.AssertNotExists(&mesh_k8s.TrafficRoute{}, "", name)

			// when
			err := s.Get(context.Background(), core_mesh.NewTrafficRouteResource(), store.GetByKey(name, mesh))

			// then
			Expect(err).To(MatchError(store.ErrorResourceNotFound(core_mesh.TrafficRouteType, name, mesh)))
		})

		It("should return an error if namespaced resource is not in the right format", func() {
			// when
			err := ks.Get(context.Background(), v1alpha1.NewMeshTraceResource(), store.GetByKey(name, mesh))

			// then
			Expect(err.Error()).To(ContainSubstring("must include namespace after the dot"))
		})

		It("should return an error if resource name is empty", func() {
			// when
			err := ks.Get(context.Background(), v1alpha1.NewMeshTraceResource(), store.GetByKey("", mesh))

			// then
			Expect(err.Error()).To(Equal("invalid: name can't be empty"))
		})

		It("should return an existing resource", func() {
			// setup
			expected := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: kuma.io/v1alpha1
            kind: TrafficRoute
            mesh: default
            metadata:
              name: %s
            spec:
              conf:
                destination:
                  path: /example
`, name))
			backend.Create(expected)

			// given
			actual := core_mesh.NewTrafficRouteResource()

			// when
			err := s.Get(context.Background(), actual, store.GetByKey(name, mesh))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual.Meta.GetName()).To(Equal(name))
			// and
			Expect(actual.Meta.GetNameExtensions()).To(Equal(core_model.ResourceNameExtensions{
				"k8s.kuma.io/namespace": "",
				"k8s.kuma.io/name":      name,
			}))
			Expect(actual.Meta.GetLabels()[mesh_proto.DisplayName]).To(Equal(name))
			// and
			Expect(actual.Spec.Conf.Destination["path"]).To(Equal("/example"))
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

		It("should return a display name from annotation", func() {
			// setup
			expected := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: kuma.io/v1alpha1
            kind: TrafficRoute
            mesh: default
            metadata:
              annotations:
                kuma.io/display-name: dn
                k8s.kuma.io/service-account: default
              name: %s
            spec:
              conf:
                destination:
                  path: /example
`, name))
			backend.Create(expected)

			// given
			actual := core_mesh.NewTrafficRouteResource()

			// when
			err := s.Get(context.Background(), actual, store.GetByKey(name, mesh))

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(actual.Meta.GetLabels()[mesh_proto.DisplayName]).To(Equal("dn"))
			Expect(actual.Meta.GetLabels()[metadata.KumaServiceAccount]).To(Equal("default"))
		})
	})

	Describe("Delete()", func() {
		It("should return en error if resource is not found", func() {
			// setup
			backend.AssertNotExists(&mesh_k8s.TrafficRoute{}, "", name)
			resource := core_mesh.NewTrafficRouteResource()

			// when
			err := s.Delete(context.Background(), resource, store.DeleteByKey(name, mesh))

			// then
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(store.ErrorResourceNotFound(resource.Descriptor().Name, name, mesh)))
		})

		It("should delete an existing resource", func() {
			// setup
			initial := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: kuma.io/v1alpha1
            kind: TrafficRoute
            mesh: default
            metadata:
              name: %s
`, name))
			backend.Create(initial)

			// when
			err := s.Delete(context.Background(), core_mesh.NewTrafficRouteResource(), store.DeleteByKey(name, mesh))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			backend.AssertNotExists(&mesh_k8s.TrafficRoute{}, "", name)
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
			backend.AssertNotExists(&mesh_k8s.Mesh{}, "", "demo")
		})
	})

	Describe("List()", func() {
		It("should return an empty list if there are no matching resources", func() {
			// given
			trl := &core_mesh.TrafficRouteResourceList{}

			// when
			err := s.List(context.Background(), trl, store.ListByMesh(mesh))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(trl.Items).To(BeEmpty())
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
            apiVersion: kuma.io/v1alpha1
            kind: TrafficRoute
            mesh: default
            metadata:
              name: %s
            spec:
              conf:
                destination:
                  path: /example
`, "one"))
			backend.Create(one)
			// and
			two := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: kuma.io/v1alpha1
            kind: TrafficRoute
            mesh: default
            metadata:
              name: %s
            spec:
              conf:
                destination:
                  path: /another
`, "two"))
			backend.Create(two)
			// and
			three := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: kuma.io/v1alpha1
            kind: TrafficRoute
            mesh: demo
            metadata:
              name: %s
            spec:
              conf:
                destination:
                  path: /third
`, "three"))
			backend.Create(three)

			By("listing resources from default mesh")
			// given
			trl := &core_mesh.TrafficRouteResourceList{}

			// when
			err := s.List(context.Background(), trl, store.ListByMesh(mesh))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(trl.Pagination.Total).To(Equal(uint32(2)))
			// and
			Expect(trl.Items).To(HaveLen(2))

			// when
			actualResources := map[string]*core_mesh.TrafficRouteResource{
				trl.Items[0].Meta.GetName(): trl.Items[0],
				trl.Items[1].Meta.GetName(): trl.Items[1],
			}
			// then
			actualResourceOne := actualResources["one"]
			Expect(actualResourceOne.Spec.Conf.Destination["path"]).To(Equal("/example"))
			Expect(actualResourceOne.Meta.GetNameExtensions()).To(Equal(core_model.ResourceNameExtensions{
				"k8s.kuma.io/namespace": "",
				"k8s.kuma.io/name":      "one",
			}))

			// and
			actualResourceTwo := actualResources["two"]
			Expect(actualResourceTwo.Spec.Conf.Destination["path"]).To(Equal("/another"))
			Expect(actualResourceTwo.Meta.GetNameExtensions()).To(Equal(core_model.ResourceNameExtensions{
				"k8s.kuma.io/namespace": "",
				"k8s.kuma.io/name":      "two",
			}))

			By("listing resources from demo mesh")

			// given
			trl = &core_mesh.TrafficRouteResourceList{}

			// when
			err = s.List(context.Background(), trl, store.ListByMesh(name))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(trl.Pagination.Total).To(Equal(uint32(1)))
			// and
			Expect(trl.Items).To(HaveLen(1))

			// when
			actualResources = map[string]*core_mesh.TrafficRouteResource{
				trl.Items[0].Meta.GetName(): trl.Items[0],
			}
			// then
			actualResourceThree := actualResources["three"]
			Expect(actualResourceThree.Spec.Conf.Destination["path"]).To(Equal("/third"))
			Expect(actualResourceThree.Meta.GetNameExtensions()).To(Equal(core_model.ResourceNameExtensions{
				"k8s.kuma.io/namespace": "",
				"k8s.kuma.io/name":      "three",
			}))

			// when
			err = s.Delete(context.Background(), core_mesh.NewMeshResource(), store.DeleteByKey("demo", core_model.NoMesh))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			backend.AssertNotExists(&mesh_k8s.Mesh{}, "", "demo")
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
