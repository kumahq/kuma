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

This MADR adds 2 kinds `to[].targetRef.kind` and Mesh*Route `backendRefs[].kind`:

* `MeshService`
* `MeshMultiZoneService`

as well as a number of new fields to the ref structure:

* `port`
* `namespace`
* `labels`
* `zone`

This MADR doesn't address using `MeshService` as a reference anywhere else.

Another option would be to have the `namespace` and `zone` fields under `labels` instead.
The motivation for having them top level is that they have especially important
semantic meaning, beyond other arbitrary, user-defined labels.

## Decision Outcome

These new kinds allow for selecting or directing traffic that's going to new,
`MeshService`-based destinations.

### `targetRef`

The policy is applied
on both the Envoy outbound listeners and clusters
that correspond to the `MeshService`.

References for `kind: MeshService` have the following structure:

```yaml
kind: MeshService
name: backend
labels: {}
zone: east
namespace: backend
port: 80
```

The `name` of a `MeshService` ref always refers to the name of the Kubernetes
object.

Exactly one of either `name` or `labels` is **required**.

#### `port`

If `port` is set in `targetRef.kind`, only traffic to that port is affected.

In `backendRefs`, `port` is required.

#### Local `MeshServices`

If the policy ref doesn't set `zone`, it refers to a `MeshService` in the local
zone.

On a k8s zone:
* without `namespace`, the policy refers to `MeshServices` in the same
  namespace as the policy.
* with `namespace`, only that namespace is searched matching `MeshServices`

Note that this means it's possible to refer to synced `MeshServices` via their
transformed `<service>-<hash-suffix>`.

#### Non-local `MeshServices`

If `zone` is set, `name` cannot be set and instead `displayName` must be set,
which matches based on the `kuma.io/display-name` of the `MeshService`.
For example, the name of the object in the zone's cluster.

Without `namespace`, refer to _all_ `MeshServices` with this name. That means we
don't infer any similarity between the local namespace and the other zone's
namespace.

#### Examples

##### `MeshService`

Users can target `MeshServices` directly by `name`:

```yaml
spec:
 to:
 - targetRef:
     kind: MeshService
     name: backend
```

or by using `labels` without `name`:

```yaml
spec:
 to:
 - targetRef:
     kind: MeshService
     labels:
       kubernetes.io/service-name: zk
```

In other zones:

```yaml
spec:
 to:
 - targetRef:
     kind: MeshService
     displayName: backend
     zone: other-zone
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
               port: 8080
               zone: east
             - kind: MeshMultiZoneService
               name: frontend
               port: 8080
```

#### Status

Similar to Gateway API conditions, we report back as status conditions
if a route targets a
`MeshService`/port tuple that doesn't exist.

### Positive Consequences

* Allows users to apply policy to traffic going to `MeshService` and
  `MeshMultiZoneService`
* We can match set of `MeshServices`, important for headless services
* We're flexible with regards to local vs non-local `MeshServices`.

### Negative Consequences

* Relative complexity of the semantics of various combinations
