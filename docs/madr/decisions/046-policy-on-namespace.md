# Policy On Namespace

- Status: accepted

## Context and Problem Statement

Kuma 2.6.0 introduced a new feature allowing users to apply policies on Zone CP.
This change unblocks us to introduce a namespace support for policies.
Namespaced-scoped policies provide a k8s-native UX:

* ability to use k8s RBAC
* policy affect only workloads in the same namespace

The CRDs for new policies are already namespace-scoped,
but they can be applied only to the `kuma-system` namespace.

## Considered Options

- allow applying new policies to the custom namespaces
- do nothing

## Decision Outcome

- allow applying new policies to the custom namespaces

## Pros and Cons of the applying policies on custom namespaces

* Good, because allows using Kubernetes RBAC
* Good, because provides more predictable behaviour as namespace-scoped policy affects only workloads in the same namespace
* Good, because provides overall more Kubernetes-native UX
* Bad, because adds complexity for users
  * what's the right namespace for the policy?
  * how cross-policy refs work?
  * how policy order works?

## Pros and Cons of doing nothing

* Good, because we have time to work on other features
* Bad, because cluster operator has to grant write permissions to `kuma-system` namespace to everyone who works with policies
* Bad, because low isolation between teams (team-a can unintentionally break polices of the team-b) 

## Implementation

User applies policy on Kubernetes zone in a custom namespace. 
The policy automatically gets a `k8s.kuma.io/namespace` label. 
The label is part of the hash suffix when policy is synced to global.

By design, several policies can select the same DPP.
In that case, policies are ordered and their `to` (or `from`) lists are merged.
Introducing namespace-scoped policies adds another step to the policy comparison process: 
1. Is top-level `targetRef.kind` more specific (MeshServiceSubset > MeshService > MeshSubset > Mesh)? 
When kinds are equal, go to the next step.
2. Zone originated policy is more specific than global originated policy. 
When policies originates at the same place, go to the next step.
3. **[new step]** Policy in a custom namespace is more specific than a policy in the `kuma-system` namespace.
When policies applied in the same namespace, go to the next step.
4. The policy is more specific if it's name is lexicographically less than other policy's name ("aaa" < "bbb" so that "aaa" policy is more specific)

Policy applied to the custom namespace `my-namespace` of  zone-a, won’t affect the custom namespace `my-namespace` of zone-b, as these namespaces are considered different. 
If such behavior is need, user can always apply the policy on global with top-level targetRef:

```yaml
targetRef:
  kind: MeshSubset
  tags:
    k8s.kuma.io/namespace: my-namespace
```

Kuma policies support cross-policy references. 
At this moment, it works only between MeshTimeout and MeshHTTPRoute, but there are plans to support it for other policies, i.e. [#6645](https://github.com/kumahq/kuma/issues/6645).
Since policies will be namespace-scoped, the targetRef schema should get a new field `Namespace`:

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
