package auth_test

import (
	"encoding/json"
	"sync"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e_env/multizone/connectivity"
	"github.com/kumahq/kuma/test/e2e_env/multizone/env"
	"github.com/kumahq/kuma/test/e2e_env/multizone/gateway"
	"github.com/kumahq/kuma/test/e2e_env/multizone/healthcheck"
	"github.com/kumahq/kuma/test/e2e_env/multizone/inbound_communication"
	"github.com/kumahq/kuma/test/e2e_env/multizone/inspect"
	"github.com/kumahq/kuma/test/e2e_env/multizone/localityawarelb"
	"github.com/kumahq/kuma/test/e2e_env/multizone/meshtrafficpermission"
	multizone_sync "github.com/kumahq/kuma/test/e2e_env/multizone/sync"
	"github.com/kumahq/kuma/test/e2e_env/multizone/trafficpermission"
	"github.com/kumahq/kuma/test/e2e_env/multizone/trafficroute"
	"github.com/kumahq/kuma/test/e2e_env/multizone/zoneegress"
	. "github.com/kumahq/kuma/test/framework"
)

func TestE2E(t *testing.T) {
	test.RunE2ESpecs(t, "E2E Multizone Suite")
}

type State struct {
	Global    UniversalNetworkingState
	UniZone1  UniversalNetworkingState
	UniZone2  UniversalNetworkingState
	KubeZone1 K8sNetworkingState
	KubeZone2 K8sNetworkingState
}

var _ = SynchronizedBeforeSuite(
	func() []byte {
		env.Global = NewUniversalCluster(NewTestingT(), Kuma3, Silent)
		E2EDeferCleanup(env.Global.DismissCluster) // clean up any containers if needed
		Expect(env.Global.Install(Kuma(core.Global,
			WithEnv("KUMA_STORE_UNSAFE_DELETE", "true"),
		))).To(Succeed())

		wg := sync.WaitGroup{}
		wg.Add(4)

		env.KubeZone1 = NewK8sCluster(NewTestingT(), Kuma1, Verbose)
		go func() {
			defer GinkgoRecover()
			Expect(env.KubeZone1.Install(Kuma(core.Zone,
				WithEnv("KUMA_STORE_UNSAFE_DELETE", "true"),
				WithIngress(),
				WithIngressEnvoyAdminTunnel(),
				WithEgress(),
				WithEgressEnvoyAdminTunnel(),
				WithGlobalAddress(env.Global.GetKuma().GetKDSServerAddress()),
			))).To(Succeed())
			wg.Done()
		}()

		env.KubeZone2 = NewK8sCluster(NewTestingT(), Kuma2, Verbose)
		go func() {
			defer GinkgoRecover()
			Expect(env.KubeZone2.Install(Kuma(core.Zone,
				WithEnv("KUMA_STORE_UNSAFE_DELETE", "true"),
				WithEnv("KUMA_DEFAULTS_ENABLE_LOCALHOST_INBOUND_CLUSTERS", "true"),
				WithIngress(),
				WithIngressEnvoyAdminTunnel(),
				WithEgress(),
				WithEgressEnvoyAdminTunnel(),
				WithGlobalAddress(env.Global.GetKuma().GetKDSServerAddress()),
				WithExperimentalCNI(),
			))).To(Succeed())
			wg.Done()
		}()

		env.UniZone1 = NewUniversalCluster(NewTestingT(), Kuma4, Silent)
		E2EDeferCleanup(env.UniZone1.DismissCluster) // clean up any containers if needed
		go func() {
			defer GinkgoRecover()
			err := NewClusterSetup().
				Install(Kuma(core.Zone,
					WithGlobalAddress(env.Global.GetKuma().GetKDSServerAddress()),
					WithEnv("KUMA_STORE_UNSAFE_DELETE", "true"),
					WithEgressEnvoyAdminTunnel(),
					WithIngressEnvoyAdminTunnel(),
				)).
				Install(IngressUniversal(env.Global.GetKuma().GenerateZoneIngressLegacyToken)).
				Install(EgressUniversal(env.Global.GetKuma().GenerateZoneEgressToken)).
				Setup(env.UniZone1)
			Expect(err).ToNot(HaveOccurred())
			wg.Done()
		}()

		env.UniZone2 = NewUniversalCluster(NewTestingT(), Kuma5, Silent)
		E2EDeferCleanup(env.UniZone2.DismissCluster) // clean up any containers if needed
		go func() {
			defer GinkgoRecover()
			err := NewClusterSetup().
				Install(Kuma(core.Zone,
					WithGlobalAddress(env.Global.GetKuma().GetKDSServerAddress()),
					WithEnv("KUMA_STORE_UNSAFE_DELETE", "true"),
					WithEnv("KUMA_DEFAULTS_ENABLE_LOCALHOST_INBOUND_CLUSTERS", "true"),
					WithEgressEnvoyAdminTunnel(),
					WithIngressEnvoyAdminTunnel(),
				)).
				Install(IngressUniversal(env.Global.GetKuma().GenerateZoneIngressToken)).
				Install(EgressUniversal(env.Global.GetKuma().GenerateZoneEgressToken)).
				Setup(env.UniZone2)
			Expect(err).ToNot(HaveOccurred())
			wg.Done()
		}()
		wg.Wait()

		state := State{
			Global: UniversalNetworkingState{
				ZoneEgress:  env.Global.GetZoneEgressNetworking(),
				ZoneIngress: env.Global.GetZoneIngressNetworking(),
				KumaCp:      env.Global.GetKuma().(*UniversalControlPlane).Networking(),
			},
			UniZone1: UniversalNetworkingState{
				ZoneEgress:  env.UniZone1.GetZoneEgressNetworking(),
				ZoneIngress: env.UniZone1.GetZoneIngressNetworking(),
				KumaCp:      env.UniZone1.GetKuma().(*UniversalControlPlane).Networking(),
			},
			UniZone2: UniversalNetworkingState{
				ZoneEgress:  env.UniZone2.GetZoneEgressNetworking(),
				ZoneIngress: env.UniZone2.GetZoneIngressNetworking(),
				KumaCp:      env.UniZone2.GetKuma().(*UniversalControlPlane).Networking(),
			},
			KubeZone1: K8sNetworkingState{
				ZoneEgress:  env.KubeZone1.GetZoneEgressPortForward(),
				ZoneIngress: env.KubeZone1.GetZoneIngressPortForward(),
				KumaCp:      env.KubeZone1.GetKuma().(*K8sControlPlane).PortFwd(),
			},
			KubeZone2: K8sNetworkingState{
				ZoneEgress:  env.KubeZone2.GetZoneEgressPortForward(),
				ZoneIngress: env.KubeZone2.GetZoneIngressPortForward(),
				KumaCp:      env.KubeZone2.GetKuma().(*K8sControlPlane).PortFwd(),
			},
		}
		bytes, err := json.Marshal(state)
		Expect(err).ToNot(HaveOccurred())
		return bytes
	},
	func(bytes []byte) {
		if env.Global != nil {
			return // cluster was already initiated with first function
		}
		state := State{}
		Expect(json.Unmarshal(bytes, &state)).To(Succeed())

		env.Global = NewUniversalCluster(NewTestingT(), Kuma3, Silent)
		E2EDeferCleanup(env.Global.DismissCluster) // clean up any containers if needed
		cp, err := NewUniversalControlPlane(
			env.Global.GetTesting(),
			core.Global,
			env.Global.Name(),
			env.Global.Verbose(),
			state.Global.KumaCp,
		)
		Expect(err).ToNot(HaveOccurred())
		env.Global.SetCp(cp)

		env.KubeZone1 = NewK8sCluster(NewTestingT(), Kuma1, Verbose)
		kubeCp := NewK8sControlPlane(
			env.KubeZone1.GetTesting(),
			core.Zone,
			env.KubeZone1.Name(),
			env.KubeZone1.GetKubectlOptions().ConfigPath,
			env.KubeZone1,
			env.KubeZone1.Verbose(),
			1,
		)
		Expect(kubeCp.FinalizeAddWithPortFwd(state.KubeZone1.KumaCp)).To(Succeed())
		env.KubeZone1.SetCP(kubeCp)
		Expect(env.KubeZone1.AddPortForward(state.KubeZone1.ZoneEgress, Config.ZoneEgressApp)).To(Succeed())
		Expect(env.KubeZone1.AddPortForward(state.KubeZone1.ZoneIngress, Config.ZoneIngressApp)).To(Succeed())

		env.KubeZone2 = NewK8sCluster(NewTestingT(), Kuma2, Verbose)
		kubeCp = NewK8sControlPlane(
			env.KubeZone2.GetTesting(),
			core.Zone,
			env.KubeZone2.Name(),
			env.KubeZone2.GetKubectlOptions().ConfigPath,
			env.KubeZone2,
			env.KubeZone2.Verbose(),
			1,
		)
		Expect(kubeCp.FinalizeAddWithPortFwd(state.KubeZone2.KumaCp)).To(Succeed())
		env.KubeZone2.SetCP(kubeCp)
		Expect(env.KubeZone2.AddPortForward(state.KubeZone2.ZoneEgress, Config.ZoneEgressApp)).To(Succeed())
		Expect(env.KubeZone2.AddPortForward(state.KubeZone2.ZoneIngress, Config.ZoneIngressApp)).To(Succeed())

		env.UniZone1 = NewUniversalCluster(NewTestingT(), Kuma4, Silent)
		E2EDeferCleanup(env.UniZone1.DismissCluster) // clean up any containers if needed
		cp, err = NewUniversalControlPlane(
			env.UniZone1.GetTesting(),
			core.Zone,
			env.UniZone1.Name(),
			env.UniZone1.Verbose(),
			state.UniZone1.KumaCp,
		)
		Expect(err).ToNot(HaveOccurred())
		env.UniZone1.SetCp(cp)
		Expect(env.UniZone1.AddNetworking(state.UniZone1.ZoneEgress, Config.ZoneEgressApp)).To(Succeed())
		Expect(env.UniZone1.AddNetworking(state.UniZone1.ZoneIngress, Config.ZoneIngressApp)).To(Succeed())

		env.UniZone2 = NewUniversalCluster(NewTestingT(), Kuma5, Silent)
		E2EDeferCleanup(env.UniZone2.DismissCluster) // clean up any containers if needed
		cp, err = NewUniversalControlPlane(
			env.UniZone2.GetTesting(),
			core.Zone,
			env.UniZone2.Name(),
			env.UniZone2.Verbose(),
			state.UniZone2.KumaCp,
		)
		Expect(err).ToNot(HaveOccurred())
		env.UniZone2.SetCp(cp)
		Expect(env.UniZone2.AddNetworking(state.UniZone2.ZoneEgress, Config.ZoneEgressApp)).To(Succeed())
		Expect(env.UniZone2.AddNetworking(state.UniZone2.ZoneIngress, Config.ZoneIngressApp)).To(Succeed())
	},
)

var _ = Describe("Gateway", gateway.GatewayHybrid, Ordered)
var _ = Describe("Cross-mesh Gateways", gateway.CrossMeshGatewayOnMultizone, Ordered)
var _ = Describe("External Service locality aware", localityawarelb.ExternalServicesWithLocalityAwareLb, Ordered)
var _ = Describe("Healthcheck", healthcheck.ApplicationOnUniversalClientOnK8s, Ordered)
var _ = Describe("Inspect", inspect.Inspect, Ordered)
var _ = Describe("TrafficPermission", trafficpermission.TrafficPermission, Ordered)
var _ = Describe("TrafficRoute", trafficroute.TrafficRoute, Ordered)
var _ = Describe("InboundPassthrough", inbound_communication.InboundPassthrough, Ordered)
var _ = Describe("InboundPassthroughDisabled", inbound_communication.InboundPassthroughDisabled, Ordered)
var _ = Describe("ZoneEgress Internal Services", zoneegress.InternalServices, Ordered)
var _ = Describe("Connectivity", connectivity.Connectivity, Ordered)
var _ = Describe("Sync", multizone_sync.Sync, Ordered)
var _ = Describe("MeshTrafficPermission", meshtrafficpermission.MeshTrafficPermission, Ordered)
