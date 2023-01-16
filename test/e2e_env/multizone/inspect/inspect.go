package inspect

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func Inspect() {
	const meshName = "inspect"

	BeforeAll(func() {
		Expect(multizone.Global.Install(MTLSMeshUniversal(meshName))).To(Succeed())
		Expect(WaitForMesh(meshName, multizone.Zones())).To(Succeed())

		err := multizone.UniZone1.Install(TestServerUniversal("test-server", meshName,
			WithArgs([]string{"echo", "--instance", "echo"}),
		))
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(multizone.UniZone1.DeleteMeshApps(meshName)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(meshName)).To(Succeed())
	})

	type testCase struct {
		cluster     func() Cluster
		args        []string
		expectedOut string
	}
	GlobalCluster := func() Cluster {
		return multizone.Global
	}
	UniZone1Cluster := func() Cluster {
		return multizone.UniZone1
	}

	Context("Dataplane", func() {
		DescribeTable("should execute envoy inspection",
			func(given testCase) {
				Eventually(func(g Gomega) {
					args := append([]string{"inspect"}, given.args...)
					out, err := given.cluster().GetKumactlOptions().RunKumactlAndGetOutput(args...)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(out).Should(ContainSubstring(given.expectedOut))
				}, "30s", "1s").Should(Succeed())
			},
			Entry("of config dump for a dataplane using Global CP", testCase{
				cluster:     GlobalCluster,
				args:        []string{"dataplane", "kuma-4.test-server", "--type", "config-dump", "--mesh", meshName},
				expectedOut: `"dataplane.proxyType": "dataplane"`,
			}),
			Entry("of stats for a dataplane using Global CP", testCase{
				cluster:     GlobalCluster,
				args:        []string{"dataplane", "kuma-4.test-server", "--type", "stats", "--mesh", meshName},
				expectedOut: `server.live: 1`,
			}),
			Entry("of clusters for a dataplane using Global CP", testCase{
				cluster:     GlobalCluster,
				args:        []string{"dataplane", "kuma-4.test-server", "--type", "clusters", "--mesh", meshName},
				expectedOut: `kuma:envoy:admin::`,
			}),
			Entry("of config dump for a dataplane using Zone CP", testCase{
				cluster:     UniZone1Cluster,
				args:        []string{"dataplane", "test-server", "--type", "config-dump", "--mesh", meshName},
				expectedOut: `"dataplane.proxyType": "dataplane"`,
			}),
			Entry("of stats for a dataplane using Zone CP", testCase{
				cluster:     UniZone1Cluster,
				args:        []string{"dataplane", "test-server", "--type", "stats", "--mesh", meshName},
				expectedOut: `server.live: 1`,
			}),
			Entry("of clusters for a dataplane using Zone CP", testCase{
				cluster:     UniZone1Cluster,
				args:        []string{"dataplane", "test-server", "--type", "clusters", "--mesh", meshName},
				expectedOut: `kuma:envoy:admin::`,
			}),
			Entry("of config dump for a zoneingress using Global CP", testCase{
				cluster:     GlobalCluster,
				args:        []string{"zoneingress", "kuma-4.ingress", "--type", "config-dump"},
				expectedOut: `"dataplane.proxyType": "ingress"`,
			}),
			Entry("of stats for a zoneingress using Global CP", testCase{
				cluster:     GlobalCluster,
				args:        []string{"zoneingress", "kuma-4.ingress", "--type", "stats"},
				expectedOut: `server.live: 1`,
			}),
			Entry("of clusters for a zoneingress using Global CP", testCase{
				cluster:     GlobalCluster,
				args:        []string{"zoneingress", "kuma-4.ingress", "--type", "clusters"},
				expectedOut: `kuma:envoy:admin::`,
			}),
			Entry("of config dump for a zoneingress using Zone CP", testCase{
				cluster:     UniZone1Cluster,
				args:        []string{"zoneingress", "ingress", "--type", "config-dump"},
				expectedOut: `"dataplane.proxyType": "ingress"`,
			}),
			Entry("of stats for a zoneingress using Global CP", testCase{
				cluster:     UniZone1Cluster,
				args:        []string{"zoneingress", "ingress", "--type", "stats"},
				expectedOut: `server.live: 1`,
			}),
			Entry("of clusters for a zoneingress using Global CP", testCase{
				cluster:     UniZone1Cluster,
				args:        []string{"zoneingress", "ingress", "--type", "clusters"},
				expectedOut: `kuma:envoy:admin::`,
			}),
			Entry("of config dump for a zoneegress using Global CP", testCase{
				cluster:     GlobalCluster,
				args:        []string{"zoneegress", "kuma-4.egress", "--type", "config-dump"},
				expectedOut: `"dataplane.proxyType": "egress"`,
			}),
			Entry("of stats for a zoneegress using Global CP", testCase{
				cluster:     GlobalCluster,
				args:        []string{"zoneegress", "kuma-4.egress", "--type", "stats"},
				expectedOut: `server.live: 1`,
			}),
			Entry("of clusters for a zoneegress using Global CP", testCase{
				cluster:     GlobalCluster,
				args:        []string{"zoneegress", "kuma-4.egress", "--type", "clusters"},
				expectedOut: `kuma:envoy:admin::`,
			}),
			Entry("of config dump for a zoneegress using Zone CP", testCase{
				cluster:     UniZone1Cluster,
				args:        []string{"zoneegress", "egress", "--type", "config-dump"},
				expectedOut: `"dataplane.proxyType": "egress"`,
			}),
			Entry("of stats for a zoneegress using Global CP", testCase{
				cluster:     UniZone1Cluster,
				args:        []string{"zoneegress", "egress", "--type", "stats"},
				expectedOut: `server.live: 1`,
			}),
			Entry("of clusters for a zoneegress using Global CP", testCase{
				cluster:     UniZone1Cluster,
				args:        []string{"zoneegress", "egress", "--type", "clusters"},
				expectedOut: `kuma:envoy:admin::`,
			}),
		)
	})
}
