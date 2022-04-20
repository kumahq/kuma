package virtual_probes

import (
	"fmt"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/test/e2e_env/kubernetes/env"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func VirtualProbes() {
	const name = "test-server"
	const namespace = "virtual-probes"
	const mesh = "virtual-probes"

	BeforeAll(func() {
		E2EDeferCleanup(func() {
			Expect(env.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
			Expect(env.Cluster.DeleteMesh(mesh))
		})

		err := NewClusterSetup().
			Install(MTLSMeshKubernetes(mesh)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(mesh),
				testserver.WithName(name),
			)).
			Setup(env.Cluster)
		Expect(err).To(Succeed())
	})

	PollPodsReady := func() error {
		pods, err := k8s.ListPodsE(env.Cluster.GetTesting(), env.Cluster.GetKubectlOptions(namespace),
			metav1.ListOptions{LabelSelector: fmt.Sprintf("app=%s", name)})
		if err != nil {
			return err
		}
		for _, p := range pods {
			err := k8s.WaitUntilPodAvailableE(env.Cluster.GetTesting(), env.Cluster.GetKubectlOptions(namespace), p.GetName(), 0, 0)
			if err != nil {
				return err
			}
		}
		return nil
	}

	It("should deploy test-server with probes", func() {
		// Sample pod readiness to ensure they stay ready to at least 10sec.
		for i := 0; i < 10; i++ {
			time.Sleep(time.Second)
			Expect(PollPodsReady()).To(Succeed())
		}
	})
}
