package meshexternalservices

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func MeshExternalServices() {
	meshName := "mesh-external-services"
	namespace := "mesh-external-services"
	clientNamespace := "client-mesh-external-services"

	hostnameGenerator := `
apiVersion: kuma.io/v1alpha1
kind: HostnameGenerator
metadata:
  labels:
    kuma.io/mesh: mesh-external-services
  name: mes-hg
spec:
  selector:
    meshExternalService:
      matchLabels:
        hostname: "true"
  template: "{{ .Name }}.mesh"
`

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshKubernetes(meshName)).
			Install(Namespace(namespace)).
			Install(NamespaceWithSidecarInjection(clientNamespace)).
			Install(democlient.Install(democlient.WithNamespace(clientNamespace), democlient.WithMesh(meshName))).
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
		meshExternalService := `
apiVersion: kuma.io/v1alpha1
kind: MeshExternalService
metadata:
  name: http-external-service
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
`

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
			Expect(kubernetes.Cluster.Install(YamlK8s(hostnameGenerator))).To(Succeed())

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

	Context("tcp non-TLS", func() {
		meshExternalService := `
apiVersion: kuma.io/v1alpha1
kind: MeshExternalService
metadata:
  name: tcp-external-service
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
`

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
			Expect(kubernetes.Cluster.Install(YamlK8s(hostnameGenerator))).To(Succeed())

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
		tlsExternalService := `
apiVersion: kuma.io/v1alpha1
kind: MeshExternalService
metadata:
  name: tls-external-service
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
`

		BeforeAll(func() {
			err := NewClusterSetup().
				Install(testserver.Install(
					testserver.WithNamespace(namespace),
					testserver.WithEchoArgs("--tls", "--crt=/kuma/server.crt", "--key=/kuma/server.key"),
					testserver.WithName("tls-external-service"),
					testserver.WithoutProbes(), // not compatible with TLS
				)).
				Install(YamlK8s(tlsExternalService)).
				Install(YamlK8s(hostnameGenerator)).
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
