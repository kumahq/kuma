package k8s_test

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	config_store "github.com/Kong/kuma/pkg/core/config/store"
	"github.com/Kong/kuma/pkg/core/resources/store"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/uuid"
	"sigs.k8s.io/controller-runtime/pkg/client"

	system_proto "github.com/Kong/kuma/api/system/v1alpha1"
	system_model "github.com/Kong/kuma/pkg/core/resources/apis/system"
	"github.com/Kong/kuma/pkg/plugins/config/k8s"
)

var _ = Describe("KubernetesStore", func() {

	var s config_store.ConfigStore
	var ns string

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

		s, err = k8s.NewStore(k8sClient, ns)
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("Create()", func() {
		It("should create a new config", func() {
			// given
			config := &system_model.ConfigResource{
				Spec: system_proto.Config{
					Config: "test",
				},
			}
			expected := backend.ParseYAML(`
            apiVersion: v1
            kind: ConfigMap
            metadata:
              name: "kuma-internal-config"
              namespace : "kuma-system"
            data:
              config: "test" 
`).(*kube_core.ConfigMap)

			// when
			err := s.Create(context.Background(), config, store.CreateByKey("kuma-internal-config", ""))

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actual := kube_core.ConfigMap{}
			backend.Get(&actual, ns, "kuma-internal-config")

			// then
			Expect(actual.Data).To(Equal(expected.Data))
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
              namespace : %s
            data:
              config: "next test" 
    `, ns)).(*kube_core.ConfigMap)

			// given
			config := &system_model.ConfigResource{}

			// when
			err := s.Get(context.Background(), config, store.GetByKey("kuma-internal-config", ""))
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
		})
	})

	Describe("Get()", func() {
		It("should return an error if resource is not found", func() {
			// setup
			backend.AssertNotExists(&kube_core.ConfigMap{}, ns, "kuma-internal-config")

			// when
			err := s.Get(context.Background(), &system_model.ConfigResource{}, store.GetByKey("kuma-internal-config", ""))

			// then
			Expect(err).To(HaveOccurred())
		})
	})

})
