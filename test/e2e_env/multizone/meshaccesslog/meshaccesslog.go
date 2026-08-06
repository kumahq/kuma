package meshaccesslog

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/sync/errgroup"

	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	. "github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/client"
	"github.com/kumahq/kuma/v3/test/framework/deployments/zoneproxy"
	"github.com/kumahq/kuma/v3/test/framework/envs/multizone"
)

func ZoneIngress() {
	const meshName = "mal-zone-ingress"
	const identityName = "mal-zone-ingress-identity"
	const demoClient = "demo-client"
	const testServer1 = "test-server-1"
	const testServer2 = "test-server-2"

	ingressWorkload := zoneproxy.IngressName(meshName)

	var zone1Name string
	var tcpSinkDockerName, testServer1SNI, testServer2SNI string

	BeforeAll(func() {
		zone1Name = multizone.UniZone1.ZoneName()
		tcpSinkDockerName = fmt.Sprintf("%s_%s_%s", multizone.UniZone1.Name(), meshName, AppModeTcpSink)
		testServer1SNI = fmt.Sprintf("sni.msvc.%s.%s.%s.80", meshName, zone1Name, testServer1)
		testServer2SNI = fmt.Sprintf("sni.msvc.%s.%s.%s.80", meshName, zone1Name, testServer2)

		zones := []Cluster{multizone.UniZone1, multizone.UniZone2}
		Expect(NewClusterSetup().
			Install(Yaml(builders.Mesh().
				WithName(meshName))).
			Install(MeshIdentityBundled(meshName, identityName)).
			Install(MeshTrafficPermissionAllowAllUniversalWorkloadIdentity(meshName, MeshIdentityTrustDomains(meshName, zones...)...)).
			Setup(multizone.Global)).To(Succeed())
		Expect(WaitForMesh(meshName, multizone.Zones())).To(Succeed())

		group := errgroup.Group{}
		NewClusterSetup().
			Install(Parallel(
				TestServerUniversal(testServer1, meshName,
					WithArgs([]string{"echo", "--instance", testServer1}),
					WithServiceName(testServer1),
					WithWorkload(testServer1),
				),
				TestServerUniversal(testServer2, meshName,
					WithArgs([]string{"echo", "--instance", testServer2}),
					WithServiceName(testServer2),
					WithWorkload(testServer2),
				),
				TcpSinkUniversal(AppModeTcpSink, WithDockerContainerName(tcpSinkDockerName)),
				zoneproxy.Install(
					zoneproxy.WithMesh(meshName),
					zoneproxy.WithIngress(),
				),
			)).
			SetupInGroup(multizone.UniZone1, &group)

		NewClusterSetup().
			Install(DemoClientUniversal(demoClient, meshName,
				WithTransparentProxy(true),
				WithWorkload(demoClient),
			)).
			SetupInGroup(multizone.UniZone2, &group)
		Expect(group.Wait()).To(Succeed())

		Expect(DistributeMeshTrusts(multizone.Global, meshName, identityName, zones...)).To(Succeed())
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
    labels:
      kuma.io/workload: %s
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

		By("traffic to test-server-2 becomes routable before checking that it is not logged")
		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(multizone.UniZone2, demoClient, urlFor(testServer2))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(Equal(testServer2))
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
