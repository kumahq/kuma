package meshexternalservices

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	meshexternalservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
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

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(YamlK8s(samples.MeshMTLSBuilder().
				WithName(meshName).
				WithEgressRoutingEnabled().KubeYaml())).
			Install(YamlK8s(samples.MeshMTLSBuilder().
				WithName(meshNameEgress).
				WithoutPassthrough().
				WithMeshExternalServiceTrafficForbidden().
				WithEgressRoutingEnabled().KubeYaml())).
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
    kuma.io/mesh: %s
    hostname: "true"
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: http
  endpoints:
    - address: external-service.mesh-external-services.svc.cluster.local
      port: 80
`, Config.KumaNamespace, meshName)

		BeforeAll(func() {
			err := kubernetes.Cluster.Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithName("external-service"),
			))
			Expect(err).ToNot(HaveOccurred())
		})

		filter := fmt.Sprintf(
			"cluster.%s_%s_%s_default_extsvc_80.upstream_rq_total",
			meshName,
			"http-external-service",
			Config.KumaNamespace,
		)

		It("should route to http external-service", func() {
			// given working communication outside the mesh with passthrough enabled and no traffic permission
			Eventually(func(g Gomega) {
				_, err := client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client", "external-service.mesh-external-services",
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
			}, "30s", "1s").Should(Succeed())

			// when apply external service
			Expect(kubernetes.Cluster.Install(YamlK8s(meshExternalService))).To(Succeed())

			// and you can also use .mesh on port of the provided host
			Eventually(func(g Gomega) {
				_, err := client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client", "http-external-service.extsvc.mesh.local",
					client.FromKubernetesPod(clientNamespace, "demo-client"),
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

	Context("http non-TLS with rbac switch", func() {
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
			"cluster.%s_%s_%s_default_extsvc_80.upstream_rq_total",
			meshNameEgress,
			"mesh-external-service-egress",
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

			// then traffic doesn't work because of missing MeshTrafficPermission
			Eventually(func(g Gomega) {
				response, err := client.CollectFailure(
					kubernetes.Cluster, "demo-client-egress", "mesh-external-service-egress.extsvc.mesh.local",
					client.FromKubernetesPod(clientNamespace, "demo-client-egress"),
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.ResponseCode).To(Equal(403))
			}, "30s", "1s").Should(Succeed())

			// when allow all traffic
			Expect(kubernetes.Cluster.Install(YamlK8s(
				samples.MeshMTLSBuilder().
					WithName(meshNameEgress).
					WithoutPassthrough().
					WithEgressRoutingEnabled().KubeYaml()),
			)).To(Succeed())

			// then traffic works
			Eventually(func(g Gomega) {
				_, err := client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client-egress", "mesh-external-service-egress.extsvc.mesh.local",
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
    kuma.io/mesh: %s
    hostname: "true"
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: tcp
  endpoints:
    - address: tcp-external-service.mesh-external-services.svc.cluster.local
      port: 80
`, Config.KumaNamespace, meshName)
		filter := fmt.Sprintf(
			"cluster.%s_%s_%s_default_extsvc_80.upstream_rq_total",
			meshName,
			"tcp-external-service",
			Config.KumaNamespace,
		)
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

			// when apply external service
			Expect(kubernetes.Cluster.Install(YamlK8s(meshExternalService))).To(Succeed())

			// and you can also use .mesh on port of the provided host
			Eventually(func(g Gomega) {
				_, err := client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client", "tcp-external-service.extsvc.mesh.local",
					client.FromKubernetesPod(clientNamespace, "demo-client"),
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

	Context("HTTPS", func() {
		tlsExternalService := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshExternalService
metadata:
  name: tls-external-service
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
    - address: tls-external-service.mesh-external-services.svc.cluster.local
      port: 80
  tls:
    enabled: true
    verification:
      mode: SkipCA # test-server certificate is not signed by a CA that is in the system trust store
`, Config.KumaNamespace, meshName)
		filter := fmt.Sprintf(
			"cluster.%s_%s_%s_default_extsvc_80.upstream_rq_total", // cx
			meshName,
			"tls-external-service",
			Config.KumaNamespace,
		)
		BeforeAll(func() {
			err := NewClusterSetup().
				Install(testserver.Install(
					testserver.WithNamespace(namespace),
					testserver.WithEchoArgs("--tls", "--crt=/kuma/server.crt", "--key=/kuma/server.key"),
					testserver.WithName("tls-external-service"),
					testserver.WithoutProbes(), // not compatible with TLS
				)).
				Install(YamlK8s(tlsExternalService)).
				Setup(kubernetes.Cluster)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should route to tls external-service", func() {
			Eventually(func(g Gomega) {
				_, err := client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client", "tls-external-service.extsvc.mesh.local",
					client.FromKubernetesPod(clientNamespace, "demo-client"),
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

	Context("http service with MeshHTTPRoute", func() {
		meshExternalService := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshExternalService
metadata:
  name: plain-external-service
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
    - address: plain-external-service.mesh-external-services.svc.cluster.local
      port: 80
`, Config.KumaNamespace, meshName)

		meshExternalService2 := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshExternalService
metadata:
  name: external-service-with-httproute
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
    - address: external-service-with-httproute.mesh-external-services.svc.cluster.local
      port: 80
`, Config.KumaNamespace, meshName)

		meshHttpRoutePolicy := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: http-route-mes-policy
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshExternalService
        name: plain-external-service
      rules:
        - matches:
            - path:
                type: PathPrefix
                value: /
          default:
            backendRefs:
              - kind: MeshExternalService
                name: external-service-with-httproute
                port: 80
                weight: 100
`, Config.KumaNamespace, meshName)

		BeforeAll(func() {
			err := NewClusterSetup().
				Install(testserver.Install(
					testserver.WithNamespace(namespace),
					testserver.WithName("plain-external-service"),
					testserver.WithEchoArgs("echo", "--instance", "plain-external-service"),
				)).
				Install(testserver.Install(
					testserver.WithNamespace(namespace),
					testserver.WithName("external-service-with-httproute"),
					testserver.WithEchoArgs("echo", "--instance", "external-service-with-httproute"),
				)).
				Setup(kubernetes.Cluster)
			Expect(err).ToNot(HaveOccurred())
		})

		E2EAfterEach(func() {
			Expect(DeleteMeshResources(kubernetes.Cluster, meshName,
				meshhttproute_api.MeshHTTPRouteResourceTypeDescriptor,
				meshexternalservice_api.MeshExternalServiceResourceTypeDescriptor,
			)).To(Succeed())
		})

		It("should route to http external-service", func() {
			// when external service added
			Expect(kubernetes.Cluster.Install(YamlK8s(meshExternalService))).To(Succeed())

			// then communication works
			Eventually(func(g Gomega) {
				resp, err := client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client", "plain-external-service.extsvc.mesh.local",
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(resp.Instance).To(Equal("plain-external-service"))
			}, "30s", "1s").Should(Succeed())

			// when
			Expect(kubernetes.Cluster.Install(YamlK8s(meshExternalService2))).To(Succeed())

			// and added route
			Expect(kubernetes.Cluster.Install(YamlK8s(meshHttpRoutePolicy))).To(Succeed())

			// then traffic is routed to the 2nd MeshExternalService
			Eventually(func(g Gomega) {
				resp, err := client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client", "plain-external-service.extsvc.mesh.local",
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(resp.Instance).To(Equal("external-service-with-httproute"))
			}, "30s", "1s").Should(Succeed())
		})
	})
}
