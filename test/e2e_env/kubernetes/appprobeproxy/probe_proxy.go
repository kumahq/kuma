package appprobeproxy

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
	. "github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/v2/test/framework/envs/kubernetes"
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
			Install(Parallel(
				testserver.Install(
					testserver.WithName(httpAppName),
					testserver.WithMesh(meshName),
					testserver.WithNamespace(namespace),
					testserver.WithArgs("echo", "--port", "80", "--probes"),
					testserver.WithProbe(testserver.ReadinessProbe, testserver.ProbeHttpGet, 80, "/probes?type=readiness"),
				),
				testserver.Install(
					testserver.WithName(tcpAppName),
					testserver.WithMesh(meshName),
					testserver.WithNamespace(namespace),
					testserver.WithArgs("health-check", "tcp", "--port", "6379"),
					testserver.WithProbe(testserver.LivenessProbe, testserver.ProbeTcpSocket, 6379, ""),
				),
				testserver.Install(
					testserver.WithName(gRPCAppName),
					testserver.WithMesh(meshName),
					testserver.WithNamespace(namespace),
					testserver.WithArgs("grpc", "server", "--port", "8080"),
					testserver.WithProbe(testserver.StartupProbe, testserver.ProbeGRPC, 8080, ""),
				),
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
