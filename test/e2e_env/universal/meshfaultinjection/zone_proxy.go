package meshfaultinjection

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	meshfaultinjection_api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshfaultinjection/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
	. "github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/client"
	"github.com/kumahq/kuma/v2/test/framework/deployments/zoneproxy"
	"github.com/kumahq/kuma/v2/test/framework/envs/universal"
)

func zoneProxyMeshIdentity(meshName string) string {
	return fmt.Sprintf(`
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
}

func zoneProxyMeshTrafficPermission(meshName, zoneName string) string {
	return fmt.Sprintf(`
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
`, meshName, meshName, zoneName)
}

func zoneProxyMeshExternalService(meshName, name, address string) string {
	return fmt.Sprintf(`
type: MeshExternalService
name: %s
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
`, name, meshName, address)
}

func zoneProxyPolicy(meshName, targetSNI string) string {
	return fmt.Sprintf(`
type: MeshFaultInjection
name: mfi-zone-egress-sni
mesh: %s
spec:
  targetRef:
    kind: Dataplane
    labels:
      kuma.io/listener-zoneegress: enabled
  rules:
    - matches:
        - sni:
            type: Exact
            value: %q
      default:
        http:
          - abort:
              httpStatus: 503
              percentage: 100
`, meshName, targetSNI)
}

func zoneProxySNI(meshName, zoneName, resourceName string) string {
	return fmt.Sprintf("sni.extsvc.%s.%s.%s.80", meshName, zoneName, resourceName)
}

// ZoneProxy verifies that MeshFaultInjection with an SNI match is enforced on
// the live mesh-scoped zone-egress listener and does not affect unmatched
// MeshExternalService hostnames sharing the same external endpoint.
func ZoneProxy() {
	const meshName = "mfi-zone-proxy"
	const demoClient = "demo-client"
	const egressWorkload = "zone-proxy-egress"
	const externalServiceApp = "mfi-zone-proxy-ext"
	const targetedMES = "mfi-target"
	const untargetedMES = "mfi-other"

	dppEnvs := map[string]string{
		"KUMA_DATAPLANE_RUNTIME_UNIFIED_RESOURCE_NAMING_ENABLED": "true",
	}

	var zoneName, externalServiceDockerName string

	trafficAllowed := func(address string) {
		GinkgoHelper()

		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(universal.Cluster, demoClient, address)
			g.Expect(err).ToNot(HaveOccurred())
		}, "60s", "2s").Should(Succeed())
	}

	trafficAllowedRepeatedly := func(address string) {
		GinkgoHelper()

		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(universal.Cluster, demoClient, address)
			g.Expect(err).ToNot(HaveOccurred())
		}, "90s", "3s").MustPassRepeatedly(3).Should(Succeed())
	}

	trafficFaultInjected := func(address string) {
		GinkgoHelper()

		Eventually(func(g Gomega) {
			failure, err := client.CollectFailure(
				universal.Cluster,
				demoClient,
				address,
				client.WithMaxTime(10),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(failure.ResponseCode).To(Equal(503))
		}, "90s", "3s").MustPassRepeatedly(3).Should(Succeed())
	}

	BeforeAll(func() {
		zoneName = universal.Cluster.ZoneName()
		externalServiceDockerName = fmt.Sprintf("%s_%s_%s", universal.Cluster.Name(), meshName, externalServiceApp)

		Expect(NewClusterSetup().
			Install(Yaml(builders.Mesh().
				WithName(meshName).
				WithoutInitialPolicies().
				WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Exclusive))).
			Install(YamlUniversal(zoneProxyMeshIdentity(meshName))).
			Install(YamlUniversal(zoneProxyMeshTrafficPermission(meshName, zoneName))).
			Install(zoneproxy.Install(
				zoneproxy.WithMesh(meshName),
				zoneproxy.WithEgressPort(11102),
				zoneproxy.WithWorkload(egressWorkload),
				zoneproxy.WithDpEnvs(dppEnvs),
			)).
			Install(TestServerExternalServiceUniversal(
				externalServiceApp,
				80,
				false,
				WithDockerContainerName(externalServiceDockerName),
			)).
			Install(DemoClientUniversal(
				demoClient,
				meshName,
				WithTransparentProxy(true),
				WithWorkload(demoClient),
				WithDpEnvs(dppEnvs),
			)).
			Install(YamlUniversal(zoneProxyMeshExternalService(meshName, targetedMES, externalServiceDockerName))).
			Install(YamlUniversal(zoneProxyMeshExternalService(meshName, untargetedMES, externalServiceDockerName))).
			Setup(universal.Cluster)).To(Succeed())

		Eventually(func(g Gomega) {
			out, err := universal.Cluster.GetKumactlOptions().
				RunKumactlAndGetOutput("get", "meshidentity", "-m", meshName, "identity", "-o", "json")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(out).To(ContainSubstring("Successfully initialized"))
		}, "30s", "1s").Should(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, meshName)
	})

	E2EAfterEach(func() {
		Expect(DeleteMeshResources(universal.Cluster, meshName, meshfaultinjection_api.MeshFaultInjectionResourceTypeDescriptor)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteApp(externalServiceApp)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("injects faults only for the MeshExternalService selected by SNI on zone egress", func() {
		targetedURL := fmt.Sprintf("http://%s.extsvc.mesh.local", targetedMES)
		untargetedURL := fmt.Sprintf("http://%s.extsvc.mesh.local", untargetedMES)

		trafficAllowed(targetedURL)
		trafficAllowed(untargetedURL)

		Expect(YamlUniversal(zoneProxyPolicy(
			meshName,
			zoneProxySNI(meshName, zoneName, targetedMES),
		))(universal.Cluster)).To(Succeed())

		trafficFaultInjected(targetedURL)
		trafficAllowedRepeatedly(untargetedURL)
	})
}
