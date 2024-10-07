package graceful

import (
	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func Eviction() {
	nsName := "eviction"
	meshName := "eviction"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(NamespaceWithSidecarInjection(nsName)).
			Install(MeshKubernetes(meshName)).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugKube(kubernetes.Cluster, meshName, nsName)
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(nsName)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("remove Dataplane of evicted Pod", func() {
		evictionPod := `apiVersion: v1
kind: Pod
metadata:
  name: to-be-evicted
  namespace: eviction
  labels:
    kuma.io/mesh: eviction
spec:
  volumes:
  containers:
  - name: alpine-evict
    image: alpine
    args:
    - /bin/ash
    - -c
    - --
    - "while true; do cat /usr/bin/* ; done"
    resources:
      limits:
        cpu: 50m
        ephemeral-storage: 10Ki
        memory: 64Mi`

		// when faulty pod is applied
		Expect(kubernetes.Cluster.Install(YamlK8s(evictionPod))).To(Succeed())

		// when it's evicted
		Eventually(func(g Gomega) {
			out, err := k8s.RunKubectlAndGetOutputE(
				kubernetes.Cluster.GetTesting(),
				kubernetes.Cluster.GetKubectlOptions(nsName),
				"get",
				"pod", "to-be-evicted",
				"-o", "go-template=\"{{.status.reason}}\"",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(out).To(ContainSubstring("Evicted"))
		}, "60s", "1s").Should(Succeed())

		// then Dataplane is removed
		Eventually(func(g Gomega) {
			dataplanes, err := kubernetes.Cluster.GetKumactlOptions().KumactlList("dataplanes", meshName)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(dataplanes).ShouldNot(ContainElement(ContainSubstring("to-be-evicted")))
		}, "60s", "1s").Should(Succeed())
	})
}
