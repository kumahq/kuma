package meshfaultinjection

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func Policy() {
	meshName := "mesh-fault-injection"
	timeout := fmt.Sprintf(`
type: MeshTimeout
mesh: "%s"
name: mesh-timeout
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshService
        name: test-server
      default:
        http:
          requestTimeout: 3s
`, meshName)
	faultInjection := fmt.Sprintf(`
type: MeshFaultInjection
mesh: "%s"
name: mesh-fault-injecton-402
spec:
  targetRef:
    kind: MeshService
    name: test-server
  from:
    - targetRef:
        kind: MeshService
        name: demo-client-blocked
      default:
        http:
          - abort:
              httpStatus: 402
              percentage: 100
    - targetRef:
        kind: MeshService
        name: demo-client-timeout
      default:
        http:
          - delay:
              value: 5s
              percentage: 100
`, meshName)
	faultInjectionAllSources := fmt.Sprintf(`
type: MeshFaultInjection
mesh: "%s"
name: mesh-fault-injecton-all
spec:
  targetRef:
    kind: MeshService
    name: test-service-block-all-sources
  from:
    - targetRef:
        kind: Mesh
      default:
        http:
          - abort:
              httpStatus: 421
              percentage: 100
`, meshName)
	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(MeshUniversal(meshName)).
			Install(YamlUniversal(faultInjection)).
			Install(YamlUniversal(faultInjectionAllSources)).
			Install(YamlUniversal(timeout)).
			Install(TestServerUniversal("test-server", meshName, WithArgs([]string{"echo", "--instance", "universal-1"}))).
			Install(TestServerUniversal("test-server-block-all-sources", meshName,
				WithArgs([]string{"echo", "--instance", "universal-1"}),
				WithServiceName("test-service-block-all-sources"),
			)).
			Install(DemoClientUniversal("demo-client", meshName, WithTransparentProxy(true))).
			Install(DemoClientUniversal("demo-client-blocked", meshName, WithTransparentProxy(true))).
			Install(DemoClientUniversal("demo-client-timeout", meshName, WithTransparentProxy(true))).
			Setup(universal.Cluster)).To(Succeed())
	})
	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should return specific error code for the demo-client-blocked", func() {
		stdout, _, err := universal.Cluster.Exec("", "", "demo-client", "curl", "-v", "test-server.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))

		stdout, _, err = universal.Cluster.Exec("", "", "demo-client-blocked", "curl", "-v", "test-server.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 402 Payment Required"))
	})

	It("should timeout for demo-client-timeout", func() {
		stdout, _, err := universal.Cluster.Exec("", "", "demo-client", "curl", "-v", "test-server.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))

		stdout, _, err = universal.Cluster.Exec("", "", "demo-client-timeout", "curl", "-v", "test-server.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("upstream request timeout"))
	})

	It("should return specific error code for all clients", func() {
		stdout, _, err := universal.Cluster.Exec("", "", "demo-client", "curl", "-v", "test-service-block-all-sources.mesh")
		println(stdout)
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 421 Misdirected Request"))

		stdout, _, err = universal.Cluster.Exec("", "", "demo-client-blocked", "curl", "-v", "test-service-block-all-sources.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 421 Misdirected Request"))

		stdout, _, err = universal.Cluster.Exec("", "", "demo-client-timeout", "curl", "-v", "test-service-block-all-sources.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 421 Misdirected Request"))
	})
}
