package auth

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/universal/env"
	. "github.com/kumahq/kuma/test/framework"
)

func DpAuth() {
	const meshName = "dp-auth"

	BeforeAll(func() {
		Expect(env.Cluster.Install(MeshUniversal(meshName))).To(Succeed())
		E2EDeferCleanup(func() {
			Expect(env.Cluster.DeleteMeshApps(meshName)).To(Succeed())
			Expect(env.Cluster.DeleteMesh(meshName)).To(Succeed())
		})
	})

	It("should not be able to override someone else Dataplane", func() {
		// given other dataplane
		dp := fmt.Sprintf(`
type: Dataplane
mesh: %s
name: dp-01
networking:
  address: 192.168.0.1
  inbound:
  - port: 8080
    tags:
      kuma.io/service: not-test-server
`, meshName)
		Expect(env.Cluster.Install(YamlUniversal(dp))).To(Succeed())

		// when trying to spin up dataplane with same name but token bound to a different service
		err := TestServerUniversal("dp-01", meshName, WithServiceName("test-server"))(env.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// then
		// todo(jakubdyszkiewicz) uncomment once we can handle CP logs across all parallel executions
		// Eventually(func() (string, error) {
		//	return env.Cluster.GetKumaCPLogs()
		// }, "30s", "1s").Should(ContainSubstring("you are trying to override existing dataplane to which you don't have an access"))
	})

	It("should be able to override old Dataplane of same service", func() {
		// given
		dp := fmt.Sprintf(`
type: Dataplane
mesh: %s
name: dp-02
networking:
  address: 192.168.0.2
  inbound:
  - port: 8080
    tags:
      kuma.io/service: test-server
`, meshName)
		Expect(env.Cluster.Install(YamlUniversal(dp))).To(Succeed())

		// when
		err := TestServerUniversal("dp-02", meshName, WithServiceName("test-server"))(env.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// then
		Eventually(func() (string, error) {
			return env.Cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes", "-oyaml")
		}, "30s", "1s").ShouldNot(ContainSubstring("192.168.0.2"))
	})
}
