# Policy matching algorithm with real resources (MeshService, MeshExternalService, Mesh*Route)

* Status: accepted

## Context and Problem Statement

We can think of the policy matching algorithm as a function
that maps a set of policies provided by the user to the following structure:

```yaml
Rules:
  - Subset:
      - Key: kuma.io/service
        Not: false
        Value: foo
    Conf:
      connectionTimeout: 2s
      http:
        requestTimeout: 15s
      idleTimeout: 20s
    Origin: [ ... ]
  - Subset:
      - Key: kuma.io/service
        Not: false
        Value: bar
    Conf:
      connectionTimeout: 2s
      http:
        requestTimeout: 20s
      idleTimeout: 20s
    Origin: [ ... ]
```

This structure is used by both "from" and "to" policies plugins.
Policy plugins use the `Rules` list to compute `Conf` for the known destination.
For example, if we want to know the `Conf` for outbound with `kuma.io/service: bar` tags,
we're going to iterate over `Rules` and check if our outbound belongs to the `Rules[].Subset`.

The approach has a number of disadvantages:

* even though it works well for "from" policies, using it for "to" policies doesn't feel natural
  * there is no concept of "subsets" in outbounds, we have to artificially build `{kuma.io/service: $svc}` subset
  * subset-based matching is build to handle subset overlapping scenarios, but it's never happening in outbounds
* when referencing the real MeshHTTPRoute we had to use subset with artificial `__rule-matches-hash__` key
* adding support for `MeshService` forces us to introduce another artificial key like `__resource-name__`
* Inspect API with subsets is tricky when referencing real resources (like MeshHTTPRoute or MeshService)

We need a new format that'd make things easier when it comes to referencing real resources in `to[]`.

## Considered Options and Decision Outcome

No changes for "from" policies, both policy matching algorithm and Inspect API stays the same.

For "to" policies the algorithm is going to return the additional structure:

```go
type ToRules struct {
    Rules         Rules 
    ResourceRules map[UniqueResourceKey]ResourceRule // new field
}

// UniqueResourceKey is a way to uniquely identify a resource, should include ResourceType, Name, Mesh and SectionName (for MeshServices)
type UniqueResourceKey struct{}

type ResourceRule struct {
    Resource            core_model.ResourceMeta
	ResourceSectionName string
    Conf                interface{}
    Origin              []Origin
}

type Origin struct {
    Resource core_model.ResourceMeta
    // RuleIndex is an index in the 'to[]' array, so we could unambiguously detect what to-item contributed to the final conf.
    // Especially useful when to-item uses `targetRef.Labels`, because there is no obvious matching between the specific resource
    // in `ResourceRule.Resource` and to-item.
    RuleIndex int 
}
```

For a simple targetRef with `name/namespace`:

```yaml
to:
  - targetRef:
      kind: MeshService
      name: backend
      namespace: finance
    conf: conf1
```

we're going to produce a single `ResourceRule` entity:

```yaml
resourceRules:
  meshservice:mesh/mesh-1:name/backend:ns/finance:
    resource:
      name: backend.finance
      mesh: mesh-1
    conf: conf-1
    origin: [...]
```

When targetRef has kind `MeshService` and uses `labels` we're going to produce multiple `ResourceRule` entities
for each real `MeshService` that satisfies `labels`, i.e for the given `targetRef`:

```yaml
to:
  - targetRef:
      kind: MeshService
      labels:
        k8s.kuma.io/namespace: finance
    conf: conf1
```

we are going to fetch all `MeshService` resources, filter them by `k8s.kuma.io/namespace: finance` and produce N rules:

```yaml
resourceRules:
  meshservice:mesh/mesh-1:name/finance-backend:ns/finance:
    resource:
      name: finance-backend.finance
      mesh: mesh-1
    conf: conf-1
    origin: [...]
  meshservice:mesh/mesh-1:name:finance-frontend:ns/finance:
    resource:
      name: finance-frontend.finance
      mesh: mesh-1
    conf: conf-1
    origin: [...]
  meshservice:mesh/mesh-1:name/finance-db:ns/finance:
    resource:
      name: finance-db.finance
      mesh: mesh-1
    conf: conf-1
    origin: [...]
```

When targetRef is using `sectionName` (supported only for MeshService):

```yaml
to:
  - targetRef:
      kind: MeshService
      name: backend
      namespace: finance
      sectionName: http-port
    conf: conf1
```

we're going to produce the following `ResourceRule`:

```yaml
resourceRules:
  meshservice:mesh/mesh-1:name/backend:ns/finance:section/http-port:
    resource:
      name: backend.finance
      mesh: mesh-1
    resourceSectionName: http-port
    conf: conf-1
    origin: [...]
```

Policy plugins should use old `ToRules.Rules` when computing configurations for legacy destinations with `kuma.io/service` tags.
Policy plugins should use new `ToRules.ResourceRule` when computing configuration for the new MeshService resources.
Eventually we're going to delete `ToRules.Rules`.

### Merging

#### Problem

The way merging works today is explained in the [initial policy matching MADR](https://github.com/kumahq/kuma/blob/master/docs/madr/decisions/005-policy-matching.md#merging).
In short, we take all matched policies, sort them by **the top-level targetRef** and concatenate their `to[]` arrays. 
After that, we merge `confs` from each to-items purely based on their order in the resulted `to[]` array, greater index has more priority.
Given the `to[]`:

```yaml
to:
  - targetRef:
      kind: MeshService
      name: backend
    default: $conf1
  - targetRef:
      kind: Mesh
    default: $conf2
```

the resulting conf for `backend` will be `merge($conf1, $conf2)`. 
IMPORTANT, it will **not** be `merge($conf2, $conf1)`, because to-item with `conf2` has greater index in `to[]`.
Basically, we discard `to[].targetRef.kind` when performing merging.

This approach has a major flaw, that was explained in https://github.com/kumahq/kuma/issues/9151.
When user creates 2 policies with the same top-level targetRef kind, the order of concatenation is decided by 
the alphabetical order of policies' names. 
Given 2 policies:

```yaml
kind: Policy
name: aaa
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: Mesh
      default: $conf2
---
kind: Policy
name: bbb
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshService
        name: backend
      default: $conf1
```

The result for `backend` will be `merge($conf1, $conf2)` because `aaa < bbb`.
While intuitively the conf shouldn't depend on the policy name in this case and be `merge($conf2, $conf1)`,
because `kind: MeshService` is more specific than `kind: Mesh` and we clearly have to prioritize `conf1`.

Now is a good moment to improve the way sorting/merging works, because ResourceRules is a new structure
and it's going to be used only when computing config for the new MeshServices.

#### New approach

We're going to take all matched policies and produce a list of the following items:

```yaml
toItemsWithTopLevel:
  - topLevelTargetRefKind: Mesh
    policyName: aaa
    toTargetRefKind: Mesh
    conf: $conf2
  - topLevelTargetRefKind: Mesh
    policyName: bbb
    toTargetRefKind: MeshService
    conf: $conf1
```

and we'll sort this list based on fields in the following order:

1. topLevelTargetRefKind
2. toTargetRefKind
3. policyName

Note: today we basically do 1 and 3, ignoring the field 2.

After that, if we want to find a conf for the resource `r`, we do:

```go
confs := []
for _, item := range toItemsWithTopLevel:
    if includes(item, r) {
        confs = append(confs, item.conf)
    }
}
resultingConf := merge(confs...)
```

Function `includes(r1, r2 core_model.Resource) bool` checks if one resource is included into another:
* every resource is included into `Mesh`
* `Mesh*Route` is included into `MeshService` if route's `to[].targetRef` references the MeshService

#### Migration

ResourceRule is a new structure, it will be used when configuring outbounds for the real MeshService resources.
That's why we don't need backwards compatibility for `to[]`.

For `from[]` policies the sorting algorithm should be changed as well. 
But the problem is less visible due to the fact `from[].targetRef.kind` is `Mesh` in the overwhelming majority of cases.
For MeshTrafficPermission it can be `Mesh` and `MeshSubset`, and technically the behaviour of the following policies is going to change:

```yaml
kind: MeshTrafficPermission
name: aaa
spec:
  targetRef:
    kind: Mesh
  from:
    - targetRef:
        kind: Mesh
      default: 
        action: Deny
---
kind: MeshTrafficPermission
name: bbb
spec:
  targetRef:
    kind: Mesh
  from:
    - targetRef:
        kind: MeshSubset
        tags:
          kuma.io/service: backend
      default: 
        action: Allow
```

Current behaviour is 'action: Deny' for all, because `aaa` happens to be more specific than `bbb`.
The new behaviour is 'action: Deny' for all except `backend` and 'action: Allow' for `backend`.
If such change in behaviour happens it probably means the user already had policies in the obscure state. 

Technically, there is a legitimate case when users could rely on the existing behaviour. 
In example, Mesh Operator decided to step in and create policy named `aaaaaa-the-most-specific-mtp` to override 
the unwanted behaviour that was configured by application developers.
We should cover the change in UPGRADE.md and docs to make sure it's visible for users,
but overall the change is harmless and doesn't require the backwards compatibility.

### Sparse ResourceRules structure

ResourceRule structure is going to have only resources that were specified in policies.
Given the following resources:

```yaml
type: MeshService
mesh: mesh-1
name: my-service
labels:
  a-common-label: foo
---
type: MeshService
mesh: mesh-1
name: labeled-service
labels:
  a-common-label: foo
  another-common-label: foo
---
type: MeshService
mesh: mesh-1
name: other-service
labels: {}
---
type: Policy
mesh: mesh-1
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
         kind: Mesh
       conf: $conf1
---
type: Policy
mesh: mesh-1
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
         kind: MeshService
         labels:
             a-common-label: foo
       conf: $conf2
---
type: Policy
mesh: mesh-1
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
         kind: MeshService
         name: my-service
       conf: $conf3
---
type: Policy
mesh: mesh-1
spec:
  targetRef:
    kind: MeshSubset
    labels:
      foo: bar
  to:
    - targetRef:
        kind: MeshService
        labels:
          another-common-label: foo
        conf: $conf4
```

Kuma CP should produce the following structure for the DPP that has the label `foo: bar`:

```yaml
resourceRules:
  meshservice:mesh/mesh-1:name/my-service:
    conf: merge($conf1, $conf2, $conf3)
  meshservice:mesh/mesh-1:name/labeled-service:
    conf: merge($conf1, $conf2, $conf4)
  mesh:name/mesh-1:
    conf: $conf1
```

Note that the structure doesn't contain `other-service`, but instead it contains entry for `mesh-1`.
The code that computes confs should handle the absence of `other-service` and replace it with conf of the `mesh-1`.

### Inspect API

When it comes to the Inspect API, we're going to add a new field `toResourceRules` to `InspectRule`.
Note, while the internal structure is a dictionary it's going to be transformed into an array in the Inspect API.

```yaml
InspectRule:
  type: object
  required: [type]
  properties:
    type:
      type: string
      example: MeshRetry
      description: the type of the policy
    proxyRule:
      description: a rule that affects the entire proxy
      $ref: '#/components/schemas/ProxyRule'
    toRules: # should be removed eventually
      type: array 
      description: a set of rules for the outbounds of this proxy
      items:
        $ref: '#/components/schemas/Rule'
    toResourceRules: # new field
      type: array
      items:
        $ref: '#/components/schemas/ResourceRule'
    fromRules:
      type: array
      description: a set of rules for each inbound of this proxy
      items:
        $ref: '#/components/schemas/FromRule'
    warnings:
      type: array
      description: a set of warnings to show in policy matching
      example: ["Mesh is not Mtls enabled this policy will have no effect"]
      items:
        type: string
Rule:
  type: object
  required: [matchers, conf, origin]
  properties:
    matchers:
      type: array
      items:
        $ref: '#/components/schemas/RuleMatcher'
    conf:
      description: The actual conf generated
      type: object
      x-go-type: 'interface{}'
    origin:
      type: array
      items:
        $ref: '#/components/schemas/Meta'
ResourceRule: # new type
  type: object
  required: [resourceMeta, conf, origin]
  properties:
    resourceMeta:
      $ref: '#/components/schemas/Meta'
    resourceSectionName:
      type: string
    conf:
      description: The actual conf generated
      type: object
      x-go-type: 'interface{}'
    origin:
      type: array
      description: |
        The list of policies that contributed to the 'conf'. The order is important as it reflects 
        in what order confs were merged to get the resulting 'conf'. 
      items:
        type: object
        properties:
          resourceMeta:
            $ref: '#/components/schemas/Meta'
          ruleIndex:
            description: index of the to-item in the policy
            type: integer
```

The example of the Inspect API response for the given MeshTimeout policy would be:

```yaml
kind: MeshTimeout
metadata:
  name: mt-1
  namespace: client-ns
  labels:
    kuma.io/mesh: mesh-1
spec:
  to:
    - targetRef:
        kind: Mesh
      conf: conf1
    - targetRef:
        kind: MeshService
        name: backend
        namespace: finance
      conf: conf2
```

```yaml
resource: 
  type: Dataplane
  name: dpp-1.client-ns
  mesh: mesh-1
  labels: {}
rules:
  - type: MeshTimeout
    toResourceRules:
      - resourceMeta:
          type: MeshService
          mesh: mesh-1
          name: backend.finance
          labels: 
            k8s.kuma.io/namespace: finance
        conf: merge(conf1, conf2)
        origin:
          - resourceMeta:
              type: MeshTimeout
              mesh: mesh-1
              name: mt-1
              labels: {}
            ruleIndex: 0
          - resourceMeta:
              type: MeshTimeout
              mesh: mesh-1
              name: mt-1
              labels: {}
            ruleIndex: 1
```

The Inspect API response should contain rules for all existing MeshServices/MeshHTTPRoutes in the cluster
even if there are no policies that'd directly reference them. 
Such complete structure potentially means duplication of confs, but on the bright side, client code in the GUI/kumactl
doesn't need to know how to compute conf for the resource – it can perform a simple lookup by key.

### Positive Consequences

* In Kuma GUI, it's possible to make links to the real MeshServices, MeshHTTPRoutes and MeshExternalServices when used in `to[]`
* Policy plugins code is simpler with the clean fallback logic
* O(1) instead of O(N) when the policy plugin is getting conf for the known MeshService
  (`ToRules.ResourceRules` is a map, while `ToRules.Rules` is a list)
* Sorting/Merging algorithm is more intuitive and doesn't fall back upon alphabetical order of names in common use cases

### Negative Consequences

* 2 implementations of the algo that build rules.
  Existing function `BuildRules` builds subset rules,
  we need another `BuildResourceRules` method that potentially doesn't share much with `BuildRules`.
