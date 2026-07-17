package multizone_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/v3/pkg/test"
	"github.com/kumahq/kuma/v3/test/e2e_env/multizone/cni"
	"github.com/kumahq/kuma/v3/test/e2e_env/multizone/connectivity"
	"github.com/kumahq/kuma/v3/test/e2e_env/multizone/defaults"
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
	_ = Describe("External Service locality aware", Label("job-0"), localityawarelb.ExternalServicesWithLocalityAwareLb, Ordered)
	_ = Describe("Healthcheck", Label("job-3"), healthcheck.ApplicationOnUniversalClientOnK8s, Ordered)
	_ = Describe("Inspect", Label("job-3"), inspect.Inspect, Ordered)
	_ = Describe("MeshHTTPRoute", Label("job-1"), meshhttproute.Test, Ordered)
	_ = Describe("MeshHTTPRoute MeshService", Label("job-1"), meshhttproute.MeshService, Ordered)
	_ = Describe("MeshTCPRoute", Label("job-3"), meshtcproute.Test, Ordered)
	_ = Describe("InboundPassthrough", Label("job-3"), inbound_communication.InboundPassthrough, Ordered)
	_ = Describe("InboundPassthroughDisabled", Label("job-3"), inbound_communication.InboundPassthroughDisabled, Ordered)
	_ = Describe("Connectivity", Label("job-1"), connectivity.Connectivity, Ordered)
	_ = Describe("Connectivity Gateway IPV6 CNI V2", Label("job-1"), connectivity.GatewayIPV6CNIV2, Ordered)
	_ = Describe("Sync", Label("job-1"), multizone_sync.Sync, Ordered)
	_ = Describe("MeshTrafficPermission", Label("job-3"), meshtrafficpermission.MeshTrafficPermission, Ordered)
	_ = Describe("MeshAccessLog on Zone Ingress", Label("job-3"), meshaccesslog.ZoneIngress, Ordered)
	_ = Describe("Zone Disable", Label("job-3"), zonedisable.ZoneDisable, Ordered)
	_ = Describe("Ownership", Label("job-2"), ownership.MultizoneUniversal, Ordered)
	_ = Describe("Resilience", Label("job-2"), resilience.ResilienceMultizoneUniversal, Ordered)
	_ = Describe("Resilience Postgres", Label("job-2"), resilience.ResilienceMultizoneUniversalPostgres, Ordered)
	_ = Describe("MeshTimeout", Label("job-0"), meshtimeout.MeshTimeout, Ordered)
	_ = Describe("LocalityAwareness with MeshLoadBalancingStrategy", Label("job-0"), localityawarelb.LocalityAwarenessWithMeshLoadBalancingStrategy, Ordered)
	_ = Describe("Advanced LocalityAwareness with MeshLoadBalancingStrategy", Label("job-0"), localityawarelb.LocalityAwareLB, Ordered)
	_ = Describe("Advanced LocalityAwareness with MeshLoadBalancingStrategy and Enabled Egress", Label("job-0"), localityawarelb.LocalityAwareLBEgress, Ordered)
	_ = Describe("Defaults", Label("job-3"), defaults.Defaults, Ordered)
	_ = Describe("MeshService Sync", Label("job-1"), meshservice.Sync, Ordered)
	_ = Describe("MeshService Connectivity", Label("job-1"), meshservice.Connectivity, Ordered)
	_ = Describe("MeshService Migration", Label("job-1"), meshservice.Migration, Ordered)
	_ = Describe("Targeting real MeshService in policies", Label("job-1"), meshservice.MeshServiceTargeting, Ordered)
	_ = Describe("MeshMultiZoneService Connectivity", Label("job-2"), meshmultizoneservice.Connectivity, Ordered)
	_ = Describe("MeshMultiZoneService MeshLbStrategy", Label("job-0"), localityawarelb.MeshMzService, Ordered)
	_ = Describe("Available services", Label("job-1"), connectivity.AvailableServices, Ordered)
	_ = Describe("ReachableBackends", Label("job-1"), reachablebackends.ReachableBackends, Ordered)
	_ = Describe("Producer Policy Flow", Label("job-2"), producer.ProducerPolicyFlow, Ordered)
	_ = Describe("MeshServiceReachableBackends", Label("job-1"), reachablebackends.MeshServicesWithReachableBackendsOption, Ordered)
	_ = Describe("MeshTLS", Label("job-3"), meshtls.MeshTLS, Ordered)
	_ = Describe("MeshIdentity", Label("job-0"), meshidentity.Identity, Ordered)
	_ = Describe("Unified Resource Naming", Label("job-3"), unifiednaming.UnifiedNaming, Ordered)
	_ = Describe("MeshIdentity Migration", Label("job-0"), meshidentity.Migration, Ordered)
	_ = Describe("CNI Configuration", Label("job-3"), Label("kind-not-supported"), cni.ExcludeOutboundPort, Ordered)
	_ = Describe("MeshProxy", Label("job-2"), meshproxy.Connectivity, Ordered)
	_ = Describe("MeshProxy Migration", Label("job-2"), meshproxy.Migration, Ordered)
	_ = Describe("MeshMetric on Zone Proxy", Label("job-3"), meshmetric.ZoneProxy, Ordered)
	_ = Describe("Resource Label Validation", Label("job-3"), Label("golden-files-e2e"), validation.ResourceValidation, Ordered)
)
