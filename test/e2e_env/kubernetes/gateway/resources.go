package gateway

import (
	"fmt"
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/client"
	"github.com/kumahq/kuma/v2/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/v2/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/v2/test/framework/envoy_admin/stats"
	"github.com/kumahq/kuma/v2/test/framework/envs/kubernetes"
	"github.com/kumahq/kuma/v2/test/framework/portforward"
)

// httpKeepAliveReader sends a complete HTTP/1.1 GET request on first reads,
// then blocks forever. This keeps the TCP connection alive as HTTP keep-alive
// without triggering the gateway's request_headers_timeout (which would drop
// a raw telnet/TCP connection that never sends HTTP headers).
type httpKeepAliveReader struct {
	headers []byte
	pos     int
}

func newHTTPKeepAliveReader(host string) *httpKeepAliveReader {
	h := fmt.Sprintf("GET / HTTP/1.1\r\nHost: %s\r\nConnection: keep-alive\r\n\r\n", host)
	return &httpKeepAliveReader{headers: []byte(h)}
}

func (r *httpKeepAliveReader) Read(p []byte) (int, error) {
	if r.pos < len(r.headers) {
		n := copy(p, r.headers[r.pos:])
		r.pos += n
		return n, nil
	}
	select {} // block forever after headers are sent
}

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
		GinkgoHelper()
		// Open TCP connections to the gateway
		defer GinkgoRecover()

		// Use Eventually to handle a transient state where the old (terminating) pod
		// and the new pod both appear in list operations, causing PodNameOfApp to
		// fail with "expected 1 pods, got 2". This is especially relevant after
		// KillAppPod, which may return before the old pod is fully removed.
		var demoClientPod string
		Eventually(func(g Gomega) {
			pod, err := PodNameOfApp(kubernetes.Cluster, "demo-client", waitingClientNamespace)
			g.Expect(err).ToNot(HaveOccurred())
			demoClientPod = pod
		}, "30s", "1s").Should(Succeed())

		// Use ncat with a proper HTTP/1.1 request so the gateway's
		// request_headers_timeout (500ms) doesn't close the connection.
		// After the server responds the TCP connection stays in HTTP
		// keep-alive mode (idle_timeout: 300s), keeping the slot occupied.
		cmd := []string{"ncat", gatewayHost, "8080"}
		_, _, _ = kubernetes.Cluster.ExecWithOptions(ExecOptions{
			Command:            cmd,
			Namespace:          waitingClientNamespace,
			PodName:            demoClientPod,
			ContainerName:      "demo-client",
			Stdin:              newHTTPKeepAliveReader(gatewayHost),
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

		// Wait until the holder connection is actually established on the gateway
		// before asserting that new requests are rejected. Without this sync point
		// the verifier races the holder: if ncat hasn't exec'd yet, the gateway
		// sees 0 active connections, curl succeeds, and the assertion fails.
		gatewayAdmin, err := kubernetes.Cluster.GetOrCreateAdminTunnel(portforward.Spec{
			AppName:   gatewayName,
			Namespace: namespace,
		})
		Expect(err).ToNot(HaveOccurred())
		Eventually(func(g Gomega) {
			s, err := gatewayAdmin.GetStats(
				fmt.Sprintf("http.%s_%s_svc.downstream_cx_active", gatewayName, namespace),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(s).To(stats.BeGreaterThanZero())
		}, "1m", "1s").Should(Succeed())

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
