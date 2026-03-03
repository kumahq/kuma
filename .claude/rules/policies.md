# Policy plugin system

Policies live in `pkg/plugins/policies/` (MeshTrafficPermission, MeshHTTPRoute, MeshTimeout, etc.).

## Directory structure per policy

```
{policyname}/
├── api/v1alpha1/
│   ├── {policyname}.go              # HAND-WRITTEN: spec struct with +kuma:policy markers
│   ├── validator.go                 # HAND-WRITTEN: validation logic
│   ├── deprecated.go                # HAND-WRITTEN: deprecation warnings (optional)
│   ├── zz_generated.resource.go     # GENERATED
│   ├── zz_generated.deepcopy.go     # GENERATED
│   ├── zz_generated.helpers.go      # GENERATED
│   └── rest.yaml                    # GENERATED: OpenAPI spec
├── k8s/v1alpha1/
│   ├── groupversion_info.go         # HAND-WRITTEN: K8s group/version
│   ├── zz_generated.deepcopy.go     # GENERATED
│   └── zz_generated.types.go        # GENERATED
├── plugin/v1alpha1/
│   ├── plugin.go                    # HAND-WRITTEN: xDS generation logic
│   ├── plugin_test.go               # HAND-WRITTEN: tests with testdata/
│   └── testdata/                    # Golden files
└── zz_generated.plugin.go           # GENERATED: plugin registration
```

**Rule**: only edit hand-written files. Never edit `zz_generated.*` or `rest.yaml`.

## After changing a policy

```bash
make generate    # Regenerate all dependent files
make check       # Lint and validate
make test TEST_PKG_LIST=./pkg/plugins/policies/{policyname}/...
```

## Policy spec markers

Add above the main struct in `api/v1alpha1/{policyname}.go`:
- `// +kuma:policy:singular_display_name=...`: UI name
- `// +kuma:policy:skip_registration=true`: test-only policies
- `// +kuma:policy:scope=Mesh`: Mesh or Global scope

Field markers: `+kuma:discriminator` (union types), `+kuma:non-mergeable-struct`

## Plugin interface

Policies implement `core_plugins.PolicyPlugin`:
- `MatchedPolicies()`: finds policies applying to a dataplane (use `matchers.MatchedPolicies()`)
- `Apply()`: modifies Envoy xDS `ResourceSet` based on matched policies

Access matched policies in Apply: `proxy.Policies.Dynamic[api.PolicyType]`

## Generation pipeline

`make generate` runs per policy: `policy-gen core-resource` → `k8s-resource` → `plugin-file` → `helpers` → `openapi`

**Gotcha**: `tools/resource-gen` depends on `tools/policy-gen`, so modifying policy-gen forces resource-gen rebuild.
