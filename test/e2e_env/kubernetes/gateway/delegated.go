package gateway

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/kubernetes/gateway/delegated"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/kic"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func Delegated() {
	config := delegated.Config{
		Namespace:            "delegated-gateway",
		NamespaceOutsideMesh: "delegated-gateway-outside-mesh",
		Mesh:                 "delegated-gateway",
		KicIP:                "",
		CpNamespace:          Config.KumaNamespace,
	}

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MTLSMeshKubernetes(config.Mesh)).
			Install(MeshTrafficPermissionAllowAllKubernetes(config.Mesh)).
			Install(NamespaceWithSidecarInjection(config.Namespace)).
			Install(Namespace(config.NamespaceOutsideMesh)).
			Install(democlient.Install(
				democlient.WithNamespace(config.NamespaceOutsideMesh),
			)).
			Install(testserver.Install(
				testserver.WithMesh(config.Mesh),
				testserver.WithNamespace(config.Namespace),
				testserver.WithName("test-server"),
			)).
			Install(kic.KongIngressController(
<<<<<<< HEAD
				kic.WithNamespace(config.namespace),
				kic.WithMesh(config.mesh),
			)).
			Install(kic.KongIngressService(kic.WithNamespace(config.namespace))).
=======
				kic.WithName("delegated"),
				kic.WithNamespace(config.Namespace),
				kic.WithMesh(config.Mesh),
			)).
			Install(kic.KongIngressService(
				kic.WithName("delegated"),
				kic.WithNamespace(config.Namespace),
			)).
>>>>>>> 67ee1be51 (fix(MeshGateway): fix MeshTCPRoute on MeshGateway (#9167))
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
`, config.Namespace, config.Mesh))).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())

		kicIP, err := kic.From(kubernetes.Cluster).IP(config.Namespace)
		Expect(err).ToNot(HaveOccurred())

		config.KicIP = kicIP
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(config.Namespace)).
			To(Succeed())
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(config.NamespaceOutsideMesh)).
			To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(config.Mesh)).To(Succeed())
	})

<<<<<<< HEAD
	Context("MeshCircuitBreaker", CircuitBreaker(config))
=======
	Context("MeshCircuitBreaker", delegated.CircuitBreaker(&config))
	Context("MeshProxyPatch", delegated.MeshProxyPatch(&config))
>>>>>>> 67ee1be51 (fix(MeshGateway): fix MeshTCPRoute on MeshGateway (#9167))
}
