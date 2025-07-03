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

- **User resource** - any resource that comes from passthrough, `Mesh*Service`, `Mesh*Route` and user defined `Secret`s
- **System resource** - any resource that is not a user resource

### Kuma system resource names

Listeners:
- inbound
  - kuma:envoy:admin
  - _kuma:dynamicconfig
  - listener./tmp/kuma-dp-728637052/kuma-mesh-metric-config.sock
  - listener.10.42.0.9_9901
  - listener.[__]_15001 # is it ipv6?
  - probe_listener
  - prometheus_listener

Clusters:
- access_log_sink
- kuma:envoy:admin
- kuma:metrics:hijacker
- kuma:readiness
- meshtrace_[zipkin|datadog|otel]
- envoy_grpc_streams (ads)

Virtual Hosts:
- _kuma:dynamicconfig
- kuma:envoy:admin

Routes (not all of them even have a name):
- _kuma:dynamicconfig:/dns
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
sum:envoy.cluster.upstream_rq.count{!envoy_cluster:_kuma*}.as_count()
```

You might be tempted to think that this use case is already covered by the `MeshMetric` profiles, 
and it partially is (the end result is the same) but this use case compliments it by making the profiles easier to implement and maintain.
It also allows users to do their own filtering more easily.

### As an operator I want to drop all stats related to internal resources

So I can use a processor like:

```yaml
processors:
  filter/drop_kuma_internal_clusters:
    metrics:
      include:
        match_type: regexp
        metric_names: [".*"]  # apply to all metrics
        attributes:
          - key: envoy_cluster
            value: "^_kuma_.*"  # matches if envoy_cluster starts with _
```

### As a Kuma developer I want to have a consistent naming scheme for all resources in Envoy

### As a Kuma developer I want to distinguish between system and user resources in Envoy

### As a Kuma developer I want to distinguish between system resources exposed and not exposed in metrics

### As a user I want the cardinality of these resource names to be low

Need to avoid:
- IPs / non constant bits
- Avoid dataplane name.

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

### Use a `^system_([a-z0-9]+_{0,1})+$` regex to name system resources

All changes will be behind the same feature flag as in [Migrating to consistent and well-defined naming for non-system Envoy resources and stats](./077-migrating-to-consistent-and-well-defined-naming-for-non-system-envoy-resources-and-stats.md).

All system resources will conform to `^system_([a-z0-9]+_{0,1})+$` regex (we shouldn't use `_system:` because of [this issue](https://github.com/kumahq/kuma/issues/2363)).

System resources that can be traced back to a Kuma resource with a valid KRI will have take the form of `system_<KRI>`.

For example:

```
system_kri_mtr_mesh-1__kuma-system_my-meshtrace_
```

Which indicates a system resource that was created for a `MeshTrace` resource.

Resources that are not correlated with any Kuma resource like:

```
kuma:envoy:admin
```

Will conform with the previously mentioned regex:

```
system_envoy_admin
```

and should describe the resource as accurately and plainly as possible (it **MUST** take into account any related configuration option or annotation if exists).

It means that `listener.0.0.0.0_15001` which can be configured by (`kuma.io/transparent-proxying-outbound-port`) will become `system_transparent_proxy_outbound_listener`.

#### Secrets

In the future we will most likely introduce new policies - `MeshTrust`, `MeshIdentity` and `SPIRE` integration.

For `Spire` we can't control the secrets as mentioned in [Non use cases section](#as-a-kuma-developer-i-want-to-rename-resources-that-are-coming-from-outside-kuma).

For a secret related to `MeshTrust`s we can't use `KRI` because it's a combination of all of them and `KRI` currently don't deal with lists of resources.

For `MeshIdentity` and user defined secrets we can use `KRI` but that will be covered by [Migrating to consistent and well-defined naming for non-system Envoy resources and stats](./077-migrating-to-consistent-and-well-defined-naming-for-non-system-envoy-resources-and-stats.md).

For all other secrets we can use `^system_([a-z0-9]+_{0,1})+$` naming.

#### Enforcement

##### Builders

To enforce these rules we will add checks in resource builders like [ClusterBuilder](https://github.com/kumahq/kuma/blob/dedaba5b9de1bd134dce813ae49b3475d5d24e6b/pkg/xds/envoy/clusters/cluster_builder.go#L80).
If anyone tries to add a new resource that doesn't conform to any type it will fail in unit tests.

Pros:
- Simple to implement

Cons:
- Can be easily skipped by not using a builder

##### Inspecting golden files

We could have a target that goes over golden files and checks the names of all resources.

Pros:
- Can catch things that are not created by a builder

Cons:
- Can be skipped by a code path not using golden files or not tested

##### Custom linter

A linter that would use `go/ssa` and `callgraph` that would figure out if an Envoy resource with a name that doesn't conform to the regex is created.

Pros:
- Should catch all the cases

Cons:
- Might be really hard to implement

### Use a `^_kuma_[a-z0-9_]+$` regex to name system resources

Same as above, but with `_kuma_` instead of `system_`.

## Implications for Kong Mesh

Kong Mesh doesn't use any special type of resource that is not used by Kuma. 
In its policies, it only defines `meshglobalratelimit:service` cluster for MeshGlobalRateLimit
(as well as an endpoint, modifications to routes, virtual hosts and filter chains).

For MeshOPA it only modifies existing resources and doesn't create any new ones.

## Decision

We will use a regex on all system resources as described in "Use a `^system_([a-z0-9]+_{0,1})+$` regex to name system resources" section.

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
