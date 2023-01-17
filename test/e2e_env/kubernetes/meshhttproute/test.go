package meshhttproute

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func Test() {
	meshName := "meshhttproute"
	namespace := "meshhttproute"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshKubernetes(meshName)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(testserver.Install(
				testserver.WithName("test-client"),
				testserver.WithMesh(meshName),
				testserver.WithNamespace(namespace),
			)).
			Install(testserver.Install(
				testserver.WithName("test-server"),
				testserver.WithMesh(meshName),
				testserver.WithNamespace(namespace),
			)).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())

		Expect(
			k8s.RunKubectlE(kubernetes.Cluster.GetTesting(), kubernetes.Cluster.GetKubectlOptions(), "delete", "trafficroute", "route-all-meshhttproute"),
		).To(Succeed())
	})
	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(meshName)).To(Succeed())
	})
	It("should use MeshHTTPRoute if any MeshHTTPRoutes are present", func() {
		Eventually(func(g Gomega) {
			_, err := client.CollectResponse(kubernetes.Cluster, "test-client", "test-server_meshhttproute_svc_80.mesh", client.FromKubernetesPod(namespace, "test-client"))
			g.Expect(err).To(HaveOccurred())
		}, "30s", "1s").Should(Succeed())

		// when
		Expect(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: route-1
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: MeshService
    name: test-client_%s_svc_80
  to:
    - targetRef:
        kind: MeshService
        name: nonexistent-service-that-activates-default
      rules: []
`, Config.KumaNamespace, meshName, meshName))(kubernetes.Cluster)).To(Succeed())

		Eventually(func(g Gomega) {
			response, err := client.CollectResponse(kubernetes.Cluster, "test-client", "test-server_meshhttproute_svc_80.mesh", client.FromKubernetesPod(namespace, "test-client"))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(HavePrefix("test-server"))
		}, "30s", "1s").Should(Succeed())
	})
}
