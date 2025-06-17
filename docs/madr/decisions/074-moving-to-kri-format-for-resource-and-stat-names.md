# Moving to KRI format for resource and stat names

* Status: accepted

## Context and problem statement

As described in the [Resource Identifier MADR](070-resource-identifier.md), we chose to standardize resource naming using the KRI (Kuma Resource Identifier) format. This change improves consistency across the system and allows for stronger cross-resource references, making data easier to present and understand. However, moving to KRI will break existing setups that rely on current Envoy stat and resource names, such as dashboards and GUI features. A smooth migration path is needed to let users adopt KRI without disruption.

### Affected systems and users

Two main groups are impacted:

* Internal systems: Kuma GUI
* External users: Dashboards, Prometheus

#### Kuma GUI

The Dataplane view displays inbound and outbound endpoints using data from Envoy stats. It matches these stats to configuration by parsing stat names, such as cluster or listener names. Changes to stat naming or resource identifiers will break this matching logic. We must coordinate changes with the GUI to support both the old and new formats during the transition.

## Scope

This decision applies to environments using the new service discovery model based on:

* `MeshService`
* `MeshExternalService`
* `MeshMultiZoneService`

These and the legacy `kuma.io/service` tag represent two different modes of describing the same types of services in the mesh. Since `kuma.io/service` is deprecated and will be removed in the future, this migration only covers naming changes for resources generated from the new model. Updating or supporting naming for the legacy mode is out of scope.

This document also does not cover renaming of internal Envoy resources that do not correspond directly to real Kuma resources such as `MeshHTTPRoute`, `Dataplane`, or `MeshExternalService`. These include:

* Secrets
* Internal listeners and clusters
* Default routes (when no `MeshHTTPRoute` or `MeshTCPRoute` is defined)
* `MeshPassthrough` resources

Renaming of these internal resources will be handled in a separate MADR.

Support for the built-in gateway is also out of scope. It may be addressed separately if we decide to include it.

## Decision outcome

Introduce a feature flag to enable the new KRI-based resource naming. This flag can be set at the control plane level, causing all data plane proxies to use the KRI format in Envoy resource and stat names. Alternatively, it can be enabled per data plane proxy using an environment variable in Universal mode or an annotation in Kubernetes mode.

## Implementation

### Feature flag

A new feature flag will be introduced to enable the KRI-based stat naming. This flag will be opt-in at first, with a clear deprecation notice: the new format will become the default in the future, and eventually the flag will be removed. The goal is to avoid breaking existing setups while giving users an early path to adopt the new behavior.

The feature flag will be sent by the data plane proxy to the control plane via xDS metadata, using the following constant:

```go
const FeatureKRIStats string = "feature-kri-stats"
```

To enable this feature globally, users can set a control plane runtime option:

```go
KRIStatsEnabled bool `json:"kriStatsEnabled" envconfig:"KUMA_MESH_SERVICE_KRI_STATS_ENABLED"`
```

This will cause all data plane proxies to include the feature flag automatically.

To enable the feature for individual data planes:

| Mode       | How to enable                                                        |
|------------|----------------------------------------------------------------------|
| Universal  | Set env var: `KUMA_DATAPLANE_RUNTIME_METRICS_KRI_STATS_ENABLED=true` |
| Kubernetes | Add annotation: `kuma.io/kri-stats-enabled: "true"`                  |

In Kubernetes, the sidecar injector will convert the annotation into the corresponding environment variable. `kuma-dp` will then detect this variable and include the feature flag in the xDS metadata sent to the control plane.

## Test scenarios required for completion

The following scenarios must be verified to consider the work complete. Each case ensures the correct generation and usage of KRI-based resource and stat names across various deployment modes.

### Single-zone deployments

* `MeshService` is targeted directly by a dataplane in the same zone
* `MeshExternalService` is targeted via ZoneEgress in the same zone

### Multi-zone with ZoneEgress and ZoneIngress (Global + 2 zones)

* Dataplane in zone 1 targets a `MeshService` in zone 2
* Dataplane in zone 1 targets a `MeshExternalService` in zone 1
* Dataplane in zone 1 targets a `MeshMultiZoneService` in zone 2
* Dataplane in zone 1 targets a `MeshMultiZoneService` in zone 2 with locality awareness disabled via `MeshLoadBalancingStrategy`

### Multi-zone with only ZoneIngress (Global + 2 zones)

* Dataplane in zone 1 targets a `MeshService` in zone 2
* Dataplane in zone 1 targets a `MeshMultiZoneService` in zone 1
* Dataplane in zone 2 targets a `MeshMultiZoneService` in zone 2 with locality awareness disabled via `MeshLoadBalancingStrategy`

### Multi-zone with ZoneIngress in both zones and ZoneEgress in one zone (Global + 2 zones)

* Dataplane in zone 1 targets a `MeshService` in zone 2
* Dataplane in zone 2 targets a `MeshService` in zone 1
* Dataplanes in both zones target a `MeshMultiZoneService` in their respective zones
* Dataplanes in both zones target a `MeshMultiZoneService` with locality awareness disabled via `MeshLoadBalancingStrategy`
