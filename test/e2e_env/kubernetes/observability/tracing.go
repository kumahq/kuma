package observability

import (
	"fmt"
	"reflect"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/test/e2e_env/kubernetes/env"
	. "github.com/kumahq/kuma/test/framework"
	obs "github.com/kumahq/kuma/test/framework/deployments/observability"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func meshWithTracing(name, zipkinURL string) string {
	return fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: %s
spec:
  tracing:
    defaultBackend: zipkin
    backends:
    - name: zipkin
      type: zipkin
      conf:
        url: %s
`, name, zipkinURL)
}

func trafficTrace(mesh, namespace string) string {
	return fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: TrafficTrace
mesh: %s
metadata:
  namespace: %s
  name: trace-all
spec:
  selectors:
  - match:
      kuma.io/service: '*'
`, mesh, namespace)
}

func Tracing() {
	ns := "tracing"
	obsNs := "obs-tracing"
	obsDeployment := "obs-tracing-deployment"
	mesh := "tracing"

	var obsClient obs.Observability
	BeforeAll(func() {
		err := NewClusterSetup().
			Install(NamespaceWithSidecarInjection(ns)).
			Install(MeshKubernetes(mesh)).
			Install(DemoClientK8s(mesh, ns)).
			Install(testserver.Install(testserver.WithMesh(mesh), testserver.WithNamespace(ns))).
			Install(obs.Install(obsDeployment, obs.WithNamespace(obsNs))).
			Setup(env.Cluster)
		obsClient = obs.From(obsDeployment, env.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})
	E2EAfterAll(func() {
		Expect(env.Cluster.TriggerDeleteNamespace(ns)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(mesh)).To(Succeed())
		Expect(env.Cluster.DeleteDeployment(obsDeployment)).To(Succeed())
	})

	It("should emit traces to jaeger", func() {
		// given TrafficTrace and mesh with tracing backend
		err := YamlK8s(meshWithTracing(mesh, obsClient.ZipkinCollectorURL()))(env.Cluster)
		Expect(err).ToNot(HaveOccurred())
		err = YamlK8s(trafficTrace(mesh, ns))(env.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// when client sends requests to server
		pods, err := k8s.ListPodsE(
			env.Cluster.GetTesting(),
			env.Cluster.GetKubectlOptions(ns),
			metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", "demo-client"),
			},
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(pods).To(HaveLen(1))

		clientPod := pods[0]

		retry.DoWithRetry(env.Cluster.GetTesting(), "curl remote service",
			DefaultRetries, DefaultTimeout,
			func() (string, error) {
				_, _, err := env.Cluster.ExecWithRetries(ns, clientPod.GetName(), "demo-client",
					"curl", "-v", "-m", "3", "--fail", "test-server")
				if err != nil {
					return "", err
				}

				// then traces are published
				services, err := obsClient.TracedServices()
				if err != nil {
					return "", err
				}

				expectedServices := []string{
					fmt.Sprintf("demo-client_%s_svc", ns),
					"jaeger-query",
					fmt.Sprintf("test-server_%s_svc_80", ns),
				}
				if !reflect.DeepEqual(services, expectedServices) {
					return "", errors.Errorf("services not traced. Expected %q, got %q", expectedServices, services)
				}
				return "ok", nil
			})
	})
}
