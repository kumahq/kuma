package graceful

import (
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/kubernetes/env"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func Graceful() {
	const name = "graceful"
	const namespace = "graceful"
	const mesh = "graceful"

	var clientPod string

	BeforeAll(func() {
		E2EDeferCleanup(func() {
			Expect(env.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
			Expect(env.Cluster.DeleteMesh(mesh))
		})

		err := NewClusterSetup().
			Install(MeshKubernetes(mesh)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(DemoClientK8s(mesh, namespace)).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(mesh),
				testserver.WithName(name),
			)).
			Setup(env.Cluster)
		Expect(err).To(Succeed())

		clientPod, err = PodNameOfApp(env.Cluster, "demo-client", namespace)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should not drop a request when scaling up and down", func() {
		// given no retries
		err := k8s.RunKubectlE(
			env.Cluster.GetTesting(),
			env.Cluster.GetKubectlOptions(),
			"delete", "retry", "retry-all-graceful",
		)
		Expect(err).ToNot(HaveOccurred())

		// and working instance of an application
		_, _, err = env.Cluster.ExecWithRetries(namespace, clientPod, "demo-client",
			"curl", "-v", "--fail", "graceful_graceful_svc_80.mesh")
		Expect(err).ToNot(HaveOccurred())

		// and constant traffic between client and server
		var failedErr error
		closeCh := make(chan struct{})
		defer close(closeCh)
		go func() {
			for {
				_, _, err := env.Cluster.Exec(namespace, clientPod, "demo-client",
					"curl", "-v", "--fail", "graceful_graceful_svc_80.mesh")
				if err != nil {
					failedErr = err
					return
				}
				select {
				case <-closeCh:
					return
				case <-time.After(100 * time.Millisecond):
				}
			}
		}()

		// when
		err = k8s.RunKubectlE(
			env.Cluster.GetTesting(),
			env.Cluster.GetKubectlOptions(namespace),
			"scale", "deployment", name, "--replicas", "2",
		)
		Expect(err).ToNot(HaveOccurred())

		// then
		Eventually(func(g Gomega) {
			g.Expect(WaitNumPods(namespace, 2, name)(env.Cluster)).To(Succeed())
			g.Expect(WaitPodsAvailable(namespace, name)(env.Cluster)).To(Succeed())
		}, "30s", "1s").Should(Succeed())
		Expect(failedErr).ToNot(HaveOccurred())

		// when
		err = k8s.RunKubectlE(
			env.Cluster.GetTesting(),
			env.Cluster.GetKubectlOptions(namespace),
			"scale", "deployment", name, "--replicas", "1",
		)
		Expect(err).ToNot(HaveOccurred())

		// then
		Eventually(func(g Gomega) {
			g.Expect(WaitNumPods(namespace, 1, name)(env.Cluster)).To(Succeed())
		}, "60s", "1s").Should(Succeed())

		Expect(failedErr).ToNot(HaveOccurred())
	})
}
