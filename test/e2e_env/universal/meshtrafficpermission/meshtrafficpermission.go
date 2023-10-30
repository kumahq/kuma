package meshtrafficpermission

import (
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

<<<<<<< HEAD
	"github.com/kumahq/kuma/test/e2e_env/universal/env"
=======
	"github.com/kumahq/kuma/pkg/test/resources/samples"
>>>>>>> 0e0489feb (fix(MeshTrafficPermission): support permissive mtls (#8171))
	. "github.com/kumahq/kuma/test/framework"
)

func MeshTrafficPermissionUniversal() {
	meshName := "meshtrafficpermission"

	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(MTLSMeshUniversal(meshName)).
			Install(TestServerUniversal("test-server", meshName, WithArgs([]string{"echo", "--instance", "echo-v1"}))).
			Install(DemoClientUniversal(AppModeDemoClient, meshName, WithTransparentProxy(true))).
			Setup(env.Cluster)).To(Succeed())

		// remove default traffic permission
		err := env.Cluster.GetKumactlOptions().KumactlDelete("traffic-permission", "allow-all-"+meshName, meshName)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(env.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	E2EAfterEach(func() {
		// remove all MeshTrafficPermissions
		items, err := env.Cluster.GetKumactlOptions().KumactlList("meshtrafficpermissions", meshName)
		Expect(err).ToNot(HaveOccurred())
		for _, item := range items {
			err := env.Cluster.GetKumactlOptions().KumactlDelete("meshtrafficpermission", item, meshName)
			Expect(err).ToNot(HaveOccurred())
		}
	})

<<<<<<< HEAD
	trafficAllowed := func() {
		stdout, _, err := env.Cluster.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "--fail", "test-server.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
=======
	trafficAllowed := func(addr string) {
		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", addr,
			)
			g.Expect(err).ToNot(HaveOccurred())
		}).Should(Succeed())
>>>>>>> 0e0489feb (fix(MeshTrafficPermission): support permissive mtls (#8171))
	}

	trafficBlocked := func() {
		Eventually(func() error {
			_, _, err := env.Cluster.Exec("", "", "demo-client",
				"curl", "-v", "--fail", "test-server.mesh")
			return err
		}, "30s", "1s").Should(HaveOccurred())
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
       action: ALLOW
`
		err := YamlUniversal(yaml)(env.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// then
		trafficAllowed("test-server.mesh")
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
        action: ALLOW
`
		err := YamlUniversal(yaml)(env.Cluster)
		Expect(err).ToNot(HaveOccurred())

		trafficAllowed("test-server.mesh")
	})

	It("should be able to allow the traffic with permissive mTLS", func() {
		// given mesh traffic permission with permissive mTLS
		trafficBlocked()
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
name: mtp-3
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
		trafficBlocked()

		// and it's still possible to access a service from outside the mesh
		publicAddress := net.JoinHostPort(universal.Cluster.GetApp("test-server").GetIP(), "80")
		trafficAllowed(publicAddress)
	})
}
