package meshaccesslog

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
	. "github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/client"
	"github.com/kumahq/kuma/v2/test/framework/deployments/zoneproxy"
	"github.com/kumahq/kuma/v2/test/framework/envs/universal"
)

func Matches() {
	meshName := "mal-matches"
	externalServiceName := "mal-matches-ext"
	egressDP := "zone-proxy-egress"
	demoClient1 := "demo-client-1"
	demoClient2 := "demo-client-2"
	testServer := "test-server"

	dppEnvs := map[string]string{
		"KUMA_DATAPLANE_RUNTIME_UNIFIED_RESOURCE_NAMING_ENABLED": "true",
	}

	var zoneName, tcpSinkDockerName, externalServiceDockerName, demoClient1SpiffeID, demoClient2SpiffeID string

	BeforeAll(func() {
		zoneName = universal.Cluster.ZoneName()
		tcpSinkDockerName = fmt.Sprintf("%s_%s_%s", universal.Cluster.Name(), meshName, AppModeTcpSink)
		externalServiceDockerName = fmt.Sprintf("%s_%s_%s", universal.Cluster.Name(), meshName, externalServiceName)

		Expect(NewClusterSetup().
			Install(Yaml(builders.Mesh().
				WithName(meshName).
				WithoutInitialPolicies().
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
          value: spiffe://%s.%s.mesh.local
`, meshName, meshName, zoneName))).
			Install(YamlUniversal(fmt.Sprintf(`
type: MeshExternalService
name: httpbin
mesh: %s
labels:
  kuma.io/origin: zone
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: http
  endpoints:
    - address: %s
      port: 80
`, meshName, externalServiceDockerName))).
			Install(zoneproxy.Install(
				zoneproxy.WithMesh(meshName),
				zoneproxy.WithEgressPort(11102),
				zoneproxy.WithWorkload(egressDP),
				zoneproxy.WithDpEnvs(dppEnvs),
			)).
			Install(TcpSinkUniversal(AppModeTcpSink, WithDockerContainerName(tcpSinkDockerName))).
			Install(TestServerExternalServiceUniversal(externalServiceName, 80, false,
				WithDockerContainerName(externalServiceDockerName))).
			Install(DemoClientUniversal(demoClient1, meshName,
				WithTransparentProxy(true),
				WithWorkload(demoClient1),
				WithDpEnvs(dppEnvs),
			)).
			Install(DemoClientUniversal(demoClient2, meshName,
				WithTransparentProxy(true),
				WithWorkload(demoClient2),
				WithDpEnvs(dppEnvs),
			)).
			Install(TestServerUniversal(testServer, meshName,
				WithArgs([]string{"echo", "--instance", testServer}),
				WithWorkload(testServer),
				WithDpEnvs(dppEnvs),
			)).
			Setup(universal.Cluster)).To(Succeed())

		Eventually(func(g Gomega) {
			out, err := universal.Cluster.GetKumactlOptions().
				RunKumactlAndGetOutput("get", "meshidentity", "-m", meshName, "identity", "-o", "json")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(out).To(ContainSubstring("Successfully initialized"))
		}, "30s", "1s").Should(Succeed())

		demoClient1SpiffeID = fmt.Sprintf("spiffe://%s.%s.mesh.local/workload/%s", meshName, zoneName, demoClient1)
		demoClient2SpiffeID = fmt.Sprintf("spiffe://%s.%s.mesh.local/workload/%s", meshName, zoneName, demoClient2)

		// Rules fire independently under parallel semantics, so each rule's
		// match is made disjoint by SpiffeID to keep `tail -1` deterministic.
		egressMAL := fmt.Sprintf(`
type: MeshAccessLog
name: mal-matches-egress
mesh: %s
spec:
  targetRef:
    kind: Dataplane
    name: %s
  rules:
    - matches:
        - spiffeID:
            type: Exact
            value: %q
      default:
        backends:
          - type: Tcp
            tcp:
              format:
                type: Plain
                plain: "on-egress spiffe=demo-client-1"
              address: "%s:9999"
    - matches:
        - spiffeID:
            type: Exact
            value: %q
          sni:
            type: Exact
            value: sni.extsvc.%s.%s.httpbin.80
      default:
        backends:
          - type: Tcp
            tcp:
              format:
                type: Plain
                plain: "on-egress sni=httpbin"
              address: "%s:9999"
`, meshName, egressDP,
			demoClient1SpiffeID, tcpSinkDockerName,
			demoClient2SpiffeID, meshName, zoneName, tcpSinkDockerName,
		)
		Expect(YamlUniversal(egressMAL)(universal.Cluster)).To(Succeed())

		sidecarMAL := fmt.Sprintf(`
type: MeshAccessLog
name: mal-matches-sidecar
mesh: %s
spec:
  targetRef:
    kind: Dataplane
    name: %s
  rules:
    - matches:
        - spiffeID:
            type: Exact
            value: %q
      default:
        backends:
          - type: Tcp
            tcp:
              format:
                type: Plain
                plain: "on-sidecar spiffe=demo-client-1"
              address: "%s:9999"
    - matches:
        - spiffeID:
            type: Exact
            value: %q
          sni:
            type: Exact
            value: sni.msvc.%s.%s.%s.80
      default:
        backends:
          - type: Tcp
            tcp:
              format:
                type: Plain
                plain: "on-sidecar sni=test-server"
              address: "%s:9999"
`, meshName, testServer,
			demoClient1SpiffeID, tcpSinkDockerName,
			demoClient2SpiffeID, meshName, zoneName, testServer, tcpSinkDockerName,
		)
		Expect(YamlUniversal(sidecarMAL)(universal.Cluster)).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, meshName)
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteApp(externalServiceName)).To(Succeed())
		Expect(universal.Cluster.DeleteApp(AppModeTcpSink)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should match rules on the zone egress listener for MeshExternalService", func() {
		By("request from demo-client-1 triggers the spiffeID rule")
		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				universal.Cluster, demoClient1, "httpbin.extsvc.mesh.local",
			)
			g.Expect(err).ToNot(HaveOccurred())

			stdout, _, err := universal.Cluster.Exec("", "", AppModeTcpSink, "tail", "-1", "/nc.out")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(strings.TrimSpace(stdout)).To(Equal("on-egress spiffe=demo-client-1"))
		}, "60s", "1s").Should(Succeed())

		By("request from demo-client-2 matches the spiffeID+SNI rule")
		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				universal.Cluster, demoClient2, "httpbin.extsvc.mesh.local",
			)
			g.Expect(err).ToNot(HaveOccurred())

			stdout, _, err := universal.Cluster.Exec("", "", AppModeTcpSink, "tail", "-1", "/nc.out")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(strings.TrimSpace(stdout)).To(Equal("on-egress sni=httpbin"))
		}, "60s", "1s").Should(Succeed())
	})

	It("should match rules on the test-server sidecar inbound listener", func() {
		testServerURL := fmt.Sprintf("%s.svc.mesh.local", testServer)

		By("request from demo-client-1 triggers the spiffeID rule")
		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				universal.Cluster, demoClient1, testServerURL,
			)
			g.Expect(err).ToNot(HaveOccurred())

			stdout, _, err := universal.Cluster.Exec("", "", AppModeTcpSink, "tail", "-1", "/nc.out")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(strings.TrimSpace(stdout)).To(Equal("on-sidecar spiffe=demo-client-1"))
		}, "60s", "1s").Should(Succeed())

		By("request from demo-client-2 matches the spiffeID+SNI rule")
		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				universal.Cluster, demoClient2, testServerURL,
			)
			g.Expect(err).ToNot(HaveOccurred())

			stdout, _, err := universal.Cluster.Exec("", "", AppModeTcpSink, "tail", "-1", "/nc.out")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(strings.TrimSpace(stdout)).To(Equal("on-sidecar sni=test-server"))
		}, "60s", "1s").Should(Succeed())
	})
}
