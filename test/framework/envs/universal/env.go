package universal

import (
	"encoding/json"
	"os"
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/universal_logs"
)

var Cluster *framework.UniversalCluster

// SetupAndGetState to be used with Ginkgo SynchronizedBeforeSuite
func SetupAndGetState() []byte {
	Cluster = framework.NewUniversalCluster(framework.NewTestingT(), framework.Kuma3, framework.Silent)
	framework.E2EDeferCleanup(Cluster.DismissCluster)
	kumaOptions := append(
		[]framework.KumaDeploymentOption{
			framework.WithEnv("KUMA_STORE_UNSAFE_DELETE", "true"),
			framework.WithEnv("KUMA_XDS_SERVER_DATAPLANE_STATUS_FLUSH_INTERVAL", "1s"), // speed up some tests by flushing stats quicker than default 10s
		}, framework.KumaDeploymentOptionsFromConfig(framework.Config.KumaCpConfig.Standalone.Universal)...)
	Expect(Cluster.Install(framework.Kuma(core.Standalone, kumaOptions...))).To(Succeed())
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

// WriteLogsIfFailed will read files used to put logs from universal tests
func WriteLogsIfFailed(ctx SpecContext) {
	logsPath := universal_logs.GetPath(
		framework.Config.UniversalE2ELogsPath,
		ctx.SpecReport().FullText(),
	)

	dir, err := os.ReadDir(logsPath)
	if err != nil {
		GinkgoWriter.Printf("Attaching component logs failed: %s", err)
		return
	}

	if !ctx.SpecReport().Failed() {
		if err := os.RemoveAll(logsPath); err != nil {
			GinkgoWriter.Printf("Failed to remove log dir: %s", err)
			return
		}
	}

	spacer := "------------------------------\n"

	for _, file := range dir {
		if !file.IsDir() {
			logFilePath := path.Join(logsPath, file.Name())

			bs, err := os.ReadFile(logFilePath)
			if err != nil {
				GinkgoWriter.Printf("Attaching component logs for %s failed: %s", logFilePath, err)
				continue
			}

			GinkgoWriter.Printf("%sLogs from: %s\n%s", spacer, logFilePath, spacer)
			Expect(GinkgoWriter.Write(bs)).Error().ToNot(HaveOccurred())
		}
	}
}

func RememberSpecID(ctx SpecContext) {
	universal_logs.GenAndSavePath(
		framework.Config.UniversalE2ELogsPath,
		ctx.SpecReport().FullText(),
	)
}
