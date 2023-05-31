# `MeshHTTPRoute` as `targetRef.kind`

- Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/5867

## Context and Problem Statement

There are Kuma policies such as MeshRateLimit, MeshRetry, MeshTimeout that support per-route configuration. 
Kuma provides a way to create an HTTP route with MeshHTTPRoute policy, but currently there is no way to
apply policies to the specific route.

### Goals

- Define the behaviour of `targetRef{kind: MeshHTTPRoute}`

### Non-goals

- MeshHTTPRoute creates only outbound (client-side) routes, that's why targeting MeshHTTPRoutes will be
  possible only for outbound policies (with `to[]` section). Configuring inbound route is not in the scope
  of this MADR.
- At this moment, MeshHTTPRoute doesn't work with MeshGateway. Defining behaviour of the MeshHTTPRoute for 
  MeshGateway is not in the scope of this MADR.

## Solution

### Should `targetRef{kind: MeshHTTPRoute}` be used as a top-level targetRef or inside the `to[]`?

#### Considered Options

- use `targetRef{kind: MeshHTTPRoute}` inside the `to[]`
- use `targetRef{kind: MeshHTTPRoute}` as a top-level targetRef ✅

#### Decision Outcome

These approaches have different semantic. When top-level targetRef selects DPPs and `to[]` selects MeshHTTPRoute:

```yaml
type: MeshRetry
spec:
  targetRef:
    kind: MeshService
    name: backend
  to:
    - targetRef:
        kind: MeshHTTPRoute
        name: route-1
      default: 
        http:
          numRetries: 12
```

then MeshRetry policy is attached to all `backend` DPPs and only if there is `route-1` on the DPP then we configure it.
This approach potentially leads to the surprising behaviour when there is no `route-1` on the `backend` DPP
and so the retry configuration won't be applied. 

Another downside of this approach is `route-1` has retry policy only when applied on `backend` DPPs. Even if other 
services have `route-1` they won't get the same retry policy. If it's fair to assume that same routes most likely have 
the same configuration then this approach leads to the increasing the number of policies with duplicated configurations.

Alternatively, top-level targetRef can select `MeshHTTPRoute`:

```yaml
type: MeshRetry
spec:
  targetRef:
    kind: MeshHTTPRoute
    name: route-1
  to:
    - targetRef:
        kind: Mesh
      default: 
        http:
          numRetries: 12
```

In this case MeshRetry policy is attached to `route-1` and so it's attached to all DPPs with `route-1`.

#### Positive Consequences

- attaching policy to route affects all the DPPs where this route is applied
- less surprising behaviour, impossible by design to target non-existing route on DPP

#### Negative Consequences

- to figure if a policy applies to a DPP we need first to get the routes that affect the DPP

### What are the allowed targetRefs inside `to[]` when top-level targetRef is MeshHTTPRoute?

#### Considered Options

- only `targetRef{kind: Mesh}` is allowed
- allow only `targetRef{kind: Mesh}` in the beginning and later add support for `targetRef{kind: MeshService}` ✅
- both `targetRef{kind: Mesh}` and `targetRef{kind: MeshService}` are allowed

#### Decision Outcome

Single MeshHTTPRoute policy can define routes for multiple outbounds:

```yaml
type: MeshHTTPRoute
name: route-1
spec:
  targetRef: 
    kind: MeshService
    name: frontend
  to:
    - targetRef:
        kind: MeshService
        name: web
      rules: [...]
    - targetRef:
        kind: MeshService
        name: orders
      rules: [...]
```

so it makes sense to let the policy specify what outbound listener is affected:

```yaml
type: MeshRetry
spec:
  targetRef:
    kind: MeshHTTPRoute
    name: route-1
  to:
    - targetRef:
        kind: MeshService
        name: web
      default: 
        http:
          numRetries: 10 # only route on 'web' outbound listener is affected ('orders' is not affected)
```

But now it's possible for `MeshRetry` to target outbound that doesn't have `route-1`, i.e:

```yaml
type: MeshRetry
spec:
  targetRef:
    kind: MeshHTTPRoute
    name: route-1
  to:
    - targetRef:
        kind: MeshService
        name: backend
      default: 
        http:
          numRetries: 12
```

in that case we should not configure routes on the `backend` listener and show error status to the user.
Since [#5870](https://github.com/kumahq/kuma/issues/5870) is not implemented and in order to reduce initial complexity
the chosen option is implementing `targetRef{kind: Mesh}` first and adding support for `targetRef{kind: MeshService}`
in the future if necessary.

#### Positive Consequences

- less surprising behaviour when targeting the outbound that wasn't targeted by route

#### Negative Consequences

- less fine-grained configuration

### Do we allow reference individual rules inside the MeshHTTPRoute?

#### Considered Options

- Don't allow referencing individual rules
- Allow referencing individual rules in the future with `sectionName` ✅
- Implement `sectionName` straightaway

#### Decision Outcome

Single MeshHTTPRoute policy can have lots of rules, but thanks to the fact that new policies are mergeable it's
not that difficult to split single MeshHTTPRoute into multiple policies. Smaller MeshHTTPRoutes are more manageable
and easier to read. Lack of ability to target individual rules is not critical.

However, there is [experimental proposal](https://gateway-api.sigs.k8s.io/geps/gep-713/#apply-policies-to-sections-of-a-resource-future-extension) 
in Gateway API to introduce `sectionName` for targeting individual rules. For Kuma this would mean adding `sectionName` 
field inside the `TargetRef`.

To sum up, we can live without the ability to target individual rules and given the fact that Gateway API proposal 
is still experimental we should provide initial implementation without the `sectionName`.

### Validation

It's important to introduce extra validation for all policies that support `spec.targetRef.kind: MeshHTTPRoute`. 
When targeting MeshHTTPRoute conf should specify only fields that are configured on routes:

```yaml
type: MeshTimeout
spec:
  targetRef:
    kind: MeshHTTPRoute
    name: route-1
  to:
    - targetRef:
        kind: Mesh
      default:
        connectTimeout: 10s # validation error, this field configures Envoy cluster 
        http:
          requestTimeout: 5s
```

### Referencing the same route from multiple policies

When several MeshHTTPRoute policies target the same DPP the resulted configuration is a merge of configuration
from these policies. For example: 

```yaml
type: MeshHTTPRoute
name: route-1
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshService
        name: web
      rules:
        - matches:
            - path:
                type: Exact
                value: /
          default:
            filters:
              - type: RequestHeaderModifier
                requestHeaderModifier:
                  set:
                    - name: x-custom-header
                      value: xyz
---
type: MeshHTTPRoute
name: route-2
spec:
  targetRef: 
    kind: MeshService
    name: frontend
  to:
    - targetRef:
        kind: MeshService
        name: web
      rules:
        - matches:
            - path:
                type: Exact
                value: /
          default:
            backendRefs:
              - kind: MeshServiceSubset
                name: backend
                tags:
                  version: "1.0"
                weight: 90
              - kind: MeshServiceSubset
                name: backend
                tags:
                  version: "2.0"
                weight: 10
```

After merging, the configuration on `frontend` DPP will have only 1 HTTP route:

```yaml
rules:
  - matches:
      - path:
          type: Exact
          value: /
    default:
      filters:
        - type: RequestHeaderModifier
          requestHeaderModifier:
            set:
              - name: x-custom-header
                value: xyz
      backendRefs:
        - kind: MeshServiceSubset
          name: backend
          tags:
            version: "1.0"
          weight: 90
        - kind: MeshServiceSubset
          name: backend
          tags:
            version: "2.0"
          weight: 10
```

so we have situation when 2 MeshHTTPRoutes `route-1` and `route-2` ended up as a single route in Envoy.

This means when 2 different MeshRetry policies are targeting `route-1` and `route-2` they're competing for the same spot in Envoy:

```yaml
type: MeshRetry
name: mr-1
spec:
  targetRef:
    kind: MeshHTTPRoute
    name: route-1
  to:
    - targetRef:
        kind: Mesh
      default: 
        http:
          numRetries: 100
          retryOn:
            - "5xx"
---
type: MeshRetry
name: mr-2
spec:
  targetRef:
    kind: MeshHTTPRoute
    name: route-2
  to:
    - targetRef:
        kind: Mesh
      default:
        http:
          numRetries: 500
```

In order to solve the ambiguity we have to define the order in which the policies are applied. 
When 2 MeshRetry are targeting the same Envoy route the order should be decided based on the MeshHTTPRoutes order.
The process of computing configuration should be somewhat like:

1. Build routes map where keys are `matches[]` and values are `conf`.
2. For each key we can say what MeshHTTPRoutes contributed to it (in our case its `route-1` and `route-2`)
3. Both for `route-1` and `route-2` compute MeshRetry configurations
4. Merge MeshRetry configurations for `route-1` and `route-2` using the order "route-2 is more specific than route-1"
5. Resulted configuration could be sent to Envoy

In our example we'll get the following configuration for the envoy route:

```yaml
default:
  http:
    numRetries: 500
    retryOn:
      - "5xx"
```

### Inspect API

In Kuma GUI in `To` column we have to display not only the outbound's name, but also `rule[].matches` list, because
this list uniquely identifies routes in Envoy:

```yaml
total: 1
items:
  - type: DestinationSubset
    name: backend
    addresses:
      - 10.3.2.3:2300
      - 240.0.0.0:80
    service: backend
    tags:
      kuma.io/service: backend
    policyType: MeshRetry
    subset: {}
    matches: # new field only for DestinationSubsets
      - path:
          value: "/v1"
          type: PathPrefix
    conf:
      http:
        numRetries: 10
    origins:
      - mesh: default
        name: mal-1
```
