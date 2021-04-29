package hybrid

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	config_core "github.com/kumahq/kuma/pkg/config/core"

	. "github.com/kumahq/kuma/test/framework"
)

func TrafficPermissionHybrid() {
	var globalCluster, remoteUniversal, remoteKube Cluster
	var optsGlobal, optsRemoteUniversal, optsRemoteKube []DeployOptionsFunc
	var clientPodName string

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

	meshMTLSOn := func(mesh string) string {
		return fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: %s
spec:
  mtls:
    enabledBackend: ca-1
    backends:
    - name: ca-1
      type: builtin
`, mesh)
	}

	E2EBeforeSuite(func() {
		k8sClusters, err := NewK8sClusters(
			[]string{Kuma1, Kuma2},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		universalClusters, err := NewUniversalClusters(
			[]string{Kuma3},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		// Global
		globalCluster = k8sClusters.GetCluster(Kuma1)
		optsGlobal = []DeployOptionsFunc{}

		err = NewClusterSetup().
			Install(Kuma(config_core.Global, optsGlobal...)).
			Install(YamlK8s(meshMTLSOn("default"))).
			Setup(globalCluster)
		Expect(err).ToNot(HaveOccurred())
		err = globalCluster.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())
		globalCP := globalCluster.GetKuma()

		echoServerToken, err := globalCP.GenerateDpToken("default", "echo-server_kuma-test_svc_8080")
		Expect(err).ToNot(HaveOccurred())
		ingressToken, err := globalCP.GenerateDpToken("default", "ingress")
		Expect(err).ToNot(HaveOccurred())

		// Remote universal
		remoteUniversal = universalClusters.GetCluster(Kuma3)
		optsRemoteUniversal = []DeployOptionsFunc{
			WithGlobalAddress(globalCP.GetKDSServerAddress()),
		}

		err = NewClusterSetup().
			Install(Kuma(config_core.Remote, optsRemoteUniversal...)).
			Install(EchoServerUniversal(AppModeEchoServer, "default", "universal", echoServerToken)).
			Install(IngressUniversal("default", ingressToken)).
			Setup(remoteUniversal)
		Expect(err).ToNot(HaveOccurred())
		err = remoteUniversal.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		// Remote kubernetes
		remoteKube = k8sClusters.GetCluster(Kuma2)
		optsRemoteKube = []DeployOptionsFunc{
			WithGlobalAddress(globalCP.GetKDSServerAddress()),
		}

		err = NewClusterSetup().
			Install(Kuma(config_core.Remote, optsRemoteKube...)).
			Install(KumaDNS()).
			Install(YamlK8s(namespaceWithSidecarInjection(TestNamespace))).
			Install(DemoClientK8s("default")).
			Setup(remoteKube)
		Expect(err).ToNot(HaveOccurred())
		err = remoteKube.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		pods, err := k8s.ListPodsE(
			remoteKube.GetTesting(),
			remoteKube.GetKubectlOptions(TestNamespace),
			metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", "demo-client"),
			},
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(pods).To(HaveLen(1))

		clientPodName = pods[0].GetName()
	})

	E2EAfterEach(func() {
		// remove all TrafficPermissions and restore the default
		err := k8s.RunKubectlE(globalCluster.GetTesting(), globalCluster.GetKubectlOptions(), "delete", "trafficpermissions", "--all")
		Expect(err).ToNot(HaveOccurred())

		err = k8s.KubectlApplyFromStringE(globalCluster.GetTesting(), globalCluster.GetKubectlOptions(), `
apiVersion: kuma.io/v1alpha1
kind: TrafficPermission
mesh: default
metadata:
  name: allow-all-default
spec:
  sources:
    - match:
        kuma.io/service: '*'
  destinations:
    - match:
        kuma.io/service: '*'

`)

		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterSuite(func() {
		err := globalCluster.DeleteKuma(optsGlobal...)
		Expect(err).ToNot(HaveOccurred())
		err = globalCluster.DismissCluster()
		Expect(err).ToNot(HaveOccurred())

		err = remoteUniversal.DeleteKuma(optsRemoteUniversal...)
		Expect(err).ToNot(HaveOccurred())
		err = remoteUniversal.DismissCluster()
		Expect(err).ToNot(HaveOccurred())

		err = remoteKube.DeleteKuma(optsRemoteKube...)
		Expect(err).ToNot(HaveOccurred())
		err = remoteKube.DismissCluster()
		Expect(err).ToNot(HaveOccurred())
	})

	trafficAllowed := func() {
		stdout, _, err := remoteKube.ExecWithRetries(TestNamespace, clientPodName, "demo-client",
			"curl", "-v", "-m", "3", "--fail", "echo-server_kuma-test_svc_8080.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("Echo universal"))
	}

	trafficBlocked := func() {
		Eventually(func() error {
			_, _, err := remoteKube.Exec(TestNamespace, clientPodName, "demo-client",
				"curl", "-v", "-m", "3", "--fail", "echo-server_kuma-test_svc_8080.mesh")
			return err
		}, "30s", "1s").Should(HaveOccurred())
	}

	removeDefaultTrafficPermission := func() {
		err := k8s.RunKubectlE(globalCluster.GetTesting(), globalCluster.GetKubectlOptions(), "delete", "trafficpermission", "allow-all-default")
		Expect(err).ToNot(HaveOccurred())
	}

	It("should allow the traffic with default traffic permission", func() {
		// given default traffic permission

		// then
		trafficAllowed()

		// when
		removeDefaultTrafficPermission()

		// then
		trafficBlocked()
	})

	It("should allow the traffic with kuma.io/zone", func() {
		// given
		removeDefaultTrafficPermission()
		trafficBlocked()

		// when
		yaml := `
apiVersion: kuma.io/v1alpha1
kind: TrafficPermission
mesh: default
metadata:
  name: example-on-zone
spec:
  sources:
    - match:
        kuma.io/zone: kuma-2-remote
  destinations:
    - match:
        kuma.io/zone: kuma-3
`
		err := YamlK8s(yaml)(globalCluster)
		Expect(err).ToNot(HaveOccurred())

		// then
		trafficAllowed()
	})

	It("should allow the traffic with kuma.io/service", func() {
		// given
		removeDefaultTrafficPermission()
		trafficBlocked()

		// when
		yaml := `
apiVersion: kuma.io/v1alpha1
kind: TrafficPermission
mesh: default
metadata:
  name: example-on-service
spec:
  sources:
    - match:
        kuma.io/service: demo-client_kuma-test_svc
  destinations:
    - match:
        kuma.io/service: echo-server_kuma-test_svc_8080
`
		err := YamlK8s(yaml)(globalCluster)
		Expect(err).ToNot(HaveOccurred())

		// then
		trafficAllowed()
	})
}
