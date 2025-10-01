# Add KRI to REST API Resources

* Status: Accepted

Technical Story: https://github.com/kumahq/kuma/issues/14462
GUI Issue: https://github.com/kumahq/kuma-gui/issues/4220

## Context and Problem Statement

The GUI needs a stable, machine-generated identifier to link any resource or policy from the REST API with the origin that produced its Envoy configuration. We already use KRI (Kuma Resource Identifier) in new components and in the Inspect API, where per-rule origins are exposed. But most REST list/get endpoints do not return the KRI field. Clients must reconstruct it manually, which is fragile and error-prone.

We need a lightweight way to expose KRI across REST API responses without major refactors or storage changes.

## Decision Drivers

* Changes should be minimal, additive, and safe for existing clients
* Avoid unnecessary work on legacy protobuf-based resources that are being deprecated
* Make KRI reliably available to GUI and tooling
* Minimize schema churn and developer confusion

## Options Considered

### Option A: Add a `kri` field to all resource/policy structs

**Pros**

* Explicit in types
* OpenAPI can immediately document it

**Cons**

* KRI is derived, not stored - we must recompute it everywhere
* Requires touching many code paths
* High risk of inconsistencies or bugs

### Option B: Compute and inject `kri` during JSON marshalling of `ResourceMeta` (chosen)

**Idea**

* In `pkg/core/resources/model/rest/v1alpha1/ResourceMeta`, override JSON marshalling to insert a computed `kri` based on `type`, `mesh`, `labels` and `name`.

**Pros**

* Minimal code changes
* Automatically appears wherever `ResourceMeta` is used
* Clients see it without needing to modify storage
* No problem with OpenAPI schema generation since our templates can be adjusted to include it

**Cons**

* The field is computed and not settable in code

### Option C: Add a `KRI` field to `ResourceMeta`, but only set it at marshalling time

**Pros**

* Easier to annotate as `readOnly` in schemas

**Cons**

* Confusing to developers (should not be set manually)
* Still needs extra scaffolding for clarity and safety

## Decision

We choose Option B: compute and inject `kri` during JSON marshalling of `ResourceMeta`. This meets our goal of minimal, safe change, while making KRI available in REST API responses without touching storage or reconciliation. Our OpenAPI templates will be updated to include the field and mark it as `readOnly`. This also aligns with how KRI is used in the Inspect API for per-rule origin.

## Design

```go
type ResourceMeta struct {
    Type             string            `json:"type"`
    Mesh             string            `json:"mesh,omitempty"`
    Name             string            `json:"name"`
    CreationTime     time.Time         `json:"creationTime"`
    ModificationTime time.Time         `json:"modificationTime"`
    Labels           map[string]string `json:"labels,omitempty"`
}

func (r ResourceMeta) MarshalJSON() ([]byte, error) {
    type Alias ResourceMeta
    return json.Marshal(&struct {
        Alias
        KRI string `json:"kri,omitempty"`
    }{
        Alias: Alias(r),
        KRI:   kri.FromResourceMeta(r, core_model.ResourceType(r.Type)).String(),
    })
}
```

* `KRI` is derived and never stored
* Anywhere `ResourceMeta` is rendered in a REST response, `meta.kri` will be included
* No changes to storage, reconciliation, or resource logic

## Scope and Legacy Resources

* We will not retrofit full KRI support into legacy protobuf-based resources beyond what [Inspect API](./075-inspect-api-redesign.md) already covers
* Adding KRI to legacy resources would require defining short names and extra logic, which we intend to remove in version 3.0.0
* Only the new Inspect API endpoints can return an origin resource by its KRI used in the GUI. Legacy resources will not appear as origins

## Compatibility

* The change is additive in JSON. Clients ignoring unknown fields continue to work
* No change to storage, hashing, reconciliation, or system behavior

## API Schema and OpenAPI

* OpenAPI templates will be updated so `meta.kri` is documented as `string` and `readOnly: true`
* Clients will see it in generated SDKs

## Security

* No new secrets or internal state are exposed
* KRI is derived from `type`, `mesh`, `labels`, and `name`, which are already public in the API

## Performance and Reliability

* The overhead is minimal (computing a small string during JSON marshalling)
* No extra database calls or caching changes
* No impact on performance or stability

## Migration

* No migrations required for users
* GUI should begin reading `meta.kri` when present
* Fallback logic: if `meta.kri` is absent, GUI may still reconstruct it temporarily during rollout

## Acceptance Criteria

* REST API responses now include `meta.kri`
* The value matches Inspect API KRI for the same resource
* GUI no longer needs to reconstruct KRI
* No breakage observed in existing API clients

## Implication for Kong Mesh

There are no product-specific implications for Kong Mesh. The change is purely additive to the REST API and affects only the representation of resources. Kong Mesh behavior, configuration, and features remain the same. Existing clients continue to work without modification, while new clients can optionally take advantage of the `meta.kri` field.
