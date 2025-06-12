# MeshLoadBalancingStrategy move `HashPolicies` field out of LB specific structs

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/13436

## Context and Problem Statement

Today, `HashPolicies` field is located inside the LB specific configs, i.e.

```yaml
type: MeshLoadBalancingStrategy
name: mlbs-1
mesh: envoyconfig
labels:
  kuma.io/effect: shadow
spec:
  to:
    - targetRef:
        kind: MeshService
        name: test-server-1
      default:
        loadBalancer:
          type: RingHash # configured per-cluster
          ringHash:
            hashPolicies: # configured per-route
              - type: Header
                header:
                  name: x-header
    - targetRef:
        kind: MeshService
        name: test-server-2
      default:
        loadBalancer:
          type: Maglev # configured per-cluster
          maglev:
            hashPolicies: # configured per-route
              - type: Header
                header:
                  name: x-header
```

Load balancer is configured on the Envoy clusters, while hash policies are configured on the Envoy routes.

Today, we simply find all routes pointing to the cluster and update all of them with the same `hashPolicies`.
But it's a viable use case to configure different hash policies depending on the route.
That's why we want to add support for `to[].targetRef.kind: MeshHTTPRoute` for MeshLoadBalancingStrategy.

But with the existing API there is a problem:

* field `type: RingHash | Maglev | ...` is required and can't be empty
* field `hashPolicies` is nested inside LB specific fields `ringHash`, `maglev`

This means if user tries to specify `hashPolicies` for their route, they're forced to pick LB type:

```yaml
spec:
  to:
    - targetRef:
        kind: MeshHTTPRoute
        name: route-1
      default:
        loadBalancer:
          type: RingHash # has to be specified because 'type' is required
          ringHash:
            hashPolicies: [...]
```

Problems with the current approach:

* user has to think about LB type while configuring routes
* from Kuma the policies point of view MeshHTTPRoute conf has more priority than MeshService conf, 
but we can't guarantee that because multiple routes are sharing the same cluster 

## Design

### API change

We can bring `hashPolicies` field up to the `default` and deprecate existing `hashPolicies` fields:

```yaml
type: MeshLoadBalancingStrategy
name: mlbs-1
mesh: envoyconfig
spec:
  to:
    - targetRef:
        kind: MeshService
        name: test-server-1
      default:
        hashPolicies:
          - type: Header
            header:
              name: x-header
        loadBalancer:
          type: RingHash 
          ringHash:
            hashFunction: MurmurHash2
    - targetRef:
        kind: MeshService
        name: test-server-2
      default:
        hashPolicies:
          - type: Header
            header:
              name: x-header
        loadBalancer:
          type: Maglev
          maglev:
            tableSize: 1000
```

With `hashPolicies` field located at the `default`'s root, targeting MeshHTTPRoute is going to look like:

```yaml
type: MeshLoadBalancingStrategy
name: mlbs-1
mesh: envoyconfig
spec:
  to:
    - targetRef:
        kind: MeshHTTPRoute
        name: route-1
      default:
        hashPolicies:
          - type: Header
            header:
              name: x-header
```

### Migration

No implications for current MeshLoadBalancingStrategy users. 

Applying policies with `default.loadBalancer.ringHash.hashPolicies` or `default.loadBalancer.maglev.hashPolicies` 
will return warning message suggesting to use `default.hashPolicies`.

Mixing both fields shouldn't be allowed within the single `default` conf.

When both fields are mixed in the final conf, `default.hashPolicies` has more priority.

#### Example 1

The following config should result in validation error:

```yaml
spec:
  to:
    - targetRef:
        kind: MeshService
        name: test-server-1
      default:
        hashPolicies:
          - type: Header
            header:
              name: x-test-header-2
        loadBalancer:
          type: RingHash 
          ringHash:
            hashPolicies: # will be overridden by 'default.hashPolicies', no point in having both at the same conf
              - type: Header
                header:
                  name: x-test-header-1
```

#### Example 2

User already has a load balancing strategy in place:

```yaml
spec:
  to:
    - targetRef:
        kind: MeshService
        name: test-server-1
      default:
        loadBalancer:
          type: RingHash 
          ringHash:
            hashPolicies:
              - type: Header
                header:
                  name: x-test-header-1
```

They start using per-route configuration:

```yaml
type: MeshLoadBalancingStrategy
name: mlbs-1
mesh: envoyconfig
spec:
  to:
    - targetRef:
        kind: MeshHTTPRoute
        name: route-1 # route to 'test-server-1'
      default:
        hashPolicies:
          - type: Header
            header:
              name: x-test-header-2
```

As a result:
* when clients consume `test-server-1` on `route-1` hashing is performed on `x-test-header-2` values
* when clients consume `test-server-1` on other routes, hashing is performed on `x-test-header-1` values


## Implications for Kong Mesh

None

## Decision

We will move the `hashPolicies` field from load balancer-specific configurations (`ringHash`, `maglev`) to the `default` level in the MeshLoadBalancingStrategy API. This change:

1. Decouples hash policy configuration from load balancer type selection
2. Enables proper per-route hash policy configuration when targeting MeshHTTPRoutes
3. Maintains backward compatibility by supporting the old configuration format with deprecation warnings
4. Ensures that when both old and new formats are used, the new `default.hashPolicies` takes precedence

This design allows users to configure different hash policies for different routes while maintaining the appropriate load balancer type at the cluster level.
