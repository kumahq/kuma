package gateway

import (
	"fmt"
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func Resources() {
	meshName := "gateway-resources"
	gatewayName := "resources-edge-gateway"
	namespace := "gateway-resources"
	waitingClientNamespace := "gateway-resources-client-wait"
	curlingClientNamespace := "gateway-resources-client-curl"
	gatewayHost := fmt.Sprintf("%s.%s", gatewayName, namespace)
	target := fmt.Sprintf("http://%s", net.JoinHostPort(gatewayHost, "8080"))

	meshGatewayWithoutLimit := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshGateway
metadata:
  name: %s
mesh: %s
spec:
  selectors:
  - match:
      kuma.io/service: %s_%s_svc
  conf:
    listeners:
    - port: 8080
      protocol: HTTP
`, gatewayName, meshName, gatewayName, namespace)

	meshGatewayWithLimit := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshGateway
metadata:
  name: %s
mesh: %s
spec:
  selectors:
  - match:
      kuma.io/service: %s_%s_svc
  conf:
    listeners:
    - port: 8080
      protocol: HTTP
      resources:
        connectionLimit: 1
`, gatewayName, meshName, gatewayName, namespace)

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
      kuma.io/service: %s_%s_svc
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
`, gatewayName, meshName, gatewayName, namespace, serverSvc)

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MTLSMeshKubernetes(meshName)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Namespace(waitingClientNamespace)).
			Install(Namespace(curlingClientNamespace)).
			Install(MeshTrafficPermissionAllowAllKubernetes(meshName)).
			Install(Parallel(
				democlient.Install(democlient.WithNamespace(waitingClientNamespace), democlient.WithMesh(meshName)),
				democlient.Install(democlient.WithNamespace(curlingClientNamespace), democlient.WithMesh(meshName)),
				testserver.Install(
					testserver.WithMesh(meshName),
					testserver.WithNamespace(namespace),
					testserver.WithName("test-server"),
					testserver.WithEchoArgs("echo", "--instance", "kubernetes"),
				),
			)).
			Install(YamlK8s(meshGatewayWithoutLimit)).
			Install(YamlK8s(MkGatewayInstance(gatewayName, namespace, meshName))).
			Install(YamlK8s(httpRoute)).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugKube(kubernetes.Cluster, meshName, namespace, waitingClientNamespace, curlingClientNamespace)
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(waitingClientNamespace)).To(Succeed())
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(curlingClientNamespace)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	keepConnectionOpen := func() {
		// Open TCP connections to the gateway
		defer GinkgoRecover()

		demoClientPod, err := PodNameOfApp(kubernetes.Cluster, "demo-client", waitingClientNamespace)
		Expect(err).ToNot(HaveOccurred())

		cmd := []string{"telnet", gatewayHost, "8080"}
		// We pass in a stdin that blocks so that telnet will keep the
		// connection open
		_, _, _ = kubernetes.Cluster.ExecWithOptions(ExecOptions{
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
			response, err := client.CollectEchoResponse(
				kubernetes.Cluster, "demo-client", target,
				client.FromKubernetesPod(curlingClientNamespace, "demo-client"),
			)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(Equal("kubernetes"))
		}, "30s", "1s").MustPassRepeatedly(3).Should(Succeed())

		By("allowing more than 1 connection without a limit")

		go keepConnectionOpen()

		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(
				kubernetes.Cluster, "demo-client", target,
				client.FromKubernetesPod(curlingClientNamespace, "demo-client"),
			)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(Equal("kubernetes"))
		}, "1m", "1s").MustPassRepeatedly(3).Should(Succeed())

		By("not allowing more than 1 connection with a limit of 1")

		Expect(kubernetes.Cluster.Install(YamlK8s(meshGatewayWithLimit))).To(Succeed())

		Expect(kubernetes.Cluster.KillAppPod("demo-client", waitingClientNamespace)).To(Succeed())

		go keepConnectionOpen()

		Eventually(func(g Gomega) {
			response, err := client.CollectFailure(
				kubernetes.Cluster, "demo-client", target,
				client.FromKubernetesPod(curlingClientNamespace, "demo-client"),
			)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Exitcode).To(Or(Equal(52), Equal(56)))
		}, "1m", "1s").MustPassRepeatedly(3).Should(Succeed())
	})
}
