package kubernetes

import (
	"encoding/json"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/utils"
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
			framework.WithEgress(),
			framework.WithEgressEnvoyAdminTunnel(),
			framework.WithCtlOpts(map[string]string{
				"--set": "controlPlane.supportGatewaySecretsInAllNamespaces=true", // needed for test/e2e_env/kubernetes/gateway/gatewayapi.go:470
			}),
			// Occasionally CP will lose a leader in the E2E test just because of this deadline,
			// which does not make sense in such controlled environment (one k3d node, one instance of the CP).
			// 100s and 80s are values that we also use in mesh-perf when we put a lot of pressure on the CP.
			framework.WithEnv("KUMA_RUNTIME_KUBERNETES_LEADER_ELECTION_LEASE_DURATION", "100s"),
			framework.WithEnv("KUMA_RUNTIME_KUBERNETES_LEADER_ELECTION_RENEW_DEADLINE", "80s"),
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
		KumaCp:     Cluster.GetKuma().(*framework.K8sControlPlane).PortFwd(),
		MADS:       Cluster.GetKuma().(*framework.K8sControlPlane).MadsPortFwd(),
		ZoneEgress: Cluster.GetPortForward(framework.Config.ZoneEgressApp),
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
	Expect(Cluster.AddPortForward(state.ZoneEgress, framework.Config.ZoneEgressApp)).To(Succeed())
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

func ExpectCpToNotPanic() {
	logs, err := Cluster.GetKumaCPLogs()
	if err != nil {
		framework.Logf("could not retrieve cp logs")
	} else {
		Expect(utils.HasPanicInCpLogs(logs)).To(BeFalse())
	}
}

func SynchronizedAfterSuite() {
	Expect(framework.CpRestarted(Cluster)).To(BeFalse(), "CP restarted in this suite, this should not happen.")
	ExpectCpToNotPanic()
}

func AfterSuite(report ginkgo.Report) {
	PrintCPLogsOnFailure(report)
	PrintKubeState(report)
}
