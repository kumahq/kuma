package matching

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

func Universal() {
	var universal Cluster

	BeforeEach(func() {
		clusters, err := NewUniversalClusters([]string{Kuma1}, Silent)
		Expect(err).ToNot(HaveOccurred())

		universal = clusters.GetCluster(Kuma1)

		err = NewClusterSetup().
			Install(Kuma(core.Standalone)).
			Setup(universal)
		Expect(err).ToNot(HaveOccurred())

		demoClientToken1, err := universal.GetKuma().GenerateDpToken("default", "demo-client-1")
		Expect(err).ToNot(HaveOccurred())
		Expect(
			DemoClientUniversal("demo-client-1", "default", demoClientToken1, WithTransparentProxy(true))(universal),
		).To(Succeed())
		demoClientToken2, err := universal.GetKuma().GenerateDpToken("default", "demo-client-2")
		Expect(err).ToNot(HaveOccurred())
		Expect(
			DemoClientUniversal("demo-client-2", "default", demoClientToken2, WithTransparentProxy(true))(universal),
		).To(Succeed())

		testServerToken, err := universal.GetKuma().GenerateDpToken("default", "test-server")
		Expect(err).ToNot(HaveOccurred())
		Expect(
			TestServerUniversal("test-server", "default", testServerToken,
				WithArgs([]string{"echo", "--instance", "echo-v1"}))(universal),
		).To(Succeed())
	})

	E2EAfterEach(func() {
		Expect(universal.DeleteKuma()).To(Succeed())
		Expect(universal.DismissCluster()).To(Succeed())
	})

	It("should both fault injections with the same destination proxy", FlakeAttempts(3), func() {
		Expect(YamlUniversal(`
type: FaultInjection
mesh: default
name: fi1
sources:
   - match:
       kuma.io/service: demo-client-1
destinations:
   - match:
       kuma.io/service: test-server
       kuma.io/protocol: http
conf:
   abort:
     httpStatus: 401
     percentage: 100`)(universal)).To(Succeed())

		Expect(YamlUniversal(`
type: FaultInjection
mesh: default
name: fi2
sources:
   - match:
       kuma.io/service: demo-client-2
destinations:
   - match:
       kuma.io/service: test-server
       kuma.io/protocol: http
conf:
   abort:
     httpStatus: 402
     percentage: 100`)(universal)).To(Succeed())

		Eventually(func() bool {
			stdout, _, err := universal.Exec("", "", "demo-client-1", "curl", "-v", "test-server.mesh")
			if err != nil {
				return false
			}
			return strings.Contains(stdout, "HTTP/1.1 401 Unauthorized")
		}, "60s", "1s").Should(BeTrue())

		Eventually(func() bool {
			stdout, _, err := universal.Exec("", "", "demo-client-2", "curl", "-v", "test-server.mesh")
			if err != nil {
				return false
			}
			return strings.Contains(stdout, "HTTP/1.1 402 Payment Required")
		}, "60s", "1s").Should(BeTrue())
	})
}
