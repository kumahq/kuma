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
* Add additional kind `NonLocalMeshService` that explicitly selects non-local
  `MeshServices` via a structured set of fields `zone`, `port`, `labels`,
  `namespace`
* Add an additional kind `MeshServiceId` to support selecting an arbitrary
  `MeshService` via a kind of UID like `kuma://backend/zone-1/backend-ns/8080`

### Pros and Cons

#### `labels` vs top level `zone`/`namespace`/`name`

Top level zone would look like the following:

```yaml
kind: NonLocalMeshService
namespace: other-zone-ns
zone: other-zone
name: name-in-its-zone
```

Using `zone` top level could allow us to be "smarter" about selecting
cross-zone `MeshServices`. Instead of matching the actual Kubernetes object
we could transparently match `name`/`namespace` to `kuma.io/display-name`/`kuma.io/namespace`, instead of
requiring:

```yaml
kind: NonLocalMeshService
labels:
  kuma.io/zone: other-zone
  kuma.io/display-name: name-in-its-zone
  k8s.kuma.io/namespace: other-zone-ns
```

which is significantly more noisy. Using `labels` is less "magic" and doesn't
look like we're matching an object that doesn't actually exist in the cluster.

##### Positive

Allows us to be explicit as well as not back ourselves into a corner
with the API.

##### Negative

* more verbose
* not always directly matching Kubernetes objects

#### `sectionName`

##### Positive

* consistency with Gateway API

##### Negative

* too generic, potentially confusing

#### `NonLocalMeshService`

```yaml
kind: NonLocalMeshService
namespace: other-zone-ns
zone: other-zone
```

##### Positive

* sidesteps issues of the k8s namespace not being the same as the original k8s
  namespace

##### Negative

* one more additional kind

#### `MeshServiceId`

```yaml
kind: MeshServiceId
name: kuma://backend/zone-1/backend-ns/8080
```

##### Positive

* least verbose of the options

##### Negative

* one more additional kind
* not structured

## Decision Outcome

We choose:

* use generic Gateway API targetRef field `sectionName` to specify `port` and
  add `labels` but use `port` in `backendRefs`.
* use exactly one of `labels` or `name`/`namespace`
* no new top level fields besides `labels`

### `targetRef`

The policy is applied
on both the Envoy outbound listeners and clusters
that correspond to the `MeshService`.
Note that more than one `MeshService` can be matched in which case it applies to
each matched `MeshService`.

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
  name: backend # only name/namespace or labels
  labels: # only name/namespace or labels
    kuma.io/zone: east
  sectionName: http
```

The `name` of a `MeshService` ref always refers to the name of one, real Kubernetes
object in the zone of the policy.

Exactly one of either `targetRef.name` or `targetRef.labels` is **required**.

#### `sectionName`

If `sectionName` is set in `targetRef.kind` with `kind: MeshService`,
it refers to an entry in `ports` by name and only traffic to that port is affected.

In `backendRefs`, `port` is required. There is no `sectionName`, in accordance
with Gateway API HTTPRoute.

#### `name`

If the policy ref sets `name`, it refers to one `MeshService` in the *local
zone*.

On a k8s zone:
* without `namespace`, the policy refers to `MeshServices` in the same
  namespace as the policy.
* with `namespace`, only that namespace is searched for matching `MeshServices`

On a universal zone, `namespace` is not valid.

Note that referring to synced `MeshServices` using transformed `<service>-<hash-suffix>` name
shouldn't be possible. 
Using hashed name in the ref would make it extremely difficult to implement Inspect API on Global.
For example, MeshTimeout is synced from Zone to Global, and it refers to the `targetRef.name: <value>`. 
There is no easy way to process `<value>` to figure out what resource on Global it refers to.
When `name/namespace` are referring to the `<service>-<hash-suffix>`, the policy is accepted but has no effect
as the search for `name/namespace` resource is happening only across locally-originated resources.

##### Backwards compatibility

If there is no `MeshService` with matching `name`, it is interpreted as
a legacy style `kuma.io/service`.

#### `labels`

If `name` is not set, `labels` must be set and `namespace` cannot be set!

Via `labels`, groups of `MeshServices` can be matched, for example by:

* `kuma.io/display-name`
* `kuma.io/zone`

With `labels`, we don't assume anything about the namespace of the object. The
namespace _must_ be selected using the `k8s.kuma.io/namespace` label.

This also means more verbosity for the headless service usecase.
It means we don't infer any similarity between the local namespace
and the other zone's namespace.

#### Examples

These examples assume the policy is in the `frontend` namespace and the zone named `local-zone`.

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

###### Multi-zone

In order to refer to objects via their display name, `labels` must be used.
In general, `labels` match _any_ `MeshService`, from any namespace or zone.

Match a specific name in a specific namespace in a specific zone:

```yaml
spec:
 to:
 - targetRef:
     kind: MeshService
     labels:
       k8s.kuma.io/service-name: zk
       k8s.kuma.io/namespace: zk-namespace # this must be set to target a namespace, even the local one
       kuma.io/zone: local-zone # must also be set to target a local zone, otherwise all zones are targeted
```

Or match a name in _any namespace_ in a specific zone:

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
               port: 8080
               labels:
                 kuma.io/display-name: frontend
                 kuma.io/zone: east
```

**NOTE**: `backendRef` must not match more than one `MeshService`.

However, whether or not a `backendRefs[].labels` could match more than one
`MeshService` is something that depends on the type of the referenced zone.
For example, if the above zone `east` is a Kubernetes zone, it's not clear
which namespaces to match.

However, in that case the above `backendRef` must _always_ be ignored, even if only
one MeshService would be matched.
Otherwise the simple creation of a second `MeshService` would break a previously working route.
We can determine this by detecting whether or not the referenced zone is a Kubernetes
zone.

In this case, the route should be applied to configuration as if no `MeshServices`
were matched by this ref.

#### Status

Similar to Gateway API conditions, we report back as status conditions
on the policy/route if a policy/route targets a
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
