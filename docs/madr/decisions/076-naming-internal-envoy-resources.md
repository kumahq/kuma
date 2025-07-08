# Standardized Naming for internal xDS Resources

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/13266

Supersedes: https://github.com/kumahq/kuma/blob/master/docs/madr/decisions/036-internal-listeners.md

## Context and Problem Statement

We use resource names to identify listeners, clusters, routes, virtual hosts and secrets in the xDS APIs.
These names appear in metrics, logs, and debug endpoints (config_dump), and they are often used for monitoring, filtering, and diagnostics.

It's important to know the difference between system resources (the ones that are defined by Kuma itself, and they usually do not concern the end users) and user resources (the ones that are defined indirectly by the user and the users are interested in).
For example, clusters like `kuma:envoy:admin` or `_kuma:dynamicconfig` are system resources, while `kri_msvc_mesh-1_us-east-2_kuma-demo_backend_httpport` cluster is user resource.

Currently, there is inconsistency in how these resource names are formed, some of them have a `_kuma` prefix, some of them just contain `kuma` and some are free form.

Some internal resources do not map clearly to Kuma resources in the data store, making it difficult to relate metrics to Kuma abstractions.

There are also resources that seem to be in a bit of a gray area, like `meshtrace_datadog` which is a cluster created for `MeshTrace` policy.
It could be traced down to MeshTrace resource `kri_mtr_mesh-1__kuma-system_my-meshtrace_` and if a cluster operator has problems with tracing collection it might be something to look at.

To make it easier to distinguish between the two types we introduce the following definition:

- **User resource** - any Envoy resource that's a result of user defined Kuma resource

Exceptions:
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

- **System resource** - any resource that is not a user resource (there will be exceptions which will require a MADR)

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

## Use cases

### All resources should have a name even if it's optional

Currently, there are routes that do not have a name - let's add a name for them.

### Looking at a resource name I want to easily relate it to Kuma concepts and understand where it's coming from

### As a user I want to easily exclude all stats related to internal resources from a query

So instead 

``` 
sum:envoy.cluster.upstream_rq.count{!envoy_cluster:kuma_readiness , !envoy_cluster:access_log_sink , !envoy_cluster:meshtrace_datadog}.as_count()
```

we can do:

``` 
sum:envoy.cluster.upstream_rq.count{!envoy_cluster:system_*}.as_count()
```

You might be tempted to think that this use case is already covered by the `MeshMetric` profiles, 
and it partially is (the end result is the same) but this use case compliments it by making the profiles easier to implement and maintain.
It also allows users to do their own filtering more easily.

### As an operator I want to drop all stats related to internal resources

So I can use a processor like:

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

### As a Kuma developer I want to have a consistent naming scheme for all resources in Envoy

### As a Kuma developer I want to distinguish between system and user resources in Envoy

### As a Kuma developer I want to distinguish between system resources exposed and not exposed in metrics

### As a user I want the cardinality of these resource names to be low

Need to avoid:
- IPs / non constant bits
- Avoid Dataplane name or its KRI.

Note: That’s not a problem for system resources since we control their names.
Right now, this only affects user resources generated from the Dataplane, which is already covered by [Defining and migrating to consistent naming for non-system Envoy resources and stats](./077-migrating-to-consistent-and-well-defined-naming-for-non-system-envoy-resources-and-stats.md).

## Non use cases

### As a user I want my MeshProxyPatch to continue working after the change

This is not possible.
In MeshProxyPatch you can use `name` matcher to match a resource name but if we change the name of the resource, it will not match anymore.
The feature will be behind a feature flag so that nothing breaks unexpectedly but migration of MeshProxyPatch resources will be required.

### As a Kuma developer I want to rename resources that are coming from outside Kuma

This is either not possible or very hard.
Envoy can connect to many xDS servers to get configuration from.
If we integrate with systems like Spire (that exposes its own SDS server) we will not be able to rename resources that are coming from there.

See:
- https://github.com/spiffe/spire/blob/f687bf21e812a4bf027d88136032fc46497e7fe0/pkg/agent/endpoints/sdsv3/handler.go#L395
- https://github.com/spiffe/spire/blob/f6a11a0ae03353b3444007ecaa8c81b34931bb49/pkg/agent/endpoints/sdsv3/handler.go#L282
- https://github.com/spiffe/spire/blob/f6a11a0ae03353b3444007ecaa8c81b34931bb49/pkg/agent/endpoints/sdsv3/handler.go#L291
- https://github.com/spiffe/spire/blob/f6a11a0ae03353b3444007ecaa8c81b34931bb49/pkg/agent/endpoints/sdsv3/handler.go#L312

## Design

### Resource naming

#### 1. User resources

Names for these resources are defined in [Defining and migrating to consistent naming for non-system Envoy resources and stats](./077-migrating-to-consistent-and-well-defined-naming-for-non-system-envoy-resources-and-stats.md).

#### 2. System resources that can be tracked back to user resource

When possible, use `system_<KRI>`. It should be justified when designing why a user resource is system and not just `<kri>`.
For example, when a resource is not on a service to service traffic path, is shared between other parts of the system (like Kong Mesh's `MeshGlobalRateLimit` service) or we want to make it easy to exclude from metrics collection.

There are other cases where a resource is the result of merging multiple resources together (MTP rbac filters, MeshPassthrough, etc.).
This is out of scope and requires a separate MADR.

#### 3. System resources that are not related to user resource

When adding such resource the name should be explicit enough to help anyone understand where this is coming from.
The name should be namespaced so that it groups similar resources together and avoids collisions with other system resources.

For example dynamic config fetcher creates one listener and one route per object type.
We will namespace it as `system_dynamicconfig` and the DNS route will be `system_dynamicconfig_dns` and MeshMetric route will be `system_dynamicconfig_meshmetric`.

If any related configuration option or annotation name exists it should be taken into account.

#### 4. Resource contextual to the Dataplane

These resources are defined in [Defining and migrating to consistent naming for non-system Envoy resources and stats](./077-migrating-to-consistent-and-well-defined-naming-for-non-system-envoy-resources-and-stats.md).

#### 5. Secrets

There is nothing special about secrets, they adhere to the same rules (1-4) and can have exceptions (6).

#### 6. Exceptions

We can't control the name of external resources (i.e. SPIRE SDS secrets).
Any new exception and case should result in a MADR which can be linked from here.

#### 7. Regex to match system names

For `system_<KRI>` resources we will use the same regex as used for KRI but prefixed with `system_`.

For other names we will use `^system_([a-z0-9-]*_?)+$`, here is an example of the usage: https://regex101.com/r/Ic1bk5/2.

### Currently existing system resources and their new naming

| Resource type | Current Name                                                 | New Name                        | Example                                                      |
|---------------|--------------------------------------------------------------|---------------------------------|--------------------------------------------------------------|
| Cluster       | kuma:envoy:admin                                             | system_envoy_admin              |                                                              | 
| Listener      | _kuma:dynamicconfig                                          | system_dynamicconfig            |                                                              | 
| Route         | _kuma:dynamicconfig:dns                                      | system_dynamicconfig_dns        |                                                              | 
| Route         | _kuma:dynamicconfig:meshmetric                               | system_dynamicconfig_meshmetric |                                                              | 
| Listener      | _kuma:metrics:prometheus:backend-default                     | system_<kri>                    | system_kri_mm_mesh-1_us-east-2_kuma-demo_default_            | 
| Listener      | _kuma:metrics:prometheus:<backendName>                       | system_<kri>                    | system_kri_mm_mesh-1_us-east-2_kuma-demo_default_            | 
| Listener      | _kuma:metrics:opentelemetry:<backendName>                    | system_<kri>                    | system_kri_mm_mesh-1_us-east-2_kuma-demo_default_            | 
| Cluster       | _kuma:metrics:opentelemetry:<backendName>                    | system_<kri>                    | system_kri_mm_mesh-1_us-east-2_kuma-demo_default_            | 
| Cluster       | _kuma:metrics:hijacker                                       | system_metrics_hijacker         |                                                              | 
| Listener      | kuma:metrics:prometheus                                      | system_metrics_prometheus       |                                                              |
| VirtualHost   | kuma:metrics:prometheus                                      | system_metrics_prometheus       |                                                              |
| Cluster       | tracing:<name>                                               | system_tracing_<name>           |                                                              | 
| Cluster       | meshtrace:<type> (zipkin, datadog, otel)                     | system_<kri>                    | system_kri_mt_mesh-1_us-east-2_kuma-demo_default_            | 
| Cluster       | meshaccesslog:opentelemetry:<index> (0, 1, 2, ...)           | system_<kri>                    | system_kri_mal_mesh-1_us-east-2_kuma-demo_multiple-backends_ | 
| Cluster       | meshglobalratelimit:service                                  | system_<kri>                    | system_kri_mgrl___kong-mesh-system_mesh-rate-limit_          |
| Cluster       | kuma:readiness                                               | system_probe_readiness          |                                                              |
| Route         | 9Zuf5Tg79OuZcQITwBbQykxAk2u4fRKrwYn3//AL4Yo= (default route) | system_route_default            |                                                              |
| Secret        | mesh_ca:secret:all                                           | system_mtls_ca_all_meshes       |                                                              |
| Secret        | mesh_ca:secret:<mesh>                                        | system_mtls_ca_<mesh>           |                                                              |
| Secret        | identity_cert:secret:<mesh>                                  | system_mtls_identity_<mesh>     |                                                              |
| Cluster       | ads_cluster                                                  | system_ads                      |                                                              |
| Cluster       | plugins:bootstrap:k8s:hooks:apiServerBypass                  | system_kube_api_server_bypass   |                                                              |
| Listener      | plugins:bootstrap:k8s:hooks:apiServerBypass                  | system_kube_api_server_bypass   |                                                              |
| Listener      | kuma:dns                                                     | system_dns_builtin              |                                                              |

Deprecated features which resource names won't be changed:
- virtual probes:
  - probe:listener
  - probe:route_configuration
  - probe

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

Kong Mesh doesn't use any special type of resource that is not used by Kuma. 
In its policies, it only defines `meshglobalratelimit:service` cluster for MeshGlobalRateLimit
(as well as an endpoint, modifications to routes, virtual hosts and filter chains).

For MeshOPA it only modifies existing resources and doesn't create any new ones.

## Decision

We will use a naming scheme as described in "Resource naming" section.

## Notes

Names in codebase:
- https://github.com/kumahq/kuma/blob/bdc95fb8b8a4da2388948041171d5b9ecf4345a5/pkg/xds/envoy/names/resource_names.go
- https://github.com/kumahq/kuma/blob/9ed757d3e5955ea57d7940badf4468057ed46663/pkg/xds/envoy/secrets.go#L13
- https://github.com/kumahq/kuma/blob/950ff3353f7e85670717557691b42208b17a6579/pkg/plugins/policies/core/xds/meshroute/listeners.go#L343
- https://github.com/kumahq/kuma/blob/26af860614cf0792c0bff004ac95e8f5115808fc/pkg/xds/envoy/tags/match.go#L37
- https://github.com/kumahq/kuma/blob/7bafa578aad6e528befcb6c96f025542fd1f6870/pkg/plugins/policies/meshtrace/plugin/xds/configurer.go#L264
- https://github.com/kumahq/kuma-gui/blob/f7f9da37c335ba14151bb4a3e546437b7eae94c7/packages/kuma-gui/src/app/connections/data/index.ts#L125-L137

Istio related issues:
- https://github.com/istio/istio/issues/5311
- https://github.com/istio/istio/issues/31112#issuecomment-1124049572

Ping FE team because it might change the implementation of:
- https://github.com/kumahq/kuma-gui/blob/f7f9da37c335ba14151bb4a3e546437b7eae94c7/packages/kuma-gui/src/app/connections/data/index.ts#L125-L137

Secrets format:
- https://github.com/kumahq/kuma/blob/9ed757d3e5955ea57d7940badf4468057ed46663/pkg/xds/envoy/secrets.go#L26
