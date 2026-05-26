package meshaccesslog

import (
	"fmt"
	"strings"
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
	"github.com/kumahq/kuma/v2/test/framework/deployments/zoneproxy"
	"github.com/kumahq/kuma/v2/test/framework/envs/multizone"
)

func ZoneIngress() {
	const meshName = "mal-zone-ingress"
	const demoClient = "demo-client"
	const testServer1 = "test-server-1"
	const testServer2 = "test-server-2"
	const ingressWorkload = "zone-proxy-ingress"
	const ingressPort = uint32(11001)

	dppEnvs := map[string]string{
		"KUMA_DATAPLANE_RUNTIME_UNIFIED_RESOURCE_NAMING_ENABLED": "true",
	}

	var zone1Name string
	var tcpSinkDockerName, testServer1SNI, testServer2SNI string

	BeforeAll(func() {
		zone1Name = multizone.UniZone1.ZoneName()
		tcpSinkDockerName = fmt.Sprintf("%s_%s_%s", multizone.UniZone1.Name(), meshName, AppModeTcpSink)
		testServer1SNI = fmt.Sprintf("sni.msvc.%s.%s.%s.80", meshName, zone1Name, testServer1)
		testServer2SNI = fmt.Sprintf("sni.msvc.%s.%s.%s.80", meshName, zone1Name, testServer2)

		Expect(NewClusterSetup().
			Install(Yaml(builders.Mesh().
				WithName(meshName).
				WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Exclusive))).
			Install(YamlUniversal(fmt.Sprintf(`
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
`, meshName))).
			Install(YamlUniversal(fmt.Sprintf(`
type: MeshTrafficPermission
name: allow-mesh
mesh: %s
spec:
  rules:
  - default:
      allow:
      - spiffeID:
          type: Prefix
          value: "spiffe://%s."
`, meshName, meshName))).
			Setup(multizone.Global)).To(Succeed())
		Expect(WaitForMesh(meshName, multizone.Zones())).To(Succeed())

		group := errgroup.Group{}
		NewClusterSetup().
			Install(Parallel(
				TestServerUniversal(testServer1, meshName,
					WithArgs([]string{"echo", "--instance", testServer1}),
					WithServiceName(testServer1),
					WithWorkload(testServer1),
					WithDpEnvs(dppEnvs),
				),
				TestServerUniversal(testServer2, meshName,
					WithArgs([]string{"echo", "--instance", testServer2}),
					WithServiceName(testServer2),
					WithWorkload(testServer2),
					WithDpEnvs(dppEnvs),
				),
				TcpSinkUniversal(AppModeTcpSink, WithDockerContainerName(tcpSinkDockerName)),
				zoneproxy.Install(
					zoneproxy.WithMesh(meshName),
					zoneproxy.WithIngressPort(ingressPort),
					zoneproxy.WithWorkload(ingressWorkload),
					zoneproxy.WithDpEnvs(dppEnvs),
				),
			)).
			SetupInGroup(multizone.UniZone1, &group)

		NewClusterSetup().
			Install(DemoClientUniversal(demoClient, meshName,
				WithTransparentProxy(true),
				WithWorkload(demoClient),
				WithDpEnvs(dppEnvs),
			)).
			SetupInGroup(multizone.UniZone2, &group)
		Expect(group.Wait()).To(Succeed())

		// MeshZoneAddress is auto-created on k8s; on Universal it must be installed by hand.
		ingressIP := multizone.UniZone1.GetApp(ingressWorkload).GetIP()
		Expect(NewClusterSetup().
			Install(YamlUniversal(fmt.Sprintf(`
type: MeshZoneAddress
name: %s
mesh: %s
labels:
  kuma.io/origin: zone
  kuma.io/zone: %s
spec:
  address: %s
  port: %d
`, ingressWorkload, meshName, zone1Name, ingressIP, ingressPort))).
			Setup(multizone.UniZone1)).To(Succeed())

		// Wait for MeshIdentity to sync, then publish per-zone MeshTrust to Global so
		// KDS distributes it to all zones and cross-zone mTLS is established.
		hashedIdentityName := hash.HashedName(meshName, "identity")
		Expect(WaitForResource(
			meshidentity_api.MeshIdentityResourceTypeDescriptor,
			model.ResourceKey{Mesh: meshName, Name: hashedIdentityName},
			multizone.UniZone1, multizone.UniZone2,
		)).To(Succeed())

		getMeshTrust := func(zoneName string) *meshtrust_api.MeshTrust {
			var trust *meshtrust_api.MeshTrust
			Eventually(func(g Gomega) {
				out, err := multizone.Global.GetKumactlOptions().RunKumactlAndGetOutput(
					"get", "meshtrust", "-m", meshName,
					hash.HashedName(meshName, hashedIdentityName, zoneName),
					"-ojson",
				)
				g.Expect(err).ToNot(HaveOccurred())
				r, err := rest.JSON.Unmarshal([]byte(out), meshtrust_api.MeshTrustResourceTypeDescriptor)
				g.Expect(err).ToNot(HaveOccurred())
				trust = r.GetSpec().(*meshtrust_api.MeshTrust)
			}, "60s", "1s").Should(Succeed())
			return trust
		}

		installTrustToGlobal := func(trust *meshtrust_api.MeshTrust, sourceZone string) {
			yaml := builders.MeshTrust().
				WithName("meshtrust-of-zone-" + sourceZone).
				WithMesh(meshName).
				WithCA(trust.CABundles[0].PEM.Value).
				WithTrustDomain(trust.TrustDomain).
				UniYaml()
			Expect(NewClusterSetup().Install(YamlUniversal(yaml)).Setup(multizone.Global)).To(Succeed())
		}

		installTrustToGlobal(getMeshTrust(multizone.UniZone1.Name()), multizone.UniZone1.Name())
		installTrustToGlobal(getMeshTrust(multizone.UniZone2.Name()), multizone.UniZone2.Name())
	})

	AfterEachFailure(func() {
		DebugUniversal(multizone.Global, meshName)
		DebugUniversal(multizone.UniZone1, meshName)
		DebugUniversal(multizone.UniZone2, meshName)
	})

	E2EAfterAll(func() {
		Expect(multizone.UniZone1.DeleteMeshApps(meshName)).To(Succeed())
		Expect(multizone.UniZone2.DeleteMeshApps(meshName)).To(Succeed())
		Expect(multizone.UniZone1.DeleteApp(AppModeTcpSink)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(meshName)).To(Succeed())
	})

	It("should log only traffic whose SNI matches the rule", func() {
		// Zone ingress does not terminate mTLS (see MADR-103 §"Zone ingress note"),
		// so DOWNSTREAM_PEER_URI_SAN is not available — log only the SNI.
		mal := fmt.Sprintf(`
type: MeshAccessLog
name: mal-on-zone-ingress
mesh: %s
labels:
  kuma.io/origin: zone
spec:
  targetRef:
    kind: Dataplane
    name: %s
  rules:
    - matches:
        - sni:
            type: Exact
            value: %s
      default:
        backends:
          - type: Tcp
            tcp:
              format:
                type: Plain
                plain: "sni=%%REQUESTED_SERVER_NAME%%"
              address: "%s:9999"
`, meshName, ingressWorkload, testServer1SNI, tcpSinkDockerName)
		Expect(YamlUniversal(mal)(multizone.UniZone1)).To(Succeed())

		urlFor := func(name string) string {
			return fmt.Sprintf("http://%s.svc.%s.mesh.local", name, zone1Name)
		}
		readLog := func() (string, error) {
			stdout, _, err := multizone.UniZone1.Exec("", "", AppModeTcpSink, "tail", "-1", "/nc.out")
			return strings.TrimSpace(stdout), err
		}

		By("traffic to test-server-1 produces a log entry with the matching SNI")
		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(multizone.UniZone2, demoClient, urlFor(testServer1))
			g.Expect(err).ToNot(HaveOccurred())

			log, err := readLog()
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(log).To(Equal(fmt.Sprintf("sni=%s", testServer1SNI)))
		}, "60s", "1s").Should(Succeed())

		By("traffic to test-server-2 does not match the rule and is not logged")
		Consistently(func(g Gomega) {
			_, err := client.CollectEchoResponse(multizone.UniZone2, demoClient, urlFor(testServer2))
			g.Expect(err).ToNot(HaveOccurred())

			log, err := readLog()
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(log).ToNot(ContainSubstring(testServer2SNI))
		}, "10s", "1s").Should(Succeed())
	})
}
