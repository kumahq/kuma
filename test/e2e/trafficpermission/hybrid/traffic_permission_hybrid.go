package hybrid

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

var globalCluster, zoneUniversal, zoneKube Cluster
var clientPodName string

func meshMTLSOn(mesh string) string {
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

var _ = E2EBeforeSuite(func() {
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

	Expect(NewClusterSetup().
		Install(Kuma(config_core.Global)).
		Install(YamlK8s(meshMTLSOn("default"))).
		Setup(globalCluster)).To(Succeed())

	E2EDeferCleanup(func() {
		Expect(globalCluster.DeleteKuma()).To(Succeed())
		Expect(globalCluster.DismissCluster()).To(Succeed())
	})

	globalCP := globalCluster.GetKuma()

	testServerToken, err := globalCP.GenerateDpToken("default", "test-server")
	Expect(err).ToNot(HaveOccurred())

	// Zone universal
	zoneUniversal = universalClusters.GetCluster(Kuma3)
	ingressTokenKuma3, err := globalCP.GenerateZoneIngressToken(Kuma3)
	Expect(err).ToNot(HaveOccurred())

	Expect(NewClusterSetup().
		Install(Kuma(config_core.Zone,
			WithGlobalAddress(globalCP.GetKDSServerAddress()),
		)).
		Install(TestServerUniversal("test-server", "default", testServerToken, WithArgs([]string{"echo", "--instance", "echo-v1"}))).
		Install(IngressUniversal(ingressTokenKuma3)).
		Setup(zoneUniversal)).To(Succeed())

	E2EDeferCleanup(zoneUniversal.DismissCluster)

	// Zone kubernetes
	zoneKube = k8sClusters.GetCluster(Kuma2)

	Expect(NewClusterSetup().
		Install(Kuma(config_core.Zone,
			WithGlobalAddress(globalCP.GetKDSServerAddress()),
		)).
		Install(NamespaceWithSidecarInjection(TestNamespace)).
		Install(DemoClientK8s("default")).
		Setup(zoneKube)).To(Succeed())

	pods, err := k8s.ListPodsE(
		zoneKube.GetTesting(),
		zoneKube.GetKubectlOptions(TestNamespace),
		metav1.ListOptions{
			LabelSelector: fmt.Sprintf("app=%s", "demo-client"),
		},
	)
	Expect(err).ToNot(HaveOccurred())
	Expect(pods).To(HaveLen(1))

	clientPodName = pods[0].GetName()

	E2EDeferCleanup(func() {
		Expect(zoneKube.DeleteKuma()).To(Succeed())
		Expect(zoneKube.DismissCluster()).To(Succeed())
	})
})

func TrafficPermissionHybrid() {
	E2EAfterEach(func() {
		// remove all TrafficPermissions and restore the default
		Expect(k8s.RunKubectlE(globalCluster.GetTesting(), globalCluster.GetKubectlOptions(), "delete", "trafficpermissions", "--all")).To(Succeed())

		Expect(k8s.KubectlApplyFromStringE(globalCluster.GetTesting(), globalCluster.GetKubectlOptions(), `
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

`)).To(Succeed())
	})

	trafficAllowed := func() {
		_, _, err := zoneKube.ExecWithRetries(TestNamespace, clientPodName, "demo-client",
			"curl", "-v", "-m", "3", "--fail", "test-server.mesh")
		Expect(err).ToNot(HaveOccurred())
	}

	trafficBlocked := func() {
		Eventually(func() error {
			_, _, err := zoneKube.Exec(TestNamespace, clientPodName, "demo-client",
				"curl", "-v", "-m", "3", "--fail", "test-server.mesh")
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
        kuma.io/zone: kuma-2-zone
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
        kuma.io/service: test-server
`
		err := YamlK8s(yaml)(globalCluster)
		Expect(err).ToNot(HaveOccurred())

		// then
		trafficAllowed()
	})

	It("should allow the traffic with k8s.kuma.io/namespace", func() {
		// given
		removeDefaultTrafficPermission()
		trafficBlocked()

		// when
		yaml := `
apiVersion: kuma.io/v1alpha1
kind: TrafficPermission
mesh: default
metadata:
  name: example-on-namespace
spec:
  sources:
    - match:
        k8s.kuma.io/namespace: kuma-test
  destinations:
    - match:
        kuma.io/service: test-server
`
		err := YamlK8s(yaml)(globalCluster)
		Expect(err).ToNot(HaveOccurred())

		// then
		trafficAllowed()
	})

	It("should allow the traffic with tags added dynamically on Kubernetes", func() {
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
        newtag: client
  destinations:
    - match:
        kuma.io/service: test-server
`
		err := YamlK8s(yaml)(globalCluster)
		Expect(err).ToNot(HaveOccurred())

		// and when Kubernetes pod is labeled
		err = k8s.RunKubectlE(zoneKube.GetTesting(), zoneKube.GetKubectlOptions(TestNamespace), "label", "pod", clientPodName, "newtag=client")
		Expect(err).ToNot(HaveOccurred())

		// then
		trafficAllowed()
	})
}
