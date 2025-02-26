package graceful

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/kumahq/kuma/pkg/plugins/policies/meshretry/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/channels"
	"github.com/kumahq/kuma/pkg/util/pointer"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func ChangeService() {
	const namespace = "changesvc"
	const mesh = "changesvc"

	firstTestServerLabels := map[string]string{
		"app":                  "test-server",
		"changesvc-test-label": "first",
	}

	secondTestServerLabels := map[string]string{
		"app":                  "test-server",
		"changesvc-test-label": "second",
	}

	thirdTestServerLabels := map[string]string{
		"kuma.io/sidecar-injection": "disabled",
		"app":                       "test-server",
		"changesvc-test-label":      "third",
	}

	newSvc := func(selector map[string]string) *corev1.Service {
		return &corev1.Service{
			TypeMeta: kube_meta.TypeMeta{
				Kind:       "Service",
				APIVersion: "v1",
			},
			ObjectMeta: kube_meta.ObjectMeta{
				Name:      "test-server",
				Namespace: namespace,
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{
						Name:        "main",
						Port:        int32(80),
						TargetPort:  intstr.FromString("main"),
						AppProtocol: pointer.To("htt"),
					},
				},
				Selector: selector,
			},
		}
	}

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MTLSMeshKubernetes(mesh)).
			Install(MeshTrafficPermissionAllowAllKubernetes(mesh)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Parallel(
				testserver.Install(
					testserver.WithNamespace(namespace),
					testserver.WithMesh(mesh),
					testserver.WithName("demo-client"),
				),
				testserver.Install(
					testserver.WithNamespace(namespace),
					testserver.WithMesh(mesh),
					testserver.WithName("test-server-first"),
					testserver.WithEchoArgs("echo", "--instance", "test-server-first"),
					testserver.WithoutService(),
					testserver.WithoutWaitingToBeReady(), // WaitForPods assumes that app label is name, but we change this in WithPodLabels
					testserver.WithPodLabels(firstTestServerLabels),
				),
				testserver.Install(
					testserver.WithNamespace(namespace),
					testserver.WithMesh(mesh),
					testserver.WithName("test-server-second"),
					testserver.WithEchoArgs("echo", "--instance", "test-server-second"),
					testserver.WithoutService(),
					testserver.WithoutWaitingToBeReady(), // WaitForPods assumes that app label is name, but we change this in WithPodLabels
					testserver.WithPodLabels(secondTestServerLabels),
				),
				testserver.Install(
					testserver.WithNamespace(namespace),
					testserver.WithName("test-server-third"),
					testserver.WithEchoArgs("echo", "--instance", "test-server-third"),
					testserver.WithoutService(),
					testserver.WithoutWaitingToBeReady(), // WaitForPods assumes that app label is name, but we change this in WithPodLabels
					testserver.WithPodLabels(thirdTestServerLabels),
				),
			)).
			Install(YamlK8sObject(newSvc(firstTestServerLabels))).
			Setup(kubernetes.Cluster)
		Expect(err).To(Succeed())

		// remove retries to avoid covering failed request
		Expect(DeleteMeshPolicyOrError(
			kubernetes.Cluster,
			v1alpha1.MeshRetryResourceTypeDescriptor,
			fmt.Sprintf("mesh-retry-all-%s", mesh),
		)).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugKube(kubernetes.Cluster, mesh, namespace)
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(mesh)).To(Succeed())
	})

	doRequest := func() (string, error) {
		resp, err := client.CollectEchoResponse(
			kubernetes.Cluster,
			"demo-client",
			"test-server:80",
			client.FromKubernetesPod(namespace, "demo-client"),
		)
		return resp.Instance, err
	}

	It("should gracefully switch to other service", func() {
		// given traffic to the first server
		Eventually(func(g Gomega) {
			instance, err := doRequest()
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(instance).To(Equal("test-server-first"))
		}, "30s", "1s").Should(Succeed())

		// and constant traffic in the background
		var failedErr error
		closeCh := make(chan struct{})
		defer close(closeCh)
		go func() {
			for {
				if channels.IsClosed(closeCh) {
					return
				}
				if _, err := doRequest(); err != nil {
					failedErr = err
					return
				}
				// add a slight delay to not overwhelm completely the host running this test and leave more resources to other tests running in parallel.
				time.Sleep(50 * time.Millisecond)
			}
		}()

		// when
		err := kubernetes.Cluster.Install(YamlK8sObject(newSvc(secondTestServerLabels)))

		// then traffic shifted
		Expect(err).To(Succeed())
		Eventually(func(g Gomega) {
			instance, err := doRequest()
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(instance).To(Equal("test-server-second"))
		}, "30s", "1s").Should(Succeed())

		// and we did not drop a single request
		Expect(failedErr).ToNot(HaveOccurred())
	})

	It("should switch to the instance of a service that in not in the mesh", func() {
		// given
		Expect(kubernetes.Cluster.Install(YamlK8sObject(newSvc(firstTestServerLabels)))).To(Succeed())
		Eventually(func(g Gomega) {
			instance, err := doRequest()
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(instance).To(Equal("test-server-first"))
		}, "30s", "1s").Should(Succeed())

		// when
		err := kubernetes.Cluster.Install(YamlK8sObject(newSvc(thirdTestServerLabels)))

		// then
		Expect(err).To(Succeed())
		Eventually(func(g Gomega) {
			instance, err := doRequest()
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(instance).To(Equal("test-server-third"))
		}, "30s", "1s").Should(Succeed())
	})
}
