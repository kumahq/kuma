package meshtimeout

import (
	"fmt"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/kubernetes/env"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func MeshTimeout() {
	namespace := "meshtimeout-namespace"
	mesh := "meshtimeout"
	var clientPodName string

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshKubernetes(mesh)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(DemoClientK8s(mesh, namespace)).
			Install(testserver.Install(testserver.WithMesh(mesh), testserver.WithNamespace(namespace))).
			Setup(env.Cluster)
		Expect(err).ToNot(HaveOccurred())

		clientPodName, err = PodNameOfApp(env.Cluster, "demo-client", namespace)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterEach(func() {
		Expect(
			k8s.RunKubectlE(env.Cluster.GetTesting(), env.Cluster.GetKubectlOptions(), "delete", "meshtimeouts", "-A", "--all"),
		).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(env.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(mesh)).To(Succeed())
	})

	DescribeTable("should add timeouts for outbound connections", func(timeoutConfig string) {
		// given no MeshTimeout
		mts, err := env.Cluster.GetKumactlOptions().KumactlList("meshtimeouts", mesh)
		Expect(err).ToNot(HaveOccurred())
		Expect(mts).To(HaveLen(0))
		Eventually(func(g Gomega) {
			start := time.Now()
			_, sterr, err := env.Cluster.Exec(namespace, clientPodName, "demo-client", "curl", "-v", "-H", "x-set-response-delay-ms: 5000", fmt.Sprintf("test-server_%s_svc_80.mesh", namespace))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(sterr).To(ContainSubstring("HTTP/1.1 200 OK"))
			g.Expect(time.Since(start)).To(BeNumerically(">", time.Second*5))
		}).WithTimeout(30 * time.Second).Should(Succeed())

		// when
		Expect(YamlK8s(timeoutConfig)(env.Cluster)).To(Succeed())

		// then
		Eventually(func(g Gomega) {
			stdout, _, err := env.Cluster.Exec(namespace, clientPodName, "demo-client", "curl", "-v", "-H", "x-set-response-delay-ms: 5000", fmt.Sprintf("test-server_%s_svc_80.mesh", namespace))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("upstream request timeout"))
		}).WithTimeout(30 * time.Second).Should(Succeed())
	},
		Entry("outbound timeout", fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: mt1
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: Mesh
      default:
        idleTimeout: 20s
        http:
          requestTimeout: 2s
          maxStreamDuration: 20s`, Config.KumaNamespace, mesh)),
		Entry("inbound timeout", fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: mt1
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Mesh
  from:
    - targetRef:
        kind: Mesh
      default:
        idleTimeout: 20s
        http:
          requestTimeout: 2s
          maxStreamDuration: 20s`, Config.KumaNamespace, mesh)),
	)
}
