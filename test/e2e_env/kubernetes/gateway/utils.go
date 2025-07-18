package gateway

import (
	"fmt"
	"net"
	"net/url"
	"path"
	"strconv"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func SuccessfullyProxyRequestToGateway(cluster Cluster, instance, gatewayAddr, namespace string) func(Gomega) {
	return func(g Gomega) {
		target := fmt.Sprintf("http://%s/%s",
			gatewayAddr, path.Join("test", url.PathEscape(GinkgoT().Name())),
		)

		response, err := client.CollectEchoResponse(
			cluster, "demo-client", target,
			client.FromKubernetesPod(namespace, "demo-client"),
		)

		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(response.Instance).To(Equal(instance))
	}
}

func FailToProxyRequestToGateway(cluster Cluster, gatewayAddr, namespace string) func(Gomega) {
	return func(g Gomega) {
		Logf("expecting failure from %q", gatewayAddr)
		target := fmt.Sprintf("http://%s/%s",
			gatewayAddr, path.Join("test", url.PathEscape(GinkgoT().Name())),
		)

		response, err := client.CollectFailure(
			cluster, "demo-client", target,
			client.FromKubernetesPod(namespace, "demo-client"),
		)

		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(response.Exitcode).To(Or(Equal(56), Equal(7), Equal(28)))
	}
}

func mkGateway(resourceName, serviceName, mesh string, crossMesh bool, hostname, backendService string, port int) string {
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
`, resourceName, mesh, serviceName, port, crossMesh, hostname)

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
`, resourceName, mesh, serviceName, backendService)
	return meshGateway + "\n---\n" + route
}

func MkGatewayInstance(name, namespace, mesh string) string {
	instance := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshGatewayInstance
metadata:
  name: %s
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  replicas: 1
  serviceType: ClusterIP
`, name, namespace, mesh)

	return instance
}

func gatewayAddress(instanceName, instanceNamespace string, port int) string {
	services, err := k8s.ListServicesE(kubernetes.Cluster.GetTesting(), kubernetes.Cluster.GetKubectlOptions(instanceNamespace), metav1.ListOptions{})
	Expect(err).ToNot(HaveOccurred())

	var rawIP string

	// Find the service that is owned by the named GatewayInstance.
	for _, svc := range services {
		for _, ref := range svc.GetOwnerReferences() {
			if ref.Kind == "MeshGatewayInstance" && ref.Name == instanceName {
				rawIP = svc.Spec.ClusterIP
			}
		}
	}

	ip := net.ParseIP(rawIP)
	Expect(ip).ToNot(BeNil(), "invalid clusterIP for gateway")

	return net.JoinHostPort(rawIP, strconv.Itoa(port))
}
