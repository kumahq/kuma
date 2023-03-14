package meshtrafficpermission

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func MeshTrafficPermissionUniversal() {
	meshName := "meshtrafficpermission"

	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(MTLSMeshUniversal(meshName)).
			Install(TestServerUniversal("test-server", meshName, WithArgs([]string{"echo", "--instance", "echo-v1"}))).
			Install(DemoClientUniversal(AppModeDemoClient, meshName, WithTransparentProxy(true))).
			Setup(universal.Cluster)).To(Succeed())

		// remove default traffic permission
		err := universal.Cluster.GetKumactlOptions().KumactlDelete("traffic-permission", "allow-all-"+meshName, meshName)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	E2EAfterEach(func() {
		// remove all MeshTrafficPermissions
		items, err := universal.Cluster.GetKumactlOptions().KumactlList("meshtrafficpermissions", meshName)
		Expect(err).ToNot(HaveOccurred())
		for _, item := range items {
			err := universal.Cluster.GetKumactlOptions().KumactlDelete("meshtrafficpermission", item, meshName)
			Expect(err).ToNot(HaveOccurred())
		}
	})

	trafficAllowed := func() {
		Eventually(func(g Gomega) {
			_, err := client.CollectResponse(
				universal.Cluster, "demo-client", "test-server.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
		}).Should(Succeed())
	}

	trafficBlocked := func() {
		Eventually(func(g Gomega) {
			response, err := client.CollectFailure(
				universal.Cluster, "demo-client", "test-server.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.ResponseCode).To(Equal(503))
		}).Should(Succeed())
	}

	It("should allow the traffic with meshtrafficpermission based on MeshService", func() {
		// given no mesh traffic permissions
		trafficBlocked()

		// when mesh traffic permission with MeshService
		yaml := `
type: MeshTrafficPermission
name: mtp-1
mesh: meshtrafficpermission
spec:
 targetRef:
   kind: MeshService
   name: test-server
 from:
   - targetRef:
       kind: MeshService
       name: demo-client
     default:
       action: Allow
`
		err := YamlUniversal(yaml)(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// then
		trafficAllowed()
	})

	It("should allow the traffic with traffic permission based on non standard tag", func() {
		// given no mesh traffic permission
		trafficBlocked()

		// when
		yaml := `
type: MeshTrafficPermission
name: mtp-2
mesh: meshtrafficpermission
spec:
  targetRef:
    kind: MeshSubset
    tags:
      team: server-owners
  from:
    - targetRef:
        kind: MeshSubset
        tags: 
          team: client-owners
      default:
        action: Allow
`
		err := YamlUniversal(yaml)(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())

		trafficAllowed()
	})
}
