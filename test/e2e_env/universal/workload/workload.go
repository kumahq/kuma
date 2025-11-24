package workload

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/envs/universal"
)

func Workload() {
	const mesh = "workload"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshUniversal(mesh)).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, mesh)
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(mesh)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(mesh)).To(Succeed())
	})

	It("should deny DPP connection when workload label is missing for MeshIdentity using workload label", func() {
		// given MeshIdentity that uses workload label in path template
		meshIdentityYaml := fmt.Sprintf(`
type: MeshIdentity
name: mi-with-workload-label
mesh: %s
spec:
  selector:
    dataplane:
      matchLabels:
        app: test-server
  spiffeID:
    trustDomain: "{{ label \"kuma.io/mesh\" }}.mesh.local"
    path: "/workload/{{ label \"kuma.io/workload\" }}"
  provider:
    type: Bundled
    bundled:
      autogenerate:
        enabled: true
`, mesh)
		Expect(YamlUniversal(meshIdentityYaml)(universal.Cluster)).To(Succeed())

		// when trying to start proxy without workload label
		err := TestServerUniversal("test-server-without-label", mesh,
			WithArgs([]string{"echo", "--instance", "test-v1"}),
			WithServiceName("test-server"),
			WithAppLabel("test-server"),
		)(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// then the dataplane should not be registered due to validator rejection
		Consistently(func(g Gomega) {
			out, err := universal.Cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes", "-m", mesh, "-ojson")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(out).ToNot(ContainSubstring("test-server-without-label"))
		}).Should(Succeed())
	})

	It("should allow DPP connection when workload label is present for MeshIdentity using workload label", func() {
		// given MeshIdentity that uses workload label in path template
		meshIdentityYaml := fmt.Sprintf(`
type: MeshIdentity
name: mi-with-workload-label-2
mesh: %s
spec:
  selector:
    dataplane:
      matchLabels:
        app: backend
  spiffeID:
    trustDomain: "{{ label \"kuma.io/mesh\" }}.mesh.local"
    path: "/workload/{{ label \"kuma.io/workload\" }}"
  provider:
    type: Bundled
    bundled:
      autogenerate:
        enabled: true
`, mesh)
		Expect(YamlUniversal(meshIdentityYaml)(universal.Cluster)).To(Succeed())

		// when trying to start test server with workload label
		err := TestServerUniversal("backend-with-label", mesh,
			WithArgs([]string{"echo", "--instance", "backend-v1"}),
			WithServiceName("backend"),
			WithAppLabel("backend"),
			WithAppendDataplaneYaml(`
labels:
  kuma.io/workload: my-workload`),
		)(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// then the dataplane proxy should connect successfully
		Eventually(func(g Gomega) {
			out, err := universal.Cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes", "-m", mesh, "-ojson")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(out).To(ContainSubstring("backend-with-label"))
		}).Should(Succeed())
	})

	It("should allow DPP connection when MeshIdentity does not use workload label", func() {
		// given MeshIdentity that does NOT use workload label in path template
		meshIdentityYaml := fmt.Sprintf(`
type: MeshIdentity
name: mi-without-workload-label
mesh: %s
spec:
  selector:
    dataplane:
      matchLabels:
        app: api
  spiffeID:
    trustDomain: "mesh.local"
    path: "/ns/default/sa/{{ label \"kuma.io/service\" }}"
  provider:
    type: Bundled
    bundled:
      autogenerate:
        enabled: true
`, mesh)
		Expect(YamlUniversal(meshIdentityYaml)(universal.Cluster)).To(Succeed())

		// when trying to start test server without workload label
		err := TestServerUniversal("api-without-label", mesh,
			WithArgs([]string{"echo", "--instance", "api-v1"}),
			WithServiceName("api"),
			WithAppLabel("api"),
		)(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// then the dataplane proxy should connect successfully
		Eventually(func(g Gomega) {
			out, err := universal.Cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes", "-m", mesh, "-ojson")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(out).To(ContainSubstring("api-without-label"))
		}).Should(Succeed())
	})

	It("should allow DPP connection when no MeshIdentity applies", func() {
		// when trying to start test server with no matching MeshIdentity
		err := TestServerUniversal("other-service", mesh,
			WithArgs([]string{"echo", "--instance", "other-v1"}),
			WithServiceName("other"),
		)(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// then the dataplane proxy should connect successfully
		Eventually(func(g Gomega) {
			out, err := universal.Cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes", "-m", mesh, "-ojson")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(out).To(ContainSubstring("other-service"))
		}).Should(Succeed())
	})
}
