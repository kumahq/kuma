package producer

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/sync/errgroup"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/kri"
	"github.com/kumahq/kuma/pkg/kds/hash"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshtimeout_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/api"
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
			out := meshtimeout_api.NewMeshTimeoutResource()
			res := api.FetchResourceByKri(g, multizone.KubeZone1, out, kri.MustFromString("kri_mt_producer-policy-flow_kuma-2_producer-policy-flow-ns_to-test-server_"))
			g.Expect(res.StatusCode).To(Equal(http.StatusOK))
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

		// should not get synced to zone 1
		Eventually(func(g Gomega) {
			out, err := k8s.RunKubectlAndGetOutputE(
				multizone.KubeZone1.GetTesting(),
				multizone.KubeZone1.GetKubectlOptions(Config.KumaNamespace),
				"get", "meshtimeouts")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(out).ToNot(ContainSubstring(hash.HashedName(mesh, "to-test-server", Kuma2, k8sZoneNamespace)))
		}).Should(Succeed())

		// should be available via KRI in zone 2
		Eventually(func(g Gomega) {
			out := meshtimeout_api.NewMeshTimeoutResource()
			res := api.FetchResourceByKri(g, multizone.KubeZone2, out, kri.MustFromString("kri_mt_producer-policy-flow_kuma-2_producer-policy-flow-ns_to-test-server_"))
			g.Expect(res.StatusCode).To(Equal(http.StatusOK))
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
                value: /with-retry
          default:
            backendRefs:
              - kind: MeshService
                name: test-server
                port: 80
`, k8sZoneNamespace, mesh))(multizone.KubeZone2)).To(Succeed())

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
            - "503"
`, k8sZoneNamespace, mesh))(multizone.KubeZone2)).To(Succeed())

		lastId := 0
		generateNewId := func() string {
			lastId++
			return fmt.Sprintf("%d", lastId)
		}

		// sending requests to / results in 503
		Eventually(func(g Gomega) {
			response, err := framework_client.CollectFailure(
				multizone.KubeZone1, "test-client", fmt.Sprintf("test-server.%s.svc.kuma-2.mesh.local", k8sZoneNamespace),
				framework_client.FromKubernetesPod(k8sZoneNamespace, "test-client"),
				framework_client.WithHeader("x-succeed-after-n", "5"),
				framework_client.WithHeader("x-succeed-after-n-id", generateNewId()),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.ResponseCode).To(Equal(503))
		}, "1m", "1s").Should(Succeed())

		// sending requests to /with-retry succeeds
		Eventually(func(g Gomega) {
			_, err := framework_client.CollectEchoResponse(
				multizone.KubeZone1, "test-client", fmt.Sprintf("test-server.%s.svc.kuma-2.mesh.local/with-retry", k8sZoneNamespace),
				framework_client.FromKubernetesPod(k8sZoneNamespace, "test-client"),
				framework_client.WithHeader("x-succeed-after-n", "5"),
				framework_client.WithHeader("x-succeed-after-n-id", generateNewId()),
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "1m", "1s").Should(Succeed())
	})
}
