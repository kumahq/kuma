package watchnamespaces

import (
	"fmt"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/random"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func WatchOnlyDefinedNamespaces() {
	watchedNamespace := "watched-namespace"
	anotherWatchedNamespace := "another-watched-namespace"
	notWatchedNamespace := "not-watched-namespace"
	meshName := "watched"
	var cluster Cluster

	E2EAfterAll(func() {
		Expect(cluster.DeleteNamespace(watchedNamespace)).To(Succeed())
		Expect(cluster.DeleteNamespace(anotherWatchedNamespace)).To(Succeed())
		Expect(cluster.DeleteNamespace(notWatchedNamespace)).To(Succeed())
		Expect(cluster.DeleteKuma()).To(Succeed())
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	BeforeAll(func() {
		cluster = NewK8sCluster(NewTestingT(), Kuma1, Silent).
			WithTimeout(6 * time.Second).
			WithRetries(60)

		releaseName := fmt.Sprintf(
			"kuma-%s",
			strings.ToLower(random.UniqueId()),
		)

		Expect(NewClusterSetup().
			Install(NamespaceWithSidecarInjection(watchedNamespace)).
			Install(NamespaceWithSidecarInjection(anotherWatchedNamespace)).
			Install(NamespaceWithSidecarInjection(notWatchedNamespace)).
			Install(Kuma(core.Zone,
				WithInstallationMode(HelmInstallationMode),
				WithHelmReleaseName(releaseName),
				WithHelmOpt("watchNamespaces[0]", watchedNamespace),
				WithHelmOpt("watchNamespaces[1]", anotherWatchedNamespace),
			)).
			Setup(cluster)).To(Succeed())

		Expect(NewClusterSetup().
			Install(Yaml(
				builders.Mesh().
					WithName(meshName).
					WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Exclusive),
			)).
			Install(testserver.Install(
				testserver.WithName("test-server-watched"),
				testserver.WithMesh(meshName),
				testserver.WithNamespace(watchedNamespace),
				testserver.WithEchoArgs("echo", "--instance", "test-server-watched"),
			)).
			Install(testserver.Install(
				testserver.WithName("test-server-another-watched"),
				testserver.WithMesh(meshName),
				testserver.WithNamespace(anotherWatchedNamespace),
				testserver.WithEchoArgs("echo", "--instance", "test-server-another-watched"),
			)).
			Install(testserver.Install(
				testserver.WithName("not-watched"),
				testserver.WithMesh(meshName),
				testserver.WithNamespace(notWatchedNamespace),
				testserver.WithEchoArgs("echo", "--instance", "not-watched"),
			)).
			Install(democlient.Install(
				democlient.WithMesh(meshName),
				democlient.WithNamespace(watchedNamespace),
				democlient.WithName("demo-client-watched"),
			)).
			Setup(cluster)).To(Succeed())
	})

	It("should not connect to non auto reachable service", func() {
		// should access service in the same namespace
		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				cluster,
				"demo-client-watched",
				fmt.Sprintf("test-server-watched.%s.svc.cluster.local", watchedNamespace),
				client.FromKubernetesPod(watchedNamespace, "demo-client-watched"),
			)
			g.Expect(err).To(Not(HaveOccurred()))
			g.Expect(resp.Instance).To(Equal("test-server-watched"))
		}, "30s", "1s").Should(Succeed())

		// should access service in another namespace
		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				cluster,
				"demo-client-watched",
				fmt.Sprintf("test-server-another-watched.%s.svc.cluster.local", anotherWatchedNamespace),
				client.FromKubernetesPod(watchedNamespace, "demo-client-watched"),
			)
			g.Expect(err).To(Not(HaveOccurred()))
			g.Expect(resp.Instance).To(Equal("test-server-another-watched"))
		}, "30s", "1s").Should(Succeed())

		// cannot find resource in not watched namespace
		Eventually(func(g Gomega) {
			stdout, err := cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes", "--mesh", meshName)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(Not(ContainSubstring("not-watched_not-watched-namespace_svc_80")))
		}, "30s", "1s").Should(Succeed())

		Expect(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: tcp-timeout-in-watched
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  to:
  - targetRef:
      kind: Mesh
    default:
      idleTimeout: 20s
      connectionTimeout: 2s
`, watchedNamespace, meshName))(cluster)).To(Succeed())

		// can find resource in watched namespace
		Eventually(func(g Gomega) {
			stdout, err := cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "meshtimeouts", "--mesh", meshName)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("tcp-timeout-in-watched"))
		}, "30s", "1s").Should(Succeed())

		// can find default resource in system namespace
		Eventually(func(g Gomega) {
			stdout, err := cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "meshtimeouts", "--mesh", meshName)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring(fmt.Sprintf("mesh-timeout-all-%s", meshName)))
		}, "30s", "1s").Should(Succeed())

		Expect(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: tcp-timeout-not-watched
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  to:
  - targetRef:
      kind: Mesh
    default:
      idleTimeout: 20s
      connectionTimeout: 2s
`, notWatchedNamespace, meshName))(cluster)).To(Succeed())

		// cannot find resource from not watched namespace
		Eventually(func(g Gomega) {
			stdout, err := cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "meshtimeouts", "--mesh", meshName)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).ToNot(ContainSubstring("tcp-timeout-not-watched"))
		}, "30s", "1s").Should(Succeed())
	})
}
