package tracing

import (
	"fmt"
	"reflect"

	"github.com/gruntwork-io/terratest/modules/retry"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/tracing"
)

func TracingUniversal() {
	meshWithTracing := func(zipkinURL string) string {
		return fmt.Sprintf(`
type: Mesh
name: default
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
type: TrafficTrace
name: traffic-trace-all
mesh: default
selectors:
- match:
   kuma.io/service: "*"
`

	var cluster Cluster

	BeforeEach(func() {
		cluster = NewUniversalCluster(NewTestingT(), Kuma3, Silent)

		err := NewClusterSetup().
			Install(Kuma(core.Standalone)).
			Install(tracing.Install()).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())

		testServerToken, err := cluster.GetKuma().GenerateDpToken("default", "test-server")
		Expect(err).ToNot(HaveOccurred())
		demoClientToken, err := cluster.GetKuma().GenerateDpToken("default", "demo-client")
		Expect(err).ToNot(HaveOccurred())

		err = TestServerUniversal("test-server", "default", testServerToken, WithArgs([]string{"echo", "--instance", "universal1"}))(cluster)
		Expect(err).ToNot(HaveOccurred())
		err = DemoClientUniversal(AppModeDemoClient, "default", demoClientToken, WithTransparentProxy(true))(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}
		Expect(cluster.DeleteKuma()).To(Succeed())
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	It("should emit traces to jaeger", func() {
		// given TrafficTrace and mesh with tracing backend
		err := YamlUniversal(meshWithTracing(tracing.From(cluster).ZipkinCollectorURL()))(cluster)
		Expect(err).ToNot(HaveOccurred())
		err = YamlUniversal(traceAll)(cluster)
		Expect(err).ToNot(HaveOccurred())

		retry.DoWithRetry(cluster.GetTesting(), "check traced services", DefaultRetries, DefaultTimeout, func() (string, error) {
			// when client sends requests to server
			_, _, err := cluster.Exec("", "", "demo-client", "curl", "-v", "-m", "3", "--fail", "test-server.mesh")
			if err != nil {
				return "", err
			}

			// then traces are published
			services, err := tracing.From(cluster).TracedServices()
			if err != nil {
				return "", err
			}

			expectedServices := []string{"demo-client", "jaeger-query", "test-server"}
			if !reflect.DeepEqual(services, expectedServices) {
				return "", errors.Errorf("services not traced. Expected %q, got %q", expectedServices, services)
			}
			return "", nil
		})
	})
}
