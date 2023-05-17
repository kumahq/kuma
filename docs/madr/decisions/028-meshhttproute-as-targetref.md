# `MeshHTTPRoute` as `targetRef.kind`

- Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/5867

## Context and Problem Statement

There are Kuma policies such as MeshRateLimit, MeshRetry, MeshTimeout that support per-route configuration. 
Kuma provides a way to create an HTTP route with MeshHTTPRoute policy, but currently there is no way to
apply policies to the specific route.

### Goals

- Define the behaviour of `targetRef{kind: MeshHTTPRoute}` in the top-level targetRef.

### Non-goals

- MeshHTTPRoute creates only outbound (client-side) routes, that's why targeting MeshHTTPRoutes will be
  possible only for outbound policies (with `to[]` section). Configuring inbound route is not in the scope
  of this MADR.
- At this moment, MeshHTTPRoute doesn't work with MeshGateway. Defining behaviour of the MeshHTTPRoute for 
  MeshGateway is not in the scope of this MADR.


## Solution

### What are the allowed targetRefs inside `to[]` when top-level targetRef is MeshHTTPRoute?

#### Considered Options

- only `targetRef{kind: Mesh}` is allowed
- both `targetRef{kind: Mesh}` and `targetRef{kind: MeshService}` are allowed ✅

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

It's possible `MeshRetry` to target outbound that doesn't have `route-1`, i.e:

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

in that case we should not configure routes on the `backend` listener and show error status to the user
when [#5870](https://github.com/kumahq/kuma/issues/5870) is done.

#### Positive Consequences

- more fine-grained configuration

#### Negative Consequences

- could be surprising if policy targets outbound that wasn't targeted by route

### Do we allow reference individual rules inside the MeshHTTPRoute?

#### Considered Options

- Don't allow referencing individual rules
- Allow referencing individual rules in the future with `sectionName` ✅
- Implement `sectionName` straightaway

#### Decision Outcome

Chosen option: allow referencing individual rules in the future with `sectionName`

Single MeshHTTPRoute policy can have lots of rules, but thanks to the fact that new policies are mergeable it's
not that difficult to split single MeshHTTPRoute into multiple policies. Smaller MeshHTTPRoutes are more manageable
and easier to read. Lack of ability to target individual rules is not critical.

However, there is [experimental proposal](https://gateway-api.sigs.k8s.io/geps/gep-713/#apply-policies-to-sections-of-a-resource-future-extension) 
in Gateway API to introduce `sectionName` for targeting individual rules. For Kuma this would mean adding `sectionName` 
field inside the `TargetRef`.

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
