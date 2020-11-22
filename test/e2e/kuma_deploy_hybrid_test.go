package e2e_test

import (
	"fmt"

	"github.com/kumahq/kuma/pkg/config/core"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/kumahq/kuma/test/framework"
)

var _ = Describe("Test Kubernetes/Universal deployment", func() {

	meshDefaulMtlsOn := `
type: Mesh
name: default
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
`

	trafficPermissionAllTo2Remote := `
type: TrafficPermission
name: all-to-2-remote
mesh: default
sources:
- match:
   kuma.io/service: "*"
destinations:
- match:
   kuma.io/service: "*"
   kuma.io/zone: kuma-2-remote
`

	namespaceWithSidecarInjection := func(namespace string) string {
		return fmt.Sprintf(`
apiVersion: v1
kind: Namespace
metadata:
  name: %s
  annotations:
    kuma.io/sidecar-injection: "enabled"
`, namespace)
	}

	var global, remote_1, remote_2, remote_3, remote_4 Cluster
	var optsGlobal, optsRemote1, optsRemote2, optsRemote3, optsRemote4 []DeployOptionsFunc

	BeforeEach(func() {
		k8sClusters, err := NewK8sClusters(
			[]string{Kuma1, Kuma2},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		universalClusters, err := NewUniversalClusters(
			[]string{Kuma3, Kuma4, Kuma5},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		// Global
		global = universalClusters.GetCluster(Kuma5)
		optsGlobal = []DeployOptionsFunc{}

		err = NewClusterSetup().
			Install(Kuma(core.Global, optsGlobal...)).
			Setup(global)
		Expect(err).ToNot(HaveOccurred())
		err = global.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		globalCP := global.GetKuma()

		echoServerToken, err := globalCP.GenerateDpToken("echo-server_kuma-test_svc_8080")
		Expect(err).ToNot(HaveOccurred())
		demoClientToken, err := globalCP.GenerateDpToken("demo-client")
		Expect(err).ToNot(HaveOccurred())
		ingressToken, err := globalCP.GenerateDpToken("ingress")
		Expect(err).ToNot(HaveOccurred())

		// K8s Cluster 1
		remote_1 = k8sClusters.GetCluster(Kuma1)
		optsRemote1 = []DeployOptionsFunc{
			WithIngress(),
			WithGlobalAddress(globalCP.GetKDSServerAddress()),
			WithCNI(),
		}

		err = NewClusterSetup().
			Install(Kuma(core.Remote, optsRemote1...)).
			Install(KumaDNS()).
			Install(YamlK8s(namespaceWithSidecarInjection(TestNamespace))).
			Install(DemoClientK8s()).
			Setup(remote_1)
		Expect(err).ToNot(HaveOccurred())
		err = remote_1.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		// K8s Cluster 2
		remote_2 = k8sClusters.GetCluster(Kuma2)
		optsRemote2 = []DeployOptionsFunc{
			WithIngress(),
			WithGlobalAddress(globalCP.GetKDSServerAddress()),
		}

		err = NewClusterSetup().
			Install(Kuma(core.Remote, optsRemote2...)).
			Install(KumaDNS()).
			Install(YamlK8s(namespaceWithSidecarInjection(TestNamespace))).
			Install(EchoServerK8s()).
			Install(DemoClientK8s()).
			Setup(remote_2)
		Expect(err).ToNot(HaveOccurred())
		err = remote_2.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		// Universal Cluster 3
		remote_3 = universalClusters.GetCluster(Kuma3)
		optsRemote3 = []DeployOptionsFunc{
			WithGlobalAddress(globalCP.GetKDSServerAddress()),
		}

		err = NewClusterSetup().
			Install(Kuma(core.Remote, optsRemote3...)).
			Install(EchoServerUniversal("universal", echoServerToken)).
			Install(DemoClientUniversal(demoClientToken)).
			Install(IngressUniversal(ingressToken)).
			Setup(remote_3)
		Expect(err).ToNot(HaveOccurred())
		err = remote_3.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		// Universal Cluster 4
		remote_4 = universalClusters.GetCluster(Kuma4)
		optsRemote4 = []DeployOptionsFunc{
			WithGlobalAddress(globalCP.GetKDSServerAddress()),
		}

		err = NewClusterSetup().
			Install(Kuma(core.Remote, optsRemote4...)).
			Install(DemoClientUniversal(demoClientToken)).
			Install(IngressUniversal(ingressToken)).
			Setup(remote_4)
		Expect(err).ToNot(HaveOccurred())
		err = remote_4.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		err = YamlUniversal(meshDefaulMtlsOn)(global)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}
		_ = k8s.KubectlDeleteFromStringE(remote_1.GetTesting(), remote_1.GetKubectlOptions(), namespaceWithSidecarInjection(TestNamespace))
		err := remote_1.DeleteKuma(optsRemote1...)
		Expect(err).ToNot(HaveOccurred())
		_ = k8s.KubectlDeleteFromStringE(remote_2.GetTesting(), remote_2.GetKubectlOptions(), namespaceWithSidecarInjection(TestNamespace))
		err = remote_2.DeleteKuma(optsRemote2...)
		Expect(err).ToNot(HaveOccurred())
		err = remote_3.DeleteKuma(optsRemote3...)
		Expect(err).ToNot(HaveOccurred())
		err = remote_4.DeleteKuma(optsRemote4...)
		Expect(err).ToNot(HaveOccurred())
		err = global.DeleteKuma(optsGlobal...)
		Expect(err).ToNot(HaveOccurred())

		err = remote_3.DismissCluster()
		Expect(err).ToNot(HaveOccurred())
		err = remote_4.DismissCluster()
		Expect(err).ToNot(HaveOccurred())
		err = global.DismissCluster()
		Expect(err).ToNot(HaveOccurred())
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
			"curl", "-v", "-m", "3", "--fail", "echo-server_kuma-test_svc_80.mesh")
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
			"curl", "-v", "-m", "3", "--fail", "echo-server_kuma-test_svc_8080.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))

		// Remote 3
		// universal access remote k8s service
		stdout, _, err := remote_3.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "-m", "3", "--fail", "localhost:4000")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))

		// Remote 4
		// universal access remote universal service
		stdout, _, err = remote_4.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "-m", "3", "--fail", "localhost:4001")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
	})

	It("should sync traffic permissions", func() {
		// Remote 4
		// universal access remote universal service
		Eventually(func() (string, error) {
			stdout, _, err := remote_4.ExecWithRetries("", "", "demo-client",
				"curl", "-v", "-m", "3", "--fail", "localhost:4001")
			return stdout, err
		}, "10s", "1s").Should(ContainSubstring("HTTP/1.1 200 OK"))

		err := global.GetKumactlOptions().KumactlDelete("traffic-permission", "allow-all-default") // remove builtin traffic permission
		Expect(err).ToNot(HaveOccurred())

		err = YamlUniversal(trafficPermissionAllTo2Remote)(global)
		Expect(err).ToNot(HaveOccurred())

		// Remote 3
		// universal access remote k8s service
		Eventually(func() (string, error) {
			stdout, _, err := remote_3.ExecWithRetries("", "", "demo-client",
				"curl", "-v", "-m", "3", "--fail", "localhost:4000")
			return stdout, err
		}, "10s", "1s").Should(ContainSubstring("HTTP/1.1 200 OK"))

		// Remote 4
		// universal can't access remote universal service
		Eventually(func() (string, error) {
			stdout, _, err := remote_4.ExecWithRetries("", "", "demo-client",
				"curl", "-v", "-m", "3", "localhost:4001")
			return stdout, err
		}, "10s", "1s").Should(ContainSubstring("HTTP/1.1 503 Service Unavailable"))
	})
})
