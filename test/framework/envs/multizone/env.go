package multizone

import (
	"encoding/json"
	"sync"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/portforward"
	"github.com/kumahq/kuma/test/framework/report"
	kssh "github.com/kumahq/kuma/test/framework/ssh"
	"github.com/kumahq/kuma/test/framework/universal"
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
	Global    universal.NetworkingState
	UniZone1  universal.NetworkingState
	UniZone2  universal.NetworkingState
	KubeZone1 K8sNetworkingState
	KubeZone2 K8sNetworkingState
}

func SetupKubeZone(wg *sync.WaitGroup, clusterName string, extraOptions ...KumaDeploymentOption) *K8sCluster {
	wg.Add(1)
	options := []KumaDeploymentOption{
		WithEnv("KUMA_MULTIZONE_ZONE_KDS_NACK_BACKOFF", "1s"),
		WithIngress(),
		WithIngressEnvoyAdminTunnel(),
		WithEgress(),
		WithEgressEnvoyAdminTunnel(),
		WithGlobalAddress(Global.GetKuma().GetKDSServerAddress()),
		// Occasionally CP will lose a leader in the E2E test just because of this deadline,
		// which does not make sense in such controlled environment (one k3d node, one instance of the CP).
		// 100s and 80s are values that we also use in mesh-perf when we put a lot of pressure on the CP.
		WithEnv("KUMA_RUNTIME_KUBERNETES_LEADER_ELECTION_LEASE_DURATION", "100s"),
		WithEnv("KUMA_RUNTIME_KUBERNETES_LEADER_ELECTION_RENEW_DEADLINE", "80s"),
		WithEnv("KUMA_MULTIZONE_ZONE_KDS_LABELS_SKIP_PREFIXES", "argocd.argoproj.io"),
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

func SetupUniZone(wg *sync.WaitGroup, clusterName string, extraOptions ...KumaDeploymentOption) *UniversalCluster {
	return SetupRemoteUniZone(wg, clusterName, nil, extraOptions...)
}

func SetupRemoteUniZone(wg *sync.WaitGroup, clusterName string, remoteHost *kssh.Host, extraOptions ...KumaDeploymentOption) *UniversalCluster {
	wg.Add(1)
	options := append(
		[]KumaDeploymentOption{
			WithGlobalAddress(Global.GetKuma().GetKDSServerAddress()),
			WithEgressEnvoyAdminTunnel(),
			WithIngressEnvoyAdminTunnel(),
			WithEnv("KUMA_XDS_DATAPLANE_DEREGISTRATION_DELAY", "0s"), // we have only 1 Kuma CP instance so there is no risk setting this to 0
			WithEnv("KUMA_MULTIZONE_ZONE_KDS_NACK_BACKOFF", "1s"),
			WithEnv("KUMA_MULTIZONE_ZONE_KDS_LABELS_SKIP_PREFIXES", "argocd.argoproj.io"),
		},
		extraOptions...,
	)

	var zone *UniversalCluster
	if remoteHost == nil {
		zone = NewUniversalCluster(NewTestingT(), clusterName, Silent)
	} else {
		zone = NewRemoteUniversalCluster(NewTestingT(), clusterName, remoteHost, Silent)
	}

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

// SetupAndGetState to be used with Ginkgo SynchronizedBeforeSuite
func SetupAndGetState() []byte {
	Global = NewUniversalCluster(NewTestingT(), Kuma3, Silent)
	globalOptions := append(
		[]KumaDeploymentOption{
			WithEnv("KUMA_MULTIZONE_GLOBAL_KDS_NACK_BACKOFF", "1s"),
			WithEnv("KUMA_MULTIZONE_GLOBAL_KDS_LABELS_SKIP_PREFIXES", "argocd.argoproj.io"),
		},
		KumaDeploymentOptionsFromConfig(Config.KumaCpConfig.Multizone.Global)...)
	Expect(Global.Install(Kuma(core.Global, globalOptions...))).To(Succeed())

	wg := sync.WaitGroup{}

	kubeZone1Options := append(
		KumaDeploymentOptionsFromConfig(Config.KumaCpConfig.Multizone.KubeZone1),
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
	KubeZone1 = SetupKubeZone(&wg, Kuma1, kubeZone1Options...)

	kubeZone2Options := append(
		KumaDeploymentOptionsFromConfig(Config.KumaCpConfig.Multizone.KubeZone2),
		WithEnv("KUMA_EXPERIMENTAL_DELTA_XDS", "true"),
		WithMemoryLimit("512Mi"),
		WithCNI(),
	)
	KubeZone2 = SetupKubeZone(&wg, Kuma2, kubeZone2Options...)

	uniZone1Options := append(
		KumaDeploymentOptionsFromConfig(Config.KumaCpConfig.Multizone.UniZone1),
		WithEnv("KUMA_EXPERIMENTAL_DELTA_XDS", "true"),
	)
	UniZone1 = SetupUniZone(&wg, Kuma4, uniZone1Options...)

	vipCIDROverride := "251.0.0.0/8"
	if Config.IPV6 {
		vipCIDROverride = "fd00:fd11::/64"
	}
	uniZone2Options := append(
		KumaDeploymentOptionsFromConfig(Config.KumaCpConfig.Multizone.UniZone2),
		WithEnv("KUMA_IPAM_MESH_SERVICE_CIDR", vipCIDROverride), // just to see that the status is not synced around
	)
	UniZone2 = SetupUniZone(&wg, Kuma5, uniZone2Options...)

	wg.Wait()

	zeSpec := portforward.Spec{
		AppName:    Config.ZoneEgressApp,
		Namespace:  Config.KumaNamespace,
		RemotePort: 9901,
	}

	ziSpec := portforward.Spec{
		AppName:    Config.ZoneIngressApp,
		Namespace:  Config.KumaNamespace,
		RemotePort: 9901,
	}

	state := State{
		Global:   Global.GetUniversalNetworkingState(),
		UniZone1: UniZone1.GetUniversalNetworkingState(),
		UniZone2: UniZone2.GetUniversalNetworkingState(),
		KubeZone1: K8sNetworkingState{
			ZoneEgress:  KubeZone1.GetPortForward(zeSpec),
			ZoneIngress: KubeZone1.GetPortForward(ziSpec),
			KumaCp:      KubeZone1.GetKuma().(*K8sControlPlane).PortFwd(),
			MADS:        KubeZone1.GetKuma().(*K8sControlPlane).MadsPortFwd(),
		},
		KubeZone2: K8sNetworkingState{
			ZoneEgress:  KubeZone2.GetPortForward(zeSpec),
			ZoneIngress: KubeZone2.GetPortForward(ziSpec),
			KumaCp:      KubeZone2.GetKuma().(*K8sControlPlane).PortFwd(),
			MADS:        KubeZone2.GetKuma().(*K8sControlPlane).MadsPortFwd(),
		},
	}
	// govet complains of marshaling with mutex, we know what we're doing here
	bytes, err := json.Marshal(state) //nolint:govet
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
	zone.AddPortForward(networkingState.ZoneEgress, portforward.Spec{
		AppName:    Config.ZoneEgressApp,
		Namespace:  Config.KumaNamespace,
		RemotePort: 9901,
	})
	zone.AddPortForward(networkingState.ZoneIngress, portforward.Spec{
		AppName:    Config.ZoneIngressApp,
		Namespace:  Config.KumaNamespace,
		RemotePort: 9901,
	})
	return zone
}

func restoreUniZone(clusterName string, networkingState *universal.NetworkingState) *UniversalCluster {
	zone := NewUniversalCluster(NewTestingT(), clusterName, Silent)
	E2EDeferCleanup(zone.DismissCluster) // clean up any containers if needed
	cp, err := NewUniversalControlPlane(
		zone.GetTesting(),
		core.Zone,
		zone.Name(),
		zone.Verbose(),
		&networkingState.KumaCp,
		nil, // headers were not configured in setup
		true,
	)
	Expect(err).ToNot(HaveOccurred())
	zone.SetCp(cp)
	Expect(zone.AddNetworking(&networkingState.ZoneEgress, Config.ZoneEgressApp)).To(Succeed())
	Expect(zone.AddNetworking(&networkingState.ZoneIngress, Config.ZoneIngressApp)).To(Succeed())
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
		&state.Global.KumaCp,
		nil,
		true,
	)
	Expect(err).ToNot(HaveOccurred())
	Global.SetCp(cp)

	KubeZone1 = restoreKubeZone(Kuma1, &state.KubeZone1)
	KubeZone2 = restoreKubeZone(Kuma2, &state.KubeZone2)

	UniZone1 = restoreUniZone(Kuma4, &state.UniZone1)
	UniZone2 = restoreUniZone(Kuma5, &state.UniZone2)
}

func SynchronizedAfterSuite() {
	for _, cluster := range append(Zones(), Global) {
		ControlPlaneAssertions(cluster)
	}
	for _, cluster := range append(Zones(), Global) {
		DebugCPLogs(cluster)
	}
	Expect(Global.DismissCluster()).To(Succeed())
	Expect(UniZone1.DismissCluster()).To(Succeed())
	Expect(UniZone2.DismissCluster()).To(Succeed())
}

func AfterSuite(r ginkgo.Report) {
	report.DumpReport(r)
}
