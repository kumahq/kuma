package k8s_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_system "github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	secret_store "github.com/kumahq/kuma/pkg/core/secrets/store"
	"github.com/kumahq/kuma/pkg/plugins/secrets/k8s"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("KubernetesStore", func() {
	var s secret_store.SecretStore
	var ns string // each test should run in a dedicated k8s namespace
	const name = "demo"
	const noMesh = ""

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
		ns = core.NewUUID()

		err := k8sClient.Create(context.Background(), &kube_core.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: ns},
		})
		Expect(err).ToNot(HaveOccurred())

		// we need to bring in the actual scheme we're using so that the Mesh CRD can be hooked up as owner,
		// otherwise we will get "no kind is registered for the type v1alpha1.Mesh in scheme"
		s, err = k8s.NewStore(k8sClient, k8sClient, runtime.NewScheme(), ns)
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("Create()", func() {
		It("should create a new secret", func() {
			// given
			secret := &core_system.SecretResource{
				Spec: &system_proto.Secret{
					Data: util_proto.Bytes([]byte("example")),
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
<<<<<<< HEAD
			err := s.Create(context.Background(), secret, store.CreateByKey(name, "demo"))
=======
			err := rs.Create(context.Background(), secret, store.CreateByKey(name, "demo"), store.CreateWithLabels(map[string]string{
				mesh_proto.DisplayName: name,
			}))
>>>>>>> e1179afd8 (fix(k8s): set annotation kuma.io/display-name for Secrets and Configs (#11923))

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
			Expect(actual.GetObjectMeta().GetLabels()).NotTo(HaveKey(mesh_proto.DisplayName))
			Expect(actual.GetObjectMeta().GetAnnotations()).To(HaveKeyWithValue(mesh_proto.DisplayName, name))

			// and
			Expect(actual.ObjectMeta.ResourceVersion).To(Equal(secret.Meta.GetVersion()))
		})

		It("should create a new global secret", func() {
			// given
			secret := &core_system.GlobalSecretResource{
				Spec: &system_proto.Secret{
					Data: util_proto.Bytes([]byte("example")),
				},
			}
			expected := backend.ParseYAML(`
            apiVersion: v1
            kind: Secret
            type: system.kuma.io/global-secret
            data:
              value: ZXhhbXBsZQ== # base64(example)
`).(*kube_core.Secret)

			// when
			err := s.Create(context.Background(), secret, store.CreateByKey(name, noMesh))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(secret.Meta.GetName()).To(Equal(name))
			Expect(secret.Meta.GetMesh()).To(Equal(noMesh))
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
			err := s.Create(context.Background(), core_system.NewSecretResource(), store.CreateByKey(name, noMesh))

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			err = s.Create(context.Background(), core_system.NewSecretResource(), store.CreateByKey(name, noMesh))

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
              annotations:
                kuma.io/display-name: %s
            data:
              value: ZXhhbXBsZQ== # base64(example)
`, ns, name, name))
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
			secret := core_system.NewSecretResource()

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
			Expect(actual.GetObjectMeta().GetLabels()).NotTo(HaveKey(mesh_proto.DisplayName))
			Expect(actual.GetObjectMeta().GetAnnotations()).To(HaveKeyWithValue(mesh_proto.DisplayName, name))
			// and
			Expect(actual.ObjectMeta.ResourceVersion).To(Equal(secret.Meta.GetVersion()))
		})

		It("should update an existing global secret", func() {
			// setup
			initial := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: v1
            kind: Secret
            type: system.kuma.io/global-secret
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
            type: system.kuma.io/global-secret
            data:
              value: YW5vdGhlcg== # base64(another)
`).(*kube_core.Secret)

			// given
			secret := core_system.NewGlobalSecretResource()

			// when
			err := s.Get(context.Background(), secret, store.GetByKey(name, noMesh))
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
			secret := core_system.NewSecretResource()

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
			secret1 := core_system.NewSecretResource()

			// when
			err := s.Get(context.Background(), secret1, store.GetByKey(name, "demo"))
			// then
			Expect(err).ToNot(HaveOccurred())

			// given
			secret2 := core_system.NewSecretResource()

			// when
			err = s.Get(context.Background(), secret2, store.GetByKey(name, "demo"))
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			secret1.Spec = &system_proto.Secret{
				Data: util_proto.Bytes([]byte("example")),
			}
			err = s.Update(context.Background(), secret1)
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			secret2.Spec = &system_proto.Secret{
				Data: util_proto.Bytes([]byte("another")),
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
			err := s.Get(context.Background(), core_system.NewSecretResource(), store.GetByKey(name, "demo"))

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
			actual := core_system.NewSecretResource()

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

		It("should return an existing global secret", func() {
			// setup
			expected := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: v1
            kind: Secret
            type: system.kuma.io/global-secret
            metadata:
              namespace: %s
              name: %s
            data:
              value: ZXhhbXBsZQ== # base64(example)
`, ns, name))
			backend.Create(expected)

			// given
			actual := core_system.NewGlobalSecretResource()

			// when
			err := s.Get(context.Background(), actual, store.GetByKey(name, noMesh))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual.Meta.GetName()).To(Equal(name))
			Expect(actual.Meta.GetMesh()).To(Equal(noMesh))
			// and
			Expect(actual.Spec.Data.Value).To(Equal([]byte("example")))
		})

		It("should not return global secret as regular secret", func() {
			// setup
			expected := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: v1
            kind: Secret
            type: system.kuma.io/global-secret
            metadata:
              namespace: %s
              name: %s
            data:
              value: ZXhhbXBsZQ== # base64(example)
`, ns, name))
			backend.Create(expected)

			// given
			actual := core_system.NewSecretResource()

			// when
			err := s.Get(context.Background(), actual, store.GetByKey(name, "default"))

			// then
			Expect(err).To(MatchError(`Resource not found: type="Secret" name="demo" mesh="default"`))

			// when
			err = s.Get(context.Background(), actual, store.GetByKey(name, noMesh))

			// then
			Expect(err).To(MatchError(`Resource not found: type="Secret" name="demo" mesh=""`))
		})

		It("should not return regular secret as global secret", func() {
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
			actual := core_system.NewGlobalSecretResource()

			// when
			err := s.Get(context.Background(), actual, store.GetByKey(name, "default"))

			// then
			Expect(err).To(MatchError(`Resource not found: type="GlobalSecret" name="demo" mesh="default"`))

			// when
			err = s.Get(context.Background(), actual, store.GetByKey(name, noMesh))

			// then
			Expect(err).To(MatchError(`Resource not found: type="GlobalSecret" name="demo" mesh=""`))
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
			actual := core_system.NewSecretResource()

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
			actual := core_system.NewSecretResource()

			// when
			err := s.Get(context.Background(), actual, store.GetByKey(name, "another-mesh"))

			// then
			Expect(err).To(MatchError(store.ErrorResourceNotFound(core_system.SecretType, name, "another-mesh")))
			Expect(actual.Spec.GetData().GetValue()).To(BeEmpty())
		})
	})

	Describe("Delete()", func() {
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
			err := s.Delete(context.Background(), core_system.NewSecretResource(), store.DeleteByKey(name, "demo"))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			backend.AssertNotExists(&kube_core.Secret{}, ns, name)
		})

		It("should delete an existing global secret", func() {
			// setup
			initial := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: v1
            kind: Secret
            type: system.kuma.io/global-secret
            metadata:
              namespace: %s
              name: %s
`, ns, name))
			backend.Create(initial)

			// when
			err := s.Delete(context.Background(), core_system.NewGlobalSecretResource(), store.DeleteByKey(name, noMesh))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			backend.AssertNotExists(&kube_core.Secret{}, ns, name)
		})

		It("should no delete an existing global secret if SecretResource is used", func() {
			// setup
			initial := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: v1
            kind: Secret
            type: system.kuma.io/global-secret
            metadata:
              namespace: %s
              name: %s
`, ns, name))
			backend.Create(initial)

			// when
			err := s.Delete(context.Background(), core_system.NewSecretResource(), store.DeleteByKey(name, noMesh))

			// then
			Expect(err).To(MatchError(`failed to delete k8s secret: Resource not found: type="Secret" name="demo" mesh=""`))
		})

		It("should no delete an existing mesh scoped secret if GlobalSecretResource", func() {
			// setup
			initial := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: v1
            kind: Secret
            type: system.kuma.io/secret
            metadata:
              namespace: %s
              name: %s
`, ns, name))
			backend.Create(initial)

			// when
			err := s.Delete(context.Background(), core_system.NewGlobalSecretResource(), store.DeleteByKey(name, "default"))

			// then
			Expect(err).To(MatchError(`failed to delete k8s secret: Resource not found: type="GlobalSecret" name="demo" mesh="default"`))
		})
	})

	Describe("List()", func() {
		It("should return an empty list if there are no matching resources", func() {
			// given
			secrets := &core_system.SecretResourceList{}

			// when
			err := s.List(context.Background(), secrets, store.ListByMesh("ignored"))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(secrets.Items).To(BeEmpty())
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

				four := backend.ParseYAML(fmt.Sprintf(`
                apiVersion: v1
                kind: Secret
                type: system.kuma.io/global-secret
                metadata:
                  namespace: %s
                  name: %s
                data:
                  value: Zm91cg== # base64(four)
`, ns, "four"))
				backend.Create(four)
			})

			It("should return a list of secrets in all meshes", func() {
				// given
				secrets := &core_system.SecretResourceList{}

				// when
				err := s.List(context.Background(), secrets)

				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(secrets.Items).To(HaveLen(2))

				// when
				items := map[string]*core_system.SecretResource{
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
				secrets := &core_system.SecretResourceList{}

				// when
				err := s.List(context.Background(), secrets, store.ListByMesh("default"))

				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(secrets.Items).To(HaveLen(1))
				Expect(secrets.Items[0].Spec.Data.Value).To(Equal([]byte("another")))
			})

			It("should return a list of global secrets", func() {
				// given
				secrets := &core_system.GlobalSecretResourceList{}

				// when
				err := s.List(context.Background(), secrets)

				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(secrets.Items).To(HaveLen(1))
				Expect(string(secrets.Items[0].Spec.Data.Value)).To(Equal("four"))
			})
		})
	})
})
