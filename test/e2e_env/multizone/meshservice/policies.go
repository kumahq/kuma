package meshservice

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshfaultinjection_api "github.com/kumahq/kuma/pkg/plugins/policies/meshfaultinjection/api/v1alpha1"
	meshhealthcheck_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhealthcheck/api/v1alpha1"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshretry_api "github.com/kumahq/kuma/pkg/plugins/policies/meshretry/api/v1alpha1"
	meshtcproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtcproute/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envoy_admin"
	"github.com/kumahq/kuma/test/framework/envoy_admin/stats"
	"github.com/kumahq/kuma/test/framework/envoy_admin/tunnel"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func MeshServiceTargeting() {
	meshName := "real-resource-mesh"
	namespace := "real-resource-ns"
	addressSuffix := "realreasource"
	addressToMeshService := func(service string) string {
		return fmt.Sprintf("%s.%s.%s.%s", service, namespace, Kuma1, addressSuffix)
	}

	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(ResourceUniversal(samples.MeshMTLSBuilder().WithName(meshName).WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Everywhere).Build())).
			Install(MeshTrafficPermissionAllowAllUniversal(meshName)).
			Setup(multizone.Global)).To(Succeed())
		Expect(WaitForMesh(meshName, multizone.Zones())).To(Succeed())

		err := NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Parallel(
				testserver.Install(
					testserver.WithName("test-client"),
					testserver.WithMesh(meshName),
					testserver.WithNamespace(namespace),
				),
				testserver.Install(
					testserver.WithName("test-server"),
					testserver.WithMesh(meshName),
					testserver.WithNamespace(namespace),
				),
				testserver.Install(
					testserver.WithName("second-test-server"),
					testserver.WithMesh(meshName),
					testserver.WithNamespace(namespace),
				),
				testserver.Install(
					testserver.WithName("kumaioservice-targeted-test-server"),
					testserver.WithMesh(meshName),
					testserver.WithNamespace(namespace),
				),
			)).
			Install(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: HostnameGenerator
metadata:
  name: e2e-connectivity
  namespace: %s
  labels:
    kuma.io/origin: zone
spec:
  template: '{{ .DisplayName }}.{{ .Namespace }}.{{ .Zone }}.%s'
  selector:
    meshService:
      matchLabels:
        kuma.io/origin: zone
        kuma.io/managed-by: k8s-controller
`, Config.KumaNamespace, addressSuffix))).
			Setup(multizone.KubeZone1)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(testserver.Install(
				testserver.WithName("test-server"),
				testserver.WithMesh(meshName),
				testserver.WithNamespace(namespace),
				testserver.WithEchoArgs("echo", "--instance", "kube-test-server-2"),
			)).Setup(multizone.KubeZone2)
		Expect(err).ToNot(HaveOccurred())

		// remove default retry policy
		Expect(DeleteMeshResources(multizone.Global, meshName, meshretry_api.MeshRetryResourceTypeDescriptor)).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(multizone.Global, meshName)
		DebugKube(multizone.KubeZone1, meshName)
	})

	E2EAfterEach(func() {
		Expect(DeleteMeshResources(multizone.KubeZone1, meshName, meshhttproute_api.MeshHTTPRouteResourceTypeDescriptor)).To(Succeed())
		Expect(DeleteMeshResources(multizone.KubeZone1, meshName, meshtcproute_api.MeshTCPRouteResourceTypeDescriptor)).To(Succeed())
		Expect(DeleteMeshResources(multizone.KubeZone1, meshName, meshhealthcheck_api.MeshHealthCheckResourceTypeDescriptor)).To(Succeed())
		Expect(DeleteMeshResources(multizone.KubeZone1, meshName, core_mesh.ExternalServiceResourceTypeDescriptor)).To(Succeed())
		Expect(DeleteMeshResources(multizone.KubeZone2, meshName, meshfaultinjection_api.MeshFaultInjectionResourceTypeDescriptor)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.KubeZone2.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(meshName)).To(Succeed())
	})

	retryStat := func(admin envoy_admin.Tunnel) *stats.Stats {
		s, err := admin.GetStats("cluster.real-resource-mesh_test-server_real-resource-ns_kuma-2_msvc_80.upstream_rq_retry_success")
		Expect(err).ToNot(HaveOccurred())
		return s
	}

	countResponseCodes := func(statusCode int) func(responses []client.FailureResponse) int {
		return func(responses []client.FailureResponse) int {
			count := 0
			for _, r := range responses {
				if r.ResponseCode == statusCode {
					count++
				}
			}
			return count
		}
	}

	It("should configure URLRewrite", func() {
		// when
		Expect(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: http-to-real-meshservice
  namespace: %s
  labels:
    kuma.io/mesh: %s
    kuma.io/origin: zone
spec:
  to:
    - targetRef:
        kind: MeshService
        name: test-server
        sectionName: main
      rules:
        - matches:
            - path:
                type: PathPrefix
                value: /prefix
          default:
            filters:
              - type: URLRewrite
                urlRewrite:
                  path:
                    type: ReplacePrefixMatch
                    replacePrefixMatch: /hello/
`, namespace, meshName))(multizone.KubeZone1)).To(Succeed())
		// then receive redirect response
		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				multizone.KubeZone1,
				"test-client",
				fmt.Sprintf("%s/prefix/world", addressToMeshService("test-server")),
				client.FromKubernetesPod(namespace, "test-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Received.Path).To(Equal("/hello/world"))
		}, "30s", "1s").Should(Succeed())
	})

	It("should configure URLRewrite if targeted to kuma.io/service", func() {
		// when
		Expect(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: http-to-kumaio-meshservice
  namespace: %s
  labels:
    kuma.io/mesh: %s
    kuma.io/origin: zone
spec:
  to:
    - targetRef:
        kind: MeshService
        name: kumaioservice-targeted-test-server_%s_svc_80
      rules:
        - matches:
            - path:
                type: PathPrefix
                value: /prefix
          default:
            filters:
              - type: URLRewrite
                urlRewrite:
                  path:
                    type: ReplacePrefixMatch
                    replacePrefixMatch: /hello/old/
`, namespace, meshName, namespace))(multizone.KubeZone1)).To(Succeed())
		// then receive redirect response
		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				multizone.KubeZone1,
				"test-client",
				fmt.Sprintf("%s/prefix/world", addressToMeshService("kumaioservice-targeted-test-server")),
				client.FromKubernetesPod(namespace, "test-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Received.Path).To(Equal("/hello/old/world"))
		}, "30s", "1s").Should(Succeed())
	})

	It("should route to second test server", func() {
		// given
		// should route to original MeshService
		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(
				multizone.KubeZone1,
				"test-client",
				addressToMeshService("test-server"),
				client.FromKubernetesPod(namespace, "test-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(HavePrefix("test-server"))
		}, "30s", "1s").Should(Succeed())

		// when
		Expect(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTCPRoute
metadata:
  name: tcp-overwrite-to-real-meshservice
  namespace: %s
  labels:
    kuma.io/mesh: %s
    kuma.io/origin: zone
spec:
  to:
  - targetRef:
      kind: MeshService
      name: test-server
      sectionName: main
    rules:
    - default:
        backendRefs:
        - kind: MeshService
          name: second-test-server
          port: 80
`, namespace, meshName))(multizone.KubeZone1)).To(Succeed())

		// then
		// should route to second MeshService
		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(
				multizone.KubeZone1,
				"test-client",
				addressToMeshService("test-server"),
				client.FromKubernetesPod(namespace, "test-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(HavePrefix("second-test-server"))
		}, "30s", "1s").Should(Succeed())
	})

	It("should configure MeshRetry for MeshService in another zone", func() {
		Expect(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshFaultInjection
metadata:
  name: mesh-fi-ms-another-zone
  namespace: %s
  labels:
    kuma.io/mesh: %s
    kuma.io/origin: zone
spec:
  targetRef:
    kind: Mesh
    proxyTypes: ["Sidecar"]
  from:
    - targetRef:
        kind: Mesh
      default:
          http:
            - abort:
                httpStatus: 503
                percentage: "50"
`, Config.KumaNamespace, meshName))(multizone.KubeZone2)).To(Succeed())
		// given
		// create a tunnel to test-client admin
		portFwd, err := multizone.KubeZone1.PortForwardApp("test-client", namespace, 9901)
		Expect(err).ToNot(HaveOccurred())

		adminTunnel, err := tunnel.NewK8sEnvoyAdminTunnel(multizone.Global.GetTesting(), portFwd.Endpoint)
		Expect(err).ToNot(HaveOccurred())
		// then
		Eventually(func() ([]client.FailureResponse, error) {
			return client.CollectResponsesAndFailures(
				multizone.KubeZone1,
				"test-client",
				fmt.Sprintf("test-server.%s.svc.%s.mesh.local", namespace, Kuma2),
				client.FromKubernetesPod(namespace, "test-client"),
				client.WithNumberOfRequests(200),
			)
		}, "30s", "5s").Should(And(
			HaveLen(200),
			WithTransform(countResponseCodes(503), BeNumerically("~", 100, 30)),
			WithTransform(countResponseCodes(200), BeNumerically("~", 100, 30)),
		))

		Eventually(func(g Gomega) {
			g.Expect(retryStat(adminTunnel)).To(stats.BeEqualZero())
		}, "5s", "1s").Should(Succeed())

		meshRetryPolicy := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshRetry
metadata:
  name: mr-for-ms-in-zone-2
  namespace: %s
  labels:
    kuma.io/mesh: %s
    kuma.io/origin: zone
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshService
        labels:
          kuma.io/display-name: test-server
          kuma.io/zone: %s
      default:
        http:
          numRetries: 6
          retryOn: ["503"]
`, Config.KumaNamespace, meshName, Kuma2)
		// and when a MeshRetry policy is applied
		Expect(YamlK8s(meshRetryPolicy)(multizone.KubeZone1)).To(Succeed())

		// then
		Eventually(func() ([]client.FailureResponse, error) {
			return client.CollectResponsesAndFailures(
				multizone.KubeZone1,
				"test-client",
				fmt.Sprintf("test-server.%s.svc.%s.mesh.local", namespace, Kuma2),
				client.FromKubernetesPod(namespace, "test-client"),
				client.WithNumberOfRequests(200),
			)
		}, "30s", "5s").Should(And(
			HaveLen(200),
			ContainElements(HaveField("ResponseCode", 200)),
		))

		// and
		Expect(retryStat(adminTunnel)).To(stats.BeGreaterThanZero())

		// remove Policies
		Expect(DeleteMeshPolicyOrError(multizone.KubeZone1, meshretry_api.MeshRetryResourceTypeDescriptor, "mr-for-ms-in-zone-2")).To(Succeed())
	})

	It("should mark MeshService from another zone as unhealthy if it doesn't reply on health checks", func() {
		// check that test-server is healthy
		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				multizone.KubeZone1,
				"test-client",
				fmt.Sprintf("test-server.%s.svc.%s.mesh.local", namespace, Kuma2),
				client.FromKubernetesPod(namespace, "test-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "30s", "1s").MustPassRepeatedly(3).Should(Succeed())

		healthCheck := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshHealthCheck
metadata:
  name: mhc-for-ms-in-kuma-2
  namespace: %s
  labels:
    kuma.io/mesh: %s
    kuma.io/origin: zone
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshService
        labels:
          kuma.io/display-name: test-server
          kuma.io/zone: %s
      default:
        interval: 3s
        timeout: 2s
        unhealthyThreshold: 3
        healthyThreshold: 1
        failTrafficOnPanic: true
        noTrafficInterval: 1s
        healthyPanicThreshold: 0
        reuseConnection: true
        http:
          path: /are-you-healthy
          expectedStatuses:
          - 500`, Config.KumaNamespace, meshName, Kuma2)
		// update HealthCheck policy to check for another status code
		Expect(YamlK8s(healthCheck)(multizone.KubeZone1)).To(Succeed())

		// check that test-server is unhealthy
		Eventually(func(g Gomega) {
			response, err := client.CollectFailure(
				multizone.KubeZone1,
				"test-client",
				fmt.Sprintf("test-server.%s.svc.%s.mesh.local", namespace, Kuma2),
				client.FromKubernetesPod(namespace, "test-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.ResponseCode).To(Equal(503))
		}, "30s", "1s").MustPassRepeatedly(3).Should(Succeed())
	})
}
