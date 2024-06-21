package meshexternalservices

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envoy_admin/stats"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func MeshExternalServices() {
	meshName := "mesh-external-services"
	meshNameEgress := "mesh-external-services-egress"
	namespace := "mesh-external-services"
	clientNamespace := "client-mesh-external-services"

	egressMesh := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: %s
spec:
  mtls:
    enabledBackend: ca-1
    backends:
      - name: ca-1
        type: builtin
  networking:
    outbound:
      passthrough: false
  routing:
    zoneEgress: true
`, meshNameEgress)

	hostnameGenerator := func(mesh string) string {
		return fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: HostnameGenerator
metadata:
  labels:
    kuma.io/mesh: %s
  name: mes-hg
  namespace: %s
spec:
  selector:
    meshExternalService:
      matchLabels:
        hostname: "true"
  template: "{{ .Name }}.mesh"
`, mesh, Config.KumaNamespace)
	}

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshKubernetes(meshName)).
			Install(YamlK8s(egressMesh)).
			Install(Namespace(namespace)).
			Install(NamespaceWithSidecarInjection(clientNamespace)).
			Install(democlient.Install(democlient.WithNamespace(clientNamespace), democlient.WithMesh(meshName))).
			Install(democlient.Install(democlient.WithNamespace(clientNamespace), democlient.WithName("demo-client-egress"), democlient.WithMesh(meshNameEgress))).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugKube(kubernetes.Cluster, meshName, namespace, clientNamespace)
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(clientNamespace)).To(Succeed())
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	Context("http non-TLS", func() {
		meshExternalService := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshExternalService
metadata:
  name: http-external-service
  namespace: %s
  labels:
    kuma.io/mesh: mesh-external-services
    hostname: "true"
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: http
  endpoints:
    - address: external-service.mesh-external-services.svc.cluster.local
      port: 80
`, Config.KumaNamespace)

		BeforeAll(func() {
			err := kubernetes.Cluster.Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithName("external-service"),
			))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should route to http external-service", func() {
			// given working communication outside the mesh with passthrough enabled and no traffic permission
			Eventually(func(g Gomega) {
				_, err := client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client", "external-service.mesh-external-services",
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
			}, "30s", "1s").Should(Succeed())

			// when apply external service and hostname generator
			Expect(kubernetes.Cluster.Install(YamlK8s(meshExternalService))).To(Succeed())
			Expect(kubernetes.Cluster.Install(YamlK8s(hostnameGenerator(meshName)))).To(Succeed())

			// and you can also use .mesh on port of the provided host
			Eventually(func(g Gomega) {
				_, err := client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client", "http-external-service.mesh",
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
			}, "30s", "1s").Should(Succeed())
		})
	})

	Context("http non-TLS with egress", func() {
		meshExternalService := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshExternalService
metadata:
  name: mesh-external-service-egress
  namespace: %s
  labels:
    kuma.io/mesh: %s
    hostname: "true"
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: http
  endpoints:
    - address: external-service-egress.mesh-external-services.svc.cluster.local
      port: 80
`, Config.KumaNamespace, meshNameEgress)

		filter := fmt.Sprintf(
			"cluster.%s_%s_%s.upstream_rq_total",
			meshNameEgress,
			"meshexternalservice_mesh-external-service-egress",
			Config.KumaNamespace,
		)
		BeforeAll(func() {
			err := kubernetes.Cluster.Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithName("external-service-egress"),
			))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should route to external-service", func() {
			// when apply external service and hostname generator
			Expect(kubernetes.Cluster.Install(YamlK8s(meshExternalService))).To(Succeed())
			Expect(kubernetes.Cluster.Install(YamlK8s(hostnameGenerator(meshNameEgress)))).To(Succeed())

			// then traffic doesn't work because of missing MeshTrafficPermission
			Eventually(func(g Gomega) {
				response, err := client.CollectFailure(
					kubernetes.Cluster, "demo-client-egress", "mesh-external-service-egress.mesh",
					client.FromKubernetesPod(clientNamespace, "demo-client-egress"),
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.ResponseCode).To(Equal(403))
			}, "30s", "1s").Should(Succeed())

			// when MTP targeting MeshExternalService added
			Expect(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTrafficPermission
metadata:
  name: mtp1
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: MeshExternalService
    name: mesh-external-service-egress
  from:
    - targetRef:
        kind: Mesh
      default:
        action: Allow
`, Config.KumaNamespace, meshNameEgress))(kubernetes.Cluster)).To(Succeed())

			// then traffic works
			Eventually(func(g Gomega) {
				_, err := client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client-egress", "mesh-external-service-egress.mesh",
					client.FromKubernetesPod(clientNamespace, "demo-client-egress"),
				)
				g.Expect(err).ToNot(HaveOccurred())
			}, "30s", "1s").Should(Succeed())

			// and flows through Egress
			Eventually(func(g Gomega) {
				stat, err := kubernetes.Cluster.GetZoneEgressEnvoyTunnel().GetStats(filter)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(stat).To(stats.BeGreaterThanZero())
			}, "30s", "1s").Should(Succeed())
		})
	})

	Context("tcp non-TLS", func() {
		meshExternalService := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshExternalService
metadata:
  name: tcp-external-service
  namespace: %s
  labels:
    kuma.io/mesh: mesh-external-services
    hostname: "true"
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: tcp
  endpoints:
    - address: tcp-external-service.mesh-external-services.svc.cluster.local
      port: 80
`, Config.KumaNamespace)

		BeforeAll(func() {
			err := kubernetes.Cluster.Install(testserver.Install(
				testserver.WithName("tcp-external-service"),
				testserver.WithServicePortAppProtocol("tcp"),
				testserver.WithMesh(meshName),
				testserver.WithNamespace(namespace),
			))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should route to tcp external-service", func() {
			// given working communication outside the mesh with passthrough enabled and no traffic permission
			Eventually(func(g Gomega) {
				_, err := client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client", "tcp-external-service.mesh-external-services",
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
			}, "30s", "1s").Should(Succeed())

			// when apply external service and hostname generator
			Expect(kubernetes.Cluster.Install(YamlK8s(meshExternalService))).To(Succeed())
			Expect(kubernetes.Cluster.Install(YamlK8s(hostnameGenerator(meshName)))).To(Succeed())

			// and you can also use .mesh on port of the provided host
			Eventually(func(g Gomega) {
				_, err := client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client", "tcp-external-service.mesh",
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
			}, "30s", "1s").Should(Succeed())
		})
	})

	Context("HTTPS", func() {
		tlsExternalService := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshExternalService
metadata:
  name: tls-external-service
  namespace: %s
  labels:
    kuma.io/mesh: mesh-external-services
    hostname: "true"
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: http
  endpoints:
    - address: tls-external-service.mesh-external-services.svc.cluster.local
      port: 80
  tls:
    enabled: true
    verification:
      mode: SkipCA # test-server certificate is not signed by a CA that is in the system trust store
`, Config.KumaNamespace)

		BeforeAll(func() {
			err := NewClusterSetup().
				Install(testserver.Install(
					testserver.WithNamespace(namespace),
					testserver.WithEchoArgs("--tls", "--crt=/kuma/server.crt", "--key=/kuma/server.key"),
					testserver.WithName("tls-external-service"),
					testserver.WithoutProbes(), // not compatible with TLS
				)).
				Install(YamlK8s(tlsExternalService)).
				Install(YamlK8s(hostnameGenerator(meshName))).
				Setup(kubernetes.Cluster)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should route to tls external-service", func() {
			Eventually(func(g Gomega) {
				_, err := client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client", "tls-external-service.mesh",
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
			}, "30s", "1s").Should(Succeed())
		})
	})
}
