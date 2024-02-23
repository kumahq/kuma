# MeshMetric profiles (predefined subsets of all metrics)

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/8845

## Context and Problem Statement

There is a lot of default Envoy metrics, even with `usedOnly` enabled.
Users might be overwhelmed by the number of metrics and not know what it's important what is less important.
Hosted providers usually charge based on the number of metrics ingested.
When using trial versions of hosted metrics users run out of free credits pretty fast.

To make it more digestible for the users and cheaper we should introduce profiles that only contain subsets of all Envoy metrics.

## Decision Drivers

* no loss of quality while minimizing number of metrics the most (ideally we want to reduce the number of metrics while still providing access to the ones that are the most valuable)

## Considered Options

### Use AI to generate Profiles

Useless, after a couple of metrics just prints "envoy_cluster_manager_warming_clusters_total" indefinitely.

### Base profiles on expert knowledge, external dashboards and our grafana dashboards

#### Where to get data from?

We built the dashboards to show what is important to look at, we could extract the list of metrics from these dashboards like this:

```bash
cat app/kumactl/data/install/k8s/metrics/grafana/kuma-dataplane.json | jq 'try .panels[] | try .targets[] | try .expr' | grep -E -o '\benvoy_[a-zA-Z0-9_]+\b' | sort | uniq
cat app/kumactl/data/install/k8s/metrics/grafana/kuma-gateway.json | jq 'try .panels[] | try .targets[] | try .expr' | grep -E -o '\benvoy_[a-zA-Z0-9_]+\b' | sort | uniq
cat app/kumactl/data/install/k8s/metrics/grafana/kuma-service-to-service.json | jq 'try .panels[] | try .targets[] | try .expr' | grep -E -o '\benvoy_[a-zA-Z0-9_]+\b' | sort | uniq
cat app/kumactl/data/install/k8s/metrics/grafana/kuma-mesh.json | jq 'try .panels[] | try .targets[] | try .expr' | grep -E -o '\benvoy_[a-zA-Z0-9_]+\b' | sort | uniq
cat app/kumactl/data/install/k8s/metrics/grafana/kuma-service.json | jq 'try .panels[] | try .targets[] | try .expr' | grep -E -o '\benvoy_[a-zA-Z0-9_]+\b' | sort | uniq
```

and put this in a profile.

We can do similar thing for other dashboards, like official Envoy Datadog dashboard:

```bash
cat docs/madr/decisions/assets/038/envoy-datadog-dashboard.json | jq 'try .widgets[] | try .definition | try .widgets[] | try .definition | try .requests[] | try .queries[] | .query' | grep -E -o '\benvoy[\._a-zA-Z0-9]+{' | tr -d '{'
```

Or for Consul Grafana dashboards:

```bash
cat docs/madr/decisions/assets/038/consul-grafana.json | jq 'try .panels[] | try .targets[] | try .expr' | grep -E -o '\benvoy_[a-zA-Z0-9_]+\b' | sort | uniq
```

All of these metrics combined result in 100 metrics.

By default, Envoy starts with 378 metrics (unfortunately it's not a complete list):

```bash
docker run --rm -it -p 9901:9901 -p 10000:10000 envoyproxy/envoy:v1.29.1
curl -s localhost:10000 > /dev/null
curl -s localhost:9901/stats/prometheus | grep -E -o '\benvoy_[a-zA-Z0-9_]+\b' | sort | uniq | wc -l
# 378
```

And datadog lists 990 metrics in total https://github.com/DataDog/integrations-core/blob/master/envoy/metadata.csv and 329 non-legacy ones

```bash
curl -s https://raw.githubusercontent.com/DataDog/integrations-core/master/envoy/metadata.csv | grep -v Legacy | wc -l
```

We can build automation on top of this, so when we update dashboards or Envoy changes metrics it emits we know about this and can adjust accordingly
(just like now official Envoy dashboards - `envoy_cluster_upstream_rq_time_99percentile` have metrics that no longer exist).

As you can see there is no easy way to track everything (Envoy does not by default print all possible metrics)
so I strongly suggest adding a feature to dynamically (by regex for example) add/remove metrics to/from existing profiles.

#### Behaviour

Profiles can be either static or additive.
A static profile means you get X metrics with profile Y and you can't combine that with other profiles.
Additive profiles allow you to mix and match which metrics you need depending on current circumstances (e.g. debugging HDS issue)
and specifics of the underlying application (e.g. `tcp` app does not need `http` metrics).

#### User defined profiles

We could allow people to define their custom profiles and reference them in `MeshMetric`.

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshMetricsProfile
metadata:
  name: my-profile
  namespace: kuma-system
spec:
  default:
    appendMetricExact:
      - envoy_cluster_upstream_rq_xx
      - envoy_http_downstream_rq_xx
    appendMetricRegex:
      - .*error.*
```

##### Static

```yaml
sidecar:
  usedOnly: true # true or false
  profile:
    ref: my-profile
```

```yaml
sidecar:
  usedOnly: true # true or false
  profile:
    ref: default # our predefined profile
```

##### Additive

```yaml
sidecar:
  usedOnly: true # true or false
  appendProfiles:
    - ref: my-profile
    - ref: default # our predefined profile
```

The question here is: do we want our builtin profiles to be defined like that (this possibly means a user can edit them a bit like default policies)
or should they be hidden from the user (and only referencable)?

#### Profiles suggested:

Suggestions to the merge / split and naming are welcomed.
The first ones are granularity based the last ones are feature focused (see [additive profiles](#behaviour)).

##### All

Nothing is removed, everything included in Envoy, people can manually remove stuff.

##### Extensive

- All available dashboards + [Charly's demo regexes](https://github.com/lahabana/demo-scene/blob/a48ec6e0079601d340f79613549e1b2a4ea715a1/mesh-localityaware/k8s/otel-collectors.yaml#L174)
  - `envoy_cluster_upstream_cx_.*`
  - `envoy_cluster_upstream_rq_.*`
  - `envoy_cluster_circuit_breakers_.*`
  - `envoy_http_downstream_.*`
  - `envoy_listener_downstream_.*`
  - `envoy_listener_http_.*`

##### Comprehensive

- All available dashboards

##### Default

- Our dashboards

##### Minimal

Only golden 4 (by regex / or exact):
- Latency
  - `.*_rq_time_.*` which is:
    - envoy_cluster_internal_upstream_rq_time_bucket
    - envoy_cluster_internal_upstream_rq_time_count
    - envoy_cluster_internal_upstream_rq_time_sum
    - envoy_cluster_external_upstream_rq_time_bucket
    - envoy_cluster_external_upstream_rq_time_count
    - envoy_cluster_external_upstream_rq_time_sum
    - envoy_cluster_upstream_rq_time_bucket
    - envoy_cluster_upstream_rq_time_count
    - envoy_cluster_upstream_rq_time_sum
    - envoy_http_downstream_rq_time_bucket
    - envoy_http_downstream_rq_time_count
    - envoy_http_downstream_rq_time_sum
  - or just `envoy_cluster_upstream_rq_time` and `envoy_http_downstream_rq_time`
  - `.*cx_length_ms.*`
    - envoy_cluster_upstream_cx_length_ms_bucket
    - envoy_cluster_upstream_cx_length_ms_count
    - envoy_cluster_upstream_cx_length_ms_sum
    - envoy_http_downstream_cx_length_ms_bucket
    - envoy_http_downstream_cx_length_ms_count
    - envoy_http_downstream_cx_length_ms_sum
    - envoy_listener_admin_downstream_cx_length_ms_bucket
    - envoy_listener_admin_downstream_cx_length_ms_count
    - envoy_listener_admin_downstream_cx_length_ms_sum
    - envoy_listener_downstream_cx_length_ms_bucket
    - envoy_listener_downstream_cx_length_ms_count
    - envoy_listener_downstream_cx_length_ms_sum
  - or just `envoy_cluster_upstream_cx_length_ms`
- Traffic
  - `.*cx_count.*` (connections total)
  - `.*cx_active.*` (active connection)
  - `.*_rq` (upstream/downstream requests broken by specific codes e.g. 200/201)
  - `.*bytes*` (bytes sent/received/not send)
- Errors (just went over combined stats from all dashboards and picked the ones that had anything to do with errors)
  - we get 5xx from `*_rq`
  - `.*timeout.*`
  - `.*health_check.*`
  - `.*lb_healthy_panic.*`
  - `.*cx_destroy.*`
  - `envoy_cluster_membership_degraded`
  - `envoy_cluster_membership_healthy`
  - `envoy_cluster_ssl_connection_error`
  - `.*error.*`
  - `.*fail.*` (has also envoy_cluster_ssl_fail envoy_cluster_update_failure)
  - `.*reset.*`
  - `.*outlier_detection_ejections.*`
  - `envoy_cluster_upstream_cx_pool_overflow_count`
  - `protocol_error`
  - `envoy_cluster_upstream_rq_cancelled`
  - `envoy_cluster_upstream_rq_max_duration_reached`
  - `envoy_cluster_upstream_rq_pending_failure_eject`
  - `.*overflow.*`
  - `.*no_cluster.*`
  - `.*no_route.*`
  - `.*reject.*`
  - `envoy_listener_no_filter_chain_match`
  - `.*denied.*` (envoy_rbac_denied, envoy_rbac_shadow_denied)
  - `envoy_server_days_until_first_cert_expiring`
- Saturation
  - `.*memory.*` (allocated, heap, physical)

##### Nothing

- Just an empty profile and people manually can add things to this.

##### ADS / XDS / Management server

All stats from: https://www.envoyproxy.io/docs/envoy/latest/configuration/overview/mgmt_server#subscription-statistics
divided into CDS/LDS/RDS/SDS/VHDS by `listener_manager.[cds, lds, rds, sds, vhds]`.

This would enable ADS (aggregated xDS) stats, might be helpful when debugging DP-CP communication on various resources (HDS, LDS, SDS).
On the other hand we would have to parse labels as well because these metrics are under normal metrics like `envoy.cluster.upstream_rq_timeout.count`,
but are published with `envoy_cluster:ads_cluster` label.

##### HTTPApp

- https://www.envoyproxy.io/docs/envoy/latest/configuration/upstream/cluster_manager/cluster_stats.html#general (HTTP related)
- https://www.envoyproxy.io/docs/envoy/latest/configuration/upstream/cluster_manager/cluster_stats.html#dynamic-http-statistics
- https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_conn_man/stats.html#http-1-codec-statistics
- https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_conn_man/stats.html#http-2-codec-statistics 
- https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_conn_man/stats.html#http-3-codec-statistics

##### GRPCApp

- https://www.envoyproxy.io/docs/envoy/latest/configuration/upstream/cluster_manager/cluster_stats.html#general (HTTP related)
- https://www.envoyproxy.io/docs/envoy/latest/configuration/upstream/cluster_manager/cluster_stats.html#dynamic-http-statistics
- https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_conn_man/stats.html#http-2-codec-statistics
- https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/grpc_stats_filter.html#grpc-statistics

##### TCPApp

- https://www.envoyproxy.io/docs/envoy/latest/configuration/upstream/cluster_manager/cluster_stats.html#general (TCP related)
- https://www.envoyproxy.io/docs/envoy/latest/configuration/upstream/cluster_manager/cluster_stats.html#tcp-statistics
- https://www.envoyproxy.io/docs/envoy/latest/configuration/listeners/stats.html#tcp-statistics

##### TLS

- https://www.envoyproxy.io/docs/envoy/latest/configuration/upstream/cluster_manager/cluster_stats.html#tls-statistics
- https://www.envoyproxy.io/docs/envoy/latest/configuration/listeners/listener_filters/tls_inspector.html#statistics
- https://www.envoyproxy.io/docs/envoy/latest/configuration/listeners/stats.html#tls-statistics

##### Fault injection

All stats from: https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/fault_filter.html#statistics

##### RBAC

All stats from:
- https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/rbac_filter.html#statistics 
- https://www.envoyproxy.io/docs/envoy/latest/configuration/listeners/network_filters/rbac_filter.html#statistics

##### EnvoyServer

All stats from: https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/statistics.html#statistics

##### AccessLog

All stats from: https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/stats.html#statistics

##### LoadBalancer

All stats from: https://www.envoyproxy.io/docs/envoy/latest/configuration/upstream/cluster_manager/cluster_stats.html#load-balancer-statistics

##### CircuitBreaker

For circuit breaker we use both circuit breaker and outlier detection so we could combine these:
- https://www.envoyproxy.io/docs/envoy/latest/configuration/upstream/cluster_manager/cluster_stats.html#circuit-breakers-statistics
- https://www.envoyproxy.io/docs/envoy/latest/configuration/upstream/cluster_manager/cluster_stats.html#outlier-detection-statistics

##### Tracing

- https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_conn_man/stats.html#tracing-statistics

#### Schema

I suggest changing the current schema of

```yaml
sidecar:
  usedOnly: true # true or false
  profile: minimal # one of minimal, default, full
  regex: http2_act.* # only profile or regex can be defined
```

to

```yaml
sidecar:
  usedOnly: true # true or false
  profile:
    name: default # one of `nothing`, `minimal`, `default`, `comprehensive`, `all`
    exclude: regex2.* # first exclude
    include: regex1.* # then include (include takes over)
```

#### Implementation

Just like we [mutate responses for metrics hijacker](https://github.com/kumahq/kuma/blob/d6c9ce64ac5e7ba1f5dbb9fb410e7d9410b67815/app/kuma-dp/pkg/dataplane/metrics/server.go#L348)
we can add a filter mutator to reduce the number of metrics (same thing for [OTEL](https://github.com/kumahq/kuma/blob/d6c9ce64ac5e7ba1f5dbb9fb410e7d9410b67815/app/kuma-dp/pkg/dataplane/metrics/metrics_producer.go#L106)).

#### Validation

After all profiles are compiled from regexes make sure that they include the ones on the lower levels (all includes default, default includes minimal etc.)
Make sure that with `default` profile (or the profile chosen for dashboards) all dashboards are populated.

Can we somehow track if users are happy with the defined profiles?

#### Additional work

We could adjust our dashboards to reflect which graphs are going to be populated in which profile.

## Decision Outcome

Chosen option: "Base profiles on expert knowledge, external dashboards and our grafana dashboards", because it provides a good mix of inputs (our recommended metrics, others recommended metrics, expert knowledge).

### Positive Consequences

* should cover all typical scenarios
* non-typical scenarios can be handled by include/exclude
* allows us to build some automation to make it more resilient to changes

### Negative Consequences

* some processing power needed to filter metrics

## Links

* https://github.com/lahabana/demo-scene/blob/a48ec6e0079601d340f79613549e1b2a4ea715a1/mesh-localityaware/k8s/otel-collectors.yaml#L174
* https://docs.datadoghq.com/integrations/envoy
* https://docs.datadoghq.com/integrations/istio
