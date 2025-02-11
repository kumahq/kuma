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
			WithName("dp-01").
			WithMesh(meshName).
			WithAddress("192.168.0.2").
			WithServices("test-server").
			Build()
		Expect(universal.Cluster.Install(ResourceUniversal(dp))).To(Succeed())

		// when
		err := TestServerUniversal("dp-02", meshName, WithServiceName("test-server"))(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// then
		Eventually(func() (string, error) {
			return universal.Cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes", "-oyaml")
		}, "30s", "1s").ShouldNot(ContainSubstring("192.168.0.2"))
	})
<<<<<<< HEAD

	It("should revoke token and kick out dataplane proxy out of the mesh", func() {
		// given
		serviceName := "test-server-to-be-revoked"
		token, err := universal.Cluster.GetKuma().GenerateDpToken(meshName, serviceName)
		Expect(err).ToNot(HaveOccurred())

		err = universal.Cluster.Install(TestServerUniversal(serviceName, meshName, WithServiceName(serviceName), WithToken(token)))
		Expect(err).ToNot(HaveOccurred())

		Eventually(func(g Gomega) {
			online, found, err := IsDataplaneOnline(universal.Cluster, meshName, serviceName)
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
		Expect(universal.Cluster.Install(YamlUniversal(yaml))).To(Succeed())

		// then DPP is disconnected
		Eventually(func(g Gomega) {
			// we need to trigger XDS config change for this DP to disconnect it
			// this limitation may be lifted in the future
			yaml = fmt.Sprintf(`
type: MeshRetry
name: retry-policy
mesh: dp-auth
spec:
  targetRef:
    kind: MeshService
    name: test-server-to-be-revoked
  to:
    - targetRef:
        kind: MeshService
        name: test-server-to-be-revoked
      default:
        http:
          numRetries: %d
          backOff:
            baseInterval: 15s
            maxInterval: 20m
          retryOn:
            - "5xx"
`, rand.Int()%100+1) // #nosec G404 -- this is for tests no need to use secure rand
			g.Expect(universal.Cluster.Install(YamlUniversal(yaml))).To(Succeed())

			online, _, err := IsDataplaneOnline(universal.Cluster, meshName, serviceName)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(online).To(BeFalse()) // either online or not found
		}).Should(Succeed())
	})
=======
>>>>>>> 0323e80f4 (fix(xds): only auth once per xds gRPC stream in kuma-cp (#12788))
}
