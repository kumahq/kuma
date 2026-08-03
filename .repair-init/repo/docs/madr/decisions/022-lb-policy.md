# Load balancing

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/5869

## Context and Problem Statement

There are 2 motivations for this document to exist.

Firstly, [MADR-011-mesh-traffic-route](./011-mesh-traffic-route.md) removed load balancing from the route policy:

> The main motivator for this is that load balancing is done on an Envoy cluster.
> Including load balancing options on a match would force creation of a new cluster
> for that set of backendRefs which has negative consequences including but not limited
> to metrics handling and a potentially large number of clusters.

This means we need a new policy for load balancing configuration.

Secondly, current way to toggle "locality-aware load balancing" is a
mesh-scoped [flag](https://kuma.io/docs/2.1.x/policies/locality-aware/#locality-aware-load-balancing).
This is not ideal, because there are situations when different services require different LB strategies.
That's why the new load balancing policy should contain functionality to enable "locality-aware load balancing".

## Considered Options

* MeshLoadBalancing

## Decision Outcome

Chosen option: "MeshLoadBalancing".

### MeshLoadBalancing

#### Top level targetRef

The policy can select any proxy that's why available values are `Mesh`,
`MeshSubset`, `MeshService`, and `MeshServiceSubset`.

#### To targetRef

The policy is applied on the Envoy outbound cluster, that's why available
values are `Mesh` and `MeshService`.

#### From targetRef

Doesn't make sense to apply load balancing on the inbound side.

#### Conf

```yaml
type: MeshLoadBalancing
name: lb-1
mesh: default
spec:
  targetRef:
    kind: Mesh | MeshSubset | MeshService | MeshServiceSubset
  to:
    - targetRef:
        kind: Mesh | MeshService
      default:
        localityAwareness: 
          disabled: true
        loadBalancer:
          type: LeastRequest
          leastRequest:
            choiceCount: 8
```

##### loadBalancer

**RoundRobin:**

```yaml
loadBalancer:
  type: RoundRobin
```

**LeastRequest:**

* choiceCount – The number of random healthy hosts from which the host with the fewest active requests will be chosen.
  Defaults to 2 so that we perform two-choice selection if the field is not set.

```yaml
loadBalancer:
  type: LeastRequest
  leastRequest:
    choiceCount: 8
```

**RingHash:**

* hashFunction – The hash function used to hash hosts onto the ketama ring.
  The value defaults to XX_HASH. Available values – XX_HASH, MURMUR_HASH_2.
* minRingSize – Minimum hash ring size. The larger the ring is (that is,
  the more hashes there are for each provided host) the better the request distribution
  will reflect the desired weights. Defaults to 1024 entries, and limited to 8M entries.
* maxRingSize – Maximum hash ring size. Defaults to 8M entries, and limited to 8M entries,
  but can be lowered to further constrain resource use.
* hashPolicies – specifies a list of request/connection properties that are used to calculate a hash.
  These hash policies are executed in the specified order. If a hash policy has the “terminal” attribute
  set to true, and there is already a hash generated, the hash is returned immediately,
  ignoring the rest of the hash policy list.

```yaml
loadBalancer:
  type: RingHash
  ringHash:
    hashFunction: MURMUR_HASH_2
    minRingSize: 64
    maxRingSize: 1024
    hashPolicies:
      - type: Header
        terminal: false
        header:
          name: header-name
      - type: Cookie
        terminal: false
        cookie:
          name: session
          ttl: 1h
          path: /
      - type: SourceIP
        terminal: false
      - type: Query
        terminal: true
        query:
          name: param
      - type: FilterState
        terminal: false
        filterState:
          key: objectNameInFilterState 
```

`hashPolicies` are configured on the Route that's why we have 2 limitations:

* can't apply different LB on different cluster where there is a MeshHTTPRoute split
* ringHash is an HTTP only policy

By default, Envoy calculates hash of the upstream host using address.
If we want something other than the host’s address to be used as the hash key 
(e.g. the semantic name of the host in a Kubernetes StatefulSet), 
then we can introduce a new DPP tag `kuma.io/hash` and set it in the "envoy.lb" LbEndpoint.Metadata e.g.:

```yaml
filter_metadata:
  envoy.lb:
    hash_key: "<value of 'kuma.io/hash' tag>"
```

**Random:**

```yaml
loadBalancer:
  type: Random
```

**Maglev:**

* tableSize – The table size for Maglev hashing. Maglev aims for “minimal disruption”
  rather than an absolute guarantee. Minimal disruption means that when the set of upstream hosts
  change, a connection will likely be sent to the same upstream as it was before.
  Increasing the table size reduces the amount of disruption. The table size must be prime number
  limited to 5000011. If it is not specified, the default is 65537.
* hashPolicies – see `RingHash.hashPolicies`

```yaml
loadBalancer:
  type: Maglev
  maglev:
    tableSize: 100
    hashPolicies: ... # see `RingHash.hashPolicies`
```

##### locality-aware load balancing

Today Kuma supports [locality-aware load balancing](https://kuma.io/docs/2.1.x/policies/locality-aware/#locality-aware-load-balancing).
But this feature has 2 major shortcomings:
* the flag is global for the entire mesh
* the actual Envoy locality-aware LB feature is not supported, see [issue #2689](https://github.com/kumahq/kuma/issues/2689)

Instead of mesh-scoped locality-aware toggle, we are going to create new conf in MeshLoadBalancing policy:

```yaml
localityAwareness:
  disabled: true
```

Since the new flag is going to work per-service, it makes more sense to make it opt-out. 
Enabled locality-aware LB is a better default option because distributing requests across all zones
can make a good demo, but is not really helpful in real-life scenarios.

In order to support the actual Envoy locality-aware LB feature, we have to extend 
`localityAwareness` configuration with something like:

```yaml
type: MeshLoadBalancing
name: extra-us-east
spec:
  targetRef:
    kind: MeshSubset
    tags:
      kuma.io/zone: us-east
  to:
    - targetRef:
        kind: MeshService
        name: payments
      default:
        localityAwareness: # or 'appendLocalityAwareness'
          matchers:
            - tagEqual:
                name: kuma.io/zone
                value: us-west
              weight: 1
            - tagEqual:
                name: kuma.io/zone
                value: eu-north
              weight: 10
```

This feature will be fully covered in a separate MADR.