package observability

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/client"
	"github.com/kumahq/kuma/v3/test/framework/deployments/democlient"
	obs "github.com/kumahq/kuma/v3/test/framework/deployments/observability"
	"github.com/kumahq/kuma/v3/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/v3/test/framework/deployments/zoneproxy"
	"github.com/kumahq/kuma/v3/test/framework/envs/kubernetes"
)

func meshTraceZoneProxyMeshIdentity(meshName string) string {
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

func meshTraceZoneProxyMTP(meshName string) string {
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

func meshTraceZoneProxyMES(meshName, extNamespace string) string {
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

func meshTraceZoneProxyPolicy(meshName, zipkinURL string) string {
	return fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTrace
metadata:
  name: trace-%[1]s
  namespace: %[2]s
  labels:
    kuma.io/mesh: %[1]s
spec:
  targetRef:
    kind: Mesh
  default:
    backends:
    - type: Zipkin
      zipkin:
        url: %[3]s
    sampling:
      overall: 100
`, meshName, Config.KumaNamespace, zipkinURL)
}

// ZoneProxyPluginTest exercises MeshTrace applied to a mesh-scoped zone-egress
// Dataplane (MADR-098). It verifies that:
//  1. The xDS pipeline generates a self_zoneegress listener with the meshtrace
//     tracing block once MeshIdentity + MeshExternalService are present.
//  2. Real outbound traffic from a regular sidecar through the zone-egress
//     produces spans at the collector with the expected kuma.* tags, including
//     kuma.workload carrying the workload-label-derived identifier (which is
//     stable across pod restarts on K8s, unlike pod.Name).
//
// Stability notes:
//   - All waits use Eventually with generous timeouts to absorb MeshIdentity
//     Ready latency and Envoy Zipkin batch flush (default 5s).
//   - The Zipkin tracer fills localEndpoint.serviceName from Envoy bootstrap,
//     so zone-egress spans land under service "unknown" in Jaeger. The
//     assertion filters by tag instead of service name; see issue #16602.
func ZoneProxyPluginTest() {
	ns := "meshtrace-zoneproxy"
	extNs := "meshtrace-zoneproxy-ext"
	obsNs := "obs-meshtrace-zoneproxy"
	obsDeployment := "obs-trace-zoneproxy-deployment"
	mesh := "meshtrace-zoneproxy"
	workload := "zone-egress"
	const egressPort = uint32(11102)

	var obsClient obs.Observability
	BeforeAll(func() {
		err := NewClusterSetup().
			Install(NamespaceWithSidecarInjection(ns)).
			Install(Namespace(extNs)).
			Install(MeshWithMeshServicesKubernetes(mesh, "Exclusive")).
			Install(YamlK8s(meshTraceZoneProxyMTP(mesh))).
			Install(Parallel(
				democlient.Install(democlient.WithNamespace(ns), democlient.WithMesh(mesh)),
				testserver.Install(
					testserver.WithNamespace(extNs),
					testserver.WithName("external-server"),
					testserver.WithEchoArgs("echo", "--instance", "external-server"),
				),
				zoneproxy.Install(
					zoneproxy.WithName("zp-meshtrace"),
					zoneproxy.WithNamespace(ns),
					zoneproxy.WithMesh(mesh),
					zoneproxy.WithWorkload(workload),
					zoneproxy.WithEgressPort(egressPort),
				),
				obs.Install(obsDeployment, obs.WithNamespace(obsNs), obs.WithComponents(obs.JaegerComponent)),
			)).
			Install(YamlK8s(meshTraceZoneProxyMeshIdentity(mesh))).
			Install(YamlK8s(meshTraceZoneProxyMES(mesh, extNs))).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())
		obsClient = obs.From(obsDeployment, kubernetes.Cluster)
	})

	AfterEachFailure(func() {
		DebugKube(kubernetes.Cluster, mesh, ns, extNs, obsNs)
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(ns)).To(Succeed())
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(extNs)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(mesh)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteDeployment(obsDeployment)).To(Succeed())
	})

	It("should emit spans from zone-egress carrying the workload-label tag", func() {
		Expect(YamlK8s(meshTraceZoneProxyPolicy(mesh, obsClient.ZipkinCollectorURL()))(kubernetes.Cluster)).To(Succeed())

		// Drive traffic from the regular sidecar through the zone-egress to the
		// MeshExternalService. In Exclusive mode with a mesh-scoped zone-egress
		// available, MES traffic is automatically routed through it.
		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				kubernetes.Cluster, "demo-client",
				fmt.Sprintf("http://external-server-%s.extsvc.mesh.local", mesh),
				client.FromKubernetesPod(ns, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "90s", "3s").Should(Succeed())

		// Verify a span with kuma.workload tag containing the workload name
		// reached the collector. We filter by tag, not service, because the
		// Zipkin tracer reports localEndpoint.serviceName="unknown" for
		// zone-egress spans (see issue #16602).
		Eventually(func(g Gomega) {
			traces, err := obsClient.TracesForService("unknown", 50)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(traces).ToNot(BeEmpty(), "no zone-egress traces at collector yet")

			var foundWorkload, foundMesh, foundZone bool
			for _, t := range traces {
				for _, span := range t.Spans {
					for _, tag := range span.Tags {
						switch tag.Key {
						case "kuma.workload":
							if strings.Contains(tag.Value, workload) {
								foundWorkload = true
							}
						case "kuma.mesh":
							if tag.Value == mesh {
								foundMesh = true
							}
						case "kuma.zone":
							if tag.Value != "" {
								foundZone = true
							}
						}
					}
				}
			}
			g.Expect(foundWorkload).To(BeTrue(),
				"no span with kuma.workload tag containing %q in %d traces", workload, len(traces))
			g.Expect(foundMesh).To(BeTrue(),
				"no span with kuma.mesh tag matching %q", mesh)
			g.Expect(foundZone).To(BeTrue(),
				"no span with non-empty kuma.zone tag")
		}, "120s", "5s").Should(Succeed())
	})
}
