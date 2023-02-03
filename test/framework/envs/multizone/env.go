package multizone

import (
	"encoding/json"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/test/framework"
	. "github.com/kumahq/kuma/test/framework"
)

var Global *UniversalCluster
var KubeZone1 *K8sCluster
var KubeZone2 *K8sCluster
var UniZone1 *UniversalCluster
var UniZone2 *UniversalCluster

func Zones() []Cluster {
	return []Cluster{KubeZone1, KubeZone2, UniZone1, UniZone2}
}

type State struct {
	Global    UniversalNetworkingState
	UniZone1  UniversalNetworkingState
	UniZone2  UniversalNetworkingState
	KubeZone1 K8sNetworkingState
	KubeZone2 K8sNetworkingState
}

// SetupAndGetState to be used with Ginkgo SynchronizedBeforeSuite
func SetupAndGetState() []byte {
	Global = NewUniversalCluster(NewTestingT(), Kuma3, Silent)
	E2EDeferCleanup(Global.DismissCluster) // clean up any containers if needed
	Expect(Global.Install(Kuma(core.Global,
		getKumaDeploymentOptions(Config.KumaCpConfig.Multizone.Global, WithEnv("KUMA_STORE_UNSAFE_DELETE", "true"))...,
	))).To(Succeed())

	wg := sync.WaitGroup{}
	wg.Add(4)

	KubeZone1 = NewK8sCluster(NewTestingT(), Kuma1, Verbose)
	go func() {
		defer GinkgoRecover()
		defer wg.Done()
		Expect(KubeZone1.Install(Kuma(core.Zone,
			getKumaDeploymentOptions(
				Config.KumaCpConfig.Multizone.KubeZone1,
				WithEnv("KUMA_STORE_UNSAFE_DELETE", "true"),
				WithIngress(),
				WithIngressEnvoyAdminTunnel(),
				WithEgress(),
				WithEgressEnvoyAdminTunnel(),
				WithGlobalAddress(Global.GetKuma().GetKDSServerAddress()),
			)...,
		))).To(Succeed())
	}()

	KubeZone2 = NewK8sCluster(NewTestingT(), Kuma2, Verbose)
	go func() {
		defer GinkgoRecover()
		defer wg.Done()
		Expect(KubeZone2.Install(Kuma(core.Zone,
			getKumaDeploymentOptions(
				Config.KumaCpConfig.Multizone.KubeZone2,
				WithEnv("KUMA_STORE_UNSAFE_DELETE", "true"),
				WithEnv("KUMA_DEFAULTS_ENABLE_LOCALHOST_INBOUND_CLUSTERS", "true"),
				WithIngress(),
				WithIngressEnvoyAdminTunnel(),
				WithEgress(),
				WithEgressEnvoyAdminTunnel(),
				WithGlobalAddress(Global.GetKuma().GetKDSServerAddress()),
				WithExperimentalCNI(),
			)...,
		))).To(Succeed())
	}()

	UniZone1 = NewUniversalCluster(NewTestingT(), Kuma4, Silent)
	E2EDeferCleanup(UniZone1.DismissCluster) // clean up any containers if needed
	go func() {
		defer GinkgoRecover()
		defer wg.Done()
		err := NewClusterSetup().
			Install(Kuma(core.Zone,
				getKumaDeploymentOptions(
					Config.KumaCpConfig.Multizone.UniZone1,
					WithGlobalAddress(Global.GetKuma().GetKDSServerAddress()),
					WithEnv("KUMA_STORE_UNSAFE_DELETE", "true"),
					WithEgressEnvoyAdminTunnel(),
					WithIngressEnvoyAdminTunnel(),
				)...,
			)).
			Install(IngressUniversal(Global.GetKuma().GenerateZoneIngressLegacyToken)).
			Install(EgressUniversal(Global.GetKuma().GenerateZoneEgressToken)).
			Setup(UniZone1)
		Expect(err).ToNot(HaveOccurred())
	}()

	UniZone2 = NewUniversalCluster(NewTestingT(), Kuma5, Silent)
	E2EDeferCleanup(UniZone2.DismissCluster) // clean up any containers if needed
	go func() {
		defer GinkgoRecover()
		defer wg.Done()
		err := NewClusterSetup().
			Install(Kuma(core.Zone,
				getKumaDeploymentOptions(
					Config.KumaCpConfig.Multizone.UniZone2,
					WithGlobalAddress(Global.GetKuma().GetKDSServerAddress()),
					WithEnv("KUMA_STORE_UNSAFE_DELETE", "true"),
					WithEnv("KUMA_DEFAULTS_ENABLE_LOCALHOST_INBOUND_CLUSTERS", "true"),
					WithEgressEnvoyAdminTunnel(),
					WithIngressEnvoyAdminTunnel(),
				)...,
			)).
			Install(IngressUniversal(Global.GetKuma().GenerateZoneIngressToken)).
			Install(EgressUniversal(Global.GetKuma().GenerateZoneEgressToken)).
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
			ZoneEgress:  KubeZone1.GetZoneEgressPortForward(),
			ZoneIngress: KubeZone1.GetZoneIngressPortForward(),
			KumaCp:      KubeZone1.GetKuma().(*K8sControlPlane).PortFwd(),
		},
		KubeZone2: K8sNetworkingState{
			ZoneEgress:  KubeZone2.GetZoneEgressPortForward(),
			ZoneIngress: KubeZone2.GetZoneIngressPortForward(),
			KumaCp:      KubeZone2.GetKuma().(*K8sControlPlane).PortFwd(),
		},
	}
	bytes, err := json.Marshal(state)
	Expect(err).ToNot(HaveOccurred())
	return bytes
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
	)
	Expect(err).ToNot(HaveOccurred())
	Global.SetCp(cp)

	KubeZone1 = NewK8sCluster(NewTestingT(), Kuma1, Verbose)
	kubeCp := NewK8sControlPlane(
		KubeZone1.GetTesting(),
		core.Zone,
		KubeZone1.Name(),
		KubeZone1.GetKubectlOptions().ConfigPath,
		KubeZone1,
		KubeZone1.Verbose(),
		1,
	)
	Expect(kubeCp.FinalizeAddWithPortFwd(state.KubeZone1.KumaCp)).To(Succeed())
	KubeZone1.SetCP(kubeCp)
	Expect(KubeZone1.AddPortForward(state.KubeZone1.ZoneEgress, Config.ZoneEgressApp)).To(Succeed())
	Expect(KubeZone1.AddPortForward(state.KubeZone1.ZoneIngress, Config.ZoneIngressApp)).To(Succeed())

	KubeZone2 = NewK8sCluster(NewTestingT(), Kuma2, Verbose)
	kubeCp = NewK8sControlPlane(
		KubeZone2.GetTesting(),
		core.Zone,
		KubeZone2.Name(),
		KubeZone2.GetKubectlOptions().ConfigPath,
		KubeZone2,
		KubeZone2.Verbose(),
		1,
	)
	Expect(kubeCp.FinalizeAddWithPortFwd(state.KubeZone2.KumaCp)).To(Succeed())
	KubeZone2.SetCP(kubeCp)
	Expect(KubeZone2.AddPortForward(state.KubeZone2.ZoneEgress, Config.ZoneEgressApp)).To(Succeed())
	Expect(KubeZone2.AddPortForward(state.KubeZone2.ZoneIngress, Config.ZoneIngressApp)).To(Succeed())

	UniZone1 = NewUniversalCluster(NewTestingT(), Kuma4, Silent)
	E2EDeferCleanup(UniZone1.DismissCluster) // clean up any containers if needed
	cp, err = NewUniversalControlPlane(
		UniZone1.GetTesting(),
		core.Zone,
		UniZone1.Name(),
		UniZone1.Verbose(),
		state.UniZone1.KumaCp,
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
	)
	Expect(err).ToNot(HaveOccurred())
	UniZone2.SetCp(cp)
	Expect(UniZone2.AddNetworking(state.UniZone2.ZoneEgress, Config.ZoneEgressApp)).To(Succeed())
	Expect(UniZone2.AddNetworking(state.UniZone2.ZoneIngress, Config.ZoneIngressApp)).To(Succeed())
}

func getKumaDeploymentOptions(config framework.ControlPlaneConfig, defaultOptions ...framework.KumaDeploymentOption) []framework.KumaDeploymentOption {
	kumaOptions := append([]framework.KumaDeploymentOption{}, defaultOptions...)
	for key, value := range config.Envs {
		kumaOptions = append(kumaOptions, framework.WithEnv(key, value))
	}
	if config.AdditionalYamlConfig != "" {
		kumaOptions = append(kumaOptions, framework.WithYamlConfig(config.AdditionalYamlConfig))
	}
	return kumaOptions
}
