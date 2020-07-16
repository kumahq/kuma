package e2e_test

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/pkg/config/mode"

	. "github.com/kumahq/kuma/test/framework"
)

var _ = XDescribe("Test Kubernetes/Universal deployment", func() {

	meshDefaulMtlsOn := `
type: Mesh
name: default
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
`
	trafficPermissionAll := `
type: TrafficPermission
name: traffic-permission-all
mesh: default
sources:
- match:
   service: "*"
destinations:
- match:
   service: "*"
`

	namespaceWithSidecarInjection := func(namespace string) string {
		return fmt.Sprintf(`
apiVersion: v1
kind: Namespace
metadata:
  name: %s
  labels:
    kuma.io/sidecar-injection: "enabled"
`, namespace)
	}

	var global, remote_1, remote_2, remote_3, remote_4 Cluster

	BeforeEach(func() {
		k8sClusters, err := NewK8sClusters(
			[]string{Kuma1, Kuma2},
			Verbose)
		Expect(err).ToNot(HaveOccurred())

		universalClusters, err := NewUniversalClusters(
			[]string{Kuma3, Kuma4, Kuma5},
			Verbose)
		Expect(err).ToNot(HaveOccurred())

		// K8s Cluster 1
		remote_1 = k8sClusters.GetCluster(Kuma1)

		err = NewClusterSetup().
			Install(Kuma(mode.Remote)).
			Install(KumaDNS()).
			Install(Ingress(nil)).
			Install(YamlK8s(namespaceWithSidecarInjection(TestNamespace))).
			Install(DemoClientK8s()).
			Setup(remote_1)
		Expect(err).ToNot(HaveOccurred())
		err = remote_1.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		// K8s Cluster 2
		remote_2 = k8sClusters.GetCluster(Kuma2)

		err = NewClusterSetup().
			Install(Kuma(mode.Remote)).
			Install(KumaDNS()).
			Install(Ingress(nil)).
			Install(YamlK8s(namespaceWithSidecarInjection(TestNamespace))).
			Install(EchoServerK8s()).
			Install(DemoClientK8s()).
			Setup(remote_2)
		Expect(err).ToNot(HaveOccurred())
		err = remote_2.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		// Universal Cluster 3
		remote_3 = universalClusters.GetCluster(Kuma3)

		err = NewClusterSetup().
			Install(Kuma(mode.Remote)).
			Install(EchoServerUniversal()).
			Install(DemoClientUniversal()).
			Setup(remote_3)
		Expect(err).ToNot(HaveOccurred())
		err = remote_3.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		// Universal Cluster 4
		remote_4 = universalClusters.GetCluster(Kuma4)

		err = NewClusterSetup().
			Install(Kuma(mode.Remote)).
			Install(DemoClientUniversal()).
			Setup(remote_4)
		Expect(err).ToNot(HaveOccurred())
		err = remote_4.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		// Global
		global = universalClusters.GetCluster(Kuma5)

		err = NewClusterSetup().
			Install(Kuma(mode.Global)).
			Setup(global)
		Expect(err).ToNot(HaveOccurred())
		err = global.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		remote_1CP := remote_1.GetKuma()
		remote_2CP := remote_2.GetKuma()
		remote_3CP := remote_3.GetKuma()
		remote_4CP := remote_4.GetKuma()

		err = global.GetKumactlOptions().KumactlApplyFromString(
			fmt.Sprintf(ZoneTemplateUniversal,
				remote_1CP.GetName(),
				remote_1CP.GetKDSServerAddress(),
				remote_1CP.GetIngressAddress()))
		Expect(err).ToNot(HaveOccurred())

		err = global.GetKumactlOptions().KumactlApplyFromString(
			fmt.Sprintf(ZoneTemplateUniversal,
				remote_2CP.GetName(),
				remote_2CP.GetKDSServerAddress(),
				remote_2CP.GetIngressAddress()))
		Expect(err).ToNot(HaveOccurred())

		err = global.GetKumactlOptions().KumactlApplyFromString(
			fmt.Sprintf(ZoneTemplateUniversal,
				remote_3CP.GetName(),
				remote_3CP.GetKDSServerAddress(),
				remote_3CP.GetIngressAddress()))
		Expect(err).ToNot(HaveOccurred())

		err = global.GetKumactlOptions().KumactlApplyFromString(
			fmt.Sprintf(ZoneTemplateUniversal,
				remote_4CP.GetName(),
				remote_4CP.GetKDSServerAddress(),
				remote_4CP.GetIngressAddress()))
		Expect(err).ToNot(HaveOccurred())

		// remove these once Zones are added dynamically
		globalCP := global.GetKuma()
		err = globalCP.SetLbAddress(remote_1CP.GetName(), globalCP.GetKDSServerAddress())
		Expect(err).ToNot(HaveOccurred())

		err = global.RestartKuma()
		Expect(err).ToNot(HaveOccurred())

		err = YamlUniversal(meshDefaulMtlsOn)(global)
		Expect(err).ToNot(HaveOccurred())

		err = YamlUniversal(trafficPermissionAll)(global)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		_ = remote_1.DeleteKuma()
		_ = k8s.KubectlDeleteFromStringE(remote_1.GetTesting(), remote_1.GetKubectlOptions(), namespaceWithSidecarInjection(TestNamespace))
		_ = remote_2.DeleteKuma()
		_ = k8s.KubectlDeleteFromStringE(remote_2.GetTesting(), remote_2.GetKubectlOptions(), namespaceWithSidecarInjection(TestNamespace))
		_ = remote_3.DeleteKuma()
		_ = remote_4.DeleteKuma()
		_ = global.DeleteKuma()

		_ = remote_1.DismissCluster()
		_ = remote_2.DismissCluster()
		_ = remote_3.DismissCluster()
		_ = remote_4.DismissCluster()
		_ = global.DismissCluster()
	})

	It("should access allservices", func() {

		// Remote 1
		pods, err := k8s.ListPodsE(
			remote_1.GetTesting(),
			remote_1.GetKubectlOptions(TestNamespace),
			metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", "demo-client"),
			},
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(pods).To(HaveLen(1))

		clientPod := pods[0]

		// k8s access remote k8s service
		_, stderr, err := remote_1.ExecWithRetries(TestNamespace, clientPod.GetName(), "demo-client",
			"curl", "-v", "-m", "3", "echo-server_kuma-test_svc_80.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))

		// Remote 2
		pods, err = k8s.ListPodsE(
			remote_2.GetTesting(),
			remote_2.GetKubectlOptions(TestNamespace),
			metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", "demo-client"),
			},
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(pods).To(HaveLen(1))

		clientPod = pods[0]

		// k8s access remote universal service
		_, stderr, err = remote_2.ExecWithRetries(TestNamespace, clientPod.GetName(), "demo-client",
			"curl", "-v", "-m", "3", "echo-server_kuma-test_svc_8080.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))

		// Remote 3
		// universal access remote k8s service
		stdout, _, err := remote_3.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "-m", "3", "localhost:4000")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))

		// Remote 4
		// universal access remote universal service
		stdout, _, err = remote_4.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "-m", "3", "localhost:4001")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
	})
})
