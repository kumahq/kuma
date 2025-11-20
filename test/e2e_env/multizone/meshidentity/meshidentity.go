package meshidentity

import (
	"fmt"

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
	"github.com/kumahq/kuma/v2/test/framework/envs/multizone"
)

func Identity() {
	namespace := "meshidentity"
	meshName := "meshidentity"

	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(Yaml(
				builders.Mesh().
					WithName(meshName).
					WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Exclusive),
			)).
			Install(MeshTrafficPermissionAllowAllUniversal(meshName)).
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
			)).
			SetupInGroup(multizone.KubeZone2, &group)

		NewClusterSetup().
			Install(Parallel(
				DemoClientUniversal("demo-client", meshName, WithTransparentProxy(true), WithWorkload("demo-client")),
				TestServerUniversal("test-server", meshName, WithArgs([]string{"echo", "--instance", "uni-test-server-zone-4"}), WithWorkload("test-server")),
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
	// identity-c2v4v6874cx8x6c8-cww8457w48b482c7
	// identity-c2v4v6874cx8x6c8-w54dw4d47449z9z8

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
		}, "30s", "1s").Should(Succeed())
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

	It("should access the service in the same zone using mTLS", func() {
		// given
		// traffic in local zones works
		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				multizone.KubeZone1, "demo-client", "test-server",
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(Equal("kube-test-server-zone-1"))
		}, "30s", "1s").MustPassRepeatedly(5).Should(Succeed())

		// and
		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				multizone.KubeZone2, "demo-client", "test-server",
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(Equal("kube-test-server-zone-2"))
		}, "30s", "1s").MustPassRepeatedly(5).Should(Succeed())

		// and
		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				multizone.UniZone1, "demo-client", "test-server.svc.mesh.local",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(Equal("uni-test-server-zone-4"))
		}, "30s", "1s").MustPassRepeatedly(5).Should(Succeed())

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
		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				multizone.KubeZone1, "demo-client", "test-server",
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(Equal("kube-test-server-zone-1"))
		}, "30s", "1s").MustPassRepeatedly(5).Should(Succeed())

		// mTLS traffic in local zone works
		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				multizone.KubeZone2, "demo-client", "test-server",
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(Equal("kube-test-server-zone-2"))
		}, "30s", "1s").MustPassRepeatedly(5).Should(Succeed())

		// and
		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				multizone.UniZone1, "demo-client", "test-server.svc.mesh.local",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(Equal("uni-test-server-zone-4"))
		}, "30s", "1s").MustPassRepeatedly(5).Should(Succeed())

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

		// time.Sleep(1 * time.Hour)
		// cross zone traffic works: kube-1 -> kube-2
		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				multizone.KubeZone1, "demo-client", "test-server.meshidentity.svc.kuma-2.mesh.local",
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(Equal("kube-test-server-zone-2"))
		}, "30s", "1s").MustPassRepeatedly(5).Should(Succeed())

		// cross zone traffic works: kube-2 -> kube-1
		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				multizone.KubeZone2, "demo-client", "test-server.meshidentity.svc.kuma-1.mesh.local",
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(Equal("kube-test-server-zone-1"))
		}, "30s", "1s").MustPassRepeatedly(5).Should(Succeed())

		// cross zone traffic works: kube-2 -> uni-1
		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				multizone.KubeZone2, "demo-client", "test-server.svc.kuma-4.mesh.local",
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(Equal("uni-test-server-zone-4"))
		}, "30s", "1s").MustPassRepeatedly(5).Should(Succeed())

		// cross zone traffic works: uni-1 -> kube-1
		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				multizone.UniZone1, "demo-client", "test-server.meshidentity.svc.kuma-1.mesh.local",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(Equal("kube-test-server-zone-1"))
		}, "30s", "1s").MustPassRepeatedly(5).Should(Succeed())
	})
}
