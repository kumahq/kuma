package auth

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
	. "github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/envs/universal"
)

func DpAuth() {
	const meshName = "dp-auth"

	BeforeAll(func() {
		Expect(universal.Cluster.Install(MeshUniversal(meshName))).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, meshName)
	})

	AfterEach(func() {
		// Clean up dataplanes between tests
		dataplanes, err := universal.Cluster.GetKumactlOptions().KumactlList("dataplanes", meshName)
		Expect(err).ToNot(HaveOccurred())
		// Best effort cleanup - ignore errors for dataplanes that may not exist or already deleted
		for _, dp := range dataplanes {
			_ = universal.Cluster.GetKumactlOptions().KumactlDelete("dataplane", dp, meshName)
		}
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
	})

	E2EAfterAll(func() {
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

	It("should allow dataplane with matching workload label", func() {
		// given dataplane with workload label
		dp := builders.Dataplane().
			WithName("dp-workload-match").
			WithMesh(meshName).
			WithLabels(map[string]string{"kuma.io/workload": "backend"}).
			WithAddress("192.168.0.3").
			WithServices("backend-service").
			Build()
		Expect(universal.Cluster.Install(ResourceUniversal(dp))).To(Succeed())

		// when generate token bound to workload
		token, err := universal.Cluster.GetKumactlOptions().RunKumactlAndGetOutput("generate", "dataplane-token",
			"--mesh", meshName,
			"--name", "dp-workload-match",
			"--workload", "backend",
			"--valid-for", "24h")
		Expect(err).ToNot(HaveOccurred())

		// then dataplane should be able to connect with the token
		Eventually(func() error {
			return universal.Cluster.DeployApp(
				WithMesh(meshName),
				WithName("dp-workload-match"),
				WithToken(token),
				WithArgs([]string{"echo", "--port", "8080"}),
			)
		}).Should(Succeed())
	})

	It("should deny dataplane with non-matching workload label", func() {
		// given dataplane with different workload label
		dp := builders.Dataplane().
			WithName("dp-workload-mismatch").
			WithMesh(meshName).
			WithLabels(map[string]string{"kuma.io/workload": "frontend"}).
			WithAddress("192.168.0.4").
			WithServices("frontend-service").
			Build()
		Expect(universal.Cluster.Install(ResourceUniversal(dp))).To(Succeed())

		// when generate token bound to different workload
		token, err := universal.Cluster.GetKumactlOptions().RunKumactlAndGetOutput("generate", "dataplane-token",
			"--mesh", meshName,
			"--name", "dp-workload-mismatch",
			"--workload", "backend",
			"--valid-for", "24h")
		Expect(err).ToNot(HaveOccurred())

		// then dataplane should NOT be able to connect with the token
		err = universal.Cluster.DeployApp(
			WithMesh(meshName),
			WithName("dp-workload-mismatch"),
			WithToken(token),
			WithArgs([]string{"echo", "--port", "8080"}),
		)
		Expect(err).ToNot(HaveOccurred())

		// Verify the dataplane never comes online due to auth failure
		Consistently(func() bool {
			online, found, err := IsDataplaneOnline(universal.Cluster, meshName, "dp-workload-mismatch")
			Expect(err).ToNot(HaveOccurred())
			return found && online
		}).Should(BeFalse(), "dataplane should never come online due to authentication failure")
	})
}
