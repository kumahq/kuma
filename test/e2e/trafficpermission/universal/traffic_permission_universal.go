package universal

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

var universalCluster Cluster

var meshDefaultMtlsOn = `
type: Mesh
name: default
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
`

var _ = E2EBeforeSuite(func() {
	universalCluster = NewUniversalCluster(NewTestingT(), Kuma1, Silent)

	Expect(NewClusterSetup().
		Install(Kuma(config_core.Standalone)).
		Install(YamlUniversal(meshDefaultMtlsOn)).
		Setup(universalCluster)).To(Succeed())

	testServerToken, err := universalCluster.GetKuma().GenerateDpToken("default", "test-server")
	Expect(err).ToNot(HaveOccurred())
	demoClientToken, err := universalCluster.GetKuma().GenerateDpToken("default", "demo-client")
	Expect(err).ToNot(HaveOccurred())

	Expect(TestServerUniversal("test-server", "default", testServerToken, WithArgs([]string{"echo", "--instance", "echo-v1"}))(universalCluster)).To(Succeed())
	Expect(DemoClientUniversal(AppModeDemoClient, "default", demoClientToken, WithTransparentProxy(true))(universalCluster)).To(Succeed())

	E2EDeferCleanup(universalCluster.DismissCluster)
})

func TrafficPermissionUniversal() {
	E2EAfterEach(func() {
		// remove all TrafficPermissions
		items, err := universalCluster.GetKumactlOptions().KumactlList("traffic-permissions", "default")
		Expect(err).ToNot(HaveOccurred())
		defaultFound := false
		for _, item := range items {
			if item == "allow-all-default" {
				defaultFound = true
				continue
			}
			err := universalCluster.GetKumactlOptions().KumactlDelete("traffic-permission", item, "default")
			Expect(err).ToNot(HaveOccurred())
		}

		if !defaultFound {
			yaml := `
type: TrafficPermission
name: allow-all-default
mesh: default
sources:
  - match:
      kuma.io/service: '*'
destinations:
  - match:
      kuma.io/service: '*'
`
			err = YamlUniversal(yaml)(universalCluster)
			Expect(err).ToNot(HaveOccurred())
		}
	})

	trafficAllowed := func() {
		stdout, _, err := universalCluster.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "--fail", "test-server.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
	}

	trafficBlocked := func() {
		Eventually(func() error {
			_, _, err := universalCluster.Exec("", "", "demo-client",
				"curl", "-v", "--fail", "test-server.mesh")
			return err
		}, "30s", "1s").Should(HaveOccurred())
	}

	removeDefaultTrafficPermission := func() {
		err := universalCluster.GetKumactlOptions().KumactlDelete("traffic-permission", "allow-all-default", "default")
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
mesh: default
sources:
  - match:
      kuma.io/service: demo-client
destinations:
  - match:
      kuma.io/service: test-server
`
		err := YamlUniversal(yaml)(universalCluster)
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
mesh: default
sources:
  - match:
      team: client-owners
destinations:
  - match:
      team: server-owners
`
		err := YamlUniversal(yaml)(universalCluster)
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
mesh: default
sources:
  - match:
      team: client-owners
      kuma.io/service: demo-client
destinations:
  - match:
      team: server-owners
      kuma.io/service: test-server
`
		err := YamlUniversal(yaml)(universalCluster)
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
mesh: default
sources:
  - match:
      kuma.io/service: non-existent-demo-client
destinations:
  - match:
      kuma.io/service: test-server
`
		err := YamlUniversal(yaml)(universalCluster)
		Expect(err).ToNot(HaveOccurred())

		// then
		trafficBlocked()

		// when traffic permission with the same "rank" is applied but later, it is preferred to the previous one
		yaml = `
type: TrafficPermission
name: example-2
mesh: default
sources:
  - match:
      kuma.io/service: demo-client
destinations:
  - match:
      kuma.io/service: test-server
`
		err = YamlUniversal(yaml)(universalCluster)
		Expect(err).ToNot(HaveOccurred())

		// then communication works because it was applied later
		trafficAllowed()
	})
}
