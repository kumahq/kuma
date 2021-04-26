package hybrid_test

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
	if IsApiV2() {
		fmt.Println("Test not supported on API v2")
		return
	}
	meshMTLSOn := func(mesh string) string {
		return fmt.Sprintf(`
type: Mesh
name: %s
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
`, mesh)
	}

	trafficPermissionAllTo2Remote := func(mesh string) string {
		return fmt.Sprintf(`
type: TrafficPermission
name: all-to-2-remote
mesh: %s
sources:
- match:
   kuma.io/service: "*"
destinations:
- match:
   kuma.io/service: "*"
   kuma.io/zone: kuma-2-remote
`, mesh)
	}

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

	const nonDefaultMesh = "non-default"
	const defaultMesh = "default"

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
			Install(YamlUniversal(meshMTLSOn(nonDefaultMesh))).
			Install(YamlUniversal(meshMTLSOn(defaultMesh))).
			Setup(global)
		Expect(err).ToNot(HaveOccurred())
		err = global.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		globalCP := global.GetKuma()

		echoServerToken, err := globalCP.GenerateDpToken(nonDefaultMesh, "echo-server_kuma-test_svc_8080")
		Expect(err).ToNot(HaveOccurred())
		demoClientToken, err := globalCP.GenerateDpToken(nonDefaultMesh, "demo-client")
		Expect(err).ToNot(HaveOccurred())
		ingressToken, err := globalCP.GenerateDpToken(defaultMesh, "ingress")
		Expect(err).ToNot(HaveOccurred())

		// K8s Cluster 1
		remote_1 = k8sClusters.GetCluster(Kuma1)
		optsRemote1 = []DeployOptionsFunc{
			WithIngress(),
			WithGlobalAddress(globalCP.GetKDSServerAddress()),
			WithCNI(),
			WithEnv("KUMA_RUNTIME_KUBERNETES_INJECTOR_BUILTIN_DNS_ENABLED", "true"),
		}

		err = NewClusterSetup().
			Install(Kuma(core.Remote, optsRemote1...)).
			Install(KumaDNS()).
			Install(YamlK8s(namespaceWithSidecarInjection(TestNamespace))).
			Install(DemoClientK8s(nonDefaultMesh)).
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
			Install(EchoServerK8s(nonDefaultMesh)).
			Install(DemoClientK8s(nonDefaultMesh)).
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
			Install(EchoServerUniversal(AppModeEchoServer, nonDefaultMesh, "universal", echoServerToken, WithTransparentProxy(true))).
			Install(DemoClientUniversal(AppModeDemoClient, nonDefaultMesh, demoClientToken, WithTransparentProxy(true))).
			Install(IngressUniversal(defaultMesh, ingressToken)).
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
			Install(DemoClientUniversal(AppModeDemoClient, nonDefaultMesh, demoClientToken)).
			Install(IngressUniversal(defaultMesh, ingressToken)).
			Setup(remote_4)
		Expect(err).ToNot(HaveOccurred())
		err = remote_4.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}
		err := remote_1.DeleteNamespace(TestNamespace)
		Expect(err).ToNot(HaveOccurred())
		err = remote_1.DeleteKuma(optsRemote1...)
		Expect(err).ToNot(HaveOccurred())
		err = remote_1.DismissCluster()
		Expect(err).ToNot(HaveOccurred())

		err = remote_2.DeleteNamespace(TestNamespace)
		Expect(err).ToNot(HaveOccurred())
		err = remote_2.DeleteKuma(optsRemote2...)
		Expect(err).ToNot(HaveOccurred())
		err = remote_2.DismissCluster()
		Expect(err).ToNot(HaveOccurred())

		err = remote_3.DeleteKuma(optsRemote3...)
		Expect(err).ToNot(HaveOccurred())
		err = remote_3.DismissCluster()
		Expect(err).ToNot(HaveOccurred())

		err = remote_4.DeleteKuma(optsRemote4...)
		Expect(err).ToNot(HaveOccurred())
		err = remote_4.DismissCluster()
		Expect(err).ToNot(HaveOccurred())

		err = global.DeleteKuma(optsGlobal...)
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
			"curl", "-v", "-m", "3", "--fail", "echo-server_kuma-test_svc_8080.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))

		// Remote 4
		// universal access remote universal service
		stdout, _, err = remote_4.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "-m", "3", "--fail", "localhost:4001")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))

		// Remote 1
		// check for job support
		// k8s access remote k8s service
		err = DemoClientJobK8s(nonDefaultMesh, "echo-server_kuma-test_svc_80.mesh")(remote_1)
		Expect(err).ToNot(HaveOccurred())

		// Remote 2
		// k8s access remote universal service
		err = DemoClientJobK8s(nonDefaultMesh, "echo-server_kuma-test_svc_8080.mesh")(remote_2)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should sync traffic permissions", func() {
		// Remote 4
		// universal access remote universal service
		Eventually(func() (string, error) {
			stdout, _, err := remote_4.ExecWithRetries("", "", "demo-client",
				"curl", "-v", "-m", "3", "--fail", "localhost:4001")
			return stdout, err
		}, "10s", "1s").Should(ContainSubstring("HTTP/1.1 200 OK"))

		err := global.GetKumactlOptions().KumactlDelete("traffic-permission", "allow-all-non-default", nonDefaultMesh) // remove builtin traffic permission
		Expect(err).ToNot(HaveOccurred())

		err = YamlUniversal(trafficPermissionAllTo2Remote(nonDefaultMesh))(global)
		Expect(err).ToNot(HaveOccurred())

		// Remote 3
		// universal access remote k8s service
		Eventually(func() (string, error) {
			stdout, _, err := remote_3.ExecWithRetries("", "", "demo-client",
				"curl", "-v", "-m", "3", "--fail", "echo-server_kuma-test_svc_8080.mesh")
			return stdout, err
		}, "10s", "1s").Should(ContainSubstring("HTTP/1.1 200 OK"))

		// Remote 4
		// universal can't access remote universal service
		Eventually(func() (string, error) {
			stdout, _, err := remote_4.ExecWithRetries("", "", "demo-client",
				"curl", "-v", "-m", "3", "localhost:4001")
			return stdout, err
		}, "10s", "1s").Should(ContainSubstring("HTTP/1.1 503 Service Unavailable"))

		// Remote 1
		// check for failing job support
		// k8s can not access remote k8s service
		err = DemoClientJobK8s(nonDefaultMesh, "echo-server_kuma-test_svc_8080.mesh")(remote_1)
		Expect(err).To(HaveOccurred())
	})

})
