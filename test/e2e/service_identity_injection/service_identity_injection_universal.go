package service_identity_injection

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/server/types"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

func ServiceIdentityInjectionUniversal() {
	meshMTLSOn := func() string {
		return `
type: Mesh
name: default
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
`
	}

	var cluster Cluster
	var deployOptsFuncs []DeployOptionsFunc

	BeforeEach(func() {
		clusters, err := NewUniversalClusters(
			[]string{Kuma3},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		// Global
		cluster = clusters.GetCluster(Kuma3)
		deployOptsFuncs = KumaUniversalDeployOpts

		err = NewClusterSetup().
			Install(Kuma(core.Standalone, deployOptsFuncs...)).
			Install(YamlUniversal(meshMTLSOn())).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())

		Expect(cluster.VerifyKuma()).To(Succeed())

		demoClientToken, err := cluster.GetKuma().GenerateDpToken("default", "dp-demo-client")
		Expect(err).ToNot(HaveOccurred())
		testServerToken, err := cluster.GetKuma().GenerateDpToken("default", "test-server")
		Expect(err).ToNot(HaveOccurred())

		err = DemoClientUniversal("dp-demo-client", "default", demoClientToken,
			WithTransparentProxy(true), WithBuiltinDNS(true))(cluster)
		Expect(err).ToNot(HaveOccurred())

		err = TestServerUniversal("test-server", "default", testServerToken,
			WithArgs([]string{"echo"}),
			WithTransparentProxy(true), WithProtocol("http"))(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}
		err := cluster.DeleteKuma(deployOptsFuncs...)
		Expect(err).ToNot(HaveOccurred())

		err = cluster.DismissCluster()
		Expect(err).ToNot(HaveOccurred())
	})

	It("server should receive X-Kuma-Forwarded-Client-* headers", func() {
		var response types.EchoResponse
		cmd := []string{"curl", "--fail", "test-server.mesh/content"}
		stdout, _, err := cluster.ExecWithRetries("", "", "dp-demo-client", cmd...)
		Expect(err).ToNot(HaveOccurred())
		Expect(json.Unmarshal([]byte(stdout), &response)).To(Succeed())

		Expect(response.Received.Headers).To(HaveKeyWithValue("X-Kuma-Forwarded-Client-Cert", []string{"spiffe://default/dp-demo-client"}))
		Expect(response.Received.Headers).To(HaveKeyWithValue("X-Kuma-Forwarded-Client-Service", []string{"dp-demo-client"}))
		Expect(response.Received.Headers).To(HaveKeyWithValue("X-Kuma-Forwarded-Client-Zone", []string{"us-east-1"}))
	})
}
