# API Server Refactoring Roadmap

## Goal

Simplify Kuma's descriptor-driven API server without changing its public HTTP contracts or replacing its routing, resource descriptor, or generated validation architecture.

## Architecture Summary

- `pkg/api-server/server.go:addResourcesEndpoints` expands registered resource descriptors into routes and constructs handlers.
- `pkg/api-server/resource_crud_endpoints.go:resourceCrudHandler` owns generic CRUD, overview merging, filtering, labels, and request validation.
- `pkg/api-server/resource_inspect_endpoints.go:resourceInspectHandler` owns policy and rules inspection.
- Generated resource `rest.yaml` schemas are embedded and used for runtime defaulting and structural validation. They are not documentation-only.
- API behavior is protected primarily by integration tests and JSON goldens under `pkg/api-server/testdata`.

## Guardrails

- Preserve paths, aliases, route order, status codes, payloads, error titles, media types, authentication/filter order, and validation precedence.
- Preserve explicit read-only PUT/DELETE routes returning `405`.
- Preserve descriptor-driven decoding, generated defaults and validation, and legacy flattened resources.
- Keep each implementation PR at seven changed files or fewer.
- Keep behavior changes separate from structural refactoring.
- Do not hand-edit `zz_generated.*`, `rest.yaml`, or `docs/generated/openapi.yaml`.
- Run generation only when OpenAPI, proto, or resource source definitions change.

## Correctness PRs

### 1. Reject Malformed Label Filters Safely

- Fix unsafe parsing in `pkg/api-server/filters/filtering.go`.
- Return validation errors for unmatched brackets, empty values, duplicate values, and unsupported syntax instead of panicking.
- Preserve valid label and status filtering.
- Add focused package tests for exact accepted and rejected query forms.

### 2. Fix API Configuration Error Aggregation

- Correct `multierr.Append(errs, ...)` usage in `pkg/config/api-server/config.go`.
- Characterize and preserve deterministic validation error ordering.
- Keep HTTPS port early-return behavior separate unless intentionally changed.

### 3. Apply TLS Maximum Version

- Apply `TlsMaxVersion` in `pkg/api-server/server.go:configureTLS`.
- Test a TLS 1.2 maximum rejecting a TLS 1.3-only client.
- Test that an empty maximum retains Go's unrestricted default.

### 4. Make Listener Startup Transactional

- Prepare HTTP and HTTPS server configuration before starting either listener.
- Use non-blocking/buffered error delivery.
- Roll back every started listener after synchronous or asynchronous sibling failure.
- Reset readiness on failure and stop.
- Use bounded shutdown rather than `context.Background()`.
- Test an occupied HTTPS port after HTTP starts and verify HTTP is closed.

### 5. Align KRI Kubernetes Mapping

- Give `kriEndpoint` the same descriptor-aware mapper selection as ordinary resource reads.
- Test `Secret` and `GlobalSecret` parity for Kubernetes and inference mappers, explicit namespaces, and the default system namespace.

### 6. Resolve Kumactl Policy Inspection Compatibility

- Prefer the current `_resources/dataplanes` endpoint and define legacy fallback behavior explicitly.
- Decide whether pagination follows `next` or exposes size/offset.
- Lock output compatibility with command/client tests.
- Update `UPGRADE.md` when behavior is selected.

## Structural Refactoring PRs

### 7. Characterize HTTP Contracts

- Add direct tests for `handle`, `withTitle`, `created`, `rawResponse`, nil bodies, title overrides, statuses, and media types.
- Add tests for the current read-only PUT/DELETE `405` behavior rather than reusing stale orphaned goldens.
- Lock CRUD decoding, lookup, validation, authorization, and error precedence.

### 8. Standardize Remaining Simple Handlers

- Migrate `addIndexWsEndpoints`, `addPoliciesWsEndpoints`, and `addWhoamiEndpoints` to `handle`.
- Retain existing handwritten DTOs where generated models differ in requiredness, pointers, integer widths, or `omitempty` behavior.
- Preserve index authentication metadata and exact JSON/media-type behavior.
- `globalInsightsEndpoints` is already migrated and is out of scope.

### 9. Split CRUD by Responsibility

- Move read/list/overview behavior into `resource_read_endpoints.go`.
- Move create/update/delete and write hooks into `resource_write_endpoints.go`.
- Move request and label validation into `resource_validation.go`.
- Preserve existence lookup before metadata validation, authorization timing, generated validation, labels, and error titles.
- Move code first; do not combine the split with new abstractions or error behavior.

### 10. Simplify CRUD Internals

- Only extract helpers after the structural split exposes proven duplication.
- Keep `SetSpec` error handling as a separate correctness change with explicit HTTP semantics and no-write tests.
- Avoid a generic lifecycle-hook framework; current type-specific maps are sufficient.

### 11. Consolidate Policy Inspection

- Share authorized Dataplane loading and base Mesh context construction between `getPoliciesConf` and `rulesForResource`.
- Extract one `matchPolicies` path and use `baseMeshContext.Resources()` consistently.
- Preserve policy plugin lookup timing and order.
- Preserve deterministic inbound, resource-rule, and HTTP-match sorting.
- Preserve initialized empty arrays in API responses.
- Split only pure response mapping blocks with focused tests.

### 12. Simplify Route Registration

- Model route roles explicitly: mesh CRUD/list, cross-mesh list, global CRUD/list, primary path, and alias.
- Replace duplicated scope and alias branches in `addResourcesEndpoints`.
- Preserve route order, route metadata, read-only handlers, and the existing `GlobalSecret` alias.
- Use a cycle-safe package or keep the route projection private to API-server code.

### 13. Decide Catalog Read-Only Semantics

- Audit consumers before changing behavior: routes and `/policies` currently differ from `/_resources`.
- Treat catalog alignment as an explicit API behavior change.
- Keep federation inclusion independent from effective route writability.
- Test Global, federated Zone, non-federated Zone, explicit API read-only, Kubernetes and Universal environments, and KDS provider flags.

## Verification

Run for every API-server PR:

```bash
make test TEST_PKG_LIST=./pkg/api-server/...
make format
make check
```

Add config, mapper, kumactl, REST-model, or policy tests when those packages are touched. Run full `make test` after the complete sequence. Run API E2E tests only for externally visible routing, authentication, or format changes.

Stop and split the work if a structural PR changes a golden, route set/order, content type, status, payload, or error precedence without an explicitly approved behavior change.

## Open Decisions

1. Should `/_resources` follow route writability or preserve its current behavior?
2. Should kumactl use current-first legacy fallback or remove legacy support?
3. Should `ApiServer.Start` support restarting after failure or shutdown?
