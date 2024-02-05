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
			Install(TestServerUniversal(
				"test-server",
				meshName,
				WithArgs([]string{"echo", "--instance", "echo-v1"}),
			)).
			Install(TestServerUniversal(
				"test-server-tcp",
				meshName,
				WithArgs([]string{"echo", "--instance", "test-server-tcp"}),
				WithServiceName("test-server-tcp"),
				WithProtocol("tcp"),
			)).
			Install(DemoClientUniversal(AppModeDemoClient, meshName, WithTransparentProxy(true))).
			Setup(universal.Cluster)).To(Succeed())
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
				universal.Cluster,
				"demo-client",
				addr,
			)
			g.Expect(err).ToNot(HaveOccurred())
		}).Should(Succeed())
	}

	httpTrafficBlocked := func(statusCode int) {
		GinkgoHelper()

		Eventually(func(g Gomega) {
			response, err := client.CollectFailure(
				universal.Cluster, "demo-client", "test-server.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.ResponseCode).To(Equal(statusCode))
		}).Should(Succeed())
	}

	tcpTrafficBlocked := func() {
		GinkgoHelper()

		Consistently(func(g Gomega) {
			stdout, _, _ := universal.Cluster.Exec(
				"",
				"",
				"dp-demo-client-mtls",
				"/bin/bash",
				"-c",
				"\"echo request | nc test-server-tcp.mesh 80\"",
			)

			// there is no real attempt to set up a connection with test-server,
			// but Envoy may return either empty response with EXIT_CODE = 0, or
			// 'Ncat: Connection reset by peer.' with EXIT_CODE = 1
			g.Expect(stdout).To(Or(
				BeEmpty(),
				ContainSubstring("Ncat: Connection reset by peer."),
			))
		}).Should(Succeed())
	}

	It("should allow the traffic with meshtrafficpermission based on MeshService (http)", func() {
		// given no mesh traffic permissions
		httpTrafficBlocked(403)

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
		tcpTrafficBlocked()

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
		httpTrafficBlocked(403)
		tcpTrafficBlocked()

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
		httpTrafficBlocked(403)
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
		httpTrafficBlocked(403)

		// and it's still possible to access a service from outside the mesh
		publicAddress := net.JoinHostPort(universal.Cluster.GetApp("test-server").GetIP(), "80")
		trafficAllowed(publicAddress)
	})

	It("should be able to allow the traffic with permissive mTLS (tcp)", func() {
		// given mesh traffic permission with permissive mTLS
		tcpTrafficBlocked()
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
		tcpTrafficBlocked()

		// and it's still possible to access a service from outside the mesh
		publicAddress := net.JoinHostPort(universal.Cluster.GetApp("test-server-tcp").GetIP(), "80")
		trafficAllowed(publicAddress)
	})
}
