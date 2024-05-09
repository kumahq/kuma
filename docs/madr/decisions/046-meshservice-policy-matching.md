# Policy matching with MeshService

* Status: accepted

Technical Story: #9707

## Context and Problem Statement

Now that MeshService exists, we need to enable users to apply policies in a way that works
with the new `MeshService` resource. This includes using `StatefulSet`.
Additionally, policy should be appliable to `MeshMultiZoneService`.
The inspect API should also reflect policies that use `MeshService`
which will be covered in a separate MADR.

Some things to remember:

* Currently `kind: MeshService` in `from[].targetRef` is deprecated.
* While `kind: MeshService` in `spec.targetRef` points to both a Dataplane and a set of inbounds,
  it is not a canonical reference.
  Its effects effectively mix with any policies applied
  to any `MeshService` that points to this `Dataplane`/inbound.
  A canonical reference can only be a direct choice of an entire `Dataplane` object
  and potentially a subset of inbounds.
* A `MeshService` only has a 1-1 relationship when used as `to[].targetRef.kind`
  since then it corresponds to a particular Envoy outbound listener/some number of clusters.

## Decision Drivers

* Targeting headless `MeshService`
* Targeting `MeshMultiZoneService`
* A balance between correctness and ease of use when configuring policy.

## Considered Options

This MADR only consider references in `to[].targetRef.kind` and MeshHTTPRoute/MeshTCPRoute `backendRefs[].kind`.

This MADR adds 1 new kind, `MeshMultiZoneService` and adds support
for matching real `MeshService` resources to the existing `MeshService` kind.
Matches on real `MeshService` objects have priority over `kuma.io/service`
matches.

A majority of the use cases for policy will involve selecting a specific *port* of the
`MeshService`, `MeshServices` from a given Kuma zone and namespace.

Note, we have already in targetRef:

```yaml
name: ...
namespace: ...
```

Now we have to consider how we attach to a port and select from a zone.

* Add top level field `zone` and `port` and `labels` for selecting groups of
  `MeshService`
* Add top level `port` field and use `labels: {"kuma.io/zone": "..."}`
* Use generic Gateway API targetRef field `sectionName` to specify a port in
  `targetRef` but `port` in `backendRef` and add `labels` for selecting
  groups of `MeshService`.

## Decision Outcome

We choose:

* use generic Gateway API targetRef field `sectionName` to specify `port` and
  add `labels` but use `port` in `backendRefs`.

### `targetRef`

The policy is applied
on both the Envoy outbound listeners and clusters
that correspond to the `MeshService`.

References for `kind: MeshService` have the following structure:

```yaml
---
kind: MeshService
spec:
  ports:
  - port: 80
    name: http
---
targetRef:
  kind: MeshService
  namespace: backend
  name: backend # only name or labels
  labels: # only name or labels
    kuma.io/zone: east
  sectionName: http
```

The `name` of a `MeshService` ref always refers to the name of one, real Kubernetes
object in the zone of the policy.

Exactly one of either `targetRef.name` or `targetRef.labels` is **required**.

#### `sectionName`

If `sectionName` is set in `targetRef.kind` with `kind: MeshService`,
it refers to an entry in `ports` by name and only traffic to that port is affected.

In `backendRefs`, `sectionName` is required.

#### `name`

If the policy ref sets `name`, it refers to one `MeshService` in the *local
zone*.

On a k8s zone:
* without `namespace`, the policy refers to `MeshServices` in the same
  namespace as the policy.
* with `namespace`, only that namespace is searched for matching `MeshServices`

On a universal zone, `namespace` is not valid.

Note that this means it's possible to refer to synced `MeshServices` via their
transformed `<service>-<hash-suffix>`.

#### `labels`

If `name` is not set, `labels` must be set.
Via `labels`, groups of `MeshServices` can be matched, for example by:

* `kuma.io/display-name`
* `kuma.io/zone`

If `kuma.io/zone` is set, a missing `namespace` field means we
refer to _all_ `MeshServices` with this name.
If no `kuma.io/zone` is set, the `namespace` is assumed to be the namespace of
the policy making the reference.

This means we don't infer any similarity between the local namespace
and the other zone's namespace.

#### Examples

##### `MeshService`

Users can target `MeshServices` in the namespace of the policy
directly with `name`:

```yaml
spec:
 to:
 - targetRef:
     kind: MeshService
     name: backend
```

or by using `labels` without `name`. The below example searches _only in the
namespace_ of the policy.

```yaml
spec:
 to:
 - targetRef:
     kind: MeshService
     labels:
       kubernetes.io/service-name: zk
```

In other zones, we search _all namespaces_:

```yaml
spec:
 to:
 - targetRef:
     kind: MeshService
     labels:
        kuma.io/display-name: backend
        kuma.io/zone: other-zone
```


##### `backendRef`

The structure of `backendRef` for `MeshService` is largely similar except that
we *require* a specific port.

```yaml
kind: MeshHTTPRoute
spec:
  targetRef:
    kind: MeshGateway
    name: edge-gateway
  to:
   - targetRef:
       kind: Mesh
     rules:
       - matches:
           - path:
               type: Prefix
               value: /v1
         default:
           backendRefs:
             - kind: MeshService
               name: frontend
               port: http
               labels:
                 kuma.io/zone: east
             - kind: MeshMultiZoneService
               name: frontend
               port: http
               labels:
                 kuma.io/zone: east
```

#### Status

Similar to Gateway API conditions, we report back as status conditions
if a route targets a
`MeshService`/port tuple that doesn't exist.

### Positive Consequences

* Allows users to apply policy to traffic going to `MeshService` and
  `MeshMultiZoneService`
* We can match a set of `MeshServices`, important for headless services
* We're flexible with regards to local vs non-local `MeshServices`.
* We diverge the least from upstream `targetRef`
* Adding fewer API fields

### Negative Consequences

* Relative complexity of the semantics of various combinations
* Verbosity in `labels`
