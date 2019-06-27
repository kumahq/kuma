package k8s_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/k8s"
	k8s_registry "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/k8s/native/pkg/registry"
	sample_k8s "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/k8s/native/test/api/sample/v1alpha1"
	sample_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/test/apis/sample/v1alpha1"
	sample_core "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/test/resources/apis/sample"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/uuid"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("KubernetesStore", func() {

	var ks *k8s.KubernetesStore
	var s store.ResourceStore
	var ns string // each test should run in a dedicated k8s namespace
	var name = "demo"

	var backend = struct {
		ParseYAML       func(yaml string) runtime.Object
		Create          func(obj runtime.Object)
		Get             func(obj runtime.Object, ns, name string)
		AssertNotExists func(obj runtime.Object, ns, name string)
		Delete          func(obj runtime.Object)
	}{
		ParseYAML: func(text string) runtime.Object {
			// setup
			decoder := serializer.NewCodecFactory(k8sClientScheme).UniversalDeserializer()
			// when
			obj, _, err := decoder.Decode([]byte(text), nil, nil)
			// then
			Expect(err).ToNot(HaveOccurred())
			return obj
		},
		Create: func(obj runtime.Object) {
			// when
			err := k8sClient.Create(context.Background(), obj)
			// then
			Expect(err).ToNot(HaveOccurred())
		},
		Get: func(obj runtime.Object, ns, name string) {
			// when
			err := k8sClient.Get(context.Background(), client.ObjectKey{Namespace: ns, Name: name}, obj)
			// then
			Expect(err).ToNot(HaveOccurred())
		},
		AssertNotExists: func(obj runtime.Object, ns, name string) {
			// when
			err := k8sClient.Get(context.Background(), client.ObjectKey{Namespace: ns, Name: name}, obj)
			// then
			Expect(apierrs.IsNotFound(err)).To(BeTrue())
		},
		Delete: func(obj runtime.Object) {
			// when
			err := k8sClient.Delete(context.Background(), obj)
			// then
			Expect(err).ToNot(HaveOccurred())
		},
	}

	BeforeEach(func() {
		kubeTypes := k8s_registry.NewTypeRegistry()
		Expect(kubeTypes.RegisterObjectType(&sample_proto.TrafficRoute{}, &sample_k8s.TrafficRoute{})).To(Succeed())
		Expect(kubeTypes.RegisterListType(&sample_proto.TrafficRoute{}, &sample_k8s.TrafficRouteList{})).To(Succeed())

		ks = &k8s.KubernetesStore{
			Client: k8sClient,
			Converter: &k8s.SimpleConverter{
				KubeFactory: &k8s.SimpleKubeFactory{
					KubeTypes: kubeTypes,
				},
			},
		}
		s = store.NewStrictResourceStore(ks)
		ns = string(uuid.NewUUID())
	})

	Describe("Create()", func() {
		It("should create a new resource", func() {
			// given
			tr := &sample_core.TrafficRouteResource{
				Spec: sample_proto.TrafficRoute{
					Path: "/example",
				},
			}
			expected := backend.ParseYAML(`
            apiVersion: sample.test.getkonvoy.io/v1alpha1
            kind: TrafficRoute
            spec:
              path: /example
`).(*sample_k8s.TrafficRoute)

			// when
			err := s.Create(context.Background(), tr, store.CreateByName(ns, name))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(tr.Meta.GetNamespace()).To(Equal(ns))
			Expect(tr.Meta.GetName()).To(Equal(name))
			Expect(tr.Meta.GetVersion()).ToNot(Equal(""))

			// when
			actual := sample_k8s.TrafficRoute{}
			backend.Get(&actual, ns, name)

			// then
			Expect(actual.Spec).To(Equal(expected.Spec))
			// and
			Expect(actual.ObjectMeta.ResourceVersion).To(Equal(tr.Meta.GetVersion()))
		})

		It("should not create a duplicate resource", func() {
			// setup
			backend.AssertNotExists(&sample_k8s.TrafficRoute{}, ns, name)

			// when
			err := s.Create(context.Background(), &sample_core.TrafficRouteResource{}, store.CreateByName(ns, name))

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			err = s.Create(context.Background(), &sample_core.TrafficRouteResource{}, store.CreateByName(ns, name))

			// then
			Expect(err).To(MatchError(store.ErrorResourceAlreadyExists(sample_core.TrafficRouteType, ns, name)))
		})
	})

	Describe("Update()", func() {
		It("should update an existing resource", func() {
			// setup
			initial := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: sample.test.getkonvoy.io/v1alpha1
            kind: TrafficRoute
            metadata:
              namespace: %s
              name: %s
            spec:
              path: /example
`, ns, name))
			backend.Create(initial)
			// and
			expected := backend.ParseYAML(`
            apiVersion: sample.test.getkonvoy.io/v1alpha1
            kind: TrafficRoute
            spec:
              path: /another
`).(*sample_k8s.TrafficRoute)

			// given
			tr := &sample_core.TrafficRouteResource{}

			// when
			err := s.Get(context.Background(), tr, store.GetByName(ns, name))
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
			actual := sample_k8s.TrafficRoute{}
			backend.Get(&actual, ns, name)

			// then
			Expect(actual.Spec).To(Equal(expected.Spec))
			// and
			Expect(actual.ObjectMeta.ResourceVersion).To(Equal(tr.Meta.GetVersion()))
		})

		It("should return an error if resource is not found", func() {
			// setup
			initial := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: sample.test.getkonvoy.io/v1alpha1
            kind: TrafficRoute
            metadata:
              namespace: %s
              name: %s
`, ns, name))
			backend.Create(initial)

			// given
			tr := &sample_core.TrafficRouteResource{}

			// when
			err := s.Get(context.Background(), tr, store.GetByName(ns, name))
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			backend.Delete(initial)
			// and
			err = s.Update(context.Background(), tr)

			// then
			Expect(err).To(MatchError(store.ErrorResourceConflict(sample_core.TrafficRouteType, ns, name)))
		})

		It("should return an error if resource has changed", func() {
			// setup
			initial := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: sample.test.getkonvoy.io/v1alpha1
            kind: TrafficRoute
            metadata:
              namespace: %s
              name: %s
`, ns, name))
			backend.Create(initial)

			// given
			tr1 := &sample_core.TrafficRouteResource{}

			// when
			err := s.Get(context.Background(), tr1, store.GetByName(ns, name))
			// then
			Expect(err).ToNot(HaveOccurred())

			// given
			tr2 := &sample_core.TrafficRouteResource{}

			// when
			err = s.Get(context.Background(), tr2, store.GetByName(ns, name))
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
			Expect(err).To(MatchError(store.ErrorResourceConflict(sample_core.TrafficRouteType, ns, name)))
		})
	})

	Describe("Get()", func() {
		It("should return an error if resource is not found", func() {
			// setup
			backend.AssertNotExists(&sample_k8s.TrafficRoute{}, ns, name)

			// when
			err := s.Get(context.Background(), &sample_core.TrafficRouteResource{}, store.GetByName(ns, name))

			// then
			Expect(err).To(MatchError(store.ErrorResourceNotFound(sample_core.TrafficRouteType, ns, name)))
		})

		It("should return an existing resource", func() {
			// setup
			expected := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: sample.test.getkonvoy.io/v1alpha1
            kind: TrafficRoute
            metadata:
              namespace: %s
              name: %s
            spec:
              path: /example
`, ns, name))
			backend.Create(expected)

			// given
			actual := &sample_core.TrafficRouteResource{}

			// when
			err := s.Get(context.Background(), actual, store.GetByName(ns, name))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual.Meta.GetNamespace()).To(Equal(ns))
			Expect(actual.Meta.GetName()).To(Equal(name))
			// and
			Expect(actual.Spec.Path).To(Equal("/example"))
		})
	})

	Describe("Delete()", func() {
		It("should succeed if resource is not found", func() {
			// setup
			backend.AssertNotExists(&sample_k8s.TrafficRoute{}, ns, name)

			// when
			err := s.Delete(context.Background(), &sample_core.TrafficRouteResource{}, store.DeleteByName(ns, name))

			// then
			Expect(err).ToNot(HaveOccurred())
		})

		It("should delete an existing resource", func() {
			// setup
			initial := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: sample.test.getkonvoy.io/v1alpha1
            kind: TrafficRoute
            metadata:
              namespace: %s
              name: %s
`, ns, name))
			backend.Create(initial)

			// when
			err := s.Delete(context.Background(), &sample_core.TrafficRouteResource{}, store.DeleteByName(ns, name))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			backend.AssertNotExists(&sample_k8s.TrafficRoute{}, ns, name)
		})
	})

	Describe("List()", func() {
		It("should return an empty list if there are no matching resources", func() {
			// given
			trl := &sample_core.TrafficRouteResourceList{}

			// when
			err := s.List(context.Background(), trl, store.ListByNamespace(ns))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(trl.Items).To(HaveLen(0))
		})

		It("should return a list of matching resource", func() {
			// setup
			one := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: sample.test.getkonvoy.io/v1alpha1
            kind: TrafficRoute
            metadata:
              namespace: %s
              name: %s
            spec:
              path: /example
`, ns, "one"))
			backend.Create(one)
			// and
			two := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: sample.test.getkonvoy.io/v1alpha1
            kind: TrafficRoute
            metadata:
              namespace: %s
              name: %s
            spec:
              path: /another
`, ns, "two"))
			backend.Create(two)

			// given
			trl := &sample_core.TrafficRouteResourceList{}

			// when
			err := s.List(context.Background(), trl, store.ListByNamespace(ns))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(trl.Items).To(HaveLen(2))

			// when
			items := map[string]*sample_core.TrafficRouteResource{
				trl.Items[0].Meta.GetName(): trl.Items[0],
				trl.Items[1].Meta.GetName(): trl.Items[1],
			}
			// then
			Expect(items["one"].Meta.GetNamespace()).To(Equal(ns))
			Expect(items["one"].Spec.Path).To(Equal("/example"))
			// and
			Expect(items["two"].Meta.GetNamespace()).To(Equal(ns))
			Expect(items["two"].Spec.Path).To(Equal("/another"))
		})
	})
})
