# Moving `HashPolicies` Field Out of Load Balancer Specific Structs in MeshLoadBalancingStrategy

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/13436

## Context and Problem Statement

Currently, the `HashPolicies` field is located inside the load balancer specific configurations, as shown below:

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

Load balancers are configured on Envoy clusters, while hash policies are configured on Envoy routes.

Currently, we find all routes pointing to a cluster and update them with the same `hashPolicies`.
However, there's a legitimate use case for configuring different hash policies depending on the specific route.
This is why we want to add support for `to[].targetRef.kind: MeshHTTPRoute` in MeshLoadBalancingStrategy.

The existing API presents two key challenges:

* The field `type: RingHash | Maglev | ...` is required and cannot be empty
* The `hashPolicies` field is nested inside load balancer specific fields (`ringHash`, `maglev`)

This means when users attempt to specify `hashPolicies` for their routes, they are forced to select a load balancer type:

```yaml
spec:
  to:
    - targetRef:
        kind: MeshHTTPRoute
        name: route-1
      default:
        loadBalancer:
          type: RingHash # must be specified because 'type' is required
          ringHash:
            hashPolicies: [...]
```

The current approach presents two significant problems:

* Users must consider load balancer type when configuring routes, which creates unnecessary complexity
* From Kuma's policy perspective, MeshHTTPRoute configuration should have higher priority than MeshService configuration, 
but we cannot guarantee this behavior because multiple routes share the same cluster

## Design

### API Change

To address these issues, we propose elevating the `hashPolicies` field to the `default` level and deprecating the existing nested `hashPolicies` fields:

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

With the `hashPolicies` field relocated to the `default` level, targeting a MeshHTTPRoute becomes more intuitive and straightforward:

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

This approach properly decouples route-specific hash policies from cluster-level load balancer configuration.

### Migration

This change has no implications for current MeshLoadBalancingStrategy users.

When users apply policies with the legacy fields (`default.loadBalancer.ringHash.hashPolicies` or `default.loadBalancer.maglev.hashPolicies`), 
they will receive a warning message suggesting migration to the new `default.hashPolicies` field.

To maintain configuration clarity, mixing both old and new fields will not be allowed within a single `default` configuration.

In cases where both field types appear in the final configuration (e.g., through policy merging), the new `default.hashPolicies` field will take precedence.

#### Example 1: Validation Error

The following configuration would result in a validation error because it mixes both old and new hash policy fields within the same configuration:

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

This validation prevents ambiguous configurations and encourages users to adopt the new field structure.

#### Example 2: Migration Path for Existing Configurations

Consider a scenario where a user already has a load balancing strategy in place using the legacy format:

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

Later, they implement per-route configuration using the new format:

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

The final configurations after merging:

* `kri_msvc_envoyconfig_us-east-1_ns-1_test-server-1_`:

    ```yaml
    loadBalancer:
      type: RingHash 
      ringHash:
        hashPolicies:
          - type: Header
            header:
              name: x-test-header-1
    ```

* `kri_mhttpr_envoyconfig___route-1_`:

    ```yaml
    hashPolicies:
      - type: Header
        header:
          name: x-test-header-2 
    loadBalancer:
      type: RingHash 
      ringHash:
        hashPolicies:
          - type: Header
            header:
              name: x-test-header-1
    ```


This results in the following behavior:
* When clients access `test-server-1` through `route-1`, hashing is performed using the `x-test-header-2` values
* When clients access `test-server-1` through other routes, hashing is performed using the `x-test-header-1` values

## Implications for Kong Mesh

None

## Decision

We will relocate the `hashPolicies` field from load balancer-specific configurations (`ringHash`, `maglev`) to the `default` level in the MeshLoadBalancingStrategy API. This architectural improvement:

1. Decouples hash policy configuration from load balancer type selection, simplifying the user experience
2. Enables proper per-route hash policy configuration when targeting MeshHTTPRoutes, enhancing flexibility
3. Maintains backward compatibility by supporting the legacy configuration format with appropriate deprecation warnings
4. Establishes clear precedence rules when both old and new formats appear in the final configuration

This design empowers users to configure different hash policies for different routes while maintaining the appropriate load balancer type at the cluster level, resulting in a more intuitive and powerful API.
