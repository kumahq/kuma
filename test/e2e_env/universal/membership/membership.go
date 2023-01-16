package membership

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func Membership() {
	meshName := "membership"
	mesh := `
type: Mesh
name: %s
constraints:
  dataplaneProxy:
    requirements:
    - tags:
        kuma.io/service: demo-client
    restrictions:
    - tags:
        kuma.io/service: test-server
`

	BeforeAll(func() {
		Expect(YamlUniversal(fmt.Sprintf(mesh, meshName))(universal.Cluster)).To(Succeed())
	})
	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})
	It("should take into account membership when dp is connecting to the CP", func() {
		// when demo client is trying to connect
		err := DemoClientUniversal(AppModeDemoClient, meshName)(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// then it's allowed
		Eventually(func() (string, error) {
			return universal.Cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes")
		}, "30s", "1s").ShouldNot(ContainSubstring(AppModeDemoClient))

		// when test server is trying to connect
		err = TestServerUniversal("test-server", meshName, WithArgs([]string{"echo", "--instance", "echo-v1"}))(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// then it's not allowed
		// todo(jakubdyszkiewicz) uncomment once we can handle CP logs across all parallel executions
		// Eventually(func() (string, error) {
		//	return env.Cluster.GetKumaCPLogs()
		// }, "30s", "1s").Should(ContainSubstring("dataplane cannot be a member of mesh"))
		Consistently(func() (string, error) {
			return universal.Cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes")
		}, "10s", "5s").ShouldNot(ContainSubstring("test-server"))
	})
}
