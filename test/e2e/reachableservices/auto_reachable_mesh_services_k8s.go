package reachableservices

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func AutoReachableMeshServices() {
	var k8sCluster Cluster
	meshName := "auto-reachable-backends"
	namespace := "auto-reachable-backends"

	hostnameGenerator := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: HostnameGenerator
metadata:
  labels:
    kuma.io/mesh: %s
  name: mes-hg
  namespace: %s
spec:
  selector:
    meshService:
      matchLabels:
        k8s.kuma.io/namespace: %s
  template: "{{ .DisplayName }}.mesh"
`, meshName, Config.KumaNamespace, namespace)

	BeforeAll(func() {
		k8sCluster = NewK8sCluster(NewTestingT(), Kuma1, Silent)

		err := NewClusterSetup().
			Install(Kuma(config_core.Zone,
				WithEnv("KUMA_EXPERIMENTAL_AUTO_REACHABLE_SERVICES", "true"),
			)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(
				Yaml(
					builders.Mesh().
						WithName(meshName).
						WithBuiltinMTLSBackend("ca-1").WithEnabledMTLSBackend("ca-1").
						WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Exclusive),
				),
			).
			Install(testserver.Install(testserver.WithName("client-server"), testserver.WithMesh(meshName), testserver.WithNamespace(namespace))).
			Install(testserver.Install(testserver.WithName("first-test-server"), testserver.WithMesh(meshName), testserver.WithNamespace(namespace))).
			Install(testserver.Install(testserver.WithName("second-test-server"), testserver.WithMesh(meshName), testserver.WithNamespace(namespace))).
			Install(YamlK8s(hostnameGenerator)).
			Setup(k8sCluster)

		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterEach(func() {
		Expect(DeleteMeshResources(k8sCluster, meshName, v1alpha1.MeshTrafficPermissionResourceTypeDescriptor)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(k8sCluster.DeleteNamespace(namespace)).To(Succeed())
		Expect(k8sCluster.DeleteKuma()).To(Succeed())
		Expect(k8sCluster.DismissCluster()).To(Succeed())
	})

	It("should not connect to non auto reachable service", func() {
		// when
		Expect(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTrafficPermission
metadata:
  name: mtp1
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: MeshSubset
    tags:
      app: first-test-server
  from:
    - targetRef:
        kind: Mesh
      default:
        action: Deny
    - targetRef:
        kind: MeshSubset
        tags:
          app: client-server
      default:
        action: Allow
`, Config.KumaNamespace, meshName))(k8sCluster)).To(Succeed())

		// then
		Eventually(func(g Gomega) {
			pod, err := PodNameOfApp(k8sCluster, "second-test-server", namespace)
			g.Expect(err).ToNot(HaveOccurred())
			stdout, err := k8sCluster.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "dataplane", pod+"."+namespace, "--type=clusters", fmt.Sprintf("--mesh=%s", meshName))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(Not(ContainSubstring(fmt.Sprintf("%s_first-test-server_%s_defaul_msvc_80", meshName, namespace))))
		}, "30s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			pod, err := PodNameOfApp(k8sCluster, "client-server", namespace)
			g.Expect(err).ToNot(HaveOccurred())
			stdout, err := k8sCluster.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "dataplane", pod+"."+namespace, "--type=clusters", fmt.Sprintf("--mesh=%s", meshName))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring(fmt.Sprintf("%s_first-test-server_%s_default_msvc_80", meshName, namespace)))
		}, "30s", "1s").Should(Succeed())

		Consistently(func(g Gomega) {
			failures, err := client.CollectFailure(
				k8sCluster,
				"second-test-server",
				"first-test-server.mesh",
				client.FromKubernetesPod(namespace, "second-test-server"),
			)
			g.Expect(err).To(Not(HaveOccurred()))
			g.Expect(failures.Exitcode).To(Equal(6))
		}, "5s", "1s").Should(Succeed())
	})
}
