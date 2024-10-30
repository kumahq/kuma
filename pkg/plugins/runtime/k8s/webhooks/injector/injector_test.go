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
	k8s_util "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
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

	DescribeTableSubtree("should inject Kuma into a Pod",
		func(given testCase) {
			// setup
			inputFile := filepath.Join("testdata", fmt.Sprintf("inject.%s.input.yaml", given.num))

			run := func(sidecarsEnabled bool) {
				var goldenFile string
				if sidecarsEnabled {
					goldenFile = filepath.Join("testdata", fmt.Sprintf("inject.sidecar-feature.%s.golden.yaml", given.num))
				} else {
					goldenFile = filepath.Join("testdata", fmt.Sprintf("inject.%s.golden.yaml", given.num))
				}

				var cfg conf.Injector
				Expect(config.Load(filepath.Join("testdata", given.cfgFile), &cfg)).To(Succeed())
				cfg.CaCertFile = caCertPath
				injector, err := inject.New(cfg, "http://kuma-control-plane.kuma-system:5681", k8sClient, sidecarsEnabled, k8s.NewSimpleConverter(), 9901, systemNamespace)
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
				if !sidecarsEnabled {
					Expect(pod.Spec.Containers[0].Name).To(BeEquivalentTo(k8s_util.KumaSidecarContainerName))
				} else {
					Expect(pod.Spec.InitContainers).To(ContainElement(
						WithTransform(func(c kube_core.Container) string { return c.Name }, Equal(k8s_util.KumaSidecarContainerName))),
					)
				}

				By("loading golden Pod")
				// when
				actual, err := yaml.Marshal(pod)
				// then
				Expect(err).ToNot(HaveOccurred())

				By("comparing actual against golden")
				Expect(actual).To(matchers.MatchGoldenYAML(goldenFile))
			}
			It("injects as traditional sidecar container", func() {
				run(false)
			})
			It("injects with sidecar containers feature", func() {
				run(true)
			})
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
                labels:
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
                labels:
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
                labels:
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
                labels:
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
                labels:
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
                labels:
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
                labels:
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
                labels:
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
                labels:
                  kuma.io/sidecar-injection: enabled`,
			cfgFile: "inject.config.yaml",
		}),
		Entry("10. Namespace - `kuma.io/sidecar-injection: disabled`, Pod - `kuma.io/sidecar-injection: enabled`", testCase{
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
                labels:
                  kuma.io/sidecar-injection: disabled`,
			cfgFile: "inject.config.yaml",
		}),
		Entry("11. Mesh name from Namespace", testCase{
			num: "11",
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
                labels:
                  kuma.io/sidecar-injection: enabled
                annotations:
                  kuma.io/mesh: mesh-name-from-ns`,
			cfgFile: "inject.config.yaml",
		}),
		Entry("12. Override mesh name in Pod", testCase{
			num: "12",
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
                labels:
                  kuma.io/sidecar-injection: enabled
                annotations:
                  kuma.io/mesh: mesh-name-from-ns`,
			cfgFile: "inject.config.yaml",
		}),
		Entry("13. Adjust Pod's probes", testCase{
			num: "13",
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
                labels:
                  kuma.io/sidecar-injection: enabled`,
			cfgFile: "inject.config.yaml",
		}),
		Entry("14. virtual probes: config - 9000, pod - 19000", testCase{
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
                labels:
                  kuma.io/sidecar-injection: enabled`,
			cfgFile: "inject.config.yaml",
		}),
		Entry("15. virtual probes: config - enabled, pod - disabled", testCase{
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
                labels:
                  kuma.io/sidecar-injection: enabled`,
			cfgFile: "inject.config.yaml",
		}),
		Entry("16. traffic.kuma.io/exclude-inbound-ports and traffic.kuma.io/exclude-outbound-ports", testCase{
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
                labels:
                  kuma.io/sidecar-injection: enabled`,
			cfgFile: "inject.config.yaml",
		}),
		Entry("17. traffic.kuma.io/exclude-inbound-ports and traffic.kuma.io/exclude-outbound-ports from config", testCase{
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
                labels:
                  kuma.io/sidecar-injection: enabled`,
			cfgFile: "inject.config-ports.yaml",
		}),
		Entry("18. traffic.kuma.io/exclude-inbound-ports and traffic.kuma.io/exclude-outbound-ports overrides config", testCase{
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
                labels:
                  kuma.io/sidecar-injection: enabled`,
			cfgFile: "inject.config-ports.yaml",
		}),
		Entry("19. virtual probes: config - disabled, pod - empty", testCase{
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
                labels:
                  kuma.io/sidecar-injection: enabled`,
			cfgFile: "inject.vp-disabled.config.yaml",
		}),
		Entry("20. virtual probes: config - disabled, pod - enabled", testCase{
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
                labels:
                  kuma.io/sidecar-injection: enabled`,
			cfgFile: "inject.vp-disabled.config.yaml",
		}),
		Entry("21. Adjust Pod's probes, named port", testCase{
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
                labels:
                  kuma.io/sidecar-injection: enabled`,
			cfgFile: "inject.config.yaml",
		}),
		Entry("22. sidecar env var config overrides", testCase{
			num: "22",
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
                labels:
                  kuma.io/sidecar-injection: enabled`,
			cfgFile: "inject.env-vars.config.yaml",
		}),
		Entry("23. sidecar with builtinDNS", testCase{
			num: "23",
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
                labels:
                  kuma.io/sidecar-injection: enabled`,
			cfgFile: "inject.builtindns.config.yaml",
		}),
		Entry("24. sidecar with high concurrency", testCase{
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
                labels:
                  kuma.io/sidecar-injection: enabled`,
			cfgFile: "inject.builtindns.config.yaml",
		}),
		Entry("25. sidecar with high resource limit", testCase{
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
                labels:
                  kuma.io/sidecar-injection: enabled`,
			cfgFile: "inject.high-resources.config.yaml",
		}),
		Entry("26. sidecar with specified service account token volume", testCase{
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
                labels:
                  kuma.io/sidecar-injection: enabled`,
			cfgFile: "inject.config.yaml",
		}),
		Entry("27. sidecar with specified drain time", testCase{
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
                labels:
                  kuma.io/sidecar-injection: enabled`,
			cfgFile: "inject.config.yaml",
		}),
		Entry("28. sidecar with patch", testCase{
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
                labels:
                  kuma.io/sidecar-injection: enabled`,
			cfgFile: "inject.config.yaml",
		}),
		Entry("29. port override #4458", testCase{
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
                labels:
                  kuma.io/sidecar-injection: enabled`,
			cfgFile: "inject.builtindns.config.yaml",
		}),
		Entry("30. with ebpf", testCase{
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
                labels:
                  kuma.io/sidecar-injection: enabled`,
			cfgFile: "inject.ebpf.config.yaml",
		}),
		Entry("31. with duplicate container/sidecar uid", testCase{
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
                labels:
                  kuma.io/sidecar-injection: enabled`,
			cfgFile: "inject.config.yaml",
		}),
		Entry("32. init first", testCase{
			num: "32",
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
                labels:
                  kuma.io/sidecar-injection: enabled`,
			cfgFile: "inject.config.yaml",
		}),
		Entry("33. kuma.io/transparent-proxying-ip-family-mode", testCase{
			num: "33",
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
                labels:
                  kuma.io/sidecar-injection: enabled`,
			cfgFile: "inject.config-ipv6-disabled.yaml",
		}),
		Entry("34. cni enabled", testCase{
			num: "34",
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
                labels:
                  kuma.io/sidecar-injection: enabled`,
			cfgFile: "inject.config-cni.yaml",
		}),
		Entry("native sidecar with probe", testCase{
			num: "35",
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
                labels:
                  kuma.io/sidecar-injection: enabled`,
			cfgFile: "inject.config.yaml",
		}),
		Entry("36. traffic.kuma.io/drop-invalid-packets overrides config", testCase{
			num: "36",
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
                labels:
                  kuma.io/sidecar-injection: enabled`,
			cfgFile: "inject.config.yaml",
		}),
		Entry("37. traffic.kuma.io/iptables-logs overrides config", testCase{
			num: "37",
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
                labels:
                  kuma.io/sidecar-injection: enabled`,
			cfgFile: "inject.config.yaml",
		}),
<<<<<<< HEAD
=======
		Entry("38. traffic.kuma.io/exclude-outbound-ips overrides config", testCase{
			num: "38",
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
                labels:
                  kuma.io/sidecar-injection: enabled`,
			cfgFile: "inject.config.yaml",
		}),
		Entry("39. traffic.kuma.io/exclude-inbound-ips overrides config", testCase{
			num: "39",
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
                labels:
                  kuma.io/sidecar-injection: enabled`,
			cfgFile: "inject.config.yaml",
		}),
		Entry("40. application probe proxy: config - disabled, pod - enabled", testCase{
			num: "40",
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
                labels:
                  kuma.io/sidecar-injection: enabled`,
			cfgFile: "inject.vp-disabled.config.yaml",
		}),
		Entry("41. gateway provided with cni enabled", testCase{
			num: "41",
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
                labels:
                  kuma.io/sidecar-injection: enabled`,
			cfgFile: "inject.config-cni.yaml",
		}),
>>>>>>> ebcc4be57 (fix(cni): delegated gateway was not correctly injected (#11922))
	)

	DescribeTable("should not inject Kuma into a Pod",
		func(given testCase) {
			// setup
			inputFile := filepath.Join("testdata", fmt.Sprintf("skip_inject.%s.input.yaml", given.num))
			goldenFile := filepath.Join("testdata", fmt.Sprintf("skip_inject.%s.golden.yaml", given.num))

			var cfg conf.Injector
			Expect(config.Load(filepath.Join("testdata", given.cfgFile), &cfg)).To(Succeed())
			cfg.CaCertFile = caCertPath
			injector, err := inject.New(cfg, "http://kuma-control-plane.kuma-system:5681", k8sClient, false, k8s.NewSimpleConverter(), 9901, systemNamespace)
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
			for _, container := range pod.Spec.Containers {
				Expect(container.Name).To(Not(BeEquivalentTo(k8s_util.KumaSidecarContainerName)))
			}

			By("loading golden Pod")
			// when
			actual, err := yaml.Marshal(pod)
			// then
			Expect(err).ToNot(HaveOccurred())

			By("comparing actual against golden")
			Expect(actual).To(matchers.MatchGoldenYAML(goldenFile))
		},
		Entry("1. Pod with `kuma.io/sidecar-injection: disabled` annotation", testCase{
			num: "1",
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
                labels:
                  kuma.io/sidecar-injection: enabled`,
			cfgFile: "inject.config.yaml",
		}),
		Entry("2. skip injection for label exception", testCase{
			num: "2",
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
                labels:
                  kuma.io/sidecar-injection: enabled`,
			cfgFile: "inject.config.yaml",
		}),
		Entry("3. skip injection when using annotations", testCase{
			num: "3",
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
	)

	DescribeTable("should fail w/error",
		func(given testCase) {
			// setup
			inputFile := filepath.Join("testdata", fmt.Sprintf("inject.shouldfail.%s.input.yaml", given.num))

			var cfg conf.Injector
			Expect(config.Load(filepath.Join("testdata", given.cfgFile), &cfg)).To(Succeed())
			cfg.CaCertFile = caCertPath
			injector, err := inject.New(cfg, "http://kuma-control-plane.kuma-system:5681", k8sClient, false, k8s.NewSimpleConverter(), 9901, systemNamespace)
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
			Expect(err).To(HaveOccurred())
			for _, container := range pod.Spec.Containers {
				Expect(container.Name).To(Not(BeEquivalentTo(k8s_util.KumaSidecarContainerName)))
			}
		},
		Entry("1. Pod annotated with name of non-existing patch", testCase{
			num: "1",
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
                labels:
                  kuma.io/sidecar-injection: enabled`,
			cfgFile: "inject.config.yaml",
		}),
		Entry("2. Pod with existing container using same UID as sidecar", testCase{
			num: "2",
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
                labels:
                  kuma.io/sidecar-injection: enabled`,
			cfgFile: "inject.config.yaml",
		}),
	)
}, Ordered)
