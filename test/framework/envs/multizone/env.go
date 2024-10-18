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

<<<<<<< HEAD
=======
func setupKubeZone(wg *sync.WaitGroup, clusterName string, extraOptions ...framework.KumaDeploymentOption) *K8sCluster {
	wg.Add(1)
	options := []framework.KumaDeploymentOption{
		WithEnv("KUMA_MULTIZONE_ZONE_KDS_NACK_BACKOFF", "1s"),
		WithIngress(),
		WithIngressEnvoyAdminTunnel(),
		WithEgress(),
		WithEgressEnvoyAdminTunnel(),
		WithGlobalAddress(Global.GetKuma().GetKDSServerAddress()),
		// Occasionally CP will lose a leader in the E2E test just because of this deadline,
		// which does not make sense in such controlled environment (one k3d node, one instance of the CP).
		// 100s and 80s are values that we also use in mesh-perf when we put a lot of pressure on the CP.
		framework.WithEnv("KUMA_RUNTIME_KUBERNETES_LEADER_ELECTION_LEASE_DURATION", "100s"),
		framework.WithEnv("KUMA_RUNTIME_KUBERNETES_LEADER_ELECTION_RENEW_DEADLINE", "80s"),
	}
	options = append(options, extraOptions...)
	zone := NewK8sCluster(NewTestingT(), clusterName, Verbose)
	go func() {
		defer ginkgo.GinkgoRecover()
		defer wg.Done()
		Expect(zone.Install(Kuma(core.Zone, options...))).To(Succeed())
	}()
	return zone
}

func setupUniZone(wg *sync.WaitGroup, clusterName string, extraOptions ...framework.KumaDeploymentOption) *UniversalCluster {
	wg.Add(1)
	options := append(
		[]framework.KumaDeploymentOption{
			WithGlobalAddress(Global.GetKuma().GetKDSServerAddress()),
			WithEgressEnvoyAdminTunnel(),
			WithIngressEnvoyAdminTunnel(),
			WithEnv("KUMA_XDS_DATAPLANE_DEREGISTRATION_DELAY", "0s"), // we have only 1 Kuma CP instance so there is no risk setting this to 0
			WithEnv("KUMA_MULTIZONE_ZONE_KDS_NACK_BACKOFF", "1s"),
		},
		extraOptions...,
	)
	zone := NewUniversalCluster(NewTestingT(), clusterName, Silent)
	go func() {
		defer ginkgo.GinkgoRecover()
		defer wg.Done()
		err := NewClusterSetup().
			Install(Kuma(core.Zone, options...)).
			Install(IngressUniversal(Global.GetKuma().GenerateZoneIngressToken)).
			Install(EgressUniversal(Global.GetKuma().GenerateZoneEgressToken, WithConcurrency(1))).
			Setup(zone)
		Expect(err).ToNot(HaveOccurred())
	}()
	return zone
}

>>>>>>> bb904d04e (test(e2e): increase leader election lease and renew duration (#11796))
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
	wg.Add(4)

	kubeZone1Options := append(
		[]framework.KumaDeploymentOption{
			WithEnv("KUMA_STORE_UNSAFE_DELETE", "true"),
			WithEnv("KUMA_MULTIZONE_ZONE_KDS_NACK_BACKOFF", "1s"),
			WithIngress(),
			WithIngressEnvoyAdminTunnel(),
			WithEgress(),
			WithEgressEnvoyAdminTunnel(),
			WithGlobalAddress(Global.GetKuma().GetKDSServerAddress()),
		},
		framework.KumaDeploymentOptionsFromConfig(framework.Config.KumaCpConfig.Multizone.KubeZone1)...,
	)
	if Config.IPV6 {
		// if the underneath clusters support IPv6, we'll configure kuma-1 with waitForDataplaneReady feature and
		// envoy admin binding to ::1 address
		kubeZone1Options = append(kubeZone1Options,
			WithEnv("KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_WAIT_FOR_DATAPLANE_READY", "true"),
			WithEnv("KUMA_BOOTSTRAP_SERVER_PARAMS_ADMIN_ADDRESS", "::1"),
		)
	}
	KubeZone1 = NewK8sCluster(NewTestingT(), Kuma1, Verbose)
	go func() {
		defer GinkgoRecover()
		defer wg.Done()
		Expect(KubeZone1.Install(Kuma(core.Zone, kubeZone1Options...))).To(Succeed())
	}()

	kubeZone2Options := append(
		[]framework.KumaDeploymentOption{
			WithEnv("KUMA_MULTIZONE_ZONE_KDS_NACK_BACKOFF", "1s"),
			WithIngress(),
			WithIngressEnvoyAdminTunnel(),
			WithEgress(),
			WithEgressEnvoyAdminTunnel(),
			WithGlobalAddress(Global.GetKuma().GetKDSServerAddress()),
			WithCNI(),
		},
		framework.KumaDeploymentOptionsFromConfig(framework.Config.KumaCpConfig.Multizone.KubeZone2)...,
	)
	KubeZone2 = NewK8sCluster(NewTestingT(), Kuma2, Verbose)
	go func() {
		defer GinkgoRecover()
		defer wg.Done()
		Expect(KubeZone2.Install(Kuma(core.Zone, kubeZone2Options...))).To(Succeed())
	}()

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
			Install(IngressUniversal(Global.GetKuma().GenerateZoneIngressLegacyToken)).
			Install(EgressUniversal(Global.GetKuma().GenerateZoneEgressToken, WithConcurrency(1))).
			Setup(UniZone1)
		Expect(err).ToNot(HaveOccurred())
	}()

	UniZone2 = NewUniversalCluster(NewTestingT(), Kuma5, Silent)
	uniZone2Options := append(
		[]framework.KumaDeploymentOption{
			WithGlobalAddress(Global.GetKuma().GetKDSServerAddress()),
			WithEgressEnvoyAdminTunnel(),
			WithIngressEnvoyAdminTunnel(),
			WithEnv("KUMA_XDS_DATAPLANE_DEREGISTRATION_DELAY", "0s"), // we have only 1 Kuma CP instance so there is no risk setting this to 0
			WithEnv("KUMA_MULTIZONE_ZONE_KDS_NACK_BACKOFF", "1s"),
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

	KubeZone1 = NewK8sCluster(NewTestingT(), Kuma1, Verbose)
	kubeCp := NewK8sControlPlane(
		KubeZone1.GetTesting(),
		core.Zone,
		KubeZone1.Name(),
		KubeZone1.GetKubectlOptions().ConfigPath,
		KubeZone1,
		KubeZone1.Verbose(),
		1,
		nil,
	)
	Expect(kubeCp.FinalizeAddWithPortFwd(state.KubeZone1.KumaCp, state.KubeZone1.KumaCp)).To(Succeed())
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
		nil, // headers were not configured in setup
	)
	Expect(kubeCp.FinalizeAddWithPortFwd(state.KubeZone2.KumaCp, state.KubeZone2.MADS)).To(Succeed())
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
