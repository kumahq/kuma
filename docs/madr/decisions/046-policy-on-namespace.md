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

### Namespace-scoped policies and multizone

Namespace-scoped policies can be applied to a custom namespace only on zones as there are simply no DPPs and custom
namespaces on Global.

After a user applied a policy on zone's custom namespace, it automatically gets a `k8s.kuma.io/namespace` label.
The label is part of the hash suffix when policy is synced to global.

Policy applied to the custom namespace `my-namespace` of zone-a, won’t affect pods in the custom
namespace `my-namespace` of zone-b,
as these namespaces are considered different.
If such behavior is need, user can always apply the policy on global with top-level targetRef:

```yaml
targetRef:
  kind: MeshSubset
  tags:
    k8s.kuma.io/namespace: my-namespace
```

### Order when merging

By design, several policies can select the same DPP.
In that case, policies are ordered and their `to` (or `from`) lists are merged.
Introducing namespace-scoped policies adds another step to the policy comparison process:

1. Is top-level `targetRef.kind` more specific (MeshServiceSubset > MeshService > MeshSubset > Mesh)?
   When kinds are equal, go to the next step.
2. Zone originated policy is more specific than global originated policy.
   When policies originates at the same place, go to the next step.
3. **[new step]** Policy in a custom namespace is more specific than a policy in the `kuma-system` namespace.
   When policies applied in the same namespace, go to the next step.
4. The policy is more specific if it's name is lexicographically less than other policy's name ("aaa" < "bbb" so that "
   aaa" policy is more specific)

### Cross-policy references

Kuma policies support cross-policy references.
At this moment, it works only between MeshTimeout and MeshHTTPRoute, but there are plans to support it for other
policies, i.e. [#6645](https://github.com/kumahq/kuma/issues/6645).
Since policies will be namespace-scoped, we need to add a new field `Namespace` in the TargetRef:

```go
// TargetRef defines structure that allows attaching policy to various objects
type TargetRef struct {
    // Kind of the referenced resource
    // +kubebuilder:validation:Enum=Mesh;MeshSubset;MeshGateway;MeshService;MeshServiceSubset;MeshHTTPRoute
    Kind TargetRefKind `json:"kind,omitempty"`
    // Name of the referenced resource. Can only be used with kinds: `MeshService`,
    // `MeshServiceSubset` and `MeshGatewayRoute`
    Name string `json:"name,omitempty"`
    // Tags used to select a subset of proxies by tags. Can only be used with kinds
    // `MeshSubset` and `MeshServiceSubset`
    Tags map[string]string `json:"tags,omitempty"`
    // Mesh is reserved for future use to identify cross mesh resources.
    Mesh string `json:"mesh,omitempty"`
    // ProxyTypes specifies the data plane types that are subject to the policy. When not specified,
    // all data plane types are targeted by the policy.
    // +kubebuilder:validation:MinItems=1
    ProxyTypes []TargetRefProxyType `json:"proxyTypes,omitempty"`
    // Namespace of the referenced resource. If not specified, then equal the value is equal to the resource's namespace
    // where this TargetRef is used.
    Namespace string `json:"namespace,omitempty"`
}
```

Let's try to review all possible types of references between policies and see if maybe some of them don't make sense.
We have 3 namespaces: `kuma-system`, `frontend-ns` and `backend-ns`. We have 2 policies: `MeshTimeout`
and `MeshHTTPRoute`,
and `MeshTimeout` always references the `MeshHTTPRoute`. Overall, there are 3^2=9 scenarios:

1. MeshTimeout in kuma-system, MeshHTTPRoute in kuma-system
2. MeshTimeout in kuma-system, MeshHTTPRoute in frontend-ns
3. MeshTimeout in kuma-system, MeshHTTPRoute in backend-ns
4. MeshTimeout in frontend-ns, MeshHTTPRoute in kuma-system
5. MeshTimeout in frontend-ns, MeshHTTPRoute in frontend-ns
6. MeshTimeout in frontend-ns, MeshHTTPRoute in backend-ns
7. MeshTimeout in backend-ns, MeshHTTPRoute in kuma-system
8. MeshTimeout in backend-ns, MeshHTTPRoute in frontend-ns
9. MeshTimeout in backend-ns, MeshHTTPRoute in backend-ns

Scenarios 2 and 3, 4 and 7, 5 and 9, 6 and 8 are describing the same setup and could be merged.
We have 5 scenarios left, so let's look closely and try to come up with meaningful use cases for each one.

**1. Both MeshTimeout and MeshHTTPRoute in "kuma-system"**

This scenario is already supported today.
The use case for this scenario can be the following:

**As** a Backend Service Owner<br>
**I want** to route all requests with `/v2` prefix to `version: v2` instances<br>
**so that** requests that need v2 version of the API ended up on correct instances.

**As** a Backend Service Owner<br>
**I want** all requests routed to `version: v2` instances to have a 5s timeout<br>
**so that** consumers didn't wait longer than max possible processing time.

This can be achieved by applying 2 policies:

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: backend-route
  namespace: kuma-system
spec:
  targetRef:
    kind: Mesh
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
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: backend-route-timeout
  namespace: kuma-system
spec:
  targetRef:
    kind: MeshHTTPRoute
    name: backend-route
  to:
    - targetRef:
        kind: Mesh
      default:
        http:
          requestTimeout: 5s
```

**2. MeshTimeout in "kuma-system", MeshHTTPRoute in "frontend-ns"**

Frontend Service Owner creates a MeshHTTPRoute in the `frontend-ns` namespace.
What are the reasons for Mesh Operator to create MeshTimeout in `kuma-system` and reference the MeshHTTPRoute policy
in `frontend-ns`?
Mesh Operator should work with policies that apply globally, knowing specifics of the `frontend` app seems incorrect.
We didn't find good use cases why this type of reference would make sense.

**3. MeshTimeout in "frontend-ns", MeshHTTPRoute in "kuma-system"**

**As** a Backend Service Owner<br>
**I want** to route all requests with `/v2` prefix to `version: v2` instances<br>
**so that** requests that need v2 version of the API ended up on correct instances.

**As** a Backend Service Owner<br>
**I want** all requests routed to `version: v2` instances to have a 5s timeout<br>
**so that** consumers didn't wait longer than max possible processing time.

**As** a Frontend Service Owner<br>
**I want** all requests routed to `version: v2` instances of `backend` to have 3s timeout<br>
**so that** my service meets SLO and responds withing 3s.

This can be achieved by applying 3 policies:

```yaml
# on kuma-system namespace
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: backend-route
  namespace: kuma-system
spec:
  targetRef:
    kind: Mesh
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
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: backend-route-timeout
  namespace: kuma-system
spec:
  targetRef:
    kind: MeshHTTPRoute
    name: backend-route
  to:
    - targetRef:
        kind: Mesh
      default:
        http:
          requestTimeout: 5s
# on frontend-ns
---
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: frontend-to-backend
  namespace: frontend-ns
spec:
  targetRef:
    kind: MeshHTTPRoute
    name: backend-route
    namespace: kuma-system
  to:
    - targetRef:
        kind: Mesh
      default:
        http:
          requestTimeout: 3s
```

This example also demonstrates the scenario when 2 MeshTimeout reference the same route,
the namespaced MeshTimeout is considered more specific.
That's why the final request timeout value should be 3s.

**4. Both MeshTimeout and MeshHTTPRoute in "frontend-ns"**

This scenario is likely to be the most common way users create policies.
The use case for this is:

**As** a Frontend Service Owner<br>
**I want** all outgoing requests from fronted (to backend) to have headers starting with `xxx-fronted-` removed<br>
**so that** requests are smaller and backend doesn't receive unnecessary data.

**As** a Frontend Service Owner<br>
**I want** all requests routed to `backend` to have 3s timeout<br>
**so that** my service meets SLO and responds withing 3s.

This can be achieved by applying 2 policies:

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: backend-route
  namespace: frontend-ns
spec:
  targetRef:
    kind: MeshService
    name: frontend
  to:
    - targetRef:
        kind: MeshService
        name: backend
      rules:
        - matches:
            - path:
                type: PathPrefix
                value: /
          default:
            filters:
              - type: RequestHeaderModifier
                requestHeaderModifier:
                  remove:
                    - xxx-frontend-version
                    - xxx-frontend-data
                    - xxx-frontend-content
            backendRefs:
              - kind: MeshService
                name: backend
---
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: frontend-to-backend
  namespace: frontend-ns
spec:
  targetRef:
    kind: MeshHTTPRoute
    name: backend-route
    namespace: frontend-ns # can be omitted as it's a default value
  to:
    - targetRef:
        kind: Mesh
      default:
        http:
          requestTimeout: 3s
```

**5. MeshTimeout in "frontend-ns", MeshHTTPRoute in "backend-ns"**

MeshHTTPRoute in `backend-ns` produces envoy routes in DPPs in `backend-ns`.
Ability to configure timeouts on these routes from the `frontend-ns` breaks the namespace isolation and seems
unnecessary.

**In the end,** only 3 scenarios have meaningful use cases and could be described by 2 principles:

* a policy in a custom namespace can reference all other policies regardless their namespace
* a policy in the system namespace can reference only policies in the system namespace

### MeshService reference

Since we didn't make any decisions on how to target MeshService in policies we will leave this out of now. We will
revisit this while discussing targeting MeshService in policies.

### MeshGateway reference

`MeshGateway` at this moment is not namespace-scoped. It can change in the future with
the [MeshBuiltinGateway resource](https://github.com/kumahq/kuma/issues/10014).

We've decided that we won't allow targeting `MeshGateway` in namspace scoped polcies except for the
`MeshHTTPRoute` and `MeshTCPRoute`. To apply policy like `MeshTimeout` or `MeshCircuitBreaker` on
`MeshGateway` you need to create it in `kuma-system` namespace.

With `MeshHTTPRoute` and `MeshTCPRoute` you can target `MeshGateway` either from `kuma-system` or custom namespace.

#### Example

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
