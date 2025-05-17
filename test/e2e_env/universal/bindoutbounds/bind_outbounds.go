package bindoutbounds

import (
	"regexp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
)

var msIPRegex = regexp.MustCompile(`ip: (.*)`)

func BindToLoopbackAddresses() {
	var universal Cluster
	mesh := "bind-outbounds"
	meshMs := "bind-outbounds-ms"
	cidr := "127.1.0.0/16"
	cidrMs := "127.2.0.0/16"
	cidrMes := "127.3.0.0/16"
	cidrMmzs := "127.4.0.0/16"

	BeforeEach(func() {
		universal = NewUniversalCluster(NewTestingT(), "kuma-bind-outbounds", Silent)
		Expect(NewClusterSetup().
			Install(Kuma(core.Zone,
				WithEnv("KUMA_DNS_SERVER_CIDR", cidr),
				WithEnv("KUMA_IPAM_MESH_SERVICE_CIDR", cidrMs),
				WithEnv("KUMA_IPAM_MESH_EXTERNAL_SERVICE_CIDR", cidrMes),
				WithEnv("KUMA_IPAM_MESH_MULTI_ZONE_SERVICE_CIDR", cidrMmzs),
			)).
			Install(MeshUniversal(mesh)).
			Install(ResourceUniversal(samples.MeshDefaultBuilder().WithName(meshMs).WithMeshServicesEnabled(v1alpha1.Mesh_MeshServices_Exclusive).Build())).
			Install(DemoClientUniversal("demo-client", mesh, WithBindOutbounds())).
			Install(DemoClientUniversal("demo-client-ms", meshMs, WithBindOutbounds())).
			Install(TestServerUniversal("test-server-ms", meshMs, WithArgs([]string{"echo", "--instance", "test-server-ms"}))).
			Setup(universal)).To(Succeed())

		Expect(NewClusterSetup().
			// deploy later so we allocate 127.1.0.1 for test-server
			Install(TestServerUniversal("test-server", mesh, WithArgs([]string{"echo", "--instance", "test-server"}))).
			Setup(universal)).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal, mesh)
		DebugUniversal(universal, meshMs)
	})

	E2EAfterEach(func() {
		Expect(universal.DismissCluster()).To(Succeed())
	})

	It("should send request through real bound listener", func() {
		// check there is no iptables
		Eventually(func(g Gomega) {
			response, err := client.CollectFailure(universal, "demo-client", "test-server.mesh")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Exitcode).To(Or(Equal(6), Equal(22)))
		}, "30s", "1s").Should(Succeed())

		// then when resolve return correct address
		Eventually(func(g Gomega) {
			stdout, _, err := client.CollectResponse(universal, "demo-client", "test-server.mesh", client.Resolve("test-server.mesh:80", "127.1.0.1"))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("test-server"))
		}, "30s", "1s").Should(Succeed())
	})

	It("should send request through real bound listener for MeshService", func() {
		testServerIp := ""
		// check there is no iptables
		Eventually(func(g Gomega) {
			response, err := client.CollectFailure(universal, "demo-client-ms", "test-server.svc.mesh.local")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Exitcode).To(Or(Equal(6), Equal(22)))
		}, "30s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			stdout, _, err := universal.Exec("", "", "kuma-cp", "kumactl", "get", "meshservice", "test-server", "-m", meshMs, "-oyaml")
			g.Expect(err).ToNot(HaveOccurred())
			value := msIPRegex.FindStringSubmatch(stdout)

			if len(value) == 2 {
				testServerIp = value[1]
			}
			g.Expect(value).To(HaveLen(2))
		}, "30s", "1s").Should(Succeed())

		// then when resolve return correct address
		Eventually(func(g Gomega) {
			stdout, _, err := client.CollectResponse(universal, "demo-client-ms", "test-server.svc.mesh.local", client.Resolve("test-server.svc.mesh.local:80", testServerIp))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("test-server"))
		}, "30s", "1s").Should(Succeed())
	})
}
