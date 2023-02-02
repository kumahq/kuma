package universal

import (
	"encoding/json"

	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/test/framework"
)

var Cluster *framework.UniversalCluster

// SetupAndGetState to be used with Ginkgo SynchronizedBeforeSuite
func SetupAndGetState(opts ...framework.KumaDeploymentOption) []byte {
	Cluster = framework.NewUniversalCluster(framework.NewTestingT(), framework.Kuma3, framework.Silent)
	framework.E2EDeferCleanup(Cluster.DismissCluster)
	kumaOpts := append([]framework.KumaDeploymentOption{
		framework.WithEnv("KUMA_STORE_UNSAFE_DELETE", "true"),
		framework.WithEnv("KUMA_XDS_SERVER_DATAPLANE_STATUS_FLUSH_INTERVAL", "1s"), // speed up some tests by flushing stats quicker than default 10s
	}, opts...)
	Expect(Cluster.Install(framework.Kuma(core.Standalone,
		kumaOpts...,
	))).To(Succeed())
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
		core.Standalone,
		Cluster.Name(),
		Cluster.Verbose(),
		state.KumaCp,
	)
	Expect(err).ToNot(HaveOccurred())
	Expect(Cluster.AddNetworking(state.ZoneEgress, framework.Config.ZoneEgressApp)).To(Succeed())
	Cluster.SetCp(cp)
}
