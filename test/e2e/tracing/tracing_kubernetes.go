package tracing

import (
	"fmt"
	"reflect"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/deployments/tracing"
)

func TracingK8S() {
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

	traceAll := `
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
		cluster = NewK8sCluster(NewTestingT(), Kuma1, Silent)
		err := NewClusterSetup().
			Install(Kuma(core.Standalone)).
			Install(NamespaceWithSidecarInjection(TestNamespace)).
			Install(DemoClientK8s("default")).
			Install(testserver.Install()).
			Install(tracing.Install()).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterEach(func() {
		Expect(cluster.DeleteKuma()).To(Succeed())
		Expect(cluster.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	It("should emit traces to jaeger", func() {
		// given TrafficTrace and mesh with tracing backend
		err := YamlK8s(meshWithTracing(tracing.From(cluster).ZipkinCollectorURL()))(cluster)
		Expect(err).ToNot(HaveOccurred())
		err = YamlK8s(traceAll)(cluster)
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
					"curl", "-v", "-m", "3", "--fail", "test-server")
				if err != nil {
					return "", err
				}

				// then traces are published
				services, err := tracing.From(cluster).TracedServices()
				if err != nil {
					return "", err
				}

				expectedServices := []string{"demo-client_kuma-test_svc", "jaeger-query", "test-server_kuma-test_svc_80"}
				if !reflect.DeepEqual(services, expectedServices) {
					return "", errors.Errorf("services not traced. Expected %q, got %q", expectedServices, services)
				}
				return "ok", nil
			})
	})
}
