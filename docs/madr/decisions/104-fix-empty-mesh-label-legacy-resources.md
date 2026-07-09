# Fix empty kuma.io/mesh label on legacy resources

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/17020

## Context and Problem Statement

Resources created before Kuma 2.9 (when `kuma.io/mesh` label computation was introduced) may have
an empty `kuma.io/mesh: ""` label stored in their metadata. This particularly affects:

- `MeshGatewayInstance` resources (also referred to as "Hosted Gateways" or HGs)
- `Secret` resources with `system.kuma.io/secret` type

### How the issue originated

1. Before 2.9, mesh-scoped resources were associated with a mesh via:
   - The `kuma.io/mesh` annotation (deprecated, now logs a warning)
   - Namespace-level `kuma.io/mesh` label
   - Fallback to `default` mesh

2. In 2.9, `ComputeLabels` was introduced to automatically set `kuma.io/mesh` on mesh-scoped
   resources at creation/update time (`pkg/core/resources/labels/compute.go`).

3. However, resources created before this change retained their original label state. If a
   resource had `kuma.io/mesh: ""` explicitly set (or was migrated with that value), it persists
   because:
   - The label is already present (albeit empty)
   - `setIfNotExist` in `ComputeLabels` skips labels that already exist
   - The resource is never updated, so recomputation never triggers

### Use cases affected

1. **API responses show empty mesh label**: When listing resources via the API, the response
   includes `labels: {"kuma.io/mesh": ""}`, which is confusing and inconsistent with newer
   resources that have the correct mesh value.

2. **Resource lookups may fail**: Code that queries resources by mesh label (e.g., `ListByMesh`)
   might not find resources with empty mesh labels, leading to "Resource not found" errors as
   reported in issue #9188.

3. **Downstream migration tooling**: External tooling (e.g., v3 compatibility scripts in
   downstream projects) needs to detect and fix these legacy resources.

## Design

### Option 1: Sanitize empty mesh label in API response layer

Add logic in the API response serialization to replace empty `kuma.io/mesh` values with the
actual mesh from `meta.mesh`. This provides a clean API response without modifying stored data.

**Implementation sketch:**

```go
// In pkg/api-server/oapi-helpers/helpers.go
func ResourceMetaToMeta(resType core_model.ResourceType, m core_model.ResourceMeta) api_common.Meta {
    labels := m.GetLabels()
    if labels == nil {
        labels = map[string]string{}
    }
    
    // Sanitize empty kuma.io/mesh label
    if labels[metadata.KumaMeshLabel] == "" && m.GetMesh() != "" {
        labels = maps.Clone(labels) // Don't modify original
        labels[metadata.KumaMeshLabel] = m.GetMesh()
    }
    
    return api_common.Meta{
        Type:   string(resType),
        Mesh:   m.GetMesh(),
        Name:   m.GetName(),
        Labels: labels,
    }
}
```

**Advantages:**
- Non-destructive: does not modify stored resources
- Immediate fix: all API consumers see correct data immediately
- Backward compatible: no migration required

**Disadvantages:**
- Stored data remains inconsistent with API response
- Does not fix the root cause in persistent storage
- May cause confusion if users inspect storage directly (e.g., kubectl get with -o yaml)
- Every API response incurs a small overhead for the check

### Option 2: Remove empty mesh label at read time in resource manager

Modify the resource manager's `Get`/`List` operations to strip or fix empty `kuma.io/mesh`
labels before returning resources. This normalizes data at a lower layer.

**Advantages:**
- All code paths see consistent data
- Single point of fix

**Disadvantages:**
- More invasive change
- Still doesn't fix stored data
- May have unintended side effects on cache behavior

### Option 3: Allow users to manually remove empty label

Document that users can remove the offending empty label themselves:

```bash
# Kubernetes
kubectl label meshgatewayinstance <name> kuma.io/mesh-

# Universal
kumactl get meshgatewayinstance <name> -o yaml | \
  yq 'del(.labels["kuma.io/mesh"])' | \
  kumactl apply -f -
```

The next reconciliation/update will compute the correct label via `ComputeLabels`.

**Advantages:**
- Simple, no code changes
- Users have control over when/if to fix
- Storage is corrected permanently

**Disadvantages:**
- Manual effort required
- Users need to know which resources are affected
- Doesn't help users who don't know about this issue

### Option 4: Add detection to v3 compatibility script (downstream)

The v3 compat script (maintained downstream) can include a check for resources with
`kuma.io/mesh: ""` and either warn or auto-fix.

**Advantages:**
- Part of formal upgrade path
- Can be combined with other migration checks
- Clear point in time when fix is applied

**Disadvantages:**
- Downstream dependency, not in Kuma core
- Users who don't run the script are not helped
- Only helps during v3 migration, not general users

### Option 5: Fix on read/write with automatic label recomputation

Trigger label recomputation when a resource is read and has an empty `kuma.io/mesh` label,
persisting the fix. This could be implemented as a one-time migration that runs on CP startup
or as a lazy fix on access.

**Advantages:**
- Permanently fixes stored data
- Self-healing system

**Disadvantages:**
- Modifying resources on read is unexpected behavior
- Write on read can cause issues with etcd watchers, caches
- May conflict with resource versioning/conflict detection
- Risky for multi-zone setups

### Recommended approach: Combine Options 1 and 3

1. **Option 1 (API sanitization)**: Implement API-layer sanitization so all consumers
   immediately see correct data. This is low-risk and provides immediate relief.

2. **Option 3 (Documentation)**: Document the manual fix for users who want to correct
   their stored resources permanently.

3. **Option 4 (v3 compat script)**: Add detection/fix to the downstream v3 compat script
   for users migrating to v3. This catches legacy resources during the major version upgrade.

## Security implications and review

N/A — This is a data normalization issue with no security implications. The mesh association
is already correctly determined via `meta.mesh`; only the label is inconsistent.

## Reliability implications

**Before fix:**
- Resources with empty `kuma.io/mesh` labels may not be found by label-based queries
- API responses show inconsistent/confusing data

**After fix:**
- API responses are consistent
- Label-based queries may still miss resources in storage until manually fixed or migrated
- Adding the option for users to fix their resources permanently eliminates the long-term issue

## Implications for Kong Mesh

The v3 compat script should include a check for resources with `kuma.io/mesh: ""`. This script
runs during enterprise upgrades and can warn users or auto-fix affected resources.

## Decision

We will implement the combined approach:

1. **API layer sanitization** (`pkg/api-server/oapi-helpers/helpers.go`): When serializing
   resource metadata for API responses, if `kuma.io/mesh` label is empty but `meta.mesh` is
   set, populate the label with the mesh value. This ensures all API consumers see consistent
   data without modifying storage.

2. **Documentation**: Add an entry to `UPGRADE.md` explaining that resources created before
   2.9 may have empty `kuma.io/mesh` labels and how to fix them manually.

3. **v3 compat script (downstream)**: Add detection for `kuma.io/mesh: ""` to the v3
   compatibility checking script, warning users to remove the empty label before upgrading.

## Notes

Open questions for human decision:

- [ ] **Auto-fix vs warn-only in compat script**: Should the v3 compat script automatically
  remove empty `kuma.io/mesh` labels, or only warn the user? Auto-fix is more convenient but
  modifies user resources; warn-only is safer but requires manual action.

- [ ] **Scope of API sanitization**: Should we sanitize only in `ResourceMetaToMeta` (API
  responses), or also in the REST/KDS layers? Broader sanitization ensures consistency but
  adds more code paths to maintain.

- [ ] **Which resource types to document**: The issue mentions MeshGatewayInstance, but
  Secrets (#9188) are also affected. Should we document all mesh-scoped resources, or focus
  on the known affected types?
