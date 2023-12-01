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
	_ = PDescribe("Virtual Probes", healthcheck.VirtualProbes, Ordered)
	_ = PDescribe("Gateway", gateway.Gateway, Ordered)
	_ = PDescribe("Gateway - Cross-mesh", gateway.CrossMeshGatewayOnKubernetes, Ordered)
	_ = PDescribe("Gateway - Gateway API", gateway.GatewayAPI, Ordered)
	_ = PDescribe("Gateway - mTLS", gateway.Mtls, Ordered)
	_ = PDescribe("Gateway - Resources", gateway.Resources, Ordered)
	_ = PDescribe("Graceful", graceful.Graceful, Ordered)
	_ = PDescribe("Eviction", graceful.Eviction, Ordered)
	_ = Describe("Change Service", graceful.ChangeService, Ordered)
	_ = PDescribe("Jobs", jobs.Jobs)
	_ = PDescribe("Membership", membership.Membership, Ordered)
	_ = PDescribe("Container Patch", container_patch.ContainerPatch, Ordered)
	_ = PDescribe("Metrics", observability.ApplicationsMetrics, Ordered)
	_ = PDescribe("Tracing", observability.Tracing, Ordered)
	_ = PDescribe("MeshTrace", observability.PluginTest, Ordered)
	_ = PDescribe("Traffic Log", trafficlog.TCPLogging, Ordered)
	_ = PDescribe("Inspect", inspect.Inspect, Ordered)
	_ = PDescribe("K8S API Bypass", k8s_api_bypass.K8sApiBypass, Ordered)
	_ = PDescribe("Reachable Services", reachableservices.ReachableServices, Ordered)
	_ = PDescribe("Defaults", defaults.Defaults, Ordered)
	_ = PDescribe("External Services", externalservices.ExternalServices, Ordered)
	_ = PDescribe("External Services Permissive MTLS", externalservices.PermissiveMTLS, Ordered)
	_ = PDescribe("ExternalName Services", externalname_services.ExternalNameServices, Ordered)
	_ = PDescribe("Virtual Outbound", virtualoutbound.VirtualOutbound, Ordered)
	_ = PDescribe("Kong Ingress Controller", Label("arm-not-supported"), kic.KICKubernetes, Ordered)
	_ = PDescribe("MeshTrafficPermission API", meshtrafficpermission.API, Ordered)
	_ = PDescribe("MeshRateLimit API", meshratelimit.API, Ordered)
	_ = PDescribe("MeshTimeout API", meshtimeout.MeshTimeout, Ordered)
	_ = PDescribe("MeshHealthCheck API", meshhealthcheck.API, Ordered)
	_ = PDescribe("MeshCircuitBreaker API", meshcircuitbreaker.API, Ordered)
	_ = PDescribe("MeshCircuitBreaker", meshcircuitbreaker.MeshCircuitBreaker, Ordered)
	_ = PDescribe("MeshRetry", meshretry.API, Ordered)
	_ = PDescribe("MeshProxyPatch", meshproxypatch.MeshProxyPatch, Ordered)
	_ = PDescribe("MeshFaultInjection", meshfaultinjection.API, Ordered)
	_ = PDescribe("MeshHTTPRoute", meshhttproute.Test, Ordered)
	_ = PDescribe("MeshTCPRoute", meshtcproute.Test, Ordered)
	_ = PDescribe("Apis", api.Api, Ordered)
	_ = PDescribe("Connectivity - Headless Services", connectivity.HeadlessServices, Ordered)
	_ = PDescribe("Connectivity - Exclude Outbound Port", connectivity.ExcludeOutboundPort, Ordered)
	_ = PDescribe("Wait for Envoy", graceful.WaitForEnvoyReady, Ordered)
)
