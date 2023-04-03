package externalservices

import (
	"fmt"
	"strings"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func ExternalServices() {
	tpName := "traffic-to-es-" + strings.ToLower(random.UniqueId())
	es1PolicyName := "external-service-1-" + strings.ToLower(random.UniqueId())
	meshName := "external-services-" + strings.ToLower(random.UniqueId())
	namespace := "external-services-" + strings.ToLower(random.UniqueId())
	clientNamespace := "client-external-services-" + strings.ToLower(random.UniqueId())

	mesh := `
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
      passthrough: %s
  routing:
    zoneEgress: true
`
	meshPassthroughEnabled := fmt.Sprintf(mesh, meshName, "true")
	meshPassthroughDisabled := fmt.Sprintf(mesh, meshName, "false")

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(YamlK8s(meshPassthroughEnabled)).
			Install(Namespace(namespace)).
			Install(NamespaceWithSidecarInjection(clientNamespace)).
			Install(democlient.Install(democlient.WithNamespace(clientNamespace), democlient.WithMesh(meshName))).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())

		err = k8s.RunKubectlE(kubernetes.Cluster.GetTesting(), kubernetes.Cluster.GetKubectlOptions(), "delete", "trafficpermission", "allow-all-"+meshName)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(clientNamespace)).To(Succeed())
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	Context("non-TLS", func() {
		externalService := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: ExternalService
mesh: %s
metadata:
  name: %s
spec:
  tags:
    kuma.io/service: external-service
    kuma.io/protocol: http
  networking:
    address: external-service.%s.svc.cluster.local:80 # .svc.cluster.local is needed, otherwise Kubernetes will resolve this to the real IP
`, meshName, es1PolicyName, namespace)

		trafficPermission := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: TrafficPermission
mesh: %s
metadata:
  name: %s
spec:
  sources:
    - match:
        kuma.io/service: '*'
  destinations:
    - match:
        kuma.io/service: external-service
`, meshName, tpName)

		BeforeAll(func() {
			err := kubernetes.Cluster.Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithName("external-service"),
			))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should route to external-service", func() {
			// given working communication outside the mesh with passthrough enabled and no traffic permission
			Eventually(func(g Gomega) {
				_, err := client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client", fmt.Sprintf("external-service.%s", namespace),
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
			}, "30s", "1s").Should(Succeed())

			// when passthrough is disabled on the Mesh
			Expect(kubernetes.Cluster.Install(YamlK8s(meshPassthroughDisabled))).To(Succeed())

			// then accessing the external service is no longer possible
			Eventually(func(g Gomega) {
				response, err := client.CollectFailure(
					kubernetes.Cluster, "demo-client", fmt.Sprintf("external-service.%s", namespace),
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.Exitcode).To(Or(Equal(56), Equal(7), Equal(28)))
			}, "30s", "1s").Should(Succeed())

			// when apply external service
			Expect(kubernetes.Cluster.Install(YamlK8s(externalService))).To(Succeed())

			// then traffic is still blocked because of lack of the traffic permission
			Eventually(func(g Gomega) {
				response, err := client.CollectFailure(
					kubernetes.Cluster, "demo-client", fmt.Sprintf("external-service.%s", namespace),
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.ResponseCode).To(Equal(503))
			}, "30s", "1s").Should(Succeed())

			// when TrafficPermission is added
			Expect(kubernetes.Cluster.Install(YamlK8s(trafficPermission))).To(Succeed())

			// then you can access external service again
			Eventually(func(g Gomega) {
				_, err := client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client", fmt.Sprintf("external-service.%s", namespace),
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
			}, "30s", "1s").Should(Succeed())

			// and you can also use .mesh on port of the provided host
			Eventually(func(g Gomega) {
				_, err := client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client", "external-service.mesh",
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
			}, "30s", "1s").Should(Succeed())
		})
	})

	Context("TLS", func() {
		tlsExternalService := fmt.Sprintf(`
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
    address: tls-external-service.%s.svc.cluster.local:80 # .svc.cluster.local is needed, otherwise Kubernetes will resolve this to the real IP
    tls:
      enabled: true
`, namespace)

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
				_, err := client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client", "tls-external-service.mesh",
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
			}, "30s", "1s").Should(Succeed())
		})
	})
}
