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

	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/metadata"
	. "github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/v3/test/framework/envs/kubernetes"
)

func ApplicationProbeProxy() {
	meshName := "application-probe-proxy"
	identityName := "application-probe-proxy-identity"
	namespace := "application-probe-proxy"
	httpAppName := "http-test-server"
	gRPCAppName := "grpc-test-server"
	tcpAppName := "tcp-test-server"

	trustDomain := fmt.Sprintf("%s.default.mesh.local", meshName)

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshKubernetes(meshName)).
			Install(MeshIdentityBundledKubernetes(meshName, identityName)).
			Install(MeshTrafficPermissionAllowAllKubernetesWorkloadIdentity(meshName, trustDomain)).
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
			httpPod, err := k8s.GetPodContextE(kubernetes.Cluster.GetTesting(), context.Background(), kubernetes.Cluster.GetKubectlOptions(namespace), httpAppPodName)
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
			tcpPod, err := k8s.GetPodContextE(kubernetes.Cluster.GetTesting(), context.Background(), kubernetes.Cluster.GetKubectlOptions(namespace), tcpAppPodName)
			g.Expect(err).ToNot(HaveOccurred(), "failed to get details of pod '%s'", tcpAppPodName)

			container := getAppContainer(tcpPod, tcpAppName)
			g.Expect(container.LivenessProbe.TCPSocket).To(BeNil())
			g.Expect(container.LivenessProbe.HTTPGet).ToNot(BeNil())
			g.Expect(container.LivenessProbe.HTTPGet.Path).To(Equal("/tcp/6379"))
		}, "30s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			grpcPod, err := k8s.GetPodContextE(kubernetes.Cluster.GetTesting(), context.Background(), kubernetes.Cluster.GetKubectlOptions(namespace), grpcAppPodName)
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

		By("fourth, assert Probes data is NOT present when using application probe proxy")
		Eventually(func(g Gomega) {
			checkDPProbes := func(podName string) {
				dpName := fmt.Sprintf("%s.%s", podName, namespace)
				dpYAML, err := kubernetes.Cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplane",
					dpName, "--mesh", meshName, "-oyaml")

				g.Expect(err).ToNot(HaveOccurred(), "failed to get dataplane '%s'", dpName)
				g.Expect(dpYAML).ToNot(ContainSubstring("probes:"), "legacy probes field should be absent")
				dpRes, err := rest.YAML.UnmarshalCore([]byte(dpYAML))
				g.Expect(err).ToNot(HaveOccurred(), "invalid dataplane object")
				_, ok := dpRes.(*core_mesh.DataplaneResource)
				g.Expect(ok).To(BeTrue(), fmt.Errorf("invalid dataplane object type: %t", dpRes))
			}

			checkDPProbes(httpAppPodName)
			checkDPProbes(tcpAppPodName)
			checkDPProbes(grpcAppPodName)
		}, "30s", "1s").Should(Succeed())
	})

	It("should leave probes untouched when application probe proxy is disabled", func() {
		By("patch the application pod and disabling application probe proxy using annotation")
		kubectlOptsApps := kubernetes.Cluster.GetKubectlOptions(namespace)
		nextTemplateHash := patchAndWait(kubernetes.Cluster.GetTesting(), Default, kubernetes.Cluster, kubectlOptsApps, httpAppName,
			`[{"op": "add", "path": "/spec/template/metadata/annotations", "value": {}},{"op":"add", "path":"/spec/template/metadata/annotations/kuma.io~1application-probe-proxy-port", "value":"0"}]`)

		By("checking probes on the new pod are left untouched")
		var nextRevPodName string
		Eventually(func(g Gomega) {
			httpPods, err := k8s.ListPodsContextE(kubernetes.Cluster.GetTesting(), context.Background(), kubectlOptsApps,
				metav1.ListOptions{LabelSelector: "pod-template-hash=" + nextTemplateHash})

			g.Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("failed to list new pods of '%s'", httpAppName))
			g.Expect(httpPods).ToNot(BeEmpty())

			httpPod := httpPods[0]
			nextRevPodName = httpPod.Name
			g.Expect(httpPod.Annotations[metadata.KumaApplicationProbeProxyPortAnnotation]).To(Equal("0"))

			container := getAppContainer(&httpPod, httpAppName)
			g.Expect(container).ToNot(BeNil())
			g.Expect(container.ReadinessProbe.HTTPGet).ToNot(BeNil())
			g.Expect(container.ReadinessProbe.HTTPGet.Port.IntValue()).To(Equal(80))
			g.Expect(container.ReadinessProbe.HTTPGet.Path).To(Equal("/probes?type=readiness"))
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
	kubeClient, err := k8s.GetKubernetesClientFromOptionsContextE(t, context.Background(), kubectlOpts)
	g.Expect(err).ToNot(HaveOccurred())

	prevDeployObj, err := kubeClient.AppsV1().Deployments(kubectlOpts.Namespace).
		Patch(context.Background(), appName, types.JSONPatchType, []byte(jsonPatch), metav1.PatchOptions{})
	g.Expect(err).ToNot(HaveOccurred())

	prevRevision := prevDeployObj.Annotations["deployment.kubernetes.io/revision"]
	prevRevisionNum, _ := strconv.Atoi(prevRevision)
	nextRevision := strconv.Itoa(prevRevisionNum + 1)
	var nextRS *appsv1.ReplicaSet
	Eventually(func() error {
		rsList := k8s.ListReplicaSetsContext(t, context.Background(), kubectlOpts, metav1.ListOptions{LabelSelector: "app=" + appName})
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
	pod, err := k8s.GetPodContextE(t, context.Background(), kubectlOpts, podName)
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
