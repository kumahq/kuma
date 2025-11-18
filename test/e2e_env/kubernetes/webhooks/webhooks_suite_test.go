package webhooks_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v2/pkg/test"
	"github.com/kumahq/kuma/v2/test/e2e_env/kubernetes/webhooks"
	. "github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/envs/kubernetes"
)

func TestE2E(t *testing.T) {
	test.RunE2ESpecs(t, "E2E Kubernetes Webhooks Suite")
}

var (
	_ = E2ESynchronizedBeforeSuite(func() []byte {
		// Don't install Kuma globally for this suite - each test will install it as needed
		kubernetes.Cluster = NewK8sCluster(NewTestingT(), Kuma1, Verbose)
		return nil
	}, func([]byte) {
		// Restore cluster reference
		kubernetes.Cluster = NewK8sCluster(NewTestingT(), Kuma1, Verbose)
	})
	_ = SynchronizedAfterSuite(func() {}, func() {
		if kubernetes.Cluster != nil {
			Expect(kubernetes.Cluster.DeleteKuma()).To(Succeed())
			Expect(kubernetes.Cluster.DismissCluster()).To(Succeed())
		}
	})
)

var _ = Describe("Webhooks Cert-Manager CA Injection", webhooks.CertManagerCAInjection, Ordered)
