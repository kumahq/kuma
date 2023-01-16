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
	"github.com/kumahq/kuma/test/e2e_env/multizone/meshtrafficpermission"
	"github.com/kumahq/kuma/test/e2e_env/multizone/ownership"
	"github.com/kumahq/kuma/test/e2e_env/multizone/resilience"
	multizone_sync "github.com/kumahq/kuma/test/e2e_env/multizone/sync"
	"github.com/kumahq/kuma/test/e2e_env/multizone/trafficpermission"
	"github.com/kumahq/kuma/test/e2e_env/multizone/trafficroute"
	"github.com/kumahq/kuma/test/e2e_env/multizone/zonedisable"
	"github.com/kumahq/kuma/test/e2e_env/multizone/zoneegress"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func TestE2E(t *testing.T) {
	test.RunE2ESpecs(t, "E2E Multizone Suite")
}

var _ = SynchronizedBeforeSuite(multizone.SetupAndGetState, multizone.RestoreState)

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
var _ = Describe("Zone Disable", zonedisable.ZoneDisable, Ordered)
var _ = Describe("External Services", externalservices.ExternalServicesOnMultizoneUniversal, Ordered)
var _ = Describe("Ownership", ownership.MultizoneUniversal, Ordered)
var _ = Describe("Resilience", resilience.ResilienceMultizoneUniversal, Ordered)
var _ = Describe("Resilience Postgres", resilience.ResilienceMultizoneUniversalPostgres, Ordered)
