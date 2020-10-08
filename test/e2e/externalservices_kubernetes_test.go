package e2e_test

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/externalservice"
)

var _ = Describe("Test ExternalServices on Kubernetes", func() {

	meshDefaulMtlsOn := `
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: default
spec:
  mtls:
    enabledBackend: ca-1
    backends:
      - name: ca-1
        type: builtin
`

	trafficPermissionAll := `
apiVersion: kuma.io/v1alpha1
kind: TrafficPermission
mesh: default
metadata:
  namespace: default
  name: allow-all-traffic
spec:
  sources:
    - match:
        kuma.io/service: '*'
  destinations:
    - match:
        kuma.io/service: '*'
`

	trafficRoute := `
apiVersion: kuma.io/v1alpha1
kind: TrafficRoute
mesh: default
metadata:
  name: rule-example
  namespace: default
spec:
  sources:
  - match:
      kuma.io/service: "*"
  destinations:
  - match:
      kuma.io/service: external-service
  conf:
    - weight: 1
      destination:
        kuma.io/service: external-service
        id: "%s"
`

	externalService := `
apiVersion: kuma.io/v1alpha1
kind: ExternalService
mesh: default
metadata:
  namespace: default
  name: external-service-%s
spec:
  tags:
    kuma.io/service: external-service
    kuma.io/protocol: http
    id: "%s"
  networking:
    address: %s:80
    tls:
      enabled: %s
`
	es1 := "1"
	es2 := "2"

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

	var cluster Cluster

	BeforeEach(func() {
		clusters, err := NewK8sClusters(
			[]string{Kuma1},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		// Global
		cluster = clusters.GetCluster(Kuma1)

		err = NewClusterSetup().
			Install(Kuma(core.Standalone)).
			Install(KumaDNS()).
			Install(YamlK8s(namespaceWithSidecarInjection(TestNamespace))).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())
		err = cluster.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(externalservice.Install(externalservice.HttpServer, []string{})).
			Install(externalservice.Install(externalservice.HttpsServer, []string{})).
			Install(DemoClientK8s()).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())

		err = YamlK8s(meshDefaulMtlsOn)(cluster)
		Expect(err).ToNot(HaveOccurred())

		err = YamlK8s(trafficPermissionAll)(cluster)
		Expect(err).ToNot(HaveOccurred())

		externalServiceAddress := externalservice.From(cluster, externalservice.HttpServer).GetExternalAppAddress()
		Expect(err).ToNot(HaveOccurred())

		err = YamlK8s(fmt.Sprintf(externalService,
			es1, es1,
			externalServiceAddress,
			"false"))(cluster)
		Expect(err).ToNot(HaveOccurred())

		externalServiceAddress = externalservice.From(cluster, externalservice.HttpsServer).GetExternalAppAddress()
		Expect(err).ToNot(HaveOccurred())

		err = YamlK8s(fmt.Sprintf(externalService,
			es2, es2,
			externalServiceAddress,
			"true"))(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		err := cluster.DeleteKuma()
		Expect(err).ToNot(HaveOccurred())

		err = cluster.DismissCluster()
		Expect(err).ToNot(HaveOccurred())

		err = k8s.DeleteNamespaceE(cluster.GetTesting(),
			cluster.GetKubectlOptions(TestNamespace),
			TestNamespace)
		Expect(err).ToNot(HaveOccurred())

		cluster.(*K8sCluster).WaitNamespaceDelete(TestNamespace)
	})

	It("Should route to external-service", func() {
		err := YamlK8s(fmt.Sprintf(trafficRoute, es1))(cluster)
		Expect(err).ToNot(HaveOccurred())

		pods, err := k8s.ListPodsE(
			cluster.GetTesting(),
			cluster.GetKubectlOptions(TestNamespace),
			metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", "demo-client"),
			},
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(pods).To(HaveLen(1))

		clientPod := pods[0]

		stdout, stderr, err := cluster.ExecWithRetries(TestNamespace, clientPod.GetName(), "demo-client",
			"curl", "-v", "-m", "3", "--fail", "http://external-service.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))
		Expect(stdout).ToNot(ContainSubstring("HTTPS"))
	})

	It("Should route to external-service over tls", func() {
		err := YamlK8s(fmt.Sprintf(trafficRoute, es2))(cluster)
		Expect(err).ToNot(HaveOccurred())

		pods, err := k8s.ListPodsE(
			cluster.GetTesting(),
			cluster.GetKubectlOptions(TestNamespace),
			metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", "demo-client"),
			},
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(pods).To(HaveLen(1))

		clientPod := pods[0]

		stdout, stderr, err := cluster.ExecWithRetries(TestNamespace, clientPod.GetName(), "demo-client",
			"curl", "-v", "-m", "3", "--fail", "http://external-service.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))
		Expect(stdout).To(ContainSubstring("HTTPS"))
	})

})
