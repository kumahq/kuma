package producer

import (
	"fmt"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/sync/errgroup"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/kds/hash"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshtimeout_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	. "github.com/kumahq/kuma/test/framework"
	framework_client "github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func ProducerPolicyFlow() {
	const mesh = "producer-policy-flow"
	const k8sZoneNamespace = "producer-policy-flow-ns"

	BeforeAll(func() {
		// Global
		Expect(NewClusterSetup().
			Install(
				Yaml(
					builders.Mesh().
						WithName(mesh).
						WithoutInitialPolicies().
						WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Exclusive).
						WithBuiltinMTLSBackend("ca-1").WithEnabledMTLSBackend("ca-1"),
				),
			).
			Install(MeshTrafficPermissionAllowAllUniversal(mesh)).
			Setup(multizone.Global)).To(Succeed())
		Expect(WaitForMesh(mesh, multizone.Zones())).To(Succeed())

		group := errgroup.Group{}
		// Kube Zone 1
		NewClusterSetup().
			Install(NamespaceWithSidecarInjection(k8sZoneNamespace)).
			Install(testserver.Install(
				testserver.WithName("test-client"),
				testserver.WithMesh(mesh),
				testserver.WithNamespace(k8sZoneNamespace),
			)).
			SetupInGroup(multizone.KubeZone1, &group)

		NewClusterSetup().
			Install(NamespaceWithSidecarInjection(k8sZoneNamespace)).
			Install(testserver.Install(
				testserver.WithName("test-server"),
				testserver.WithMesh(mesh),
				testserver.WithNamespace(k8sZoneNamespace),
				testserver.WithEchoArgs("echo", "--instance", "kube-test-server-2"),
			)).
			SetupInGroup(multizone.KubeZone2, &group)
		Expect(group.Wait()).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(multizone.Global, mesh)
		DebugKube(multizone.KubeZone1, mesh, k8sZoneNamespace)
		DebugKube(multizone.KubeZone2, mesh, k8sZoneNamespace)
	})

	E2EAfterEach(func() {
		Expect(DeleteMeshResources(multizone.Global, mesh, meshtimeout_api.MeshTimeoutResourceTypeDescriptor)).To(Succeed())
		Expect(DeleteMeshResources(multizone.KubeZone1, mesh, meshtimeout_api.MeshTimeoutResourceTypeDescriptor)).To(Succeed())
		Expect(DeleteMeshResources(multizone.KubeZone2, mesh, meshtimeout_api.MeshTimeoutResourceTypeDescriptor)).To(Succeed())
		Expect(DeleteMeshResources(multizone.Global, mesh, meshhttproute_api.MeshHTTPRouteResourceTypeDescriptor)).To(Succeed())
		Expect(DeleteMeshResources(multizone.KubeZone1, mesh, meshhttproute_api.MeshHTTPRouteResourceTypeDescriptor)).To(Succeed())
		Expect(DeleteMeshResources(multizone.KubeZone2, mesh, meshhttproute_api.MeshHTTPRouteResourceTypeDescriptor)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(k8sZoneNamespace)).To(Succeed())
		Expect(multizone.KubeZone2.TriggerDeleteNamespace(k8sZoneNamespace)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(mesh)).To(Succeed())
	})

	It("should sync producer policy to other clusters", func() {
		Expect(YamlK8s(fmt.Sprintf(`
kind: MeshTimeout
apiVersion: kuma.io/v1alpha1
metadata:
  name: to-test-server
  namespace: %s
  labels:
    kuma.io/origin: zone
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshService
        name: test-server
      default:
        http:
          requestTimeout: 2s
`, k8sZoneNamespace, mesh))(multizone.KubeZone2)).To(Succeed())

		Eventually(func(g Gomega) {
			out, err := k8s.RunKubectlAndGetOutputE(
				multizone.KubeZone1.GetTesting(),
				multizone.KubeZone1.GetKubectlOptions(Config.KumaNamespace),
				"get", "meshtimeouts")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(out).To(ContainSubstring(hash.HashedName(mesh, "to-test-server", Kuma2, k8sZoneNamespace)))
		}).Should(Succeed())

		Eventually(func(g Gomega) {
			response, err := framework_client.CollectFailure(
				multizone.KubeZone1, "test-client", fmt.Sprintf("test-server.%s.svc.kuma-2.mesh.local", k8sZoneNamespace),
				framework_client.FromKubernetesPod(k8sZoneNamespace, "test-client"),
				framework_client.WithHeader("x-set-response-delay-ms", "5000"),
				framework_client.WithMaxTime(10), // we don't want 'curl' to return early
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.ResponseCode).To(Equal(504))
		}, "1m", "1s", MustPassRepeatedly(5)).Should(Succeed())

		Expect(YamlK8s(fmt.Sprintf(`
kind: MeshTimeout
apiVersion: kuma.io/v1alpha1
metadata:
  name: to-test-server
  namespace: %s
  labels:
    kuma.io/origin: zone
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshService
        name: test-server
        namespace: random-ns-name
      default:
        http:
          requestTimeout: 2s
`, k8sZoneNamespace, mesh))(multizone.KubeZone2)).To(Succeed())

		Eventually(func(g Gomega) {
			out, err := k8s.RunKubectlAndGetOutputE(
				multizone.KubeZone1.GetTesting(),
				multizone.KubeZone1.GetKubectlOptions(Config.KumaNamespace),
				"get", "meshtimeouts")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(out).ToNot(ContainSubstring(hash.HashedName(mesh, "to-test-server", Kuma2, k8sZoneNamespace)))
		}).Should(Succeed())
	})

	It("should sync producer MeshTimeout that targets producer route to other clusters", func() {
		Expect(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: add-response-delay-header
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  to:
    - targetRef:
        kind: MeshService
        name: test-server
      rules:
        - matches:
            - path:
                type: PathPrefix
                value: /
          default:
            filters:
              - type: RequestHeaderModifier
                requestHeaderModifier:
                  add:
                    - name: x-set-response-delay-ms
                      value: "3000"
`, k8sZoneNamespace, mesh))(multizone.KubeZone2)).To(Succeed())

		// check that MeshHTTPRoute 'add-response-delay-header' makes response time more than 3s
		Eventually(func(g Gomega) {
			start := time.Now()
			g.Expect(framework_client.CollectEchoResponse(
				multizone.KubeZone1, "test-client", fmt.Sprintf("test-server.%s.svc.kuma-2.mesh.local", k8sZoneNamespace),
				framework_client.FromKubernetesPod(k8sZoneNamespace, "test-client"),
				framework_client.WithMaxTime(10), // we don't want 'curl' to return early
			)).Should(HaveField("Instance", ContainSubstring("test-server")))
			g.Expect(time.Since(start)).To(BeNumerically(">", time.Second*3))
		}, "30s", "1s").Should(Succeed())

		Expect(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: timeout-on-http-route
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  to:
    - targetRef:
        kind: MeshHTTPRoute
        name: add-response-delay-header
      default:
        http:
          requestTimeout: 2s
`, k8sZoneNamespace, mesh))(multizone.KubeZone2)).To(Succeed())

		// check 'timeout-on-http-route' synced to test-client's zone
		Eventually(func(g Gomega) {
			out, err := k8s.RunKubectlAndGetOutputE(
				multizone.KubeZone1.GetTesting(),
				multizone.KubeZone1.GetKubectlOptions(Config.KumaNamespace),
				"get", "meshtimeouts")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(out).To(ContainSubstring(hash.HashedName(mesh, "timeout-on-http-route", Kuma2, k8sZoneNamespace)))
		}).Should(Succeed())

		// check 'timeout-on-http-route' is applied
		Eventually(func(g Gomega) {
			response, err := framework_client.CollectFailure(
				multizone.KubeZone1, "test-client", fmt.Sprintf("test-server.%s.svc.kuma-2.mesh.local", k8sZoneNamespace),
				framework_client.FromKubernetesPod(k8sZoneNamespace, "test-client"),
				framework_client.WithMaxTime(10), // we don't want 'curl' to return early
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.ResponseCode).To(Equal(504))
		}, "1m", "1s", MustPassRepeatedly(5)).Should(Succeed())
	})

	It("should sync producer MeshRetry that targets producer route to other clusters", func() {
		// Function CollectResponsesAndFailures sends requests concurrently. If we want to accurately check
		// that MeshRetry works we need to increase 'maxRetries' using MeshCircuitBreaker policy.
		Expect(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshCircuitBreaker
metadata:
  name: increased-max-retries
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  to:
  - targetRef:
      kind: MeshService
      name: test-server
    default:
      connectionLimits:
        maxRetries: 20`, k8sZoneNamespace, mesh))(multizone.KubeZone2)).Should(Succeed())

		Expect(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshFaultInjection
metadata:
  name: mesh-fault-injecton
  namespace: %s
  labels:
    kuma.io/mesh: "%s"
spec:
  targetRef:
    kind: Dataplane
    labels:
      app: test-server
  from:
    - targetRef:
        kind: Mesh
      default:
        http:
          - abort:
              httpStatus: 500
              percentage: "50.0"
`, k8sZoneNamespace, mesh))(multizone.KubeZone2)).Should(Succeed())

		Expect(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: to-test-server
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  to:
    - targetRef:
        kind: MeshService
        name: test-server
      rules:
        - matches:
            - path:
                type: PathPrefix
                value: /
          default:
            backendRefs:
              - kind: MeshService
                name: test-server
                port: 80
`, k8sZoneNamespace, mesh))(multizone.KubeZone2)).To(Succeed())

		Eventually(func(g Gomega) {
			responses, err := framework_client.CollectResponsesAndFailures(
				multizone.KubeZone1, "test-client", fmt.Sprintf("test-server.%s.svc.kuma-2.mesh.local", k8sZoneNamespace),
				framework_client.FromKubernetesPod(k8sZoneNamespace, "test-client"),
				framework_client.WithNumberOfRequests(100),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses).To(And(
				HaveLen(100),
				WithTransform(framework_client.CountResponseCodes(500), BeNumerically("~", 50, 15)),
				WithTransform(framework_client.CountResponseCodes(200), BeNumerically("~", 50, 15)),
			))
		}, "30s", "5s").Should(Succeed())

		Expect(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshRetry
metadata:
  name: retry-on-5xx
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  to:
    - targetRef:
        kind: MeshHTTPRoute
        name: to-test-server
      default:
        http:
          numRetries: 5
          retryOn:
            - "5xx"
`, k8sZoneNamespace, mesh))(multizone.KubeZone2)).To(Succeed())

		Eventually(func(g Gomega) {
			responses, err := framework_client.CollectResponsesAndFailures(
				multizone.KubeZone1, "test-client", fmt.Sprintf("test-server.%s.svc.kuma-2.mesh.local", k8sZoneNamespace),
				framework_client.FromKubernetesPod(k8sZoneNamespace, "test-client"),
				framework_client.WithNumberOfRequests(100),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses).To(And(
				HaveLen(100),
				WithTransform(framework_client.CountResponseCodes(200), BeNumerically("~", 100, 10)),
			))
		}, "30s", "5s").Should(Succeed())
	})
}
