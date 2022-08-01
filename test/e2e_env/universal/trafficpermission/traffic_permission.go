package trafficpermission

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/universal/env"
	. "github.com/kumahq/kuma/test/framework"
)

func TrafficPermissionUniversal() {
	meshName := "trafficpermission"

	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(MTLSMeshUniversal(meshName)).
			Install(TestServerUniversal("test-server", meshName, WithArgs([]string{"echo", "--instance", "echo-v1"}))).
			Install(DemoClientUniversal(AppModeDemoClient, meshName, WithTransparentProxy(true))).
			Setup(env.Cluster)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(env.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	E2EAfterEach(func() {
		// remove all TrafficPermissions
		items, err := env.Cluster.GetKumactlOptions().KumactlList("traffic-permissions", meshName)
		Expect(err).ToNot(HaveOccurred())
		defaultFound := false
		for _, item := range items {
			if item == "allow-all-"+meshName {
				defaultFound = true
				continue
			}
			err := env.Cluster.GetKumactlOptions().KumactlDelete("traffic-permission", item, meshName)
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
			err = YamlUniversal(yaml)(env.Cluster)
			Expect(err).ToNot(HaveOccurred())
		}
	})

	trafficAllowed := func() {
		stdout, _, err := env.Cluster.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "--fail", "test-server.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
	}

	trafficBlocked := func() {
		Eventually(func() error {
			_, _, err := env.Cluster.Exec("", "", "demo-client",
				"curl", "-v", "--fail", "test-server.mesh")
			return err
		}, "30s", "1s").Should(HaveOccurred())
	}

	removeDefaultTrafficPermission := func() {
		err := env.Cluster.GetKumactlOptions().KumactlDelete("traffic-permission", "allow-all-"+meshName, meshName)
		Expect(err).ToNot(HaveOccurred())
	}

	It("should allow the traffic with default traffic permission", func() {
		// given default traffic permission

		// then
		trafficAllowed()

		// when
		removeDefaultTrafficPermission()

		// then
		trafficBlocked()
	})

	It("should allow the traffic with traffic permission based on kuma.io/service tag", func() {
		// given no default traffic permission
		removeDefaultTrafficPermission()
		trafficBlocked()

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
		err := YamlUniversal(yaml)(env.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// then
		trafficAllowed()
	})

	It("should allow the traffic with traffic permission based on non standard tag", func() {
		// given no default traffic permission
		removeDefaultTrafficPermission()
		trafficBlocked()

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
		err := YamlUniversal(yaml)(env.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// then
		trafficAllowed()
	})

	It("should allow the traffic with traffic permission based on many tags", func() {
		// given no default traffic permission
		removeDefaultTrafficPermission()
		trafficBlocked()

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
		err := YamlUniversal(yaml)(env.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// then
		trafficAllowed()
	})

	It("should use most specific traffic permission", func() {
		// given default traffic permission
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
		err := YamlUniversal(yaml)(env.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// then
		trafficBlocked()

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
		err = YamlUniversal(yaml)(env.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// then communication works because it was applied later
		trafficAllowed()
	})
}
