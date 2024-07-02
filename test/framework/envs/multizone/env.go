package multizone

import (
	"encoding/json"
	"sync"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/test/framework"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/utils"
)

var (
	Global    *UniversalCluster
	KubeZone1 *K8sCluster
	KubeZone2 *K8sCluster
	UniZone1  *UniversalCluster
	UniZone2  *UniversalCluster
)

func Zones() []Cluster {
	return []Cluster{KubeZone1, KubeZone2, UniZone1, UniZone2}
}

type ZoneInfo struct {
	Mesh      string
	KubeZone1 string
	KubeZone2 string
	UniZone1  string
	UniZone2  string
}

func ZoneInfoForMesh(mesh string) ZoneInfo {
	return ZoneInfo{
		Mesh:      mesh,
		KubeZone1: KubeZone1.ZoneName(),
		KubeZone2: KubeZone2.ZoneName(),
		UniZone1:  UniZone1.ZoneName(),
		UniZone2:  UniZone2.ZoneName(),
	}
}

type State struct {
	Global    UniversalNetworkingState
	UniZone1  UniversalNetworkingState
	UniZone2  UniversalNetworkingState
	KubeZone1 K8sNetworkingState
	KubeZone2 K8sNetworkingState
}

func setupKubeZone(wg *sync.WaitGroup, clusterName string, extraOptions ...framework.KumaDeploymentOption) *K8sCluster {
	wg.Add(1)
	options := []framework.KumaDeploymentOption{
		WithEnv("KUMA_MULTIZONE_ZONE_KDS_NACK_BACKOFF", "1s"),
		WithIngress(),
		WithIngressEnvoyAdminTunnel(),
		WithEgress(),
		WithEgressEnvoyAdminTunnel(),
		WithGlobalAddress(Global.GetKuma().GetKDSServerAddress()),
	}
	options = append(options, extraOptions...)
	zone := NewK8sCluster(NewTestingT(), clusterName, Verbose)
	go func() {
		defer GinkgoRecover()
		defer wg.Done()
		Expect(zone.Install(Kuma(core.Zone, options...))).To(Succeed())
	}()
	return zone
}

// SetupAndGetState to be used with Ginkgo SynchronizedBeforeSuite
func SetupAndGetState() []byte {
	Global = NewUniversalCluster(NewTestingT(), Kuma3, Silent)
	E2EDeferCleanup(Global.DismissCluster) // clean up any containers if needed
	globalOptions := append(
		[]framework.KumaDeploymentOption{
			WithEnv("KUMA_MULTIZONE_GLOBAL_KDS_NACK_BACKOFF", "1s"),
		},
		framework.KumaDeploymentOptionsFromConfig(framework.Config.KumaCpConfig.Multizone.Global)...)
	Expect(Global.Install(Kuma(core.Global, globalOptions...))).To(Succeed())

	wg := sync.WaitGroup{}
	wg.Add(2)

	kubeZone1Options := append(
		framework.KumaDeploymentOptionsFromConfig(framework.Config.KumaCpConfig.Multizone.KubeZone1),
		WithEnv("KUMA_STORE_UNSAFE_DELETE", "true"),
	)
	if Config.IPV6 {
		// if the underneath clusters support IPv6, we'll configure kuma-1 with waitForDataplaneReady feature and
		// envoy admin binding to ::1 address
		kubeZone1Options = append(kubeZone1Options,
			WithEnv("KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_WAIT_FOR_DATAPLANE_READY", "true"),
			WithEnv("KUMA_BOOTSTRAP_SERVER_PARAMS_ADMIN_ADDRESS", "::1"),
		)
	}
	KubeZone1 = setupKubeZone(&wg, Kuma1, kubeZone1Options...)

	kubeZone2Options := framework.KumaDeploymentOptionsFromConfig(framework.Config.KumaCpConfig.Multizone.KubeZone2)
	KubeZone2 = setupKubeZone(&wg, Kuma2, kubeZone2Options...)

	UniZone1 = NewUniversalCluster(NewTestingT(), Kuma4, Silent)
	uniZone1Options := append(
		[]framework.KumaDeploymentOption{
			WithGlobalAddress(Global.GetKuma().GetKDSServerAddress()),
			WithEgressEnvoyAdminTunnel(),
			WithIngressEnvoyAdminTunnel(),
			WithEnv("KUMA_XDS_DATAPLANE_DEREGISTRATION_DELAY", "0s"), // we have only 1 Kuma CP instance so there is no risk setting this to 0
			WithEnv("KUMA_MULTIZONE_ZONE_KDS_NACK_BACKOFF", "1s"),
		},
		framework.KumaDeploymentOptionsFromConfig(framework.Config.KumaCpConfig.Multizone.UniZone1)...,
	)
	E2EDeferCleanup(UniZone1.DismissCluster) // clean up any containers if needed
	go func() {
		defer GinkgoRecover()
		defer wg.Done()
		err := NewClusterSetup().
			Install(Kuma(core.Zone, uniZone1Options...)).
			Install(IngressUniversal(Global.GetKuma().GenerateZoneIngressToken)).
			Install(EgressUniversal(Global.GetKuma().GenerateZoneEgressToken, WithConcurrency(1))).
			Setup(UniZone1)
		Expect(err).ToNot(HaveOccurred())
	}()

	vipCIDROverride := "251.0.0.0/8"
	if Config.IPV6 {
		vipCIDROverride = "fd00:fd11::/64"
	}
	UniZone2 = NewUniversalCluster(NewTestingT(), Kuma5, Silent)
	uniZone2Options := append(
		[]framework.KumaDeploymentOption{
			WithGlobalAddress(Global.GetKuma().GetKDSServerAddress()),
			WithEgressEnvoyAdminTunnel(),
			WithIngressEnvoyAdminTunnel(),
			WithEnv("KUMA_XDS_DATAPLANE_DEREGISTRATION_DELAY", "0s"), // we have only 1 Kuma CP instance so there is no risk setting this to 0
			WithEnv("KUMA_MULTIZONE_ZONE_KDS_NACK_BACKOFF", "1s"),
			WithEnv("KUMA_IPAM_MESH_SERVICE_CIDR", vipCIDROverride), // just to see that the status is not synced around
		},
		framework.KumaDeploymentOptionsFromConfig(framework.Config.KumaCpConfig.Multizone.UniZone2)...,
	)
	E2EDeferCleanup(UniZone2.DismissCluster) // clean up any containers if needed
	go func() {
		defer GinkgoRecover()
		defer wg.Done()
		err := NewClusterSetup().
			Install(Kuma(core.Zone, uniZone2Options...)).
			Install(IngressUniversal(Global.GetKuma().GenerateZoneIngressToken)).
			Install(EgressUniversal(Global.GetKuma().GenerateZoneEgressToken, WithConcurrency(1))).
			Setup(UniZone2)
		Expect(err).ToNot(HaveOccurred())
	}()
	wg.Wait()

	state := State{
		Global: UniversalNetworkingState{
			ZoneEgress:  Global.GetZoneEgressNetworking(),
			ZoneIngress: Global.GetZoneIngressNetworking(),
			KumaCp:      Global.GetKuma().(*UniversalControlPlane).Networking(),
		},
		UniZone1: UniversalNetworkingState{
			ZoneEgress:  UniZone1.GetZoneEgressNetworking(),
			ZoneIngress: UniZone1.GetZoneIngressNetworking(),
			KumaCp:      UniZone1.GetKuma().(*UniversalControlPlane).Networking(),
		},
		UniZone2: UniversalNetworkingState{
			ZoneEgress:  UniZone2.GetZoneEgressNetworking(),
			ZoneIngress: UniZone2.GetZoneIngressNetworking(),
			KumaCp:      UniZone2.GetKuma().(*UniversalControlPlane).Networking(),
		},
		KubeZone1: K8sNetworkingState{
			ZoneEgress:  KubeZone1.GetPortForward(Config.ZoneEgressApp),
			ZoneIngress: KubeZone1.GetPortForward(Config.ZoneIngressApp),
			KumaCp:      KubeZone1.GetKuma().(*K8sControlPlane).PortFwd(),
			MADS:        KubeZone1.GetKuma().(*K8sControlPlane).MadsPortFwd(),
		},
		KubeZone2: K8sNetworkingState{
			ZoneEgress:  KubeZone2.GetPortForward(Config.ZoneEgressApp),
			ZoneIngress: KubeZone2.GetPortForward(Config.ZoneIngressApp),
			KumaCp:      KubeZone2.GetKuma().(*K8sControlPlane).PortFwd(),
			MADS:        KubeZone2.GetKuma().(*K8sControlPlane).MadsPortFwd(),
		},
	}
	bytes, err := json.Marshal(state)
	Expect(err).ToNot(HaveOccurred())
	return bytes
}

func restoreKubeZone(clusterName string, networkingState *K8sNetworkingState) *K8sCluster {
	zone := NewK8sCluster(NewTestingT(), clusterName, Verbose)
	kubeCp := NewK8sControlPlane(
		zone.GetTesting(),
		core.Zone,
		zone.Name(),
		zone.GetKubectlOptions().ConfigPath,
		zone,
		zone.Verbose(),
		1,
		nil,
	)
	Expect(kubeCp.FinalizeAddWithPortFwd(networkingState.KumaCp, networkingState.KumaCp)).To(Succeed())
	zone.SetCP(kubeCp)
	Expect(zone.AddPortForward(networkingState.ZoneEgress, Config.ZoneEgressApp)).To(Succeed())
	Expect(zone.AddPortForward(networkingState.ZoneIngress, Config.ZoneIngressApp)).To(Succeed())
	return zone
}

// RestoreState to be used with Ginkgo SynchronizedBeforeSuite
func RestoreState(bytes []byte) {
	if Global != nil {
		return // cluster was already initiated with first function
	}
	state := State{}
	Expect(json.Unmarshal(bytes, &state)).To(Succeed())

	Global = NewUniversalCluster(NewTestingT(), Kuma3, Silent)
	E2EDeferCleanup(Global.DismissCluster) // clean up any containers if needed
	cp, err := NewUniversalControlPlane(
		Global.GetTesting(),
		core.Global,
		Global.Name(),
		Global.Verbose(),
		state.Global.KumaCp,
		nil,
		true,
	)
	Expect(err).ToNot(HaveOccurred())
	Global.SetCp(cp)

	KubeZone1 = restoreKubeZone(Kuma1, &state.KubeZone1)
	KubeZone2 = restoreKubeZone(Kuma2, &state.KubeZone2)

	UniZone1 = NewUniversalCluster(NewTestingT(), Kuma4, Silent)
	E2EDeferCleanup(UniZone1.DismissCluster) // clean up any containers if needed
	cp, err = NewUniversalControlPlane(
		UniZone1.GetTesting(),
		core.Zone,
		UniZone1.Name(),
		UniZone1.Verbose(),
		state.UniZone1.KumaCp,
		nil, // headers were not configured in setup
		true,
	)
	Expect(err).ToNot(HaveOccurred())
	UniZone1.SetCp(cp)
	Expect(UniZone1.AddNetworking(state.UniZone1.ZoneEgress, Config.ZoneEgressApp)).To(Succeed())
	Expect(UniZone1.AddNetworking(state.UniZone1.ZoneIngress, Config.ZoneIngressApp)).To(Succeed())

	UniZone2 = NewUniversalCluster(NewTestingT(), Kuma5, Silent)
	E2EDeferCleanup(UniZone2.DismissCluster) // clean up any containers if needed
	cp, err = NewUniversalControlPlane(
		UniZone2.GetTesting(),
		core.Zone,
		UniZone2.Name(),
		UniZone2.Verbose(),
		state.UniZone2.KumaCp,
		nil, // headers were not configured in setup
		true,
	)
	Expect(err).ToNot(HaveOccurred())
	UniZone2.SetCp(cp)
	Expect(UniZone2.AddNetworking(state.UniZone2.ZoneEgress, Config.ZoneEgressApp)).To(Succeed())
	Expect(UniZone2.AddNetworking(state.UniZone2.ZoneIngress, Config.ZoneIngressApp)).To(Succeed())
}

func PrintCPLogsOnFailure(report Report) {
	if !report.SuiteSucceeded {
		for _, cluster := range append(Zones(), Global) {
			Logf("\n\n\n\n\nCP logs of: " + cluster.Name())
			logs, err := cluster.GetKumaCPLogs()
			if err != nil {
				Logf("could not retrieve cp logs")
			} else {
				Logf(logs)
			}
		}
	}
}

func PrintKubeState(report Report) {
	if !report.SuiteSucceeded {
		for _, cluster := range []Cluster{KubeZone1, KubeZone2} {
			Logf("Kube state of cluster: " + cluster.Name())
			// just running it, prints the logs
			if err := k8s.RunKubectlE(cluster.GetTesting(), cluster.GetKubectlOptions(), "get", "pods", "-A"); err != nil {
				framework.Logf("could not retrieve kube pods")
			}
		}
	}
}

func ExpectCpsToNotCrash() {
	restartCount := framework.RestartCount(KubeZone1.GetKuma().(*framework.K8sControlPlane).GetKumaCPPods())
	Expect(restartCount).To(Equal(0), "Zone 1 CP restarted in this suite, this should not happen.")
	restartCount = framework.RestartCount(KubeZone2.GetKuma().(*framework.K8sControlPlane).GetKumaCPPods())
	Expect(restartCount).To(Equal(0), "Zone 2 CP restarted in this suite, this should not happen.")
}

func ExpectCpsToNotPanic() {
	for _, cluster := range append(Zones(), Global) {
		logs, err := cluster.GetKumaCPLogs()
		if err != nil {
			Logf("could not retrieve cp logs")
		} else {
			Expect(utils.HasPanicInCpLogs(logs)).To(BeFalse())
		}
	}
}
