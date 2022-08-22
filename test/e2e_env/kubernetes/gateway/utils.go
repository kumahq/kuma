package gateway

import (
	"fmt"
	"net"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/test/e2e_env/kubernetes/env"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
)

func GatewayAPICRDs(cluster Cluster) error {
	return k8s.RunKubectlE(
		cluster.GetTesting(),
		cluster.GetKubectlOptions(),
		"apply", "-f", "https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.5.0/experimental-install.yaml")
}

func SuccessfullyProxyRequestToGateway(cluster Cluster, instance string, gatewayAddr string, namespace string) {
	Logf("expecting 200 response from %q", gatewayAddr)
	target := fmt.Sprintf("http://%s/%s",
		gatewayAddr, path.Join("test", url.PathEscape(GinkgoT().Name())),
	)

	response, err := client.CollectResponse(
		cluster, "demo-client", target,
		client.FromKubernetesPod(namespace, "demo-client"),
	)

	Expect(err).ToNot(HaveOccurred())
	Expect(response.Instance).To(Equal(instance))
}

func FailToProxyRequestToGateway(cluster Cluster, gatewayAddr string, namespace string) func(Gomega) {
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

func mkGateway(name, mesh string, crossMesh bool, hostname, backendService string, port int) string {
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
	return strings.Join([]string{meshGateway, route}, "\n---\n")
}

func MkGatewayInstance(name, namespace, mesh string) string {
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

	return instance
}

func gatewayAddress(instanceName, instanceNamespace string, port int) string {
	services, err := k8s.ListServicesE(env.Cluster.GetTesting(), env.Cluster.GetKubectlOptions(instanceNamespace), metav1.ListOptions{})
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
