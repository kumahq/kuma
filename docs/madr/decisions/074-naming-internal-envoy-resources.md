# {short title of solved problem and solution}

* Status: {rejected | accepted} <!-- recommended to have the status as accepted proactively and then to change it if needed -->

Technical Story: {ticket/issue URL} <!-- link to the github issue -->

## Context and Problem Statement

{Describe the context and problem statement. You need to be as thorough as possible. If possible create set of use cases.
Always try thinking from end user point of view. Remember to not think in terms of solutions in this section. Try to be as 
objective as possible only describing current state and what we need.}

Some resources in Envoy don't correlate with the real resources in the store. That's why we can't use KRI formatted name for these resources.

Is it really the case for everything?
I think secrets have a relationship:
- mesh_ca in the store is `mesh-1.ca-builtin-cert-ca-1` that is defined in `mesh.backens[]` (CRD) so the kri could be `kri_secret_mesh-1___mesh-ca_`

Current:

```
mesh_ca:secret:mesh-1
```

```
kri_<resource-type>_<mesh>_<zone>_<namespace>_<resource-name>_<section-name>
```

Secrets format:
https://github.com/kumahq/kuma/blob/9ed757d3e5955ea57d7940badf4468057ed46663/pkg/xds/envoy/secrets.go#L26

Internal listeners:
- inbound
  - kuma:envoy:admin
  - _kuma:dynamicconfig

Internal clusters:
- access_log_sink
- kuma:envoy:admin
- kuma:readiness
- meshtrace_[zipkin|datadog|otel]

Internal virtual Hosts:
- _kuma:dynamicconfig
- kuma:envoy:admin

Routes (not all of them even have a name):
- _kuma:dynamicconfig:/dns
- 9Zuf5Tg79OuZcQITwBbQykxAk2u4fRKrwYn3//AL4Yo= (default route)

Names in codebase:
- https://github.com/kumahq/kuma/blob/bdc95fb8b8a4da2388948041171d5b9ecf4345a5/pkg/xds/envoy/names/resource_names.go
- https://github.com/kumahq/kuma/blob/9ed757d3e5955ea57d7940badf4468057ed46663/pkg/xds/envoy/secrets.go#L13
- https://github.com/kumahq/kuma/blob/950ff3353f7e85670717557691b42208b17a6579/pkg/plugins/policies/core/xds/meshroute/listeners.go#L343
- https://github.com/kumahq/kuma/blob/26af860614cf0792c0bff004ac95e8f5115808fc/pkg/xds/envoy/tags/match.go#L37
- https://github.com/kumahq/kuma/blob/7bafa578aad6e528befcb6c96f025542fd1f6870/pkg/plugins/policies/meshtrace/plugin/xds/configurer.go#L264

## Use cases

### As a user I want to exclude all stats related to internal resources from a query easily

So instead 

``` 
sum:envoy.cluster.upstream_rq.count{!envoy_cluster:kuma_readiness , !envoy_cluster:access_log_sink , !envoy_cluster:meshtrace_datadog}.as_count()
```

we can do:

``` 
sum:envoy.cluster.upstream_rq.count{!envoy_cluster:_*}.as_count()
```

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

### As a Kuma developer I want to have a consistent naming scheme for all resources in Envoy



## Design

{This is a place for main design, it is best to present multiple solutions. 
- Remember about examples. It is best to present then in context of previously defined use cases. 
- Add advantages and disadvantages sections to proposed solutions.
- When writing design remember to include history of how it evolved. This doc should be understandable without looking into git history.}

## Implications for Kong Mesh

{In this section we should look into how this changes affect Kong Mesh. For example we might need to update Kong Mesh policies to new API.}

## Decision

{Fill this section as last. This section should contain simplified description of selected solution.}

## Notes <!-- optional -->

{This section could include notes from meeting or open topics for discussion}
