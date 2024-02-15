# MeshMetric profiles (predefined subsets of all metrics)

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/8845

## Context and Problem Statement

There is a lot of default Envoy metrics, even with `usedOnly` enabled.
Users might be overwhelmed by the number of metrics and not know what it's important what is less important.
Hosted providers usually charge based on the number of metrics ingested.
When using trial versions of hosted metrics users run out of free credits pretty fast.

To make it more digestible for the users and cheaper we should introduce profiles that only contain subsets of all Envoy metrics.

## Decision Drivers <!-- optional -->

* {driver 1, e.g., a force, facing concern, …}
* {driver 2, e.g., a force, facing concern, …}
* … <!-- numbers of drivers can vary -->

## Considered Options

* Base profiles on expert knowledge and current grafana dashboards
* Base profiles on stats popularity

## Decision Outcome

Chosen option: "{option 1}", because {justification. e.g., only option, which meets k.o. criterion decision driver | which resolves force {force} | … | comes out best (see below)}.

### Positive Consequences <!-- optional -->

* {e.g., improvement of quality attribute satisfaction, follow-up decisions required, …}
* …

### Negative Consequences <!-- optional -->

* {e.g., compromising quality attribute, follow-up decisions required, …}
* …

## Pros and Cons of the Options

### Base profiles on expert knowledge, external dashboards and our grafana dashboards

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
curl -s https://raw.githubusercontent.com/DataDog/integrations-core/master/envoy/metadata.csv | grep -v Legacy | wc -l                                                                                                                                         -- INSERT --
```

We can build automation on top of this, so when we update dashboards or Envoy changes metrics it emits we know about this and can adjust accordingly
(just like now official Envoy dashboards - `envoy_cluster_upstream_rq_time_99percentile` have metrics that no longer exist).

As you can see there is no easy way to track everything (Envoy does not by default print all possible metrics)
so I strongly suggest adding a feature to dynamically (by regex for example) add/remove metrics to/from existing profiles.

#### Profiles selected:

##### All

Nothing is removed, everything included in Envoy, people can manually remove stuff.

##### Extensive

- All available dashboards + [Charly's demo regexes](https://github.com/lahabana/demo-scene/blob/a48ec6e0079601d340f79613549e1b2a4ea715a1/mesh-localityaware/k8s/otel-collectors.yaml#L174)

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
- Errors
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

#### Schema



### {option 2}

{example | description | pointer to more information | …} <!-- optional -->

* Good, because {argument a}
* Good, because {argument b}
* Bad, because {argument c}
* … <!-- numbers of pros and cons can vary -->

### {option 3}

{example | description | pointer to more information | …} <!-- optional -->

* Good, because {argument a}
* Good, because {argument b}
* Bad, because {argument c}
* … <!-- numbers of pros and cons can vary -->

## Links <!-- optional -->

* {Link type} {Link to ADR} <!-- example: Refined by [ADR-0005](0005-example.md) -->
* … <!-- numbers of links can vary -->
