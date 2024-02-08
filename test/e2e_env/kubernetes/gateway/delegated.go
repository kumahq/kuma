package gateway

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/kubernetes/gateway/delegated"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/kic"
	"github.com/kumahq/kuma/test/framework/deployments/observability"
	"github.com/kumahq/kuma/test/framework/deployments/otelcollector"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func Delegated() {
	config := delegated.Config{
		Namespace:                   "delegated-gateway",
		NamespaceOutsideMesh:        "delegated-gateway-outside-mesh",
		Mesh:                        "delegated-gateway",
		KicIP:                       "",
		CpNamespace:                 Config.KumaNamespace,
		ObservabilityDeploymentName: "observability-delegated-meshtrace",
	}

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MTLSMeshKubernetes(config.Mesh)).
			Install(MeshTrafficPermissionAllowAllKubernetes(config.Mesh)).
			Install(NamespaceWithSidecarInjection(config.Namespace)).
			Install(Namespace(config.NamespaceOutsideMesh)).
			Install(democlient.Install(
				democlient.WithNamespace(config.NamespaceOutsideMesh),
				democlient.WithService(true),
			)).
			Install(testserver.Install(
				testserver.WithMesh(config.Mesh),
				testserver.WithNamespace(config.Namespace),
				testserver.WithName("test-server"),
				testserver.WithStatefulSet(true),
				testserver.WithReplicas(3),
			)).
			Install(testserver.Install(
				testserver.WithNamespace(config.NamespaceOutsideMesh),
				testserver.WithName("external-service"),
			)).
			Install(testserver.Install(
				testserver.WithNamespace(config.NamespaceOutsideMesh),
				testserver.WithName("external-tcp-service"),
			)).
			Install(otelcollector.Install(
				otelcollector.WithNamespace(config.NamespaceOutsideMesh),
				otelcollector.WithIPv6(Config.IPV6),
			)).
			Install(observability.Install(
				config.ObservabilityDeploymentName,
				observability.WithNamespace(config.NamespaceOutsideMesh),
				observability.WithComponents(observability.JaegerComponent),
			)).
			Install(kic.KongIngressController(
				kic.WithName("delegated"),
<<<<<<< HEAD
				kic.WithNamespace(config.Namespace),
				kic.WithMesh(config.Mesh),
			)).
			Install(kic.KongIngressService(
				kic.WithName("delegated"),
				kic.WithNamespace(config.Namespace),
			)).
=======
				kic.WithNamespace(config.namespace),
				kic.WithMesh(config.mesh),
			)).
			Install(kic.KongIngressService(
				kic.WithName("delegated"),
				kic.WithNamespace(config.namespace),
			)).
>>>>>>> 6cf0b3eea (test(e2e): upgrade KIC (#9157))
			Install(YamlK8s(fmt.Sprintf(`
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  namespace: %s
  name: %s-ingress
  annotations:
    kubernetes.io/ingress.class: delegated
<<<<<<< HEAD
    konghq.com/strip-path: 'true'
=======
>>>>>>> 6cf0b3eea (test(e2e): upgrade KIC (#9157))
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
		Expect(kubernetes.Cluster.DeleteDeployment(config.ObservabilityDeploymentName)).
			To(Succeed())
	})

	Context("MeshCircuitBreaker", delegated.CircuitBreaker(&config))
	Context("MeshProxyPatch", delegated.MeshProxyPatch(&config))
	Context("MeshHealthCheck", delegated.MeshHealthCheck(&config))
	Context("MeshRetry", delegated.MeshRetry(&config))
	Context("MeshHTTPRoute", delegated.MeshHTTPRoute(&config))
	Context("MeshTimeout", delegated.MeshTimeout(&config))
	Context("MeshMetric", delegated.MeshMetric(&config))
	Context("MeshTrace", delegated.MeshTrace(&config))
	Context("MeshLoadBalancingStrategy", delegated.MeshLoadBalancingStrategy(&config))
	Context("MeshAccessLog", delegated.MeshAccessLog(&config))
	XContext("MeshTCPRoute", delegated.MeshTCPRoute(&config))
}
