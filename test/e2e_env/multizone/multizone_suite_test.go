package auth_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	multizone_sync "github.com/kumahq/kuma/test/e2e_env/multizone/sync"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func TestE2E(t *testing.T) {
	test.RunE2ESpecs(t, "E2E Multizone Suite")
}

var _ = SynchronizedBeforeSuite(multizone.SetupAndGetState, multizone.RestoreState)

var (
	// _ = Describe("Gateway", gateway.GatewayHybrid, Ordered)
	// _ = Describe("Cross-mesh Gateways", gateway.CrossMeshGatewayOnMultizone, Ordered)
	// _ = Describe("External Service locality aware", localityawarelb.ExternalServicesWithLocalityAwareLb, Ordered)
	// _ = Describe("Healthcheck", healthcheck.ApplicationOnUniversalClientOnK8s, Ordered)
	// _ = Describe("Inspect", inspect.Inspect, Ordered)
	// _ = Describe("TrafficPermission", trafficpermission.TrafficPermission, Ordered)
	// _ = Describe("TrafficRoute", trafficroute.TrafficRoute, Ordered)
	// _ = Describe("MeshHTTPRoute", meshhttproute.Test, Ordered)
	// _ = Describe("InboundPassthrough", inbound_communication.InboundPassthrough, Ordered)
	// _ = Describe("InboundPassthroughDisabled", inbound_communication.InboundPassthroughDisabled, Ordered)
	// _ = Describe("ZoneEgress Internal Services", zoneegress.InternalServices, Ordered)
	// _ = Describe("Connectivity", connectivity.Connectivity, Ordered)
	// _ = Describe("Sync", multizone_sync.Sync, Ordered)
	_ = Describe("Sync", multizone_sync.SyncV2, Ordered)
	// _ = Describe("MeshTrafficPermission", meshtrafficpermission.MeshTrafficPermission, Ordered)
	// _ = Describe("Zone Disable", zonedisable.ZoneDisable, Ordered)
	// _ = Describe("External Services", externalservices.ExternalServicesOnMultizoneUniversal, Ordered)
	// _ = Describe("Ownership", ownership.MultizoneUniversal, Ordered)
	// _ = Describe("Resilience", resilience.ResilienceMultizoneUniversal, Ordered)
	// _ = Describe("Resilience Postgres", resilience.ResilienceMultizoneUniversalPostgres, Ordered)
)
