package externalservices

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func ExternalServices() {
	meshName := "external-services"
	namespace := "external-services"
	clientNamespace := "client-external-services"

	var clientPodName string

	mesh := `
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: external-services
spec:
  mtls:
    enabledBackend: ca-1
    backends:
      - name: ca-1
        type: builtin
  networking:
    outbound:
      passthrough: %s
  routing:
    zoneEgress: true
`
	meshPassthroughEnabled := fmt.Sprintf(mesh, "true")
	meshPassthroughDisabled := fmt.Sprintf(mesh, "false")

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(YamlK8s(meshPassthroughEnabled)).
			Install(Namespace(namespace)).
			Install(NamespaceWithSidecarInjection(clientNamespace)).
			Install(democlient.Install(democlient.WithNamespace(clientNamespace), democlient.WithMesh(meshName))).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())

		clientPodName, err = PodNameOfApp(kubernetes.Cluster, "demo-client", clientNamespace)
		Expect(err).ToNot(HaveOccurred())

		err = k8s.RunKubectlE(kubernetes.Cluster.GetTesting(), kubernetes.Cluster.GetKubectlOptions(), "delete", "trafficpermission", "allow-all-external-services")
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(clientNamespace)).To(Succeed())
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	Context("non-TLS", func() {
		externalService := `
apiVersion: kuma.io/v1alpha1
kind: ExternalService
mesh: external-services
metadata:
  name: external-service-1
spec:
  tags:
    kuma.io/service: external-service
    kuma.io/protocol: http
  networking:
    address: external-service.external-services.svc.cluster.local:80 # .svc.cluster.local is needed, otherwise Kubernetes will resolve this to the real IP
`

		trafficPermission := `
apiVersion: kuma.io/v1alpha1
kind: TrafficPermission
mesh: external-services
metadata:
  name: traffic-to-es
spec:
  sources:
    - match:
        kuma.io/service: '*'
  destinations:
    - match:
        kuma.io/service: external-service
`
		trafficBlocked := func() error {
			_, _, err := kubernetes.Cluster.Exec(clientNamespace, clientPodName, "demo-client",
				"curl", "-v", "-m", "3", "--fail", "http://external-service.external-services:80")
			return err
		}

		BeforeAll(func() {
			err := kubernetes.Cluster.Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithName("external-service"),
			))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should route to external-service", func() {
			// given working communication outside of the mesh with passthrough enabled and no traffic permission
			Eventually(func(g Gomega) {
				_, stderr, err := kubernetes.Cluster.Exec(clientNamespace, clientPodName, "demo-client",
					"curl", "-v", "-m", "3", "--fail", "http://external-service.external-services:80")
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))
			}).Should(Succeed())

			// when passthrough is disabled on the Mesh
			Expect(kubernetes.Cluster.Install(YamlK8s(meshPassthroughDisabled))).To(Succeed())

			// then accessing the external service is no longer possible
			Eventually(trafficBlocked, "30s", "1s").Should(HaveOccurred())

			// when apply external service
			Expect(kubernetes.Cluster.Install(YamlK8s(externalService))).To(Succeed())

			// then traffic is still blocked because of lack of the traffic permission
			Consistently(trafficBlocked, "5s", "1s").Should(HaveOccurred())

			// when TrafficPermission is added
			Expect(kubernetes.Cluster.Install(YamlK8s(trafficPermission))).To(Succeed())

			// then you can access external service again
			Eventually(func(g Gomega) {
				_, stderr, err := kubernetes.Cluster.Exec(clientNamespace, clientPodName, "demo-client",
					"curl", "-v", "-m", "3", "--fail", "http://external-service.external-services:80")
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))
			}).Should(Succeed())

			// and you can also use .mesh on port of the provided host
			Eventually(func(g Gomega) {
				_, stderr, err := kubernetes.Cluster.Exec(clientNamespace, clientPodName, "demo-client",
					"curl", "-v", "-m", "3", "--fail", "http://external-service.mesh:80")
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))
			}).Should(Succeed())
		})
	})

	Context("TLS", func() {
		tlsExternalService := `
apiVersion: kuma.io/v1alpha1
kind: ExternalService
mesh: external-services
metadata:
  name: tls-external-service
spec:
  tags:
    kuma.io/service: tls-external-service
    kuma.io/protocol: http
  networking:
    address: tls-external-service.external-services.svc.cluster.local:80 # .svc.cluster.local is needed, otherwise Kubernetes will resolve this to the real IP
    tls:
      enabled: true
`

		tlsTrafficPermission := `
apiVersion: kuma.io/v1alpha1
kind: TrafficPermission
mesh: external-services
metadata:
  name: traffic-to-es
spec:
  sources:
    - match:
        kuma.io/service: '*'
  destinations:
    - match:
        kuma.io/service: tls-external-service
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
				Install(YamlK8s(tlsTrafficPermission)).
				Setup(kubernetes.Cluster)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should access tls external service", func() {
			Eventually(func(g Gomega) {
				_, stderr, err := kubernetes.Cluster.Exec(clientNamespace, clientPodName, "demo-client",
					"curl", "-v", "-m", "3", "--fail", "http://tls-external-service.mesh:80")
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))
			}).Should(Succeed())
		})
	})
}
