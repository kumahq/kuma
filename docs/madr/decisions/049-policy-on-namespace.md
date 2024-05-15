# Policy On Namespace

- Status: accepted

## Context and Problem Statement

Kuma 2.6.0 introduced a new feature allowing users to apply policies on Zone CP.
This change unblocks us to introduce a namespace support for policies.
The CRDs for new policies are already namespace-scoped, but they can be applied only on the `kuma-system` namespace.

## Considered Options

- allow applying new policies to the custom namespaces
- do nothing

## Decision Outcome

- allow applying new policies to the custom namespaces

## Pros and Cons of the applying policies on custom namespaces

* Good, because allows using Kubernetes RBAC
* Good, because provides more predictable behaviour as namespace-scoped policy affects only workloads in the same
  namespace
* Good, because provides more Kubernetes-native UX
* Bad, because adds complexity for users
    * what's the right namespace for the policy?
    * how cross-policy refs work?
    * how policy order works?

## Pros and Cons of doing nothing

* Good, because we have time to work on other features
* Bad, because cluster operator has to grant write permissions to `kuma-system` namespace to everyone who works with
  policies
* Bad, because low isolation between teams (team-a can unintentionally break polices of the team-b)

## Implementation

Applying a policy on the custom namespace **affects requests to the pods in the custom namespace**.
Since some policies are applied on the client-side, the scope of policy's effect is not limited to the namespace
where it was applied. 
For example, the following policy sets the timeout on outbounds of all consumers of the backend service:

```yaml
kind: MeshTimeout
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
      default:
        connectionTimeout: 5s
```

Important to note, even though policy in `backend-ns` changes proxy configurations in `frontend-ns` 
the top-level targetRef still servers the purpose of showing what proxy configurations are going to change.
Policy's namespace **does not** implicitly turn the top-level targetRef to `MeshSubset{tags:{"kuma.io/namespace":"backend-ns"}}`.

### Namespace-scoped policies and multizone

Namespace-scoped policies can be applied to a custom namespace only on zones as there are simply no DPPs and custom
namespaces on Global.

After a user applied a policy on zone's custom namespace, it automatically gets a `k8s.kuma.io/namespace` label.
The label is part of the hash suffix when policy is synced to global.

As mentioned previously, a policy applied in producer namespace can affect envoy configs of consumers.
Consumers can reside not only in different namespaces, but also in different zones.
That's why, the policy applied in `backend-ns` in `zone-1` has to be synced to global and then to other zones.
With the increased blast radius of the policies comes great responsibility, 
we don't want small mistake in the policy to break traffic on the other side of the world.
A few things could be done to protect users from impactful mistakes:
* Enhanced validation. If policy is created in `backend-ns` of the `zone-1` and the top-level targetRef is `Mesh`, 
then `to[].targetRef` should have both `k8s.kuma.io/namespace: backend-ns` and `kuma.io/zone: zone-1`. 
* Ability to opt-in or opt-out from syncing the policy to other zones.

#### Enhanced validation

Since applying a policy to the custom namespace of a zone has a potential to change envoy configs of all pods in the mesh including pods in other zones,
we need to restrict "to" section, so it affect only the requests going to the custom namespace.
For example, the following policy when applied to `backend-ns` is too intrusive toward the backend consumers:

```yaml
kind: MeshTimeout
metadata:
  namespace: backend-ns
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshService
        labels:
          kuma.io/display-name: redis
          k8s.kuma.io/namespace: redis-ns
      default:
        connectionTimeout: 5s # why does the Backend Service Owner try to change the way clients use the 'redis' service?
```

Such policy should not be allowed, if Backend Service Owner needs to change the way clients use `redis`
they have to put the policy into `kuma-system` namespace by checking with Mesh Operator first.

The validation rules for namespace-scoped policy are the following:
* to-policy when placed into the custom namespace must have the namespace and zone either in the top-level targetRef or in each to-item
* from-policy when placed into the custom namespace must always have the namespace and zone in the top-level targetRef

For example, allowed policies are:
```yaml
metadata:
  namespace: backend-ns
  labels:
    kuma.io/origin: zone
    kuma.io/zone: producer-zone
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshService
        labels:
          kuma.io/display-name: backend
          k8s.kuma.io/namespace: backend-ns
          kuma.io/zone: producer-zone
---
metadata:
  namespace: backend-ns
  labels:
    kuma.io/origin: zone
    kuma.io/zone: producer-zone
spec:
  targetRef:
    kind: MeshSubset
    tags:
      k8s.kuma.io/namespace: backend-ns
  to:
    - targetRef:
        kind: MeshService
        labels:
          kuma.io/display-name: redis
          kuma.io/zone: producer-zone
---
metadata:
  namespace: backend-ns
  labels:
    kuma.io/origin: zone
    kuma.io/zone: producer-zone
spec:
  targetRef:
    kind: MeshSubset
    tags:
      k8s.kuma.io/namespace: backend-ns
      kuma.io/zone: producer-zone
  to:
    - targetRef:
        kind: MeshService
        name: redis
        namespace: redis-ns
    - targetRef:
        kind: Mesh
---
metadata:
  namespace: backend-ns
  labels:
    kuma.io/origin: zone
    kuma.io/zone: producer-zone
spec:
  targetRef:
    kind: MeshSubset
    tags:
      k8s.kuma.io/namespace: backend-ns
      kuma.io/zone: producer-zone
  from:
    - targetRef:
        kind: MeshService
        name: redis
        namespace: redis-ns
```

Not allowed:

```yaml
# The top-level targetRef references the Mesh, so consumers in all zones will be affected by the policy.
# Label 'kuma.io/zone: producer-zone' is missing from 'to[0].targetRef.labels'. 
# Without it, the policy is going to affect requests to 'backend' service in any zone.
metadata:
  namespace: backend-ns
  labels:
    kuma.io/origin: zone
    kuma.io/zone: producer-zone
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshService
        labels:
          kuma.io/display-name: backend
          k8s.kuma.io/namespace: backend-ns
---
# The top-level targetRef references the Mesh, but this is an inbound policy, 
# and we try to affect requests from all consumers to redis service.
# This should not be allowed.
metadata:
  namespace: backend-ns
  labels:
    kuma.io/origin: zone
    kuma.io/zone: producer-zone
spec:
  targetRef:
    kind: Mesh
  from:
    - targetRef:
        kind: MeshService
        name: redis
        namespace: redis-ns
```

#### Considered options on opting-in(out) to cross-zone policy syncing

* Introduce `spec.export.mode` for all policies, the default is `None`
* Introduce `spec.export.mode` for all policies, the default is `All`
* Automatically detect if a policy needs to be exported

##### Decision Outcome

Automatically detect if a policy needs to be exported. 
Backwards compatibility is tricky, but manageable. 
We can introduce this behaviour behind the feature flag. 
This gives users time after the upgrade to make conscious choices on what policies should be exported.
When the feature flag is switched, Global is going to validate policies against the new enhanced rules.

TODO: how does the Global behave when policy doesn't pass the validation to be synced to other zones?

##### Pros and Cons of introducing `spec.export.mode`, the default is `None`

* Good, because backwards compatible. 
If the field is not set, the behaviour is similar to what we have in 2.7.
* Good, because it's the least surprising behaviour. 
No matter what policy looks like you won't find it affecting pods in other zones. 
* Bad, because user can forget about it. 
User assumes MeshHTTPRoute affects all consumers while in reality it affects only consumers in the same zone.

##### Pros and Cons of introducing `spec.export.mode`, the default is `All`

* Good, because coherent with MeshService.
* Good, because treats all consumers equally by default without favoring the local ones.
* Bad, because it's not backwards compatible.
* Bad, because it doesn't add much value.
It can be easily replaced with the top-level `targetRef{kind:MeshSubset,tags{"kuma.io/zone":"my-zone"}}` when user wants to opt-out.

##### Pros and Cons of automatically detecting if a policy needs to be exported

Having `spec.export.mode` might be excessive taking into account we already have a top-level targetRef.
If the top-level targetRef specifies a tag `kuma.io/zone` then it's clear whether we need to export the policy or not.

* Good, because we don't introduce new fields, no need for users to learn new concepts.
* Good, because all consumers are treated equally by default without favoring the local ones.
* Bad, because it's not backward compatible.

### Order when merging

By design, several policies can select the same DPP.
In that case, policies are ordered and their `to` (or `from`) lists are merged.
Introducing namespace-scoped policies adds another steps to the policy comparison process:

1. Is top-level `targetRef.kind` more specific (MeshServiceSubset > MeshService > MeshSubset > Mesh)?
   When kinds are equal, go to the next step.
2. Zone originated policy is more specific than global originated policy.
   When policies originates at the same place, go to the next step.
3. **[new step]** Policy in any custom namespace is more specific than any policy in the `kuma-system` namespace.
   When both policies either from custom namespaces or from `kuma-system`, go to the next step.
4. **[new step]** Policy from local zone is more specific than a policy from another zone. 
   When policies have the same zone, go to the next step.
5. **[new step]** Policy from local namespace is more specific than a policy from another namespace.
   When namespaces are equal, go to the next step.
6. The policy is more specific if it's name is lexicographically less than other policy's name ("aaa" < "bbb" so that "
   aaa" policy is more specific)

### Cross-policy references

Kuma policies support cross-policy references.
At this moment, it works only between MeshTimeout and MeshHTTPRoute, but there are plans to support it for other
policies, i.e. [#6645](https://github.com/kumahq/kuma/issues/6645).
Referencing a policy from another namespace or another zone should be done in a similar way MeshService is referenced.
Fields `name/namespace` and `labels` are mutually exclusive.

```yaml
# reference an object with name/namespace in the storage
targetRef:
  kind: MeshHTTPRoute
  name: orders-route
  namespace: backend-ns
---
# reference MeshHTTPRoute that was synced from 'us-east-1' 
targetRef:
  kind: MeshHTTPRoute
  labels:
    kuma.io/display-name: orders-route
    k8s.kuma.io/namespace: backend-ns
    kuma.io/zone: us-east-1
```

As mentioned previously, applying a policy on the custom namespace affects requests to the pods in the custom namespace.
This means, no matter where MeshHTTPRoute is created it can be a "producer" route and affect consumers' configurations.
It's always justifiable to reference any MeshHTTPRoute from any MeshTimeout.

### MeshGateway reference

`MeshGateway` at this moment is not namespace-scoped. It can change in the future with
the [MeshBuiltinGateway resource](https://github.com/kumahq/kuma/issues/10014).

For policies in custom namespaces, we've decided that we would allow targeting `MeshGateway` in `MeshHTTPRoute` and `MeshTCPRoute`, and policies 
that configures only Envoy cluster (services that allows configuring: `to[].targetRef.kind: MeshService` like `MeshCircuitBreaker`). 
For other policies we would allow referencing routes that reference `MeshGateway`.

## Examples

## Producer and consumer routes

As a **Backend Service Owner** I want to route all requests with `/v2` prefix to `version: v2` instances
so that requests that need v2 version of the API ended up on correct instances.

As a **Backend Service Owner** I want all requests routed to `version: v2` instances to have a 5s timeout
so that consumers didn't wait longer than max possible processing time.

As a **Frontend Service Owner** I want all requests routed to `version: v2` instances of `backend` to have 3s timeout
so that my service meets SLO and responds withing 3s.

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: backend-v2
  namespace: backend-ns
  labels:
    kuma.io/zone: zone-1
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshService
        labels:
          kuma.io/display-name: backend
          k8s.kuma.io/namespace: backend-ns
          kuma.io/zone: zone-1
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
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: backend-route-timeout
  namespace: backend-ns
  labels:
    kuma.io/zone: zone-1
spec:
  targetRef:
    kind: MeshHTTPRoute
    name: backend-route # name is enough as both policies are in the same namespace
  to:
    - targetRef:
        kind: Mesh
      default:
        http:
          requestTimeout: 5s
---
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: frontend-to-backend
  namespace: frontend-ns
  labels:
    kuma.io/zone: zone-2
spec:
  targetRef:
    kind: MeshHTTPRoute
    labels:
      kuma.io/display-name: backend-route
      k8s.kuma.io/namespace: backend-ns
      kuma.io/zone: zone-2
  to:
    - targetRef:
        kind: Mesh
      default:
        http:
          requestTimeout: 3s
```

### MeshGateway Examples

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

#### Targeting MeshGateway in MeshHttpRoute

As a **Frontend Service Owner** I want to configure routes on `MeshGateway` without the access to
cluster-scoped CRDs so that I could manage my routes independently.

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: http-route-1
  namespace: frontend-ns
  labels:
    kuma.io/origin: zone
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
                value: /gui
          default:
            backendRefs:
              - kind: MeshService
                name: frontend
                namespace: frontend-ns
```

#### Targeting MeshGateway in other policies

As a **Frontend Service Owner** I want to configure circuit breaker on `MeshGateway` without the access to
cluster-scoped CRDs so that I could manage my routes independently.

```yaml
type: MeshCircuitBreaker
mesh: default
name: circuit-breaker
namespace: frontend-ns
spec:
  targetRef:
    kind: MeshGateway
    name: edge-gateway
  to:
  - targetRef:
      kind: MeshService # you need to specify MeshService in to[].targetRef.kind
      name: frontend
    default:
      outlierDetection:
        detectors:
          totalFailures:
            consecutive: 10
```

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
not be possible. We want to forbid selecting `backendRef` from namespace different than policy namespace if top level
targetRef is `MeshGateway`.

Since our `MeshGateway` is cluster scoped it is not limited to single namespace. You can direct traffic to any namespace
by applying policy in `kuma-system` namespace, to which not everyone should have access. Because of this we don't need
to implement `ReferenceGrant` yet. But this will probably be needed after adding namespace
scoped [MeshBuiltinGateway](https://github.com/kumahq/kuma/issues/10014) in the future.
