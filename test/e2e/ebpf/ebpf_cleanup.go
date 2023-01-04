package ebpf

import (
	"fmt"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/random"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/test/e2e/ebpf/ebpf_checker"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func CleanupEbpfConfigFromNode() {
	namespace := "ebpf-cleanup"
	// IPv6 currently not supported by our eBPF
	// https://github.com/kumahq/kuma-net/issues/72
	if Config.IPV6 {
		fmt.Println("Test not supported on IPv6")
		return
	}

	var cluster Cluster
	releaseName := fmt.Sprintf(
		"kuma-%s",
		strings.ToLower(random.UniqueId()),
	)

	BeforeAll(func() {
		cluster = NewK8sCluster(NewTestingT(), Kuma1, Silent).
			WithTimeout(6 * time.Second).
			WithRetries(60)

		err := NewClusterSetup().
			Install(Kuma(core.Standalone,
				WithInstallationMode(HelmInstallationMode),
				WithHelmReleaseName(releaseName),
				WithHelmOpt("experimental.ebpf.enabled", "true"))).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh("default"),
				testserver.WithName("test-server"),
				// k3d has some problems with ebpf and sometimes mounting doesn't work
				// it might not start and because of if fails
				// ebpf files are created so we can still check if cleanup works
				testserver.WithoutWaitingToBeReady(),
			)).
			Install(ebpf_checker.Install(
				ebpf_checker.WithNamespace(namespace),
				ebpf_checker.WithoutSidecar(),
			)).Setup(cluster)

		Expect(err).ToNot(HaveOccurred())
	})

	AfterAll(func() {
		Expect(cluster.DeleteNamespace(namespace)).To(Succeed())
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	It("should cleanup ebpf files from node", func() {
		ebpfCheckerPodName, err := PodNameOfApp(cluster, "ebpf-checker", namespace)
		Expect(err).ToNot(HaveOccurred())

		// when remove application
		Expect(cluster.DeleteDeployment("test-server")).To(Succeed())

		// then should have bpf files left on the node
		stdout, _, err := cluster.Exec(namespace, ebpfCheckerPodName, "ebpf-checker", "ls", "/sys/fs/bpf")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("cookie_orig_dst"))

		// when kuma is deleted
		Expect(cluster.DeleteKuma()).To(Succeed())

		// then should not have ebpf files left
		Eventually(func(g Gomega) {
			stdout, _, err = cluster.Exec(namespace, ebpfCheckerPodName, "ebpf-checker", "ls", "/sys/fs/bpf")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).ToNot(ContainSubstring("cookie_orig_dst"))
		}, "30s", "1s").Should(Succeed())
	})
}
