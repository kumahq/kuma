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

## Tips

- Golden files in `testdata/` show exact expected Envoy configs - use these to derive validation commands.
- The `+kuma:policy` markers on the spec struct indicate scope (Mesh vs Global), display name, etc.
- `validator.go` returns `admission.Warnings` for deprecations - these become G7 test cases.
- `plugin.go`'s `Apply()` method reveals which xDS resource types are affected (listeners, clusters, routes, endpoints).
- When the code references delegated gateways (`IsDelegatedGateway()`, `gateway.type: DELEGATED`), the test workload should be Kong Gateway - not nginx or a generic proxy. See `references/suite-structure.md` (Domain knowledge) for setup details.
