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

	meshGatewayWithoutLimit := fmt.Sprintf(`
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
      protocol: HTTP
`, gatewayName, meshName, gatewayName)

	meshGatewayWithLimit := fmt.Sprintf(`
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
      protocol: HTTP
      resources:
        connectionLimit: 1
`, gatewayName, meshName, gatewayName)

	serverSvc := fmt.Sprintf("test-server_%s_svc_80", namespace)

	httpRoute := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshGatewayRoute
metadata:
  name: %s
mesh: %s
spec:
  selectors:
  - match:
      kuma.io/service: %s
  conf:
    http:
      rules:
      - matches:
          - path:
              match: PREFIX
              value: /
        backends:
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
			Install(YamlK8s(meshGatewayWithoutLimit)).
			Install(YamlK8s(MkGatewayInstance(gatewayName, namespace, meshName))).
			Install(YamlK8s(httpRoute)).
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

	gatewayHost := fmt.Sprintf("%s.%s", gatewayName, namespace)
	target := fmt.Sprintf("http://%s:8080", gatewayHost)

	keepConnectionOpen := func() {
		// Open TCP connections to the gateway
		defer GinkgoRecover()

		demoClientPod, err := PodNameOfApp(env.Cluster, "demo-client", waitingClientNamespace)
		Expect(err).ToNot(HaveOccurred())

		cmd := []string{"telnet", gatewayHost, "8080"}
		// We pass in a stdin that blocks so that telnet will keep the
		// connection open
		_, _, _ = env.Cluster.ExecWithOptions(ExecOptions{
			Command:            cmd,
			Namespace:          waitingClientNamespace,
			PodName:            demoClientPod,
			ContainerName:      "demo-client",
			Stdin:              &BlockingReader{},
			CaptureStdout:      true,
			CaptureStderr:      true,
			PreserveWhitespace: false,
		})
	}

	Specify("connection limit is respected", func() {
		By("allowing connections without a limit")

		Eventually(func(g Gomega) {
			response, err := client.CollectResponse(
				env.Cluster, "demo-client", target,
				client.FromKubernetesPod(curlingClientNamespace, "demo-client"),
			)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(Equal("kubernetes"))
		}, "20s", "1s")

		By("allowing more than 1 connection without a limit")

		go keepConnectionOpen()

		Eventually(func(g Gomega) {
			_, err := client.CollectResponse(
				env.Cluster, "demo-client", target,
				client.FromKubernetesPod(curlingClientNamespace, "demo-client"),
			)

			g.Expect(err).ToNot(HaveOccurred())
		}).Should(Succeed())
		Consistently(func(g Gomega) {
			response, err := client.CollectResponse(
				env.Cluster, "demo-client", target,
				client.FromKubernetesPod(curlingClientNamespace, "demo-client"),
			)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(Equal("kubernetes"))
		}, "40s", "1s").Should(Succeed())

		By("not allowing more than 1 connection with a limit of 1")

		Expect(env.Cluster.Install(YamlK8s(meshGatewayWithLimit))).To(Succeed())

		Expect(env.Cluster.KillAppPod("demo-client", waitingClientNamespace)).To(Succeed())

		go keepConnectionOpen()

		Eventually(func(g Gomega) {
			response, err := client.CollectFailure(
				env.Cluster, "demo-client", target,
				client.FromKubernetesPod(curlingClientNamespace, "demo-client"),
			)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Exitcode).To(Or(Equal(52), Equal(56)))
		}, "40s", "1s").Should(Succeed())
	})
}
