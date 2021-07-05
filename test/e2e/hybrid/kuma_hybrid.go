package hybrid

import (
	"fmt"
	"regexp"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config/core"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/kumahq/kuma/test/framework"
)

func KubernetesUniversalDeployment() {
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

	var global, zone1, zone2, zone3, zone4 Cluster
	var optsGlobal, optsZone1, optsZone2, optsZone3, optsZone4 = KumaUniversalDeployOpts, KumaZoneK8sDeployOpts, KumaZoneK8sDeployOpts, KumaUniversalDeployOpts, KumaUniversalDeployOpts

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

		// K8s Cluster 1
		zone1 = k8sClusters.GetCluster(Kuma1)
		optsZone1 = append(optsZone1,
			WithIngress(),
			WithGlobalAddress(globalCP.GetKDSServerAddress()),
			WithCNI(),
			WithEnv("KUMA_RUNTIME_KUBERNETES_INJECTOR_BUILTIN_DNS_ENABLED", "false")) // check if old resolving still works

		err = NewClusterSetup().
			Install(Kuma(core.Zone, optsZone1...)).
			Install(KumaDNS()).
			Install(YamlK8s(namespaceWithSidecarInjection(TestNamespace))).
			Install(DemoClientK8s(nonDefaultMesh)).
			Setup(zone1)
		Expect(err).ToNot(HaveOccurred())
		err = zone1.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		// K8s Cluster 2
		zone2 = k8sClusters.GetCluster(Kuma2)
		optsZone2 = append(optsZone2,
			WithIngress(),
			WithGlobalAddress(globalCP.GetKDSServerAddress()),
			WithEnv("KUMA_RUNTIME_KUBERNETES_INJECTOR_BUILTIN_DNS_ENABLED", "false"))

		err = NewClusterSetup().
			Install(Kuma(core.Zone, optsZone2...)).
			Install(KumaDNS()).
			Install(YamlK8s(namespaceWithSidecarInjection(TestNamespace))).
			Install(EchoServerK8s(nonDefaultMesh)).
			Install(DemoClientK8s(nonDefaultMesh)).
			Setup(zone2)
		Expect(err).ToNot(HaveOccurred())
		err = zone2.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		// Universal Cluster 3
		zone3 = universalClusters.GetCluster(Kuma3)
		optsZone3 = append(optsZone3,
			WithGlobalAddress(globalCP.GetKDSServerAddress()))
		ingressTokenKuma3, err := globalCP.GenerateZoneIngressToken(Kuma3)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(Kuma(core.Zone, optsZone3...)).
			Install(EchoServerUniversal(AppModeEchoServer, nonDefaultMesh, "universal", echoServerToken, WithTransparentProxy(true), WithBuiltinDNS(false))).
			Install(DemoClientUniversal(AppModeDemoClient, nonDefaultMesh, demoClientToken, WithTransparentProxy(true), WithBuiltinDNS(false))).
			Install(IngressUniversal(ingressTokenKuma3)).
			Setup(zone3)
		Expect(err).ToNot(HaveOccurred())
		err = zone3.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		// Universal Cluster 4
		zone4 = universalClusters.GetCluster(Kuma4)
		optsZone4 = append(optsZone4,
			WithGlobalAddress(globalCP.GetKDSServerAddress()))
		ingressTokenKuma4, err := globalCP.GenerateZoneIngressToken(Kuma4)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(Kuma(core.Zone, optsZone4...)).
			Install(DemoClientUniversal(AppModeDemoClient, nonDefaultMesh, demoClientToken)).
			Install(IngressUniversal(ingressTokenKuma4)).
			Setup(zone4)
		Expect(err).ToNot(HaveOccurred())
		err = zone4.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}
		err := zone1.DeleteNamespace(TestNamespace)
		Expect(err).ToNot(HaveOccurred())
		err = zone1.DeleteKuma(optsZone1...)
		Expect(err).ToNot(HaveOccurred())
		err = zone1.DismissCluster()
		Expect(err).ToNot(HaveOccurred())

		err = zone2.DeleteNamespace(TestNamespace)
		Expect(err).ToNot(HaveOccurred())
		err = zone2.DeleteKuma(optsZone2...)
		Expect(err).ToNot(HaveOccurred())
		err = zone2.DismissCluster()
		Expect(err).ToNot(HaveOccurred())

		err = zone3.DeleteKuma(optsZone3...)
		Expect(err).ToNot(HaveOccurred())
		err = zone3.DismissCluster()
		Expect(err).ToNot(HaveOccurred())

		err = zone4.DeleteKuma(optsZone4...)
		Expect(err).ToNot(HaveOccurred())
		err = zone4.DismissCluster()
		Expect(err).ToNot(HaveOccurred())

		err = global.DeleteKuma(optsGlobal...)
		Expect(err).ToNot(HaveOccurred())
		err = global.DismissCluster()
		Expect(err).ToNot(HaveOccurred())
	})

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

		// todo (lobkovilya): echo-server is restarting because of ncat problem, uncomment this as soon as it will be replaces with test-server
		// Eventually(func() error {
		//	output, err := global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "dataplanes", "--mesh", "non-default")
		//	if err != nil {
		//		return err
		//	}
		//
		//	re := regexp.MustCompile(`Online`)
		//	if len(re.FindAllString(output, -1)) != 6 {
		//		return errors.New("not all dataplanes are online")
		//	}
		//	return nil
		// }, "30s", "1s").ShouldNot(HaveOccurred())
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
			"curl", "-v", "-m", "3", "--fail", "echo-server_kuma-test_svc_80.mesh")
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
			"curl", "-v", "-m", "3", "--fail", "echo-server_kuma-test_svc_8080.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))

		// Zone 3
		// universal access remote k8s service
		stdout, _, err := zone3.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "-m", "3", "--fail", "echo-server_kuma-test_svc_8080.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))

		// Zone 4
		// universal access remote universal service
		stdout, _, err = zone4.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "-m", "3", "--fail", "localhost:4001")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
	})

	PIt("should support jobs with a sidecar", func() {
		// when deploy job that connects to a service on other K8S cluster
		err := DemoClientJobK8s(nonDefaultMesh, "echo-server_kuma-test_svc_80.mesh")(zone1)

		// then job is properly cleaned up and finished
		Expect(err).ToNot(HaveOccurred())

		// when deploy job that connects to a service on other Universal cluster
		err = DemoClientJobK8s(nonDefaultMesh, "echo-server_kuma-test_svc_8080.mesh")(zone2)

		// then job is properly cleaned up and finished
		Expect(err).ToNot(HaveOccurred())
	})
}
