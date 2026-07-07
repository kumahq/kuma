package meshproxy

import (
	"context"
	"fmt"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/sync/errgroup"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	meshidentity_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	meshtrust_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshtrust/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/v3/pkg/kds/hash"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	. "github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/client"
	"github.com/kumahq/kuma/v3/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/v3/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/v3/test/framework/deployments/zoneproxy"
	"github.com/kumahq/kuma/v3/test/framework/envoy_admin/stats"
	"github.com/kumahq/kuma/v3/test/framework/envs/multizone"
)

func Migration() {
	namespace := "meshproxy-migration"

	const (
		redMeshName  = "meshproxymig-red"
		blueMeshName = "meshproxymig-blue"

		redClientZone1  = "democlient-red-zone1"
		blueClientZone1 = "democlient-blue-zone1"

		redServerZone2  = "testserver-red-zone2"
		blueServerZone2 = "testserver-blue-zone2"
	)

	type trafficCheck struct {
		meshName          string
		clientName        string
		destinationServer string
		expectedInstance  string
	}

	redTraffic := trafficCheck{
		meshName:          redMeshName,
		clientName:        redClientZone1,
		destinationServer: redServerZone2,
		expectedInstance:  redServerZone2,
	}
	blueTraffic := trafficCheck{
		meshName:          blueMeshName,
		clientName:        blueClientZone1,
		destinationServer: blueServerZone2,
		expectedInstance:  blueServerZone2,
	}

	var (
		crossZoneAddress            func(check trafficCheck) string
		collectInstances            func(check trafficCheck, requests int) (map[string]int, error)
		assertCrossZoneTraffic      func(g Gomega, check trafficCheck)
		assertTrafficForBothMeshes  func()
		deployMeshScopedZoneProxies func(meshName string)
		assertMeshScopedProxyUsed   func(check trafficCheck)
		runDuringMigration          func(ctx context.Context, checks []trafficCheck, do func())
		setupMeshIdentity           func(meshName string)
		installSpiffeMTP            func(meshName string)
	)

	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(Yaml(
				builders.Mesh().
					WithName(redMeshName).
					WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Exclusive),
			)).
			Install(Yaml(
				builders.Mesh().
					WithName(blueMeshName).
					WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Exclusive),
			)).
			Setup(multizone.Global)).To(Succeed())

		Expect(WaitForMesh(redMeshName, []Cluster{multizone.KubeZone1, multizone.KubeZone2})).To(Succeed())
		Expect(WaitForMesh(blueMeshName, []Cluster{multizone.KubeZone1, multizone.KubeZone2})).To(Succeed())

		group := errgroup.Group{}
		NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Parallel(
				democlient.Install(
					democlient.WithNamespace(namespace),
					democlient.WithMesh(redMeshName),
					democlient.WithName(redClientZone1),
				),
				democlient.Install(
					democlient.WithNamespace(namespace),
					democlient.WithMesh(blueMeshName),
					democlient.WithName(blueClientZone1),
				),
			)).
			SetupInGroup(multizone.KubeZone1, &group)
		NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Parallel(
				testserver.Install(
					testserver.WithNamespace(namespace),
					testserver.WithMesh(redMeshName),
					testserver.WithName(redServerZone2),
					testserver.WithEchoArgs("echo", "--instance", redServerZone2),
				),
				testserver.Install(
					testserver.WithNamespace(namespace),
					testserver.WithMesh(blueMeshName),
					testserver.WithName(blueServerZone2),
					testserver.WithEchoArgs("echo", "--instance", blueServerZone2),
				),
			)).
			SetupInGroup(multizone.KubeZone2, &group)
		Expect(group.Wait()).To(Succeed())

		setupMeshIdentity(redMeshName)
		setupMeshIdentity(blueMeshName)

		installSpiffeMTP(redMeshName)
		installSpiffeMTP(blueMeshName)
	})

	AfterEachFailure(func() {
		DebugUniversal(multizone.Global, redMeshName)
		DebugUniversal(multizone.Global, blueMeshName)
		DebugKube(multizone.KubeZone1, redMeshName, namespace)
		DebugKube(multizone.KubeZone1, blueMeshName, namespace)
		DebugKube(multizone.KubeZone2, redMeshName, namespace)
		DebugKube(multizone.KubeZone2, blueMeshName, namespace)
	})

	E2EAfterAll(func() {
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.KubeZone2.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(redMeshName)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(blueMeshName)).To(Succeed())
	})

	It("should deploy zone proxies to meshproxymig-red with no interruptions for meshproxymig-blue", func(ctx SpecContext) {
		assertTrafficForBothMeshes()

		runDuringMigration(ctx, []trafficCheck{redTraffic, blueTraffic}, func() {
			deployMeshScopedZoneProxies(redMeshName)
			assertTrafficForBothMeshes()
			assertMeshScopedProxyUsed(redTraffic)
		})
	})

	It("should deploy zone proxies to meshproxymig-blue with no interruptions for meshproxymig-red", func(ctx SpecContext) {
		Skip("flaky zero-downtime check (transient empty response during migration); tracked in https://github.com/kumahq/kuma/issues/17161")
		assertTrafficForBothMeshes()

		runDuringMigration(ctx, []trafficCheck{redTraffic, blueTraffic}, func() {
			deployMeshScopedZoneProxies(blueMeshName)
			assertTrafficForBothMeshes()
			assertMeshScopedProxyUsed(blueTraffic)
		})
	})

	crossZoneAddress = func(check trafficCheck) string {
		GinkgoHelper()

		return fmt.Sprintf(
			"http://%s.%s.svc.%s.mesh.local:80",
			check.destinationServer,
			namespace,
			multizone.KubeZone2.ZoneName(),
		)
	}

	collectInstances = func(check trafficCheck, requests int) (map[string]int, error) {
		GinkgoHelper()

		return client.CollectResponsesByInstance(
			multizone.KubeZone1,
			check.clientName,
			crossZoneAddress(check),
			client.FromKubernetesPod(namespace, check.clientName),
			client.WithNumberOfRequests(uint(requests)),
		)
	}

	assertCrossZoneTraffic = func(g Gomega, check trafficCheck) {
		GinkgoHelper()

		const requests = 10

		instances, err := collectInstances(check, requests)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(instances).To(HaveLen(1))
		g.Expect(instances).To(HaveKey(check.expectedInstance))
		g.Expect(instances[check.expectedInstance]).To(Equal(requests))
	}

	assertTrafficForBothMeshes = func() {
		GinkgoHelper()

		Eventually(func(g Gomega) {
			assertCrossZoneTraffic(g, redTraffic)
		}, "60s", "1s").MustPassRepeatedly(3).Should(Succeed())

		Eventually(func(g Gomega) {
			assertCrossZoneTraffic(g, blueTraffic)
		}, "60s", "1s").MustPassRepeatedly(3).Should(Succeed())
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
	assertMeshScopedProxyUsed = func(check trafficCheck) {
		GinkgoHelper()

		ingressApp := fmt.Sprintf("mesh-zone-proxy-%s-ingress", check.meshName)
		ingressFilter := fmt.Sprintf(
			"cluster.kri_msvc_%s_%s_%s_%s_main.upstream_rq_total",
			check.meshName,
			multizone.KubeZone2.ZoneName(),
			namespace,
			check.destinationServer,
		)

		Eventually(func(g Gomega) {
			assertCrossZoneTraffic(g, check)
			stat, err := multizone.KubeZone2.GetEnvoyAdminTunnel(ingressApp, namespace).GetStats(ingressFilter)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeGreaterThanZero())
		}, "60s", "1s").Should(Succeed())
	}

	// runDuringMigration runs `do` while background goroutines continuously
	// send cross-zone requests and then verifies no request errors occurred.
	runDuringMigration = func(ctx context.Context, checks []trafficCheck, do func()) {
		GinkgoHelper()

		migrationCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		var (
			wg       sync.WaitGroup
			mu       sync.Mutex
			firstErr error
		)
		setFirstErr := func(err error) {
			if err == nil {
				return
			}
			mu.Lock()
			if firstErr == nil {
				firstErr = err
			}
			mu.Unlock()
		}

		for _, check := range checks {
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

					const requests = 5
					instances, err := collectInstances(check, requests)
					if err != nil {
						setFirstErr(err)
					} else if len(instances) != 1 || instances[check.expectedInstance] != requests {
						setFirstErr(fmt.Errorf(
							"unexpected instances for mesh %q traffic from %q to %q: got=%v expected=%q",
							check.meshName,
							check.clientName,
							check.destinationServer,
							instances,
							check.expectedInstance,
						))
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

	installSpiffeMTP = func(meshName string) {
		GinkgoHelper()

		meshTrafficPermission := fmt.Sprintf(`
type: MeshTrafficPermission
name: allow-all-spiffe-%[1]s
mesh: %[1]s
spec:
  targetRef:
    kind: Mesh
  rules:
    - default:
        allow:
          - spiffeID:
              type: Prefix
              value: spiffe://%[1]s.
`, meshName)

		Expect(NewClusterSetup().
			Install(YamlUniversal(meshTrafficPermission)).
			Setup(multizone.Global)).To(Succeed())
	}
}
