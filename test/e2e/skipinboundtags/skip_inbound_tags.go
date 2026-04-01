package skipinboundtags

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
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

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Yaml(
				builders.Mesh().
					WithName(meshName).
					WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Exclusive),
			)).
			Install(MeshTrafficPermissionAllowAllKubernetes(meshName)).
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
			)).
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
			out, err := k8s.RunKubectlAndGetOutputE(
				KubeCluster.GetTesting(),
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
			out, err := k8s.RunKubectlAndGetOutputE(
				KubeCluster.GetTesting(),
				KubeCluster.GetKubectlOptions(namespace),
				"get", "meshservices", "-n", namespace,
				"-l", fmt.Sprintf("kuma.io/mesh=%s", meshName),
				"-o", "jsonpath={.items[*].spec.selector.dataplaneTags}",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(out).To(BeEmpty(), "MeshService should not have dataplaneTags selector")
		}, "60s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			out, err := k8s.RunKubectlAndGetOutputE(
				KubeCluster.GetTesting(),
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
}
