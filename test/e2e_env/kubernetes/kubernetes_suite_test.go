package auth_test

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/container_patch"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/env"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/graceful"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/healthcheck"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/jobs"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/membership"
	. "github.com/kumahq/kuma/test/framework"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Kubernetes Suite")
}

var _ = SynchronizedBeforeSuite(
	func() []byte {
		env.Cluster = NewK8sCluster(NewTestingT(), Kuma1, Verbose)
		Expect(env.Cluster.Install(Kuma(core.Standalone,
			WithEnv("KUMA_STORE_UNSAFE_DELETE", "true"),
			WithCtlOpts(map[string]string{"--experimental-meshgateway": "true"}),
		))).To(Succeed())
		portFwd := env.Cluster.GetKuma().(*K8sControlPlane).PortFwd()

		bytes, err := json.Marshal(portFwd)
		Expect(err).ToNot(HaveOccurred())
		// Deliberately do not delete Kuma to save execution time (30s).
		// If everything is fine, K8S cluster will be deleted anyways
		// If something went wrong, we want to investigate it.
		return bytes
	},
	func(bytes []byte) {
		if env.Cluster != nil {
			return // cluster was already initiated with first function
		}
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

// SynchronizedAfterSuite keeps the main process alive until all other processes finish.
// Otherwise, we would close port-forward to the CP and remaining tests executed in different processes may fail.
var _ = SynchronizedAfterSuite(func() {}, func() {})

var _ = Describe("Virtual Probes", healthcheck.VirtualProbes, Ordered)
var _ = Describe("Graceful", graceful.Graceful, Ordered)
var _ = Describe("Jobs", jobs.Jobs)
var _ = Describe("Membership", membership.Membership, Ordered)
var _ = FDescribe("Container Patch", container_patch.ContainerPatch, Ordered)
