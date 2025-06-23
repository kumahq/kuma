# Standardized Naming for internal xDS Resources

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/13266

## Context and Problem Statement

We use resource names to identify listeners, clusters, routes, virtual hosts and secrets in the xDS APIs.
These names appear in metrics, logs, and debug endpoints (config_dump), and they are often used for monitoring, filtering, and diagnostics.

It's important to know the difference between internal resources (the ones that are defined by Kuma itself, and they usually do not concern the end users) and external resources (the ones that are defined indirectly by the user and the users are interested in).
For example, clusters like `kuma:envoy:admin` or `_kuma:dynamicconfig` are internal to Kuma, while `kri_msvc_mesh-1_us-east-2_kuma-demo_backend_httpport` cluster is an external resource.

Currently, there is inconsistency in how these resource names are formed, some of them have a `_kuma` prefix, some of them just contain `kuma` and some are free form.

Some internal resources do not map clearly to Kuma resources in the data store, making it difficult to relate metrics to Kuma abstractions.

There are also resources that are in a bit of a gray area, like `meshtrace_datadog` which is a cluster created for `MeshTrace` policy.
It seems internal, but it could be traced down to MeshTrace resource `kri_meshtrace_mesh-1__kuma-system_my-meshtrace_`.

#### Examples of kuma resource names

Listeners:
- inbound
  - kuma:envoy:admin
  - _kuma:dynamicconfig
  - listener./tmp/kuma-dp-728637052/kuma-mesh-metric-config.sock
  - listener.0.0.0.0_15001
  - listener.0.0.0.0_15006
  - listener.10.42.0.9_9901
  - listener.[__]_15001 # is it ipv6?
  - probe_listener
  - prometheus_listener
- outbound
  - outbound:passthrough:ipv4
  - outbound:passthrough:ipv6

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

### Looking at a resource name I want to easily relate it to Kuma concepts

### As a user I want my MeshProxyPatch to continue working after the change

I don't think that's going to be possible.
In MeshProxyPatch you can use `name` matcher to match a resource name but if we change the name of the resource, it will not match anymore.

### As a user I want to easily exclude all stats related to internal resources from a query

So instead 

``` 
sum:envoy.cluster.upstream_rq.count{!envoy_cluster:kuma_readiness , !envoy_cluster:access_log_sink , !envoy_cluster:meshtrace_datadog}.as_count()
```

we can do:

``` 
sum:envoy.cluster.upstream_rq.count{!envoy_cluster:_kuma*}.as_count()
```

**Is this a valid use case if we have profiles in MeshMetric?**

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
            value: "^_.*"  # matches if envoy_cluster starts with _
```

**Is this a valid use case if we have profiles in MeshMetric?**

### As a Kuma developer I want to have a consistent naming scheme for all resources in Envoy

### As a Kuma developer I want to distinguish between internal and external resources in Envoy

### As a Kuma developer I want to distinguish between external resources exposed and not exposed in metrics

## Design

TBA

## Implications for Kong Mesh

TBA

## Decision

TBA

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
