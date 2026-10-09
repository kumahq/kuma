package skipinboundtags

import (
	"context"
	"fmt"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshservice/api/v1alpha1"
	meshretry_api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshretry/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
	"github.com/kumahq/kuma/v2/pkg/util/channels"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
	. "github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/client"
	"github.com/kumahq/kuma/v2/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/v2/test/framework/deployments/testserver"
)

var KubeCluster *K8sCluster

func SkipInboundTags() {
	meshName := "skip-inbound-tags"
	namespace := "skip-inbound-tags-ns"

	meshIdentity := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshIdentity
metadata:
  name: identity-skip-inbound-tags
  namespace: %s
  labels:
    kuma.io/mesh: %s
    kuma.io/origin: zone
spec:
  selector:
    dataplane:
      matchLabels: {}
  spiffeID:
    trustDomain: "{{ .Mesh }}.{{ .Zone }}.mesh.local"
    path: "/ns/{{ .Namespace }}/sa/{{ .ServiceAccount }}"
  provider:
    type: Bundled
    bundled:
      meshTrustCreation: Enabled
      insecureAllowSelfSigned: true
      certificateParameters:
        expiry: 24h
      autogenerate:
        enabled: true
`, Config.KumaNamespace, meshName)

	// changesvc-test-label is in ignoredServiceSelectorLabels, so the Service selects
	// both servers once it is stripped, and only one of them with the full selector
	changeSvcLabels := func(instance string) map[string]string {
		return map[string]string{
			"app":                  "changesvc",
			"changesvc-test-label": instance,
		}
	}
	changeSvcServer := func(instance string) InstallFunc {
		return testserver.Install(
			testserver.WithName("changesvc-"+instance),
			testserver.WithMesh(meshName),
			testserver.WithNamespace(namespace),
			testserver.WithEchoArgs("echo", "--instance", "changesvc-"+instance),
			testserver.WithoutService(),
			testserver.WithoutWaitingToBeReady(), // WaitForPods assumes that app label is name, but we change this in WithPodLabels
			testserver.WithPodLabels(changeSvcLabels(instance)),
		)
	}
	changeSvc := func(instance string) *corev1.Service {
		return &corev1.Service{
			Kind: "Service", APIVersion: "v1",
			Name:      "changesvc",
			Namespace: namespace,
			Labels:    map[string]string{"kuma.io/mesh": meshName},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{{
					Name:        "main",
					Port:        80,
					TargetPort:  intstr.FromString("main"),
					AppProtocol: pointer.To("http"),
				}},
				Selector: changeSvcLabels(instance),
			},
		}
	}

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Yaml(
				builders.Mesh().
					WithName(meshName).
					WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Exclusive),
			)).
			// standalone zone CP without global resolves {{ .Zone }} to "default"
			Install(MeshTrafficPermissionAllowAllKubernetesWorkloadIdentity(meshName,
				fmt.Sprintf("%s.default.mesh.local", meshName))).
			Install(YamlK8s(meshIdentity)).
			Install(Parallel(
				testserver.Install(
					testserver.WithName("test-server"),
					testserver.WithMesh(meshName),
					testserver.WithNamespace(namespace),
				),
				democlient.Install(
					democlient.WithName("demo-client"),
					democlient.WithMesh(meshName),
					democlient.WithNamespace(namespace),
				),
				changeSvcServer("first"),
				changeSvcServer("second"),
			)).
			Install(YamlK8sObject(changeSvc("first"))).
			Setup(KubeCluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugKube(KubeCluster, meshName, namespace)
	})

	E2EAfterAll(func() {
		Expect(KubeCluster.DeleteNamespace(namespace)).To(Succeed())
		Expect(KubeCluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should generate dataplane with empty inbound tags", func() {
		Eventually(func(g Gomega) {
			out, err := k8s.RunKubectlAndGetOutputContextE(
				KubeCluster.GetTesting(), context.Background(),
				KubeCluster.GetKubectlOptions(Config.KumaNamespace),
				"get", "dataplanes", "-n", Config.KumaNamespace,
				"-l", fmt.Sprintf("k8s.kuma.io/namespace=%s", namespace),
				"-o", "jsonpath={.items[*].spec.networking.inbound[*].tags}",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(out).To(BeEmpty(), "inbound tags should be empty")
		}, "60s", "1s").Should(Succeed())
	})

	It("should generate MeshService with dataplaneLabels selector", func() {
		Eventually(func(g Gomega) {
			out, err := k8s.RunKubectlAndGetOutputContextE(
				KubeCluster.GetTesting(), context.Background(),
				KubeCluster.GetKubectlOptions(namespace),
				"get", "meshservices", "-n", namespace,
				"-l", fmt.Sprintf("kuma.io/mesh=%s", meshName),
				"-o", "jsonpath={.items[*].spec.selector.dataplaneTags}",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(out).To(BeEmpty(), "MeshService should not have dataplaneTags selector")
		}, "60s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			out, err := k8s.RunKubectlAndGetOutputContextE(
				KubeCluster.GetTesting(), context.Background(),
				KubeCluster.GetKubectlOptions(namespace),
				"get", "meshservices", "-n", namespace,
				"-l", fmt.Sprintf("kuma.io/mesh=%s", meshName),
				"-o", "jsonpath={.items[*].spec.selector.dataplaneLabels}",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(out).ToNot(BeEmpty(), "MeshService should have dataplaneLabels selector")
		}, "60s", "1s").Should(Succeed())
	})

	It("should allow traffic between services", func() {
		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				KubeCluster,
				"demo-client",
				fmt.Sprintf("test-server.%s.svc.cluster.local", namespace),
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "60s", "1s").MustPassRepeatedly(5).Should(Succeed())
	})

	It("should not drop requests when the Service selector moves to other pods", func() {
		doRequest := func() (string, error) {
			resp, err := client.CollectEchoResponse(
				KubeCluster,
				"demo-client",
				fmt.Sprintf("changesvc.%s.svc.cluster.local", namespace),
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			return resp.Instance, err
		}

		// remove retries to avoid covering failed requests
		Expect(DeleteMeshPolicyOrError(
			KubeCluster,
			meshretry_api.MeshRetryResourceTypeDescriptor,
			fmt.Sprintf("mesh-retry-all-%s", meshName),
		)).To(Succeed())

		// keep service bootstrap outside the selector move
		Eventually(func(g Gomega) {
			_, status, err := GetMeshServiceStatus(KubeCluster, "changesvc."+namespace, meshName)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(status.TLS.Status).To(Equal(meshservice_api.TLSReady))
		}, "60s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			instance, err := doRequest()
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(instance).To(Equal("changesvc-first"))
		}, "60s", "1s").Should(Succeed())

		var failedErr error
		closeCh := make(chan struct{})
		defer close(closeCh)
		go func() {
			for !channels.IsClosed(closeCh) {
				if _, err := doRequest(); err != nil {
					failedErr = err
					return
				}
				time.Sleep(50 * time.Millisecond)
			}
		}()

		// when
		Expect(KubeCluster.Install(YamlK8sObject(changeSvc("second")))).To(Succeed())

		// then
		Eventually(func(g Gomega) {
			instance, err := doRequest()
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(instance).To(Equal("changesvc-second"))
		}, "60s", "1s").Should(Succeed())
		Expect(failedErr).ToNot(HaveOccurred())
	})
}
