package auth_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e_env/multizone/connectivity"
	"github.com/kumahq/kuma/test/e2e_env/multizone/externalservices"
	"github.com/kumahq/kuma/test/e2e_env/multizone/gateway"
	"github.com/kumahq/kuma/test/e2e_env/multizone/healthcheck"
	"github.com/kumahq/kuma/test/e2e_env/multizone/inbound_communication"
	"github.com/kumahq/kuma/test/e2e_env/multizone/inspect"
	"github.com/kumahq/kuma/test/e2e_env/multizone/localityawarelb"
	"github.com/kumahq/kuma/test/e2e_env/multizone/meshhttproute"
	"github.com/kumahq/kuma/test/e2e_env/multizone/meshtrafficpermission"
	"github.com/kumahq/kuma/test/e2e_env/multizone/ownership"
	"github.com/kumahq/kuma/test/e2e_env/multizone/resilience"
	multizone_sync "github.com/kumahq/kuma/test/e2e_env/multizone/sync"
	"github.com/kumahq/kuma/test/e2e_env/multizone/trafficpermission"
	"github.com/kumahq/kuma/test/e2e_env/multizone/trafficroute"
	"github.com/kumahq/kuma/test/e2e_env/multizone/zonedisable"
	"github.com/kumahq/kuma/test/e2e_env/multizone/zoneegress"
)

func TestE2E(t *testing.T) {
	test.RunE2ESpecs(t, "E2E Multizone Suite")
}

// var _ = SynchronizedBeforeSuite(multizone.SetupAndGetState, multizone.RestoreState)

var (
	_ = PDescribe("Gateway", gateway.GatewayHybrid, Ordered)
	_ = PDescribe("Cross-mesh Gateways", gateway.CrossMeshGatewayOnMultizone, Ordered)
	_ = PDescribe("External Service locality aware", localityawarelb.ExternalServicesWithLocalityAwareLb, Ordered)
	_ = PDescribe("Healthcheck", healthcheck.ApplicationOnUniversalClientOnK8s, Ordered)
	_ = PDescribe("Inspect", inspect.Inspect, Ordered)
	_ = PDescribe("TrafficPermission", trafficpermission.TrafficPermission, Ordered)
	_ = PDescribe("TrafficRoute", trafficroute.TrafficRoute, Ordered)
	_ = PDescribe("MeshHTTPRoute", meshhttproute.Test, Ordered)
	_ = PDescribe("InboundPassthrough", inbound_communication.InboundPassthrough, Ordered)
	_ = PDescribe("InboundPassthroughDisabled", inbound_communication.InboundPassthroughDisabled, Ordered)
	_ = PDescribe("ZoneEgress Internal Services", zoneegress.InternalServices, Ordered)
	_ = PDescribe("Connectivity", connectivity.Connectivity, Ordered)
	_ = PDescribe("Connectivity Gateway IPV6 CNI V2", connectivity.GatewayIPV6CNIV2, Ordered)
	_ = PDescribe("Sync", multizone_sync.Sync, Ordered)
	_ = PDescribe("Sync V2", multizone_sync.SyncV2, Ordered)
	_ = PDescribe("MeshTrafficPermission", meshtrafficpermission.MeshTrafficPermission, Ordered)
	_ = PDescribe("Zone Disable", zonedisable.ZoneDisable, Ordered)
	_ = FDescribe("External Services", externalservices.ExternalServicesOnMultizoneUniversal, Ordered)
	_ = PDescribe("Ownership", ownership.MultizoneUniversal, Ordered)
	_ = PDescribe("Resilience", resilience.ResilienceMultizoneUniversal, Ordered)
	_ = PDescribe("Resilience Postgres", resilience.ResilienceMultizoneUniversalPostgres, Ordered)
)
