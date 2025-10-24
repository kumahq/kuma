# Policy On Namespace

- Status: accepted

## Context and Problem Statement

Kuma 2.6.0 enable users to apply policies on Zone CP.
This change unblocks us to introduce namespace support for policies.
The CRDs for new policies are already namespace-scoped, but they can be applied only on the `kuma-system` namespace.

## Decision Drivers

* Kubernetes-native UX, all app-related resources should be applied in the app's namespace
* Ability to use Kubernetes RBAC with Kuma policies
* It's not sustainable to grant access to `kuma-system` namespace to everyone who works with Kuma policies
* Service owners should have a way to affect Envoy configuration of their clients without applying policies to `kuma-system`

## Considered Options

* implicitly convert policy's namespace to `k8s.kuma.io/namespace` tag for top-level targetRef
* adopt producer/consumer concept from [GAMMA](https://gateway-api.sigs.k8s.io/mesh/#gateway-api-for-service-mesh)

## Decision Outcome

- adopt producer/consumer concept from GAMMA

## Pros and Cons of implicitly converting policy's namespace to `k8s.kuma.io/namespace` tag for top-level targetRef

* Good, because it easy to implement
* Good, because it works similar to zone-originated policies (we implicitly add `kuma.io/zone` tag for top-level targetRef)
* Bad, because it doesn't remove necessity to create policies in `kuma-system` namespace

## Pros and Cons of adopting producer/consumer concept from GAMMA

* Good, because it's familiar to users who know GAMMA
* Good, because creating policies in `kuma-system` becomes an edge use-case, the vast majority of use cases is covered by namespace-scoped policies
* Bad, because the logic of syncing policies cross zones can be complex

## Implementation

### Producers and consumers

[GAMMA](https://gateway-api.sigs.k8s.io/mesh/#gateway-api-for-service-mesh) introduces a concept of producer and consumer routes:

> A Route in the same Namespace as its Service is called a **producer route**, 
> since it is typically created by the creator of the workload in order to define acceptable usage of the workload.

> A Route in a different Namespace than its Service is called a **consumer route**. 
> Typically, this is a Route meant to refine how a consumer of a given workload makes request of that workload.

The concept of producer/consumer routes is very convenient. 
We can extend this to producer/consumer policies in general:

* A **producer policy** is a policy created in the same namespace as the service it's referring to. 
Typically created by the service owner in order to define acceptable usage of the service.

* A **consumer policy** is a policy created in a different namespace than the service it's referring to.
Typically, the policy is meant to refine how a consumer of a given service makes request to it.

Applying this concept to Kuma policies, we'll get the following examples:

```yaml
# Backend Service Owner creates a producer policy 
metadata:
  namespace: backend-ns
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshService
        name: backend
        namespace: backend-ns
      default: $conf1 # 'conf1' is going to be applied to all consumers of the backend service
---
# Frontend Service Owner creates a consumer policy 
metadata:
  namespace: frontend-ns
spec:
  targetRef:
    kind: MeshSubset
    tags:
      kuma.io/service: frontend_frontend-ns_svc_8080
  to:
    - targetRef:
        kind: MeshService
        name: backend
        namespace: backend-ns
      default: $conf2 # 'conf2' refines 'conf1' that was already provided by the producer
```

We can formally define what's producer and consumer Kuma policies:

* if for each `idx` the value of `spec.to[idx].targetRef.namespace` is equal to `metadata.namespace` then it's a producer policy
* if for each `idx` the value of `spec.to[idx].targetRef.namespace` is **not** equal to `metadata.namespace` then it's a consumer policy

The concept of producer/consumer policies is applied only to policies with `to` array.
Other policies, i.e. with `from` or without `from/to` are neither producer nor consumer policies.
For convenience when referring to such type of policies we're going to call them **workload-owner policies**.
The examples of workload-owner policies are MeshTrace, MeshMetric, MeshProxyPatch, MeshTrafficPermission, etc. 

Top-level targetRef in Kuma policies always references a group of sidecars that are subjected to the change.
New namespace-scoped policies are no exception, but we're going to implicitly add `k8s.kuma.io/namespace` and `kuma.io/zone` for **consumer** and **workload-owner policies**.

```yaml
# a valid consumer policy
metadata:
  namespace: frontend-ns
  labels:
    kuma.io/zone: zone-1
spec:
  targetRef:
    kind: Mesh # despite it targets 'Mesh' the policy will be applied to pods in the 'frontend-ns' of 'zone-1'
  to:
    - targetRef:
        kind: MeshService
        name: backend
        namespace: backend-ns
---
# workload-owner policy with 'from'
metadata:
  namespace: frontend-ns
  labels:
    kuma.io/zone: zone-1
spec:
  targetRef:
    kind: Mesh # despite it targets 'Mesh' the policy will be applied to pods in the 'frontend-ns' of 'zone-1'
  from:
    - targetRef:
        kind: Mesh
```

### Validation

Policies with `to` array that have items for local and non-local namespaces shouldn't be allowed:

```yaml
# invalid policy example
metadata:
  namespace: finance-ns
  labels:
    kuma.io/zone: zone-1
spec:
  to:
    - targetRef:
        kind: MeshService
        name: backend
        namespace: finance-ns # this indicates the policy is a producer policy
    - targetRef:
        kind: MeshService
        name: redis
        namespace: redis-ns # this indicates the policy is a consumer policy
```

Namespace-scoped policies that support both `from` and `to` and specify them at the same time shouldn't be allowed as well.

### Syntactic sugar

There are few things in policies syntax Kuma can figure out without requiring user to specify them explicitly.
The syntactic sugar is very important as it allows user to not specify zone's name and namespace explicitly in the policy's yaml. 

We don't want to change `spec` on the fly, because it might be surprising for user to see completely different resource in the store.
Also, in case of the future changes in the logic, it's preferable to store the less explicit version. 
But we probably have to implement a new feature in the Inspect API that allows viewing the most explicit version of the resource.

#### Empty namespace in `to[].targetRef`

If the `spec.to[idx].targetRef.namespace` is not specified then it's equal to `metadata.namespace`.

#### Empty top-level targetRef

* if the top-level targetRef is not specified in **producer policy**
  then it's assumed to be `Mesh`
* if the top-level targetRef is not specified  or is `targetRef.kind: Mesh` in **consumer policy** or **workload-owner policy**
  then it's assumed to be `MeshSubset` with `kuma.io/zone` and `k8s.kuma.io/namespace` tags.

The following producer policies semantically equal:

```yaml
metadata:
  namespace: backend-ns
spec:
  to:
    - targetRef:
        kind: MeshService
        name: backend
---
metadata:
  namespace: backend-ns
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshService
        name: backend
        namespace: backend-ns
```

The following workload-owner policies are semantically equal:

```yaml
metadata:
  namespace: backend-ns
spec:
  from:
    - targetRef:
        kind: Mesh
---
metadata:
  namespace: backend-ns
spec:
  targetRef:
    kind: MeshSubset
    tags:
      k8s.kuma.io/namespace: backend-ns
      kuma.io/zone: zone-with-backend
  from:
    - targetRef:
        kind: Mesh
```

The following consumer policies are semantically equal:

```yaml
metadata:
  namespace: frontend-ns
spec:
  to:
    - targetRef:
        kind: MeshService
        name: backend
        namespace: backend-ns
---
metadata:
  namespace: frontend-ns
spec:
  targetRef:
    kind: MeshSubset
    tags:
      k8s.kuma.io/namespace: frontend-ns
      kuma.io/zone: zone-with-frontend
  to:
    - targetRef:
        kind: MeshService
        name: backend
        namespace: backend-ns
```

#### Empty `from[].targetRef`

Most of the policies support only `kind: Mesh` in `from[].targetRef` so it makes sense to allow the field to be empty.
If `from[].targetRef` is not specified we're going to assume it's equal to `kind: Mesh`.
The following policies are semantically equal:

```yaml
metadata:
  namespace: backend-ns
spec:
  from:
    - default: {}
---
metadata:
  namespace: backend-ns
spec:
  targetRef:
    kind: MeshSubset
    tags:
      k8s.kuma.io/namespace: backend-ns
      kuma.io/zone: zone-with-backend
  from:
    - targetRef:
        kind: Mesh
      default: {}
```

### Namespace-scoped policies and multizone

Namespace-scoped policies can be applied to a custom namespace only on zones as there are simply no DPPs and custom
namespaces on Global. This will be enforced by validation webhook.

After a user applied a policy on zone's custom namespace, the webhook automatically adds a `k8s.kuma.io/namespace` label to the policy.
The label is part of the hash suffix when policy is synced to global.

Either implicitly or explicitly a policy can have a `kuma.io/zone` tag in the top-level targetRef.
Based on the actual or assumed value of `kuma.io/zone` tag Kuma makes a decision if policy needs to be synced to other zones.

```yaml
# Producer policy, zone tag is not specified which means the policy has to be synced to all other zones,
metadata:
  namespace: backend-ns
spec:
  to:
    - targetRef:
        kind: MeshService
        name: backend
---
# consumer policy, 'kuma.io/zone' tag is implicitly equal to the zone's name where policy is applied. 
# No need to sync this policy to other zones.
metadata:
  namespace: frontend-ns
spec:
  to:
    - targetRef:
        kind: MeshService
        name: backend
        namespace: backend-ns
```

Policies that reference `MeshExternalService` in `to[].targetRef` won't be synced to other zones.
MeshExternalService can be created only in `kuma-system` namespace and that's why the policy it's referenced by is always a consumer policy.
That's why the policies that have potential to indirectly change ZoneEgress configuration are never synced to other zones.

Zone-originated policies in `kuma-system` won't be synced to other zones.
Zone-originated policies on Universal won't be synced to other zones as well.
We might want to introduce namespaces to Universal in the future.

### Order when merging

By design, several policies can select the same DPP.
In that case, policies are ordered and their `to` (or `from`) arrays are merged.
Introducing namespace-scoped policies adds other steps to the policy comparison process.

1. Is top-level `targetRef.kind` more specific (MeshServiceSubset > MeshService > MeshSubset > Mesh)?
   When kinds are equal, go to the next step.
2. Zone originated policy is more specific than global originated policy.
   When policies originates at the same place, go to the next step.
3. **[new step]** Compare namespaces so that `consumer-ns` > `producer-ns` > `kuma-system`.
   If namespaces are equal, go to the next step.
4. The policy is more specific if it's name is lexicographically less than other policy's name ("aaa" < "bbb" so that "
   aaa" policy is more specific).

### Cross-resource references

When a policy in `zone-1` references a resource (MeshService or MeshHTTPRoute), 
the reference should work after policy is synced to `zone-2`.

Using `to[].targetRef.labels` is a reliable way to reference the resource, because labels are constant across zones. 
For example, the label `kuma.io/display-name` has the same value no matter what zone or global we're using to read the resource. 
That's why targetRefs like

```yaml
to:
  - targetRef:
      kind: MeshService
      labels:
        kuma.io/display-name: backend
```
are going to work even when policy is synced to another zone.
The downside of this approach is that multiple MeshServices can be selected, 
because we didn't specify `kuma.io/zone` and `k8s.kuma.io/namespace` labels.
Also, using labels is a more verbose approach especially when it comes to referencing the resources in the same namespace.

Another way to reference the resource is by using `name` and `namespace` fields:

```yaml
to:
  - targetRef:
      kind: MeshService
      name: backend
      namespace: backend-ns
```

but we have to implement "smart" de-referencing when the policy is synced to another zone.

#### Considered Options

* update all `to[].targetRef.name` with hashed name before syncing the resource to another zone
* de-reference `to[].targetRef.name` in the destination zone when needed (policy's spec is not changed)

#### Decision Outcome

* update all `to[].targetRef.name` with hashed name before syncing the resource to another zone

#### Pros and Cons of updating all `to[].targetRef.name` with hashed name before syncing the resource to another zone

* Good, because `name` and `namespace` always refer to the real object
* Good, because more predictable for users. They can always copy the name with hash suffix and find the resource in the store
* Bad, because we're modifying the original spec provided by user

#### Pros and Cons of de-referencing `to[].targetRef.name` in the destination zone when needed (policy's spec is not changed)

* Good, because we're not modifying the spec on the fly
* Bad, because de-referencing is not straightforward and hides extra logic

#### Implementation

The update of `to[].targetRef` is going to happen in 2 steps:

1. Global KDS client updates `to[].targetRef.name` and `to[].targetRef.namespace` when receives policies from the source zone
2. KDS client of the destination zone updates `to[].targetRef.namespace` when receives policies from Global

All clients should update `to[].targetRef.namespace` because this is an environment specific information 
(i.e. system namespaces can be different, Universal deployments don't have namespaces at all).

#### Examples

##### Policy in zone-1 references MeshService in zone-1 (and reference works after both resources synced to zone-2)

```yaml
# zone-1, producer-timeout references backend service
kind: MeshService
metadata:
  name: backend
  namespace: backend-ns
---
kind: MeshTimeout
metadata:
  name: producer-timeout
  namespace: backend-ns
spec:
  to:
    - targetRef:
        kind: MeshService
        name: backend
```

these resources are synced to `zone-2`:

```yaml
# zone-2, producer-timeout references backend service
kind: MeshService
metadata:
  name: backend-w3r3jwi90l8dyuix
  namespace: kuma-system
  labels:
    kuma.io/zone: zone-1
    k8s.kuma.io/namespace: backend-ns
    kuma.io/display-name: backend
---
kind: MeshTimeout
metadata:
  name: producer-timeout-qwea89uy739jdgmu
  namespace: kuma-system
  labels:
    kuma.io/zone: zone-1
    k8s.kuma.io/namespace: backend-ns
    kuma.io/display-name: producer-timeout
spec:
  to:
    - targetRef:
        kind: MeshService
        name: backend-w3r3jwi90l8dyuix
        namespace: kuma-system
```

##### Policy in zone-1 references MeshHTTPRoute from zone-2

As a **Backend Service Owner** I want to route all requests with `/v2` prefix to `version: v2` instances
so that requests that need v2 version of the API ended up on correct instances.

As a **Backend Service Owner** I want all requests routed to `version: v2` instances to have a 5s timeout
so that consumers didn't wait longer than max possible processing time.

As a **Frontend Service Owner** I want all requests routed to `version: v2` instances of `backend` to have 3s timeout
so that my service meets SLO and responds withing 3s.

```yaml
# zone-1, producer route 
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: backend-v2
  namespace: backend-ns
  labels:
    kuma.io/zone: zone-1
spec:
  to:
    - targetRef:
        kind: MeshService
        name: backend
      rules:
        - matches:
            - path:
                type: PathPrefix
                value: /v2
          default:
            backendRefs:
              - kind: MeshServiceSubset
                name: backend
                tags:
                  version: v2
---
# zone-1, producer policy
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: backend-route-timeout
  namespace: backend-ns
  labels:
    kuma.io/zone: zone-1
spec:
  to:
    - targetRef:
        kind: MeshHTTPRoute # doesn't work like this today, we need https://github.com/kumahq/kuma/issues/10247
        name: backend-route
      default:
        http:
          requestTimeout: 5s
```

```yaml
# zone-2, consumer policy
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: frontend-to-backend
  namespace: frontend-ns
  labels:
    kuma.io/zone: zone-2
spec:
  to:
    - targetRef:
        kind: MeshHTTPRoute # technically we can use 'name' and 'namespace' but we'd have to know hashed name of the route after syncing
        labels:
          kuma.io/display-name: backend-route
          k8s.kuma.io/namespace: backend-ns
          kuma.io/zone: zone-1
      default:
        http:
          requestTimeout: 3s
```

### Referencing MeshGateway

`MeshGateway` at this moment is not namespace-scoped. It can change in the future with
the [MeshBuiltinGateway resource](https://github.com/kumahq/kuma/issues/10014).

We decide that we won't allow creating Mesh*Routes in custom namespace. Mesh Operator should be responsible for deciding
which traffic should enter the Mesh, and create route in system namespace.

For policies in custom namespaces, we've decided that we would allow targeting `MeshGateway` in topLevel targetRef only in 
policies that configures only Envoy clusters (policies that allows configuring: `to[].targetRef.kind: MeshService` like `MeshCircuitBreaker`).

For other policies we would allow referencing Mesh*Routes that reference `MeshGateway` in topLevel targetRef.

#### Examples

As a **Cluster Operator** I want to deploy a single global `MeshGateway` to the cluster so that I’m able to manage a
single deployment for the whole cluster without knowing specific details about apps and routes.

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshGateway
mesh: default
metadata:
  name: edge-gateway
  labels:
    kuma.io/origin: zone
spec:
  conf:
    listeners:
      - port: 80
        protocol: HTTP
  selectors:
    - match:
        kuma.io/service: edge-gateway
```

##### Targeting MeshGateway policies

###### Producer policies

As a **Frontend Service Owner** I want to configure circuit breaker on `MeshGateway` without the access to
cluster-scoped CRDs so that I could manage my policies independently. `MeshGateway` behaves like `MeshSubset`. 

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshCircuitBreaker
metadata:
  name: circuit-breaker
  namespace: frontend-ns
spec:
  targetRef:
    kind: MeshGateway
    name: edge-gateway
  to:
  - targetRef:
      kind: MeshService 
      name: frontend 
      # namespace: frontend-ns - is added automatically
    default:
      outlierDetection:
        detectors:
          totalFailures:
            consecutive: 10
```

This is straightforward example that will work like any other producer policies and will be applied to MeshGateway

###### Consumer policies

Let's look at the example:

```yaml
metadata:
  namespace: finance-ns
  labels:
    kuma.io/zone: zone-1
spec:
  targetRef:
    kind: MeshGateway
    name: edge-gateway
    tags:
      k8s.kuma.io/namespace: finance-ns # should we assume this?
  to:
    - targetRef:
        kind: MeshService
        name: backend
        namespace: backend-ns
```

We are creating policy in `frontend-ns` namespace, and we are modifying traffic to backend service in different namespace. 
Right now MeshGateway is cluster scoped, so it is not clear what should we do. I think we have three options here:

1. forbid user from configuring this, `MeshGateway` is cluster scoped so this does not make sense
2. allow and try to apply config only to `MeshGatewayInstances` in `frontend-ns`, feels wrong since policy is not really applied on MeshGateway but on MeshGatewayInstance, 
but this can be done now with matching gateway + listener + dataplane tags, and specifying namespace tag 
3. allow but ignore policy, as there is no namespace scoped MeshGateway present

**We've decided to go with option no 1. We will forbid user from configuring this `MeshGateway` is cluster scoped so this does not make sense**

## Cross namespace traffic forwarding and ReferenceGrant

There is a [CVE on cross namespace traffic](https://github.com/kubernetes/kubernetes/issues/103675). To solve this issue
Kubernetes Gateway API introduced [ReferenceGrant](https://gateway-api.sigs.k8s.io/api-types/referencegrant/).

For the east-west traffic this is not needed as we can use `MeshTrafficPermission` to handle permissions. It is more
broadly explained in [Gateway API docs](https://gateway-api.sigs.k8s.io/geps/gep-1426/#namespace-boundaries).

The problem could emerge in MeshGateway. Lets look at the example:

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: http-route-1
  namespace: payments-ns
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
                type: PathPrefix
                value: /internal
          default:
            backendRefs:
              - kind: MeshService
                name: internal
                namespace: internal-ns
```

This `MeshHTTPRoute` is created in a `payment-ns` namespace and routes traffic straight to `internal-ns` which should
not be possible. We want to forbid selecting `backendRef` from namespace different from policy namespace if top level
targetRef is `MeshGateway`.

Since our `MeshGateway` is cluster scoped it is not limited to single namespace. You can direct traffic to any namespace
by applying policy in `kuma-system` namespace, to which not everyone should have access. Because of this we don't need
to implement `ReferenceGrant` yet. But this will probably be needed after adding namespace
scoped [MeshBuiltinGateway](https://github.com/kumahq/kuma/issues/10014) in the future.
