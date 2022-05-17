package membership

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/kubernetes/env"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func Membership() {
	const ns1 = "membership-1"
	const ns2 = "membership-2"
	const mesh1 = "membership-1"
	const mesh2 = "membership-2"

	meshAllowingNamespace := func(mesh, ns string) string {
		return fmt.Sprintf(`apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: %s
spec:
  constraints:
    dataplaneProxy:
      requirements:
      - tags:
          k8s.kuma.io/namespace: %s`, mesh, ns)
	}
	BeforeAll(func() {
		err := NewClusterSetup().
			Install(NamespaceWithSidecarInjection(ns1)).
			Install(YamlK8s(meshAllowingNamespace(mesh1, ns1))).
			Install(NamespaceWithSidecarInjection(ns2)).
			Install(YamlK8s(meshAllowingNamespace(mesh2, ns2))).
			Setup(env.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})
	E2EAfterAll(func() {
		Expect(env.Cluster.TriggerDeleteNamespace(ns1)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(mesh1)).To(Succeed())
		Expect(env.Cluster.TriggerDeleteNamespace(ns2)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(mesh2)).To(Succeed())
	})

	It("should take into account membership when dp is connecting to the CP", func() {
		// when
		err := testserver.Install(
			testserver.WithNamespace(ns1),
			testserver.WithMesh(mesh1),
		)(env.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// then
		Eventually(func() (string, error) {
			return env.Cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes", "--mesh", mesh1)
		}, "30s", "1s").Should(ContainSubstring("test-server"))

		// when trying to change mesh to demo
		err = testserver.Install(
			testserver.WithNamespace(ns1),
			testserver.WithMesh(mesh2),
			testserver.WithoutWaitingToBeReady(),
		)(env.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// then the client is not allowed to do it
		// then it's not allowed
		// todo(jakubdyszkiewicz) uncomment once we can handle CP logs across all parallel executions
		// Eventually(func() (string, error) {
		//	return env.Cluster.GetKumaCPLogs()
		// }, "30s", "1s").Should(ContainSubstring("dataplane cannot be a member of mesh"))
		Consistently(func() (string, error) {
			return env.Cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes", "--mesh", mesh2)
		}, "10s", "5s").ShouldNot(ContainSubstring("test-server"))

		// when a new client is deployed in demo namespace in demo mesh
		err = testserver.Install(
			testserver.WithNamespace(ns2),
			testserver.WithMesh(mesh2),
		)(env.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// then it's allowed
		Eventually(func() (string, error) {
			return env.Cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes", "--mesh", mesh2)
		}, "30s", "1s").Should(ContainSubstring("test-server"))
	})
}
