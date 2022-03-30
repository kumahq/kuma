package hybrid

import (
	"fmt"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

func ApplicationHealthCheckOnKubernetesUniversal() {
	meshMTLSOn := func(mesh string) string {
		return fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: %s
spec:
  mtls:
    enabledBackend: ca-1
    backends:
    - name: ca-1
      type: builtin
`, mesh)
	}

	var globalK8s, zoneK8s, zoneUniversal Cluster

	BeforeEach(func() {
		k8sClusters, err := NewK8sClusters([]string{Kuma1, Kuma2}, Silent)
		Expect(err).ToNot(HaveOccurred())

		universalClusters, err := NewUniversalClusters([]string{Kuma3}, Silent)
		Expect(err).ToNot(HaveOccurred())

		globalK8s = k8sClusters.GetCluster(Kuma1)
		err = NewClusterSetup().
			Install(Kuma(core.Global)).
			Install(YamlK8s(meshMTLSOn("default"))).
			Setup(globalK8s)
		Expect(err).ToNot(HaveOccurred())

		zoneK8s = k8sClusters.GetCluster(Kuma2)
		err = NewClusterSetup().
			Install(Kuma(core.Zone,
				WithIngress(),
				WithGlobalAddress(globalK8s.GetKuma().GetKDSServerAddress()),
			)).
			Install(NamespaceWithSidecarInjection(TestNamespace)).
			Install(DemoClientK8s("default")).
			Setup(zoneK8s)
		Expect(err).ToNot(HaveOccurred())

		testServerToken, err := globalK8s.GetKuma().GenerateDpToken("default", "test-server")
		Expect(err).ToNot(HaveOccurred())

		zoneUniversal = universalClusters.GetCluster(Kuma3)
		ingressTokenKuma3, err := globalK8s.GetKuma().GenerateZoneIngressToken(Kuma3)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(Kuma(core.Zone,
				WithGlobalAddress(globalK8s.GetKuma().GetKDSServerAddress()),
			)).
			Install(TestServerUniversal("test-server-1", "default", testServerToken,
				WithArgs([]string{"echo", "--instance", "dp-universal-1"}),
				WithProtocol("tcp"))).
			Install(TestServerUniversal("test-server-2", "default", testServerToken,
				WithArgs([]string{"echo", "--instance", "dp-universal-2"}),
				WithProtocol("tcp"),
				ProxyOnly(),
				ServiceProbe())).
			Install(TestServerUniversal("test-server-3", "default", testServerToken,
				WithArgs([]string{"echo", "--instance", "dp-universal-3"}),
				WithProtocol("tcp"))).
			Install(IngressUniversal(ingressTokenKuma3)).
			Setup(zoneUniversal)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}

		Expect(zoneK8s.DeleteNamespace(TestNamespace)).To(Succeed())
		err := zoneK8s.DeleteKuma()
		Expect(err).ToNot(HaveOccurred())
		err = zoneK8s.DismissCluster()
		Expect(err).ToNot(HaveOccurred())

		err = zoneUniversal.DeleteKuma()
		Expect(err).ToNot(HaveOccurred())
		err = zoneUniversal.DismissCluster()
		Expect(err).ToNot(HaveOccurred())

		err = globalK8s.DeleteKuma()
		Expect(err).ToNot(HaveOccurred())
		err = globalK8s.DismissCluster()
		Expect(err).ToNot(HaveOccurred())
	})

	It("should not load balance requests to unhealthy instance", func() {
		pods, err := k8s.ListPodsE(zoneK8s.GetTesting(), zoneK8s.GetKubectlOptions(TestNamespace), metav1.ListOptions{
			LabelSelector: fmt.Sprintf("app=%s", "demo-client"),
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(pods).To(HaveLen(1))

		cmd := []string{"curl", "-v", "-m", "3", "--fail", "test-server.mesh"}

		instances := []string{"dp-universal-1", "dp-universal-3"}
		instanceSet := map[string]bool{}

		_, err = retry.DoWithRetryE(zoneK8s.GetTesting(), fmt.Sprintf("kubectl exec %s -- %s", pods[0].GetName(), strings.Join(cmd, " ")),
			100, 500*time.Millisecond, func() (string, error) {
				stdout, _, err := zoneK8s.Exec(TestNamespace, pods[0].GetName(), "demo-client", cmd...)
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

			stdout, _, err = zoneK8s.Exec(TestNamespace, pods[0].GetName(), "demo-client", cmd...)
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
