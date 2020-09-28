package injector_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config"
	conf "github.com/kumahq/kuma/pkg/config/plugins/runtime/k8s"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	inject "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/webhooks/injector"

	kube_core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	"github.com/ghodss/yaml"
)

var _ = Describe("Injector", func() {

	var injector *inject.KumaInjector

	BeforeEach(func() {
		var cfg conf.Injector
		Expect(config.Load(filepath.Join("testdata", "inject.config.yaml"), &cfg)).To(Succeed())
		injector = inject.New(cfg, "http://kuma-control-plane.kuma-system:5681", k8sClient)
	})

	type testCase struct {
		num       string
		mesh      string
		namespace string
	}

	BeforeEach(func() {
		err := k8sClient.DeleteAllOf(context.Background(), &v1alpha1.Mesh{})
		Expect(err).ToNot(HaveOccurred())
	})

	DescribeTable("should inject Kuma into a Pod",
		func(given testCase) {
			// setup
			inputFile := filepath.Join("testdata", fmt.Sprintf("inject.%s.input.yaml", given.num))
			goldenFile := filepath.Join("testdata", fmt.Sprintf("inject.%s.golden.yaml", given.num))

			// and create mesh
			decoder := serializer.NewCodecFactory(k8sClientScheme).UniversalDeserializer()
			obj, _, errMesh := decoder.Decode([]byte(given.mesh), nil, nil)
			Expect(errMesh).ToNot(HaveOccurred())
			errCreate := k8sClient.Create(context.Background(), obj)
			Expect(errCreate).ToNot(HaveOccurred())
			ns, _, errNs := decoder.Decode([]byte(given.namespace), nil, nil)
			Expect(errNs).ToNot(HaveOccurred())
			errUpd := k8sClient.Update(context.Background(), ns)
			Expect(errUpd).ToNot(HaveOccurred())

			// given
			pod := &kube_core.Pod{}

			By("loading input Pod")
			// when
			input, err := ioutil.ReadFile(inputFile)
			// then
			Expect(err).ToNot(HaveOccurred())
			// when
			err = yaml.Unmarshal(input, pod)
			// then
			Expect(err).ToNot(HaveOccurred())

			By("injecting Kuma")
			// when
			err = injector.InjectKuma(pod)
			// then
			Expect(err).ToNot(HaveOccurred())

			By("loading golden Pod")
			// when
			actual, err := yaml.Marshal(pod)
			// then
			Expect(err).ToNot(HaveOccurred())

			By("comparing actual against golden")
			// when
			expected, err := ioutil.ReadFile(goldenFile)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(expected))
		},
		Entry("01. Pod without init containers and annotations", testCase{
			num: "01",
			mesh: `
              apiVersion: kuma.io/v1alpha1
              kind: Mesh
              metadata:
                name: default`,
			namespace: `
              apiVersion: v1
              kind: Namespace
              metadata:
                name: default
                annotations:
                  kuma.io/sidecar-injection: enabled`,
		}),
		Entry("02. Pod with init containers and annotations", testCase{
			num: "02",
			mesh: `
              apiVersion: kuma.io/v1alpha1
              kind: Mesh
              metadata:
                name: default`,
			namespace: `
              apiVersion: v1
              kind: Namespace
              metadata:
                name: default
                annotations:
                  kuma.io/sidecar-injection: enabled`,
		}),
		Entry("03. Pod without Namespace and Name", testCase{
			num: "03",
			mesh: `
              apiVersion: kuma.io/v1alpha1
              kind: Mesh
              metadata:
                name: default`,
			namespace: `
              apiVersion: v1
              kind: Namespace
              metadata:
                name: default
                annotations:
                  kuma.io/sidecar-injection: enabled`,
		}),
		Entry("04. Pod with explicitly selected Mesh", testCase{
			num: "04",
			mesh: `
              apiVersion: kuma.io/v1alpha1
              kind: Mesh
              metadata:
                name: demo`,
			namespace: `
              apiVersion: v1
              kind: Namespace
              metadata:
                name: default
                annotations:
                  kuma.io/sidecar-injection: enabled`,
		}),
		Entry("05. Pod without ServiceAccount token", testCase{
			num: "05",
			mesh: `
              apiVersion: kuma.io/v1alpha1
              kind: Mesh
              metadata:
                name: default`,
			namespace: `
              apiVersion: v1
              kind: Namespace
              metadata:
                name: default
                annotations:
                  kuma.io/sidecar-injection: enabled`,
		}),
		Entry("06. Pod with kuma.io/gateway annotation", testCase{
			num: "06",
			mesh: `
              apiVersion: kuma.io/v1alpha1
              kind: Mesh
              metadata:
                name: default`,
			namespace: `
              apiVersion: v1
              kind: Namespace
              metadata:
                name: default
                annotations:
                  kuma.io/sidecar-injection: enabled`,
		}),
		Entry("07. Pod with mesh with metrics enabled", testCase{
			num: "07",
			mesh: `
              apiVersion: kuma.io/v1alpha1
              kind: Mesh
              metadata:
                name: default
              spec:
                metrics:
                  prometheus:
                    port: 1234
                    path: /metrics`,
			namespace: `
              apiVersion: v1
              kind: Namespace
              metadata:
                name: default
                annotations:
                  kuma.io/sidecar-injection: enabled`}),
		Entry("08. Pod with prometheus annotation already defined so injector won't override those", testCase{
			num: "08",
			mesh: `
              apiVersion: kuma.io/v1alpha1
              kind: Mesh
              metadata:
                name: default
              spec:
                metrics:
                  prometheus:
                    port: 1234
                    path: /metrics`,
			namespace: `
              apiVersion: v1
              kind: Namespace
              metadata:
                name: default
                annotations:
                  kuma.io/sidecar-injection: enabled`,
		}),
		Entry("09. Pod with Kuma metrics annotation overrides", testCase{
			num: "09",
			mesh: `
              apiVersion: kuma.io/v1alpha1
              kind: Mesh
              metadata:
                name: default
              spec:
                metrics:
                  prometheus:
                    port: 1234
                    path: /metrics`,
			namespace: `
              apiVersion: v1
              kind: Namespace
              metadata:
                name: default
                annotations:
                  kuma.io/sidecar-injection: enabled`,
		}),
		Entry("10. Pod with `kuma.io/sidecar-injection: disabled` annotation", testCase{
			num: "10",
			mesh: `
              apiVersion: kuma.io/v1alpha1
              kind: Mesh
              metadata:
                name: default
              spec: {}`,
			namespace: `
              apiVersion: v1
              kind: Namespace
              metadata:
                name: default
                annotations:
                  kuma.io/sidecar-injection: enabled`,
		}),
		Entry("11. Namespace - `kuma.io/sidecar-injection: disabled`, Pod - `kuma.io/sidecar-injection: enabled`", testCase{
			num: "11",
			mesh: `
              apiVersion: kuma.io/v1alpha1
              kind: Mesh
              metadata:
                name: default
              spec: {}`,
			namespace: `
              apiVersion: v1
              kind: Namespace
              metadata:
                name: default
                annotations:
                  kuma.io/sidecar-injection: disabled`,
		}),
		Entry("12. Mesh name from Namespace", testCase{
			num: "12",
			mesh: `
              apiVersion: kuma.io/v1alpha1
              kind: Mesh
              metadata:
                name: mesh-name-from-ns
              spec: {}`,
			namespace: `
              apiVersion: v1
              kind: Namespace
              metadata:
                name: default
                annotations:
                  kuma.io/sidecar-injection: enabled
                  kuma.io/mesh: mesh-name-from-ns`,
		}),
		Entry("13. Override mesh name in Pod", testCase{
			num: "13",
			mesh: `
              apiVersion: kuma.io/v1alpha1
              kind: Mesh
              metadata:
                name: mesh-name-from-pod
              spec: {}`,
			namespace: `
              apiVersion: v1
              kind: Namespace
              metadata:
                name: default
                annotations:
                  kuma.io/sidecar-injection: enabled
                  kuma.io/mesh: mesh-name-from-ns`,
		}),
		Entry("14. Adjust Pod's probes", testCase{
			num: "14",
			mesh: `
              apiVersion: kuma.io/v1alpha1
              kind: Mesh
              metadata:
                name: default
              spec: {}`,
			namespace: `
              apiVersion: v1
              kind: Namespace
              metadata:
                name: default
                annotations:
                  kuma.io/sidecar-injection: enabled`,
		}),
		Entry("15. Override virtual probes port", testCase{
			num: "15",
			mesh: `
              apiVersion: kuma.io/v1alpha1
              kind: Mesh
              metadata:
                name: default
              spec: {}`,
			namespace: `
              apiVersion: v1
              kind: Namespace
              metadata:
                name: default
                annotations:
                  kuma.io/sidecar-injection: enabled`,
		}),
		Entry("16. Override virtual probes enabled state", testCase{
			num: "16",
			mesh: `
              apiVersion: kuma.io/v1alpha1
              kind: Mesh
              metadata:
                name: default
              spec: {}`,
			namespace: `
              apiVersion: v1
              kind: Namespace
              metadata:
                name: default
                annotations:
                  kuma.io/sidecar-injection: enabled`,
		}),
	)
})
