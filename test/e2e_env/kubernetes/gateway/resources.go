package gateway

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/kubernetes/env"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func Resources() {
	meshName := "gateway-resources"
	gatewayName := "resources-edge-gateway"
	namespace := "gateway-resources"
	waitingClientNamespace := "gateway-resources-client-wait"
	curlingClientNamespace := "gateway-resources-client-curl"

	meshGateway := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshGateway
metadata:
  name: %s
mesh: %s
spec:
  selectors:
  - match:
      kuma.io/service: %s
  conf:
    listeners:
    - port: 8080
      protocol: TCP
`, gatewayName, meshName, gatewayName)

	serverSvc := fmt.Sprintf("test-server-%s_svc_80", namespace)

	tcpRoute := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshGatewayRoute
metadata:
  name: %s
mesh: %s
spec:
  selectors:
  - match:
      kuma.io/service: %s
      protocol: http
  conf:
    tcp:
      rules:
      - backends:
        - destination:
            kuma.io/service: %s
`, gatewayName, meshName, gatewayName, serverSvc)

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MTLSMeshKubernetes(meshName)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Namespace(waitingClientNamespace)).
			Install(Namespace(curlingClientNamespace)).
			Install(DemoClientK8s(meshName, waitingClientNamespace)).
			Install(DemoClientK8s(meshName, curlingClientNamespace)).
			Install(YamlK8s(meshGateway)).
			Install(YamlK8s(MkGatewayInstance(gatewayName, namespace, meshName))).
			Install(YamlK8s(tcpRoute)).
			Install(testserver.Install(
				testserver.WithMesh(meshName),
				testserver.WithNamespace(namespace),
				testserver.WithName("test-server"),
				testserver.WithEchoArgs("echo", "--instance", "kubernetes"),
			)).
			Setup(env.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(env.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(env.Cluster.TriggerDeleteNamespace(waitingClientNamespace)).To(Succeed())
		Expect(env.Cluster.TriggerDeleteNamespace(curlingClientNamespace)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	Context("connection limit", func() {
		gatewayHost := fmt.Sprintf("%s.%s", gatewayName, namespace)
		target := fmt.Sprintf("http://%s:8080", gatewayHost)

		It("should allow 1 connection", func() {
			Eventually(func(g Gomega) {
				response, err := client.CollectResponse(
					env.Cluster, "demo-client", target,
					client.FromKubernetesPod(curlingClientNamespace, "demo-client"),
				)

				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.Instance).To(Equal("kubernetes"))
			})
		})

		It("should not allow more than 1 connection", func() {
			// Open a long-living TCP connection to the gateway
			go func() {
				defer GinkgoRecover()

				demoClientPod, err := PodNameOfApp(env.Cluster, "demo-client", waitingClientNamespace)
				Expect(err).ToNot(HaveOccurred())

				// this pod will be killed when we delete the namespace
				cmd := []string{"nc", "-w", "30", gatewayHost, "8080"}
				_, _, _ = env.Cluster.Exec(waitingClientNamespace, demoClientPod, "demo-client", cmd...)
			}()

			Eventually(func(g Gomega) {
				response, err := client.CollectFailure(
					env.Cluster, "demo-client", target,
					client.FromKubernetesPod(curlingClientNamespace, "demo-client"),
				)

				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.Exitcode).To(Equal(56))
			}, "20s", "1s").Should(Succeed())
		})
	})
}
