package meshproxypatch

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	meshproxypatch_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshproxypatch/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	. "github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/deployments/zoneproxy"
	"github.com/kumahq/kuma/v3/test/framework/envs/universal"
)

// ZoneProxy verifies that MeshProxyPatch is applied to mesh-scoped zone proxy
// Dataplane resources (those carrying networking.listeners[] of type
// ZoneIngress/ZoneEgress). Unlike the envoyconfig suite, which inspects the
// shadow-generated config, this exercises the full path: a real policy is
// applied, pushed via xDS, and observed on the live Envoy of each zone proxy.
func ZoneProxy() {
	const mesh = "mpp-zone-proxy"
	// proxyName is the zone-proxy deployment base name. zoneproxy.Install
	// derives the app and tunnel keys from it as <proxyName>-ingress and
	// <proxyName>-egress.
	const proxyName = "mpp-zone-proxy"
	const ingressWorkload = proxyName + "-ingress"
	const egressWorkload = proxyName + "-egress"

	// clusterPatch renders a MeshProxyPatch that adds a STATIC cluster. The
	// cluster carries an explicit loadAssignment with an unreachable dummy
	// endpoint so the live Envoy accepts it unconditionally — no traffic is
	// ever sent to it, the cluster's presence is the assertion.
	clusterPatch := func(name, targetRef, clusterName string) string {
		return fmt.Sprintf(`
type: MeshProxyPatch
name: %s
mesh: %s
spec:
  targetRef:
%s
  default:
    appendModifications:
      - cluster:
          operation: Add
          value: |
            name: %s
            connectTimeout: 7s
            type: STATIC
            loadAssignment:
              clusterName: %s
              endpoints:
                - lbEndpoints:
                    - endpoint:
                        address:
                          socketAddress:
                            address: 240.0.0.1
                            portValue: 65000
`, name, mesh, targetRef, clusterName, clusterName)
	}

	// allProxiesCluster targets every Dataplane in the mesh — the only
	// Dataplanes here are the two zone proxies.
	allProxiesCluster := clusterPatch("mpp-all-zone-proxies", "    kind: Dataplane", "mpp-all-cluster")

	// egressScoped / ingressScoped scope to a single zone proxy DPP via the
	// computed labels (kuma.io/listener-zoneegress / -zoneingress).
	egressScoped := clusterPatch("mpp-egress-only", "    kind: Dataplane\n    labels:\n      kuma.io/listener-zoneegress: enabled", "mpp-egress-cluster")
	ingressScoped := clusterPatch("mpp-ingress-only", "    kind: Dataplane\n    labels:\n      kuma.io/listener-zoneingress: enabled", "mpp-ingress-cluster")

	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(Yaml(builders.Mesh().WithName(mesh).WithoutInitialPolicies())).
			Install(MeshTrafficPermissionAllowAllUniversal(mesh)).
			Install(zoneproxy.Install(
				zoneproxy.WithMesh(mesh),
				zoneproxy.WithName(proxyName),
				zoneproxy.WithIngressPort(12001),
				zoneproxy.WithEgressPort(12002),
			)).
			Setup(universal.Cluster)).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, mesh)
	})

	E2EAfterEach(func() {
		Expect(DeleteMeshResources(universal.Cluster, mesh, meshproxypatch_api.MeshProxyPatchResourceTypeDescriptor)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(mesh)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(mesh)).To(Succeed())
	})

	// hasCluster reports whether a cluster with the given name is present on
	// the live Envoy of a zone proxy. g must come from an Eventually callback
	// so xDS propagation delays are retried rather than failing outright.
	hasCluster := func(g Gomega, workload, clusterName string) bool {
		GinkgoHelper()
		cs, err := universal.Cluster.GetAppEnvoyTunnel(workload).GetClusters()
		g.Expect(err).ToNot(HaveOccurred())
		return cs.GetCluster(clusterName) != nil
	}

	It("applies a Dataplane-targeted MeshProxyPatch to every zone proxy", func() {
		// when
		Expect(universal.Cluster.Install(YamlUniversal(allProxiesCluster))).To(Succeed())

		// then — the patched cluster reaches both zone proxy DPPs
		Eventually(func(g Gomega) {
			g.Expect(hasCluster(g, egressWorkload, "mpp-all-cluster")).To(BeTrue())
			g.Expect(hasCluster(g, ingressWorkload, "mpp-all-cluster")).To(BeTrue())
		}, "60s", "2s").Should(Succeed())
	})

	It("scopes a MeshProxyPatch to a single zone proxy via computed labels", func() {
		// when
		Expect(universal.Cluster.Install(YamlUniversal(egressScoped))).To(Succeed())
		Expect(universal.Cluster.Install(YamlUniversal(ingressScoped))).To(Succeed())

		// then — each zone proxy gets only the cluster scoped to it
		Eventually(func(g Gomega) {
			g.Expect(hasCluster(g, egressWorkload, "mpp-egress-cluster")).To(BeTrue())
			g.Expect(hasCluster(g, ingressWorkload, "mpp-ingress-cluster")).To(BeTrue())
		}, "60s", "2s").Should(Succeed())

		// and — the scoping holds: neither proxy carries the other's cluster
		Consistently(func(g Gomega) {
			g.Expect(hasCluster(g, egressWorkload, "mpp-ingress-cluster")).To(BeFalse())
			g.Expect(hasCluster(g, ingressWorkload, "mpp-egress-cluster")).To(BeFalse())
		}, "10s", "2s").Should(Succeed())
	})
}
