package graceful

import (
	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/kubernetes/env"
	. "github.com/kumahq/kuma/test/framework"
)

func Eviction() {
	nsName := "eviction"
	meshName := "eviction"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(NamespaceWithSidecarInjection(nsName)).
			Install(MeshKubernetes(meshName)).
			Setup(env.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(env.Cluster.TriggerDeleteNamespace(nsName)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("remove Dataplane of evicted Pod", func() {
		evictionPod := `apiVersion: v1
kind: Pod
metadata:
  name: to-be-evicted
  namespace: eviction
  annotations:
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
		Expect(env.Cluster.Install(YamlK8s(evictionPod))).To(Succeed())

		// then Dataplane should be created
		Eventually(func(g Gomega) {
			dataplanes, err := env.Cluster.GetKumactlOptions().KumactlList("dataplanes", meshName)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(dataplanes).Should(ContainElement(ContainSubstring("to-be-evicted")))
		}, "30s", "1s").Should(Succeed())

		// when it's evicted
		Eventually(func(g Gomega) {
			out, err := k8s.RunKubectlAndGetOutputE(env.Cluster.GetTesting(), env.Cluster.GetKubectlOptions(nsName), "get", "pods")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(out).To(ContainSubstring("Evicted"))
		}, "30s", "1s").Should(Succeed())

		// then Dataplane is removed
		Eventually(func(g Gomega) {
			dataplanes, err := env.Cluster.GetKumactlOptions().KumactlList("dataplanes", meshName)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(dataplanes).ShouldNot(ContainElement(ContainSubstring("to-be-evicted")))
		}, "60s", "1s").Should(Succeed())
	})
}
