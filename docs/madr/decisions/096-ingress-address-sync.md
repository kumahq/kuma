# Syncing Zone Ingress Address Across Zones

* Status: accepted

Technical Story: https://github.com/Kong/kong-mesh/issues/9151

## Context and Problem Statement

In multizone, other zones need to know the addresses of zone ingresses to route traffic.
Since zone ingress will be represented by the Dataplane resource and Dataplanes are not synced from one zone to another,
we need to find another way to share `advertisedAddress` and `advertisedPort` with other zones.

## Decision Drivers

- reduce duplication of `advertisedAddress`, changing the public zone address should not trigger a mesh wide reconciliation storm

## Design

### Decision

Chosen option: Option 1.

The downside of introducing a new resource is mitigated by the fact that creation of new resource type is already automated in Kuma.

### Option 1: Create new MeshZoneAddress resource

Resource descriptor
```golang
var MeshZoneAddressResourceTypeDescriptor = model.ResourceTypeDescriptor{
	Name:                         MeshZoneAddressType,
	Resource:                     NewMeshZoneAddressResource(),
	ResourceList:                 &MeshZoneAddressResourceList{},
	Scope:                        model.ScopeMesh,
	KDSFlags:                     model.ZoneToGlobalFlag | model.SyncedAcrossZonesFlag,
	WsPath:                       "meshzoneaddresses",
	KumactlArg:                   "meshzoneaddress",
	KumactlListArg:               "meshzoneaddresses",
	AllowToInspect:               false,
	IsPolicy:                     false,
	IsExperimental:               false,
	SingularDisplayName:          "Mesh Zone Address",
	PluralDisplayName:            "Mesh Zone Addresses",
	IsPluginOriginated:           true,
	IsTargetRefBased:             false,
	HasToTargetRef:               false,
	HasFromTargetRef:             false,
	HasRulesTargetRef:            false,
	HasStatus:                    false,
	AllowedOnSystemNamespaceOnly: false,
	IsReferenceableInTo:          false,
	ShortName:                    "mza",
	IsFromAsRules:                false,
}
```

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
  address: 192.168.0.1 # string field, can support either IPv4 or IPv6
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

## Address and port resolution

Today, the `ZoneIngress` resource derives its `advertisedAddress` and `advertisedPort` from the Service
using the following algorithm (see `pkg/plugins/runtime/k8s/controllers/ingress_converter.go`):

1. **Pod annotations (highest priority)**: If the Pod has both `kuma.io/ingress-public-address` and
   `kuma.io/ingress-public-port` annotations, use those values. Both must be present or neither.

2. **Service-based resolution (fallback)**: If annotations are not present, derive from the Service based on its type:
   - **LoadBalancer**: Use `status.loadBalancer.ingress[0]`. If both `ip` and `hostname` are present, `ip` takes precedence.
     If the LoadBalancer is not yet ready (no ingress entries), skip setting advertised coordinates.
   - **NodePort**: List all cluster Nodes and use the first Node's (`nodes.Items[0]`) address from `node.Status.Addresses`.
     Address type priority: `NodeExternalIP` > `NodeInternalIP`.
     Port is taken from `service.Spec.Ports[0].NodePort`.
   - **Other types (ClusterIP, etc.)**: No advertised coordinates are set. A log message suggests using LoadBalancer,
     NodePort, or annotation overrides.

3. **Port resolution**: For LoadBalancer, use `service.Spec.Ports[0].Port`. For NodePort, use `service.Spec.Ports[0].NodePort`.

Note that the current implementation always takes the first element (index 0) when multiple addresses or ports are available.

For `MeshZoneAddress`, we have two options for handling Services with multiple addresses or ports:

### Option 1: Take first address and port

Follow the existing `ZoneIngress` behavior: always use index 0 for both `status.loadBalancer.ingress` and `service.Spec.Ports`.
This is simple and consistent with existing behavior.

* Good, because it maintains consistency with existing ZoneIngress behavior
* Good, because it's simple to implement
* Bad, because silently ignoring additional addresses/ports may be confusing to users

### Option 2: Validate and reject multiple addresses/ports

The `MeshZoneAddress` controller validates the Service and emits an error (via status condition or event)
if the Service has more than one LoadBalancer ingress entry or more than one port.
Users must ensure their Service has exactly one address and one port.

* Good, because it forces explicit configuration and avoids ambiguity
* Bad, because it's more restrictive and may require users to restructure their Services
* Bad, because LoadBalancer services may temporarily have multiple ingress entries during cloud provider updates

### Decision

Chosen option: Option 1 (take first address and port).

The `MeshZoneAddress` controller derives address and port from the Service using the following algorithm:

1. **`spec.externalIPs[0]`** (highest priority): If the Service has `externalIPs` defined, use the first one.
   Port is taken from `service.Spec.Ports[0].Port`.

2. **LoadBalancer**: If Service type is LoadBalancer, use `status.loadBalancer.ingress[0]`.
   If both `ip` and `hostname` are present, `hostname` takes precedence (more stable, e.g., AWS ELB IPs can change).
   If the LoadBalancer is not yet ready (no ingress entries), do not generate `MeshZoneAddress` yet.
   Port is taken from `service.Spec.Ports[0].Port`.

3. **NodePort**: If Service type is NodePort, list all cluster Nodes and use the first Node's (`nodes.Items[0]`)
   address from `node.Status.Addresses`. Address type priority: `NodeExternalIP` > `NodeInternalIP`.
   Port is taken from `service.Spec.Ports[0].NodePort`.

4. **Other types (ClusterIP, etc.)**: Do not generate `MeshZoneAddress`. Emit a warning event on the Service
   suggesting to use LoadBalancer, NodePort, or set `externalIPs`.

If multiple addresses or ports exist, always use the first one (index 0).

Note: Key differences from the legacy `ZoneIngress` address resolution:
- **No Pod annotations**: `ZoneIngress` uses Pod annotations (`kuma.io/ingress-public-address` and
  `kuma.io/ingress-public-port`) for overrides. `MeshZoneAddress` uses the native Kubernetes `spec.externalIPs`
  field on the Service instead, since it watches Services directly.
  **Migration requirement**: Users currently relying on Pod annotation overrides must switch to setting
  `spec.externalIPs` on the Kubernetes Service when adopting `MeshZoneAddress`.
- **Hostname over IP**: `ZoneIngress` prefers `ip` over `hostname` for LoadBalancer ingress entries.
  `MeshZoneAddress` prefers `hostname` over `ip` for better stability (e.g., AWS ELB IPs can change).

## Migration

Mesh-scoped zone proxies are an opt-in feature in Kuma 2.14.
Once enabled, both the legacy `ZoneIngress` and the new mesh-scoped zone proxy resources are supported during a transition period to ensure a smooth migration.

When the CP detects both a legacy `ZoneIngress` and a new `MeshZoneAddress` for a given zone,
it prioritizes `MeshZoneAddress` and shifts all traffic to the new zone proxies.

### Rollback

Because `ZoneIngress` resources persist throughout the transition period, rollback is supported:
1. Delete the `MeshZoneAddress` resources (or disable the mesh-scoped zone proxy feature).
2. The CP stops finding `MeshZoneAddress` for the zone and reverts to propagating the legacy `ZoneIngress` addresses.

Note: this requires `ZoneIngress` to still be present. If the user has already removed `ZoneIngress` as part of a completed migration, it must be redeployed before rollback is possible.

## Implications for Kong Mesh

None

