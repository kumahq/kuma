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

type delegatedE2EConfig struct {
	namespace            string
	namespaceOutsideMesh string
	mesh                 string
	kicIP                string
}

func Delegated() {
	config := &delegatedE2EConfig{
		namespace:            "delegated-gateway",
		namespaceOutsideMesh: "delegated-gateway-outside-mesh",
		mesh:                 "delegated-gateway",
		kicIP:                "",
	}

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MTLSMeshKubernetes(config.mesh)).
			Install(MeshTrafficPermissionAllowAllKubernetes(config.mesh)).
			Install(NamespaceWithSidecarInjection(config.namespace)).
			Install(Namespace(config.namespaceOutsideMesh)).
			Install(democlient.Install(
				democlient.WithNamespace(config.namespaceOutsideMesh),
			)).
			Install(testserver.Install(
				testserver.WithMesh(config.mesh),
				testserver.WithNamespace(config.namespace),
				testserver.WithName("test-server"),
			)).
			Install(kic.KongIngressController(
				kic.WithName("delegated"),
				kic.WithNamespace(config.namespace),
				kic.WithMesh(config.mesh),
			)).
			Install(kic.KongIngressService(
				kic.WithName("delegated"),
				kic.WithNamespace(config.namespace),
			)).
			Install(YamlK8s(fmt.Sprintf(`
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  namespace: %s
  name: %s-ingress
  annotations:
    kubernetes.io/ingress.class: delegated
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
`, config.namespace, config.mesh))).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())

		kicIP, err := kic.From(kubernetes.Cluster).IP(config.namespace)
		Expect(err).To(Succeed())

		config.kicIP = kicIP
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(config.namespace)).
			To(Succeed())
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(config.namespaceOutsideMesh)).
			To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(config.mesh)).To(Succeed())
	})

	Context("MeshCircuitBreaker", CircuitBreaker(config))
}
