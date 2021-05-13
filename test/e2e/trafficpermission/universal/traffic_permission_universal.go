package universal

import (
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core"
	. "github.com/kumahq/kuma/test/framework"
)

func TrafficPermissionUniversal() {
	var universalCluster Cluster
	var deployOptsFuncs []DeployOptionsFunc

	meshDefaulMtlsOn := `
type: Mesh
name: default
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
`

	E2EBeforeSuite(func() {
		core.SetLogger = func(l logr.Logger) {}
		logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))

		universalCluster = NewUniversalCluster(NewTestingT(), Kuma1, Silent)
		deployOptsFuncs = []DeployOptionsFunc{}

		err := NewClusterSetup().
			Install(Kuma(config_core.Standalone, deployOptsFuncs...)).
			Install(YamlUniversal(meshDefaulMtlsOn)).
			Setup(universalCluster)
		Expect(err).ToNot(HaveOccurred())
		err = universalCluster.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		echoServerToken, err := universalCluster.GetKuma().GenerateDpToken("default", "echo-server_kuma-test_svc_8080")
		Expect(err).ToNot(HaveOccurred())
		demoClientToken, err := universalCluster.GetKuma().GenerateDpToken("default", "demo-client")
		Expect(err).ToNot(HaveOccurred())

		err = EchoServerUniversal(AppModeEchoServer, "default", "universal-1", echoServerToken)(universalCluster)
		Expect(err).ToNot(HaveOccurred())
		err = DemoClientUniversal(AppModeDemoClient, "default", demoClientToken, WithTransparentProxy(true))(universalCluster)
		Expect(err).ToNot(HaveOccurred())
	})

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

	E2EAfterSuite(func() {
		Expect(universalCluster.DeleteKuma(deployOptsFuncs...)).To(Succeed())
		Expect(universalCluster.DismissCluster()).To(Succeed())
	})

	trafficAllowed := func() {
		stdout, _, err := universalCluster.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "--fail", "echo-server_kuma-test_svc_8080.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
	}

	trafficBlocked := func() {
		Eventually(func() error {
			_, _, err := universalCluster.Exec("", "", "demo-client",
				"curl", "-v", "--fail", "echo-server_kuma-test_svc_8080.mesh")
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
      kuma.io/service: echo-server_kuma-test_svc_8080
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
      kuma.io/service: echo-server_kuma-test_svc_8080
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
      kuma.io/service: echo-server_kuma-test_svc_8080
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
      kuma.io/service: echo-server_kuma-test_svc_8080
`
		err = YamlUniversal(yaml)(universalCluster)
		Expect(err).ToNot(HaveOccurred())

		// then communication works because it was applied later
		trafficAllowed()
	})
}
