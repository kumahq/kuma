package mtls

import (
	"fmt"
	"regexp"
	"strings"

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

	createMesh := func(name, mode string) {
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

	runTestServer := func(mesh string) {
		echoServerToken, err := universal.GetKuma().GenerateDpToken(mesh, "test-server")
		Expect(err).ToNot(HaveOccurred())

		Expect(TestServerUniversal("test-server", mesh, echoServerToken,
			WithArgs([]string{"echo", "--instance", "universal-1"}))(universal)).To(Succeed())
	}

	checkInsideMeshConnection := func() {
		Eventually(func() bool {
			stdout, _, err := universal.Exec("", "", "demo-client", "curl", "-v", "-m", "8", "--fail", "test-server.mesh")
			if err != nil {
				return false
			}
			return strings.Contains(stdout, "HTTP/1.1 200 OK")
		}, "30s", "1s").Should(BeTrue())
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

	checkOutsideMeshConnection := func() {
		addr := getServiceEndpoint()

		Eventually(func() bool {
			stdout, _, err := universal.Exec("", "", "demo-client", "curl", "-v", "-m", "8", "--fail", addr)
			if err != nil {
				return false
			}
			return strings.Contains(stdout, "HTTP/1.1 200 OK")
		}, "30s", "1s").Should(BeTrue())
	}

	checkNoOutsideMeshConnection := func() {
		addr := getServiceEndpoint()

		Eventually(func() bool {
			stdout, _, err := universal.Exec("", "", "demo-client", "curl", "-v", "-m", "8", "--fail", addr)
			if err != nil {
				return false
			}
			return strings.Contains(stdout, "HTTP/1.1 200 OK")
		}, "30s", "1s").Should(BeFalse())
	}

	It("should support STRICT mTLS mode", func() {
		createMesh("default", "STRICT")

		runTestServer("default")

		runDemoClient("default")

		checkInsideMeshConnection()

		checkNoOutsideMeshConnection()
	})

	It("should support PERMISSIVE mTLS mode", func() {
		createMesh("default", "PERMISSIVE")

		runTestServer("default")

		runDemoClient("default")

		checkInsideMeshConnection()

		checkOutsideMeshConnection()
	})
}
