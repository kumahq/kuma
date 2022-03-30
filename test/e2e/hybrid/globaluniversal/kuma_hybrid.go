package globaluniversal

import (
	"fmt"
	"regexp"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func meshMTLSOn(mesh string) string {
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

var global, zone1, zone2, zone3, zone4 Cluster

const nonDefaultMesh = "non-default"
const defaultMesh = "default"

var _ = E2EBeforeSuite(func() {
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

	Expect(NewClusterSetup().
		Install(Kuma(core.Global)).
		Install(YamlUniversal(meshMTLSOn(nonDefaultMesh))).
		Install(YamlUniversal(meshMTLSOn(defaultMesh))).
		Setup(global)).To(Succeed())

	E2EDeferCleanup(global.DismissCluster)

	globalCP := global.GetKuma()

	echoServerToken, err := globalCP.GenerateDpToken(nonDefaultMesh, "test-server")
	Expect(err).ToNot(HaveOccurred())
	demoClientToken, err := globalCP.GenerateDpToken(nonDefaultMesh, "demo-client")
	Expect(err).ToNot(HaveOccurred())

	// K8s Cluster 1
	zone1 = k8sClusters.GetCluster(Kuma1)
	Expect(NewClusterSetup().
		Install(Kuma(core.Zone,
			WithIngress(),
			WithGlobalAddress(globalCP.GetKDSServerAddress()),
			WithCNI(),
			WithEnv("KUMA_RUNTIME_KUBERNETES_INJECTOR_BUILTIN_DNS_ENABLED", "false"), // check if old resolving still works
		)).
		Install(KumaDNS()).
		Install(NamespaceWithSidecarInjection(TestNamespace)).
		Install(DemoClientK8s(nonDefaultMesh)).
		Setup(zone1)).To(Succeed())

	E2EDeferCleanup(func() {
		Expect(zone1.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(zone1.DeleteKuma()).To(Succeed())
		Expect(zone1.DismissCluster()).To(Succeed())
	})

	// K8s Cluster 2
	zone2 = k8sClusters.GetCluster(Kuma2)

	Expect(NewClusterSetup().
		Install(Kuma(core.Zone,
			WithIngress(),
			WithGlobalAddress(globalCP.GetKDSServerAddress()),
			WithEnv("KUMA_RUNTIME_KUBERNETES_INJECTOR_BUILTIN_DNS_ENABLED", "false"),
		)).
		Install(KumaDNS()).
		Install(NamespaceWithSidecarInjection(TestNamespace)).
		Install(testserver.Install(testserver.WithMesh(nonDefaultMesh), testserver.WithServiceAccount("sa-test"))).
		Install(DemoClientK8s(nonDefaultMesh)).
		Setup(zone2)).To(Succeed())

	E2EDeferCleanup(func() {
		Expect(zone2.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(zone2.DeleteKuma()).To(Succeed())
		Expect(zone2.DismissCluster()).To(Succeed())
	})

	// Universal Cluster 3
	zone3 = universalClusters.GetCluster(Kuma3)
	ingressTokenKuma3, err := globalCP.GenerateZoneIngressToken(Kuma3)
	Expect(err).ToNot(HaveOccurred())

	Expect(NewClusterSetup().
		Install(Kuma(core.Zone, WithGlobalAddress(globalCP.GetKDSServerAddress()))).
		Install(TestServerUniversal("dp-echo", nonDefaultMesh, echoServerToken,
			WithArgs([]string{"echo", "--instance", "echo-v1"}),
			WithServiceName("test-server"),
		)).
		Install(DemoClientUniversal(AppModeDemoClient, nonDefaultMesh, demoClientToken, WithTransparentProxy(true), WithBuiltinDNS(false))).
		Install(IngressUniversal(ingressTokenKuma3)).
		Setup(zone3)).To(Succeed())

	E2EDeferCleanup(zone3.DismissCluster)

	// Universal Cluster 4
	zone4 = universalClusters.GetCluster(Kuma4)
	ingressTokenKuma4, err := globalCP.GenerateZoneIngressToken(Kuma4)
	Expect(err).ToNot(HaveOccurred())

	Expect(NewClusterSetup().
		Install(Kuma(core.Zone, WithGlobalAddress(globalCP.GetKDSServerAddress()))).
		Install(DemoClientUniversal(AppModeDemoClient, nonDefaultMesh, demoClientToken, WithTransparentProxy(true))).
		Install(IngressUniversal(ingressTokenKuma4)).
		Setup(zone4)).To(Succeed())

	E2EDeferCleanup(zone4.DismissCluster)
})

func KubernetesUniversalDeployment() {
	It("should correctly synchronize Dataplanes and ZoneIngresses and their statuses", func() {
		Eventually(func() error {
			output, err := global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zone-ingresses")
			if err != nil {
				return err
			}

			re := regexp.MustCompile(`Online`)
			if len(re.FindAllString(output, -1)) != 4 {
				return errors.New("not all zone-ingresses are online")
			}
			return nil
		}, "30s", "1s").ShouldNot(HaveOccurred())

		Eventually(func() error {
			output, err := global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "dataplanes", "--mesh", "non-default")
			if err != nil {
				return err
			}

			re := regexp.MustCompile(`Online`)
			if len(re.FindAllString(output, -1)) != 6 {
				return errors.New("not all dataplanes are online")
			}
			return nil
		}, "30s", "1s").ShouldNot(HaveOccurred())
	})

	It("should access allservices", func() {
		// Zone 1
		pods, err := k8s.ListPodsE(
			zone1.GetTesting(),
			zone1.GetKubectlOptions(TestNamespace),
			metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", "demo-client"),
			},
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(pods).To(HaveLen(1))

		clientPod := pods[0]

		// k8s access remote k8s service
		_, stderr, err := zone1.ExecWithRetries(TestNamespace, clientPod.GetName(), "demo-client",
			"curl", "-v", "-m", "3", "--fail", "test-server_kuma-test_svc_80.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))

		// Zone 2
		pods, err = k8s.ListPodsE(
			zone2.GetTesting(),
			zone2.GetKubectlOptions(TestNamespace),
			metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", "demo-client"),
			},
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(pods).To(HaveLen(1))

		clientPod = pods[0]

		// k8s access remote universal service
		_, stderr, err = zone2.ExecWithRetries(TestNamespace, clientPod.GetName(), "demo-client",
			"curl", "-v", "-m", "3", "--fail", "test-server.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))

		// Zone 3
		// universal access remote k8s service
		stdout, _, err := zone3.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "-m", "3", "--fail", "test-server_kuma-test_svc_80.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))

		// Zone 4
		// universal access remote universal service
		stdout, _, err = zone4.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "-m", "3", "--fail", "test-server.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
	})
}
