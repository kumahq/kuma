# Mesh Scoped Zone Ingress and Zone Egress

* Status: accepted

Technical Story: <!-- link to the github issue -->

## Context and Problem Statement

Zone Ingress and Zone Egress are currently global-scoped resources that exist outside of any mesh context. This creates problems for:

1. **Identity** - ZoneIngress/ZoneEgress cannot have a proper mesh-scoped identity, which breaks the intended trust and identity model
2. **Policies** - Cannot apply mesh-scoped policies to ZoneIngress/ZoneEgress since they don't belong to a mesh

We need to rework the API to make Zone Ingress and Zone Egress mesh-scoped by leveraging the existing Dataplane resource with additional fields.

## Design

### New Schema for Zone Ingress and Zone Egress

It doesn't seem possible to keep using ZoneIngress/ZoneEgress resources and change their scope from `global` to `mesh`.
Main reason for that is migration. Migration requires CP to work with old and new resources at the same time,
however resource in Kuma can't have 2 scopes:

```go
var ZoneIngressResourceTypeDescriptor = model.ResourceTypeDescriptor{
	Name:                ZoneIngressType,
	Scope:               model.ScopeGlobal, // or model.ScopeMesh
}
```

Introducing new resource kinds like `MeshZoneIngress` and `MeshZoneEgress` is possible, but feels excessive.
The best option is to extend already existing kind `Dataplane` and add fields to hold zone proxy specific information.

#### Option 1 (selected): dedicated `zoneIngress` and `zoneEgress` sections

```yaml
type: Dataplane
mesh: default
name: zone-ingress-1
spec:
  networking:
    zoneIngress:
      address: 10.0.0.1 # required, address listener binds to 
      port: 10001 # required, port listener binds to
      advertisedAddress: 192.168.1.100  # required, address visible to other zones
      advertisedPort: 30001  # required, port visible to other zones
      name: zi-port # optional, user should be able to set name since `port` can be the same when `address` are different
    zoneEgress:
      address: 10.0.0.2 # required
      port: 10002 # required
      name: ze-port # optional
```

##### Open Questions

1. Should `networking.zoneIngress` and `networking.zoneEgress` be an array?
2. There is also `networking.advertisedAddress` field (but no `networking.advertisedPort`). 
It was a very niche Universal-only field [contributed by community](https://github.com/kumahq/kuma/pull/2116) that I think we should remove in v3. 
I think `networking.zoneIngress.advertisedAddress` should be `required` so it never falls back to `networking.advertisedAddress`.

##### Pros and Cons

* Good, because `zoneIngress` and `zoneEgress` have different schema 
* Bad, because `name` looks awkward, it needs to be unique across `networking.zoneIngress` and `networking.zoneEgress`

#### Option 2: `listeners` array

```
type: Dataplane
mesh: default
name: zone-ingress-1
spec:
  networking:
    listeners:
      - type: ZoneIngress # required
        address: 10.0.0.1 # required
        port: 10001 # required
        advertisedAddress: 192.168.1.100 # optional, because when `type` is ZoneEgress we don't need it
        advertisedPort: 30001 # optional
        name: zi-port
      - type: ZoneEgress
        address: 10.0.0.2
        port: 10002
        name: ze-port
```

##### Pros and Cons

* Good, because `name` doesn't look awkward, it's obvious it has to be unique within the `listeners` array
* Good, because in the future if we want clean DPP spec we can deprecate `networking.inbound` and add listeners with `type: Inbound`
* Bad, because items in `listeners` need to have the same schema so we'd have to validate `advertisedAddress` based on `type`

### Policy Targeting for Zone Ingress and Zone Egress

In Kuma, it is fine if `spec.targetRef` is too broad.
The policy plugin determines whether the policy actually applies to a matched DPP and simply ignores it if it does not.

So it's okay to create a policy:

```yaml
spec:
  targetRef:
    kind: Dataplane # targets all DPPs in the mesh
  to:
    - targetRef:
        kind: MeshHTTPRoute
        name: route-1
      default: $conf # will be ignored on DPPs that don't have `route-1`
```

We should follow the same approach with policies targeting zone ingress and zone egress.

```
spec:
  targetRef:
    kind: Dataplane
   rules:
     - matches: [] # select specific FilterChain
       default: # ignored if DPP doesn't have the requested FilterChain
         deny:
           - spiffeId:
               type: Prefix
               value: "spiffe://default/"
```

Field `targetRef.sectionName` can be used to select only zone egress or zone ingress listener. 

#### Labels on Zone Ingress/Egress Dataplanes

TBD

### Syncing Zone Ingress Addresses via MeshService

In multizone, other zones need to know the addresses of zone ingresses to route traffic.
Since zone ingress will be represented by the Dataplane resource and Dataplanes are not synced from one zone to another,
we need to find another way to share `advertisedAddress` and `advertisedPort` with other zones.

#### Proposed Approach: MeshService as the Carrier

TBD

## Deprecation of ZoneIngress and ZoneEgress Resources

TBD

## Implications for Kong Mesh

None

