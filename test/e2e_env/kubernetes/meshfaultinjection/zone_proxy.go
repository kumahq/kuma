package meshfaultinjection

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	meshfaultinjection_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshfaultinjection/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	. "github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/client"
	"github.com/kumahq/kuma/v3/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/v3/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/v3/test/framework/deployments/zoneproxy"
	"github.com/kumahq/kuma/v3/test/framework/envs/kubernetes"
)

func zoneProxyMeshIdentity(meshName string) string {
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

func zoneProxyMeshExternalService(resourceName, meshName, extNamespace string) string {
	return fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshExternalService
metadata:
  name: %[1]s
  namespace: %[2]s
  labels:
    kuma.io/mesh: %[3]s
    kuma.io/origin: zone
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: http
  endpoints:
    - address: external-server.%[4]s.svc.cluster.local
      port: 80
`, resourceName, Config.KumaNamespace, meshName, extNamespace)
}

func zoneProxyMeshTrafficPermission(meshName string, allowedSNIs ...string) string {
	allow := make([]string, 0, len(allowedSNIs))
	for _, sni := range allowedSNIs {
		allow = append(allow, fmt.Sprintf(`
          - sni:
              type: Exact
              value: %q`, sni))
	}

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
        allow:%[3]s
`, meshName, Config.KumaNamespace, strings.Join(allow, ""))
}

func zoneProxyPolicy(meshName, targetSNI string) string {
	return fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshFaultInjection
metadata:
  name: mfi-zone-proxy-%[1]s
  namespace: %[2]s
  labels:
    kuma.io/mesh: %[1]s
spec:
  targetRef:
    kind: Dataplane
    labels:
      kuma.io/listener-zoneegress: enabled
  rules:
    - matches:
        - sni:
            type: Exact
            value: %[3]q
      default:
        http:
          - abort:
              httpStatus: 503
              percentage: 100
`, meshName, Config.KumaNamespace, targetSNI)
}

func zoneProxySNI(meshName, resourceName string) string {
	return fmt.Sprintf("sni.extsvc.%s.default.%s.%s.80", meshName, Config.KumaNamespace, resourceName)
}

func ZoneProxy() {
	const ns = "mfi-zoneproxy"
	const extNs = "mfi-zoneproxy-ext"
	const mesh = "mfi-zoneproxy"
	const workload = "zone-egress"
	const targetedMES = "external-target-mfi-zoneproxy"
	const untargetedMES = "external-other-mfi-zoneproxy"
	const demoClient = "demo-client"
	const externalServer = "external-server"
	const egressPort = uint32(11102)

	fromDemoClient := client.FromKubernetesPod(ns, demoClient)

	trafficAllowedWithPasses := func(address string, mustPassRepeatedly int) {
		GinkgoHelper()

		eventually := Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				kubernetes.Cluster,
				demoClient,
				address,
				fromDemoClient,
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(ContainSubstring(externalServer))
		}, "90s", "3s")

		if mustPassRepeatedly > 0 {
			eventually.MustPassRepeatedly(mustPassRepeatedly).Should(Succeed())
			return
		}

		eventually.Should(Succeed())
	}

	trafficAllowed := func(address string) {
		GinkgoHelper()

		trafficAllowedWithPasses(address, 0)
	}

	trafficAllowedRepeatedly := func(address string) {
		GinkgoHelper()

		trafficAllowedWithPasses(address, 3)
	}

	trafficFaultInjected := func(address string) {
		GinkgoHelper()

		Eventually(func(g Gomega) {
			failure, err := client.CollectFailure(
				kubernetes.Cluster,
				demoClient,
				address,
				fromDemoClient,
				client.WithMaxTime(10),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(failure.ResponseCode).To(Equal(503))
		}, "90s", "3s").MustPassRepeatedly(3).Should(Succeed())
	}

	BeforeAll(func() {
		targetedSNI := zoneProxySNI(mesh, targetedMES)
		untargetedSNI := zoneProxySNI(mesh, untargetedMES)

		Expect(NewClusterSetup().
			Install(NamespaceWithSidecarInjection(ns)).
			Install(Namespace(extNs)).
			Install(Yaml(builders.Mesh().
				WithName(mesh).
				WithoutInitialPolicies().
				WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Exclusive))).
			Install(YamlK8s(zoneProxyMeshTrafficPermission(mesh, targetedSNI, untargetedSNI))).
			Install(Parallel(
				democlient.Install(democlient.WithNamespace(ns), democlient.WithMesh(mesh)),
				testserver.Install(
					testserver.WithNamespace(extNs),
					testserver.WithName(externalServer),
					testserver.WithEchoArgs("echo", "--instance", externalServer),
				),
				zoneproxy.Install(
					zoneproxy.WithName("zp-meshfaultinjection"),
					zoneproxy.WithNamespace(ns),
					zoneproxy.WithMesh(mesh),
					zoneproxy.WithWorkload(workload),
					zoneproxy.WithEgressPort(egressPort),
				),
			)).
			Install(YamlK8s(zoneProxyMeshIdentity(mesh))).
			Install(YamlK8s(zoneProxyMeshExternalService(targetedMES, mesh, extNs))).
			Install(YamlK8s(zoneProxyMeshExternalService(untargetedMES, mesh, extNs))).
			Setup(kubernetes.Cluster)).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugKube(kubernetes.Cluster, mesh, ns, extNs)
	})

	E2EAfterEach(func() {
		Expect(DeleteMeshResources(kubernetes.Cluster, mesh, meshfaultinjection_api.MeshFaultInjectionResourceTypeDescriptor)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(ns)).To(Succeed())
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(extNs)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(mesh)).To(Succeed())
	})

	It("injects faults only for the MeshExternalService selected by SNI on zone egress", func() {
		targetedURL := fmt.Sprintf("http://%s.extsvc.mesh.local", targetedMES)
		untargetedURL := fmt.Sprintf("http://%s.extsvc.mesh.local", untargetedMES)

		trafficAllowed(targetedURL)
		trafficAllowed(untargetedURL)

		Expect(YamlK8s(zoneProxyPolicy(
			mesh,
			zoneProxySNI(mesh, targetedMES),
		))(kubernetes.Cluster)).To(Succeed())

		trafficFaultInjected(targetedURL)
		trafficAllowedRepeatedly(untargetedURL)
	})
}
