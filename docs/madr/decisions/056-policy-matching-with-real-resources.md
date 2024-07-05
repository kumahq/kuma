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
In example, if we want to know the `Conf` for outbound with `kuma.io/service: bar` tags,
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

// UniqueResourceKey is a way to uniquely identify a resource, should include ResourceType, Name and Mesh
type UniqueResourceKey struct{}

type ResourceRule struct {
    Resource core_model.ResourceMeta
    Conf     interface{}
    Origin   []core_model.ResourceMeta
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
  - resource:
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
  - resource:
      name: finance-backend.finance
      mesh: mesh-1
    conf: conf-1
    origin: [...]
  - resource:
      name: finance-frontend.finance
      mesh: mesh-1
    conf: conf-1
    origin: [...]
  - resource:
      name: finance-db.finance
      mesh: mesh-1
    conf: conf-1
    origin: [...]
```

Policy plugins should use old `ToRules.Rules` when computing configurations for legacy destinations with `kuma.io/service` tags.
Policy plugins should use new `ToRules.ResourceRule` when computing configuration for the new MeshService resources.
Eventually we're going to delete `ToRules.Rules`.

When it comes to the Inspect API, we're going to add a new field to `InspectRule`

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
    conf:
      description: The actual conf generated
      type: object
      x-go-type: 'interface{}'
    origin:
      type: array
      items:
        $ref: '#/components/schemas/Meta'
```

### Positive Consequences

* In Kuma GUI, it's possible to make links to the real MeshServices, MeshHTTPRoutes and MeshExternalServices when used in `to[]`
* Policy plugins code is simpler with the clean fallback logic
* O(1) instead of O(N) when the policy plugin is getting conf for the known MeshService
  (`ToRules.ResourceRules` is a map, while `ToRules.Rules` is a list)

### Negative Consequences

* 2 implementations of the algo that build rules.
  Existing function `BuildRules` builds subset rules,
  we need another `BuildResourceRules` method that potentially doesn't share much with `BuildRules`.
