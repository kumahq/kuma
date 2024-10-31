package localityawarelb

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	mesh_http_route_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	mesh_loadbalancing_api "github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancingstrategy/api/v1alpha1"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
	"github.com/kumahq/kuma/test/framework/utils"
)

func LocalityAwareLB() {
	const mesh = "locality-aware-lb"
	const namespace = "locality-aware-lb"

	BeforeAll(func() {
		// Global
		Expect(NewClusterSetup().
			Install(MTLSMeshUniversal(mesh)).
			Install(MeshTrafficPermissionAllowAllUniversal(mesh)).
			Setup(multizone.Global)).To(Succeed())
		Expect(WaitForMesh(mesh, multizone.Zones())).To(Succeed())

		group := errgroup.Group{}
		// Universal Zone 4
		NewClusterSetup().
			Install(Parallel(
				DemoClientUniversal(
					"demo-client_locality-aware-lb_svc",
					mesh,
					WithTransparentProxy(true),
					WithAdditionalTags(
						map[string]string{
							"k8s.io/node": "node-1",
							"k8s.io/az":   "az-1",
						}),
				),
				TestServerUniversal("test-server-node-1", mesh, WithAdditionalTags(
					map[string]string{
						"k8s.io/node": "node-1",
						"k8s.io/az":   "az-1",
					}),
					WithServiceName("test-server_locality-aware-lb_svc_80"),
					WithArgs([]string{"echo", "--instance", "test-server-node-1-zone-4"}),
				),
				TestServerUniversal("test-server-az-1", mesh, WithAdditionalTags(
					map[string]string{
						"k8s.io/az": "az-1",
					}),
					WithServiceName("test-server_locality-aware-lb_svc_80"),
					WithArgs([]string{"echo", "--instance", "test-server-az-1-zone-4"}),
				),
				TestServerUniversal("test-server-node-2", mesh, WithAdditionalTags(
					map[string]string{
						"k8s.io/node": "node-2",
					}),
					WithServiceName("test-server_locality-aware-lb_svc_80"),
					WithArgs([]string{"echo", "--instance", "test-server-node-2-zone-4"}),
				),
				TestServerUniversal("test-server-no-tags", mesh,
					WithArgs([]string{"echo", "--instance", "test-server-no-tags-zone-4"}),
					WithServiceName("test-server_locality-aware-lb_svc_80"),
				),
			)).
			SetupInGroup(multizone.UniZone1, &group)

		// Universal Zone 5
		NewClusterSetup().
			Install(Parallel(
				DemoClientUniversal(
					"demo-client_locality-aware-lb_svc",
					mesh,
					WithTransparentProxy(true),
				),
				TestServerUniversal("test-server", mesh,
					WithServiceName("test-server_locality-aware-lb_svc_80"),
					WithArgs([]string{"echo", "--instance", "test-server-zone-5"}),
				),
				TestServerUniversal("test-server-mesh-route", mesh,
					WithServiceName("test-server-mesh-route_locality-aware-lb_svc_80"),
					WithArgs([]string{"echo", "--instance", "test-server-mesh-route-zone-5"}),
				),
			)).
			SetupInGroup(multizone.UniZone2, &group)

		// Kubernetes Zone 1
		NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Parallel(
				democlient.Install(democlient.WithMesh(mesh), democlient.WithNamespace(namespace)),
				testserver.Install(
					testserver.WithName("test-server"),
					testserver.WithMesh(mesh),
					testserver.WithNamespace(namespace),
					testserver.WithEchoArgs("echo", "--instance", "test-server-zone-1"),
				),
				testserver.Install(
					testserver.WithName("test-server-mesh-route"),
					testserver.WithMesh(mesh),
					testserver.WithNamespace(namespace),
					testserver.WithEchoArgs("echo", "--instance", "test-server-mesh-route-zone-1"),
				),
			)).
			SetupInGroup(multizone.KubeZone1, &group)

		// Kubernetes Zone 2
		NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Parallel(
				democlient.Install(democlient.WithMesh(mesh), democlient.WithNamespace(namespace)),
				democlient.Install(
					democlient.WithMesh(mesh),
					democlient.WithNamespace(namespace),
					democlient.WithName("demo-client-mesh-route"),
				),
				testserver.Install(
					testserver.WithName("test-server"),
					testserver.WithMesh(mesh),
					testserver.WithNamespace(namespace),
					testserver.WithEchoArgs("echo", "--instance", "test-server-zone-2"),
				),
			)).
			SetupInGroup(multizone.KubeZone2, &group)
		Expect(group.Wait()).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(multizone.Global, mesh)
		DebugUniversal(multizone.UniZone1, mesh)
		DebugUniversal(multizone.UniZone2, mesh)
		DebugKube(multizone.KubeZone1, mesh, namespace)
		DebugKube(multizone.KubeZone2, mesh, namespace)
	})

	E2EAfterAll(func() {
		Expect(multizone.UniZone1.DeleteMeshApps(mesh)).To(Succeed())
		Expect(multizone.UniZone2.DeleteMeshApps(mesh)).To(Succeed())
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.KubeZone2.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(mesh)).To(Succeed())
	})

	E2EAfterEach(func() {
		Expect(DeleteMeshResources(
			multizone.Global,
			mesh,
			mesh_loadbalancing_api.MeshLoadBalancingStrategyResourceTypeDescriptor,
		)).To(Succeed())
		Expect(DeleteMeshResources(
			multizone.Global,
			mesh,
			mesh_http_route_api.MeshHTTPRouteResourceTypeDescriptor,
		)).To(Succeed())
	})

	It("should route based on defined strategy", func() {
		// should load balance traffic equally when no policy
		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(multizone.UniZone1, "demo-client_locality-aware-lb_svc", "test-server_locality-aware-lb_svc_80.mesh", client.WithNumberOfRequests(100))
		}, "30s", "5s").Should(
			And(
				HaveKeyWithValue(Equal(`test-server-node-1-zone-4`), BeNumerically("~", 20, 5)),
				HaveKeyWithValue(Equal(`test-server-az-1-zone-4`), BeNumerically("~", 20, 5)),
				HaveKeyWithValue(Equal(`test-server-node-2-zone-4`), BeNumerically("~", 20, 5)),
				HaveKeyWithValue(Equal(`test-server-no-tags-zone-4`), BeNumerically("~", 20, 5)),
				HaveKeyWithValue(Equal(`test-server-zone-5`), BeNumerically("~", 6, 2)),
				HaveKeyWithValue(Equal(`test-server-zone-1`), BeNumerically("~", 6, 2)),
				HaveKeyWithValue(Equal(`test-server-zone-2`), BeNumerically("~", 6, 2)),
			),
		)

		// when load balancing policy created
		meshLoadBalancingStrategyDemoClient := utils.FromTemplate(Default, `
type: MeshLoadBalancingStrategy
name: mlbs-1
mesh: {{.Mesh}} 
spec:
  targetRef:
    kind: MeshService
    name: demo-client_locality-aware-lb_svc
  to:
    - targetRef:
        kind: MeshService
        name: test-server_locality-aware-lb_svc_80
      default:
        localityAwareness:
          localZone:
            affinityTags:
            - key: k8s.io/node
            - key: k8s.io/az
          crossZone:
            failover:
              - from:
                  zones: ["{{ .UniZone1 }}", "{{ .UniZone2 }}"]
                to:
                  type: Only
                  zones: ["{{ .UniZone1 }}", "{{ .UniZone2 }}"]
              - from:
                  zones: ["{{ .KubeZone1 }}", "{{ .KubeZone2 }}"]
                to:
                  type: Only
                  zones: ["{{ .KubeZone1 }}", "{{ .KubeZone2 }}"]
              - from:
                  zones: ["{{ .UniZone1 }}"]
                to:
                  type: Only
                  zones: ["{{ .KubeZone1 }}"]`,
			multizone.ZoneInfoForMesh(mesh))
		Expect(NewClusterSetup().Install(YamlUniversal(meshLoadBalancingStrategyDemoClient)).Setup(multizone.Global)).To(Succeed())

		// then traffic should be routed based on the policy
		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(multizone.UniZone1, "demo-client_locality-aware-lb_svc", "test-server_locality-aware-lb_svc_80.mesh", client.WithNumberOfRequests(100))
		}, "30s", "5s").Should(
			And(
				HaveKeyWithValue(Equal(`test-server-node-1-zone-4`), BeNumerically("~", 93, 5)),
				HaveKeyWithValue(Equal(`test-server-az-1-zone-4`), BeNumerically("~", 7, 5)),
				Or(
					HaveKeyWithValue(Equal(`test-server-node-2-zone-4`), BeNumerically("~", 2, 2)),
					HaveKeyWithValue(Equal(`test-server-no-tags-zone-4`), BeNumerically("~", 2, 2)),
					Not(HaveKeyWithValue(Equal(`test-server-node-2-zone-4`), BeNumerically("~", 2, 2))),
					Not(HaveKeyWithValue(Equal(`test-server-no-tags-zone-4`), BeNumerically("~", 2, 2))),
				),
			),
		)

		// when app with the highest weight is down
		Expect(multizone.UniZone1.DeleteApp("test-server-node-1")).To(Succeed())

		// then traffic goes to the next highest
		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(multizone.UniZone1, "demo-client_locality-aware-lb_svc", "test-server_locality-aware-lb_svc_80.mesh", client.WithNumberOfRequests(100))
		}, "30s", "5s").Should(
			And(
				HaveKeyWithValue(Equal(`test-server-az-1-zone-4`), BeNumerically("~", 90, 5)),
				Or(
					HaveKeyWithValue(Equal(`test-server-node-2-zone-4`), BeNumerically("~", 5, 3)),
					HaveKeyWithValue(Equal(`test-server-no-tags-zone-4`), BeNumerically("~", 5, 3)),
				),
			),
		)

		// when the next app with the highest weight is down
		Expect(multizone.UniZone1.DeleteApp("test-server-az-1")).To(Succeed())

		// then traffic goes to the next highest
		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(multizone.UniZone1, "demo-client_locality-aware-lb_svc", "test-server_locality-aware-lb_svc_80.mesh", client.WithNumberOfRequests(100))
		}, "30s", "5s").Should(
			And(
				HaveKeyWithValue(Equal(`test-server-node-2-zone-4`), BeNumerically("~", 50, 5)),
				HaveKeyWithValue(Equal(`test-server-no-tags-zone-4`), BeNumerically("~", 50, 5)),
			),
		)

		// when all apps in local zone down
		Expect(multizone.UniZone1.DeleteApp("test-server-node-2")).To(Succeed())
		Expect(multizone.UniZone1.DeleteApp("test-server-no-tags")).To(Succeed())

		// then traffic goes to the zone with the next priority, kuma-5
		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(multizone.UniZone1, "demo-client_locality-aware-lb_svc", "test-server_locality-aware-lb_svc_80.mesh", client.WithNumberOfRequests(100))
		}, "30s", "5s").Should(HaveKeyWithValue(Equal(`test-server-zone-5`), BeNumerically("==", 100)))

		// when zone kuma-5 is disabled
		Expect(multizone.UniZone2.DeleteApp("test-server")).To(Succeed())

		// then traffic should goes to k8s
		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(multizone.UniZone1, "demo-client_locality-aware-lb_svc", "test-server_locality-aware-lb_svc_80.mesh", client.WithNumberOfRequests(100))
		}, "30s", "5s").Should(HaveKeyWithValue(Equal(`test-server-zone-1`), BeNumerically("==", 100)))

		// when zone kuma-1-zone is disabled
		Expect(DeleteK8sApp(multizone.KubeZone1, "test-server", namespace)).To(Succeed())

		// then traffic should goes to k8s
		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				multizone.UniZone1, "demo-client_locality-aware-lb_svc", "test-server_locality-aware-lb_svc_80.mesh",
			)
			g.Expect(err).To(HaveOccurred())
		}).Should(Succeed())
	})

	It("should route based on the strategy when split defined", FlakeAttempts(3), func() {
		// given
		Expect(YamlUniversal(fmt.Sprintf(`
type: MeshHTTPRoute
name: route-1
mesh: %s
spec:
  targetRef:
    kind: MeshService
    name: demo-client-mesh-route_locality-aware-lb_svc
  to:
    - targetRef:
        kind: MeshService
        name: test-server-mesh-route_locality-aware-lb_svc_80
      rules:
        - matches:
          - path:
              value: /
              type: PathPrefix
          default:
            backendRefs:
              - kind: MeshServiceSubset
                name: test-server-mesh-route_locality-aware-lb_svc_80
                weight: 50
                tags:
                  kuma.io/zone: "%s" 
              - kind: MeshServiceSubset
                name: test-server-mesh-route_locality-aware-lb_svc_80
                weight: 50
                tags:
                  kuma.io/zone: "%s"
`, mesh, multizone.UniZone2.ZoneName(), multizone.KubeZone1.ZoneName()))(multizone.Global)).To(Succeed())

		// then traffic should be routed equally
		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(
				multizone.KubeZone2, "demo-client-mesh-route", "test-server-mesh-route_locality-aware-lb_svc_80.mesh",
				client.WithNumberOfRequests(200),
				client.FromKubernetesPod(namespace, "demo-client-mesh-route"),
			)
		}, "30s", "10s").Should(
			And(
				HaveKeyWithValue(Equal(`test-server-mesh-route-zone-1`), BeNumerically("~", 100, 20)),
				HaveKeyWithValue(Equal(`test-server-mesh-route-zone-5`), BeNumerically("~", 100, 20)),
			),
		)

		// clean stats
		Expect(resetCounter(multizone.KubeZone2, "demo-client-mesh-route", namespace)).To(Succeed())

		// when load balancing policy created
		meshLoadBalancingStrategyDemoClientMeshRoute := utils.FromTemplate(Default, `
type: MeshLoadBalancingStrategy
name: mlbs-2
mesh: "{{ .Mesh }}"
spec:
  targetRef:
    kind: MeshService
    name: demo-client-mesh-route_{{ .Mesh }}_svc
  to:
    - targetRef:
        kind: MeshService
        name: test-server-mesh-route_{{ .Mesh }}_svc_80
      default:
        localityAwareness:
          crossZone:
            failover:
              - to:
                  type: Only
                  zones: ["{{ .KubeZone1 }}"]`, multizone.ZoneInfoForMesh(mesh))

		Expect(NewClusterSetup().Install(YamlUniversal(meshLoadBalancingStrategyDemoClientMeshRoute)).Setup(multizone.Global)).To(Succeed())

		// and generate some traffic
		_, err := client.CollectResponsesAndFailures(
			multizone.KubeZone2, "demo-client-mesh-route", "test-server-mesh-route_locality-aware-lb_svc_80.mesh",
			client.WithNumberOfRequests(200),
			client.NoFail(),
			client.WithoutRetries(),
			client.FromKubernetesPod(namespace, "demo-client-mesh-route"),
		)
		Expect(err).ToNot(HaveOccurred())

		// then
		failedRequests, err := collectMetric(multizone.KubeZone2, "demo-client-mesh-route", namespace, "test-server-mesh-route_locality-aware-lb_svc_80-ce3d32a0f959e460.upstream_rq_5xx")
		Expect(err).ToNot(HaveOccurred())
		Expect(failedRequests).To(BeNumerically("~", 100, 25))
		successRequests, err := collectMetric(multizone.KubeZone2, "demo-client-mesh-route", namespace, "test-server-mesh-route_locality-aware-lb_svc_80-70a8d85bc2519528.upstream_rq_2xx")
		Expect(err).ToNot(HaveOccurred())
		Expect(successRequests).To(BeNumerically("~", 100, 25))
	})
}

func collectMetric(cluster Cluster, name string, namespace string, metricName string) (int, error) {
	resp, _, err := client.CollectResponse(cluster, name, fmt.Sprintf("http://localhost:9901/stats?filter=%s", metricName), client.FromKubernetesPod(namespace, name))
	if err != nil {
		return -1, err
	}
	split := strings.Split(resp, ": ")
	if len(split) == 2 {
		i, err := strconv.Atoi(split[1])
		if err != nil {
			return -1, err
		}
		return i, nil
	} else {
		return -1, errors.New("no metric found")
	}
}

func resetCounter(cluster Cluster, name string, namespace string) error {
	_, _, err := client.CollectResponse(
		cluster, name, "http://localhost:9901/reset_counters",
		client.FromKubernetesPod(namespace, name),
		client.WithMethod("POST"),
	)
	return err
}

// TODO(lukidzi): use test-server implementation: https://github.com/kumahq/kuma/issues/8245
func DeleteK8sApp(c Cluster, name string, namespace string) error {
	if err := k8s.RunKubectlE(c.GetTesting(), c.GetKubectlOptions(namespace), "delete", "service", name); err != nil {
		return err
	}
	if err := k8s.RunKubectlE(c.GetTesting(), c.GetKubectlOptions(namespace), "delete", "deployment", name); err != nil {
		return err
	}
	return nil
}
