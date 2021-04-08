package healthcheck_test

import (
	"fmt"
	"strings"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

var _ = Describe("Test application HealthCheck on Kubernetes/Universal", func() {
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

	namespaceWithSidecarInjection := func(namespace string) string {
		return fmt.Sprintf(`
apiVersion: v1
kind: Namespace
metadata:
  name: %s
  annotations:
    kuma.io/sidecar-injection: "enabled"
`, namespace)
	}

	var global, remoteK8s, remoteUniversal Cluster
	var optsRemoteK8s, optsRemoteUniversal []DeployOptionsFunc

	BeforeEach(func() {
		k8sClusters, err := NewK8sClusters([]string{Kuma1, Kuma2}, Silent)
		Expect(err).ToNot(HaveOccurred())

		universalClusters, err := NewUniversalClusters([]string{Kuma3}, Silent)
		Expect(err).ToNot(HaveOccurred())

		global = k8sClusters.GetCluster(Kuma1)
		err = NewClusterSetup().
			Install(Kuma(core.Global)).
			Install(YamlK8s(meshMTLSOn("default"))).
			Setup(global)
		Expect(err).ToNot(HaveOccurred())

		optsRemoteK8s = []DeployOptionsFunc{
			WithIngress(),
			WithGlobalAddress(global.GetKuma().GetKDSServerAddress()),
		}
		remoteK8s = k8sClusters.GetCluster(Kuma2)
		err = NewClusterSetup().
			Install(Kuma(core.Remote, optsRemoteK8s...)).
			Install(KumaDNS()).
			Install(YamlK8s(namespaceWithSidecarInjection(TestNamespace))).
			Install(DemoClientK8s("default")).
			Setup(remoteK8s)
		Expect(err).ToNot(HaveOccurred())

		echoServerToken, err := global.GetKuma().GenerateDpToken("default", "echo-server_kuma-test_svc_8080")
		Expect(err).ToNot(HaveOccurred())
		ingressToken, err := global.GetKuma().GenerateDpToken("default", "ingress")
		Expect(err).ToNot(HaveOccurred())

		optsRemoteUniversal = []DeployOptionsFunc{
			WithGlobalAddress(global.GetKuma().GetKDSServerAddress()),
		}
		remoteUniversal = universalClusters.GetCluster(Kuma3)
		err = NewClusterSetup().
			Install(Kuma(core.Remote, optsRemoteUniversal...)).
			Install(EchoServerUniversal("dp-echo-1", "default", "echo-universal-1", echoServerToken, WithProtocol("tcp"))).
			Install(EchoServerUniversal("dp-echo-2", "default", "echo-universal-2", echoServerToken, WithProtocol("tcp"), ProxyOnly(), ServiceProbe())).
			Install(EchoServerUniversal("dp-echo-3", "default", "echo-universal-3", echoServerToken, WithProtocol("tcp"))).
			Install(IngressUniversal("default", ingressToken)).
			Setup(remoteUniversal)
		Expect(err).ToNot(HaveOccurred())
		err = remoteUniversal.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}

		Expect(remoteK8s.DeleteNamespace(TestNamespace)).To(Succeed())
		err := remoteK8s.DeleteKuma(optsRemoteK8s...)
		Expect(err).ToNot(HaveOccurred())
		err = remoteK8s.DismissCluster()
		Expect(err).ToNot(HaveOccurred())

		err = remoteUniversal.DeleteKuma(optsRemoteUniversal...)
		Expect(err).ToNot(HaveOccurred())
		err = remoteUniversal.DismissCluster()
		Expect(err).ToNot(HaveOccurred())

		err = global.DeleteKuma()
		Expect(err).ToNot(HaveOccurred())
		err = global.DismissCluster()
		Expect(err).ToNot(HaveOccurred())

	})

	It("should not load balance requests to unhealthy instance", func() {
		pods, err := k8s.ListPodsE(remoteK8s.GetTesting(), remoteK8s.GetKubectlOptions(TestNamespace), metav1.ListOptions{
			LabelSelector: fmt.Sprintf("app=%s", "demo-client"),
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(pods).To(HaveLen(1))

		cmd := []string{"curl", "-v", "-m", "3", "--fail", "echo-server_kuma-test_svc_8080.mesh"}

		checkInstance := func(instance string) {
			_, err = retry.DoWithRetryE(remoteK8s.GetTesting(), fmt.Sprintf("kubectl exec %s -- %s", pods[0].GetName(), strings.Join(cmd, " ")),
				DefaultRetries, DefaultTimeout, func() (string, error) {
					stdout, _, err := remoteK8s.Exec(TestNamespace, pods[0].GetName(), "demo-client", cmd...)
					if err != nil {
						return "", err
					}
					if !strings.Contains(stdout, instance) {
						return "", errors.New("wrong instance")
					}
					return "", nil
				},
			)
		}

		checkInstance("echo-universal-1")
		checkInstance("echo-universal-2")

		var counter1, counter2, counter3 int
		const numOfRequest = 100

		for i := 0; i < numOfRequest; i++ {
			var stdout string

			stdout, _, err = remoteK8s.Exec(TestNamespace, pods[0].GetName(), "demo-client", cmd...)
			Expect(err).ToNot(HaveOccurred())

			switch {
			case strings.Contains(stdout, "Echo echo-universal-1"):
				counter1++
			case strings.Contains(stdout, "Echo echo-universal-2"):
				counter2++
			case strings.Contains(stdout, "Echo echo-universal-3"):
				counter3++
			}
		}

		Expect(counter2).To(Equal(0))
		Expect(counter1 > 0).To(BeTrue())
		Expect(counter3 > 0).To(BeTrue())
		Expect(counter1 + counter3).To(Equal(numOfRequest))
	})
})
