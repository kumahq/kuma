package gateway

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/test/e2e_env/kubernetes/env"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
)

func successfullyProxyRequestToGateway(cluster Cluster, instance string, gateway string, port int, namespace string) {
	Logf("expecting 200 response from %q", gateway)
	target := fmt.Sprintf("http://%s:%d/%s",
		gateway, port, path.Join("test", url.PathEscape(GinkgoT().Name())),
	)

	response, err := client.CollectResponse(
		cluster, "demo-client", target,
		client.FromKubernetesPod(namespace, "demo-client"),
	)

	Expect(err).ToNot(HaveOccurred())
	Expect(response.Instance).To(Equal(instance))
}

func failToProxyRequestToGateway(cluster Cluster, gateway string, port int, namespace string) func(Gomega) {
	return func(g Gomega) {
		Logf("expecting failure from %q", gateway)
		target := fmt.Sprintf("http://%s:%d/%s",
			gateway, port, path.Join("test", url.PathEscape(GinkgoT().Name())),
		)

		response, err := client.CollectFailure(
			cluster, "demo-client", target,
			client.FromKubernetesPod(namespace, "demo-client"),
		)

		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(response.Exitcode).To(Or(Equal(56), Equal(7), Equal(28)))
	}
}

func mkGateway(name, namespace, mesh string, crossMesh bool, hostname, backendService string, port int) string {
	meshGateway := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshGateway
metadata:
  name: %s
mesh: %s
spec:
  selectors:
  - match:
      kuma.io/service: %s
  conf:
    listeners:
    - port: %d
      protocol: HTTP
      crossMesh: %t
      hostname: %s
`, name, mesh, name, port, crossMesh, hostname)

	route := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshGatewayRoute
metadata:
  name: %s
mesh: %s
spec:
  selectors:
  - match:
      kuma.io/service: %s
  conf:
    http:
      rules:
      - matches:
        - path:
            match: PREFIX
            value: /
        backends:
        - destination:
            kuma.io/service: %s # Matches the echo-server we deployed.
`, name, mesh, name, backendService)

	instance := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshGatewayInstance
metadata:
  name: %s
  namespace: %s
  annotations:
    kuma.io/mesh: %s
spec:
  replicas: 1
  serviceType: ClusterIP
  tags:
    kuma.io/service: %s
`, name, namespace, mesh, name)

	return strings.Join([]string{meshGateway, route, instance}, "\n---\n")
}

func gatewayAddress(instanceName, instanceNamespace string) string {
	services, err := k8s.ListServicesE(env.Cluster.GetTesting(), env.Cluster.GetKubectlOptions(instanceNamespace), metav1.ListOptions{})
	Expect(err).ToNot(HaveOccurred())

	// Find the service that is owned by the named GatewayInstance.
	for _, svc := range services {
		for _, ref := range svc.GetOwnerReferences() {
			if ref.Kind == "MeshGatewayInstance" && ref.Name == instanceName {
				return svc.Spec.ClusterIP
			}
		}
	}

	return "0.0.0.0"
}
