package helm

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/random"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func AppDeploymentWithHelmChart() {
	var cluster Cluster
	var skip bool

	minReplicas := 3

	var setup = func(withCni KumaDeploymentOption, shouldSkip func(cluster Cluster) (string, bool)) {
		cluster = NewK8sCluster(NewTestingT(), Kuma1, Silent).
			WithTimeout(6 * time.Second).
			WithRetries(60)

		if shouldSkip != nil {
			var msg string
			if msg, skip = shouldSkip(cluster); skip {
				Skip(msg)
			}
		}

		err := NewClusterSetup().
			Install(Kuma(core.Standalone,
				WithInstallationMode(HelmInstallationMode),
				WithHelmReleaseName(fmt.Sprintf("kuma-%s", strings.ToLower(random.UniqueId()))),
				WithSkipDefaultMesh(true), // it's common case for HELM deployments that Mesh is also managed by HELM therefore it's not created by default
				WithHelmOpt("controlPlane.autoscaling.enabled", "true"),
				WithHelmOpt("controlPlane.autoscaling.minReplicas", strconv.Itoa(minReplicas)),
				withCni,
			)).
			Install(MeshKubernetes("default")).
			Install(NamespaceWithSidecarInjection(TestNamespace)).
			Install(DemoClientK8s("default", TestNamespace)).
			Install(testserver.Install()).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())

		Expect(cluster.(*K8sCluster).WaitApp(Config.KumaServiceName, Config.KumaNamespace, minReplicas)).To(Succeed())
	}

	E2EAfterEach(func() {
		if !skip {
			Expect(cluster.DeleteNamespace(TestNamespace)).To(Succeed())
			Expect(cluster.DeleteKuma()).To(Succeed())
		}

		Expect(cluster.DismissCluster()).To(Succeed())
	})

	DescribeTable(
		"Should deploy two apps",
		func(withCni KumaDeploymentOption, shouldSkip func(cluster Cluster) (string, bool)) {
			setup(withCni, shouldSkip)

			clientPodName, err := PodNameOfApp(cluster, "demo-client", TestNamespace)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func(g Gomega) {
				_, stderr, err := cluster.Exec(TestNamespace, clientPodName, "demo-client",
					"curl", "-v", "-m", "3", "--fail", "test-server")
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))
			}).Should(Succeed())

			Eventually(func(g Gomega) {
				_, stderr, err := cluster.Exec(TestNamespace, clientPodName, "demo-client",
					"curl", "-v", "-m", "3", "--fail", "test-server_kuma-test_svc_80.mesh")
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))
			}).Should(Succeed())

			Eventually(func(g Gomega) {
				_, stderr, err := cluster.Exec(TestNamespace, clientPodName, "demo-client",
					"curl", "-v", "-m", "3", "--fail", "test-server.kuma-test.svc.80.mesh")
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))
			}).Should(Succeed())
		},
		// Entry("with default cni", WithCNI(), func(cluster Cluster) (string, bool) {
		// 	version, err := cluster.GetK8sVersion()
		// 	Expect(err).To(Succeed())
		//
		// 	// Just future proofing
		// 	if version.Major != 1 {
		// 		return fmt.Sprintf(
		// 			"default cni is not supported in version %d.%d.%d [supported <= 1.21.x]",
		// 			version.Major, version.Minor, version.Patch,
		// 		), true
		// 	}
		//
		// 	// k3s from version 1.22 comes with flannel CNI plugin in version 1,
		// 	// which is not supported with our default/legacy kuma-cni plugin
		// 	// (max supported version is 0.4)
		// 	if version.Minor > 21 {
		// 		return fmt.Sprintf(
		// 			"default cni is not supported in version 1.%d.%d [supported <= 1.21.x]",
		// 			version.Minor, version.Patch,
		// 		), true
		// 	}
		//
		// 	return "", false
		// }),
		Entry("with new cni (experimental)", WithExperimentalCNI(), nil),
	)
}
