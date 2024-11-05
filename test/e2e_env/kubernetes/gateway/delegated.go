package gateway

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	mcb_api "github.com/kumahq/kuma/pkg/plugins/policies/meshcircuitbreaker/api/v1alpha1"
	mr_api "github.com/kumahq/kuma/pkg/plugins/policies/meshretry/api/v1alpha1"
	mt_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
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
		IPV6:                        Config.IPV6,
		MeshServiceMode:             mesh_proto.Mesh_MeshServices_Disabled,
		UseEgress:                   false,
	}

	configMs := delegated.Config{
		Namespace:                   "delegated-gateway-ms",
		NamespaceOutsideMesh:        "delegated-gateway-outside-mesh-ms",
		Mesh:                        "delegated-gateway-ms",
		KicIP:                       "",
		CpNamespace:                 Config.KumaNamespace,
		ObservabilityDeploymentName: "observability-delegated-meshtrace-ms",
		IPV6:                        Config.IPV6,
		MeshServiceMode:             mesh_proto.Mesh_MeshServices_Exclusive,
		UseEgress:                   true,
	}
	contextFor := func(name string, config *delegated.Config, testMatrix map[string]func()) {
		Context(name, func() {
			externalNameService := func(serviceName string) string {
				return fmt.Sprintf(`apiVersion: v1
kind: Service
metadata:
  name: %s
  namespace: %s
spec:
  type: ExternalName
  externalName: %s.%s.svc.cluster.local`, serviceName, config.Namespace, serviceName, config.NamespaceOutsideMesh)
			}
			BeforeAll(func() {
				mesh := samples.MeshMTLSBuilder().WithName(config.Mesh).WithMeshServicesEnabled(config.MeshServiceMode)
				if config.UseEgress {
					mesh.WithEgressRoutingEnabled()
				}
				err := NewClusterSetup().
					Install(Yaml(mesh)).
					Install(MeshTrafficPermissionAllowAllKubernetes(config.Mesh)).
					Install(NamespaceWithSidecarInjection(config.Namespace)).
					Install(Namespace(config.NamespaceOutsideMesh)).
					Install(Parallel(
						democlient.Install(
							democlient.WithNamespace(config.NamespaceOutsideMesh),
							democlient.WithService(true),
						),
						testserver.Install(
							testserver.WithMesh(config.Mesh),
							testserver.WithNamespace(config.Namespace),
							testserver.WithName("test-server"),
							testserver.WithStatefulSet(),
							testserver.WithReplicas(3),
						),
						testserver.Install(
							testserver.WithNamespace(config.NamespaceOutsideMesh),
							testserver.WithName("external-service"),
						),
						testserver.Install(
							testserver.WithNamespace(config.NamespaceOutsideMesh),
							testserver.WithName("another-external-service"),
						),
						testserver.Install(
							testserver.WithNamespace(config.NamespaceOutsideMesh),
							testserver.WithName("external-tcp-service"),
						),
						otelcollector.Install(
							otelcollector.WithNamespace(config.NamespaceOutsideMesh),
							otelcollector.WithIPv6(Config.IPV6),
						),
						observability.Install(
							config.ObservabilityDeploymentName,
							observability.WithNamespace(config.NamespaceOutsideMesh),
							observability.WithComponents(observability.JaegerComponent),
						),
						kic.KongIngressController(
							kic.WithName(config.Mesh),
							kic.WithNamespace(config.Namespace),
							kic.WithMesh(config.Mesh),
						),
						kic.KongIngressService(
							kic.WithName(config.Mesh),
							kic.WithNamespace(config.Namespace),
						),
					)).
					Install(YamlK8s(externalNameService("external-service"))).
					Install(YamlK8s(externalNameService("another-external-service"))).
					Install(YamlK8s(fmt.Sprintf(`
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  namespace: %s
  name: %s-ingress
  annotations:
    kubernetes.io/ingress.class: %s
    konghq.com/strip-path: 'true'
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
  - http:
      paths:
      - path: /external-service
        pathType: Prefix
        backend:
          service:
            name: external-service
            port:
              number: 80
  - http:
      paths:
      - path: /another-external-service
        pathType: Prefix
        backend:
          service:
            name: another-external-service
            port:
              number: 80`, config.Namespace, config.Mesh, config.Mesh))).
					Setup(kubernetes.Cluster)
				Expect(err).ToNot(HaveOccurred())

				kicIP, err := kic.From(kubernetes.Cluster).IP(config.Namespace)
				Expect(err).ToNot(HaveOccurred())

				config.KicIP = kicIP
				Expect(DeleteMeshResources(
					kubernetes.Cluster,
					config.Mesh,
					mcb_api.MeshCircuitBreakerResourceTypeDescriptor,
					mt_api.MeshTimeoutResourceTypeDescriptor,
					mr_api.MeshRetryResourceTypeDescriptor,
				)).To(Succeed())
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

			// If you copy the test case from a non-gateway test or create a new test,
			// remember the the name of policies needs to be unique.
			// If they have the same name, one might override the other, causing a flake.
			for policyName, test := range testMatrix {
				Context(policyName, test)
			}
		})
	}

	contextFor("delegated with kuma.io/service", &config, map[string]func(){
		"MeshCircuitBreaker":        delegated.CircuitBreaker(&config),
		"MeshProxyPatch":            delegated.MeshProxyPatch(&config),
		"MeshHealthCheck":           delegated.MeshHealthCheck(&config),
		"MeshRetry":                 delegated.MeshRetry(&config),
		"MeshHTTPRoute":             delegated.MeshHTTPRoute(&config),
		"MeshTimeout":               delegated.MeshTimeout(&config),
		"MeshMetric":                delegated.MeshMetric(&config),
		"MeshTrace":                 delegated.MeshTrace(&config),
		"MeshLoadBalancingStrategy": delegated.MeshLoadBalancingStrategy(&config),
		"MeshAccessLog":             delegated.MeshAccessLog(&config),
		"MeshPassthrough":           delegated.MeshPassthrough(&config),
		"MeshTLS":                   delegated.MeshTLS(&config),
	})
	contextFor("delegated with MeshService", &configMs, map[string]func(){
		"MeshHTTPRoute": delegated.MeshHTTPRouteMeshService(&configMs),
	})
}
