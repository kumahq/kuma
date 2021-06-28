package service_identity_injection

import (
	"encoding/json"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/server/types"
)

func ServiceIdentityInjectionUniversal() {
	mtlsEnabled := `
type: Mesh
name: default
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
`
	mesh := "default"
	demoClientAppName := "dp-demo-client"
	demoClientAppZone := "eu-west-1"

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
			Install(YamlUniversal(mtlsEnabled)).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())

		Expect(cluster.VerifyKuma()).To(Succeed())

		demoClientToken, err := cluster.GetKuma().GenerateDpToken(mesh, demoClientAppName)
		Expect(err).ToNot(HaveOccurred())
		testServerToken, err := cluster.GetKuma().GenerateDpToken(mesh, "test-server")
		Expect(err).ToNot(HaveOccurred())

		demoClientDataplaneYaml := fmt.Sprintf(`
type: Dataplane
mesh: %s
name: {{ name }}
networking:
  address: {{ address }}
  inbound:
  - port: 3000
    tags:
      kuma.io/service: %s
      kuma.io/zone: %s
      team: client-owners
  transparentProxying:
    redirectPortInbound: %s
    redirectPortInboundV6: %s
    redirectPortOutbound: %s
`,
			mesh, demoClientAppName, demoClientAppZone, RedirectPortInbound, RedirectPortInboundV6, RedirectPortOutbound)
		err = DemoClientUniversalYaml(demoClientAppName, mesh, demoClientToken, demoClientDataplaneYaml,
			WithTransparentProxy(true), WithBuiltinDNS(true))(cluster)
		Expect(err).ToNot(HaveOccurred())

		err = TestServerUniversal("test-server", mesh, testServerToken,
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
		stdout, _, err := cluster.ExecWithRetries("", "", demoClientAppName, cmd...)
		Expect(err).ToNot(HaveOccurred())
		Expect(json.Unmarshal([]byte(stdout), &response)).To(Succeed())

		Expect(response.Received.Headers).To(HaveKeyWithValue("X-Kuma-Forwarded-Client-Cert", []string{fmt.Sprintf("spiffe://%s/%s", mesh, demoClientAppName)}))
		Expect(response.Received.Headers).To(HaveKeyWithValue("X-Kuma-Forwarded-Client-Service", []string{demoClientAppName}))
		Expect(response.Received.Headers).To(HaveKeyWithValue("X-Kuma-Forwarded-Client-Zone", []string{demoClientAppZone}))
	})
}
