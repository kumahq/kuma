package meshproxy

import (
	"context"
	"fmt"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/sync/errgroup"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	meshidentity_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	meshtrust_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshtrust/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/v2/pkg/kds/hash"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
	. "github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/client"
	"github.com/kumahq/kuma/v2/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/v2/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/v2/test/framework/deployments/zoneproxy"
	"github.com/kumahq/kuma/v2/test/framework/envoy_admin/stats"
	"github.com/kumahq/kuma/v2/test/framework/envs/multizone"
)

func Migration() {
	namespace := "meshproxy-migration"

	type crossZoneRoute struct {
		fromCluster      func() *K8sCluster
		fromClientName   string
		toCluster        func() *K8sCluster
		toServiceName    string
		expectedInstance string
	}

	type meshScenario struct {
		meshName string
		route    crossZoneRoute
	}

	scenarios := []meshScenario{
		{
			meshName: "meshproxymig-red",
			route: crossZoneRoute{
				fromCluster:      func() *K8sCluster { return multizone.KubeZone1 },
				fromClientName:   "democlient-red-zone1",
				toCluster:        func() *K8sCluster { return multizone.KubeZone2 },
				toServiceName:    "testserver-red-zone2",
				expectedInstance: "testserver-red-zone2",
			},
		},
		{
			meshName: "meshproxymig-blue",
			route: crossZoneRoute{
				fromCluster:      func() *K8sCluster { return multizone.KubeZone1 },
				fromClientName:   "democlient-blue-zone1",
				toCluster:        func() *K8sCluster { return multizone.KubeZone2 },
				toServiceName:    "testserver-blue-zone2",
				expectedInstance: "testserver-blue-zone2",
			},
		},
	}

	var (
		crossZoneAddress            func(route crossZoneRoute) string
		expectCrossZoneTraffic      func(g Gomega, route crossZoneRoute)
		assertAllMeshesReachable    func()
		deployMeshScopedZoneProxies func(meshName string)
		assertMeshScopedProxyUsed   func(scenario meshScenario)
		runDuringMigration          func(ctx context.Context, routes []crossZoneRoute, do func())
		setupMeshIdentity           func(meshName string)
	)

	It("should migrate traffic to mesh-scoped zone proxies without request drops for two meshes", func(ctx SpecContext) {
		assertAllMeshesReachable()

		routes := make([]crossZoneRoute, 0, len(scenarios))
		for _, scenario := range scenarios {
			routes = append(routes, scenario.route)
		}

		runDuringMigration(ctx, routes, func() {
			for _, scenario := range scenarios {
				deployMeshScopedZoneProxies(scenario.meshName)
				assertAllMeshesReachable()
				assertMeshScopedProxyUsed(scenario)
			}
		})
	})

	BeforeAll(func() {
		setup := NewClusterSetup()
		for _, scenario := range scenarios {
			setup.
				Install(Yaml(
					builders.Mesh().
						WithName(scenario.meshName).
						WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Exclusive),
				)).
				Install(MeshTrafficPermissionAllowAllUniversal(scenario.meshName))
		}
		Expect(setup.Setup(multizone.Global)).To(Succeed())

		for _, scenario := range scenarios {
			Expect(WaitForMesh(scenario.meshName, multizone.Zones())).To(Succeed())
		}

		kubeZone1Install := []InstallFunc{}
		kubeZone2Install := []InstallFunc{}
		for _, scenario := range scenarios {
			kubeZone1Install = append(kubeZone1Install,
				democlient.Install(
					democlient.WithNamespace(namespace),
					democlient.WithMesh(scenario.meshName),
					democlient.WithName(scenario.route.fromClientName),
				),
			)
			kubeZone2Install = append(kubeZone2Install,
				testserver.Install(
					testserver.WithNamespace(namespace),
					testserver.WithMesh(scenario.meshName),
					testserver.WithName(scenario.route.toServiceName),
					testserver.WithEchoArgs("echo", "--instance", scenario.route.expectedInstance),
				),
			)
		}

		group := errgroup.Group{}
		NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Parallel(kubeZone1Install...)).
			SetupInGroup(multizone.KubeZone1, &group)
		NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Parallel(kubeZone2Install...)).
			SetupInGroup(multizone.KubeZone2, &group)

		Expect(group.Wait()).To(Succeed())

		for _, scenario := range scenarios {
			setupMeshIdentity(scenario.meshName)
		}
	})

	AfterEachFailure(func() {
		for _, scenario := range scenarios {
			DebugUniversal(multizone.Global, scenario.meshName)
			DebugKube(multizone.KubeZone1, scenario.meshName, namespace)
			DebugKube(multizone.KubeZone2, scenario.meshName, namespace)
		}
	})

	E2EAfterAll(func() {
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.KubeZone2.TriggerDeleteNamespace(namespace)).To(Succeed())
		for _, scenario := range scenarios {
			Expect(multizone.Global.DeleteMesh(scenario.meshName)).To(Succeed())
		}
	})

	crossZoneAddress = func(route crossZoneRoute) string {
		GinkgoHelper()
		toCluster := route.toCluster()

		return fmt.Sprintf(
			"http://%s.%s.svc.%s.mesh.local:80",
			route.toServiceName,
			namespace,
			toCluster.ZoneName(),
		)
	}

	expectCrossZoneTraffic = func(g Gomega, route crossZoneRoute) {
		GinkgoHelper()
		fromCluster := route.fromCluster()

		response, err := client.CollectEchoResponse(
			fromCluster,
			route.fromClientName,
			crossZoneAddress(route),
			client.FromKubernetesPod(namespace, route.fromClientName),
		)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(response.Instance).To(Equal(route.expectedInstance))
	}

	assertAllMeshesReachable = func() {
		GinkgoHelper()

		for _, scenario := range scenarios {
			Eventually(func(g Gomega) {
				expectCrossZoneTraffic(g, scenario.route)
			}, "60s", "1s").MustPassRepeatedly(3).Should(Succeed())
		}
	}

	deployMeshScopedZoneProxies = func(meshName string) {
		GinkgoHelper()

		const ingressPort = uint32(11001)
		const egressPort = uint32(11002)

		proxyName := fmt.Sprintf("mesh-zone-proxy-%s", meshName)
		group := errgroup.Group{}
		NewClusterSetup().
			Install(zoneproxy.Install(
				zoneproxy.WithName(proxyName),
				zoneproxy.WithNamespace(namespace),
				zoneproxy.WithMesh(meshName),
				zoneproxy.WithIngressPort(ingressPort),
				zoneproxy.WithEgressPort(egressPort),
			)).
			SetupInGroup(multizone.KubeZone1, &group)
		NewClusterSetup().
			Install(zoneproxy.Install(
				zoneproxy.WithName(proxyName),
				zoneproxy.WithNamespace(namespace),
				zoneproxy.WithMesh(meshName),
				zoneproxy.WithIngressPort(ingressPort),
				zoneproxy.WithEgressPort(egressPort),
			)).
			SetupInGroup(multizone.KubeZone2, &group)
		Expect(group.Wait()).To(Succeed())
	}

	// assertMeshScopedProxyUsed verifies that the destination zone's new
	// mesh-scoped ingress receives cross-zone requests once deployed.
	assertMeshScopedProxyUsed = func(scenario meshScenario) {
		GinkgoHelper()
		toCluster := scenario.route.toCluster()

		ingressApp := fmt.Sprintf("mesh-zone-proxy-%s-ingress", scenario.meshName)
		ingressFilter := fmt.Sprintf(
			"cluster.kri_msvc_%s_%s_%s_%s_main.upstream_rq_total",
			scenario.meshName,
			toCluster.ZoneName(),
			namespace,
			scenario.route.toServiceName,
		)

		Eventually(func(g Gomega) {
			expectCrossZoneTraffic(g, scenario.route)
			stat, err := toCluster.GetEnvoyAdminTunnel(ingressApp, namespace).GetStats(ingressFilter)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeGreaterThanZero())
		}, "60s", "1s").Should(Succeed())
	}

	// runDuringMigration runs `do` while background goroutines continuously
	// send cross-zone requests and then verifies no request errors occurred.
	runDuringMigration = func(ctx context.Context, routes []crossZoneRoute, do func()) {
		GinkgoHelper()

		migrationCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		var (
			wg       sync.WaitGroup
			mu       sync.Mutex
			firstErr error
		)
		for _, route := range routes {
			wg.Add(1)
			go func() {
				defer GinkgoRecover()
				defer wg.Done()

				for {
					select {
					case <-migrationCtx.Done():
						return
					default:
					}

					fromCluster := route.fromCluster()
					_, err := client.CollectEchoResponse(
						fromCluster,
						route.fromClientName,
						crossZoneAddress(route),
						client.FromKubernetesPod(namespace, route.fromClientName),
					)
					if err != nil {
						mu.Lock()
						if firstErr == nil {
							firstErr = err
						}
						mu.Unlock()
					}

					select {
					case <-migrationCtx.Done():
						return
					case <-time.After(200 * time.Millisecond):
					}
				}
			}()
		}

		do()

		Consistently(func(g Gomega) {
			mu.Lock()
			defer mu.Unlock()
			g.Expect(firstErr).ToNot(HaveOccurred())
		}, "10s", "500ms").Should(Succeed())

		cancel()
		wg.Wait()
	}

	setupMeshIdentity = func(meshName string) {
		GinkgoHelper()

		// Use a per-mesh identity name so the bundled provider's
		// auto-generated CA secrets ("<identity>-root-ca",
		// "<identity>-private-key") don't collide across meshes in the same
		// kuma-system namespace.
		identityName := "identity-" + meshName

		meshIdentityYAML := fmt.Sprintf(`
type: MeshIdentity
name: %[2]s
mesh: %[1]s
spec:
  selector:
    dataplane:
      matchLabels: {}
  spiffeID:
    trustDomain: "{{ .Mesh }}.{{ .Zone }}.mesh.local"
  provider:
    type: Bundled
    bundled:
      meshTrustCreation: Enabled
      insecureAllowSelfSigned: true
      certificateParameters:
        expiry: 24h
      autogenerate:
        enabled: true
`, meshName, identityName)
		Expect(NewClusterSetup().
			Install(YamlUniversal(meshIdentityYAML)).
			Setup(multizone.Global)).To(Succeed())

		hashedIdentityName := hash.HashedName(meshName, identityName)
		Expect(WaitForResource(
			meshidentity_api.MeshIdentityResourceTypeDescriptor,
			model.ResourceKey{Mesh: meshName, Name: fmt.Sprintf("%s.%s", hashedIdentityName, Config.KumaNamespace)},
			multizone.KubeZone1, multizone.KubeZone2,
		)).To(Succeed())

		// Keep these helpers local to this suite.
		// If the same MeshIdentity trust bootstrap is needed in more multizone
		// suites, extract them into a shared helper under
		// test/e2e_env/multizone/meshproxy/.
		getMeshTrust := func(hashValues ...string) *meshtrust_api.MeshTrust {
			var trust *meshtrust_api.MeshTrust
			Eventually(func(g Gomega) {
				out, err := multizone.Global.GetKumactlOptions().RunKumactlAndGetOutput(
					"get", "meshtrust", "-m", meshName,
					hash.HashedName(meshName, hashedIdentityName, hashValues...),
					"-ojson",
				)
				g.Expect(err).ToNot(HaveOccurred())
				r, err := rest.JSON.Unmarshal([]byte(out), meshtrust_api.MeshTrustResourceTypeDescriptor)
				g.Expect(err).ToNot(HaveOccurred())
				trust = r.GetSpec().(*meshtrust_api.MeshTrust)
			}, "2m", "1s").Should(Succeed())
			return trust
		}

		installTrustToGlobal := func(trust *meshtrust_api.MeshTrust, sourceZoneName string) {
			yaml := builders.MeshTrust().
				WithName("meshtrust-of-zone-" + sourceZoneName).
				WithMesh(meshName).
				WithCA(trust.CABundles[0].PEM.Value).
				WithTrustDomain(trust.TrustDomain).
				UniYaml()
			Expect(NewClusterSetup().
				Install(YamlUniversal(yaml)).
				Setup(multizone.Global)).To(Succeed())
		}

		trustZone1 := getMeshTrust(multizone.KubeZone1.Name(), Config.KumaNamespace)
		installTrustToGlobal(trustZone1, multizone.KubeZone1.Name())

		trustZone2 := getMeshTrust(multizone.KubeZone2.Name(), Config.KumaNamespace)
		installTrustToGlobal(trustZone2, multizone.KubeZone2.Name())
	}
}
