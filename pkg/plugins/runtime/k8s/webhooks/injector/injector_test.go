package injector_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/pkg/config"
	conf "github.com/kumahq/kuma/pkg/config/plugins/runtime/k8s"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	inject "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/webhooks/injector"
	"github.com/kumahq/kuma/pkg/test/matchers"
)

var _ = Describe("Injector", func() {

	systemNamespace := "kuma-system"

	caCertPath := filepath.Join(
		"..", "..", "..", "..", "..", "..",
		"test", "certs", "server-cert.pem",
	)

	type testCase struct {
		num       string
		mesh      string
		cfgFile   string
		namespace string
	}

	BeforeAll(func() {
		err := k8sClient.Create(context.Background(), &kube_core.Namespace{ObjectMeta: kube_meta.ObjectMeta{Name: systemNamespace}})
		Expect(err).ToNot(HaveOccurred())

		cPatch := `
apiVersion: kuma.io/v1alpha1
kind: ContainerPatch
metadata:
  namespace: kuma-system
  name: container-patch-1
spec:
  sidecarPatch:
    - op: add
      path: /securityContext/privileged
      value: "false"
    - op: add
      path: "/volumeMounts/-"
      value: '{ "mountPath": "/var/run/secrets/kubernetes.io/serviceaccount/", "name": "{{ template \"kong.serviceAccountTokenName\" . }}", "readOnly": true }'
  initPatch:
    - op: remove
      path: /securityContext/runAsUser`
		decoder := serializer.NewCodecFactory(k8sClientScheme).UniversalDeserializer()
		pobj, _, errCPatch := decoder.Decode([]byte(cPatch), nil, nil)
		Expect(errCPatch).ToNot(HaveOccurred())
		errCPatchCreate := k8sClient.Create(context.Background(), pobj.(kube_client.Object))
		Expect(errCPatchCreate).ToNot(HaveOccurred())
	})

	AfterAll(func() {
		err := k8sClient.DeleteAllOf(context.Background(), &v1alpha1.ContainerPatch{}, kube_client.InNamespace("default"))
		Expect(err).ToNot(HaveOccurred())
	})

	BeforeEach(func() {
		err := k8sClient.DeleteAllOf(context.Background(), &v1alpha1.Mesh{})
		Expect(err).ToNot(HaveOccurred())
	})

	DescribeTable("should inject Kuma into a Pod",
		func(given testCase) {
			// setup
			inputFile := filepath.Join("testdata", fmt.Sprintf("inject.%s.input.yaml", given.num))
			goldenFile := filepath.Join("testdata", fmt.Sprintf("inject.%s.golden.yaml", given.num))

			var cfg conf.Injector
			Expect(config.Load(filepath.Join("testdata", given.cfgFile), &cfg)).To(Succeed())
			cfg.CaCertFile = caCertPath
			injector, err := inject.New(cfg, "http://kuma-control-plane.kuma-system:5681", k8sClient, k8s.NewSimpleConverter(), 9901, systemNamespace)
			Expect(err).ToNot(HaveOccurred())

			// and create mesh
			decoder := serializer.NewCodecFactory(k8sClientScheme).UniversalDeserializer()
			obj, _, errMesh := decoder.Decode([]byte(given.mesh), nil, nil)
			Expect(errMesh).ToNot(HaveOccurred())
			errCreate := k8sClient.Create(context.Background(), obj.(kube_client.Object))
			Expect(errCreate).ToNot(HaveOccurred())
			ns, _, errNs := decoder.Decode([]byte(given.namespace), nil, nil)
			Expect(errNs).ToNot(HaveOccurred())
			errUpd := k8sClient.Update(context.Background(), ns.(kube_client.Object))
			Expect(errUpd).ToNot(HaveOccurred())

			// given
			pod := &kube_core.Pod{}

			By("loading input Pod")
			// when
			input, err := os.ReadFile(inputFile)
			// then
			Expect(err).ToNot(HaveOccurred())
			// when
			err = yaml.Unmarshal(input, pod)
			// then
			Expect(err).ToNot(HaveOccurred())

			By("injecting Kuma")
			// when
			err = injector.InjectKuma(context.Background(), pod)
			// then
			Expect(err).ToNot(HaveOccurred())

			By("loading golden Pod")
			// when
			actual, err := yaml.Marshal(pod)
			// then
			Expect(err).ToNot(HaveOccurred())

			By("comparing actual against golden")
			Expect(actual).To(matchers.MatchGoldenYAML(goldenFile))
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
			cfgFile: "inject.config.yaml",
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
			cfgFile: "inject.config.yaml",
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
			cfgFile: "inject.config.yaml",
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
			cfgFile: "inject.config.yaml",
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
			cfgFile: "inject.config.yaml",
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
			cfgFile: "inject.config.yaml",
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
                  kuma.io/sidecar-injection: enabled`,
			cfgFile: "inject.config.yaml",
		}),
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
			cfgFile: "inject.config.yaml",
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
			cfgFile: "inject.config.yaml",
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
			cfgFile: "inject.config.yaml",
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
			cfgFile: "inject.config.yaml",
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
			cfgFile: "inject.config.yaml",
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
			cfgFile: "inject.config.yaml",
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
			cfgFile: "inject.config.yaml",
		}),
		Entry("15. virtual probes: config - 9000, pod - 19000", testCase{
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
			cfgFile: "inject.config.yaml",
		}),
		Entry("16. virtual probes: config - enabled, pod - disabled", testCase{
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
			cfgFile: "inject.config.yaml",
		}),
		Entry("17. traffic.kuma.io/exclude-inbound-ports and traffic.kuma.io/exclude-outbound-ports", testCase{
			num: "17",
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
			cfgFile: "inject.config.yaml",
		}),
		Entry("18. traffic.kuma.io/exclude-inbound-ports and traffic.kuma.io/exclude-outbound-ports from config", testCase{
			num: "18",
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
			cfgFile: "inject.config-ports.yaml",
		}),
		Entry("19. traffic.kuma.io/exclude-inbound-ports and traffic.kuma.io/exclude-outbound-ports overrides config", testCase{
			num: "19",
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
			cfgFile: "inject.config-ports.yaml",
		}),
		Entry("20. skip injection for label exception", testCase{
			num: "20",
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
			cfgFile: "inject.config.yaml",
		}),
		Entry("21. virtual probes: config - disabled, pod - empty", testCase{
			num: "21",
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
			cfgFile: "inject.vp-disabled.config.yaml",
		}),
		Entry("22. virtual probes: config - disabled, pod - enabled", testCase{
			num: "22",
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
			cfgFile: "inject.vp-disabled.config.yaml",
		}),
		Entry("23. Adjust Pod's probes, named port", testCase{
			num: "23",
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
			cfgFile: "inject.config.yaml",
		}),
		Entry("24. sidecar env var config overrides", testCase{
			num: "24",
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
			cfgFile: "inject.env-vars.config.yaml",
		}),
		Entry("25. sidecar with builtinDNS", testCase{
			num: "25",
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
			cfgFile: "inject.builtindns.config.yaml",
		}),
		Entry("26. sidecar with high concurrency", testCase{
			num: "26",
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
			cfgFile: "inject.builtindns.config.yaml",
		}),
		Entry("27. sidecar with high resource limit", testCase{
			num: "27",
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
			cfgFile: "inject.high-resources.config.yaml",
		}),
		Entry("28. sidecar with specified service account token volume", testCase{
			num: "28",
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
			cfgFile: "inject.config.yaml",
		}),
		Entry("29. sidecar with specified drain time", testCase{
			num: "29",
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
			cfgFile: "inject.config.yaml",
		}),
		Entry("30. sidecar with patch", testCase{
			num: "30",
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
			cfgFile: "inject.config.yaml",
		}),
		Entry("31. port override #4458", testCase{
			num: "31",
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
			cfgFile: "inject.builtindns.config.yaml",
		}),
	)

	Describe("should fail", func() {
		It("when pod is annotated with the name of non-existing patch", func() {
			// setup
			ctx := context.Background()
			configFile := "testdata/inject.config.yaml"
			inputFile := "testdata/inject.shouldfail.1.input.yaml"

			var cfg conf.Injector
			Expect(config.Load(configFile, &cfg)).To(Succeed())

			cfg.CaCertFile = caCertPath

			injector, err := inject.New(
				cfg,
				"http://kuma-control-plane.kuma-system:5681",
				k8sClient,
				k8s.NewSimpleConverter(),
				9901,
				systemNamespace,
			)
			Expect(err).To(Succeed())

			mesh := []byte(`
              apiVersion: kuma.io/v1alpha1
              kind: Mesh
              metadata:
                name: default`)

			namespace := []byte(`
              apiVersion: v1
              kind: Namespace
              metadata:
                name: default
                annotations:
                  kuma.io/sidecar-injection: enabled`)

			// and create mesh
			decoder := serializer.
				NewCodecFactory(k8sClientScheme).
				UniversalDeserializer()

			obj, _, errMesh := decoder.Decode(mesh, nil, nil)
			Expect(errMesh).To(Succeed())

			Expect(k8sClient.Create(ctx, obj.(kube_client.Object))).To(Succeed())

			ns, _, errNs := decoder.Decode(namespace, nil, nil)
			Expect(errNs).To(Succeed())

			Expect(k8sClient.Update(ctx, ns.(kube_client.Object))).To(Succeed())

			// given
			pod := &kube_core.Pod{}

			By("loading input Pod")
			// when
			input, err := os.ReadFile(inputFile)
			// then
			Expect(err).To(Succeed())
			// when
			err = yaml.Unmarshal(input, pod)
			// then
			Expect(err).To(Succeed())

			By("injecting Kuma")
			// when
			err = injector.InjectKuma(ctx, pod)
			// then
			Expect(err).ToNot(Succeed())
		})
	})
}, Ordered)
