package auth

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/test/resources/builders"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func DpAuth() {
	const meshName = "dp-auth"

	BeforeAll(func() {
		Expect(universal.Cluster.Install(MeshUniversal(meshName))).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, meshName)
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should not be able to override someone else Dataplane", func() {
		// given other dataplane
		dp := builders.Dataplane().
			WithName("dp-01").
			WithMesh(meshName).
			WithAddress("192.168.0.1").
			WithServices("not-test-server").
			Build()
		Expect(universal.Cluster.Install(ResourceUniversal(dp))).To(Succeed())

		// when trying to spin up dataplane with same name but token bound to a different service
		err := TestServerUniversal("dp-01", meshName, WithServiceName("test-server"))(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// then
		// todo(jakubdyszkiewicz) uncomment once we can handle CP logs across all parallel executions
		// Eventually(func() (string, error) {
		//	return env.Cluster.GetKumaCPLogs()
		// }, "30s", "1s").Should(ContainSubstring("you are trying to override existing dataplane to which you don't have an access"))
	})

	It("should be able to override old Dataplane of same service", func() {
		// given
		dp := builders.Dataplane().
			WithName("dp-02").
			WithMesh(meshName).
			WithLabels(map[string]string{"app": "not-test-server"}).
			WithAddress("192.168.0.2").
			WithServices("test-server").
			Build()
		Expect(universal.Cluster.Install(ResourceUniversal(dp))).To(Succeed())

		// when
		err := TestServerUniversal("dp-02", meshName, WithServiceName("test-server"), WithAppLabel("test-server"))(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// then
		Eventually(func() (string, error) {
			return universal.Cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes", "--mesh", meshName, "-oyaml")
		}, "30s", "1s").Should(
			And(
				Not(ContainSubstring("192.168.0.2")),
				Not(ContainSubstring("not-test-server")),
				ContainSubstring("test-server"),
			),
		)
	})
}
