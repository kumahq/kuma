package universal

import (
	"encoding/json"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/test/framework"
)

var Cluster *framework.UniversalCluster

// SetupAndGetState to be used with Ginkgo SynchronizedBeforeSuite
func SetupAndGetState() []byte {
	Cluster = framework.NewUniversalCluster(framework.NewTestingT(), framework.Kuma3, framework.Silent)
	framework.E2EDeferCleanup(Cluster.DismissCluster)
	kumaOptions := append(
		[]framework.KumaDeploymentOption{
			framework.WithEnv("KUMA_XDS_SERVER_DATAPLANE_STATUS_FLUSH_INTERVAL", "1s"), // speed up some tests by flushing stats quicker than default 10s
			framework.WithEnv("KUMA_XDS_DATAPLANE_DEREGISTRATION_DELAY", "0s"),         // we have only 1 Kuma CP instance so there is no risk setting this to 0
		}, framework.KumaDeploymentOptionsFromConfig(framework.Config.KumaCpConfig.Standalone.Universal)...)
	Expect(Cluster.Install(framework.Kuma(core.Zone, kumaOptions...))).To(Succeed())
	Expect(Cluster.Install(framework.EgressUniversal(func(zone string) (string, error) {
		return Cluster.GetKuma().GenerateZoneEgressToken("")
	}))).To(Succeed())
	state := framework.UniversalNetworkingState{
		ZoneEgress: Cluster.GetZoneEgressNetworking(),
		KumaCp:     Cluster.GetKuma().(*framework.UniversalControlPlane).Networking(),
	}
	bytes, err := json.Marshal(state)
	Expect(err).ToNot(HaveOccurred())
	return bytes
}

// RestoreState to be used with Ginkgo SynchronizedBeforeSuite
func RestoreState(bytes []byte) {
	if Cluster != nil {
		return // cluster was already initiated with first function
	}
	state := framework.UniversalNetworkingState{}
	Expect(json.Unmarshal(bytes, &state)).To(Succeed())
	Cluster = framework.NewUniversalCluster(framework.NewTestingT(), framework.Kuma3, framework.Silent)
	framework.E2EDeferCleanup(Cluster.DismissCluster) // clean up any containers if needed
	cp, err := framework.NewUniversalControlPlane(
		Cluster.GetTesting(),
		core.Zone,
		Cluster.Name(),
		Cluster.Verbose(),
		state.KumaCp,
		nil, // headers were not configured in setup
		true,
	)
	Expect(err).ToNot(HaveOccurred())
	Expect(Cluster.AddNetworking(state.ZoneEgress, framework.Config.ZoneEgressApp)).To(Succeed())
	Cluster.SetCp(cp)
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
