package multizone_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/v3/pkg/test"
	"github.com/kumahq/kuma/v3/test/e2e_env/multizone/cni"
	"github.com/kumahq/kuma/v3/test/e2e_env/multizone/connectivity"
	"github.com/kumahq/kuma/v3/test/e2e_env/multizone/defaults"
	"github.com/kumahq/kuma/v3/test/e2e_env/multizone/externalservices"
	"github.com/kumahq/kuma/v3/test/e2e_env/multizone/gateway"
	"github.com/kumahq/kuma/v3/test/e2e_env/multizone/healthcheck"
	"github.com/kumahq/kuma/v3/test/e2e_env/multizone/inbound_communication"
	"github.com/kumahq/kuma/v3/test/e2e_env/multizone/inspect"
	"github.com/kumahq/kuma/v3/test/e2e_env/multizone/localityawarelb"
	"github.com/kumahq/kuma/v3/test/e2e_env/multizone/meshaccesslog"
	"github.com/kumahq/kuma/v3/test/e2e_env/multizone/meshhttproute"
	"github.com/kumahq/kuma/v3/test/e2e_env/multizone/meshidentity"
	"github.com/kumahq/kuma/v3/test/e2e_env/multizone/meshmetric"
	"github.com/kumahq/kuma/v3/test/e2e_env/multizone/meshmultizoneservice"
	"github.com/kumahq/kuma/v3/test/e2e_env/multizone/meshproxy"
	"github.com/kumahq/kuma/v3/test/e2e_env/multizone/meshservice"
	"github.com/kumahq/kuma/v3/test/e2e_env/multizone/meshtcproute"
	"github.com/kumahq/kuma/v3/test/e2e_env/multizone/meshtimeout"
	"github.com/kumahq/kuma/v3/test/e2e_env/multizone/meshtls"
	"github.com/kumahq/kuma/v3/test/e2e_env/multizone/meshtrafficpermission"
	"github.com/kumahq/kuma/v3/test/e2e_env/multizone/ownership"
	"github.com/kumahq/kuma/v3/test/e2e_env/multizone/producer"
	"github.com/kumahq/kuma/v3/test/e2e_env/multizone/reachablebackends"
	"github.com/kumahq/kuma/v3/test/e2e_env/multizone/resilience"
	multizone_sync "github.com/kumahq/kuma/v3/test/e2e_env/multizone/sync"
	"github.com/kumahq/kuma/v3/test/e2e_env/multizone/unifiednaming"
	"github.com/kumahq/kuma/v3/test/e2e_env/multizone/validation"
	"github.com/kumahq/kuma/v3/test/e2e_env/multizone/zonedisable"
	"github.com/kumahq/kuma/v3/test/e2e_env/multizone/zoneegress"
	. "github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/envs/multizone"
)

func TestE2E(t *testing.T) {
	test.RunE2ESpecs(t, "E2E Multizone Suite")
}

var (
	_ = E2ESynchronizedBeforeSuite(multizone.SetupAndGetState, multizone.RestoreState)
	_ = SynchronizedAfterSuite(func() {}, multizone.SynchronizedAfterSuite)
	_ = ReportAfterSuite("multizone after suite", multizone.AfterSuite)
	// Opt-in (KUMA3_PREFLIGHT_BIN + KUMA3_PREFLIGHT_DIR): snapshot the GLOBAL CP after each
	// spec — one audit of the global covers every zone (resources sync over KDS). No-op otherwise.
	_ = AfterEach(func() {
		CapturePreflightCluster(CurrentSpecReport().FullText(), multizone.Global)
	})
)

var (
	_ = Describe("Gateway", gateway.GatewayHybrid, Ordered)
	_ = Describe("Cross-mesh Gateways", gateway.CrossMeshGatewayOnMultizone, Ordered)
	_ = Describe("External Service locality aware", localityawarelb.ExternalServicesWithLocalityAwareLb, Ordered)
	_ = Describe("Healthcheck", healthcheck.ApplicationOnUniversalClientOnK8s, Ordered)
	_ = Describe("Inspect", inspect.Inspect, Ordered)
	_ = Describe("MeshHTTPRoute", meshhttproute.Test, Ordered)
	_ = Describe("MeshHTTPRoute MeshService", meshhttproute.MeshService, Ordered)
	_ = Describe("MeshTCPRoute", meshtcproute.Test, Ordered)
	_ = Describe("InboundPassthrough", inbound_communication.InboundPassthrough, Ordered)
	_ = Describe("InboundPassthroughDisabled", inbound_communication.InboundPassthroughDisabled, Ordered)
	_ = Describe("ZoneEgress Internal Services", zoneegress.InternalServices, Ordered)
	_ = Describe("Connectivity", connectivity.Connectivity, Ordered)
	_ = Describe("Connectivity Gateway IPV6 CNI V2", connectivity.GatewayIPV6CNIV2, Ordered)
	_ = Describe("Sync", multizone_sync.Sync, Ordered)
	_ = Describe("MeshTrafficPermission", meshtrafficpermission.MeshTrafficPermission, Ordered)
	_ = Describe("MeshAccessLog on Zone Ingress", meshaccesslog.ZoneIngress, Ordered)
	_ = Describe("Zone Disable", zonedisable.ZoneDisable, Ordered)
	_ = Describe("External Services", externalservices.ExternalServicesOnMultizoneUniversal, Ordered)
	_ = Describe("Ownership", ownership.MultizoneUniversal, Ordered)
	_ = Describe("Resilience", resilience.ResilienceMultizoneUniversal, Ordered)
	_ = Describe("Resilience Postgres", resilience.ResilienceMultizoneUniversalPostgres, Ordered)
	_ = Describe("MeshTimeout", meshtimeout.MeshTimeout, Ordered)
	_ = Describe("LocalityAwareness with MeshLoadBalancingStrategy", localityawarelb.LocalityAwarenessWithMeshLoadBalancingStrategy, Ordered)
	_ = Describe("Advanced LocalityAwareness with MeshLoadBalancingStrategy", localityawarelb.LocalityAwareLB, Ordered)
	_ = Describe("Advanced LocalityAwareness with MeshLoadBalancingStrategy with Gateway", localityawarelb.LocalityAwareLBGateway, Ordered)
	_ = Describe("Advanced LocalityAwareness with MeshLoadBalancingStrategy and Enabled Egress", localityawarelb.LocalityAwareLBEgress, Ordered)
	_ = Describe("Defaults", defaults.Defaults, Ordered)
	_ = Describe("MeshService Sync", meshservice.Sync, Ordered)
	_ = Describe("MeshService Connectivity", meshservice.Connectivity, Ordered)
	_ = Describe("MeshService Migration", meshservice.Migration, Ordered)
	_ = Describe("Targeting real MeshService in policies", meshservice.MeshServiceTargeting, Ordered)
	_ = Describe("MeshMultiZoneService Connectivity", meshmultizoneservice.Connectivity, Ordered)
	_ = Describe("MeshMultiZoneService MeshLbStrategy", localityawarelb.MeshMzService, Ordered)
	_ = Describe("Available services", connectivity.AvailableServices, Ordered)
	_ = Describe("ReachableBackends", reachablebackends.ReachableBackends, Ordered)
	_ = Describe("Producer Policy Flow", producer.ProducerPolicyFlow, Ordered)
	_ = Describe("MeshServiceReachableBackends", reachablebackends.MeshServicesWithReachableBackendsOption, Ordered)
	_ = Describe("MeshTLS", meshtls.MeshTLS, Ordered)
	_ = Describe("MeshIdentity", meshidentity.Identity, Ordered)
	_ = Describe("Unified Resource Naming", unifiednaming.UnifiedNaming, Ordered)
	_ = Describe("MeshIdentity Migration", meshidentity.Migration, Ordered)
	_ = Describe("CNI Configuration", Label("kind-not-supported"), cni.ExcludeOutboundPort, Ordered)
	_ = Describe("MeshProxy", meshproxy.Connectivity, Ordered)
	_ = Describe("MeshProxy Migration", meshproxy.Migration, Ordered)
	_ = Describe("MeshMetric on Zone Proxy", meshmetric.ZoneProxy, Ordered)
	_ = Describe("Resource Label Validation", Label("golden-files-e2e"), validation.ResourceValidation, Ordered)
)
