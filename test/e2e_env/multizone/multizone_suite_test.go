package auth_test

import (
	"encoding/json"
	"sync"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e_env/multizone/env"
	"github.com/kumahq/kuma/test/e2e_env/multizone/healthcheck"
	"github.com/kumahq/kuma/test/e2e_env/multizone/inspect"
	"github.com/kumahq/kuma/test/e2e_env/multizone/localityawarelb"
	"github.com/kumahq/kuma/test/e2e_env/multizone/trafficpermission"
	"github.com/kumahq/kuma/test/e2e_env/multizone/trafficroute"
	. "github.com/kumahq/kuma/test/framework"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Universal Suite")
}

type State struct {
	Global    UniversalCPNetworking
	UniZone1  UniversalCPNetworking
	UniZone2  UniversalCPNetworking
	KubeZone1 PortFwd
	KubeZone2 PortFwd
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
				WithIngress(),
				WithIngressEnvoyAdminTunnel(),
				WithEgress(),
				WithEgressEnvoyAdminTunnel(),
				WithGlobalAddress(env.Global.GetKuma().GetKDSServerAddress()),
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
				Install(IngressUniversal(env.Global.GetKuma().GenerateZoneIngressToken)).
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
				)).
				Install(IngressUniversal(env.Global.GetKuma().GenerateZoneIngressToken)).
				Install(EgressUniversal(env.Global.GetKuma().GenerateZoneEgressToken)).
				Setup(env.UniZone2)
			Expect(err).ToNot(HaveOccurred())
			wg.Done()
		}()
		wg.Wait()

		state := State{
			Global:    env.Global.GetKuma().(*UniversalControlPlane).Networking(),
			UniZone1:  env.UniZone1.GetKuma().(*UniversalControlPlane).Networking(),
			UniZone2:  env.UniZone2.GetKuma().(*UniversalControlPlane).Networking(),
			KubeZone1: env.KubeZone1.GetKuma().(*K8sControlPlane).PortFwd(),
			KubeZone2: env.KubeZone2.GetKuma().(*K8sControlPlane).PortFwd(),
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
			state.Global,
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
		Expect(kubeCp.FinalizeAddWithPortFwd(state.KubeZone1)).To(Succeed())
		env.KubeZone1.SetCP(kubeCp)

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
		Expect(kubeCp.FinalizeAddWithPortFwd(state.KubeZone2)).To(Succeed())
		env.KubeZone2.SetCP(kubeCp)

		env.UniZone1 = NewUniversalCluster(NewTestingT(), Kuma4, Silent)
		E2EDeferCleanup(env.UniZone1.DismissCluster) // clean up any containers if needed
		cp, err = NewUniversalControlPlane(
			env.UniZone1.GetTesting(),
			core.Zone,
			env.UniZone1.Name(),
			env.UniZone1.Verbose(),
			state.UniZone1,
		)
		Expect(err).ToNot(HaveOccurred())
		env.UniZone1.SetCp(cp)

		env.UniZone2 = NewUniversalCluster(NewTestingT(), Kuma5, Silent)
		E2EDeferCleanup(env.UniZone2.DismissCluster) // clean up any containers if needed
		cp, err = NewUniversalControlPlane(
			env.UniZone2.GetTesting(),
			core.Zone,
			env.UniZone2.Name(),
			env.UniZone2.Verbose(),
			state.UniZone2,
		)
		Expect(err).ToNot(HaveOccurred())
		env.UniZone2.SetCp(cp)
	},
)

var _ = Describe("TrafficPermission", trafficpermission.TrafficPermission, Ordered)
var _ = Describe("TrafficRoute", trafficroute.TrafficRoute, Ordered)
var _ = Describe("Healtcheck", healthcheck.ApplicationOnUniversalClientOnK8s, Ordered)
var _ = Describe("Inspect", inspect.Inspect, Ordered)
var _ = Describe("External Service locality aware", localityawarelb.ExternalServicesWithLocalityAwareLb, Ordered)
