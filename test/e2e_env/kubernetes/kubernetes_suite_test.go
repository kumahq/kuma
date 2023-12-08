package kubernetes_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/api"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/connectivity"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/container_patch"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/defaults"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/externalname-services"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/externalservices"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/gateway"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/graceful"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/healthcheck"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/inspect"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/jobs"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/k8s_api_bypass"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/kic"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/membership"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/meshcircuitbreaker"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/meshfaultinjection"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/meshhealthcheck"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/meshhttproute"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/meshproxypatch"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/meshratelimit"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/meshretry"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/meshtcproute"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/meshtimeout"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/meshtrafficpermission"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/observability"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/reachableservices"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/trafficlog"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/virtualoutbound"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func TestE2E(t *testing.T) {
	test.RunE2ESpecs(t, "E2E Kubernetes Suite")
}

var _ = E2ESynchronizedBeforeSuite(kubernetes.SetupAndGetState, kubernetes.RestoreState)

// SynchronizedAfterSuite keeps the main process alive until all other processes finish.
// Otherwise, we would close port-forward to the CP and remaining tests executed in different processes may fail.
var _ = SynchronizedAfterSuite(func() {}, func() {})

var (
	_ = Describe("Virtual Probes", healthcheck.VirtualProbes, Ordered)
	_ = Describe("Gateway", gateway.Gateway, Ordered)
	_ = Describe("Gateway - Cross-mesh", gateway.CrossMeshGatewayOnKubernetes, Ordered)
	_ = Describe("Gateway - Gateway API", gateway.GatewayAPI, Ordered)
	_ = Describe("Gateway - mTLS", gateway.Mtls, Ordered)
	_ = Describe("Gateway - Resources", gateway.Resources, Ordered)
	_ = Describe("Graceful", graceful.Graceful, Ordered)
	_ = Describe("Eviction", graceful.Eviction, Ordered)
	_ = XDescribe("Change Service", graceful.ChangeService, Ordered)
	_ = Describe("Jobs", jobs.Jobs)
	_ = Describe("Membership", membership.Membership, Ordered)
	_ = Describe("Container Patch", container_patch.ContainerPatch, Ordered)
	_ = Describe("Metrics", observability.ApplicationsMetrics, Ordered)
	_ = Describe("Tracing", observability.Tracing, Ordered)
	_ = Describe("MeshTrace", observability.PluginTest, Ordered)
	_ = Describe("Traffic Log", trafficlog.TCPLogging, Ordered)
	_ = Describe("Inspect", inspect.Inspect, Ordered)
	_ = Describe("K8S API Bypass", k8s_api_bypass.K8sApiBypass, Ordered)
	_ = Describe("Reachable Services", reachableservices.ReachableServices, Ordered)
	_ = Describe("Defaults", defaults.Defaults, Ordered)
	_ = Describe("External Services", externalservices.ExternalServices, Ordered)
	_ = Describe("External Services Permissive MTLS", externalservices.PermissiveMTLS, Ordered)
	_ = Describe("ExternalName Services", externalname_services.ExternalNameServices, Ordered)
	_ = Describe("Virtual Outbound", virtualoutbound.VirtualOutbound, Ordered)
	_ = Describe("Kong Ingress Controller", Label("arm-not-supported"), kic.KICKubernetes, Ordered)
	_ = Describe("MeshTrafficPermission API", meshtrafficpermission.API, Ordered)
	_ = Describe("MeshRateLimit API", meshratelimit.API, Ordered)
	_ = Describe("MeshTimeout API", meshtimeout.MeshTimeout, Ordered)
	_ = Describe("MeshHealthCheck API", meshhealthcheck.API, Ordered)
	_ = Describe("MeshCircuitBreaker API", meshcircuitbreaker.API, Ordered)
	_ = Describe("MeshCircuitBreaker", meshcircuitbreaker.MeshCircuitBreaker, Ordered)
	_ = Describe("MeshRetry", meshretry.API, Ordered)
	_ = Describe("MeshProxyPatch", meshproxypatch.MeshProxyPatch, Ordered)
	_ = Describe("MeshFaultInjection", meshfaultinjection.API, Ordered)
	_ = Describe("MeshHTTPRoute", meshhttproute.Test, Ordered)
	_ = Describe("MeshTCPRoute", meshtcproute.Test, Ordered)
	_ = Describe("Apis", api.Api, Ordered)
	_ = Describe("Connectivity - Headless Services", connectivity.HeadlessServices, Ordered)
	_ = Describe("Connectivity - Exclude Outbound Port", connectivity.ExcludeOutboundPort, Ordered)
	_ = Describe("Wait for Envoy", graceful.WaitForEnvoyReady, Ordered)
)
