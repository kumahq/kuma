package gateway

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/kic"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func Delegated() {
	namespace := "delegated-gateway"
	namespaceOutsideMesh := "delegated-gateway-outside-mesh"
	mesh := "delegated-gateway"

	var kicIP string

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MTLSMeshKubernetes(mesh)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Namespace(namespaceOutsideMesh)).
			Install(democlient.Install(
				democlient.WithNamespace(namespaceOutsideMesh),
			)).
			Install(testserver.Install(
				testserver.WithMesh(mesh),
				testserver.WithNamespace(namespace),
				testserver.WithName("test-server"),
			)).
			Install(kic.KongIngressController(
				kic.WithNamespace(namespace),
				kic.WithMesh(mesh),
			)).
			Install(kic.KongIngressService(kic.WithNamespace(namespace))).
			Install(YamlK8s(fmt.Sprintf(`
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  namespace: %s
  name: %s-ingress
  annotations:
    kubernetes.io/ingress.class: kong
spec:
  rules:
  - http:
      paths:
      - path: /test-server
        pathType: Prefix
        backend:
          service:
            name: test-server
            port:
              number: 80
`, namespace, mesh))).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())

		kicIP, err = kic.From(kubernetes.Cluster).IP(namespace)
		Expect(err).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).
			To(Succeed())
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespaceOutsideMesh)).
			To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(mesh)).To(Succeed())
	})

	Context("MeshCircuitBreaker", CircuitBreaker(namespaceOutsideMesh, mesh, kicIP))
}
