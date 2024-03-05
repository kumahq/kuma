package kubernetes

import (
	"encoding/json"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/onsi/ginkgo/v2"
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
	Expect(Cluster.Install(
		framework.GatewayAPICRDs,
	)).To(Succeed())

	kumaOptions := append(
		[]framework.KumaDeploymentOption{
			framework.WithCtlOpts(map[string]string{
				"--experimental-gatewayapi": "true",
			}),
			framework.WithEgress(),
		},
		framework.KumaDeploymentOptionsFromConfig(framework.Config.KumaCpConfig.Standalone.Kubernetes)...,
	)
	if framework.Config.KumaExperimentalSidecarContainers {
		kumaOptions = append(kumaOptions, framework.WithEnv("KUMA_EXPERIMENTAL_SIDECAR_CONTAINERS", "true"))
	}

	Eventually(func() error {
		return Cluster.Install(framework.Kuma(core.Zone, kumaOptions...))
	}, "90s", "3s").Should(Succeed())

	state := framework.K8sNetworkingState{
		KumaCp: Cluster.GetKuma().(*framework.K8sControlPlane).PortFwd(),
		MADS:   Cluster.GetKuma().(*framework.K8sControlPlane).MadsPortFwd(),
	}

	bytes, err := json.Marshal(state)
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
	state := framework.K8sNetworkingState{}
	Expect(json.Unmarshal(bytes, &state)).To(Succeed())

	Cluster = framework.NewK8sCluster(framework.NewTestingT(), framework.Kuma1, framework.Verbose)
	cp := framework.NewK8sControlPlane(
		Cluster.GetTesting(),
		core.Zone,
		Cluster.Name(),
		Cluster.GetKubectlOptions().ConfigPath,
		Cluster,
		Cluster.Verbose(),
		1,
		nil, // headers were not configured in setup
	)
	Expect(cp.FinalizeAddWithPortFwd(state.KumaCp, state.MADS)).To(Succeed())
	Cluster.SetCP(cp)
}

func PrintCPLogsOnFailure(report ginkgo.Report) {
	if !report.SuiteSucceeded {
		logs, err := Cluster.GetKumaCPLogs()
		if err != nil {
			framework.Logf("could not retrieve cp logs")
		} else {
			framework.Logf(logs)
		}
	}
}

func PrintKubeState(report ginkgo.Report) {
	if !report.SuiteSucceeded {
		// just running it, prints the logs
		if err := k8s.RunKubectlE(Cluster.GetTesting(), Cluster.GetKubectlOptions(), "get", "pods", "-A"); err != nil {
			framework.Logf("could not retrieve kube pods")
		}
	}
}
