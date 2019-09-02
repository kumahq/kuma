package k8s_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/secrets/k8s"

	core_system "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/system"
	secret_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/secrets/model"
	secret_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/secrets/store"
	proto_types "github.com/gogo/protobuf/types"

	kube_core "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/uuid"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("KubernetesStore", func() {

	var s secret_store.SecretStore
	var ns string // each test should run in a dedicated k8s namespace
	const name = "demo"
	const noMesh = ""

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
		var err error
		s, err = k8s.NewStore(k8sClient)
		Expect(err).ToNot(HaveOccurred())

		ns = string(uuid.NewUUID())
	})

	Describe("Create()", func() {
		It("should create a new secret", func() {
			// given
			secret := &secret_model.Secret{
				Spec: proto_types.BytesValue{
					Value: []byte("example"),
				},
			}
			expected := backend.ParseYAML(`
            apiVersion: v1
            kind: Secret
            type: system.getkonvoy.io/secret
            data:
              value: ZXhhbXBsZQ== # base64(example)
`).(*kube_core.Secret)

			// when
			err := s.Create(context.Background(), secret, store.CreateByKey(ns, name, noMesh))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(secret.Meta.GetNamespace()).To(Equal(ns))
			Expect(secret.Meta.GetName()).To(Equal(name))
			Expect(secret.Meta.GetMesh()).To(Equal(""))
			Expect(secret.Meta.GetVersion()).ToNot(Equal(""))

			// when
			actual := kube_core.Secret{}
			backend.Get(&actual, ns, name)

			// then
			Expect(actual.Data).To(Equal(expected.Data))
			Expect(actual.Type).To(Equal(expected.Type))
			// and
			Expect(actual.ObjectMeta.ResourceVersion).To(Equal(secret.Meta.GetVersion()))
		})

		It("should not create a duplicate resource", func() {
			// setup
			backend.AssertNotExists(&kube_core.Secret{}, ns, name)

			// when
			err := s.Create(context.Background(), &secret_model.Secret{}, store.CreateByKey(ns, name, noMesh))

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			err = s.Create(context.Background(), &secret_model.Secret{}, store.CreateByKey(ns, name, noMesh))

			// then
			Expect(err).To(MatchError(store.ErrorResourceAlreadyExists(core_system.SecretType, ns, name, noMesh)))
		})
	})

	Describe("Update()", func() {
		It("should update an existing resource", func() {
			// setup
			initial := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: v1
            kind: Secret
            type: system.getkonvoy.io/secret
            metadata:
              namespace: %s
              name: %s
            data:
              value: ZXhhbXBsZQ== # base64(example)
`, ns, name))
			backend.Create(initial)
			// and
			expected := backend.ParseYAML(`
            apiVersion: v1
            kind: Secret
            type: system.getkonvoy.io/secret
            data:
              value: YW5vdGhlcg== # base64(another)
`).(*kube_core.Secret)

			// given
			secret := &secret_model.Secret{}

			// when
			err := s.Get(context.Background(), secret, store.GetByKey(ns, name, noMesh))
			// then
			Expect(err).ToNot(HaveOccurred())
			version := secret.Meta.GetVersion()

			// when
			secret.Spec.Value = []byte("another")
			err = s.Update(context.Background(), secret)

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(secret.Meta.GetVersion()).ToNot(Equal(""))
			Expect(secret.Meta.GetVersion()).ToNot(Equal(version))

			// when
			actual := kube_core.Secret{}
			backend.Get(&actual, ns, name)

			// then
			Expect(actual.Data).To(Equal(expected.Data))
			Expect(actual.Type).To(Equal(expected.Type))
			// and
			Expect(actual.ObjectMeta.ResourceVersion).To(Equal(secret.Meta.GetVersion()))
		})

		It("should return an error if resource is not found", func() {
			// setup
			initial := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: v1
            kind: Secret
            type: system.getkonvoy.io/secret
            metadata:
              namespace: %s
              name: %s
`, ns, name))
			backend.Create(initial)

			// given
			secret := &secret_model.Secret{}

			// when
			err := s.Get(context.Background(), secret, store.GetByKey(ns, name, noMesh))
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			backend.Delete(initial)
			// and
			err = s.Update(context.Background(), secret)

			// then
			Expect(err).To(MatchError(store.ErrorResourceConflict(core_system.SecretType, ns, name, noMesh)))
		})

		It("should return an error if resource has changed", func() {
			// setup
			initial := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: v1
            kind: Secret
            type: system.getkonvoy.io/secret
            metadata:
              namespace: %s
              name: %s
`, ns, name))
			backend.Create(initial)

			// given
			secret1 := &secret_model.Secret{}

			// when
			err := s.Get(context.Background(), secret1, store.GetByKey(ns, name, noMesh))
			// then
			Expect(err).ToNot(HaveOccurred())

			// given
			secret2 := &secret_model.Secret{}

			// when
			err = s.Get(context.Background(), secret2, store.GetByKey(ns, name, noMesh))
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			secret1.Spec.Value = []byte("example")
			err = s.Update(context.Background(), secret1)
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			secret2.Spec.Value = []byte("another")
			err = s.Update(context.Background(), secret2)
			// then
			Expect(err).To(MatchError(store.ErrorResourceConflict(core_system.SecretType, ns, name, noMesh)))
		})
	})

	Describe("Get()", func() {
		It("should return an error if resource is not found", func() {
			// setup
			backend.AssertNotExists(&kube_core.Secret{}, ns, name)

			// when
			err := s.Get(context.Background(), &secret_model.Secret{}, store.GetByKey(ns, name, noMesh))

			// then
			Expect(err).To(MatchError(store.ErrorResourceNotFound(core_system.SecretType, ns, name, noMesh)))
		})

		It("should return an existing resource", func() {
			// setup
			expected := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: v1
            kind: Secret
            type: system.getkonvoy.io/secret
            metadata:
              namespace: %s
              name: %s
            data:
              value: ZXhhbXBsZQ== # base64(example)
`, ns, name))
			backend.Create(expected)

			// given
			actual := &secret_model.Secret{}

			// when
			err := s.Get(context.Background(), actual, store.GetByKey(ns, name, noMesh))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual.Meta.GetNamespace()).To(Equal(ns))
			Expect(actual.Meta.GetName()).To(Equal(name))
			// and
			Expect(actual.Spec.Value).To(Equal([]byte("example")))
		})
	})

	Describe("Delete()", func() {
		It("should succeed if resource is not found", func() {
			// setup
			backend.AssertNotExists(&kube_core.Secret{}, ns, name)

			// when
			err := s.Delete(context.Background(), &secret_model.Secret{}, store.DeleteByKey(ns, name, noMesh))

			// then
			Expect(err).ToNot(HaveOccurred())
		})

		It("should delete an existing resource", func() {
			// setup
			initial := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: v1
            kind: Secret
            type: system.getkonvoy.io/secret
            metadata:
              namespace: %s
              name: %s
`, ns, name))
			backend.Create(initial)

			// when
			err := s.Delete(context.Background(), &secret_model.Secret{}, store.DeleteByKey(ns, name, noMesh))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			backend.AssertNotExists(&kube_core.Secret{}, ns, name)
		})
	})

	Describe("List()", func() {
		It("should return an empty list if there are no matching resources", func() {
			// given
			secrets := &secret_model.SecretList{}

			// when
			err := s.List(context.Background(), secrets, store.ListByNamespace(ns), store.ListByMesh(noMesh))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(secrets.Items).To(HaveLen(0))
		})

		It("should return a list of matching resource", func() {
			// setup
			one := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: v1
            kind: Secret
            type: system.getkonvoy.io/secret
            metadata:
              namespace: %s
              name: %s
            data:
              value: ZXhhbXBsZQ== # base64(example)
`, ns, "one"))
			backend.Create(one)
			// and
			two := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: v1
            kind: Secret
            type: system.getkonvoy.io/secret
            metadata:
              namespace: %s
              name: %s
            data:
              value: YW5vdGhlcg== # base64(another)
`, ns, "two"))
			backend.Create(two)

			// given
			secrets := &secret_model.SecretList{}

			// when
			err := s.List(context.Background(), secrets, store.ListByNamespace(ns))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(secrets.Items).To(HaveLen(2))

			// when
			items := map[string]*secret_model.Secret{
				secrets.Items[0].Meta.GetName(): secrets.Items[0],
				secrets.Items[1].Meta.GetName(): secrets.Items[1],
			}
			// then
			Expect(items["one"].Meta.GetNamespace()).To(Equal(ns))
			Expect(items["one"].Spec.Value).To(Equal([]byte("example")))
			// and
			Expect(items["two"].Meta.GetNamespace()).To(Equal(ns))
			Expect(items["two"].Spec.Value).To(Equal([]byte("another")))
		})
	})
})
