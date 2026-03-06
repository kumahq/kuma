# Contents

1. [Policy-based features](#policy-based-features)
2. [Non-policy features](#non-policy-features)
3. [What to read for each group type](#what-to-read-for-each-group-type)
4. [Variant detection signals](#variant-detection-signals)
5. [Schema verification sources](#schema-verification-sources)
6. [Tips](#tips)

---

# Code reading guide

Where to find things in a Kuma repo for generating test suites.

## Policy-based features

| What                          | Path                                                                |
| :---------------------------- | :------------------------------------------------------------------ |
| Policy spec (struct, markers) | `pkg/plugins/policies/<name>/api/v1alpha1/<name>.go`                |
| Validation logic              | `pkg/plugins/policies/<name>/api/v1alpha1/validator.go`             |
| Deprecation warnings          | `pkg/plugins/policies/<name>/api/v1alpha1/deprecated.go` (optional) |
| xDS generation                | `pkg/plugins/policies/<name>/plugin/v1alpha1/plugin.go`             |
| Test golden files             | `pkg/plugins/policies/<name>/plugin/v1alpha1/testdata/`             |
| K8s types                     | `pkg/plugins/policies/<name>/k8s/v1alpha1/`                         |
| CRDs                          | `deployments/charts/kuma/crds/`                                     |

## Non-policy features

Start from changed files (PR diff or branch diff against master). Common locations:

| Area                   | Path                            |
| :--------------------- | :------------------------------ |
| Core resource types    | `pkg/core/resources/apis/mesh/` |
| Common API types       | `api/common/v1alpha1/`          |
| Mesh proto definitions | `api/mesh/v1alpha1/`            |
| xDS config generation  | `pkg/xds/`                      |
| KDS sync               | `pkg/kds/`                      |
| Control plane runtime  | `pkg/plugins/runtime/`          |
| Data plane config      | `pkg/config/app/kuma-dp/`       |
| Transparent proxy      | `pkg/transparentproxy/`         |
| REST API               | `pkg/api-server/`               |

## What to read for each group type

| Suite group        | Read from code                                                                       |
| :----------------- | :----------------------------------------------------------------------------------- |
| G1 CRUD            | API spec struct fields, CRD schema                                                   |
| G2 Validation      | `validator.go` - every `Err()` call is a rejection case                              |
| G3 Runtime config  | `plugin.go` - what xDS resources get generated, what config dump sections to inspect |
| G4 E2E flow        | Test golden files - expected xDS output, plus traffic generation patterns            |
| G5 Edge cases      | Validator edge cases, nil/empty handling in plugin.go                                |
| G6 Multi-zone      | KDS sync markers on the resource type, `pkg/kds/` for sync behavior                  |
| G7 Backward compat | `deprecated.go`, old field names in API spec, migration notes                        |

## Variant detection signals

While reading code for G1-G7, also scan for patterns that indicate feature variants. Variants expand the suite with G8+ groups.

| Signal                 | Where to look                                                 | What it means            |
| :--------------------- | :------------------------------------------------------------ | :----------------------- |
| S1 Deployment topology | KDS markers, resource registration, `pkg/kds/`                | Multi-zone groups needed |
| S2 Feature modes       | Enum/string fields in API spec, switch/case in plugin.go      | Per-mode groups          |
| S3 Backend variants    | Multiple backend types in spec struct, `backendRef` kinds     | Per-backend groups       |
| S4 Feature flags       | Conditional branches in `Apply()` checking flags/config       | Per-flag groups          |
| S5 Policy roles        | `targetRef` section, producer/consumer/workload-owner markers | Role-specific groups     |
| S6 Protocol variants   | HTTP/TCP/gRPC branching in plugin.go                          | Per-protocol groups      |
| S7 Backward compat     | `deprecated.go`, old field names, version-gated logic         | Legacy path groups       |

See `references/variant-detection.md` for the full methodology, signal strength classification, and a worked MOTB example.

## Schema verification sources

Every generated manifest (Kubernetes or Universal) must be verified against the actual schema before inclusion in the suite. Common mistakes: wrong field names, wrong namespace, wrong enum values, missing required fields.

### Kubernetes manifests

| Check | Source | How to verify |
| :---- | :----- | :------------ |
| apiVersion/kind | CRD in `deployments/charts/kuma/crds/kuma.io_<plural>.yaml` | `spec.group` + `spec.versions[].name` |
| Scope (Namespaced/Cluster) | CRD `spec.scope` field | Namespaced = needs `metadata.namespace`; Cluster = no namespace |
| Field names | CRD `openAPIV3Schema.properties.spec` tree OR Go struct JSON tags | YAML field = JSON tag value (e.g., `json:"targetRef"` = `targetRef:` in YAML) |
| Required fields | CRD `required` arrays at each level | Non-pointer Go fields without `omitempty` are required |
| Enum values | CRD `enum` arrays OR `+kubebuilder:validation:Enum=X;Y;Z` markers | Only listed values are valid |
| Label requirements | `+kuma:policy:scope=Mesh` marker + CRD | Mesh-scoped resources need `kuma.io/mesh` label |

### Universal manifests

| Check | Source | How to verify |
| :---- | :----- | :------------ |
| Resource type | Go resource type registration | `type` field matches registered resource kind |
| Mesh field | `+kuma:policy:scope=Mesh` marker | Mesh-scoped resources need `mesh` field at top level |
| Spec fields | Go API spec struct JSON tags | Same field name rules as Kubernetes |
| Enum values | `+kubebuilder:validation:Enum` markers | Same values apply in Universal format |

### Common pitfalls

- Go field `TargetRef *common_api.TargetRef` with tag `json:"targetRef,omitempty"` means the YAML key is `targetRef` (camelCase), not `target_ref` or `TargetRef`
- Pointer fields (`*Type`) are optional; non-pointer fields are required
- `omitempty` means the field can be absent; without it, the field must be present even if zero-valued
- `kuma.io/mesh` label goes in `metadata.labels` for Kubernetes, `mesh` field at top level for Universal
- Policy scope is NOT the same as CRD scope - a policy can be Mesh-scoped (conceptually) but Namespaced (in K8s CRD scope)
- `targetRef.kind: Dataplane` with labels targets individual dataplanes; `targetRef.kind: MeshGateway` targets builtin gateways only

### Verification workflow

For each manifest generated for the suite:

1. Read the CRD file (K8s) or Go API spec (Universal) for the resource kind
2. Walk every field in the manifest and confirm it exists in the schema at that path
3. Check enum fields against allowed values
4. Confirm namespace/mesh placement matches the resource scope
5. Confirm required fields are present
6. For `targetRef` sections, verify the `kind` value is valid for the policy type

## Tips

- Golden files in `testdata/` show exact expected Envoy configs - use these to derive validation commands.
- The `+kuma:policy` markers on the spec struct indicate scope (Mesh vs Global), display name, etc.
- `validator.go` returns `admission.Warnings` for deprecations - these become G7 test cases.
- `plugin.go`'s `Apply()` method reveals which xDS resource types are affected (listeners, clusters, routes, endpoints).
- When the code references delegated gateways (`IsDelegatedGateway()`, `gateway.type: DELEGATED`), the test workload should be Kong Gateway - not nginx or a generic proxy. See `references/suite-structure.md` (Domain knowledge) for setup details.
