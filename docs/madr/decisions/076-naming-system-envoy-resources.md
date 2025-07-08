# Standardized Naming for system xDS Resources

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/13266

Supersedes: https://github.com/kumahq/kuma/blob/master/docs/madr/decisions/036-internal-listeners.md

## Context and Problem Statement

The point of this document is to define a naming strategy of Envoy resources, existing exceptions and guidelines for future exceptions.

We use Envoy resource names to identify listeners, clusters, routes, virtual hosts and secrets in the xDS APIs.
These names appear in metrics, logs, and debug endpoints (config_dump), and they are often used for monitoring, filtering, and diagnostics.

We classify Envoy resources in two groups:
- System resources - these are internal to Kuma and users are unlikely to be interested in them except when debugging Kuma.
- User resources - the rest of the resources which are most commonly derived from a Kuma resource and will likely be monitored by users.

For example a cluster named `kuma:envoy:admin` is a system resource, while a cluster `kri_msvc_mesh-1_us-east-2_kuma-demo_backend_httpport` is a user resource.

From now on when we use "system resource" or "user resource" we're talking about Envoy resources, not Kuma resources.

Currently, there is no consistency in how these resources are named, see [additional context](#additional-context) for detailed list.

## Driving factors

### 1. All resources should have a name

Even though Envoy resource names are sometimes optional we should always set them.
This can improve debugging, readability and reusability.

### 2. Looking at a resource name I want to easily relate it to Kuma concepts and understand where it's coming from

This means that if a resource comes directly from a Kuma resource we should relate to this Kuma resource directly.
It should be easy when reading a `config_dump` to understand relationship between Envoy resources.

### 3. As a user I want to easily filter out system resources

So for example in Prometheus instead of

``` 
sum:envoy.cluster.upstream_rq.count{!envoy_cluster:kuma_readiness , !envoy_cluster:access_log_sink , !envoy_cluster:meshtrace_datadog}.as_count()
```

we can do:

``` 
sum:envoy.cluster.upstream_rq.count{!envoy_cluster:system_*}.as_count()
```

Another example would be to have an OTEL processor like this to drop the system related metrics:

```yaml
processors:
  filter/drop_kuma_system_resources:
    metrics:
      exclude:
        match_type: regexp
        metric_names: [".*"]  # apply to all metrics
        attributes:
          - key: envoy_cluster
            value: "^system_.*"  # matches if envoy_cluster starts with _
```

A side benefit of this is that profiles implementation is a lot simpler.

### 4. As a user I want the cardinality of these resource names to be low

High cardinality is problematic for metrics collection systems and dataplane memory.

Need to avoid:
- IPs / non constant bits
- Dataplane name or its KRI.

There is existing solution for `Dataplane` covered by [MADR-077](./077-migrating-to-consistent-and-well-defined-naming-for-non-system-envoy-resources-and-stats.md).

#### 5. Rename existing system resources in well-defined patterns

Here is a list of exceptions we will back-fill:
- Dynamic Config Fetcher listener and routes
- MeshMetric clusters and listeners
- Metrics Hijacker cluster
- MeshTrace clusters
- MeshAccessLog cluster
- MeshGlobalRateLimit cluster
- Dataplane probe listener and route
- Prometheus listener
- Kuma DNS listener
- Kuma readiness cluster

## Unsupported use cases

### As a user I want my MeshProxyPatch to continue working after the change

This is not possible.
In MeshProxyPatch you can use `name` matcher to match a resource name but if we change the name of the resource, it will not match anymore.
The feature will be behind a feature flag so that nothing breaks unexpectedly but migration of MeshProxyPatch resources will be required.
However, this will make `MeshProxyPatch` matches easier to write.

### As a Kuma developer I want to rename Envoy resources that are provided by xDS components outside Kuma

This is either not possible or very hard.
Envoy can connect to many xDS servers to get configuration from (example: SPIRE uses its own SDS server).

## Design

### Resource naming

#### 1. System resources that can be tracked back to Kuma resource

These should be exceptional cases when we consider such resource as system and not user resource.
Such exceptions should be defined in a MADR and must use `system_<kri>` naming scheme.

For example the cluster for `MeshGlobalRateLimit` service is a system resource because it is not part of the regular service-to-service traffic path,
but it is still related to a Kuma resource (the `MeshGlobalRateLimit` policy).
The cluster is currently called `meshglobalratelimit:service`, but it will be renamed to `system_kri_mgrl___kong-mesh-system_global-rate-limit-policy_`.

There are other cases where a resource is the result of merging multiple resources together (MeshTrafficPermission RBAC filters, MeshPassthrough, etc.).
This is out of scope and [requires a separate MADR](https://github.com/kumahq/kuma/issues/13886).

Refer to [currently existing system resources and their new naming](#currently-existing-system-resources-and-their-new-naming) for the full list.

#### 2. System resources that are not related to Kuma resource

When adding such resource the name should be explicit enough to help anyone understand where this is coming from.
The name should be namespaced so that it groups similar resources together and avoids collisions with other system resources.

For example dynamic config creates one listener and one route per object type.
We will namespace it as `system_dynamicconfig` and the DNS route will be `system_dynamicconfig_dns` and MeshMetric route will be `system_dynamicconfig_meshmetric`.

If the name is coming from a reusable component like it is now the namespacing is the responsibility of the reusable component, not the user of it.

It's important to balance the specificity of the name with the risk of high cardinality.
For example the socket path of the internal listeners are configurable,
but we don't want to include it in the name because could lead to high cardinality and prevents reusability of dashboards.

#### 3. User resources

Names for these resources are defined in [MADR-077](./077-migrating-to-consistent-and-well-defined-naming-for-non-system-envoy-resources-and-stats.md).

### Useful references for naming

- Resource contextual to the Dataplane are defined in [MADR-077](./077-migrating-to-consistent-and-well-defined-naming-for-non-system-envoy-resources-and-stats.md).
- Secrets adhere to the same rules (1-3).

### Regex to match system names

For `system_<kri>` resources we will use the same regex as used for KRI but prefixed with `system_`.

For other names we will use `^system_([a-z0-9-]*_?)+$`, here is an example of the usage: https://regex101.com/r/Ic1bk5/2.

### Currently existing system resources and their new naming

| Resource type | Current Name                                                 | New Name                        | Example                                                                                     |
|---------------|--------------------------------------------------------------|---------------------------------|---------------------------------------------------------------------------------------------|
| Cluster       | kuma:envoy:admin                                             | system_envoy_admin              |                                                                                             | 
| Listener      | _kuma:dynamicconfig                                          | system_dynamicconfig            |                                                                                             | 
| Route         | _kuma:dynamicconfig:dns                                      | system_dynamicconfig_dns        |                                                                                             | 
| Route         | _kuma:dynamicconfig:meshmetric                               | system_dynamicconfig_meshmetric |                                                                                             | 
| Listener      | _kuma:metrics:prometheus:backend-default                     | system_<kri>                    | system_kri_mm_mesh-1_us-east-2_kuma-demo_default_prometheus_backend-default                 |
| Listener      | _kuma:metrics:prometheus:<clientId>                          | system_<kri>                    | system_kri_mm_mesh-1_us-east-2_kuma-demo_default_prometheus_prometheus-secondary            |
| Listener      | _kuma:metrics:opentelemetry:<sanitizedBackendUrl>            | system_<kri>                    | system_kri_mm_mesh-1_us-east-2_kuma-demo_default_otel_otel-collector.observability.svc_4317 |
| Cluster       | _kuma:metrics:opentelemetry:<sanitizedBackendUrl>            | system_<kri>                    | system_kri_mm_mesh-1_us-east-2_kuma-demo_default_otel_otel-collector.observability.svc_4317 |
| Cluster       | _kuma:metrics:hijacker                                       | system_metrics_hijacker         |                                                                                             | 
| Listener      | kuma:metrics:prometheus                                      | system_metrics_prometheus       |                                                                                             |
| VirtualHost   | kuma:metrics:prometheus                                      | system_metrics_prometheus       |                                                                                             |
| Cluster       | tracing:<name>                                               | system_tracing_<name>           |                                                                                             | 
| Cluster       | meshtrace:<type> (zipkin, datadog, otel)                     | system_<kri>                    | system_kri_mt_mesh-1_us-east-2_kuma-demo_default_                                           | 
| Cluster       | meshaccesslog:opentelemetry:<index> (0, 1, 2, ...)           | system_<kri>                    | system_kri_mal_mesh-1_us-east-2_kuma-demo_multiple-backends_                                | 
| Cluster       | meshglobalratelimit:service                                  | system_<kri>                    | system_kri_mgrl___kong-mesh-system_mesh-rate-limit_                                         |
| Cluster       | kuma:readiness                                               | system_probe_readiness          |                                                                                             |
| Route         | 9Zuf5Tg79OuZcQITwBbQykxAk2u4fRKrwYn3//AL4Yo= (default route) | system_route_default            |                                                                                             |
| Secret        | mesh_ca:secret:all                                           | system_mtls_ca_all_meshes       |                                                                                             |
| Secret        | mesh_ca:secret:<mesh>                                        | system_mtls_ca_<mesh>           |                                                                                             |
| Secret        | identity_cert:secret:<mesh>                                  | system_mtls_identity_<mesh>     |                                                                                             |
| Cluster       | ads_cluster                                                  | system_ads                      |                                                                                             |
| Cluster       | plugins:bootstrap:k8s:hooks:apiServerBypass                  | system_kube_api_server_bypass   |                                                                                             |
| Listener      | plugins:bootstrap:k8s:hooks:apiServerBypass                  | system_kube_api_server_bypass   |                                                                                             |
| Listener      | kuma:dns                                                     | system_dns_builtin              |                                                                                             |

Deprecated features whose resource names will remain unchanged:

- [Virtual Probes](https://kuma.io/docs/2.11.x/policies/service-health-probes/#virtual-probes):
  - `probe:listener`
  - `probe:route_configuration`
  - `probe`

### Enforcement

#### Builders

To enforce these rules we will add checks in resource builders like [ClusterBuilder](https://github.com/kumahq/kuma/blob/dedaba5b9de1bd134dce813ae49b3475d5d24e6b/pkg/xds/envoy/clusters/cluster_builder.go#L80).
If anyone tries to add a new resource that doesn't conform to any type it will fail tests.

Pros:
- Simple to implement

Cons:
- Can be easily skipped by not using a builder

#### Inspecting golden files

We could have a target that goes over golden files and checks the names of all resources.

Pros:
- Can catch things that are not created by a builder

Cons:
- Can be skipped by a code path not using golden files or not tested

#### Custom linter

A linter that would use `go/ssa` and `callgraph` that would figure out if an Envoy resource with a name that doesn't conform to the regex is created.

Pros:
- Should catch all the cases

Cons:
- Might be really hard to implement

#### Chosen option

We're choosing "Inspecting golden files" as it's the most feasible option that can be implemented quickly and will catch most of the cases.

## Rejected alternatives

- [Using `:`](https://github.com/kumahq/kuma/issues/2363)

## Implications for Kong Mesh

The changes introduced by this document impact the Envoy resources generated by the `MeshGlobalRateLimit` policy.
This policy currently adds a cluster with a static name `meshglobalratelimit:service`,
which is then referenced in route-level [rate limit configuration](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-field-config-route-v3-routeaction-rate-limits). As a result of these changes, the cluster name will be updated to follow the `system_<kri>` format, using the valid `kri` of the `MeshGlobalRateLimit` policy that is the source of this cluster (i.e. `system_kri_mgrl___kong-mesh-system_global-rate-limit-policy_`)

Depending on the changes of `dynamicconfig` `MeshOPA` resource creation might change.

Other Kong Mesh-specific features or policies do not rely on or modify Envoy resource names or stat prefixes that fall within the scope of this document.

## Implications for GUI

GUI will need to reflect these new rules because we will ship them at the same time as [MADR-077](./077-migrating-to-consistent-and-well-defined-naming-for-non-system-envoy-resources-and-stats.md).

## Additional context

### Kuma system resource names

Listeners:
- inbound
  - kuma:envoy:admin
  - _kuma:dynamicconfig
  - _kuma:metrics:prometheus:default-backend
  - probe:listener
  - kuma:metrics:prometheus
  - plugins:bootstrap:k8s:hooks:apiServerBypass
  - kuma:dns

Clusters:
- access_log_sink
- kuma:envoy:admin
- kuma:metrics:hijacker
- kuma:readiness
- meshtrace_[zipkin|datadog|otel]
- ads_cluster
- plugins:bootstrap:k8s:hooks:apiServerBypass

Virtual Hosts:
- _kuma:dynamicconfig
- kuma:envoy:admin

Routes (not all of them even have a name):
- _kuma:dynamicconfig:dns
- _kuma:dynamicconfig:meshtrace
- 9Zuf5Tg79OuZcQITwBbQykxAk2u4fRKrwYn3//AL4Yo= (default route)

Secrets format:
- https://github.com/kumahq/kuma/blob/9ed757d3e5955ea57d7940badf4468057ed46663/pkg/xds/envoy/secrets.go#L26

Names in codebase:
- https://github.com/kumahq/kuma/blob/bdc95fb8b8a4da2388948041171d5b9ecf4345a5/pkg/xds/envoy/names/resource_names.go
- https://github.com/kumahq/kuma/blob/9ed757d3e5955ea57d7940badf4468057ed46663/pkg/xds/envoy/secrets.go#L13
- https://github.com/kumahq/kuma/blob/950ff3353f7e85670717557691b42208b17a6579/pkg/plugins/policies/core/xds/meshroute/listeners.go#L343
- https://github.com/kumahq/kuma/blob/26af860614cf0792c0bff004ac95e8f5115808fc/pkg/xds/envoy/tags/match.go#L37
- https://github.com/kumahq/kuma/blob/7bafa578aad6e528befcb6c96f025542fd1f6870/pkg/plugins/policies/meshtrace/plugin/xds/configurer.go#L264
- https://github.com/kumahq/kuma-gui/blob/f7f9da37c335ba14151bb4a3e546437b7eae94c7/packages/kuma-gui/src/app/connections/data/index.ts#L125-L137

#### Other meshes

##### Istio

Istio doesn't seem to be following any naming convention:
- clusters
  - some are in form of [inbound|outbound]|port (like `outbound|9080`)
  - PassthroughCluster / InboundPassthroughCluster
  - BlackHoleCluster
  - prometheus_stats
  - sds-grpc
  - xds-grpc
  - agent
- listeners
  - ip_port
  - virtualOutbound
  - virtualInbound

Istio related issues:
- https://github.com/istio/istio/issues/5311
- https://github.com/istio/istio/issues/31112#issuecomment-1124049572

