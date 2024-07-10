# Rework of Mesh*Routes matching

* Status: accepted

Technical story: https://github.com/kumahq/kuma/issues/10247

## Context and Problem Statement

At the moment when targeting Mesh*Route you need to put it in topLevel targetRef. 
We want to iterate over policy design to fully utilize the fact that we are referencing real objects.
The current implementation has made things a little tricky because policy matching is convoluted where the top level 
MeshHTTPRoute gets converted to a fake `to`.

## Considered options

- add possibility to put Mesh*Routes in `spec.to[].targetRef`
- leave Mesh*Routes in top level targetRef as it is now

## Decision Outcome

- add possibility to put Mesh*Routes in `spec.to[].targetRef`

### Positive consequences

- improved policy validation, we can enforce applying only http config on MeshHttpRoute
- simplifying topLevel targetRef as it will now match only whole proxies
- no need for specifying topLevel targetRef when targeting `Mesh`
- easier policy matching code

## Design

The design is grounded in multiple recent MADRs around improvements to policies: 

- [MeshService policy matching](https://github.com/kumahq/kuma/blob/master/docs/madr/decisions/046-meshservice-policy-matching.md)
- [Policy on Namespace](https://github.com/kumahq/kuma/blob/master/docs/madr/decisions/052-policy-on-namespace.md) - which introduced consumer/producer policies
- [Policy matching with real resources](https://github.com/kumahq/kuma/pull/10726/files)

### API

New API for policy that selects Mesh*Route

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

specifying `Mesh` in topLevel targetRef will be optional (this probably will be most used case). Minimal policy, targeting
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

To introduce this change we need to deprecate specifying Mesh*Route in topLevel targetRef.

Should we allow selecting MeshGateway in topLevel targetRef? probably this is obsolete as Route will select gateway, and we
want to apply policy on a route. Is there a need to make it more specific than route? 

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
    default:
      http:
        requestTimeout: 2s
        streamIdleTimeout: 1h
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
    default:
      http:
        requestTimeout: 10s
```

As described in targeting real resources MADR, while merging policies we will create `ResourceRules map[UniqueResourceKey]ResourceRule` 
field that holds configuration for a given resource identified by `UniqueResourceKey`. So when user creates policy that has 
real resource in `to` section we will create confs by resources

```go
map[UniqueResourceKey]ResourceRule{
    "backend.backend-ns": {  // this is just a placeholder name for visualisation. UniqueResourceKey is build from ResourceType, Name and Mesh, I will update it after real resource matching madr is merged
        requestTimeout: "2s"
        streamIdleTimeout: "1h"
    },
    "backend-route.backend-ns": {
		requestTimeout: "10s"
    },
}
```

Route configuration will be applied when route was created for a given DPP, if not MeshService config will be applied. 
Route may not be present if Mesh*Route was applied only for subset of proxies, for example:

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

After applying this two policies timeout won't be applied to any proxy, since Mesh*Route and MeshTimeout is applied
on different subset of proxies. We should detect this situation and add warning in Inspect API.

#### Overriding configuration

For policies on namespace this is as simple as creating consumer policy that targets MeshService or Mesh*Route:
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
    default:
      http:
        streamIdleTimeout: 2h
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
    default:
      http:
        requestTimeout: 15s
```

After merging conf we will get:

```go
map[UniqueResourceKey]ResourceRule{
    "backend.backend-ns": {
        requestTimeout: "2s" // from producer policy
        streamIdleTimeout: "2h" // from consumer policy
    },
    "backend-route.backend-ns": {
		requestTimeout: "15s" // from consumer policy
    },
}
```

Configuration priority (from most to least important): 
1. consumer policy targeting Mesh*Route
2. producer policy targeting Mesh*Route
3. consumer policy targeting MeshService
4. producer policy targeting MeshService

We always prioritize policies targeting Mesh*Route than the one targeting MeshService because route can have more specific
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
      default:
        http:
          requestTimeout: 2s
          streamIdleTimeout: 1h
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
      default:
        http:
          requestTimeout: 10s
```

To override this configuration for a service consuming this route we have to create policy with topLevel targetRef MeshSubset: 

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
      default:
        http:
          requestTimeout: 2s
          streamIdleTimeout: 1h
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
      default:
        http:
          requestTimeout: 10s
```

This will result in the same configuration as in previous example, because `MeshSubset` kind is more specific than `Mesh`:

```go
map[UniqueResourceKey]ResourceRule{
    "backend": {
        requestTimeout: "2s" // from timeout-on-backend-service policy
        streamIdleTimeout: "2h" // from timeout-on-backend-service-override policy
    },
    "backend-route": {
		requestTimeout: "15s" // from timeout-on-backend-route-override policy
    },
}
```

### Backward compatibility and migration

First we need to start by deprecating putting Mesh*Route in topLevel targetRef.
At the moment we are building `Rules` object with Subset for matching. This will change with implementation of 
`policy matching with real resources` MADR. For real resources we will build `ResourceRule` object. To smoothen the transition
we can still build old `Rules` for old policies and new `ResourceRule` for newly created policies with Mesh*Route in `to` section.
At the policy level we can prioritize new `ResourceRule` over old `Rules`. Thanks to this new way of targeting Mesh*Routes
will take precedence, so that users can easily switch to new targeting mechanism without any downtime.
