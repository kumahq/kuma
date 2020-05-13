package k8s_test

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/golang/protobuf/ptypes/wrappers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/uuid"
	"sigs.k8s.io/controller-runtime/pkg/client"

	system_proto "github.com/Kong/kuma/api/system/v1alpha1"
	core_system "github.com/Kong/kuma/pkg/core/resources/apis/system"
	secret_model "github.com/Kong/kuma/pkg/core/resources/apis/system"
	"github.com/Kong/kuma/pkg/core/resources/store"
	secret_store "github.com/Kong/kuma/pkg/core/secrets/store"
	"github.com/Kong/kuma/pkg/plugins/secrets/k8s"
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
		ns = string(uuid.NewUUID())

		err := k8sClient.Create(context.Background(), &kube_core.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: ns},
		})
		Expect(err).ToNot(HaveOccurred())

		s, err = k8s.NewStore(k8sClient, k8sClient, ns)
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("Create()", func() {
		It("should create a new secret", func() {
			// given
			secret := &secret_model.SecretResource{
				Spec: system_proto.Secret{
					Data: &wrappers.BytesValue{
						Value: []byte("example"),
					},
				},
			}
			expected := backend.ParseYAML(`
            apiVersion: v1
            kind: Secret
            type: system.kuma.io/secret
            metadata:
              labels:
                kuma.io/mesh: demo
            data:
              value: ZXhhbXBsZQ== # base64(example)
`).(*kube_core.Secret)

			// when
			err := s.Create(context.Background(), secret, store.CreateByKey(name, "demo"))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(secret.Meta.GetName()).To(Equal(name))
			Expect(secret.Meta.GetMesh()).To(Equal("demo"))
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
			backend.AssertNotExists(&kube_core.Secret{}, "ignored", name)

			// when
			err := s.Create(context.Background(), &secret_model.SecretResource{}, store.CreateByKey(name, noMesh))

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			err = s.Create(context.Background(), &secret_model.SecretResource{}, store.CreateByKey(name, noMesh))

			// then
			Expect(err).To(MatchError(store.ErrorResourceAlreadyExists(core_system.SecretType, name, noMesh)))
		})
	})

	Describe("Update()", func() {
		It("should update an existing resource", func() {
			// setup
			initial := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: v1
            kind: Secret
            type: system.kuma.io/secret
            metadata:
              namespace: %s
              name: %s
              labels:
                kuma.io/mesh: demo
            data:
              value: ZXhhbXBsZQ== # base64(example)
`, ns, name))
			backend.Create(initial)
			// and
			expected := backend.ParseYAML(`
            apiVersion: v1
            kind: Secret
            type: system.kuma.io/secret
            data:
              value: YW5vdGhlcg== # base64(another)
`).(*kube_core.Secret)

			// given
			secret := &secret_model.SecretResource{}

			// when
			err := s.Get(context.Background(), secret, store.GetByKey(name, "demo"))
			// then
			Expect(err).ToNot(HaveOccurred())
			version := secret.Meta.GetVersion()

			// when
			secret.Spec.Data.Value = []byte("another")
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
            type: system.kuma.io/secret
            metadata:
              namespace: %s
              name: %s
              labels:
                kuma.io/mesh: demo
`, ns, name))
			backend.Create(initial)

			// given
			secret := &secret_model.SecretResource{}

			// when
			err := s.Get(context.Background(), secret, store.GetByKey(name, "demo"))
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			backend.Delete(initial)
			// and
			err = s.Update(context.Background(), secret)

			// then
			Expect(err).To(MatchError(store.ErrorResourceConflict(core_system.SecretType, name, "demo")))
		})

		It("should return an error if resource has changed", func() {
			// setup
			initial := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: v1
            kind: Secret
            type: system.kuma.io/secret
            metadata:
              namespace: %s
              name: %s
              labels:
                kuma.io/mesh: demo             
`, ns, name))
			backend.Create(initial)

			// given
			secret1 := &secret_model.SecretResource{}

			// when
			err := s.Get(context.Background(), secret1, store.GetByKey(name, "demo"))
			// then
			Expect(err).ToNot(HaveOccurred())

			// given
			secret2 := &secret_model.SecretResource{}

			// when
			err = s.Get(context.Background(), secret2, store.GetByKey(name, "demo"))
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			secret1.Spec = system_proto.Secret{
				Data: &wrappers.BytesValue{
					Value: []byte("example"),
				},
			}
			err = s.Update(context.Background(), secret1)
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			secret2.Spec = system_proto.Secret{
				Data: &wrappers.BytesValue{
					Value: []byte("another"),
				},
			}
			err = s.Update(context.Background(), secret2)
			// then
			Expect(err).To(MatchError(store.ErrorResourceConflict(core_system.SecretType, name, "demo")))
		})
	})

	Describe("Get()", func() {
		It("should return an error if resource is not found", func() {
			// setup
			backend.AssertNotExists(&kube_core.Secret{}, ns, name)

			// when
			err := s.Get(context.Background(), &secret_model.SecretResource{}, store.GetByKey(name, "demo"))

			// then
			Expect(err).To(MatchError(store.ErrorResourceNotFound(core_system.SecretType, name, "demo")))
		})

		It("should return an existing resource with explicit mesh label", func() {
			// setup
			expected := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: v1
            kind: Secret
            type: system.kuma.io/secret
            metadata:
              namespace: %s
              name: %s
              labels:
                kuma.io/mesh: demo
            data:
              value: ZXhhbXBsZQ== # base64(example)
`, ns, name))
			backend.Create(expected)

			// given
			actual := &secret_model.SecretResource{}

			// when
			err := s.Get(context.Background(), actual, store.GetByKey(name, "demo"))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual.Meta.GetName()).To(Equal(name))
			Expect(actual.Meta.GetMesh()).To(Equal("demo"))
			// and
			Expect(actual.Spec.Data.Value).To(Equal([]byte("example")))
		})

		It("should return an existing resource with implicit default mesh", func() {
			// setup
			expected := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: v1
            kind: Secret
            type: system.kuma.io/secret
            metadata:
              namespace: %s
              name: %s
            data:
              value: ZXhhbXBsZQ== # base64(example)
`, ns, name))
			backend.Create(expected)

			// given
			actual := &secret_model.SecretResource{}

			// when
			err := s.Get(context.Background(), actual, store.GetByKey(name, "default"))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual.Meta.GetName()).To(Equal(name))
			Expect(actual.Meta.GetMesh()).To(Equal("default"))
			// and
			Expect(actual.Spec.Data.Value).To(Equal([]byte("example")))
		})

		It("should return an error if resource is in another mesh", func() {
			// setup
			expected := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: v1
            kind: Secret
            type: system.kuma.io/secret
            metadata:
              namespace: %s
              name: %s
              labels:
                kuma.io/mesh: demo
            data:
              value: ZXhhbXBsZQ== # base64(example)
`, ns, name))
			backend.Create(expected)

			// given
			actual := &secret_model.SecretResource{}

			// when
			err := s.Get(context.Background(), actual, store.GetByKey(name, "another-mesh"))

			// then
			Expect(err).To(MatchError(store.ErrorResourceNotFound(core_system.SecretType, name, "another-mesh")))
			Expect(actual.Spec.GetData().GetValue()).To(BeEmpty())
		})
	})

	Describe("Delete()", func() {
		It("should succeed if resource is not found", func() {
			// setup
			backend.AssertNotExists(&kube_core.Secret{}, ns, name)

			// when
			err := s.Delete(context.Background(), &secret_model.SecretResource{}, store.DeleteByKey(name, "demo"))

			// then
			Expect(err).ToNot(HaveOccurred())
		})

		It("should delete an existing resource", func() {
			// setup
			initial := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: v1
            kind: Secret
            type: system.kuma.io/secret
            metadata:
              namespace: %s
              name: %s
              labels:
                kuma.io/mesh: demo
`, ns, name))
			backend.Create(initial)

			// when
			err := s.Delete(context.Background(), &secret_model.SecretResource{}, store.DeleteByKey(name, "demo"))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			backend.AssertNotExists(&kube_core.Secret{}, ns, name)
		})
	})

	Describe("List()", func() {
		It("should return an empty list if there are no matching resources", func() {
			// given
			secrets := &secret_model.SecretResourceList{}

			// when
			err := s.List(context.Background(), secrets, store.ListByMesh("ignored"))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(secrets.Items).To(HaveLen(0))
		})

		Describe("with resources loaded", func() {
			BeforeEach(func() {
				// given secret in demo mesh
				one := backend.ParseYAML(fmt.Sprintf(`
                apiVersion: v1
                kind: Secret
                type: system.kuma.io/secret
                metadata:
                  namespace: %s
                  name: %s
                  labels:
                    kuma.io/mesh: demo
                data:
                  value: ZXhhbXBsZQ== # base64(example)
`, ns, "one"))
				backend.Create(one)
				// and secret in default mesh
				two := backend.ParseYAML(fmt.Sprintf(`
                apiVersion: v1
                kind: Secret
                type: system.kuma.io/secret
                metadata:
                  namespace: %s
                  name: %s
                  labels:
                    kuma.io/mesh: default
                data:
                  value: YW5vdGhlcg== # base64(another)
`, ns, "two"))
				backend.Create(two)
				// and secret which is not a Kuma secret but resides in the kuma-system namespace - this should be ignored
				three := backend.ParseYAML(fmt.Sprintf(`
                apiVersion: v1
                kind: Secret
                type: some-other-type
                metadata:
                  namespace: %s
                  name: %s
                  labels:
                    kuma.io/mesh: default
                data:
                  value: YW5vdGhlcg== # base64(another)
`, ns, "three"))
				backend.Create(three)
			})

			It("should return a list of secrets in all meshes", func() {
				// given
				secrets := &secret_model.SecretResourceList{}

				// when
				err := s.List(context.Background(), secrets)

				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(secrets.Items).To(HaveLen(2))

				// when
				items := map[string]*secret_model.SecretResource{
					secrets.Items[0].Meta.GetName(): secrets.Items[0],
					secrets.Items[1].Meta.GetName(): secrets.Items[1],
				}
				// then
				Expect(items["one"].Spec.Data.Value).To(Equal([]byte("example")))
				// and
				Expect(items["two"].Spec.Data.Value).To(Equal([]byte("another")))
			})

			It("should return a list of secrets in given mesh", func() {
				// given
				secrets := &secret_model.SecretResourceList{}

				// when
				err := s.List(context.Background(), secrets, store.ListByMesh("default"))

				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(secrets.Items).To(HaveLen(1))
				Expect(secrets.Items[0].Spec.Data.Value).To(Equal([]byte("another")))
			})
		})

	})
})
