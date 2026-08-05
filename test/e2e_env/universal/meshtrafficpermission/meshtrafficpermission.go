package meshtrafficpermission

import (
	"fmt"
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/client"
	"github.com/kumahq/kuma/v3/test/framework/envs/universal"
)

func MeshTrafficPermissionUniversal() {
	meshName := "meshtrafficpermission"
	identityName := "meshtrafficpermission-identity"

	// SPIFFE ID of the demo client: the Bundled MeshIdentity renders the zone
	// trust domain and Universal defaults the path to the workload label. The
	// cluster only exists once the suite is running, so this is resolved lazily.
	demoClientSpiffeID := func() string {
		return fmt.Sprintf(
			"spiffe://%s/workload/%s",
			MeshIdentityTrustDomain(meshName, universal.Cluster),
			AppModeDemoClient,
		)
	}

	permissiveMeshTLS := fmt.Sprintf(`
type: MeshTLS
name: permissive
mesh: %s
spec:
  targetRef:
    kind: Mesh
  rules:
    - default:
        mode: Permissive
`, meshName)

	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(MeshUniversal(meshName)).
			Install(MeshIdentityBundled(meshName, identityName)).
			Install(TestServerUniversal(
				"test-server",
				meshName,
				WithArgs([]string{"echo", "--instance", "echo-v1"}),
				WithLabels(map[string]string{"kuma.io/service": "test-server", "team": "server-owners"}),
			)).
			Install(TestServerUniversal(
				"test-server-tcp",
				meshName,
				WithArgs([]string{"echo", "--instance", "test-server-tcp"}),
				WithServiceName("test-server-tcp"),
				WithProtocol("tcp"),
				WithLabels(map[string]string{"kuma.io/service": "test-server-tcp", "team": "server-owners"}),
			)).
			Install(DemoClientUniversal(AppModeDemoClient, meshName, WithTransparentProxy(true))).
			Setup(universal.Cluster)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, meshName)
	})

	E2EAfterEach(func() {
		// remove all MeshTrafficPermissions and MeshTLSes, so a permissive case
		// does not leak its mode into the next test
		for _, policy := range []struct{ plural, singular string }{
			{"meshtrafficpermissions", "meshtrafficpermission"},
			{"meshtlses", "meshtls"},
		} {
			items, err := universal.Cluster.GetKumactlOptions().KumactlList(policy.plural, meshName)
			Expect(err).ToNot(HaveOccurred())
			for _, item := range items {
				Expect(universal.Cluster.GetKumactlOptions().KumactlDelete(policy.singular, item, meshName)).To(Succeed())
			}
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
				universal.Cluster, "demo-client", "test-server.svc.mesh.local",
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
				AppModeDemoClient,
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
		yaml := fmt.Sprintf(`
type: MeshTrafficPermission
name: mtp-1
mesh: meshtrafficpermission
spec:
 targetRef:
   kind: Dataplane
   labels:
     kuma.io/service: test-server
 rules:
   - default:
       allow:
         - spiffeID:
             type: Prefix
             value: %s
`, demoClientSpiffeID())
		err := YamlUniversal(yaml)(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())
		trafficAllowed("test-server.svc.mesh.local")
	})

	It("should allow the traffic with meshtrafficpermission based on MeshService (tcp)", func() {
		// given no mesh traffic permissions
		tcpTrafficBlocked()

		// when mesh traffic permission with MeshService
		yaml := fmt.Sprintf(`
type: MeshTrafficPermission
name: mtp-2
mesh: meshtrafficpermission
spec:
 targetRef:
   kind: Dataplane
   labels:
     kuma.io/service: test-server-tcp
 rules:
   - default:
       allow:
         - spiffeID:
             type: Prefix
             value: %s
`, demoClientSpiffeID())
		err := YamlUniversal(yaml)(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// then
		trafficAllowed("test-server-tcp.svc.mesh.local")
	})

	It("should be able to allow the traffic with permissive mTLS (http)", func() {
		// given permissive mTLS
		httpTrafficBlocked(403)
		Expect(universal.Cluster.Install(YamlUniversal(permissiveMeshTLS))).To(Succeed())

		// when specific MTP is applied
		yaml := fmt.Sprintf(`
type: MeshTrafficPermission
name: mtp-4
mesh: meshtrafficpermission
spec:
 targetRef:
   kind: Dataplane
   labels:
     kuma.io/service: test-server
 rules:
   - default:
       deny:
         - spiffeID:
             type: Prefix
             value: %s`, demoClientSpiffeID())
		Expect(universal.Cluster.Install(YamlUniversal(yaml))).To(Succeed())

		// then
		httpTrafficBlocked(403)

		// and it's still possible to access a service from outside the mesh
		publicAddress := net.JoinHostPort(universal.Cluster.GetApp("test-server").GetIP(), "80")
		trafficAllowed(publicAddress)
	})

	It("should be able to allow the traffic with permissive mTLS (tcp)", func() {
		// given permissive mTLS
		tcpTrafficBlocked()
		Expect(universal.Cluster.Install(YamlUniversal(permissiveMeshTLS))).To(Succeed())

		// when specific MTP is applied
		yaml := fmt.Sprintf(`
type: MeshTrafficPermission
name: mtp-5
mesh: meshtrafficpermission
spec:
 targetRef:
   kind: Dataplane
   labels:
     kuma.io/service: test-server-tcp
 rules:
   - default:
       deny:
         - spiffeID:
             type: Prefix
             value: %s`, demoClientSpiffeID())
		Expect(universal.Cluster.Install(YamlUniversal(yaml))).To(Succeed())

		// then
		tcpTrafficBlocked()

		// and it's still possible to access a service from outside the mesh
		publicAddress := net.JoinHostPort(universal.Cluster.GetApp("test-server-tcp").GetIP(), "80")
		trafficAllowed(publicAddress)
	})
}
