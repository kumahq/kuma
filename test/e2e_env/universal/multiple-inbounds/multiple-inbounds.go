package multiple_inbounds

import (
	"github.com/kumahq/kuma/test/e2e_env/universal/env"
	. "github.com/kumahq/kuma/test/framework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func Test() {
	meshName := "multiple-inbounds"
	grpcPort := "8889"
	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshUniversal(meshName)).
			Install(DemoClientUniversal("http-client", meshName, WithTransparentProxy(true))).
			// both http and grpc server
			Install(TestServerUniversal("test-server", meshName,
				WithArgs([]string{"echo", "--instance", "universal-1", "--port", "8080"}),
				WithSecondaryApp(
					grpcPort,
					grpcPort,
					"1",
					"",
					"v1",
					"test-server-grpc",
					"grpc",
					[]string{"test-server", "grpc", "server", "--port", grpcPort},
				),
			)).
			Setup(env.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// Delete the default retry policy
		Eventually(func() error {
			return env.Cluster.GetKumactlOptions().RunKumactl("delete", "retry", "--mesh", meshName, "retry-all-"+meshName)
		}).Should(Succeed())
	})

	E2EAfterAll(func() {
		//Expect(env.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		//Expect(env.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should not mix up inbounds on upstream restart", func() {
		By("Checking initially requests succeed")
		Eventually(func(g Gomega) {
			stdout, _, err := env.Cluster.Exec("", "", "http-client",
				"curl", "-v", "-m", "3", "--fail", "test-server.mesh")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
		}).Should(Succeed())

		Eventually(func(g Gomega) {
			stdout, _, err := env.Cluster.Exec("", "", "http-client",
				"grpcurl", "-plaintext", "test-server-grpc.mesh:80", "test.server.grpc.apo.Greeter/SayHello")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("Hello  from"))
		}).Should(Succeed())
	})
}
