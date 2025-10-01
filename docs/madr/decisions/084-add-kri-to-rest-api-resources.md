# Add KRI to REST API Resources

* Status: Accepted

Technical Story: https://github.com/kumahq/kuma/issues/14462
GUI Issue: https://github.com/kumahq/kuma-gui/issues/4220

## Context and Problem Statement

The GUI needs a stable, machine-generated identifier to link any resource or policy from the REST API with the origin that produced its Envoy configuration. We already use KRI (Kuma Resource Identifier) in new components and in the Inspect API, where per-rule origins are exposed. But most REST list/get endpoints do not return the KRI field. Clients must reconstruct it manually, which is fragile and error-prone.

This task focuses only on **changing the REST API responses** to include KRI. It does not cover modifying resource stores or managers such as Kubernetes, In-Memory, or Postgres.

Having KRI available directly in CRD resources returned by the Kubernetes API, or more generally in resources returned by any store/manager, would also be useful. However, that would likely require storing KRI as part of the resource itself. Storing KRI comes with drawbacks: it depends on values that are already part of the resource (`name`, `mesh`, `labels`) as well as the resource type. This means that every time those values change, the stored KRI would have to be recalculated and updated. Ensuring this consistency would require many code changes and additional design effort.

Because of that complexity, storing KRI in backends is not included in the scope of this MADR. The current decision only addresses returning KRI in REST API responses by computing it at marshalling time. At the same time, this MADR does not prevent us from exploring persistent KRI storage in the future if it becomes valuable. That broader discussion will be handled in a separate MADR.

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

### Scope and Legacy Resources

* Resources defined in protobuf that are part of the core data model, such as `Dataplane`, `Mesh`, `ZoneIngress`, and `ZoneEgress`, are **not** considered deprecated or legacy. When returned by the REST API, they will also include the `kri` field
* We will not retrofit full KRI support into truly legacy protobuf-based policies that are being phased out. Those remain covered only by what the [Inspect API](./075-inspect-api-redesign.md) already provides
* Adding KRI to deprecated policies would require defining short names and introducing additional logic, which we plan to remove in version 3.0.0
* Only the new Inspect API endpoints can return an origin resource by its KRI for use in the GUI. Deprecated policy types will not appear as origins

### Compatibility

* The change is **additive**: a new `kri` field is included in JSON responses. Clients that ignore unknown fields will continue to work without modification.
* There is **no impact** on storage, reconciliation, hashing, or overall system behavior.
* OpenAPI templates are updated so that `meta.kri` is fully documented as a `string` with `readOnly: true`. This ensures the field is visible in generated specifications and SDKs.
* SDKs and tooling generated from the OpenAPI schema will include the `kri` field, allowing typed clients to consume it directly.

Schema updates include:

```patch
diff --git a/tools/openapi/templates/schema.yaml b/tools/openapi/templates/schema.yaml
--- a/tools/openapi/templates/schema.yaml
+++ b/tools/openapi/templates/schema.yaml
@@ -11,6 +11,15 @@
 type: string
 default: default
 {{- end}}
+  {{- if ne .ShortName ""}}
+  kri:
+    description: 'A unique identifier for this resource instance, derived from
+      resource attributes such as type, mesh, and name. Used by internal tooling
+      and integrations for cross-references or indexing'
+    type: string
+    readOnly: true
+    example: 'kri_{{ .ShortName }}_default__kuma-system_{{ .ShortName }}123_'
+  {{- end}}
   name:
   description: 'Name of the Kuma resource'
   type: string
```

```patch
diff --git a/api/openapi/specs/common/resource.yaml b/api/openapi/specs/common/resource.yaml
--- a/api/openapi/specs/common/resource.yaml
+++ b/api/openapi/specs/common/resource.yaml
@@ -61,6 +61,11 @@
           type: string
           example: my-resource
           description: the name of the resource
+        kri:
+          type: string
+          readOnly: true
+          example: kri_mtp_default_zone1_kuma-system_mtp_
+          description: a unique identifier for this resource (KRI)
         labels:
           type: object
           additionalProperties:
```

Example of how the field appears in a policy schema:

```patch
diff --git a/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1/schema.yaml b/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1/schema.yaml
--- a/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1/schema.yaml
+++ b/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1/schema.yaml
@@ -10,6 +10,11 @@
     description: 'Mesh is the name of the Kuma mesh this resource belongs to. It may be omitted for cluster-scoped resources.'
     type: string
     default: default
+  kri:
+    description: 'A unique identifier for this resource instance, derived from
+      attributes such as type, mesh, and name. Useful for cross-references and
+      indexing by tooling or integrations'
+    type: string
+    readOnly: true
+    example: 'kri_mtp_default__kuma-system_mtp123_'
   name:
     description: 'Name of the Kuma resource'
     type: string
```

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
