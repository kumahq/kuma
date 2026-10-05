# Envoy Zone Aware Routing Integration

**Status:** accepted

## Context and Problem Statement

Kuma’s existing **locality-aware load balancing** uses Envoy’s priority-based routing via `MeshLoadBalancingStrategy` to favor dataplanes sharing the same `kuma.io/zone` tag. However, it does **not** account for the relative number of instances per availability zone when distributing traffic. In environments where tasks are unevenly distributed across zones, this can overload individual instances and degrade performance.

**Example scenario:**
- **Source service:** 4 tasks in AZ-A, 1 in AZ-B
- **Destination service:** 1 task in AZ-A, 4 in AZ-B  
  Priority-based routing would still favor AZ-A—even though it has a single instance—resulting in uneven per-instance load. Envoy’s **zone aware routing** would balance evenly across all endpoints while favoring the local zone.

## Decision Drivers

- **True zone-awareness:** factor zone-level instance counts into routing decisions.
- **Minimal infrastructure changes:** no additional proxies or sidecars.
- **Leverage Envoy’s native implementation:** avoid reimplementing complex load-balancing logic.
- **Compatibility:** support both Kubernetes and universal modes.
- **Maintain existing `kuma.io/zone` tags:** avoid forcing users to reconfigure zones.

## Considered Options

1. **Enhance priority-based strategy**  
   Extend `MeshLoadBalancingStrategy` to calculate custom weights mimicking zone-aware behavior.  
   *Pros:* No external flags. *Cons:* Reinventing Envoy’s algorithm, maintenance burden.

2. **Expose Envoy’s `zone_aware_lb_config` directly**  
   Add a new `type: ZoneAware` mode in `MeshLoadBalancingStrategy` that sets `common_lb_config.zone_aware_lb_config`.  
   *Pros:* Minimal CP logic; leverages Envoy. *Cons:* Requires documentation and policy extension.

## Decision Outcome

Chosen option: **Expose Envoy’s `zone_aware_lb_config` directly** as a new `ZoneAware` mode in `MeshLoadBalancingStrategy`.

### Implementation Details

- Extend `MeshLoadBalancingStrategy.spec.default.localityAwareness.localZone` to accept:
    - `type: ZoneAware`
    - `minClusterSize`: mapped to Envoy’s `zone_aware_lb_config.min_cluster_size`
    - `zoneIdentifier`: tag key for zone (e.g., `topology/availability-zone`)
  - Kuma CP will inject into each dataplane’s bootstrap:
    ```yaml
    cluster_manager:
      local_cluster_name: {{ service_name }}
    node:
      locality:
        zone: {{ zone_tag }}
    ```
  - Kuma CP will inject into each relevant cluster:
    ```yaml
    common_lb_config:
      zone_aware_lb_config:
        min_cluster_size: {{ minClusterSize }}
    ```

### Example Configuration

```yaml
type: MeshLoadBalancingStrategy
name: zone-aware
mesh: default
spec:
  targetRef:
    kind: Dataplane
    labels:
      app: frontend
  to:
    - targetRef:
        kind: MeshService
        name: backend
        sectionName: http
      default:
        loadBalancer:
          type: ZoneAware
          zoneAware:
            # Minimum total endpoints before zone-aware logic activates
            minClusterSize: 1
            # Tag used on dataplanes to identify their zone
            zoneIdentifier: topology/availability-zone
```

## Consequences

- **Backward compatibility:** Existing weighted strategies remain unchanged.
- **Policy changes:** Users opt in via `type: ZoneAware`.
- **Documentation:** Need examples, guides, and upgrade notes.
