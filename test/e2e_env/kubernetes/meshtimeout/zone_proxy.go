package meshtimeout

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	meshretry_api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshretry/api/v1alpha1"
	meshtimeout_api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	. "github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/client"
	"github.com/kumahq/kuma/v2/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/v2/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/v2/test/framework/deployments/zoneproxy"
	"github.com/kumahq/kuma/v2/test/framework/envs/kubernetes"
)

func meshTimeoutZoneProxyMeshIdentity(meshName string) string {
	return fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshIdentity
metadata:
  name: identity-%[1]s
  namespace: %[2]s
  labels:
    kuma.io/mesh: %[1]s
    kuma.io/origin: zone
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
`, meshName, Config.KumaNamespace)
}

func meshTimeoutZoneProxyMTP(meshName string) string {
	return fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTrafficPermission
metadata:
  name: allow-all-ze-%[1]s
  namespace: %[2]s
  labels:
    kuma.io/mesh: %[1]s
spec:
  targetRef:
    kind: Mesh
  rules:
    - default:
        allow:
          - sni:
              type: Exact
              value: "sni.extsvc.%[1]s.default.%[2]s.external-server-%[1]s.80"
`, meshName, Config.KumaNamespace)
}

func meshTimeoutZoneProxyMES(meshName, extNamespace string) string {
	return fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshExternalService
metadata:
  name: external-server-%[1]s
  namespace: %[2]s
  labels:
    kuma.io/mesh: %[1]s
    kuma.io/origin: zone
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: http
  endpoints:
  - address: external-server.%[3]s.svc.cluster.local
    port: 80
`, meshName, Config.KumaNamespace, extNamespace)
}

func meshTimeoutZoneProxyPolicy(meshName string) string {
	return fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: timeout-zone-proxy-%[1]s
  namespace: %[2]s
  labels:
    kuma.io/mesh: %[1]s
spec:
  targetRef:
    kind: Mesh
  rules:
    - matches:
        - sni:
            type: Exact
            value: "sni.extsvc.%[1]s.default.%[2]s.external-server-%[1]s.80"
      default:
        idleTimeout: 20s
        http:
          requestTimeout: 2s
          maxStreamDuration: 20s
`, meshName, Config.KumaNamespace)
}

func ZoneProxyMeshTimeout() {
	ns := "meshtimeout-zoneproxy"
	extNs := "meshtimeout-zoneproxy-ext"
	mesh := "meshtimeout-zoneproxy"
	workload := "zone-egress"
	const egressPort = uint32(11102)

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(NamespaceWithSidecarInjection(ns)).
			Install(Namespace(extNs)).
			Install(MeshWithMeshServicesKubernetes(mesh, "Exclusive")).
			Install(YamlK8s(meshTimeoutZoneProxyMTP(mesh))).
			Install(Parallel(
				democlient.Install(democlient.WithNamespace(ns), democlient.WithMesh(mesh)),
				testserver.Install(
					testserver.WithNamespace(extNs),
					testserver.WithName("external-server"),
					testserver.WithEchoArgs("echo", "--instance", "external-server"),
				),
				zoneproxy.Install(
					zoneproxy.WithName("zp-meshtimeout"),
					zoneproxy.WithNamespace(ns),
					zoneproxy.WithMesh(mesh),
					zoneproxy.WithWorkload(workload),
					zoneproxy.WithEgressPort(egressPort),
				),
			)).
			Install(YamlK8s(meshTimeoutZoneProxyMeshIdentity(mesh))).
			Install(YamlK8s(meshTimeoutZoneProxyMES(mesh, extNs))).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugKube(kubernetes.Cluster, mesh, ns, extNs)
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(ns)).To(Succeed())
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(extNs)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(mesh)).To(Succeed())
	})

	It("zone-proxy should timeout MeshExternalService traffic selected by SNI", FlakeAttempts(3), func() {
		url := fmt.Sprintf("http://external-server-%s.extsvc.mesh.local", mesh)

		Expect(DeleteMeshResources(kubernetes.Cluster, mesh,
			meshtimeout_api.MeshTimeoutResourceTypeDescriptor,
			meshretry_api.MeshRetryResourceTypeDescriptor,
		)).To(Succeed())

		Eventually(func(g Gomega) {
			start := time.Now()
			resp, err := client.CollectEchoResponse(
				kubernetes.Cluster,
				"demo-client",
				url,
				client.FromKubernetesPod(ns, "demo-client"),
				client.WithHeader("x-set-response-delay-ms", "5000"),
				client.WithMaxTime(10),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(ContainSubstring("external-server"))
			g.Expect(time.Since(start)).To(BeNumerically(">", 5*time.Second))
		}, "90s", "3s").Should(Succeed())

		Expect(YamlK8s(meshTimeoutZoneProxyPolicy(mesh))(kubernetes.Cluster)).To(Succeed())

		Eventually(func(g Gomega) {
			start := time.Now()
			failure, err := client.CollectFailure(
				kubernetes.Cluster,
				"demo-client",
				url,
				client.FromKubernetesPod(ns, "demo-client"),
				client.WithHeader("x-set-response-delay-ms", "5000"),
				client.WithMaxTime(10),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(failure.ResponseCode).To(Equal(504))
			g.Expect(time.Since(start)).To(BeNumerically("<", 5*time.Second))
		}, "90s", "3s", MustPassRepeatedly(5)).Should(Succeed())
	})
}
