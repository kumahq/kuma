package kubernetes_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/v3/pkg/test"
	"github.com/kumahq/kuma/v3/test/e2e_env/kubernetes/api"
	"github.com/kumahq/kuma/v3/test/e2e_env/kubernetes/appprobeproxy"
	"github.com/kumahq/kuma/v3/test/e2e_env/kubernetes/connectivity"
	"github.com/kumahq/kuma/v3/test/e2e_env/kubernetes/container_patch"
	"github.com/kumahq/kuma/v3/test/e2e_env/kubernetes/defaults"
	externalname_services "github.com/kumahq/kuma/v3/test/e2e_env/kubernetes/externalname-services"
	"github.com/kumahq/kuma/v3/test/e2e_env/kubernetes/gateway"
	"github.com/kumahq/kuma/v3/test/e2e_env/kubernetes/graceful"
	"github.com/kumahq/kuma/v3/test/e2e_env/kubernetes/healthcheck"
	"github.com/kumahq/kuma/v3/test/e2e_env/kubernetes/inspect"
	"github.com/kumahq/kuma/v3/test/e2e_env/kubernetes/jobs"
	"github.com/kumahq/kuma/v3/test/e2e_env/kubernetes/k8s_api_bypass"
	"github.com/kumahq/kuma/v3/test/e2e_env/kubernetes/kic"
	"github.com/kumahq/kuma/v3/test/e2e_env/kubernetes/meshcircuitbreaker"
	meshexternalservices "github.com/kumahq/kuma/v3/test/e2e_env/kubernetes/meshexternalservice"
	"github.com/kumahq/kuma/v3/test/e2e_env/kubernetes/meshfaultinjection"
	"github.com/kumahq/kuma/v3/test/e2e_env/kubernetes/meshhealthcheck"
	"github.com/kumahq/kuma/v3/test/e2e_env/kubernetes/meshhttproute"
	"github.com/kumahq/kuma/v3/test/e2e_env/kubernetes/meshidentity"
	"github.com/kumahq/kuma/v3/test/e2e_env/kubernetes/meshmetric"
	"github.com/kumahq/kuma/v3/test/e2e_env/kubernetes/meshpassthrough"
	"github.com/kumahq/kuma/v3/test/e2e_env/kubernetes/meshproxypatch"
	"github.com/kumahq/kuma/v3/test/e2e_env/kubernetes/meshretry"
	"github.com/kumahq/kuma/v3/test/e2e_env/kubernetes/meshtcproute"
	"github.com/kumahq/kuma/v3/test/e2e_env/kubernetes/meshtimeout"
	"github.com/kumahq/kuma/v3/test/e2e_env/kubernetes/meshtrafficpermission"
	"github.com/kumahq/kuma/v3/test/e2e_env/kubernetes/observability"
	"github.com/kumahq/kuma/v3/test/e2e_env/kubernetes/reachableservices"
	"github.com/kumahq/kuma/v3/test/e2e_env/kubernetes/workload"
	. "github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/envs/kubernetes"
)

func TestE2E(t *testing.T) {
	test.RunE2ESpecs(t, "E2E Kubernetes Suite")
}

var (
	_ = E2ESynchronizedBeforeSuite(kubernetes.SetupAndGetState, kubernetes.RestoreState)
	_ = SynchronizedAfterSuite(func() {}, kubernetes.SynchronizedAfterSuite)
	_ = ReportAfterSuite("kubernetes after suite", kubernetes.AfterSuite)
	// Opt-in (KUMA3_PREFLIGHT_BIN + KUMA3_PREFLIGHT_DIR): snapshot the CP after each spec
	// so kuma3-preflight can classify which tests use Kuma-3.0-removed features. No-op otherwise.
	_ = AfterEach(func() {
		CapturePreflightCluster(CurrentSpecReport().FullText(), kubernetes.Cluster)
	})
)

var (
	_ = Describe("Virtual Probes", Label("job-2"), healthcheck.VirtualProbes, Ordered)
	_ = Describe("Delegated Gateway", Label("job-2"), Label("kind-not-supported", "ipv6-not-supported"), gateway.Delegated, Ordered)
	_ = Describe("Graceful", Label("job-1"), graceful.Graceful, Ordered)
	_ = Describe("Eviction", Label("job-1"), graceful.Eviction, Ordered)
	_ = XDescribe("Change Service", graceful.ChangeService, Ordered)
	_ = Describe("Jobs", Label("job-3"), jobs.Jobs)
	_ = Describe("Container Patch", Label("job-3"), container_patch.ContainerPatch, Ordered)
	_ = Describe("MeshTrace", Label("job-3"), observability.PluginTest, Ordered)
	_ = Describe("MeshTrace Zone Proxy", Label("job-3"), observability.ZoneProxyPluginTest, Ordered)
	_ = Describe("Inspect", Label("job-3"), inspect.Inspect, Ordered)
	_ = Describe("K8S API Bypass", Label("job-3"), k8s_api_bypass.K8sApiBypass, Ordered)
	_ = Describe("Reachable Services", Label("job-1"), reachableservices.ReachableServices, Ordered)
	_ = Describe("Defaults", Label("job-3"), defaults.Defaults, Ordered)
	_ = Describe("Mesh External Services", Label("job-2"), meshexternalservices.MeshExternalServices, Ordered)
	_ = Describe("ExternalName Services", Label("job-3"), externalname_services.ExternalNameServices, Ordered)
	_ = Describe("Kong Ingress Controller", Label("job-0"), kic.KICKubernetes, Ordered)
	_ = Describe("MeshTrafficPermission API", Label("job-3"), meshtrafficpermission.API, Ordered)
	_ = Describe("MeshTrafficPermission Rules", Label("job-3"), meshtrafficpermission.MeshTrafficPermissionRules, Ordered)
	_ = Describe("MeshTimeout API", Label("job-1"), meshtimeout.MeshTimeout, Ordered)
	_ = Describe("MeshTimeout Zone Proxy", Label("job-1"), meshtimeout.ZoneProxyMeshTimeout, Ordered)
	_ = Describe("MeshHealthCheck API", Label("job-3"), meshhealthcheck.API, Ordered)
	_ = Describe("MeshCircuitBreaker", Label("job-0"), meshcircuitbreaker.MeshCircuitBreaker, Ordered)
	_ = Describe("MeshMetric", Label("job-1"), meshmetric.MeshMetric, Ordered)
	_ = Describe("MeshRetry", Label("job-3"), meshretry.API, Ordered)
	_ = Describe("MeshProxyPatch", Label("job-3"), meshproxypatch.MeshProxyPatch, Ordered)
	_ = Describe("MeshFaultInjection", Label("job-2"), meshfaultinjection.MeshFaultInjection, Ordered)
	_ = Describe("MeshFaultInjection Zone Proxy", Label("job-2"), meshfaultinjection.ZoneProxy, Ordered)
	_ = Describe("MeshHTTPRoute", Label("job-3"), meshhttproute.Test, Ordered)
	_ = Describe("MeshTCPRoute", Label("job-3"), meshtcproute.Test, Ordered)
	_ = Describe("Apis", Label("job-3"), api.Api, Ordered)
	_ = Describe("Connectivity - Headless Services", Label("job-1"), connectivity.HeadlessServices, Ordered)
	_ = Describe("Connectivity - Exclude Outbound Port", Label("job-1"), connectivity.ExcludeOutboundPort, Ordered)
	_ = Describe("Wait for Envoy", Label("job-1"), graceful.WaitForEnvoyReady, Ordered)
	_ = Describe("MeshPassthrough", Label("job-3"), meshpassthrough.MeshPassthrough, Ordered)
	_ = Describe("ApplicationProbeProxy", Label("job-0"), appprobeproxy.ApplicationProbeProxy, Ordered)
	_ = Describe("MeshIdentity Spire", Label("job-3"), meshidentity.Spire, Ordered)
	_ = Describe("Workload", Label("job-2"), workload.Workload, Ordered)
)
