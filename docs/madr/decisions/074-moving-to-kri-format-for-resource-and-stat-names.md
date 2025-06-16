# Moving to KRI format for resource and stat names

* Status: accepted

## Context and Problem Statement

As stated in the [MADR](070-resource-identifier.md) about resource identifiers, we are not consistent in how we set names for resources. By improving the naming, we can also enable much more powerful cross-resource referencing, which would allow better presentation of data.
We would like to make this change, but it would break all places where users rely on Envoy metrics (e.g., dashboards) and might also impact the GUI.
Therefore, we need a path forward that allows users to transition safely to the new model.

## Considered Options

* Option 1 - Rename resources based on KRI without introducing a breaking change, and introduce a feature flag to enable the new resource naming.

## Decision Outcome

* Option 1 - ename resources based on KRI without introducing a breaking change, and introduce a feature flag to enable the new resource naming.

## Pros and Cons of the Options

### Option 1 - ename resources based on KRI without introducing a breaking change, and introduce a feature flag to enable the new resource naming.

We currently have two modes to consider:

* `kuma.io/service`
* `MeshService`, `MeshExternalService`, and `MeshMultiZoneService`

We will not change anything for `kuma.io/service`, since it is not a real resource.

During this task, we should split the work into two main parts:
* Rename places that are not exposed to or used by clients, so that we can adopt the new model by default internally.
* Introduce a feature flag to switch all resource names and metrics to the KRI-based naming for users.

#### Rename places that are not exposed to or used by clients, so that we can adopt the new model by default internally.

We can distinguish two types of clients:

* Internal — GUI
* External — Dashboards, Prometheus

On the Dataplane view, we show the outbounds and inbounds.
We retrieve this information from the stats and later reference the configuration based on the name of the stats — either the cluster name or the listener name.
Changing either the resource name or the statistic name might impact this functionality.
We need to coordinate changes with the UI team to ensure we support both situations.

What can we do?

>[!NOTE]
> Format of resource names is defined in [MADR](070-resource-identifier.md)

* `Dataplane`
Cluster:
```
  name: kri_msvc_mesh-1_us-east-2_kuma-demo_backend_httpport
  alt_stats_name: mesh-1_backend_kuma-demo_us-east2-_msvc_9090
```
Listener:
```
  name: kri_msvc_mesh-1_us-east-2_kuma-demo_backend_httpport
  stat_prefix: mesh-1_backend_kuma-demo_us-east2-_msvc_9090
```
The GUI should be updated to support looking up resources not just by name, but also by `alt_stats_name` (for Clusters) or `stat_prefix` (for Listeners).
This way, we avoid breaking the stats view for metrics even when naming changes.

* `ZoneIngress` and `ZoneEgress`
In these cases, we should be able to rename the fields without any impact, since they are not exposed to external consumers.
Cluster:
```
  name: kri_msvc_mesh-1_us-east-2_kuma-demo_backend_httpport
```
Listener:
```
  name: kri_msvc_mesh-1_us-east-2_kuma-demo_backend_httpport
```

#### Introduce a feature flag to switch all resource names and metrics to the KRI-based naming for users.

In `2.11.x`, we want to introduce a flag that allows switching to the new format, along with a note that starting from `2.13.x`, this will become the default behavior.
We do not want to break existing user setups, but we want to give users a heads-up that this change is coming, and that switching earlier might make the transition less painful.

```
KRIStatsEnabled bool `json:"kriStatsEnabled" envconfig:"KUMA_MESH_SERVICE_KRI_STATS_ENABLED"
```

By default, the flag will have a `false` value, but in `2.13.x` we will change the default to `true`.
In `2.15.x`, we will completely remove the flag — the new behavior will be the only supported mode.

When the variable is set to `true`, metrics will use only the new format.
This means we will no longer set `alt_stats_name` or `stat_prefix`; instead, only the resource name (for Cluster, Listener, Routes and Endpoint) will be set and used in both metrics and configuration.

We will also introduce an option to migrate a single dataplane to the new model:

| Universal | Kubernetes |
| --- | --- |
| Set environment variable: `KUMA_DATAPLANE_RUNTIME_METRICS_KRI_STATS_ENABLED=true` | Set annotation: `kuma.io/kri-stats-enabled: "true"` |

On Kubernetes, we will translate the annotation into the environment variable `KUMA_DATAPLANE_RUNTIME_METRICS_KRI_STATS_ENABLED` during sidecar injection.
Based on this environment variable, `kuma-dp` will enable the feature and propagate it via metadata:

```golang 
const FeatureKRIStats string = "feature-kri-stats"`
```

After releasing `2.11.x`, we should start adjusting dashboards and metric references so that by the time the new behavior becomes default in `2.13.x`, users will not encounter issues with observability.

#### Pros and Cons of the Option 1

**Pros:**
* Enables gradual migration of stats without breaking existing dashboards.
* Does not change the `kuma.io/service` behavior.
* Lays the foundation for building better tooling around metrics and observability.

**Cons:**
* Might make the code more complicated.
* Migration might be troublesome for users who have custom dashboards.
