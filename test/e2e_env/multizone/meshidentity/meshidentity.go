package meshidentity

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/sync/errgroup"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	meshtrust_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshtrust/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/kds/hash"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
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

		Expect(group.Wait()).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(multizone.Global, meshName)
		DebugKube(multizone.KubeZone1, meshName, namespace)
		DebugKube(multizone.KubeZone2, meshName, namespace)
	})

	E2EAfterAll(func() {
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.KubeZone2.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(meshName)).To(Succeed())
	})
	// identity-c2v4v6874cx8x6c8-cww8457w48b482c7
	// identity-c2v4v6874cx8x6c8-w54dw4d47449z9z8

	getMeshTrust := func(zone string) *meshtrust_api.MeshTrust {
		trust, err := multizone.Global.GetKumactlOptions().RunKumactlAndGetOutput("get", "meshtrust", "-m", meshName, hash.HashedName(meshName, hash.HashedName(meshName, "identity"), zone, Config.KumaNamespace), "-ojson")
		println(fmt.Sprintf("ALL IDEN: %v", trust))
		Expect(err).ToNot(HaveOccurred())
		println(fmt.Sprintf("MATCHED IDEN: %v", trust))
		r, err := rest.JSON.Unmarshal([]byte(trust), meshtrust_api.MeshTrustResourceTypeDescriptor)
		Expect(err).ToNot(HaveOccurred())
		return r.GetSpec().(*meshtrust_api.MeshTrust)
	}

	indent := func(pem string, spaces int) string {
		pad := strings.Repeat(" ", spaces)
		return pad + strings.ReplaceAll(pem, "\n", "\n"+pad)
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
    path: "/ns/{{ .Namespace }}/sa/{{ .ServiceAccount }}"
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

		// when
		// added Trust from zone 1 to zone 2
		trustTmpl := `
apiVersion: kuma.io/v1alpha1
kind: MeshTrust
metadata:
  name: identity-trust-%s
  namespace: kuma-system
  labels:
    kuma.io/mesh: %s
    kuma.io/origin: zone
    kuma.io/zone: %s
spec:
  caBundles:
    - type: Pem
      pem:
        value: |-
%s
  trustDomain: %s
`
		trustZone1 := getMeshTrust(multizone.KubeZone1.Name())
		Expect(NewClusterSetup().
			Install(YamlK8s(fmt.Sprintf(trustTmpl, multizone.KubeZone1.Name(), meshName, multizone.KubeZone2.Name(), indent(trustZone1.CABundles[0].PEM.Value, 10), trustZone1.TrustDomain))).
			Setup(multizone.KubeZone2)).To(Succeed())

		trustZone2 := getMeshTrust(multizone.KubeZone2.Name())
		Expect(NewClusterSetup().
			Install(YamlK8s(fmt.Sprintf(trustTmpl, multizone.KubeZone2.Name(), meshName, multizone.KubeZone1.Name(), indent(trustZone2.CABundles[0].PEM.Value, 10), trustZone2.TrustDomain))).
			Setup(multizone.KubeZone1)).To(Succeed())

		// and Trust from zone 2 to zone 1
		// cross zone traffic works
		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				multizone.KubeZone1, "demo-client", "test-server.meshidentity.svc.kuma-2.mesh.local",
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(Equal("kube-test-server-zone-2"))
		}, "30s", "1s").MustPassRepeatedly(5).Should(Succeed())

		// cross zone traffic works
		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				multizone.KubeZone2, "demo-client", "test-server.meshidentity.svc.kuma-1.mesh.local",
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(Equal("kube-test-server-zone-1"))
		}, "30s", "1s").MustPassRepeatedly(5).Should(Succeed())
	})
}
