package skipinboundtags

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/meshretry/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
	"github.com/kumahq/kuma/v2/pkg/util/channels"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
	. "github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/client"
	"github.com/kumahq/kuma/v2/test/framework/deployments/testserver"
)

// ChangeService is test/e2e_env/kubernetes/graceful/change_service.go run with MeshService
// Exclusive, inbound tags disabled and "changesvc-test-label" in ignoredServiceSelectorLabels.
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
			Kind:       "Service",
			APIVersion: "v1",
			Name:       "test-server",
			Namespace:  namespace,
			Labels: map[string]string{
				"kuma.io/mesh": mesh,
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{
						Name:        "main",
						Port:        int32(80),
						TargetPort:  intstr.FromString("main"),
						AppProtocol: pointer.To("http"),
					},
				},
				Selector: selector,
			},
		}
	}

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(Yaml(
				builders.Mesh().
					WithName(mesh).
					WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Exclusive),
			)).
			// standalone zone CP without global resolves {{ .Zone }} to "default"
			Install(MeshTrafficPermissionAllowAllKubernetesWorkloadIdentity(mesh,
				fmt.Sprintf("%s.default.mesh.local", mesh))).
			Install(YamlK8s(meshIdentityYAML(mesh))).
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
			Setup(KubeCluster)
		Expect(err).To(Succeed())

		// remove retries to avoid covering failed request
		Expect(DeleteMeshPolicyOrError(
			KubeCluster,
			v1alpha1.MeshRetryResourceTypeDescriptor,
			fmt.Sprintf("mesh-retry-all-%s", mesh),
		)).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugKube(KubeCluster, mesh, namespace)
	})

	E2EAfterAll(func() {
		Expect(KubeCluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(KubeCluster.DeleteMesh(mesh)).To(Succeed())
	})

	doRequest := func() (string, error) {
		resp, err := client.CollectEchoResponse(
			KubeCluster,
			"demo-client",
			"test-server:80",
			client.FromKubernetesPod(namespace, "demo-client"),
		)
		return resp.Instance, err
	}

	It("should gracefully switch to other service", func() {
		// Keep service bootstrap outside the selector transition. A successful
		// plaintext request does not mean clients have received the mTLS cluster.
		Eventually(func(g Gomega) {
			_, status, err := GetMeshServiceStatus(KubeCluster, "test-server."+namespace, mesh)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(status.TLS.Status).To(Equal(meshservice_api.TLSReady))
		}, "30s", "1s").Should(Succeed())

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
		err := KubeCluster.Install(YamlK8sObject(newSvc(secondTestServerLabels)))

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

	It("should fail fast when the service selects instances outside the mesh", func() {
		// given
		Expect(KubeCluster.Install(YamlK8sObject(newSvc(firstTestServerLabels)))).To(Succeed())
		Eventually(func(g Gomega) {
			instance, err := doRequest()
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(instance).To(Equal("test-server-first"))
		}, "30s", "1s").Should(Succeed())

		// when the selector moves onto pods that have sidecar injection disabled
		err := KubeCluster.Install(YamlK8sObject(newSvc(thirdTestServerLabels)))

		// then the MeshService selects dataplanes, those pods have none, so
		// requests fail fast instead of hanging
		Expect(err).To(Succeed())
		Eventually(func(g Gomega) {
			resp, err := client.CollectFailure(
				KubeCluster,
				"demo-client",
				"test-server:80",
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.ResponseCode).To(Equal(503))
		}, "30s", "1s").Should(Succeed())
	})
}
