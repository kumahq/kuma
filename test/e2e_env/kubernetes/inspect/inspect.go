package inspect

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func Inspect() {
	nsName := "inspect"
	meshName := "inspect"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(NamespaceWithSidecarInjection(nsName)).
			Install(MeshKubernetes(meshName)).
			Install(DemoClientK8s(meshName, nsName)).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(nsName)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should return envoy config_dump", func() {
		// Synchronize on the dataplanes coming up.
		Eventually(func(g Gomega) {
			dataplanes, err := kubernetes.Cluster.GetKumactlOptions().KumactlList("dataplanes", meshName)
			g.Expect(err).ToNot(HaveOccurred())
			// Dataplane names are generated, so we check for a partial match.
			g.Expect(dataplanes).Should(ContainElement(ContainSubstring("demo-client")))
		}, "10s", "250ms").Should(Succeed())

		podName, err := PodNameOfApp(kubernetes.Cluster, "demo-client", nsName)
		Expect(err).ToNot(HaveOccurred())
		dataplaneName := fmt.Sprintf("%s.%s", podName, nsName)
		stdout, err := kubernetes.Cluster.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "dataplane", "-m", meshName, dataplaneName, "--type=config-dump")
		Expect(err).ToNot(HaveOccurred())

		Expect(stdout).To(ContainSubstring(fmt.Sprintf(`"name": "demo-client_%s_svc"`, nsName)))
		Expect(stdout).To(ContainSubstring(`"name": "inbound:passthrough:ipv4"`))
		Expect(stdout).To(ContainSubstring(`"name": "inbound:passthrough:ipv6"`))
		Expect(stdout).To(ContainSubstring(`"name": "kuma:envoy:admin"`))
		Expect(stdout).To(ContainSubstring(`"name": "outbound:passthrough:ipv4"`))
		Expect(stdout).To(ContainSubstring(`"name": "outbound:passthrough:ipv6"`))
	})
}
