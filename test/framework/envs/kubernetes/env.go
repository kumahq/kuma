package kubernetes

import (
	"encoding/json"
	"runtime"

	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/test/framework"
)

var Cluster *framework.K8sCluster

// SetupAndGetState to be used with Ginkgo SynchronizedBeforeSuite
func SetupAndGetState() []byte {
	Cluster = framework.NewK8sCluster(framework.NewTestingT(), framework.Kuma1, framework.Verbose)
	// The Gateway API webhook needs to start before we can create
	// GatewayClasses
	gatewayAPI := "false"
	// There's no arm64 webhook image yet
	if runtime.GOARCH == "amd64" {
		gatewayAPI = "true"
		Expect(Cluster.Install(
			framework.GatewayAPICRDs,
		)).To(Succeed())
	}
	kumaOptions := append([]framework.KumaDeploymentOption{
		framework.WithEnv("KUMA_STORE_UNSAFE_DELETE", "true"),
		framework.WithCtlOpts(map[string]string{
			"--experimental-gatewayapi": gatewayAPI,
		}),
		framework.WithEgress(),
	}, framework.KumaDeploymentOptionsFromConfig(framework.Config.KumaCpConfig.Standalone.Kubernetes)...)
	Eventually(func() error {
		return Cluster.Install(framework.Kuma(core.Standalone, kumaOptions...))
	}, "90s", "3s").Should(Succeed())
	portFwd := Cluster.GetKuma().(*framework.K8sControlPlane).PortFwd()

	bytes, err := json.Marshal(portFwd)
	Expect(err).ToNot(HaveOccurred())
	// Deliberately do not delete Kuma to save execution time (30s).
	// If everything is fine, K8S cluster will be deleted anyways
	// If something went wrong, we want to investigate it.
	return bytes
}

// RestoreState to be used with Ginkgo SynchronizedBeforeSuite
func RestoreState(bytes []byte) {
	if Cluster != nil {
		return // cluster was already initiated with first function
	}
	// Only one process should manage Kuma deployment
	// Other parallel processes should just replicate CP with its port forwards
	portFwd := framework.PortFwd{}
	Expect(json.Unmarshal(bytes, &portFwd)).To(Succeed())

	Cluster = framework.NewK8sCluster(framework.NewTestingT(), framework.Kuma1, framework.Verbose)
	cp := framework.NewK8sControlPlane(
		Cluster.GetTesting(),
		core.Standalone,
		Cluster.Name(),
		Cluster.GetKubectlOptions().ConfigPath,
		Cluster,
		Cluster.Verbose(),
		1,
	)
	Expect(cp.FinalizeAddWithPortFwd(portFwd)).To(Succeed())
	Cluster.SetCP(cp)
}
