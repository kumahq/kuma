package meshidentity

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	meshidentity_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/spire"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envoy_admin/stats"
	"github.com/kumahq/kuma/test/framework/envoy_admin/tunnel"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func Spire() {
	meshName := "meshidentity-spire"
	namespace := "meshidentity-spire"
	spireNamespace := "spire-system"
	trustDomain := fmt.Sprintf("%s.local-zone.mesh.local", meshName)

	workflowRegistration := fmt.Sprintf(`
apiVersion: spire.spiffe.io/v1alpha1
kind: ClusterSPIFFEID
metadata:
  name: spire-registration
spec:
  spiffeIDTemplate: "spiffe://{{ .TrustDomain }}/ns/{{ .PodMeta.Namespace }}/sa/{{ .PodSpec.ServiceAccountName }}"
  podSelector:
    matchLabels:
      k8s.kuma.io/spire-support: "true"
  workloadSelectorTemplates:
    - "k8s:ns:%s"
`, namespace)

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(YamlK8s(samples.MeshDefaultBuilder().WithName(meshName).WithMeshServicesEnabled(v1alpha1.Mesh_MeshServices_Exclusive).KubeYaml())).
			Install(MeshTrafficPermissionAllowAllKubernetes(meshName)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Namespace(spireNamespace)).
			Install(spire.Install(
				spire.WithName("spire"),
				spire.WithNamespace(spireNamespace),
				spire.WithTrustDomain(trustDomain),
				// default spire helm uses 1.27.16 which fails since there is no image
				spire.WithKubectlVersion("v1.31.11"),
			)).
			Install(YamlK8s(workflowRegistration)).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugKube(kubernetes.Cluster, meshName, namespace)
	})

	E2EAfterEach(func() {
		Expect(DeleteMeshResources(kubernetes.Cluster, meshName, meshidentity_api.MeshIdentityResourceTypeDescriptor)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(spireNamespace)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should use MeshHTTPRoute if no TrafficRoutes are present", func() {
		// when
		yaml := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshIdentity
metadata:
  name: identity-spire
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  selector:
    dataplane:
      matchLabels: {}
  spiffeID:
    trustDomain: %s
    path: "/ns/{{ .Namespace }}/sa/{{ .ServiceAccount }}"
  provider:
    type: Spire
    spire: {}
`, Config.KumaNamespace, meshName, trustDomain)
		Expect(NewClusterSetup().
			Install(YamlK8s(yaml)).
			Install(testserver.Install(
				testserver.WithName("test-server"),
				testserver.WithMesh(meshName),
				testserver.WithNamespace(namespace),
				testserver.WithEchoArgs("echo", "--instance", "test-server-spire"),
				testserver.WithPodAnnotations(map[string]string{
					metadata.KumaSpireSupport: "true",
				}),
			)).
			Install(democlient.Install(
				democlient.WithNamespace(namespace),
				democlient.WithMesh(meshName),
				democlient.WithPodAnnotations(map[string]string{
					metadata.KumaSpireSupport: "true",
				}),
			)).
			Setup(kubernetes.Cluster)).To(Succeed())

		// then
		// traffic works
		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(kubernetes.Cluster, "demo-client", fmt.Sprintf("test-server.%s:80", namespace), client.FromKubernetesPod(namespace, "demo-client"))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(Equal("test-server-spire"))
		}, "30s", "1s", MustPassRepeatedly(5)).Should(Succeed())

		portFwd, err := kubernetes.Cluster.PortForwardApp("test-server", namespace, 9901)
		Expect(err).ToNot(HaveOccurred())

		adminTunnel, err := tunnel.NewK8sEnvoyAdminTunnel(kubernetes.Cluster.GetTesting(), portFwd.Endpoint())
		Expect(err).ToNot(HaveOccurred())

		// and it's a tls traffic
		Eventually(func(g Gomega) {
			s, err := adminTunnel.GetStats("listener.*_80.ssl.handshake")
			Expect(err).ToNot(HaveOccurred())
			Expect(s).To(stats.BeGreaterThanZero())
		}, "5s", "1s").Should(Succeed())
	})
}
