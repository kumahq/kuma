package inspect

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/multizone/env"
	. "github.com/kumahq/kuma/test/framework"
)

func Inspect() {
	const meshName = "inspect"

	BeforeAll(func() {
		Expect(env.Global.Install(MTLSMeshUniversal(meshName))).To(Succeed())
		Expect(WaitForMesh(meshName, env.Zones())).To(Succeed())

		err := env.UniZone1.Install(TestServerUniversal("test-server", meshName,
			WithArgs([]string{"echo", "--instance", "echo"}),
		))
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(env.UniZone1.DeleteMeshApps(meshName)).To(Succeed())
		Expect(env.Global.DeleteMesh(meshName)).To(Succeed())
	})

	type testCase struct {
		typ         string
		expectedOut string
	}

	Context("Dataplane", func() {
		DescribeTable("should execute envoy inspection from Global CP",
			func(given testCase) {
				Eventually(func(g Gomega) {
					out, err := env.Global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "dataplane", "kuma-4.test-server", "--type", given.typ, "--mesh", meshName)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(out).Should(ContainSubstring(given.expectedOut))
				}, "30s", "1s").Should(Succeed())
			},
			Entry("config dump", testCase{
				typ:         "config-dump",
				expectedOut: `"dataplane.proxyType": "dataplane"`,
			}),
			Entry("stats", testCase{
				typ:         "stats",
				expectedOut: `server.live: 1`,
			}),
			Entry("clusters", testCase{
				typ:         "clusters",
				expectedOut: `kuma:envoy:admin`,
			}),
		)

		DescribeTable("should execute envoy inspection from Zone CP",
			func(given testCase) {
				Eventually(func(g Gomega) {
					out, err := env.UniZone1.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "dataplane", "test-server", "--type", given.typ, "--mesh", meshName)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(out).Should(ContainSubstring(given.expectedOut))
				}, "30s", "1s").Should(Succeed())
			},
			Entry("config dump", testCase{
				typ:         "config-dump",
				expectedOut: `"dataplane.proxyType": "dataplane"`,
			}),
			Entry("stats", testCase{
				typ:         "stats",
				expectedOut: `server.live: 1`,
			}),
			Entry("clusters", testCase{
				typ:         "clusters",
				expectedOut: `kuma:envoy:admin`,
			}),
		)
	})

	Context("ZoneIngress", func() {
		DescribeTable("should execute envoy inspection from Global CP",
			func(given testCase) {
				Eventually(func(g Gomega) {
					out, err := env.Global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zoneingress", "kuma-4.ingress", "--type", given.typ)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(out).Should(ContainSubstring(given.expectedOut))
				}, "30s", "1s").Should(Succeed())
			},
			Entry("config dump", testCase{
				typ:         "config-dump",
				expectedOut: `"dataplane.proxyType": "ingress"`,
			}),
			Entry("stats", testCase{
				typ:         "stats",
				expectedOut: `server.live: 1`,
			}),
			Entry("clusters", testCase{
				typ:         "clusters",
				expectedOut: `kuma:envoy:admin`,
			}),
		)

		DescribeTable("should execute envoy inspection from Zone CP",
			func(given testCase) {
				Eventually(func(g Gomega) {
					out, err := env.UniZone1.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zoneingress", "ingress", "--type", given.typ)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(out).Should(ContainSubstring(given.expectedOut))
				}, "30s", "1s").Should(Succeed())
			},
			Entry("config dump", testCase{
				typ:         "config-dump",
				expectedOut: `"dataplane.proxyType": "ingress"`,
			}),
			Entry("stats", testCase{
				typ:         "stats",
				expectedOut: `server.live: 1`,
			}),
			Entry("clusters", testCase{
				typ:         "clusters",
				expectedOut: `kuma:envoy:admin`,
			}),
		)
	})

	Context("ZoneEgress", func() {
		DescribeTable("should execute envoy inspection from Global CP",
			func(given testCase) {
				Eventually(func(g Gomega) {
					out, err := env.Global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zoneegress", "kuma-4.egress", "--type", given.typ)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(out).Should(ContainSubstring(given.expectedOut))
				}, "30s", "1s").Should(Succeed())
			},
			Entry("config dump", testCase{
				typ:         "config-dump",
				expectedOut: `"dataplane.proxyType": "egress"`,
			}),
			Entry("stats", testCase{
				typ:         "stats",
				expectedOut: `server.live: 1`,
			}),
			Entry("clusters", testCase{
				typ:         "clusters",
				expectedOut: `kuma:envoy:admin`,
			}),
		)

		DescribeTable("should execute envoy inspection from Zone CP",
			func(given testCase) {
				Eventually(func(g Gomega) {
					out, err := env.UniZone1.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zoneegress", "egress", "--type", given.typ)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(out).Should(ContainSubstring(given.expectedOut))
				}, "30s", "1s").Should(Succeed())
			},
			Entry("config dump", testCase{
				typ:         "config-dump",
				expectedOut: `"dataplane.proxyType": "egress"`,
			}),
			Entry("stats", testCase{
				typ:         "stats",
				expectedOut: `server.live: 1`,
			}),
			Entry("clusters", testCase{
				typ:         "clusters",
				expectedOut: `kuma:envoy:admin`,
			}),
		)
	})
}
