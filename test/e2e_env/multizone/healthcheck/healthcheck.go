package healthcheck

import (
	"fmt"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/retry"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/test/e2e_env/multizone/env"
	. "github.com/kumahq/kuma/test/framework"
)

func ApplicationOnUniversalClientOnK8s() {
	namespace := "healthcheck-app-on-universal"
	meshName := "healthcheck-app-on-universal"

	BeforeAll(func() {
		err := env.Global.Install(MTLSMeshUniversal(meshName))
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(DemoClientK8s(meshName, namespace)).
			Setup(env.KubeZone1)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(TestServerUniversal("test-server-1", meshName,
				WithArgs([]string{"echo", "--instance", "dp-universal-1"}),
				WithProtocol("tcp"))).
			// This instance doesn't actually start the app
			Install(TestServerUniversal("test-server-2", meshName,
				WithArgs([]string{"echo", "--instance", "dp-universal-2"}),
				WithProtocol("tcp"),
				ProxyOnly(),
				ServiceProbe())).
			Install(TestServerUniversal("test-server-3", meshName,
				WithArgs([]string{"echo", "--instance", "dp-universal-3"}),
				WithProtocol("tcp"))).
			Setup(env.UniZone1)
		Expect(err).ToNot(HaveOccurred())
	})
	E2EAfterAll(func() {
		Expect(env.Global.DeleteMesh(meshName)).To(Succeed())
		Expect(env.KubeZone1.DeleteNamespace(namespace)).To(Succeed())
		Expect(env.UniZone1.DeleteMeshApps(meshName)).To(Succeed())
	})

	It("should not load balance requests to unhealthy instance", func() {
		pod, err := PodNameOfApp(env.KubeZone1, "demo-client", namespace)
		Expect(err).ToNot(HaveOccurred())

		cmd := []string{"curl", "-v", "-m", "3", "--fail", "test-server.mesh"}

		instances := []string{"dp-universal-1", "dp-universal-3"}
		instanceSet := map[string]bool{}

		_, err = retry.DoWithRetryE(env.KubeZone1.GetTesting(), fmt.Sprintf("kubectl exec %s -- %s", pod, strings.Join(cmd, " ")),
			100, 500*time.Millisecond, func() (string, error) {
				stdout, _, err := env.KubeZone1.Exec(namespace, pod, "demo-client", cmd...)
				if err != nil {
					return "", err
				}
				for _, instance := range instances {
					if strings.Contains(stdout, instance) {
						instanceSet[instance] = true
					}
				}
				if len(instanceSet) != len(instances) {
					return "", errors.Errorf("checked %d/%d instances", len(instanceSet), len(instances))
				}
				return "", nil
			},
		)
		Expect(err).ToNot(HaveOccurred())

		var counter1, counter2, counter3 int
		const numOfRequest = 100

		for i := 0; i < numOfRequest; i++ {
			var stdout string

			stdout, _, err = env.KubeZone1.Exec(namespace, pod, "demo-client", cmd...)
			Expect(err).ToNot(HaveOccurred())

			switch {
			case strings.Contains(stdout, "dp-universal-1"):
				counter1++
			case strings.Contains(stdout, "dp-universal-2"):
				counter2++
			case strings.Contains(stdout, "dp-universal-3"):
				counter3++
			}
		}

		Expect(counter2).To(Equal(0))
		Expect(counter1 > 0).To(BeTrue())
		Expect(counter3 > 0).To(BeTrue())
		Expect(counter1 + counter3).To(Equal(numOfRequest))
	})
}
