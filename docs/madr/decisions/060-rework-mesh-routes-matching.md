# Rework of Mesh*Routes matching

* Status: accepted

Technical story: https://github.com/kumahq/kuma/issues/10247

## Context and Problem Statement

This MADR is focused to targeting outbound routes configured on client proxy. 

At the moment when targeting `Mesh*Route` you need to put it in topLevel targetRef. 
We want to iterate over policy design and move `Mesh*Route` to `spec.to[].targetRef` section follow model where `spec.targetRef`
selects whole proxies, `spec.from[].targetRef` selects clients and `spec.to[].targetRef` selects outbounds.


## Decision Outcome

- add possibility to put `Mesh*Routes` in `spec.to[].targetRef`

### Positive Consequences

* We have a relatively small migration
* a policy can target a subset of the sidecars that a `MeshHTTPRoute` targets
  * this is especially important on universal where the zone isn't granular enough and there are no namespaces to implicitly limit the effects of a policy
* easy to target multiple `MeshHTTPRoutes` with a policy

### Negative Consequences

* `spec.targetRef` **must be combined** with `spec.to[].targetRef` to figure out which sidecars are actually affected since policies referenced in `spec.to[].targetRef` also select sidecars.
* `spec.targetRef`, given that it's likely to be empty, is relegated to not being very useful on its own for figuring out what a policy affects in the policy attachment sense
## Considered options

1. add possibility to put `Mesh*Routes` in `spec.to[].targetRef` -> `spec.targetRef` selects proxies, `spec.from[].targetRef` selects inbounds, `spec.to[].targetRef` selects outbounds.
2. leave `Mesh*Routes` in top level targetRef as it is now, improve the semantics of `spec.targetRef` and introduce `spec.workloadSelector` to select whole proxies

### Add possibility to put Mesh*Routes in `spec.to[].targetRef`

#### Advantages

- ability to set different confs on different proxy for the same route (overriding default conf for route without producer/consumer policies, on universal)
- gives more flexibility as you can target multiple outbounds in single policy. Reducing number of consumer policies
- easier policy matching code
- least problematic migration

#### Disadvantages

- can be harder to understand because `spec.targetRef` can be MeshSubset (which is optional and probably will rarely be used)
- which sidecars are actually configured depends on both `spec.targetRef` of the policy _and_ the `spec.targetRef` of the `MeshHTTPRoute`.
- yet another migration for users (only for MeshTimeout system policy as for namespaced policies this is already not allowed) 
- mixing outbounds and routes in `spec.to[].targetRef`
- hides the most often most important information under `to[]` adding noise when targeting Mesh in `spec.targetRef`
- is it consistent in the case of MeshGateways? MeshHTTPRoutes aren't on outbound listeners in the case of MeshGateways, they're on inbound listeners

### Leave Mesh\*Routes in top level targetRef as it is now and refine semantics of `spec.targetRef`

Instead of having `targetRef` always refer to a workload,
`targetRef` selects some part of the Envoy configuration of a workload,
where that part is represented by a real Kuma resource.

In order to allow limiting the effects of a policy further we can have
`spec.workloadRef`.

Maybe `targetRef` is the wrong concept for selecting workloads,
since "targetRef refers to a workload" means we aren't targeting
an arbitrary `Kind` of resource but instead always a `Dataplane` or a
`Gateway`, it doesn't depend on the policy. We could potentially introduce
`spec.workloadSelector` which would basically be some variation of `labelSelector`,
though how this works with Dataplane vs Gateway is not clear, maybe some kind of
`spec.gatewaySelector`.

What's not clear in this model is how to handle policies that can reasonably be both inbound
and outbound, such as `MeshTimeout`, and be applied to the whole mesh.
In order to target the whole mesh, we'd need to split such a policy
into `InboundTimeout` and `OutboundTimeout` and set `spec.targetRef.kind: Mesh`.
Otherwise a simple `spec.targetRef.kind: Mesh` wouldn't be enough to determine
whether an inbound or outbound timeout is wanted.

#### Advantages:

- the workloads targeted by the policy is often "all workloads" or can be
  inferred from the namespace so we don't waste the top level, "special" `spec.targetRef`
- Doesn't force every potential resource into a "to" or a "from", instead using the semantics of the `kind` itself
	- `HTTPRoute` is a route, regardless of whether it's inbound or outbound
        - in case the route itself targes a `MeshService`, it is an outbound route
        - if it targets a `MeshGateway`, it's a gateway route
        - whatever an inbound route is, it would target a different resource
        - if I'm targeting a route, I don't need to specify on the policy again
          whether it's inbound or outbound
	- `MeshService` _is_ a destination so `spec.targetRef.kind: MeshService` targets the configuration for a given destination
	- `WorkloadIdentity` (obviously doesn't exist yet but we want to change how we manage identity)
      would represent an identity so fits as `spec.targetRef` for e.g. `MeshTrafficPermission`

#### Disadvantages:

- `targetRef` is a single field, so it's not easy to share configuration
  in a single resource to avoid duplication.
  Though the very WIP policy attachment GEPs do propose a `targetRefs` field
- requires a huge migration

### Don't use `spec.targetRef` and have `spec.sourceTargetRef` and `spec.destinationTargetRef`

It's clear that it's very important when applying a policy, that we often need
to apply the policy based on its source _and_ its destination.
We always end up having to choose one of these to put in `spec.targetRef`. In
the `MeshTimeout.to` case it's the source of the traffic. With
`MeshTrafficPermission` it's the destination.

But it could be argued that this distinction is an implementation detail. The user doesn't need
to know which sidecar is used to apply the policy, as long as its effects are
as desired.

Therefore maybe the usage of our policies would be most clear if the user
didn't sometimes see `to` and sometimes `from` and instead always thought about
both source and destination equally.

## Comparing UX of two approaches

In both cases we assume that either `spec.targetRef` or `spec.workloadSelector` selects whole Mesh which can be omitted as default.
With examples fist yaml is always with `to` section and second is with `spec.workloadSelector` approach.

Looking at the examples we can see that when we omit selecting Mesh as default both approaches look almost the same. With 
first approach giving a little bit more flexibility as `spec.to[]` is a list.

### Examples

1. Producer timeout to MeshService

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: timeout-on-backend
  namespace: backend-ns
  labels:
    kuma.io/mesh: default
spec:
  to:
  - targetRef:
      kind: MeshService
      name: backend
    default:
      http:
        requestTimeout: 10s
```

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: timeout-on-backend
  namespace: backend-ns
  labels:
    kuma.io/mesh: default
spec:
  targetRef:
    kind: MeshService
    name: backend
  default:
    http:
      requestTimeout: 10s
```

2. Consumer timeout to MeshService

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: timeout-on-backend
  namespace: frontend-ns
  labels:
    kuma.io/mesh: default
spec:
  to:
  - targetRef:
      kind: MeshService
      name: backend
      namespace: backend-ns
    default:
      http:
        requestTimeout: 10s
```

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: timeout-on-backend
  namespace: frontend-ns
  labels:
    kuma.io/mesh: default
spec:
  targetRef:
    kind: MeshService
    name: backend
    namespace: backend-ns
  default:
    http:
      requestTimeout: 10s
```

3. Consumer timeout to MeshService applied to labeled proxies

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: timeout-on-backend
  namespace: frontend-ns
  labels:
    kuma.io/mesh: default
spec:
  targetRef:
    kind: MeshSubset
    tag:
      service-type: ui
  to:
  - targetRef:
      kind: MeshService
      name: backend
      namespace: backend-ns
    default:
      http:
        requestTimeout: 10s
```

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: timeout-on-backend
  namespace: frontend-ns
  labels:
    kuma.io/mesh: default
spec:
  workloadSelector:
    labels:
      service-type: ui
  targetRef:
    kind: MeshService
    name: backend
    namespace: backend-ns
  default:
    http:
      requestTimeout: 10s
```

4. Producer timeout to MeshHTTPRoute

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: timeout-on-backend
  namespace: backend-ns
  labels:
    kuma.io/mesh: default
spec:
  to:
  - targetRef:
      kind: MeshHTTPRoute
      name: backend-route
    default:
      http:
        requestTimeout: 10s
```

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: timeout-on-backend
  namespace: backend-ns
  labels:
    kuma.io/mesh: default
spec:
  targetRef:
    kind: MeshHTTPRoute
    name: backend-route
  default:
    http:
      requestTimeout: 10s
```

5. Consumer timeout to MeshHTTPRoute

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: timeout-on-backend
  namespace: frontend-ns
  labels:
    kuma.io/mesh: default
spec:
  to:
  - targetRef:
      kind: MeshHTTPRoute
      name: backend-route
      namespace: backend-ns
    default:
      http:
        requestTimeout: 10s
```

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: timeout-on-backend
  namespace: frontend-ns
  labels:
    kuma.io/mesh: default
spec:
  targetRef:
    kind: MeshHTTPRoute
    name: backend-route
    namespace: backend-ns
  default:
    http:
      requestTimeout: 10s
```

6. Consumer timeout to MeshHTTPRoute applied to labeled proxies

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: timeout-on-backend
  namespace: frontend-ns
  labels:
    kuma.io/mesh: default
spec:
  targetRef:
    kind: MeshSubset
    tag:
      service-type: ui
  to:
  - targetRef:
      kind: MeshHTTPRoute
      name: backend-route
      namespace: backend-ns
    default:
      http:
        requestTimeout: 10s
```

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: timeout-on-backend
  namespace: frontend-ns
  labels:
    kuma.io/mesh: default
spec:
  workloadSelector:
    labels:
      service-type: ui
  targetRef:
    kind: MeshHTTPRoute
    name: backend-route
    namespace: backend-ns
  default:
    http:
      requestTimeout: 10s
```

7. Consumer timeout to both MeshService and MeshHttpRoute

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: timeout-on-backend
  namespace: frontend-ns
  labels:
    kuma.io/mesh: default
spec:
  to:
  - targetRef:
      kind: MeshService
      name: backend
      namespace: backend-ns
    default:
      http:
        requestTimeout: 5s
  - targetRef:
      kind: MeshHTTPRoute
      name: slow-backend-route
      namespace: backend-ns
    default:
      http:
        requestTimeout: 15s
```

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: timeout-on-backend
  namespace: frontend-ns
  labels:
    kuma.io/mesh: default
spec:
  targetRef:
    kind: MeshService
    name: backend
    namespace: backend-ns
  default:
    http:
      requestTimeout: 5s
---
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: timeout-on-backend
  namespace: frontend-ns
  labels:
    kuma.io/mesh: default
spec:
  targetRef:
    kind: MeshHTTPRoute
    name: slow-backend-route
    namespace: backend-ns
  default:
    http:
      requestTimeout: 15s
```

When looking at functionality both approaches covers all our examples. We can imagine moving in the future to `apiVersion: v1`, 
and translating things over to new version.

## Mesh*Route in spec.to[].targetRef design

The design is grounded in multiple recent MADRs around improvements to policies: 

- [MeshService policy matching](https://github.com/kumahq/kuma/blob/master/docs/madr/decisions/046-meshservice-policy-matching.md)
- [Policy on Namespace](https://github.com/kumahq/kuma/blob/master/docs/madr/decisions/052-policy-on-namespace.md) - which introduced consumer/producer policies
- [Policy matching with real resources](https://github.com/kumahq/kuma/pull/10726/files)

### API

New API for policy that selects `Mesh*Route`

```yaml
type: MeshTimeout
name: timeouts-on-route
mesh: default
spec:
  targetRef:
    kind: Mesh | MeshSubset # either Mesh, MeshSubset or MeshGateway
  to:
    - targetRef:
        kind: MeshHTTPRoute | MeshTCPRoute # either MeshHTTPRoute or MeshTCPRoute
        name: route-to-backend
      default:
        http:
          requestTimeout: 5s
          streamIdleTimeout: 1h
```

specifying `Mesh` in `spec.targetRef` will be optional (this probably will be most used case). Minimal policy, targeting
`MeshHTTPRoute` will look like this:

```yaml
type: MeshTimeout
name: timeouts-on-route
mesh: default
spec:
  to:
    - targetRef:
        kind: MeshHTTPRoute
        name: route-to-backend
      default:
        http:
          requestTimeout: 5s
          streamIdleTimeout: 1h
```

To introduce this change we need to deprecate specifying `Mesh*Route` in `spec.targetRef`.

#### Additional validation

When referencing Mesh*Route we should validate policy configuration to make sure user is applying only configuration 
that can be applied on route.

### Affected policies

Policies that can be configured per route: 

- `MeshTimeout` - at the moment only implemented policy, we should enforce validation to only accept config that can be set per route.  
- `MeshRetry` - would need to be implemented. Like timeout should only accept route related config.
- `MeshLoadBalancingStrategy` - would need to be implemented. Like timeout should only accept route related config.
- `MeshRateLimit` - while it's possible to set rate limit on a route. We currently only set rate limits on the inbound so 
supporting `spec.to[].target.kind: Mesh*Route` doesn't make sense.
- `MeshAccessLog` - would need to be implemented, with access log filter

#### Producer/consumer model for targeting Mesh*Route in other policies

For MeshRetry our producer/consumer makes sense since producer config is a suggestion from owner of the service. If consumer
wants to override it because they don't want to retry, they just want to fail fast or they want to retry on bigger number of errors
they can configure it by applying consumer policy.

MeshRetry example:

```yaml
# Producer policy with retry on 500 HTTP error code
apiVersion: kuma.io/v1alpha1
kind: MeshRetry
metadata:
  name: producer-retry
  namespace: backend-ns
  labels:
    kuma.io/mesh: default
spec:
  to:
    - targetRef:
        kind: MeshHTTPRoute
        name: backend-route # backend-route must be a producer route, for it to be synced to other zones and work
      default:
        http:
          numRetries: 3
          backOff:
            baseInterval: 10ms
            maxInterval: 1s
          retryOn:
            - "500"
---
# Consumer policy with override to retry on all 5xx errors
apiVersion: kuma.io/v1alpha1
kind: MeshRetry
metadata:
  name: consumer-retry
  namespace: frontend-ns
  labels:
    kuma.io/mesh: default
spec:
  to:
    - targetRef:
        kind: MeshHTTPRoute
        name: backend-route
        namespace: backend-ns
      default:
        http:
          retryOn:
            - "5xx"
```

For MeshLoadBalancingStrategy adding producer/consumer model can be problematic, for load balancing and hashing. Since this
is configured on outbound it falls into producer/consumer policy model. But who should have more power on deciding hashing rules?
It should be producer who is an owner of service who decides how to apply this configuration which is not compliant
with our consumer/producer model. 

For example, we could have a service that each instance is responsible for handling specific clients. We can use `RingHash`
to make sure that specific clients requests are processed by the same instance. This should be configured in producer policy. 
Changing this at consumer level can influence correctness of computations in our service. We've decided to allow this as 
consumer can broke things in other ways and they will be affecting only themselves

### Merging and applying configurations

Let's start with an example:

```yaml
# Producer route in backend-ns namespace
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: route-to-backend
  namespace: backend-ns
  labels:
    kuma.io/mesh: default
spec:
  to:
  - targetRef:
      kind: MeshService
      name: backend
    rules:
    - matches:
      - path:
          type: PathPrefix
          value: "/slow-endpoint"
      default:
        backendRefs:
          - kind: MeshService
            name: backend
---
# Producer MeshTimeout targeting whole MeshService
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: timeout-on-backend-service
  namespace: backend-ns
  labels:
    kuma.io/mesh: default
spec:
  to:
  - targetRef:
      kind: MeshService
      name: backend
    default: conf1
---
# Producer MeshTimeout targeting producer route
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: timeout-on-backend-route
  namespace: backend-ns
  labels:
    kuma.io/mesh: default
spec:
  to:
  - targetRef:
      kind: MeshHTTPRoute
      name: route-to-backend
    default: producer_conf
# Consumer MeshTimeout targeting producer route
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: timeout-on-backend-route
  namespace: frontend-ns
  labels:
    kuma.io/mesh: default
spec:
  to:
    - targetRef:
        kind: MeshHTTPRoute
        name: route-to-backend
      default: consumer_conf
```

All of these policies have `spec.targetRef` `Mesh` and these are producer policies so this configuration will be applied 
to all dataplanes in a mesh. 
As described in targeting real resources MADR, while merging policies we will create `ResourceRules map[UniqueResourceKey]ResourceRule` 
field that holds configuration for a given resource identified by `UniqueResourceKey`. So when user creates policy that has 
real resource in `spec.to[].targetRef` we will create confs by resources: (this is configuration that will be applied to selected dataplane)

```yaml
resourceRules:
  backend.backend-ns:
    conf: $conf1
  backend-route.backend-ns:
    conf: merge($producer_conf, $consumer_conf)
```

Because of this when inspecting configuration we will get only route specific config, without defaults from Mesh or
MeshService as they could not be even applicable on route. 

Route configuration will be applied when route was created for a given DPP, if not MeshService config will be applied. 
Route may not be present if `Mesh*Route` was applied only for subset of proxies, for example:

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: route-to-backend
  namespace: backend-ns
  labels:
    kuma.io/mesh: default
spec:
  targetRef:
    kind: MeshSubset
    tags:
      k8s.kuma.io/namespace: frontend-ns
---
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: timout-for-backend
  namespace: backend-ns
  labels:
    kuma.io/mesh: default
spec:
  targetRef:
    kind: MeshSubset
    tags:
      k8s.kuma.io/namespace: some-other-ns
  to:
    - targetRef:
        kind: MeshHTTPRoute
        name: route-to-backend
```

After applying these two policies timeout won't be applied to any proxy, since `Mesh*Route` and MeshTimeout is applied
on different subset of proxies. We should ignore this situation as this will just have no effect. Also, we don't have possibility
to show warnings with this information to users. 

#### Mesh*Route with multiple spec.to[].targetRef

At the moment users can create `Mesh*Route` that references multiple MeshServices, like in example:

```yaml
kind: MeshHTTPRoute
metadata:
  name: route-1
spec:
  to:
    - targetRef:
        kind: MeshService
        name: backend
      rules: [...]
    - targetRef:
        kind: MeshService
        name: frontend
      rules: [...]
```

When we create MeshTimeout policy that configures `Mesh*Route` and each of `MeshService` when merging configuration we will
have problem which conf should we pick for merging. With example timeout policy:

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: timeout-on-backend
  namespace: backend-ns
  labels:
    kuma.io/mesh: default
spec:
  to:
  - targetRef:
      kind: MeshService
      name: backend
    default: conf1
  - targetRef:
      kind: MeshService
      name: frontend
    default: conf2
  - targetRef:
      kind: MeshHTTPRoute
      name: slow-backend-route
      namespace: backend-ns
    default: conf3
```

When merging we will merge `$conf1` and `$conf2` for MeshService with `$conf3` for `Mesh*Route`. After changes to `ResourceRules`
which no longer merge confs for different resources we will get config:

```yaml
resourceRules:
  backend.backend-ns: 
    conf: $conf1
  frontend.backend-ns:
    conf: $conf2
  backend-route.backend-ns:
    conf: $conf3
```

Decision on which conf should be applied should be made on plugin level. As we know to which MeshService outbound we are 
applying configuration.

There is also question of deprecating and later on removing possibility to put multiple items in `spec.to[]`
section of `Mesh*Route`. This will align `Mesh*Route` design with GatewayAPI. In the future we should also discuss removing 
`spec.to[]` section entirely leaving only `spec.targetRef` section in `Mesh*Route`. This is tracked by issue: https://github.com/kumahq/kuma/issues/11021

#### Overriding configuration

For policies on namespace this is as simple as creating consumer policy that targets MeshService or `Mesh*Route`:
(This can be useful when we consume MeshService from other zone)

```yaml
# Producer MeshTimeout targeting whole MeshHTTPRoute
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: default-timeout-on-backend-route
  namespace: backend-ns
  labels:
    kuma.io/mesh: default
spec:
  to:
  - targetRef:
      kind: MeshHTTPRoute
      name: route-to-backend
      namespace: backend-ns
    default: $producer_conf
---
# Consumer MeshTimeout targeting producer route
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: timeout-on-backend-route
  namespace: frontend-ns
  labels:
    kuma.io/mesh: default
spec:
  to:
  - targetRef:
      kind: MeshHTTPRoute
      name: route-to-backend
      namespace: backend-ns
    default: $consumer_conf
```

After merging conf we will get:

```yaml
resourceRules:
  backend-route.backend-ns:
    conf: merge($producer_conf, $consumer_conf)
```

Configuration priority (from most to least important):
1. consumer policy targeting consumer `Mesh*Route`
2. consumer policy targeting producer `Mesh*Route`
3. producer policy targeting producer `Mesh*Route`
4. consumer policy targeting `MeshService`
5. producer policy targeting `MeshService`

**When configuring routes consumer route always takes precedence over producer route** 

We always prioritize policies targeting `Mesh*Route` than the one targeting MeshService because route can have more specific
requirements than a whole MeshService.

##### Overriding on universal

Since all policies on universal are system policies (we don't have consumer/producer policies as we don't have namespaces),
our example will look like: 

```yaml
# route to backend MeshService
type: MeshHTTPRoute
name: route-to-backend
mesh: default
spec:
  to:
    - targetRef:
        kind: MeshService
        name: backend
      rules:
        - matches:
            - path:
                type: PathPrefix
                value: "/slow-endpoint"
          default:
            backendRefs:
              - kind: MeshService
                name: backend
---
# MeshTimeout targeting whole MeshService
type: MeshTimeout
name: timeout-on-backend-service
mesh: default
spec:
  to:
    - targetRef:
        kind: MeshService
        name: backend
      default: conf1
---
# MeshTimeout targeting route
type: MeshTimeout
name: timeout-on-backend-route
mesh: default
spec:
  to:
    - targetRef:
        kind: MeshHTTPRoute
        name: route-to-backend
      default: conf2
```

To override this configuration for a service consuming this route we have to create policy with `spec.targetRef` MeshSubset: 

```yaml
# MeshTimeout targeting whole MeshService
type: MeshTimeout
name: timeout-on-backend-service-override
mesh: default
spec:
  targetRef:
    kind: MeshSubset
    tags:
      services-group: frontend # custom tag on fronted services
  to:
    - targetRef:
        kind: MeshService
        name: backend
      default: conf3
---
# MeshTimeout targeting route
type: MeshTimeout
name: timeout-on-backend-route-override
mesh: default
spec:
  targetRef:
    kind: MeshSubset
    tags:
      services-group: frontend # custom tag on fronted services
  to:
    - targetRef:
        kind: MeshHTTPRoute
        name: route-to-backend
      default: conf4
```

This will result in the same configuration as in previous example, because `MeshSubset` kind is more specific than `Mesh`:

```yaml
resourceRules:
  backend:
    conf: merge($conf1, $conf3)
  backend-route:
    conf: merge($conf2, $conf4)
```

### Backward compatibility and migration

First we need to start by deprecating putting `Mesh*Route` in `spec.targetRef`.
At the moment we are building `Rules` object with Subset for matching. This will change with implementation of 
`policy matching with real resources` MADR. For real resources we will build `ResourceRule` object. To smoothen the transition
we can still build old `Rules` for old policies and new `ResourceRule` for newly created policies with `Mesh*Route` in `spec.to[].targetRef`.
At the policy level we can prioritize new `ResourceRule` over old `Rules`. Thanks to this new way of targeting `Mesh*Routes`
will take precedence, so that users can easily switch to new targeting mechanism without any downtime.

With migration, we can take advantage of producer/consumer policies. We can enforce rules from this MADR like:

- putting `Mesh*Route` in `spec.to[].targetRef`
- single entry in `Mesh*Route` policy `spec.to[]` section

in producer/consumer policies, as no one is using them at the moment, and deprecate it in system policies.


## Addendum

### Role of the policy that references MeshHTTPRoute in spec.to[].targetRef

Given MeshTimeout references MeshHTTPRoute with `name` and `namespace`, i.e.:

```yaml
kind: MeshTimeout
metadata:
  name: timeout-1
  namespace: app-ns
spec:
  to:
    - targetRef:
        kind: MeshHTTPRoute
        name: route-1
        namespace: app-ns # can be omitted
```

How do we pick a policy role for `timeout-1`?

The only thing that matters is: if `route-1` is a producer policy, 
then `timeout-1` must also be a producer policy – otherwise, 
`route-1` gets synced to other zones without the right configuration.

There are two ways we can determine whether a policy that references a MeshHTTPRoute is a producer policy:

1. The same principle we use for referencing Mesh*Service: if the policy is in the same namespace as the MeshHTTPRoute, it's considered a producer policy.
2. A stricter approach: the policy is in the same namespace **and** the MeshHTTPRoute itself is a producer policy — only then is the policy considered a producer policy.

#### Pros and Cons of Option 1

* Good, because it's simple and consistent with existing behavior
* Bad, because a policy referencing a consumer MeshHTTPRoute would still be considered a producer policy — harmless, but potentially confusing.

#### Pros and Cons of Option 2

* Good, because it's less surprising behavior
* Bad, because requires resolving the referenced MeshHTTPRoute to determine the policy role. What if the route doesn’t exist? What if its role changes over time?  

#### Decision outcome

We chose Option 1. 
The same kind of confusion can already happen when referencing Mesh*Service — see [this issue](https://github.com/kumahq/kuma/issues/11570).
We just need to establish and document a clear rule of thumb: 
when creating a consumer policy, 
always prefer using `labels` over `name/namespace` when referencing the resource.
