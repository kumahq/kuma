package meshmetric

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/sync/errgroup"

	meshmetric_api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshmetric/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
	util_proto "github.com/kumahq/kuma/v2/pkg/util/proto"
	. "github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/deployments/zoneproxy"
	"github.com/kumahq/kuma/v2/test/framework/envs/multizone"
)

func ZoneProxy() {
	const mesh = "mm-zone-proxy"

	policyWithApplications := fmt.Sprintf(`
type: MeshMetric
name: mm-zone-proxy-policy
mesh: %s
spec:
  targetRef:
    kind: Mesh
  default:
    applications:
      - name: ignored-on-zone-proxy
        path: /metrics
        port: 8888
    backends:
      - type: Prometheus
        prometheus:
          port: 5670
          path: /metrics
          tls:
            mode: Disabled
`, mesh)

	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(Yaml(builders.Mesh().WithName(mesh))).
			Install(MeshTrafficPermissionAllowAllUniversal(mesh)).
			Setup(multizone.Global)).To(Succeed())
		Expect(WaitForMesh(mesh, multizone.Zones())).To(Succeed())

		group := errgroup.Group{}
		NewClusterSetup().
			Install(Parallel(
				zoneproxy.Install(
					zoneproxy.WithMesh(mesh),
					zoneproxy.WithIngressPort(11001),
					zoneproxy.WithWorkload("zone-proxy-ingress"),
				),
				zoneproxy.Install(
					zoneproxy.WithMesh(mesh),
					zoneproxy.WithEgressPort(11002),
					zoneproxy.WithWorkload("zone-proxy-egress"),
				),
			)).
			SetupInGroup(multizone.UniZone1, &group)
		Expect(group.Wait()).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(multizone.Global, mesh)
		DebugUniversal(multizone.UniZone1, mesh)
	})

	E2EAfterEach(func() {
		Expect(DeleteMeshResources(multizone.Global, mesh, meshmetric_api.MeshMetricResourceTypeDescriptor)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(multizone.UniZone1.DeleteMeshApps(mesh)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(mesh)).To(Succeed())
	})

	// dynConfigJSON returns the listeners config_dump of the given zone-proxy
	// DPP as JSON. The meshmetric dynconf payload is embedded as the
	// `inlineString` body of the `_kuma:dynamicconfig` listener's direct
	// response, so a substring search over the marshaled listeners is enough
	// to assert on the contract without walking the proto structure.
	// g must come from an Eventually callback so that failures are retryable.
	dynConfigJSON := func(g Gomega, appName string) string {
		GinkgoHelper()
		tnl := multizone.UniZone1.GetAppEnvoyTunnel(appName)
		cd, err := tnl.GetConfigDump()
		g.Expect(err).ToNot(HaveOccurred())
		if cd == nil {
			return ""
		}
		raw, err := util_proto.ToJSON(&cd.Listeners)
		g.Expect(err).ToNot(HaveOccurred())
		return string(raw)
	}

	It("MeshMetric on zone-proxy-only DPP clears applications and uses DPP name as service label", func() {
		// given
		Expect(multizone.Global.Install(YamlUniversal(policyWithApplications))).To(Succeed())

		// then — zone-egress
		Eventually(func(g Gomega) {
			payload := dynConfigJSON(g, "zone-proxy-egress")
			// MeshMetric reached the proxy and emitted the dynconf listener.
			g.Expect(payload).To(ContainSubstring("_kuma:dynamicconfig"))
			// applications[] must be cleared on a zone-proxy-only DPP.
			g.Expect(payload).To(ContainSubstring(`"applications":null`))
			g.Expect(payload).ToNot(ContainSubstring("ignored-on-zone-proxy"))
			// proxy_role label identifies the proxy's purpose.
			g.Expect(payload).To(ContainSubstring(`"kuma.proxy_role":"zone-egress"`))
			// service label falls back to the DPP name, never to ServiceUnknown.
			g.Expect(payload).To(ContainSubstring(`"service":"zone-proxy-egress"`))
			g.Expect(payload).ToNot(ContainSubstring(`"service":"unknown"`))
			// kuma.workload is suppressed: there is no co-located workload.
			g.Expect(payload).ToNot(ContainSubstring(`"kuma.workload":`))
		}, "60s", "2s").Should(Succeed())

		// then — zone-ingress
		Eventually(func(g Gomega) {
			payload := dynConfigJSON(g, "zone-proxy-ingress")
			g.Expect(payload).To(ContainSubstring("_kuma:dynamicconfig"))
			g.Expect(payload).To(ContainSubstring(`"applications":null`))
			g.Expect(payload).ToNot(ContainSubstring("ignored-on-zone-proxy"))
			g.Expect(payload).To(ContainSubstring(`"kuma.proxy_role":"zone-ingress"`))
			g.Expect(payload).To(ContainSubstring(`"service":"zone-proxy-ingress"`))
			g.Expect(payload).ToNot(ContainSubstring(`"service":"unknown"`))
			g.Expect(payload).ToNot(ContainSubstring(`"kuma.workload":`))
		}, "60s", "2s").Should(Succeed())
	})
}
