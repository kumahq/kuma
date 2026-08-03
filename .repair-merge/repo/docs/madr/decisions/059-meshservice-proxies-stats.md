# MeshService proxies stats

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/10915

## Context and Problem Statement

[ServiceInsight](https://github.com/kumahq/kuma/blob/master/api/mesh/v1alpha1/service_insight.proto) has detailed information about the Service.
We need an equivalent of it in MeshService.

ServiceInsight includes the Status field
* Online - all dataplane proxies are connected to CP and have healthy inbounds
* Offline - all dataplane proxies are not connected or have no healthy inbounds
* Partially degraded - not all are "online" and not all are "offline"

This mixing of whether the DPP is connected to CP and whether is healthy is confusing.

When aggregating multiple `MeshService` into `MeshMultiZoneService` we need to know to which zones we can send the traffic to.
A zone can have `MeshService` but have no healthy endpoints, in this case we should try other zone instead.

## Considered Options

* Count in status
* State in spec

## Decision Outcome

Chosen option: "Count in status" and "state in spec".

## Pros and Cons of the Options

### Count in status and state in spec

```
kind: MeshService
status:
  dataplaneProxies:
    connected: 10
    healthy: 8
    total: 10
```

We now separate connected proxies to the zone control plane from proxies that have healthy inbounds.
This information is computed by the source zone and synced to global for visibility.

Fields:
* `dataplaneProxies.connected` - number of DPPs connected to the Zone CP
* `dataplaneProxies.healthy` - number of DPPs connected with all healthy inbounds selected by the MeshService

Fields that are in `ServiceInsights`, but won't be included in status of `MeshService`
* `status` - If we want to present in the GUI if `MeshService` is available to consume, we can use `spec.state` described in the next section
* `zones` - `MeshService` is bound to specific zone, so it makes no sense to include this
* `serviceType` - `MeshService` is always `internal`
* `addressPort` - `MeshService` already defines ports in `spec`
* `issuedBackends` - Permissive mTLS will be covered by a separate MADR

### State in spec

```
kind: MeshService
spec:
  state: Available | Unavailable
```

We need to pass an information for client's zone if there is any healthy endpoint available.
Because `status` of MeshService is synced from Zone CP to Global CP only (it's not synced cross zone),
we need to put this in `spec` just like we did with `spec.identities`.
Client's zone do not care how many endpoints are in the zone, it only cares if there is at least one endpoint available.
Not sending the exact number of endpoints cross zone helps with excessive KDS updates of remote MeshServices.
`state` gets two states:
* `Available` - at least one DPP selected by `MeshService` is healthy.
  While technically health status can be different for each port, assuming each port is exposed from different container,
  we want to avoid this complexity, and we want to use just one status. 
* `Unavailable` - no DPPs selected by `MeshService` are healthy
If `state` field is missing, we assume that it is `Unavailable`.

This is a lightweight equivalent of `ZoneIngress#availableServices`

`spec.state` and `status.dataplaneProxies` are computed using `StatusUpdater` component that already updates the resource with `spec.identities`.
