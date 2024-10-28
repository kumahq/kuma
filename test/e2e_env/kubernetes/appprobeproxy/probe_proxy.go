package appprobeproxy

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func ApplicationProbeProxy() {
	meshName := "application-probe-proxy"
	namespace := "application-probe-proxy"
	httpAppName := "http-test-server"
	gRPCAppName := "grpc-test-server"
	tcpAppName := "tcp-test-server"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MTLSMeshKubernetes(meshName)).
			Install(MeshTrafficPermissionAllowAllKubernetes(meshName)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(testserver.Install(
				testserver.WithName(httpAppName),
				testserver.WithMesh(meshName),
				testserver.WithNamespace(namespace),
				testserver.WithArgs("echo", "--port", "80", "--probes"),
				testserver.WithProbe(testserver.ReadinessProbe, testserver.ProbeHttpGet, 80, "/probes?type=readiness"),
			)).
			Install(testserver.Install(
				testserver.WithName(tcpAppName),
				testserver.WithMesh(meshName),
				testserver.WithNamespace(namespace),
				testserver.WithArgs("health-check", "tcp", "--port", "6379"),
				testserver.WithProbe(testserver.LivenessProbe, testserver.ProbeTcpSocket, 6379, ""),
			)).
			Install(testserver.Install(
				testserver.WithName(gRPCAppName),
				testserver.WithMesh(meshName),
				testserver.WithNamespace(namespace),
				testserver.WithArgs("grpc", "server", "--port", "8080"),
				testserver.WithProbe(testserver.StartupProbe, testserver.ProbeGRPC, 8080, ""),
			)).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugKube(kubernetes.Cluster, meshName, namespace)
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should setup application app proxy", func() {
		var httpAppPodName string
		var tcpAppPodName string
		var grpcAppPodName string

		By("first, we get the pod names")
		Eventually(func(g Gomega) {
			var err error
			httpAppPodName, err = PodNameOfApp(kubernetes.Cluster, httpAppName, namespace)
			g.Expect(err).ToNot(HaveOccurred(), "failed to get pod of '%s'", httpAppName)

			tcpAppPodName, err = PodNameOfApp(kubernetes.Cluster, tcpAppName, namespace)
			g.Expect(err).ToNot(HaveOccurred(), "failed to get pod of '%s'", tcpAppName)

			grpcAppPodName, err = PodNameOfApp(kubernetes.Cluster, gRPCAppName, namespace)
			g.Expect(err).ToNot(HaveOccurred(), "failed to get pod of '%s'", gRPCAppName)
		}, "30s", "1s").Should(Succeed())

		By("second, assert probes are converted to HTTPGet")
		Eventually(func(g Gomega) {
			httpPod, err := k8s.GetPodE(kubernetes.Cluster.GetTesting(), kubernetes.Cluster.GetKubectlOptions(namespace), httpAppPodName)
			g.Expect(err).ToNot(HaveOccurred(), "failed to get details of pod '%s'", httpAppPodName)
			g.Expect(httpPod).ToNot(BeNil())

			probeProxyPortAnno := httpPod.Annotations[metadata.KumaApplicationProbeProxyPortAnnotation]
			g.Expect(probeProxyPortAnno).ToNot(BeEmpty())
			g.Expect(httpPod.Annotations[metadata.KumaVirtualProbesPortAnnotation]).ToNot(BeEmpty())

			container := getAppContainer(httpPod, httpAppName)
			g.Expect(container).ToNot(BeNil())
			g.Expect(container.ReadinessProbe.HTTPGet).ToNot(BeNil())
			port := intstr.FromString(probeProxyPortAnno)
			g.Expect(container.ReadinessProbe.HTTPGet.Port.IntValue()).To(Equal(port.IntValue()))
			g.Expect(container.ReadinessProbe.HTTPGet.Path).To(Equal("/80/probes?type=readiness"))
		}, "30s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			tcpPod, err := k8s.GetPodE(kubernetes.Cluster.GetTesting(), kubernetes.Cluster.GetKubectlOptions(namespace), tcpAppPodName)
			g.Expect(err).ToNot(HaveOccurred(), "failed to get details of pod '%s'", tcpAppPodName)

			container := getAppContainer(tcpPod, tcpAppName)
			g.Expect(container.LivenessProbe.TCPSocket).To(BeNil())
			g.Expect(container.LivenessProbe.HTTPGet).ToNot(BeNil())
			g.Expect(container.LivenessProbe.HTTPGet.Path).To(Equal("/tcp/6379"))
		}, "30s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			grpcPod, err := k8s.GetPodE(kubernetes.Cluster.GetTesting(), kubernetes.Cluster.GetKubectlOptions(namespace), grpcAppPodName)
			g.Expect(err).ToNot(HaveOccurred(), "failed to get details of pod '%s'", grpcAppPodName)

			container := getAppContainer(grpcPod, gRPCAppName)
			g.Expect(container.StartupProbe.GRPC).To(BeNil())
			g.Expect(container.StartupProbe.HTTPGet).ToNot(BeNil())
			g.Expect(container.StartupProbe.HTTPGet.Path).To(Equal("/grpc/8080"))
		}, "30s", "1s").Should(Succeed())

		By("third, assert pods are ready and live")
		Eventually(func(g Gomega) {
			var err error
			err = checkIfAppReady(kubernetes.Cluster.GetTesting(), kubernetes.Cluster.GetKubectlOptions(namespace), httpAppPodName, httpAppName)
			g.Expect(err).ToNot(HaveOccurred())

			err = checkIfAppReady(kubernetes.Cluster.GetTesting(), kubernetes.Cluster.GetKubectlOptions(namespace), tcpAppPodName, tcpAppName)
			g.Expect(err).ToNot(HaveOccurred())

			err = checkIfAppReady(kubernetes.Cluster.GetTesting(), kubernetes.Cluster.GetKubectlOptions(namespace), grpcAppPodName, gRPCAppName)
			g.Expect(err).ToNot(HaveOccurred())
		}, "30s", "3s").Should(Succeed())

		By("fourth, assert Probes data is present for HTTP Probes")
		Eventually(func(g Gomega) {
			checkDPProbes := func(podName string, shouldHasProbes bool) {
				dpName := fmt.Sprintf("%s.%s", podName, namespace)
				dpYAML, err := kubernetes.Cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplane",
					dpName, "--mesh", meshName, "-oyaml")

				g.Expect(err).ToNot(HaveOccurred(), "failed to get dataplane '%s'", dpName)
				dpRes, err := rest.YAML.UnmarshalCore([]byte(dpYAML))
				g.Expect(err).ToNot(HaveOccurred(), "invalid dataplane object")
				dp, ok := dpRes.(*core_mesh.DataplaneResource)
				g.Expect(ok).To(BeTrue(), fmt.Errorf("invalid dataplane object type: %t", dpRes))

				// 9000 is the default port of Virtual Probes
				// Probes field should always be available regardless if it has any HTTP probes on the Pod
				g.Expect(dp.Spec.Probes.Port).To(Equal(uint32(9000)))

				if shouldHasProbes {
					g.Expect(dp.Spec.Probes.Endpoints).ToNot(BeEmpty())
				} else {
					g.Expect(dp.Spec.Probes.Endpoints).To(BeEmpty())
				}
			}

			checkDPProbes(httpAppPodName, true)
			checkDPProbes(tcpAppPodName, false)
			checkDPProbes(grpcAppPodName, false)
		}, "30s", "1s").Should(Succeed())
	})

	It("should fallback to virtual probes when application probe proxy is disabled", func() {
		By("patch the application pod and disabling application probe proxy using annotation")
		kubectlOptsApps := kubernetes.Cluster.GetKubectlOptions(namespace)
		nextTemplateHash := patchAndWait(kubernetes.Cluster.GetTesting(), Default, kubernetes.Cluster, kubectlOptsApps, httpAppName,
			`[{"op": "add", "path": "/spec/template/metadata/annotations", "value": {}},{"op":"add", "path":"/spec/template/metadata/annotations/kuma.io~1application-probe-proxy-port", "value":"0"}]`)

		By("checking virtual probes annotations on the new pod")
		var nextRevPodName string
		// assert the Pod has application probe proxy disabled and virtual probes replaces
		Eventually(func(g Gomega) {
			httpPods, err := k8s.ListPodsE(kubernetes.Cluster.GetTesting(), kubectlOptsApps,
				metav1.ListOptions{LabelSelector: "pod-template-hash=" + nextTemplateHash})

			g.Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("failed to list new pods of '%s'", httpAppName))
			g.Expect(httpPods).ToNot(BeEmpty())

			httpPod := httpPods[0]
			nextRevPodName = httpPod.Name
			virtualProbesPortAnno := httpPod.Annotations[metadata.KumaVirtualProbesPortAnnotation]
			g.Expect(virtualProbesPortAnno).To(Equal("9000"))

			container := getAppContainer(&httpPod, httpAppName)
			g.Expect(container).ToNot(BeNil())
			g.Expect(container.ReadinessProbe.HTTPGet).ToNot(BeNil())

			port := intstr.FromString(virtualProbesPortAnno)
			g.Expect(container.ReadinessProbe.HTTPGet.Port.IntValue()).To(Equal(port.IntValue()))
			g.Expect(container.ReadinessProbe.HTTPGet.Path).To(Equal("/80/probes?type=readiness"))
		}, "30s", "1s").Should(Succeed())

		By("making sure the new pod is ready")
		Eventually(func(g Gomega) {
			err := checkIfAppReady(kubernetes.Cluster.GetTesting(), kubernetes.Cluster.GetKubectlOptions(namespace), nextRevPodName, httpAppName)
			g.Expect(err).ToNot(HaveOccurred())
		}, "30s", "3s").Should(Succeed())
	})
}

func getAppContainer(pod *corev1.Pod, appName string) *corev1.Container {
	for _, c := range pod.Spec.Containers {
		if c.Name == appName {
			return &c
		}
	}
	return nil
}

func patchAndWait(t testing.TestingT, g Gomega, cluster Cluster, kubectlOpts *k8s.KubectlOptions, appName string, jsonPatch string) string {
	kubeClient, err := k8s.GetKubernetesClientFromOptionsE(t, kubectlOpts)
	g.Expect(err).ToNot(HaveOccurred())

	prevDeployObj, err := kubeClient.AppsV1().Deployments(kubectlOpts.Namespace).
		Patch(context.Background(), appName, types.JSONPatchType, []byte(jsonPatch), metav1.PatchOptions{})
	g.Expect(err).ToNot(HaveOccurred())

	prevRevision := prevDeployObj.Annotations["deployment.kubernetes.io/revision"]
	prevRevisionNum, _ := strconv.Atoi(prevRevision)
	nextRevision := strconv.Itoa(prevRevisionNum + 1)
	var nextRS *appsv1.ReplicaSet
	Eventually(func() error {
		rsList := k8s.ListReplicaSets(t, kubectlOpts, metav1.ListOptions{LabelSelector: "app=" + appName})
		for _, rs := range rsList {
			if rs.Annotations["deployment.kubernetes.io/revision"] == nextRevision {
				nextRS = &rs
				break
			}
		}
		if nextRS != nil {
			return nil
		}
		return fmt.Errorf("failed to find the latest ReplicaSet for Deployment %s", appName)
	}, "30s", "2s").ShouldNot(HaveOccurred(), "failed to find the latest ReplicaSet for Deployment %s", appName)

	nextRSHash := nextRS.Labels["pod-template-hash"]
	g.Expect(WaitPodsAvailableWithLabel(kubectlOpts.Namespace, "pod-template-hash", nextRSHash)(cluster)).To(Succeed())

	return nextRSHash
}

func checkIfAppReady(t testing.TestingT, kubectlOpts *k8s.KubectlOptions, podName, appName string) error {
	pod, err := k8s.GetPodE(t, kubectlOpts, podName)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get details of pod '%s'", podName))
	}

	if !isTestServerReady(pod, appName) {
		return errors.Errorf("pod '%s' is not ready", podName)
	}
	return nil
}

func isTestServerReady(pod *corev1.Pod, appName string) bool {
	for _, c := range pod.Status.ContainerStatuses {
		if c.Name == appName {
			return c.Ready
		}
	}
	return false
}
