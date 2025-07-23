# Internal listeners naming convention
* status: **superseded by [MADR-076](./076-naming-system-envoy-resources.md)**

## Context and Problem statement
At the moment we don't clearly distinguish between listeners used for service to service communication
and internal listeners, like Prometheus listeners.

## Decision Outcome

We will add prefix `_` to internal listeners.

## Solution
At the moments internal listeners names start with `kuma` prefix, problem is that service can also have this prefix.
To be able to easily distinguish between service listeners and internal listeners, we can
add `_` prefix to listener name.

I've looked into usage of these names in metrics, dashboards and other places and didn't find anything we 
should worry about. Also this convention is already used in MeshMetric and we didn't find any issues.

When migrating resources names remember to add info on how to migrate in UPGRADE.md.
