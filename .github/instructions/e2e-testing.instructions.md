---
applyTo:
  - "test/**"
---

# E2E Testing Guidelines

## Test Organization

### Directory Structure

**Preferred: e2e_env (Parallel Tests)**
```
test/e2e_env/
├── universal/           # Universal mode tests (44 subdirectories)
├── kubernetes/          # Kubernetes mode tests (36 subdirectories)
├── multizone/           # Multi-zone tests (30 subdirectories)
└── gatewayapi/          # Gateway API conformance
```

**Legacy: e2e (Sequential Tests)**
```
test/e2e/
└── <feature>/           # Legacy sequential tests (avoid)
```

**Framework & Utilities**
```
test/framework/
├── interface.go         # Cluster, ControlPlane interfaces
├── setup.go             # ClusterSetup builder pattern
├── client/              # HTTP/TCP client helpers
├── kumactl/             # Kumactl integration
└── envoy_admin/         # Envoy admin API
```

### When to Use Each

**Use e2e_env:**
- Policy functionality tests
- Traffic routing tests
- mTLS tests
- Service connectivity tests
- Most feature tests

**Use e2e (only when necessary):**
- CP restart/resilience tests
- Special deployment configurations
- Zone Egress absence tests
- Default VIPs disabled tests

**Key Differences:**
- `e2e_env`: Parallel on shared deployments, own mesh/namespace per test
- `e2e`: Sequential, deploys own Kuma instance (slower)

## Test Framework Architecture

### Cluster Interface

**Location:** `test/framework/interface.go:15`

```go
type Cluster interface {
    DeployKuma(mode CpMode, opts ...DeployOptionsFunc) error
    Deploy(app Deployment) error
    Delete(app Deployment) error
    DeleteMesh(name string) error
    DeleteMeshApps(mesh string) error
    Exec(namespace, pod, container string, cmd ...string) (string, string, error)
    GetKumactlOptions() *KumactlOptions
}
```

**Implementations:**
- `K8sCluster` - Kubernetes clusters
- `UniversalCluster` - Universal (Docker) clusters

### Setup Pattern (Builder)

**Location:** `test/framework/setup.go:20`

```go
// Type definition
type InstallFunc func(Cluster) error

// Usage
err := NewClusterSetup().
    Install(MTLSMeshUniversal(meshName)).
    Install(TestServerUniversal("test-server", meshName,
        WithArgs([]string{"echo", "--instance", "echo-v1"}))).
    Install(DemoClientUniversal("demo-client", meshName,
        WithTransparentProxy(true))).
    Setup(universal.Cluster)
Expect(err).ToNot(HaveOccurred())
```

**Common InstallFuncs:**
- `MTLSMeshUniversal(name)` / `MTLSMeshKubernetes(name)`
- `MeshTrafficPermissionAllowAllUniversal(name)` / `MeshTrafficPermissionAllowAllKubernetes(name)`
- `TestServerUniversal(name, mesh, opts...)` - Echo server
- `DemoClientUniversal(name, mesh, opts...)` - Client
- `YamlUniversal(yaml)` / `YamlK8s(yaml)` - Raw YAML
- `Combine(fns...)` - Sequential execution
- `Parallel(fns...)` - Parallel execution

**Deployment Options:**
- `WithArgs([]string)` - Custom command arguments
- `WithServiceName(name)` - Override service name
- `WithProtocol(protocol)` - Set protocol (http, tcp, grpc)
- `WithTransparentProxy(bool)` - Enable/disable tproxy
- `WithMesh(name)` - Target mesh
- `WithNamespace(ns)` - Kubernetes namespace
- `WithReachableServices(services...)` - Limit reachable services

## Ginkgo/Gomega Patterns

### Suite Setup

**Suite Registration** (`*_suite_test.go`):
```go
func TestE2E(t *testing.T) {
    test.RunE2ESpecs(t, "E2E Universal Suite")
}

var (
    _ = E2ESynchronizedBeforeSuite(universal.SetupAndGetState, universal.RestoreState)
    _ = SynchronizedAfterSuite(func() {}, universal.SynchronizedAfterSuite)
    _ = ReportAfterSuite("universal after suite", universal.AfterSuite)
)

// Test registration
var (
    _ = Describe("MeshTrafficPermission", meshtrafficpermission.MeshTrafficPermissionUniversal, Ordered)
    _ = Describe("Gateway", gateway.Gateway, Ordered)
)
```

### Test Structure

**Standard Pattern:**
```go
func MeshTrafficPermissionUniversal() {
    meshName := "meshtrafficpermission"  // Unique per test suite

    BeforeAll(func() {
        // One-time setup for all tests in this describe block
        Expect(NewClusterSetup().
            Install(MTLSMeshUniversal(meshName)).
            Install(TestServerUniversal("test-server", meshName)).
            Setup(universal.Cluster)).To(Succeed())
    })

    E2EAfterAll(func() {
        // Cleanup (respects fail-fast)
        Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
        Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
    })

    AfterEachFailure(func() {
        // Debug on failure ONLY
        DebugUniversal(universal.Cluster, meshName)
    })

    E2EAfterEach(func() {
        // Per-test cleanup (respects fail-fast)
        items, err := universal.Cluster.GetKumactlOptions().KumactlList("meshtrafficpermissions", meshName)
        Expect(err).ToNot(HaveOccurred())
        for _, item := range items {
            err := universal.Cluster.GetKumactlOptions().KumactlDelete("meshtrafficpermission", item, meshName)
            Expect(err).ToNot(HaveOccurred())
        }
    })

    It("should allow traffic", func() {
        // Test implementation
    })
}
```

### Lifecycle Hooks

**Hook Types:**
- `BeforeAll` - Setup once before all tests
- `E2EAfterAll` - Cleanup after all tests (skips on fail-fast)
- `E2EAfterEach` - Cleanup after each test (skips on fail-fast)
- `AfterEachFailure` - Runs ONLY on test failure (debugging)
- `E2ESynchronizedBeforeSuite` - Global suite setup (first process only)
- `SynchronizedAfterSuite` - Global suite teardown

**IMPORTANT:** Use `E2EAfterAll`/`E2EAfterEach` instead of `AfterAll`/`AfterEach` to respect fail-fast mode.

## Common Test Patterns

### Eventually/Consistently Pattern

**Location:** All test files

```go
// Wait for condition to become true
Eventually(func(g Gomega) {
    response, err := client.CollectEchoResponse(cluster, "demo-client", "test-server.mesh")
    g.Expect(err).ToNot(HaveOccurred())
    g.Expect(response.Instance).To(Equal("echo-v1"))
}).Should(Succeed())

// Assert condition remains true
Consistently(func(g Gomega) {
    _, err := client.CollectEchoResponse(cluster, "demo-client", "test-server.mesh")
    g.Expect(err).ToNot(HaveOccurred())
}).Should(Succeed())
```

**Why `func(g Gomega)`:** Proper failure messages within Eventually/Consistently.

### Helper Functions

**Location:** Test files

```go
trafficAllowed := func(addr string) {
    GinkgoHelper()  // CRITICAL: Proper stack traces

    Eventually(func(g Gomega) {
        _, err := client.CollectEchoResponse(universal.Cluster, "demo-client", addr)
        g.Expect(err).ToNot(HaveOccurred())
    }).Should(Succeed())
}

trafficBlocked := func(expectedStatus int) {
    GinkgoHelper()  // CRITICAL: Proper stack traces

    Eventually(func(g Gomega) {
        response, err := client.CollectFailure(universal.Cluster, "demo-client", "test-server.mesh")
        g.Expect(err).ToNot(HaveOccurred())
        g.Expect(response.ResponseCode).To(Equal(expectedStatus))
    }).Should(Succeed())
}
```

**ALWAYS call `GinkgoHelper()` in test helpers for correct stack traces.**

## Client Helpers

### HTTP/TCP Client

**Location:** `test/framework/client/collect.go:15`

**Functions:**
```go
// Single request
response, err := client.CollectEchoResponse(
    cluster,
    "demo-client",                  // Source container/pod
    "test-server.mesh",             // Destination (*.mesh, http://addr)
    client.WithNumberOfRequests(1),
    client.WithHeader("x-custom", "value"),
)

// Multiple requests (load balancing)
instances, err := client.CollectResponsesByInstance(
    cluster,
    "demo-client",
    "backend.mesh",
    client.WithNumberOfRequests(100),
)
Expect(instances).To(HaveLen(2))  // Two backend instances

// Failure cases
failure, err := client.CollectFailure(
    cluster,
    "demo-client",
    "blocked-service.mesh",
)
Expect(failure.ResponseCode).To(Equal(403))
```

**Client Options:**
- `WithNumberOfRequests(n)` - Number of requests (default: 1)
- `WithHeader(key, value)` - Custom headers
- `FromKubernetesPod(namespace, app)` - Execute from K8s pod
- `WithVerbose()` - Verbose curl output
- `Insecure()` - Skip TLS verification
- `WithMaxTime(seconds)` - Request timeout

**Response Types:**
- `EchoResponse` - Instance, received headers, zone
- `FailureResponse` - ResponseCode, body

## Environment Access

### Shared Clusters

**Universal:**
```go
import "github.com/kumahq/kuma/v2/test/framework/envs/universal"

universal.Cluster  // Shared universal cluster
```

**Kubernetes:**
```go
import "github.com/kumahq/kuma/v2/test/framework/envs/kubernetes"

kubernetes.Cluster  // Shared kubernetes cluster
```

**Multizone:**
```go
import "github.com/kumahq/kuma/v2/test/framework/envs/multizone"

multizone.Global      // Global CP (Universal)
multizone.KubeZone1   // Kubernetes zone 1
multizone.KubeZone2   // Kubernetes zone 2
multizone.UniZone1    // Universal zone 1
multizone.UniZone2    // Universal zone 2
```

### Multizone Topology

**Default Setup:**
- Global CP: Universal
- 2 Kubernetes zones (KubeZone1, KubeZone2)
- 2 Universal zones (UniZone1, UniZone2)
- Zone Egress: Enabled by default

**IMPORTANT:** Call `WaitForMesh(meshName)` before deploying apps in multizone tests.

## Policy Application

### YAML Application

**Pattern:**
```go
yaml := `
type: MeshTrafficPermission
name: mtp-1
mesh: meshtrafficpermission
spec:
  targetRef:
    kind: MeshService
    name: test-server
  from:
    - targetRef:
        kind: Mesh
      default:
        action: Allow
`
err := YamlUniversal(yaml)(universal.Cluster)
Expect(err).ToNot(HaveOccurred())

// Kubernetes
err := YamlK8s(yaml)(kubernetes.Cluster)
Expect(err).ToNot(HaveOccurred())
```

### Kumactl Integration

**Location:** `test/framework/kumactl/kumactl.go:10`

```go
// List resources
items, err := cluster.GetKumactlOptions().KumactlList("meshtrafficpermissions", meshName)
Expect(err).ToNot(HaveOccurred())

// Delete resource
err := cluster.GetKumactlOptions().KumactlDelete("meshtrafficpermission", "mtp-1", meshName)
Expect(err).ToNot(HaveOccurred())

// Apply from string
err := cluster.GetKumactlOptions().KumactlApplyFromString(yaml)
Expect(err).ToNot(HaveOccurred())
```

## Debug Helpers

### Debug Functions

**Universal:**
```go
AfterEachFailure(func() {
    DebugUniversal(universal.Cluster, meshName)
})
```

**Kubernetes:**
```go
AfterEachFailure(func() {
    DebugKube(kubernetes.Cluster, meshName, "kuma-test", "app-namespace")
})
```

**Control Plane Assertions:**
```go
It("should not crash control plane", func() {
    ControlPlaneAssertions(universal.Cluster)
})
```

**What They Do:**
- Dump Envoy configs
- Print dataplane resources
- Show pod/container status
- Collect CP logs
- Check for panics/restarts

## Testing Scenarios

### Load Balancing Test

```go
It("should load balance across instances", func() {
    // when: send many requests
    instances, err := client.CollectResponsesByInstance(
        universal.Cluster,
        "demo-client",
        "backend.mesh",
        client.WithNumberOfRequests(100),
    )

    // then: distributed across instances
    Expect(err).ToNot(HaveOccurred())
    Expect(instances).To(HaveLen(2))
    Expect(instances["echo-v1"]).To(BeNumerically(">", 40))
    Expect(instances["echo-v2"]).To(BeNumerically(">", 40))
})
```

### Cross-Zone Test

```go
It("should route cross-zone", func() {
    // given: client in zone1, server in zone2
    Eventually(func(g Gomega) {
        response, err := client.CollectEchoResponse(
            multizone.KubeZone1,
            "demo-client",
            "backend.mesh",
        )
        g.Expect(err).ToNot(HaveOccurred())
        g.Expect(response.Zone).To(Equal("kube-zone-2"))
    }).Should(Succeed())
})
```

### mTLS Test

```go
It("should enforce mTLS", func() {
    // given: mTLS enabled mesh
    Expect(NewClusterSetup().
        Install(MTLSMeshUniversal(meshName)).
        Install(TestServerUniversal("backend", meshName)).
        Install(DemoClientUniversal("client", meshName)).
        Setup(universal.Cluster)).To(Succeed())

    // when: client connects
    // then: uses mTLS
    Eventually(func(g Gomega) {
        response, err := client.CollectEchoResponse(
            universal.Cluster,
            "client",
            "backend.mesh",
        )
        g.Expect(err).ToNot(HaveOccurred())
        g.Expect(response.TLS).To(BeTrue())
    }).Should(Succeed())
})
```

## Best Practices

### Test Isolation

**DO:**
- ✅ Use unique mesh name per test suite
- ✅ Use own namespace in Kubernetes
- ✅ Clean up resources in `E2EAfterEach`/`E2EAfterAll`
- ✅ Use `TriggerDeleteNamespace` instead of `DeleteNamespace` (K8s)

**DON'T:**
- ❌ Share mesh between test suites
- ❌ Rely on specific execution order (unless `Ordered`)
- ❌ Leave resources behind after tests

### Setup/Teardown

**DO:**
- ✅ Use `BeforeAll` for expensive setup
- ✅ Use `E2EAfterAll` for cleanup (respects fail-fast)
- ✅ Use `AfterEachFailure` for debug output
- ✅ Call `GinkgoHelper()` in all test helpers

**DON'T:**
- ❌ Use `AfterAll`/`AfterEach` (use E2E variants)
- ❌ Forget to clean up policies in `E2EAfterEach`
- ❌ Run debug helpers on success (use `AfterEachFailure`)

### Client Helpers

**DO:**
- ✅ Use `Eventually(func(g Gomega) { ... })` for async assertions
- ✅ Use `Consistently` to verify condition remains true
- ✅ Use `client.CollectFailure` for expected failures
- ✅ Check specific error codes (403, 503, etc.)

**DON'T:**
- ❌ Use `Eventually(func() { Expect(...) })` (use `func(g Gomega)`)
- ❌ Forget `GinkgoHelper()` in helper functions
- ❌ Use `time.Sleep` instead of `Eventually`

### Multizone

**DO:**
- ✅ Call `WaitForMesh(meshName)` before deploying apps
- ✅ Set conservative resource limits
- ✅ Use `KUMA_STORE_UNSAFE_DELETE=true` for faster cleanup

**DON'T:**
- ❌ Assume immediate sync between zones
- ❌ Deploy to zones before mesh syncs

## Common Mistakes

### ❌ NEVER Do

**No GinkgoHelper:**
```go
// ❌ No GinkgoHelper - wrong line numbers in failures
trafficAllowed := func(addr string) {
    Eventually(func(g Gomega) {
        _, err := client.CollectEchoResponse(cluster, "demo-client", addr)
        g.Expect(err).ToNot(HaveOccurred())
    }).Should(Succeed())
}
```

**Wrong Eventually Signature:**
```go
// ❌ Wrong signature - no Gomega parameter
Eventually(func() {
    _, err := client.CollectEchoResponse(cluster, "demo-client", "backend.mesh")
    Expect(err).ToNot(HaveOccurred())
}).Should(Succeed())
```

**Using AfterAll Instead of E2EAfterAll:**
```go
// ❌ Doesn't respect fail-fast
AfterAll(func() {
    Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
})
```

**Shared Mesh Between Suites:**
```go
// ❌ Causes test interference
meshName := "default"  // Multiple suites use same mesh
```

**No Cleanup:**
```go
// ❌ Leaves policies behind for next test
It("should apply policy", func() {
    yaml := `type: MeshTimeout...`
    err := YamlUniversal(yaml)(cluster)
    // No cleanup!
})
```

**Debug in AfterEach:**
```go
// ❌ Runs on every test (slow)
AfterEach(func() {
    DebugUniversal(universal.Cluster, meshName)
})
```

### ✅ ALWAYS Do

**With GinkgoHelper:**
```go
// ✅ Proper stack traces
trafficAllowed := func(addr string) {
    GinkgoHelper()

    Eventually(func(g Gomega) {
        _, err := client.CollectEchoResponse(cluster, "demo-client", addr)
        g.Expect(err).ToNot(HaveOccurred())
    }).Should(Succeed())
}
```

**Correct Eventually Signature:**
```go
// ✅ Gomega parameter for proper failures
Eventually(func(g Gomega) {
    _, err := client.CollectEchoResponse(cluster, "demo-client", "backend.mesh")
    g.Expect(err).ToNot(HaveOccurred())
}).Should(Succeed())
```

**E2E Lifecycle Hooks:**
```go
// ✅ Respects fail-fast mode
E2EAfterAll(func() {
    Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
})

E2EAfterEach(func() {
    // Cleanup policies per test
})
```

**Unique Mesh Per Suite:**
```go
// ✅ No test interference
meshName := "meshtrafficpermission"  // Unique to this suite
```

**Proper Cleanup:**
```go
// ✅ Clean up after each test
E2EAfterEach(func() {
    items, err := cluster.GetKumactlOptions().KumactlList("meshtimeouts", meshName)
    Expect(err).ToNot(HaveOccurred())
    for _, item := range items {
        err := cluster.GetKumactlOptions().KumactlDelete("meshtimeout", item, meshName)
        Expect(err).ToNot(HaveOccurred())
    }
})
```

**Debug Only on Failure:**
```go
// ✅ Only runs when test fails
AfterEachFailure(func() {
    DebugUniversal(universal.Cluster, meshName)
})
```

## Review Focus

### Structure
- Correct directory (e2e_env vs e2e)
- Suite registration with `RunE2ESpecs`
- `E2ESynchronizedBeforeSuite` for shared setup
- Test functions with unique mesh names

### Lifecycle
- `BeforeAll` for setup
- `E2EAfterAll` for cleanup (not `AfterAll`)
- `E2EAfterEach` for per-test cleanup
- `AfterEachFailure` for debug (not `AfterEach`)

### Patterns
- `Eventually(func(g Gomega) { ... })` signature
- `GinkgoHelper()` in all test helpers
- Proper client helper usage
- Resource cleanup in lifecycle hooks

### Isolation
- Unique mesh per test suite
- Proper namespace usage (K8s)
- No shared state between tests
- Complete cleanup

### Multizone
- `WaitForMesh()` before app deployment
- Cross-zone routing verification
- Zone-specific assertions

## Quick Reference

**Framework:**
- Cluster Interface: `test/framework/interface.go:15`
- Setup Pattern: `test/framework/setup.go:20`
- Client Helpers: `test/framework/client/collect.go:15`

**Environments:**
- Universal: `test/framework/envs/universal/env.go:10`
- Kubernetes: `test/framework/envs/kubernetes/env.go:10`
- Multizone: `test/framework/envs/multizone/env.go:10`

**Constants:**
- `test/framework/constants.go:10`

**Example Tests:**
- Universal: `test/e2e_env/universal/meshtrafficpermission/`
- Kubernetes: `test/e2e_env/kubernetes/meshhttproute/`
- Multizone: `test/e2e_env/multizone/inspect/`
