package meshtrafficpermission

import (
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/test/resources/samples"
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
			Install(TestServerUniversal("test-server-tcp", meshName, WithArgs([]string{"echo", "--instance", "test-server-tcp"}), WithServiceName("test-server-tcp"), WithProtocol("tcp"))).
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

	trafficAllowed := func(addr string) {
		GinkgoHelper()

		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", addr,
			)
			g.Expect(err).ToNot(HaveOccurred())
		}).Should(Succeed())
	}

	trafficBlocked := func(addr string, statusCode int) {
		GinkgoHelper()

		Eventually(func(g Gomega) {
			response, err := client.CollectFailure(
				universal.Cluster, "demo-client", addr,
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.ResponseCode).To(Equal(statusCode))
		}).Should(Succeed())
	}

	It("should allow the traffic with meshtrafficpermission based on MeshService (http)", func() {
		// given no mesh traffic permissions
		trafficBlocked("test-server.mesh", 503)

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
		trafficAllowed("test-server.mesh")
	})

	It("should allow the traffic with meshtrafficpermission based on MeshService (tcp)", func() {
		// given no mesh traffic permissions
		trafficBlocked("test-server-tcp.mesh", 503)

		// when mesh traffic permission with MeshService
		yaml := `
type: MeshTrafficPermission
name: mtp-2
mesh: meshtrafficpermission
spec:
 targetRef:
   kind: MeshService
   name: test-server-tcp
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
		trafficAllowed("test-server-tcp.mesh")
	})

	It("should allow the traffic with traffic permission based on non standard tag", func() {
		// given no mesh traffic permission
		trafficBlocked("test-server.mesh", 503)
		trafficBlocked("test-server-tcp.mesh", 503)

		// when
		yaml := `
type: MeshTrafficPermission
name: mtp-3
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

		trafficAllowed("test-server.mesh")
		trafficAllowed("test-server-tcp.mesh")
	})

	It("should be able to allow the traffic with permissive mTLS (http)", func() {
		// given mesh traffic permission with permissive mTLS
		trafficBlocked("test-server.mesh", 503)
		permissive := samples.MeshDefaultBuilder().
			WithName(meshName).
			WithEnabledMTLSBackend("ca-1").
			WithBuiltinMTLSBackend("ca-1").
			WithPermissiveMTLSBackends().
			Build()
		Expect(universal.Cluster.Install(ResourceUniversal(permissive))).To(Succeed())

		// when specific MTP is applied
		yaml := `
type: MeshTrafficPermission
name: mtp-4
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
       action: Deny`
		Expect(universal.Cluster.Install(YamlUniversal(yaml))).To(Succeed())

		// then
		trafficBlocked("test-server.mesh", 403)

		// and it's still possible to access a service from outside the mesh
		publicAddress := net.JoinHostPort(universal.Cluster.GetApp("test-server").GetIP(), "80")
		trafficAllowed(publicAddress)
	})

	It("should be able to allow the traffic with permissive mTLS (tcp)", func() {
		// given mesh traffic permission with permissive mTLS
		trafficBlocked("test-server-tcp.mesh", 503)
		permissive := samples.MeshDefaultBuilder().
			WithName(meshName).
			WithEnabledMTLSBackend("ca-1").
			WithBuiltinMTLSBackend("ca-1").
			WithPermissiveMTLSBackends().
			Build()
		Expect(universal.Cluster.Install(ResourceUniversal(permissive))).To(Succeed())

		// when specific MTP is applied
		yaml := `
type: MeshTrafficPermission
name: mtp-5
mesh: meshtrafficpermission
spec:
 targetRef:
   kind: MeshService
   name: test-server-tcp
 from:
   - targetRef:
       kind: MeshService
       name: demo-client
     default:
       action: Deny`
		Expect(universal.Cluster.Install(YamlUniversal(yaml))).To(Succeed())

		// then
		trafficBlocked("test-server-tcp.mesh", 503)

		// and it's still possible to access a service from outside the mesh
		publicAddress := net.JoinHostPort(universal.Cluster.GetApp("test-server-tcp").GetIP(), "80")
		trafficAllowed(publicAddress)
	})
}
