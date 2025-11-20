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
	meshName := "workload-test"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshUniversal(meshName)).
			Install(MeshTrafficPermissionAllowAllUniversal(meshName)).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, meshName)
	})

	E2EAfterAll(func() {
		// Delete manually created dataplanes first
		_ = universal.Cluster.GetKumactlOptions().RunKumactl("delete", "dataplane", "dp-with-workload", "-m", meshName)
		_ = universal.Cluster.GetKumactlOptions().RunKumactl("delete", "dataplane", "manual-dp-1", "-m", meshName)
		_ = universal.Cluster.GetKumactlOptions().RunKumactl("delete", "dataplane", "manual-dp-2", "-m", meshName)

		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
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
`, meshName)
		Expect(universal.Cluster.Install(YamlUniversal(dataplane))).To(Succeed())

		// when workload generator runs
		// then a Workload resource should be auto-created
		Eventually(func(g Gomega) {
			expectManagedWorkload(g, universal.Cluster, "auto-workload", meshName)
		}, "1m", "1s").Should(Succeed())
	})

	It("should delete Workload after grace period when Dataplane is removed", func() {
		// given a dataplane with workload label
		dataplane := fmt.Sprintf(`
type: Dataplane
mesh: %s
name: dp-for-deletion
networking:
  address: 192.168.0.11
  inbound:
  - port: 80
    tags:
      kuma.io/service: deletion-test-service
      kuma.io/protocol: http
labels:
  kuma.io/workload: workload-for-deletion
`, meshName)
		Expect(universal.Cluster.Install(YamlUniversal(dataplane))).To(Succeed())

		// when workload is created
		Eventually(func(g Gomega) {
			expectManagedWorkload(g, universal.Cluster, "workload-for-deletion", meshName)
		}, "30s", "1s").Should(Succeed())

		// and dataplane is deleted
		Expect(universal.Cluster.GetKumactlOptions().RunKumactl("delete", "dataplane", "dp-for-deletion", "-m", meshName)).To(Succeed())

		// then workload should be marked with deletion grace period label
		Eventually(func(g Gomega) {
			workload, err := GetWorkload(universal.Cluster, "workload-for-deletion", meshName)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(workload.GetMeta().GetLabels()).To(HaveKey("kuma.io/deletion-grace-period-started-at"))
		}, "30s", "1s").Should(Succeed())

		// Note: Full deletion with grace period would take 1 hour by default,
		// so we only verify the grace period label is set
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
`, meshName)
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
`, meshName)

		Expect(universal.Cluster.Install(YamlUniversal(dataplane1))).To(Succeed())
		Expect(universal.Cluster.Install(YamlUniversal(dataplane2))).To(Succeed())

		// when workload generator runs
		// then only one Workload should be created
		Eventually(func(g Gomega) {
			expectManagedWorkload(g, universal.Cluster, "shared-workload", meshName)
		}, "30s", "1s").Should(Succeed())

		// verify only one workload named "shared-workload" exists
		workloads, err := universal.Cluster.GetKumactlOptions().KumactlList("workloads", meshName)
		Expect(err).ToNot(HaveOccurred())
		sharedWorkloadCount := 0
		for _, name := range workloads {
			if name == "shared-workload" {
				sharedWorkloadCount++
			}
		}
		Expect(sharedWorkloadCount).To(Equal(1))
	})

	It("should not manage manually created Workloads", func() {
		// given a manually created Workload without managed-by label
		manualWorkload := fmt.Sprintf(`
type: Workload
mesh: %s
name: manual-workload
spec: {}
`, meshName)
		Expect(universal.Cluster.Install(YamlUniversal(manualWorkload))).To(Succeed())

		// when workload generator runs
		// then the manual workload should persist
		Consistently(func(g Gomega) {
			workload, err := GetWorkload(universal.Cluster, "manual-workload", meshName)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(workload.GetMeta().GetLabels()).ToNot(HaveKeyWithValue("kuma.io/managed-by", "workload-generator"))
		}, "10s", "1s").Should(Succeed())
	})
}
