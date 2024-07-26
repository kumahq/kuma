# Rework of Mesh*Routes matching

* Status: accepted

Technical story: https://github.com/kumahq/kuma/issues/10247

## Context and Problem Statement

At the moment when targeting `Mesh*Route` you need to put it in topLevel targetRef. 
We want to iterate over policy design and move `Mesh*Route` to `spec.to[].targetRef` section follow model where `spec.targetRef`
selects whole proxies, `spec.from[].targetRef` selects inbounds and `spec.to[].targetRef` selects outbounds

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

- simplifying policy api as `spec.targetRef` will always pick proxy and `spec.to[].targetRef` section will select outgoing traffic
- ability to set different confs on different proxy for the same route (overriding default conf for route without producer/consumer policies, on universal)
- gives more flexibility as you can target multiple outbounds in single policy. Reducing number of consumer policies
- more consistent with current approach
- easier policy matching code
- least problematic migration

#### Disadvantages

- can be harder to understand because `spec.targetRef` can be MeshSubset (which is optional and probably will rarely be used)
- which sidecars are actually configured depends on both `spec.targetRef` of the policy _and_ the `spec.targetRef` of the `MeshHTTPRoute`.
- yet another migration for users (only for system policies as for namespaced policies this is already not allowed) 
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

Should we allow selecting MeshGateway in `spec.targetRef` when `Mesh*Route` is configured in `spec.to[].targetRef`? probably this 
is obsolete as Route will select gateway, and we want to apply policy on a route. Is there a need to make it more specific than route? 
For example when user creates route targeting MeshGateway:

```yaml
kind: MeshHTTPRoute
metadata:
  name: demo-app-route
  namespace: kuma-system
  labels:
    kuma.io/origin: zone
    kuma.io/mesh: default
spec:
  targetRef:
    kind: MeshGateway
    name: demo-app
  to:
    - targetRef:
        kind: Mesh
      rules:
        - matches:
            - path:
                type: PathPrefix
                value: "/"
          default:
            backendRefs:
              - kind: MeshService
                name: demo-app_kuma-demo_svc_5000
```

then when you create timeout that targets this route: 

```yaml
kind: MeshTimeout
metadata:
  name: demo-app-timeout
  namespace: kuma-system
  labels:
    kuma.io/origin: zone
    kuma.io/mesh: default
spec:
  to:
    - targetRef:
        kind: MeshHTTPRoute
        name: demo-app-route
      default:
        requestTimeout: 5s
```

This route already will be created only on MeshGateway, so is there a need to specify anything else then a Mesh in `spec.targetRef`
of timeout policy?

#### Additional validation

When referencing Mesh*Route we should validate policy configuration to make sure user is applying only configuration 
that can be applied on route.

### Affected policies

At the moment we can only target `Mesh*Routes` in MeshTimeout policy. This MADR is based on MeshTimeout policy.
Looking at what can be configured at Envoy route we can also apply this to our MeshRetry policy. Rate limit is also configurable on route,
but we are configuring rate limit on inbound traffic, which does not apply to this MADR. `MeshLoadBalancingStrategy` also
configures hashing per route and we should probably add support for routes in the future

#### Producer/consumer model for targeting Mesh*Route in other policies

For MeshRetry our producer/consumer makes sense since producer config is a suggestion from owner of the service. If consumer
wants to override it because they don't want to retry, they just want to fail fast or the want to retry on bigger number of errors
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
        name: backend-route
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
I think it should be producer who is an owner of service who decides how to apply this configuration which is not compliant
with our consumer/producer model.

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
# MeshTimeout targeting producer route
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
    default: conf2
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
    conf: merge($conf1, $conf2)
```

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

After applying this two policies timeout won't be applied to any proxy, since `Mesh*Route` and MeshTimeout is applied
on different subset of proxies. We should ignore this situation as this will just have no effect. Also we don't have possibility
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

When merging we will merge `$conf1` and `$conf2` for MeshService with `$conf3` for `Mesh*Route`. 

```yaml
resourceRules:
  backend.backend-ns: 
    conf: $conf1
  frontend.backend-ns:
    conf: $conf2
  backend-route.backend-ns:
    conf: merge($conf1, $conf2, $conf3)
```

Which makes it problematic when applying conf since we are applying policy per outbound, and we have merged config for 
multiple outbounds. 

Possible solutions:

1. We can deprecate and then remove possibility to have multiple `spec.to[]` items in `Mesh*Route` policy to make it referencable. 
(this is basically step into removing `to` section altogether, why do we need it if we cannot put multiple items in it)
2. When merging we should only merge confs for given resource, don't treat `Mesh*Route` as subtype of `MeshService` and then
when applying configuration to resource we should be more mindful on which outbound the route is configured and configure 
it based on conf for that MeshService

#### Overriding configuration

For policies on namespace this is as simple as creating consumer policy that targets MeshService or `Mesh*Route`:
(This can be useful when we consume MeshService from other zone)

```yaml
# Consumer MeshTimeout targeting whole MeshService
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: timeout-on-backend-service
  namespace: frontend-ns
  labels:
    kuma.io/mesh: default
spec:
  to:
  - targetRef:
      kind: MeshService
      name: backend
      namespace: backend-ns
    default: $conf1
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
    default: $conf2
```

After merging conf we will get:

```yaml
resourceRules:
  backend.backend-ns: 
    conf: $conf1
  backend-route.backend-ns:
    conf: merge($conf1, $conf2)
```

Configuration priority (from most to least important): 
1. consumer policy targeting `Mesh*Route`
2. producer policy targeting `Mesh*Route`
3. consumer policy targeting MeshService
4. producer policy targeting MeshService

We always prioritize policies targeting `Mesh*Route` than the one targeting MeshService because route can have more specific
requirements than a whole MeshService

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
    conf: merge($conf1, $conf2, $conf3, $conf4)
```

### Backward compatibility and migration

First we need to start by deprecating putting `Mesh*Route` in `spec.targetRef`.
At the moment we are building `Rules` object with Subset for matching. This will change with implementation of 
`policy matching with real resources` MADR. For real resources we will build `ResourceRule` object. To smoothen the transition
we can still build old `Rules` for old policies and new `ResourceRule` for newly created policies with `Mesh*Route` in `spec.to[].targetRef`.
At the policy level we can prioritize new `ResourceRule` over old `Rules`. Thanks to this new way of targeting `Mesh*Routes`
will take precedence, so that users can easily switch to new targeting mechanism without any downtime.
