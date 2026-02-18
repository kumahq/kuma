# Syncing Zone Ingress Address Across Zones

* Status: accepted

Technical Story: https://github.com/Kong/kong-mesh/issues/9151

## Context and Problem Statement

In multizone, other zones need to know the addresses of zone ingresses to route traffic.
Since zone ingress will be represented by the Dataplane resource and Dataplanes are not synced from one zone to another,
we need to find another way to share `advertisedAddress` and `advertisedPort` with other zones.

## Decision Drivers

- reduce duplication of `advertisedAddress`,
changing the public zone address should not trigger a mesh wide reconciliation storm

## Design

### Decision

Chosen option: Option 1.

The downside of introducing a new resource is mitigated by the fact that creation of new resource type is already automated in Kuma.

### Option 1: Create new MeshZoneAddress resource

We can run a controller on the Zone CP that watches Service resources labeled `k8s.kuma.io/zone-proxy-type: ingress`
and creates a corresponding MeshZoneAddress resource.

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshZoneAddress
metadata:
  name: zone-ingress-service-name # name of the service
  namespace: kuma-system # namespace of the service
  labels:
    kuma.io/mesh: default
    kuma.io/zone: zone-name
spec:
  address: 192.168.0.1
  port: 10001
```

By generating `MeshZoneAddress` in a one-to-one relationship with the Kubernetes Service,
we guarantee there are no naming collisions within the Kubernetes cluster.

This means if a user has multiple Kubernetes Services labeled with `k8s.kuma.io/zone-proxy-type: ingress`,
the CP is going to generate multiple `MeshZoneAddress` for that zone.

#### Outbound configuration

CP fetches `MeshService` that corresponds to the outbound.
If `MeshService` is synced from another zone then CP has to fetch all `MeshZoneAddress` of that zone
to get the public addresses of zone's ingresses.

#### Pros and Cons

* Good, because the new resource has a clear purpose
* Bad, because need to create a new resource

### Option 2: New type of MeshService (with kuma.io/zone-proxy-type label)

We already have a resource generated one to one with a Kubernetes Service, it's `MeshService`.
For each Kubernetes Service with `kuma.io/zone-proxy-type: ingress` label we will generate a `MeshService` with the same `kuma.io/zone-proxy-type` label:

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshService
metadata:
  name: zone-ingress-service-name # name of the service
  namespace: kuma-system # namespace of the service
  labels:
    kuma.io/mesh: default
    kuma.io/zone: zone-name
    kuma.io/zone-proxy-type: ingress
spec:
  selector: {} # selects zone ingress deployment pods
  ports:
    - port: 10001
      targetPort: 6379
      appProtocol: tcp
  externalIPs: # this field already exist in Kubernetes Service but has to be added to MeshService
    - 192.168.0.1
```

We already sync `MeshService` resources across zones, so no extra work is required here.

#### Outbound configuration

CP fetches `MeshService` that corresponds to the outbound.
If `MeshService` is synced from another zone then CP has to fetch `MeshService` with `kuma.io/zone-proxy-type: ingress`.

#### Pros and Cons

* Good, because no need to add a new resource type
* Bad, because backward compatibility becomes tricky.
Existing code that lists all `MeshService` resources and generates artifacts from them (for example outbounds) may break.
A new `MeshService` labeled with `kuma.io/zone-proxy-type: ingress` does not behave like a regular outbound service,
but older logic will still treat it as one.

## Migration

TODO

## Implications for Kong Mesh

None



