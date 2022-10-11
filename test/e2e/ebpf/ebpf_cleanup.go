package ebpf

import (
	"fmt"
	"strings"

	"github.com/gruntwork-io/terratest/modules/random"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/test/e2e/ebpf/ebpf_checker"
	"github.com/kumahq/kuma/test/e2e/ebpf/ebpf_cleaner"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func CleanupEbpfConfigFromNode() {
	cleanupNamespace := "cleanup"
	meshName := "cleanup-ebpf"
	defaultMesh := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: %s
`, meshName)

	var cluster Cluster

	BeforeEach(func() {
		clusters, err := NewK8sClusters(
			[]string{Kuma1},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		releaseName := fmt.Sprintf(
			"kuma-%s",
			strings.ToLower(random.UniqueId()),
		)

		cluster = clusters.GetCluster(Kuma1)
		Expect(NewClusterSetup().
			Install(Kuma(core.Standalone,
				WithInstallationMode(HelmInstallationMode),
				WithHelmReleaseName(releaseName),
				WithSkipDefaultMesh(true), // it's common case for HELM deployments that Mesh is also managed by HELM therefore it's not created by default
				WithHelmOpt("experimental.ebpf.enabled", "true"))).
			Install(NamespaceWithSidecarInjection(TestNamespace)).
			Install(Namespace(cleanupNamespace)).
			Setup(cluster)).To(Succeed())
		Expect(YamlK8s(defaultMesh)(cluster)).To(Succeed())
		Expect(NewClusterSetup().
			Install(testserver.Install(
				testserver.WithNamespace(TestNamespace),
				testserver.WithMesh(meshName),
				testserver.WithName("test-server"),
			)).
			Install(ebpf_checker.Install(
				ebpf_checker.WithNamespace(cleanupNamespace),
			)).
			Setup(cluster)).To(Succeed())
	})

	E2EAfterEach(func() {
		Expect(cluster.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(cluster.DeleteNamespace(cleanupNamespace)).To(Succeed())
		Expect(cluster.DeleteKuma()).To(Succeed())
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	It("should cleanup ebpf files from node", func() {
		ebpfCheckerPodName, err := PodNameOfApp(cluster, "ebpf-checker", cleanupNamespace)
		Expect(err).ToNot(HaveOccurred())

		// when remove application
		Expect(cluster.DeleteDeployment("test-server")).To(Succeed())

		// then should have bpf files left on the node
		stdout, _, err := cluster.Exec(cleanupNamespace, ebpfCheckerPodName, "ebpf-checker", "ls", "/sys/fs/bpf")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("netns_cleanup_link"))

		// when run uninstall ebpf program
		Expect(NewClusterSetup().
			Install(ebpf_cleaner.Install(ebpf_cleaner.WithNamespace(cleanupNamespace))).
			Setup(cluster)).To(Succeed())

		// then should not have ebpf files left
		Eventually(func() bool {
			stdout, _, err = cluster.Exec(cleanupNamespace, ebpfCheckerPodName, "ebpf-checker", "ls", "/sys/fs/bpf")
			if err != nil {
				return false
			}
			return !strings.Contains(stdout, "netns_cleanup_link")
		}, "30s", "1s").Should(BeTrue())
	})
}
