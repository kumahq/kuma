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

var timeoutNamespace = "mesh-timeout-namespace"
var meshName = "meshtimeout"

func MeshTimeout() {

	var clientPodName string

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshKubernetes(meshName)).
			Install(NamespaceWithSidecarInjection(timeoutNamespace)).
			Install(DemoClientK8s(meshName, timeoutNamespace)).
			Install(testserver.Install(testserver.WithMesh(meshName), testserver.WithNamespace(timeoutNamespace))).
			Setup(env.Cluster)
		Expect(err).ToNot(HaveOccurred())

		clientPodName, err = PodNameOfApp(env.Cluster, "demo-client", timeoutNamespace)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterEach(func() {
		Expect(
			k8s.RunKubectlE(env.Cluster.GetTesting(), env.Cluster.GetKubectlOptions(), "delete", "meshtimeouts", "-A", "--all"),
		).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(env.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should add timeouts for outbound connections", func() {
		// given no MeshTimeout
		mts, err := env.Cluster.GetKumactlOptions().KumactlList("meshtimeouts", meshName)
		Expect(err).ToNot(HaveOccurred())
		Expect(mts).To(HaveLen(0))
		Eventually(func(g Gomega) {
			start := time.Now()
			stdout, _, err := env.Cluster.ExecWithOptions(execOptions(clientPodName))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
			g.Expect(time.Since(start)).To(BeNumerically(">", time.Second*5))
		}).Should(Succeed())

		// when
		Expect(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: mt1
  namespace: %s
  labels:
    kuma.io/mesh: meshtimeout
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
          maxStreamDuration: 20s
`, timeoutNamespace))(env.Cluster)).To(Succeed())

		// then
		Eventually(func(g Gomega) {
			stdout, _, err := env.Cluster.ExecWithOptions(execOptions(clientPodName))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("upstream request timeout"))
		}).Should(Succeed())
	})
}

func execOptions(clientPodName string) ExecOptions {
	return ExecOptions{
		Command:            []string{"curl", "-v", "-H", "\"x-set-response-delay-ms: 5000\"", fmt.Sprintf("test-server_%s_svc_80.mesh", timeoutNamespace)},
		Namespace:          timeoutNamespace,
		PodName:            clientPodName,
		ContainerName:      "demo-client",
		CaptureStdout:      true,
		CaptureStderr:      true,
		PreserveWhitespace: false,
		Retries:            4,
		Timeout:            8 * time.Second,
	}
}
