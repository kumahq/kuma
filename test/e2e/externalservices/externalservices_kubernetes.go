package externalservices

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/externalservice"
)

func ExternalServicesOnKubernetes() {
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
  networking:
    outbound:
      passthrough: %s
`

	externalService := `
apiVersion: kuma.io/v1alpha1
kind: ExternalService
mesh: default
metadata:
  name: external-service-%s
spec:
  tags:
    kuma.io/service: external-service
    kuma.io/protocol: http
    id: "%s"
  networking:
    address: %s:%d
    tls:
      enabled: %s
`
	es1 := "1"
	es2 := "2"

	trafficPermission := `
apiVersion: kuma.io/v1alpha1
kind: TrafficPermission
mesh: default
metadata:
  name: traffic-to-es
spec:
  sources:
    - match:
        kuma.io/service: '*'
  destinations:
    - match:
        kuma.io/service: external-service

`

	var cluster Cluster
	var clientPod *v1.Pod

	BeforeEach(func() {
		clusters, err := NewK8sClusters(
			[]string{Kuma1},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		// Global
		cluster = clusters.GetCluster(Kuma1)
		err = NewClusterSetup().
			Install(Kuma(core.Standalone)).
			Install(NamespaceWithSidecarInjection(TestNamespace)).
			Install(DemoClientK8s("default")).
			Install(externalservice.Install(externalservice.HttpServer, []string{})).
			Install(externalservice.Install(externalservice.HttpsServer, []string{})).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())

		err = YamlK8s(fmt.Sprintf(meshDefaulMtlsOn, "false"))(cluster)
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

		clientPod = &pods[0]
	})

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}

		err := cluster.DeleteNamespace(TestNamespace)
		Expect(err).ToNot(HaveOccurred())

		err = cluster.DeleteKuma()
		Expect(err).ToNot(HaveOccurred())

		err = cluster.DismissCluster()
		Expect(err).ToNot(HaveOccurred())
	})

	trafficBlocked := func() error {
		_, _, err := cluster.Exec(TestNamespace, clientPod.GetName(), "demo-client",
			"curl", "-v", "-m", "3", "--fail", "http://externalservice-http-server.externalservice-namespace:10080")
		return err
	}

	It("should route to external-service", func() {
		// given Mesh with passthrough enabled
		err := YamlK8s(fmt.Sprintf(meshDefaulMtlsOn, "true"))(cluster)
		Expect(err).ToNot(HaveOccurred())

		// and no default traffic permission
		err = k8s.RunKubectlE(cluster.GetTesting(), cluster.GetKubectlOptions(), "delete", "trafficpermission", "allow-all-default")
		Expect(err).ToNot(HaveOccurred())

		// then communication outside of the Mesh works
		_, stderr, err := cluster.ExecWithRetries(TestNamespace, clientPod.GetName(), "demo-client",
			"curl", "-v", "-m", "3", "--fail", "http://externalservice-http-server.externalservice-namespace:10080")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))

		// when passthrough is disabled on the Mesh
		err = YamlK8s(fmt.Sprintf(meshDefaulMtlsOn, "false"))(cluster)
		Expect(err).ToNot(HaveOccurred())

		// then accessing the external service is no longer possible
		Eventually(trafficBlocked, "30s", "1s").Should(HaveOccurred())

		// when apply external service
		err = YamlK8s(fmt.Sprintf(externalService,
			es1, es1,
			"externalservice-http-server.externalservice-namespace.svc.cluster.local", 10080, // .svc.cluster.local is needed, otherwise Kubernetes will resolve this to the real IP
			"false"))(cluster)
		Expect(err).ToNot(HaveOccurred())

		// then traffic is still blocked because of lack of the traffic permission
		Consistently(trafficBlocked, "10s", "1s").Should(HaveOccurred())

		// when TrafficPermission is added
		err = YamlK8s(trafficPermission)(cluster)
		Expect(err).ToNot(HaveOccurred())

		// then you can access external service again
		stdout, stderr, err := cluster.ExecWithRetries(TestNamespace, clientPod.GetName(), "demo-client",
			"curl", "-v", "-m", "3", "--fail", "http://externalservice-http-server.externalservice-namespace:10080")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))
		Expect(stdout).ToNot(ContainSubstring("externalservice-https-server"))

		// and you can also use .mesh on port of the provided host
		stdout, stderr, err = cluster.ExecWithRetries(TestNamespace, clientPod.GetName(), "demo-client",
			"curl", "-v", "-m", "3", "--fail", "http://external-service.mesh:10080")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))
		Expect(stdout).ToNot(ContainSubstring("externalservice-https-server"))

		// and you can also use .mesh on port 80
		// todo (lobkovilya): check of backward compatibility, could be deleted in the next major release Kuma 1.2.x
		stdout, stderr, err = cluster.ExecWithRetries(TestNamespace, clientPod.GetName(), "demo-client",
			"curl", "-v", "-m", "3", "--fail", "http://external-service.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))
		Expect(stdout).ToNot(ContainSubstring("externalservice-https-server"))
	})

	It("should route to external-service over tls", func() {
		err := YamlK8s(fmt.Sprintf(externalService,
			es2, es2,
			"externalservice-https-server.externalservice-namespace.svc.cluster.local", 10080,
			"true"))(cluster)
		Expect(err).ToNot(HaveOccurred())

		stdout, stderr, err := cluster.ExecWithRetries(TestNamespace, clientPod.GetName(), "demo-client",
			"curl", "-v", "-m", "3", "--fail", "http://external-service.mesh:10080")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))
		Expect(stdout).To(ContainSubstring("externalservice-https-server"))
	})
}
