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
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	system_model "github.com/kumahq/kuma/pkg/core/resources/apis/system"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/config/k8s"
	k8s_resources "github.com/kumahq/kuma/pkg/plugins/resources/k8s"
)

var _ = Describe("KubernetesStore", func() {
	var s core_store.ResourceStore
	var ns string

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

		s, err = k8s.NewStore(k8sClient, ns, k8sClientScheme, k8s_resources.NewSimpleConverter())
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("Create()", func() {
		It("should create a new config", func() {
			// given
			config := &system_model.ConfigResource{
				Spec: &system_proto.Config{
					Config: "test",
				},
			}
			expected := backend.ParseYAML(`
            apiVersion: v1
            kind: ConfigMap
            metadata:
              name: "kuma-internal-config"
              namespace: "kuma-system"
            data:
              config: "test" 
`).(*kube_core.ConfigMap)

			// when
			err := s.Create(context.Background(), config, core_store.CreateByKey("kuma-internal-config", ""), core_store.CreateWithLabels(map[string]string{
				mesh_proto.DisplayName: "kuma-internal-config",
			}))

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actual := kube_core.ConfigMap{}
			backend.Get(&actual, ns, "kuma-internal-config")

			// then
			Expect(actual.Data).To(Equal(expected.Data))
			Expect(actual.GetLabels()).NotTo(HaveKey(mesh_proto.DisplayName))
			Expect(actual.GetAnnotations()).To(HaveKeyWithValue(mesh_proto.DisplayName, "kuma-internal-config"))
		})
	})

	Describe("Update()", func() {
		It("should update an existing resource", func() {
			// setup
			initial := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: v1
            kind: ConfigMap
            metadata:
              name: "kuma-internal-config"
              namespace : %s
              annotations:
                kuma.io/display-name: "kuma-internal-config"
            data:
              config: "test" 
    `, ns))
			backend.Create(initial)
			// and
			expected := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: v1
            kind: ConfigMap
            metadata:
              name: "kuma-internal-config"
              namespace: %s
              annotations:
                kuma.io/display-name: "kuma-internal-config"
            data:
              config: "next test" 
    `, ns)).(*kube_core.ConfigMap)

			// given
			config := system_model.NewConfigResource()

			// when
			err := s.Get(context.Background(), config, core_store.GetByKey("kuma-internal-config", ""))
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			config.Spec.Config = "next test"
			err = s.Update(context.Background(), config)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actual := kube_core.ConfigMap{}
			backend.Get(&actual, ns, "kuma-internal-config")

			// then
			Expect(actual.Data).To(Equal(expected.Data))
			Expect(actual.GetLabels()).NotTo(HaveKey(mesh_proto.DisplayName))
			Expect(actual.GetAnnotations()).To(HaveKeyWithValue(mesh_proto.DisplayName, "kuma-internal-config"))
		})

		It("should return error in case of resource conflict", func() {
			// setup
			initial := backend.ParseYAML(fmt.Sprintf(`
            apiVersion: v1
            kind: ConfigMap
            metadata:
              name: "kuma-internal-config"
              namespace : %s
            data:
              config: "test" 
    `, ns))
			backend.Create(initial)

			// given
			config := system_model.NewConfigResource()

			err := s.Get(context.Background(), config, core_store.GetByKey("kuma-internal-config", ""))
			Expect(err).ToNot(HaveOccurred())

			config.Meta.(*k8s.KubernetesMetaAdapter).ResourceVersion = config.Meta.(*k8s.KubernetesMetaAdapter).ResourceVersion + "1"
			config.Spec.Config = "next test"

			// when
			err = s.Update(context.Background(), config)

			// then
			Expect(err).To(MatchError(core_store.ErrorResourceConflict(system_model.ConfigType, "kuma-internal-config", "")))
		})
	})

	Describe("Get()", func() {
		It("should return an error if resource is not found", func() {
			// setup
			backend.AssertNotExists(&kube_core.ConfigMap{}, ns, "kuma-internal-config")

			// when
			err := s.Get(context.Background(), system_model.NewConfigResource(), core_store.GetByKey("kuma-internal-config", ""))

			// then
			Expect(err).To(HaveOccurred())
		})
	})
})
