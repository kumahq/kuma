package mtls

import (
	"fmt"
	"net"
	"regexp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

func StandaloneUniversal() {
	var universal Cluster
	var universalOpts = KumaUniversalDeployOpts

	BeforeEach(func() {
		clusters, err := NewUniversalClusters([]string{Kuma1}, Silent)
		Expect(err).ToNot(HaveOccurred())

		universal = clusters.GetCluster(Kuma1)
		Expect(Kuma(core.Standalone, universalOpts...)(universal)).To(Succeed())
		Expect(universal.VerifyKuma()).To(Succeed())
	})

	E2EAfterEach(func() {
		Expect(universal.DeleteKuma(universalOpts...)).To(Succeed())
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

	getServiceEndpoint := func() string {
		var addr string
		r := regexp.MustCompile(`::([0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}):?([0-9]{1,5})?::`)
		Eventually(func() bool {
			cmd := []string{"/bin/bash", "-c", "\"curl -s localhost:30001/clusters | grep test-server.*cx_active\""}
			stdout, _, err := universal.Exec("", "", "demo-client", cmd...)
			if err != nil {
				return false
			}
			submatch := r.FindStringSubmatch(stdout)
			if len(submatch) < 3 {
				return false
			}
			addr = fmt.Sprintf("%s:%s", submatch[1], submatch[2])
			return true
		}, "30s", "1s").Should(BeTrue())
		return addr
	}

	It("should support STRICT mTLS mode", func() {
		createMeshMTLS("default", "STRICT")

		runTestServer("default", false)

		runDemoClient("default")

		// check the inside-mesh communication
		Eventually(func() error {
			_, _, err := universal.Exec("", "", "demo-client", "curl", "-v", "-m", "3", "--fail", "test-server.mesh")
			return err
		}, "30s", "1s").ShouldNot(HaveOccurred())

		// check the outside-mesh communication (using direct IP:PORT allows bypassing outbound listeners)
		addr := getServiceEndpoint()
		Eventually(func() error {
			_, _, err := universal.Exec("", "", "demo-client", "curl", "-v", "-m", "3", "--fail", addr)
			return err
		}, "30s", "1s").ShouldNot(Succeed())
	})

	It("should support PERMISSIVE mTLS mode", func() {
		createMeshMTLS("default", "PERMISSIVE")

		runTestServer("default", false)

		runDemoClient("default")

		// check the inside-mesh communication
		Eventually(func() error {
			_, _, err := universal.Exec("", "", "demo-client", "curl", "-v", "-m", "3", "--fail", "test-server.mesh")
			return err
		}, "30s", "1s").ShouldNot(HaveOccurred())

		// check the outside-mesh communication (using direct IP:PORT allows bypassing outbound listeners)
		addr := getServiceEndpoint()
		Eventually(func() error {
			_, _, err := universal.Exec("", "", "demo-client", "curl", "-v", "-m", "3", "--fail", addr)
			return err
		}, "30s", "1s").ShouldNot(HaveOccurred())
	})

	It("should support mTLS if connection already TLS", func() {
		createMeshMTLS("default", "STRICT")

		runTestServer("default", true)

		runDemoClient("default")

		Eventually(func() error {
			cmd := []string{"curl", "-v", "-m", "3", "--fail", "--cacert", "/kuma/server.crt", "https://test-server.mesh:80"}
			_, _, err := universal.Exec("", "", "demo-client", cmd...)
			return err
		}, "30s", "1s").ShouldNot(HaveOccurred())
	})

	It("should support PERMISSIVE mTLS mode if the client is using TLS", func() {
		createMeshMTLS("default", "PERMISSIVE")

		runTestServer("default", true)

		runDemoClient("default")

		// check the inside-mesh communication with mTLS over TLS
		Eventually(func() error {
			cmd := []string{"curl", "-v", "-m", "3", "--fail", "--cacert", "/kuma/server.crt", "https://test-server.mesh:80"}
			_, _, err := universal.Exec("", "", "demo-client", cmd...)
			return err
		}, "30s", "1s").ShouldNot(HaveOccurred())

		// check the outside-mesh communication with mTLS over TLS
		// we're using curl with '--resolve' flag to verify certificate Common Name 'test-server.mesh'
		host, _, _ := net.SplitHostPort(getServiceEndpoint())
		Eventually(func() error {
			cmd := []string{"curl", "-v", "-m", "3", "--resolve", fmt.Sprintf("test-server.mesh:80:%s", host), "--fail", "--cacert", "/kuma/server.crt", "https://test-server.mesh:80"}
			_, _, err := universal.Exec("", "", "demo-client", cmd...)
			return err
		}, "30s", "1s").ShouldNot(HaveOccurred())
	})
}
