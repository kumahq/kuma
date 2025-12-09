---
applyTo:
  - "pkg/api-server/**"
---

# API Server Guidelines

## Architecture

**Purpose:** REST API (go-restful/v3) for CP resource management, inspection, debugging

**Structure:**
```
pkg/api-server/
├── authn/customization/filters/mappers/types/  # Support modules
├── server.go                                    # Lifecycle (setup, TLS)
├── resource_endpoints.go                        # CRUD (1,200+ lines)
├── inspect_*_endpoints.go                       # Policy/Envoy inspection
└── pagination.go, *_endpoint.go                 # Utilities, endpoints
```

**Key Types:**
```go
type ApiServer struct {
    mux        *http.ServeMux
    config     api_server.ApiServerConfig
    httpReady  atomic.Bool
    httpsReady atomic.Bool
}

func NewApiServer(rt runtime.Runtime, meshContextBuilder, defs, cfg) (*ApiServer, error)
```

**Extension Points:** `rt.APIWebServiceCustomize()(ws)`, `rt.APIInstaller().Install(container)`

## Core Patterns

### Endpoint Registration (`resource_endpoints.go`)

```go
func addResourcesEndpoints(ws *restful.WebService, defs, rm, cfg, ...) {
    endpoints := &resourceEndpoints{resManager: rm, descriptor: descriptor, ...}
    ws.Route(ws.GET("/meshes/{mesh}/dataplanes/{name}").To(endpoints.findResource(false))...)
}

// Handler closure pattern
func (r *resourceEndpoints) findResource(withInsight bool) func(*restful.Request, *restful.Response) {
    return func(request *restful.Request, response *restful.Response) {
        name := request.PathParameter("name")
        meshName, err := r.meshFromRequest(request)
        // ... handler logic
    }
}
```

### CRUD Operations

**Endpoints:** GET/PUT/DELETE `/meshes/{mesh}/{resource}/{name}`, GET `/meshes/{mesh}/{resource}`

**Dual Format:**
```go
func formatResource(resource, format, k8sMapper, namespace) (any, error) {
    switch format {
    case "k8s", "kubernetes": return k8sMapper(resource, namespace), nil  // K8s YAML
    case "universal", "": return rest.From.Resource(resource), nil        // REST JSON
    default: return nil, errors.Errorf("unknown format: %s", format)
    }
}
```

**Pagination:** Default 100, max 1000 • `?size=100&offset=0` • `InvalidPageSizeError`
**Filtering:** Labels `?labels=k:v`, tags `?tags=k:v`, gateway `?gateway=true/false/builtin/delegated`

### Access Control (8 validation points)

```go
// Every CRUD operation validates access
r.resourceAccess.ValidateGet(ctx, key, descriptor, user)       // Get
r.resourceAccess.ValidateList(ctx, mesh, descriptor, user)     // List
r.resourceAccess.ValidateCreate(ctx, key, spec, descriptor, user)  // Create/Update
r.resourceAccess.ValidateDelete(ctx, key, spec, descriptor, user)  // Delete

// User from context
user := user_model.FromCtx(request.Request.Context())
```

### Error Handling (118+ usages)

```go
if err != nil {
    rest_errors.HandleError(request.Request.Context(), response, err, "User-friendly message")
    return
}

// Custom error types (pkg/api-server/types/)
type InvalidPageSizeError struct { Reason string }
func (e *InvalidPageSizeError) Is(err error) bool { _, ok := err.(*InvalidPageSizeError); return ok }
```

## Inspection Endpoints

**Policy Inspection** (`inspect_endpoints.go`):
```go
GET /meshes/{mesh}/dataplanes/{dp}/_policies      # Matched policies
GET /meshes/{mesh}/dataplanes/{dp}/_config        # Full Envoy config
GET /meshes/{mesh}/dataplanes/{dp}/_rules         # Inbound/outbound rules
GET /meshes/{mesh}/{policy}/{name}/dataplanes     # Policy → dataplanes

// Pattern: BuildMeshContext → BuildProxy → Generate config/policies → Format
meshContext := meshContextBuilder.BuildMeshContext(ctx, meshName)
proxy := sync.DefaultDataplaneProxyBuilder(*cfg, envoy.APIV3).Build(ctx, key, metadata, meshContext)
```

**Envoy Admin** (`inspect_envoy_admin_endpoints.go`):
```go
GET /meshes/{mesh}/dataplanes/{dp}/clusters|stats|config_dump|xds
GET /zoneingresses/{zi}/{type}, /zoneegresses/{ze}/{type}

type inspectClient struct {
    adminClient admin.EnvoyAdminClient; access access.EnvoyAdminAccess; rm manager.ResourceManager
}
// Pattern: Extract key → Validate access → Call Envoy admin API → Return raw response
```

**MeshService** (`inspect_mesh_service.go`):
```go
GET /meshes/{mesh}/meshservices/{svc}/_resources/dataplanes
GET /meshes/{mesh}/{serviceType}/{name}/_hostnames
```

## Authentication & Authorization

**Auth Filters** (`server.go:122-132`):
```go
container.Filter(authn.LocalhostAuthenticator)  // 127.0.0.1/::1 = admin (if LocalhostIsAdmin)
container.Filter(rt.APIServerAuthenticator())    // Pluggable (certs, tokens)
```

**Types:** `authn/localhost.go`, `authn/skip.go`, custom via runtime
**Access Validators:** `access.ResourceAccess` (CRUD), `access.EnvoyAdminAccess` (Envoy admin)

**Read-Only Mode:**
```go
if r.descriptor.ReadOnly || readOnly {
    rest_errors.HandleError(ctx, response, &types.MethodNotAllowed{Reason: "resource is read-only"}, "...")
}
// Cases: K8s dataplanes, global CP zone resources, federated zone, ApiServer.ReadOnly=true
```

## Common Patterns

**Request/Response:**
```go
// Extraction
name := request.PathParameter("name"); mesh := request.PathParameter("mesh")
meshName, err := r.meshFromRequest(request)  // With default mesh fallback
format := request.QueryParameter("format")   // k8s or universal
pageSize, offset := pagination(request); filter, err := r.filter(request)

// Mesh name resolution (default for global-scoped resources)
func (r *resourceEndpoints) meshFromRequest(request *restful.Request) (string, error) {
    if mesh := request.PathParameter("mesh"); mesh != "" { return mesh, nil }
    if r.descriptor.Scope == model.ScopeMesh { return "", errors.New("mesh is required") }
    return core_model.DefaultMesh, nil
}

// Response writing
res, err := formatResource(resource, format, k8sMapper, namespace)
if err != nil { rest_errors.HandleError(ctx, response, err, "Format error"); return }
if err := response.WriteAsJson(res); err != nil { log.Error(err, "Could not write response") }
```

**Overview (Insight) Handling:**
```go
if withInsight {
    insight := r.descriptor.NewInsight()
    r.resManager.Get(ctx, insight, store.GetByKey(name, mesh))  // Optional, log but don't fail
    if overview, ok := r.descriptor.NewOverview().(core_model.OverviewResource); ok {
        overview.SetOverviewSpec(resource, insight)
        resource = overview.(core_model.Resource)
    }
}
```

## Testing

**Utilities** (`api_server_suite_test.go`):
```go
type testApiServerConfigurer struct {
    store store.ResourceStore; config *config_api_server.ApiServerConfig
    metrics func() core_metrics.Metrics; zone string; global bool
}
func NewTestApiServerConfigurer() *testApiServerConfigurer {
    return &testApiServerConfigurer{store: memory.NewStore(), config: defaultApiServerConfig()}
}
```

**Golden Files:** `testdata/` (66+ files) • Update: `UPDATE_GOLDEN_FILES=true make test`

**Pattern:**
```go
var _ = Describe("Resource Endpoints", func() {
    BeforeEach(func() { apiServer, _ = NewTestApiServerConfigurer().Start() })
    It("should list resources", func() {
        store.Create(ctx, NewDataplane().WithName("dp1").Build(), ...)
        resp, _ := client.Get("/meshes/default/dataplanes")
        Expect(resp.StatusCode).To(Equal(200))
        body, _ := io.ReadAll(resp.Body)
        Expect(body).To(MatchGoldenJSON("testdata", "list-dataplanes.golden.json"))
    })
})
```

## Configuration & Integration

**Config** (`pkg/config/api-server/config.go`):
```go
type ApiServerConfig struct {
    HTTP struct { Enabled bool; Interface string; Port uint32 }  // 0.0.0.0:5681
    HTTPS struct { Enabled bool; Interface string; Port uint32; TlsCertFile, TlsKeyFile string
        TlsMinVersion string; TlsCipherSuites []string }  // 5682, TLSv1_2
    Auth struct { ClientCertsDir string }
    Authn struct { LocalhostIsAdmin bool; Type string }  // 127.0.0.1/::1 = admin
    CorsAllowedDomains []string; ReadOnly bool
    GUI struct { Enabled bool; BasePath, RootUrl string }
}
// TLS: Min version, cipher suites, client cert pool, RequireAndVerifyClientCert/VerifyClientCertIfGiven
```

**Runtime Dependencies:**
```go
rt.ResourceManager()           // CRUD
rt.Metrics()                   // Prometheus
rt.APIServerAuthenticator()    // Auth plugin
rt.GlobalInsightService()      // Statistics
rt.Access()                    // Resource/Envoy admin access
rt.EnvoyAdminClient()          // Envoy admin client
```

**Resource Manager:**
```go
resManager.Get(ctx, resource, store.GetByKey(name, mesh))
resManager.List(ctx, list, store.ListByMesh(mesh), store.ListByPage(size, offset))
resManager.Create|Update|Delete(ctx, resource, ...)
```

**XDS Context** (for `_config/_policies/_rules`):
```go
meshContext := meshContextBuilder.BuildMeshContext(ctx, meshName)
proxy := sync.DefaultDataplaneProxyBuilder(*cfg, envoy.APIV3).Build(ctx, key, metadata, meshContext)
```

## Endpoints

| Category | Endpoints | File |
|----------|-----------|------|
| CRUD | GET/PUT/DELETE `/meshes/{mesh}/{resource}/{name}`, GET `/meshes/{mesh}/{resource}` | `resource_endpoints.go` |
| Policy Inspection | `/{mesh}/{policy}/{name}/dataplanes`, `/{mesh}/dataplanes/{dp}/_config\|_policies\|_rules` | `resource_endpoints.go` |
| Envoy Admin | `/{mesh}/dataplanes/{dp}/{type}`, `/zoneingresses/{zi}/{type}`, `/zoneegresses/{ze}/{type}` | `inspect_envoy_admin_endpoints.go` |
| MeshService | `/{mesh}/meshservices/{svc}/_resources/dataplanes`, `/{mesh}/{svcType}/{name}/_hostnames` | `inspect_mesh_service.go` |
| Other | `/global-insights`, `/policies`, `/_kri/{kri}`, `/config`, `/who-am-i`, `/tokens/*`, `/gui/*` | Various |

## Before Implementation

**Search Patterns:** CRUD (`resource_endpoints.go`), inspection (`inspect_*.go`), errors (`types/errors.go`), access (`Validate*`)

**Checklist:**
- Dual format support (`formatResource()`) • Pagination if listing • Filtering (labels, tags, gateway)
- Read-only enforcement • Access control (`resourceAccess.Validate*()`) • Error handling (`rest_errors.HandleError()`)

## Review Focus

**Security:** Access control (8 points) • User context • Read-only mode • TLS config • No sensitive logs
**Correctness:** K8s YAML + Universal JSON • Pagination (max 1000) • Mesh name resolution • Insight handling (optional)
**Integration:** Runtime deps • ResourceManager ops • MeshContext for xDS • Envoy admin client • Access validators
**Testing:** Golden files • Both formats • Pagination edges • Access control • Error scenarios (404, 403, 400)

## Quick Reference

**Interfaces:** `resourceEndpoints:70`, `inspectClient:40`, `ApiServer:60`
**Functions:** `formatResource():257`, `pagination():20`, `rest_errors.HandleError()`
**Access:** `Validate{Get,List,Create,Update,Delete}()`
**Errors:** `InvalidPageSizeError`, `MethodNotAllowed`
