# Migrating to KRI-based Envoy resource and stat naming

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

Introduce a feature flag to enable the new KRI-based resource naming. This flag can be enabled per data plane proxy using an environment variable in Universal mode or an annotation in Kubernetes mode. In Kubernetes, the control plane can be configured to automatically inject the annotation into all workloads during sidecar injection, enabling the feature by default for all proxies in the zone.

## Implementation

### Feature flag and propagation

A new feature flag will control the use of KRI-based stat and resource naming. It will be optional at first, with a clear deprecation plan: KRI naming will become the default in the future, and the flag will eventually be removed. This gives users time to migrate without breaking existing setups.

The flag will be passed from the data plane proxy to the control plane via xDS metadata using:

```go
const FeatureKRINaming string = "feature-kri-naming"
```

#### Per-proxy opt-in

Users can enable the feature for individual data plane proxies:

| Mode       | How to enable                                                 |
|------------|---------------------------------------------------------------|
| Universal  | Set env var: `KUMA_DATAPLANE_RUNTIME_KRI_NAMING_ENABLED=true` |
| Kubernetes | Add annotation: `kuma.io/kri-naming: "enabled"`               |

In Kubernetes, the sidecar injector will translate the annotation into the corresponding environment variable. `kuma-dp` will detect the variable and include the feature flag in xDS metadata.

#### Zone-wide opt-in for data plane proxies in Kubernetes

To enable KRI naming for all injected data plane proxies in a Kubernetes zone, users can configure the control plane to automatically set the annotation during sidecar injection:

```go
KRINamingEnabled bool `json:"kriNamingEnabled" envconfig:"KUMA_RUNTIME_KUBERNETES_INJECTOR_KRI_NAMING_ENABLED"`
```

#### ZoneIngress and ZoneEgress proxies

`ZoneIngress` and `ZoneEgress` proxies must also include the KRI naming feature flag to ensure consistent naming across the mesh. In both Universal and Kubernetes modes, this is done by setting the following environment variable in their deployments:

```env
KUMA_DATAPLANE_RUNTIME_KRI_NAMING_ENABLED=true
```

#### Helper setting for Kubernetes installations

To simplify configuration in Kubernetes environments, a new installation setting will be introduced:

```yaml
experimental.kriNaming.enabled
```

When set to `true`, this will:

* Set the `KUMA_DATAPLANE_RUNTIME_KRI_NAMING_ENABLED="true"` environment variable in all `ZoneIngress` and `ZoneEgress` deployments
* Set the `KUMA_RUNTIME_KUBERNETES_INJECTOR_KRI_NAMING_ENABLED="true"` environment variable in the control plane deployment

Once KRI naming becomes the default, this setting will also default to `true`. Users will still be able to disable it by setting it to `false`. Eventually, this setting will be removed when disabling the feature is no longer supported.

### Impact on MeshProxyPatch policies

Enabling the KRI naming feature can immediately break existing `MeshProxyPatch` policies if they were written using legacy resource names and are not updated ahead of time. Since KRI changes the structure and names of xDS resources, any patch relying on previous names may stop applying correctly.

Unfortunately, there is currently no reliable way to detect and warn users about this. Several approaches were considered:

* Log or show a warning if a `Dataplane` is matched by a `MeshProxyPatch`, but none of the patch modifiers apply to any xDS resources

  *Problem:* It's hard to determine whether complex patches like JSON Patch match anything. There's no existing way to surface this to the GUI, and logging such cases could be noisy and lead to false positives. Even in debug logs, it could quickly become more annoying than helpful.

* Search for the `kri_` prefix in patch contents as a hint that the policy is KRI-aware

  *Problem:* This is unreliable, and all the concerns above still apply.

* Show a general warning in the dataplane or policy view when KRI is enabled and any `MeshProxyPatch` applies

  *Problem:* This might be useful the first time but would quickly become noise once the user verifies their configuration.

Because none of these options are satisfactory, the chosen path is to document this behavior clearly:

* Include a warning in the upgrade notes to alert users that existing `MeshProxyPatch` policies must be reviewed and updated before enabling KRI naming
* Add a strong warning to the documentation for `MeshProxyPatch` policies

### Updating ZoneIgress and ZoneEgress overview resources with feature flags

To support the Kuma GUI in adapting to KRI-based naming, we need to expose feature flag information in `ZoneIngressOverview` and `ZoneEgressOverview` resources, similar to how it's already done for `DataplaneOverview`. These overview resources are available via the control plane API and provide a summary of runtime state and metadata.

We will update both `ZoneIngressOverview` and `ZoneEgressOverview` to include metadata with active feature flags. This will allow the GUI to detect whether KRI naming is enabled for each proxy and adjust its behavior accordingly.

## Migration

### Kuma GUI

The Kuma GUI relies on Envoy stat names to display inbound and outbound endpoint details for data plane, `ZoneIngress`, and `ZoneEgress` proxies. These names are used to associate metrics with specific resources and visualize traffic paths in the GUI.

To support the KRI naming format, we need to update the GUI parsers that extract resource information from stat names such as listeners, clusters, and endpoints. These parsers must recognize and handle the new KRI format defined in the [Resource Identifier MADR](https://github.com/kumahq/kuma/blob/d19b78a4556962f4d9d3cc5921c7bdc73dc93d26/docs/madr/decisions/070-resource-identifier.md?plain=1#L328):

```
kri_<resource-type>_<mesh>_<zone>_<namespace>_<resource-name>_<section-name>
```

The updated parsers will be activated when the `feature-kri-naming` flag is present in the metadata of the `DataplaneOverview`, `ZoneIngressOverview`, and `ZoneEgressOverview` resources.

The GUI does not rely on the names of xDS resources themselves. It only uses stat names, so no additional changes are required to support KRI naming.

## Test scenarios required for completion

The following scenarios must be verified to consider the work complete. Each case ensures the correct generation and usage of KRI-based resource and stat names across various deployment modes.

### Single-zone deployments

* `MeshService` is targeted directly by a `Dataplane` in the same zone
* `MeshExternalService` is targeted by a `Dataplane` via the `ZoneEgress` in the same zone

### Multi-zone with ZoneEgress and ZoneIngress (Global + 2 zones)

* `Dataplane` in zone 1 targets a `MeshService` in zone 2
* `Dataplane` in zone 1 targets a `MeshExternalService` in zone 1
* `Dataplane` in zone 1 targets a `MeshMultiZoneService` in zone 2
* `Dataplane` in zone 1 targets a `MeshMultiZoneService` in zone 2 with locality awareness disabled via `MeshLoadBalancingStrategy`

### Multi-zone with only ZoneIngress (Global + 2 zones)

* `Dataplane` in zone 1 targets a `MeshService` in zone 2
* `Dataplane` in zone 1 targets a `MeshMultiZoneService` in zone 1
* `Dataplane` in zone 2 targets a `MeshMultiZoneService` in zone 2 with locality awareness disabled via `MeshLoadBalancingStrategy`

### Multi-zone with ZoneIngress in both zones and ZoneEgress in one zone (Global + 2 zones)

* `Dataplane` in zone 1 targets a `MeshService` in zone 2
* `Dataplane` in zone 2 targets a `MeshService` in zone 1
* `Dataplanes` in both zones target a `MeshMultiZoneService` in their respective zones
* `Dataplanes` in both zones target a `MeshMultiZoneService` with locality awareness disabled via `MeshLoadBalancingStrategy`
