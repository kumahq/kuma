package meshidentity

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/sync/errgroup"

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
	"github.com/kumahq/kuma/v3/test/framework/envs/multizone"
)

func Identity() {
	namespace := "meshidentity"
	meshName := "meshidentity"

	zoneIngress := func() InstallFunc {
		return zoneproxy.Install(
			zoneproxy.WithMesh(meshName),
			zoneproxy.WithNamespace(namespace),
			zoneproxy.WithIngress(),
		)
	}

	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(Yaml(
				builders.Mesh().
					WithName(meshName),
			)).
			Install(MeshTrafficPermissionAllowAllUniversalWorkloadIdentity(meshName,
				fmt.Sprintf("%s.%s.mesh.local", meshName, multizone.KubeZone1.ZoneName()),
				fmt.Sprintf("%s.%s.mesh.local", meshName, multizone.KubeZone2.ZoneName()),
				fmt.Sprintf("%s.%s.mesh.local", meshName, multizone.UniZone1.ZoneName()),
			)).
			Install(YamlUniversal(fmt.Sprintf(`
type: MeshMultiZoneService
name: test-server-mi
mesh: %s
spec:
  selector:
    meshService:
      matchLabels:
        kuma.io/display-name: test-server
  ports:
  - port: 80
    appProtocol: http
`, meshName))).
			Install(YamlUniversal(fmt.Sprintf(`
type: MeshLoadBalancingStrategy
name: disable-la-to-test-server
mesh: %s
spec:
  to:
  - targetRef:
      kind: MeshMultiZoneService
      labels:
        kuma.io/display-name: test-server-mi
      sectionName: '80'
    default:
      localityAwareness:
        disabled: true`, meshName))).
			Setup(multizone.Global)).To(Succeed())
		Expect(WaitForMesh(meshName, multizone.Zones())).To(Succeed())

		group := errgroup.Group{}
		NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Parallel(
				testserver.Install(
					testserver.WithNamespace(namespace),
					testserver.WithMesh(meshName),
					testserver.WithEchoArgs("echo", "--instance", "kube-test-server-zone-1"),
				),
				democlient.Install(democlient.WithNamespace(namespace), democlient.WithMesh(meshName)),
				zoneIngress(),
			)).
			SetupInGroup(multizone.KubeZone1, &group)

		NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Parallel(
				testserver.Install(
					testserver.WithNamespace(namespace),
					testserver.WithMesh(meshName),
					testserver.WithEchoArgs("echo", "--instance", "kube-test-server-zone-2"),
				),
				democlient.Install(democlient.WithNamespace(namespace), democlient.WithMesh(meshName)),
				zoneIngress(),
			)).
			SetupInGroup(multizone.KubeZone2, &group)

		NewClusterSetup().
			Install(Parallel(
				DemoClientUniversal("demo-client", meshName, WithTransparentProxy(true), WithWorkload("demo-client")),
				TestServerUniversal("test-server", meshName, WithArgs([]string{"echo", "--instance", "uni-test-server-zone-4"}), WithWorkload("test-server")),
				zoneIngress(),
			)).
			SetupInGroup(multizone.UniZone1, &group)

		Expect(group.Wait()).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(multizone.Global, meshName)
		DebugKube(multizone.KubeZone1, meshName, namespace)
		DebugKube(multizone.KubeZone2, meshName, namespace)
		DebugUniversal(multizone.UniZone1, meshName)
	})

	E2EAfterAll(func() {
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.KubeZone2.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.UniZone1.DeleteMeshApps(meshName)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(meshName)).To(Succeed())
	})

	getMeshTrust := func(hashValues ...string) (*meshtrust_api.MeshTrust, error) {
		trust, err := multizone.Global.GetKumactlOptions().RunKumactlAndGetOutput("get", "meshtrust", "-m", meshName, hash.HashedName(meshName, hash.HashedName(meshName, "identity"), hashValues...), "-ojson")
		if err != nil {
			return nil, err
		}
		r, err := rest.JSON.Unmarshal([]byte(trust), meshtrust_api.MeshTrustResourceTypeDescriptor)
		if err != nil {
			return nil, err
		}
		return r.GetSpec().(*meshtrust_api.MeshTrust), nil
	}

	waitForMeshTrust := func(hashValues ...string) *meshtrust_api.MeshTrust {
		var trust *meshtrust_api.MeshTrust
		Eventually(func(g Gomega) {
			var err error
			trust, err = getMeshTrust(hashValues...)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(trust).ToNot(BeNil())
		}, "60s", "1s").Should(Succeed())
		return trust
	}

	buildMeshTrustYaml := func(trust *meshtrust_api.MeshTrust, sourceZoneName, targetZoneName string, isK8s bool) string {
		builder := builders.MeshTrust().
			WithName("identity-trust-" + sourceZoneName).
			WithMesh(meshName).
			WithLabels(map[string]string{
				"kuma.io/origin": "zone",
				"kuma.io/zone":   targetZoneName,
			}).
			WithCA(trust.CABundles[0].PEM.Value).
			WithTrustDomain(trust.TrustDomain)
		if isK8s {
			return builder.WithNamespace(Config.KumaNamespace).KubeYaml()
		}
		return builder.UniYaml()
	}

	installTrustToZone := func(trust *meshtrust_api.MeshTrust, sourceZoneName string, targetZone Cluster, isK8s bool) error {
		yaml := buildMeshTrustYaml(trust, sourceZoneName, targetZone.ZoneName(), isK8s)
		if isK8s {
			return NewClusterSetup().Install(YamlK8s(yaml)).Setup(targetZone)
		} else {
			return NewClusterSetup().Install(YamlUniversal(yaml)).Setup(targetZone)
		}
	}

	expectTraffic := func(from Cluster, destination string, instance string, opts ...client.CollectResponsesOptsFn) {
		GinkgoHelper()
		reachable := func(g Gomega) {
			resp, err := client.CollectEchoResponse(from, "demo-client", destination, opts...)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(Equal(instance))
		}
		// MustPassRepeatedly rather than Consistently: applying a MeshIdentity rotates
		// certificates and rebuilds listeners, so a request can fail while the route is
		// being replaced. Requiring five consecutive successes proves the route settled
		// without failing on the churn that gets there.
		Eventually(reachable, "2m", "1s").MustPassRepeatedly(5).Should(Succeed())
	}

	It("should access the service in the same zone using mTLS", func() {
		// given
		// traffic in local zones works
		expectTraffic(multizone.KubeZone1, "test-server", "kube-test-server-zone-1", client.FromKubernetesPod(namespace, "demo-client"))

		// and
		expectTraffic(multizone.KubeZone2, "test-server", "kube-test-server-zone-2", client.FromKubernetesPod(namespace, "demo-client"))

		// and
		expectTraffic(multizone.UniZone1, "test-server.svc.mesh.local", "uni-test-server-zone-4")

		// when
		yaml := fmt.Sprintf(`
type: MeshIdentity
name: identity
mesh: %s
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
`, meshName)
		Expect(NewClusterSetup().
			Install(YamlUniversal(yaml)).
			Setup(multizone.Global)).To(Succeed())
		hashedName := hash.HashedName(meshName, "identity")
		Expect(WaitForResource(meshidentity_api.MeshIdentityResourceTypeDescriptor, model.ResourceKey{Mesh: meshName, Name: fmt.Sprintf("%s.%s", hashedName, Config.KumaNamespace)}, multizone.KubeZone1, multizone.KubeZone2)).To(Succeed())

		// then
		// mTLS traffic in local zone works
		expectTraffic(multizone.KubeZone1, "test-server", "kube-test-server-zone-1", client.FromKubernetesPod(namespace, "demo-client"))

		// mTLS traffic in local zone works
		expectTraffic(multizone.KubeZone2, "test-server", "kube-test-server-zone-2", client.FromKubernetesPod(namespace, "demo-client"))

		// and
		expectTraffic(multizone.UniZone1, "test-server.svc.mesh.local", "uni-test-server-zone-4")

		// when
		// added Trust from zone 1 to zone 2
		trustZone1 := waitForMeshTrust(multizone.KubeZone1.Name(), Config.KumaNamespace)
		Expect(installTrustToZone(trustZone1, multizone.KubeZone1.Name(), multizone.KubeZone2, true)).To(Succeed())
		Expect(installTrustToZone(trustZone1, multizone.KubeZone1.Name(), multizone.UniZone1, false)).To(Succeed())

		trustZone2 := waitForMeshTrust(multizone.KubeZone2.Name(), Config.KumaNamespace)
		Expect(installTrustToZone(trustZone2, multizone.KubeZone2.Name(), multizone.KubeZone1, true)).To(Succeed())
		Expect(installTrustToZone(trustZone2, multizone.KubeZone2.Name(), multizone.UniZone1, false)).To(Succeed())

		trustZone4 := waitForMeshTrust(multizone.UniZone1.Name())
		Expect(installTrustToZone(trustZone4, multizone.UniZone1.Name(), multizone.KubeZone1, true)).To(Succeed())
		Expect(installTrustToZone(trustZone4, multizone.UniZone1.Name(), multizone.KubeZone2, true)).To(Succeed())

		// cross zone traffic works: kube-1 -> kube-2
		expectTraffic(multizone.KubeZone1, "test-server.meshidentity.svc.kuma-2.mesh.local", "kube-test-server-zone-2", client.FromKubernetesPod(namespace, "demo-client"))

		// cross zone traffic works: kube-2 -> kube-1
		expectTraffic(multizone.KubeZone2, "test-server.meshidentity.svc.kuma-1.mesh.local", "kube-test-server-zone-1", client.FromKubernetesPod(namespace, "demo-client"))

		// cross zone traffic works: kube-2 -> uni-1
		expectTraffic(multizone.KubeZone2, "test-server.svc.kuma-4.mesh.local", "uni-test-server-zone-4", client.FromKubernetesPod(namespace, "demo-client"))

		// cross zone traffic works: uni-1 -> kube-1
		expectTraffic(multizone.UniZone1, "test-server.meshidentity.svc.kuma-1.mesh.local", "kube-test-server-zone-1")

		// meshmultizone works
		Expect(client.CollectResponsesByInstance(multizone.UniZone1, "demo-client", "test-server-mi.mzsvc.mesh.local", client.WithNumberOfRequests(50))).
			Should(And(
				HaveLen(3),
				HaveKey(Equal(`kube-test-server-zone-1`)),
				HaveKey(Equal(`kube-test-server-zone-2`)),
				HaveKey(Equal(`uni-test-server-zone-4`)),
			))
	})
}
