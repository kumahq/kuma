package auth

import (
	"encoding/base64"
	"fmt"
	"math/rand"

	"github.com/golang-jwt/jwt/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/universal/env"
	. "github.com/kumahq/kuma/test/framework"
)

func DpAuth() {
	const meshName = "dp-auth"

	BeforeAll(func() {
		Expect(env.Cluster.Install(MeshUniversal(meshName))).To(Succeed())
	})
	E2EAfterAll(func() {
		Expect(env.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(meshName)).To(Succeed())
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

	It("should revoke token and kick out dataplane proxy out of the mesh", func() {
		// given
		serviceName := "test-server-to-be-revoked"
		token, err := env.Cluster.GetKuma().GenerateDpToken(meshName, serviceName)
		Expect(err).ToNot(HaveOccurred())

		err = env.Cluster.Install(TestServerUniversal(serviceName, meshName, WithServiceName(serviceName), WithToken(token)))
		Expect(err).ToNot(HaveOccurred())

		Eventually(func(g Gomega) {
			online, found, err := IsDataplaneOnline(env.Cluster, meshName, serviceName)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(found).To(BeTrue())
			g.Expect(online).To(BeTrue())
		}).Should(Succeed())

		// when token ID is added to revocation list
		claims := &jwt.RegisteredClaims{}
		_, _, err = jwt.NewParser().ParseUnverified(token, claims)
		Expect(err).ToNot(HaveOccurred())

		yaml := fmt.Sprintf(`
type: Secret
mesh: dp-auth
name: dataplane-token-revocations-dp-auth
data: %s`, base64.StdEncoding.EncodeToString([]byte(claims.ID)))
		Expect(env.Cluster.Install(YamlUniversal(yaml))).To(Succeed())

		// then DPP is disconnected
		Eventually(func(g Gomega) {
			// we need to trigger XDS config change for this DP to disconnect it
			// this limitation may be lifted in the future
			yaml = fmt.Sprintf(`
type: Retry
name: retry-policy
mesh: dp-auth
sources:
- match:
    kuma.io/service: test-server-to-be-revoked
destinations:
- match:
    kuma.io/service: test-server-to-be-revoked
conf:
  http:
    numRetries: %d
`, rand.Int()%100+1)
			g.Expect(env.Cluster.Install(YamlUniversal(yaml))).To(Succeed())

			online, _, err := IsDataplaneOnline(env.Cluster, meshName, serviceName)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(online).To(BeFalse()) // either online or not found
		}).Should(Succeed())
	})
}
