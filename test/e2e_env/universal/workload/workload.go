package workload

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/envs/universal"
)

func expectManagedWorkload(g Gomega, cluster Cluster, name, mesh string) {
	workload, err := GetWorkload(cluster, name, mesh)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(workload.GetMeta().GetLabels()).To(HaveKeyWithValue("kuma.io/managed-by", "workload-generator"))
}

func Workload() {
	const mesh = "workload"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshUniversal(mesh)).
			Install(MeshTrafficPermissionAllowAllUniversal(mesh)).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, mesh)
	})

	E2EAfterAll(func() {
		// Delete manually created dataplanes first
		_ = universal.Cluster.GetKumactlOptions().RunKumactl("delete", "dataplane", "dp-with-workload", "-m", mesh)
		_ = universal.Cluster.GetKumactlOptions().RunKumactl("delete", "dataplane", "manual-dp-1", "-m", mesh)
		_ = universal.Cluster.GetKumactlOptions().RunKumactl("delete", "dataplane", "manual-dp-2", "-m", mesh)

		Expect(universal.Cluster.DeleteMeshApps(mesh)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(mesh)).To(Succeed())
	})

	It("should auto-generate Workload from Dataplane", func() {
		// given a dataplane with workload label
		dataplane := fmt.Sprintf(`
type: Dataplane
mesh: %s
name: dp-with-workload
networking:
  address: 192.168.0.10
  inbound:
  - port: 80
    tags:
      kuma.io/service: test-service
      kuma.io/protocol: http
labels:
  kuma.io/workload: auto-workload
`, mesh)
		Expect(universal.Cluster.Install(YamlUniversal(dataplane))).To(Succeed())

		// when workload generator runs
		// then a Workload resource should be auto-created
		Eventually(func(g Gomega) {
			expectManagedWorkload(g, universal.Cluster, "auto-workload", mesh)
		}, "1m", "1s").Should(Succeed())
	})

	It("should create single Workload for multiple Dataplanes with same workload label", func() {
		// given we manually create dataplanes with the same workload label
		dataplane1 := fmt.Sprintf(`
type: Dataplane
mesh: %s
name: manual-dp-1
networking:
  address: 192.168.0.1
  inbound:
  - port: 80
    tags:
      kuma.io/service: manual-service
      kuma.io/protocol: http
labels:
  kuma.io/workload: shared-workload
`, mesh)
		dataplane2 := fmt.Sprintf(`
type: Dataplane
mesh: %s
name: manual-dp-2
networking:
  address: 192.168.0.2
  inbound:
  - port: 80
    tags:
      kuma.io/service: manual-service
      kuma.io/protocol: http
labels:
  kuma.io/workload: shared-workload
`, mesh)

		Expect(universal.Cluster.Install(YamlUniversal(dataplane1))).To(Succeed())
		Expect(universal.Cluster.Install(YamlUniversal(dataplane2))).To(Succeed())

		// when workload generator runs
		// then only one Workload should be created
		Eventually(func(g Gomega) {
			expectManagedWorkload(g, universal.Cluster, "shared-workload", mesh)
		}, "30s", "1s").Should(Succeed())

		// verify only one workload named "shared-workload" exists
		workloads, err := universal.Cluster.GetKumactlOptions().KumactlList("workloads", mesh)
		Expect(err).ToNot(HaveOccurred())
		sharedWorkloadCount := 0
		for _, name := range workloads {
			if name == "shared-workload" {
				sharedWorkloadCount++
			}
		}
		Expect(sharedWorkloadCount).To(Equal(1))
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

	It("should update Workload status as dataplanes are added and removed", func() {
		const workloadName = "status-test-workload"
		const appName1 = "test-server-1"
		const appName2 = "test-server-2"

		// when deploy first test server with workload label
		err := TestServerUniversal(appName1, mesh,
			WithArgs([]string{"echo", "--instance", "v1"}),
			WithServiceName(appName1),
			WithAppLabel(appName1),
			WithAppendDataplaneYaml(fmt.Sprintf(`
labels:
  kuma.io/workload: %s`, workloadName)),
		)(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// then verify Workload status shows 1 dataplane
		Eventually(func(g Gomega) {
			workload, err := GetWorkload(universal.Cluster, workloadName, mesh)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(workload.Status.DataplaneProxies.Total).To(Equal(int32(1)))
		}).Should(Succeed())

		// when deploy second test server with same workload label
		err = TestServerUniversal(appName2, mesh,
			WithArgs([]string{"echo", "--instance", "v2"}),
			WithServiceName(appName2),
			WithAppLabel(appName2),
			WithAppendDataplaneYaml(fmt.Sprintf(`
labels:
  kuma.io/workload: %s`, workloadName)),
		)(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// then verify Workload status shows 2 dataplanes
		Eventually(func(g Gomega) {
			workload, err := GetWorkload(universal.Cluster, workloadName, mesh)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(workload.Status.DataplaneProxies.Total).To(Equal(int32(2)))
		}).Should(Succeed())

		// when delete first test server
		err = universal.Cluster.DeleteApp(appName1)
		Expect(err).ToNot(HaveOccurred())

		// then verify Workload status shows 1 dataplane again
		Eventually(func(g Gomega) {
			workload, err := GetWorkload(universal.Cluster, workloadName, mesh)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(workload.Status.DataplaneProxies.Total).To(Equal(int32(1)))
		}).Should(Succeed())
	})
}
