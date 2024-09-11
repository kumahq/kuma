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

		// first, we get the pod names
		Eventually(func() error {
			var err error
			httpAppPodName, err = PodNameOfApp(kubernetes.Cluster, httpAppName, namespace)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to get pod of '%s'", httpAppName))
			}

			tcpAppPodName, err = PodNameOfApp(kubernetes.Cluster, tcpAppName, namespace)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to get pod of '%s'", tcpAppName))
			}

			grpcAppPodName, err = PodNameOfApp(kubernetes.Cluster, gRPCAppName, namespace)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to get pod of '%s'", gRPCAppName))
			}
			return nil
		}, "30s", "1s").ShouldNot(HaveOccurred())

		// second, assert probes are converted to HTTPGet
		Eventually(func() error {
			httpPod, err := k8s.GetPodE(kubernetes.Cluster.GetTesting(), kubernetes.Cluster.GetKubectlOptions(namespace), httpAppPodName)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to get details of pod '%s'", httpAppPodName))
			}
			Expect(httpPod).ToNot(BeNil())

			probeProxyPortAnno := httpPod.Annotations[metadata.KumaApplicationProbeProxyPortAnnotation]
			Expect(probeProxyPortAnno).ToNot(BeEmpty())
			Expect(httpPod.Annotations[metadata.KumaVirtualProbesPortAnnotation]).ToNot(BeEmpty())

			container := getAppContainer(httpPod, httpAppName)
			Expect(container).ToNot(BeNil())
			Expect(container.ReadinessProbe.HTTPGet).ToNot(BeNil())
			probeProxyPort, _ := strconv.Atoi(probeProxyPortAnno)
			Expect(container.ReadinessProbe.HTTPGet.Port).To(Equal(intstr.FromInt32(int32(probeProxyPort)))) //nolint:gosec  // we never overflow here
			Expect(container.ReadinessProbe.HTTPGet.Path).To(Equal("/80/probes?type=readiness"))
			return nil
		}, "30s", "1s").ShouldNot(HaveOccurred())

		Eventually(func() error {
			tcpPod, err := k8s.GetPodE(kubernetes.Cluster.GetTesting(), kubernetes.Cluster.GetKubectlOptions(namespace), tcpAppPodName)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to get details of pod '%s'", tcpAppPodName))
			}

			container := getAppContainer(tcpPod, tcpAppName)
			Expect(container.LivenessProbe.TCPSocket).To(BeNil())
			Expect(container.LivenessProbe.HTTPGet).ToNot(BeNil())
			Expect(container.LivenessProbe.HTTPGet.Path).To(Equal("/tcp/6379"))
			return nil
		}, "30s", "1s").ShouldNot(HaveOccurred())

		Eventually(func() error {
			grpcPod, err := k8s.GetPodE(kubernetes.Cluster.GetTesting(), kubernetes.Cluster.GetKubectlOptions(namespace), grpcAppPodName)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to get details of pod '%s'", grpcAppPodName))
			}

			container := getAppContainer(grpcPod, gRPCAppName)
			Expect(container.StartupProbe.GRPC).To(BeNil())
			Expect(container.StartupProbe.HTTPGet).ToNot(BeNil())
			Expect(container.StartupProbe.HTTPGet.Path).To(Equal("/grpc/8080"))
			return nil
		}, "30s", "1s").ShouldNot(HaveOccurred())

		// third, assert pods are ready and live
		Consistently(func() error {
			if err := checkIfAppReady(kubernetes.Cluster.GetTesting(), kubernetes.Cluster.GetKubectlOptions(namespace),
				httpAppPodName, httpAppName); err != nil {
				return err
			}
			if err := checkIfAppReady(kubernetes.Cluster.GetTesting(), kubernetes.Cluster.GetKubectlOptions(namespace),
				tcpAppPodName, tcpAppName); err != nil {
				return err
			}
			if err := checkIfAppReady(kubernetes.Cluster.GetTesting(), kubernetes.Cluster.GetKubectlOptions(namespace),
				grpcAppPodName, gRPCAppName); err != nil {
				return err
			}
			return nil
		}, "30s", "3s", MustPassRepeatedly(2)).ShouldNot(HaveOccurred())

		// fourth, assert Probes data is present for HTTP Probes
		Eventually(func() error {
			checkDPProbes := func(podName string, shouldHasProbes bool) error {
				dpName := fmt.Sprintf("%s.%s", podName, namespace)
				dpYAML, err := kubernetes.Cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplane",
					dpName, "--mesh", meshName, "-oyaml")
				if err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to get dataplane '%s'", dpName))
				}

				dpRes, err := rest.YAML.UnmarshalCore([]byte(dpYAML))
				Expect(err).ToNot(HaveOccurred())
				dp, ok := dpRes.(*core_mesh.DataplaneResource)
				Expect(ok).To(BeTrue())

				// 9000 is the default port of Virtual Probes
				// Probes field should always be available regardless if it has any HTTP probes on the Pod
				Expect(dp.Spec.Probes.Port).To(Equal(uint32(9000)))
				if shouldHasProbes {
					Expect(dp.Spec.Probes.Endpoints).ToNot(BeEmpty())
				} else {
					Expect(dp.Spec.Probes.Endpoints).To(BeEmpty())
				}
				return nil
			}

			if err := checkDPProbes(httpAppPodName, true); err != nil {
				return err
			}
			if err := checkDPProbes(tcpAppPodName, false); err != nil {
				return err
			}
			if err := checkDPProbes(grpcAppPodName, false); err != nil {
				return err
			}
			return nil
		}, "30s", "1s").ShouldNot(HaveOccurred())
	})

	It("should fallback to virtual probes when application probe proxy is disabled", func() {
		kubectlOptsApps := kubernetes.Cluster.GetKubectlOptions(namespace)
		nextTemplateHash := patchAndWait(kubernetes.Cluster.GetTesting(), kubernetes.Cluster, kubectlOptsApps, httpAppName,
			`[{"op":"add", "path":"/spec/template/metadata/annotations/kuma.io~1application-probe-proxy-port", "value":"0"}]`)

		var nextRevPodName string
		// assert the Pod has application probe proxy disabled and virtual probes replaces
		Eventually(func() error {
			httpPods, err := k8s.ListPodsE(kubernetes.Cluster.GetTesting(), kubectlOptsApps,
				metav1.ListOptions{LabelSelector: "pod-template-hash=" + nextTemplateHash})
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to list new pods of '%s'", httpAppName))
			}

			Expect(httpPods).ToNot(BeEmpty())
			httpPod := httpPods[0]
			nextRevPodName = httpPod.Name
			virtualProbesPortAnno := httpPod.Annotations[metadata.KumaVirtualProbesPortAnnotation]
			Expect(virtualProbesPortAnno).To(Equal("9000"))

			container := getAppContainer(&httpPod, httpAppName)
			Expect(container).ToNot(BeNil())
			Expect(container.ReadinessProbe.HTTPGet).ToNot(BeNil())
			port, _ := strconv.Atoi(virtualProbesPortAnno)
			Expect(container.ReadinessProbe.HTTPGet.Port).To(Equal(intstr.FromInt32(int32(port)))) //nolint:gosec  // we never overflow here
			Expect(container.ReadinessProbe.HTTPGet.Path).To(Equal("/80/probes?type=readiness"))
			return nil
		}, "30s", "1s").ShouldNot(HaveOccurred())

		Consistently(func() error {
			if err := checkIfAppReady(kubernetes.Cluster.GetTesting(), kubernetes.Cluster.GetKubectlOptions(namespace),
				nextRevPodName, httpAppName); err != nil {
				return err
			}
			return nil
		}, "30s", "3s", MustPassRepeatedly(2)).ShouldNot(HaveOccurred())
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

func patchAndWait(t testing.TestingT, cluster Cluster, kubectlOpts *k8s.KubectlOptions, appName string, jsonPatch string) string {
	kubeClient, err := k8s.GetKubernetesClientFromOptionsE(t, kubectlOpts)
	Expect(err).ToNot(HaveOccurred())

	updatedDeployObj, err := kubeClient.AppsV1().Deployments(kubectlOpts.Namespace).
		Patch(context.Background(), appName, types.JSONPatchType, []byte(jsonPatch), metav1.PatchOptions{})
	Expect(err).ToNot(HaveOccurred())

	prevRevision := updatedDeployObj.Annotations["deployment.kubernetes.io/revision"]
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
	Expect(WaitPodsAvailableWithLabel(kubectlOpts.Namespace, "pod-template-hash", nextRSHash)(cluster)).To(Succeed())

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
