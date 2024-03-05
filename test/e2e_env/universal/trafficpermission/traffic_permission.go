package trafficpermission

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func TrafficPermission() {
	meshName := "trafficpermission"

	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(MTLSMeshUniversal(meshName)).
			Install(TrafficRouteUniversal(meshName)).
			Install(TestServerUniversal("test-server", meshName, WithArgs([]string{"echo", "--instance", "echo-v1"}))).
			Install(DemoClientUniversal(AppModeDemoClient, meshName, WithTransparentProxy(true))).
			Setup(universal.Cluster)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	E2EAfterEach(func() {
		// remove all TrafficPermissions
		items, err := universal.Cluster.GetKumactlOptions().KumactlList("traffic-permissions", meshName)
		Expect(err).ToNot(HaveOccurred())
		defaultFound := false
		for _, item := range items {
			if item == "allow-all-"+meshName {
				defaultFound = true
				continue
			}
			err := universal.Cluster.GetKumactlOptions().KumactlDelete("traffic-permission", item, meshName)
			Expect(err).ToNot(HaveOccurred())
		}

		if !defaultFound {
			yaml := `
type: TrafficPermission
name: allow-all-trafficpermission
mesh: trafficpermission
sources:
  - match:
      kuma.io/service: '*'
destinations:
  - match:
      kuma.io/service: '*'
`
			err = YamlUniversal(yaml)(universal.Cluster)
			Expect(err).ToNot(HaveOccurred())
		}
	})

	trafficAllowed := func() {
		GinkgoHelper()

		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				universal.Cluster, AppModeDemoClient, "test-server.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
		}).Should(Succeed())
	}

	trafficBlocked := func(statusCode int) {
		GinkgoHelper()

		Eventually(func(g Gomega) {
			response, err := client.CollectFailure(
				universal.Cluster, AppModeDemoClient, "test-server.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.ResponseCode).To(Equal(statusCode))
		}, "30s", "1s").Should(Succeed())
	}

	removeDefaultTrafficPermission := func() {
		GinkgoHelper()

		err := universal.Cluster.GetKumactlOptions().KumactlDelete("traffic-permission", "allow-all-"+meshName, meshName)
		Expect(err).ToNot(HaveOccurred())
	}

	addAllowAllTrafficPermission := func() {
		GinkgoHelper()

		Expect(NewClusterSetup().Install(TrafficPermissionUniversal(meshName)).Setup(universal.Cluster)).ToNot(HaveOccurred())
	}

	It("should allow the traffic with default traffic permission", func() {
		// given allow-all traffic permission
		addAllowAllTrafficPermission()

		// then
		trafficAllowed()

		// when
		removeDefaultTrafficPermission()

		// then
		trafficBlocked(403)
	})

	It("should allow the traffic with traffic permission based on kuma.io/service tag", func() {
		// given no default traffic permission
		removeDefaultTrafficPermission()
		trafficBlocked(403)

		// when traffic permission on service tag is applied
		yaml := `
type: TrafficPermission
name: example
mesh: trafficpermission
sources:
  - match:
      kuma.io/service: demo-client
destinations:
  - match:
      kuma.io/service: test-server
`
		err := YamlUniversal(yaml)(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// then
		trafficAllowed()
	})

	It("should allow the traffic with traffic permission based on non standard tag", func() {
		// given no default traffic permission
		removeDefaultTrafficPermission()
		trafficBlocked(403)

		// when
		yaml := `
type: TrafficPermission
name: other-tag-example
mesh: trafficpermission
sources:
  - match:
      team: client-owners
destinations:
  - match:
      team: server-owners
`
		err := YamlUniversal(yaml)(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// then
		trafficAllowed()
	})

	It("should allow the traffic with traffic permission based on many tags", func() {
		// given no default traffic permission
		removeDefaultTrafficPermission()
		trafficBlocked(403)

		// when
		yaml := `
type: TrafficPermission
name: other-tag-example
mesh: trafficpermission
sources:
  - match:
      team: client-owners
      kuma.io/service: demo-client
destinations:
  - match:
      team: server-owners
      kuma.io/service: test-server
`
		err := YamlUniversal(yaml)(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// then
		trafficAllowed()
	})

	It("should use most specific traffic permission", func() {
		// given allow-all traffic permission
		addAllowAllTrafficPermission()

		// then
		trafficAllowed()

		// when more specific traffic permission on service tag is applied
		yaml := `
type: TrafficPermission
name: example
mesh: trafficpermission
sources:
  - match:
      kuma.io/service: non-existent-demo-client
destinations:
  - match:
      kuma.io/service: test-server
`
		err := YamlUniversal(yaml)(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// then it's blocked with 503, because TrafficPermission configures L4 RBAC, not L7 RBAC.
		trafficBlocked(503)

		// when traffic permission with the same "rank" is applied but later, it is preferred to the previous one
		yaml = `
type: TrafficPermission
name: example-2
mesh: trafficpermission
sources:
  - match:
      kuma.io/service: demo-client
destinations:
  - match:
      kuma.io/service: test-server
`
		err = YamlUniversal(yaml)(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// then communication works because it was applied later
		trafficAllowed()
	})
}
