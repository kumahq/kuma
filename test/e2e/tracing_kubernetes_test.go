package e2e_test

import (
	"fmt"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/kumahq/kuma/pkg/config/mode"
	. "github.com/kumahq/kuma/test/framework"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
)

var _ = Describe("Test Tracing K8S", func() {

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

	meshWithTracing := func(zipkinURL string) string {
		return fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: default
spec:
  tracing:
    defaultBackend: zipkin
    backends:
    - name: zipkin
      type: zipkin
      conf:
        url: %s
`, zipkinURL)
	}

	zipkinAll := `
apiVersion: kuma.io/v1alpha1
kind: TrafficTrace
mesh: default
metadata:
  namespace: default
  name: trace-all
spec:
  selectors:
  - match:
      kuma.io/service: '*'
`

	var cluster Cluster

	BeforeEach(func() {
		c, err := NewK8SCluster(NewTestingT(), Kuma1, Silent)
		Expect(err).ToNot(HaveOccurred())
		cluster = c

		err = NewClusterSetup().
			Install(Kuma(mode.Standalone)).
			Install(KumaDNS()).
			Install(YamlK8s(namespaceWithSidecarInjection(TestNamespace))).
			Install(DemoClientK8s()).
			Install(EchoServerK8s()).
			Install(JaegerTracing()).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(cluster.DismissCluster()).To(Succeed())
		Expect(cluster.DeleteKuma()).To(Succeed())
		Expect(k8s.KubectlDeleteFromStringE(cluster.GetTesting(), cluster.GetKubectlOptions(), namespaceWithSidecarInjection(TestNamespace))).To(Succeed())
	})

	It("should emit traces to jaeger", func() {
		// given TrafficTrace and mesh with tracing backend
		err := YamlK8s(meshWithTracing(cluster.Tracing().ZipkinCollectorURL()))(cluster)
		Expect(err).ToNot(HaveOccurred())
		err = YamlK8s(zipkinAll)(cluster)
		Expect(err).ToNot(HaveOccurred())

		// when client sends requests to server
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

		retry.DoWithRetry(cluster.GetTesting(), "curl remote service",
			DefaultRetries, DefaultTimeout,
			func() (string, error) {
				_, _, err := cluster.ExecWithRetries(TestNamespace, clientPod.GetName(), "demo-client",
					"curl", "-v", "-m", "3", "echo-server")
				if err != nil {
					return "", err
				}

				// then traces are published
				services, err := cluster.Tracing().TracedServices()
				if err != nil {
					return "", err
				}

				expectedServices := []string{"demo-client_kuma-test_svc_3000", "echo-server_kuma-test_svc_80", "jaeger-query"}
				if !reflect.DeepEqual(services, expectedServices) {
					return "", errors.Errorf("services not traced. Expected %q, got %q", expectedServices, services)
				}
				return "ok", nil
			})
	})
})
