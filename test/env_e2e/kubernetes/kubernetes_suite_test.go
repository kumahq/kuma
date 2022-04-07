package auth_test

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/env_e2e/kubernetes/env"
	healthcheck "github.com/kumahq/kuma/test/env_e2e/kubernetes/healthcheck"
	"github.com/kumahq/kuma/test/env_e2e/kubernetes/jobs"
	. "github.com/kumahq/kuma/test/framework"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Kubernetes Suite")
}

var _ = SynchronizedBeforeSuite(
	func() []byte {
		env.Cluster = NewK8sCluster(NewTestingT(), Kuma1, Verbose)
		E2EDeferCleanup(env.Cluster.DeleteKuma)
		Expect(env.Cluster.Install(Kuma(core.Standalone, WithEnv("KUMA_STORE_UNSAFE_DELETE", "true")))).To(Succeed())
		portFwd := env.Cluster.GetKuma().(*K8sControlPlane).PortFwd()

		bytes, err := json.Marshal(portFwd)
		Expect(err).ToNot(HaveOccurred())
		return bytes
	},
	func(bytes []byte) {
		// Only one process should manage Kuma deployment
		// Other parallel processes should just replicate CP with its port forwards
		portFwd := PortFwd{}
		Expect(json.Unmarshal(bytes, &portFwd)).To(Succeed())

		env.Cluster = NewK8sCluster(NewTestingT(), Kuma1, Verbose)
		cp := NewK8sControlPlane(
			env.Cluster.GetTesting(),
			core.Standalone,
			env.Cluster.Name(),
			env.Cluster.GetKubectlOptions().ConfigPath,
			env.Cluster,
			env.Cluster.Verbose(),
			1,
		)
		Expect(cp.FinalizeAddWithPortFwd(portFwd)).To(Succeed())
		env.Cluster.SetCP(cp)
	},
)

var _ = Describe("Virtual Probes", healthcheck.VirtualProbes, Ordered)
var _ = Describe("Jobs", jobs.Jobs)
