package permissive

import (
	"fmt"
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

func PermissiveMode() {
	var universal Cluster

	BeforeEach(func() {
		clusters, err := NewUniversalClusters([]string{Kuma1}, Silent)
		Expect(err).ToNot(HaveOccurred())

		universal = clusters.GetCluster(Kuma1)
		// This option is important for introducing update delays into to enable PERMISSIVE mTLS test
		Expect(Kuma(core.Standalone, WithEnv("KUMA_XDS_SERVER_DATAPLANE_CONFIGURATION_REFRESH_INTERVAL", "1s"))(universal)).To(Succeed())
	})

	E2EAfterEach(func() {
		Expect(universal.DeleteKuma()).To(Succeed())
		Expect(universal.DismissCluster()).To(Succeed())
	})

	createMeshMTLS := func(name, mode string) {
		meshYaml := fmt.Sprintf(
			`
type: Mesh
name: %s
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
    mode: %s`, name, mode)
		Expect(YamlUniversal(meshYaml)(universal)).To(Succeed())
	}

	runDemoClient := func(mesh string) {
		demoClientToken, err := universal.GetKuma().GenerateDpToken(mesh, "demo-client")
		Expect(err).ToNot(HaveOccurred())
		Expect(
			DemoClientUniversal(AppModeDemoClient, mesh, demoClientToken, WithTransparentProxy(true))(universal),
		).To(Succeed())
	}

	runTestServer := func(mesh string, tls bool) {
		echoServerToken, err := universal.GetKuma().GenerateDpToken(mesh, "test-server")
		Expect(err).ToNot(HaveOccurred())

		args := []string{"echo", "--instance", "universal-1"}
		if tls {
			args = append(args, "--tls", "--crt=/kuma/server.crt", "--key=/kuma/server.key")
		}
		Expect(TestServerUniversal("test-server", mesh, echoServerToken, WithArgs(args), WithProtocol("tcp"))(universal)).To(Succeed())
	}

	curlAddr := func(addr string, opts ...string) func() error {
		return func() error {
			cmd := []string{"curl", "-v", "-m", "3", "--fail"}
			cmd = append(cmd, opts...)
			cmd = append(cmd, addr)
			_, _, err := universal.Exec("", "", "demo-client", cmd...)
			if err != nil {
				return fmt.Errorf("request from client to %s failed", addr)
			}
			return nil
		}
	}

	clientToServer := curlAddr("test-server.mesh")
	clientToServerDirect := func() error {
		// using direct IP:PORT allows bypassing outbound listeners
		addr := net.JoinHostPort(universal.(*UniversalCluster).GetApp("test-server").GetIP(), "80")
		return curlAddr(addr)()
	}
	clientToServerTLS := curlAddr("https://test-server.mesh:80", "--cacert", "/kuma/server.crt")
	clientToServerTLSDirect := func() error {
		host := universal.(*UniversalCluster).GetApp("test-server").GetIP()
		// we're using curl with '--resolve' flag to verify certificate Common Name 'test-server.mesh'
		return curlAddr("https://test-server.mesh:80", "--cacert", "/kuma/server.crt", "--resolve", fmt.Sprintf("test-server.mesh:80:[%s]", host))()
	}

	It("should support STRICT mTLS mode", func() {
		createMeshMTLS("default", "STRICT")

		runTestServer("default", false)

		runDemoClient("default")

		// check the inside-mesh communication
		Eventually(clientToServer, "30s", "1s").Should(Succeed())

		// check the outside-mesh communication
		Eventually(clientToServerDirect, "30s", "1s").ShouldNot(Succeed())
	})

	It("should support PERMISSIVE mTLS mode", func() {
		createMeshMTLS("default", "PERMISSIVE")

		runTestServer("default", false)

		runDemoClient("default")

		// check the inside-mesh communication
		Eventually(clientToServer, "30s", "1s").Should(Succeed())

		// check the outside-mesh communication (using direct IP:PORT allows bypassing outbound listeners)
		Eventually(clientToServerDirect, "30s", "1s").Should(Succeed())
	})

	It("should support mTLS if connection already TLS", func() {
		createMeshMTLS("default", "STRICT")

		runTestServer("default", true)

		runDemoClient("default")

		Eventually(clientToServerTLS, "30s", "1s").Should(Succeed())
	})

	It("should support PERMISSIVE mTLS mode if the client is using TLS", func() {
		createMeshMTLS("default", "PERMISSIVE")

		runTestServer("default", true)

		runDemoClient("default")

		// check the inside-mesh communication with mTLS over TLS
		Eventually(clientToServerTLS, "30s", "1s").Should(Succeed())

		// check the outside-mesh communication with mTLS over TLS
		Eventually(clientToServerTLSDirect, "30s", "1s").Should(Succeed())
	})

	It("should support enabling PERMISSIVE mTLS mode with no failed requests", func() {
		// Disable retries so that we see every failed request
		kumactl := universal.GetKumactlOptions()
		Expect(kumactl.KumactlDelete("retry", "retry-all-default", "default")).To(Succeed())

		// We must start client before server to test this properly. The client
		// should get XDS refreshes first to trigger the race condition fixed by
		// kumahq/kuma#3019
		runDemoClient("default")

		runTestServer("default", false)

		Eventually(clientToServer, "30s", "1s").Should(Succeed())

		createMeshMTLS("default", "PERMISSIVE")

		Consistently(clientToServer, "20s", "100ms").Should(Succeed())
	})
}
