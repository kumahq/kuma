# Prometheus Metrics

## Context

* At the moment, metrics collected by `Envoy` dataplanes are not available for scraping by Prometheus
* The reason for that is that `Envoy` exposes metrics as part of its Admin API; for security reasons, it is not possible to make the entire Admin API publicly accessible
* As a solution, it is possible to configure an extra `Listener` on every `Envoy` dataplane
  * that will be publicly accessible (i.e., accessible to Prometheus)
  * that will only respond to requests for `Envoy` metrics (original `/stats/prometheus` endpoint in Admin API)

## Requirements

1. Support "one-click" enablement of Prometheus metrics
2. Support `Mesh`-wide configuration that will be a default for all `Dataplane`s in it
3. Support per-`Dataplane` configuration overrides
   * e.g., the most practical use case is the ability to override a port on which Envoy metrics should be exposed
4. Support integration with Prometheus discovery on Kubernetes
   * technically, annotate `Pods` to enable auto-discovery by Prometheus, e.g.
     * `prometheus.io/scrape: "true"` 
     * `prometheus.io/port: "1234"` 
     * `prometheus.io/path: "/metrics"` 
     * see [Prometheus Helm Chart](https://github.com/helm/charts/blob/2f46f4fa4f7a995381f5add685bb762265b9ff15/stable/prometheus/values.yaml) for more details
   * TODO: What if a user already defined those annotations for his own app ?
     Do these annotations support multiple values ?
5. Support integration with Prometheus discovery on Universal
   * technically, it should be possible to configure Prometheus to discover targets for scraping sthrough integration of some kind with `Kuma Control Plane`, e.g.
     * we could build a helper tool (to run as a cron job) that would extract scrape targets info out of `Kuma Control Plane` and put it into a file Prometheus could pick it up from (seems to be a [recommend approach](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#file_sd_config))
     * we could contribute to Prometheus support for a discovery protocol convenient for Control Planes
     * `Kuma Control Plane` could implement a discovery protocol already supported by Prometheus (although, there is no obvious candidate)

## Out of scope

* Securing access to metrics with mTLS
* Support for a use case where Prometheus is deployed inside Service Mesh

## Proposed configuration model

### Mesh-wide defaults

#### Universal

```yaml
type: Mesh
name: default
mtls:
  ...
metrics:
  prometheus:
    port: 5670
    path: /metrics
```

Notice that the minimal configuration users will have to provide is actually much simpler:

```yaml
type: Mesh
name: default
metrics:
  prometheus: {}
```

`Kuma Control Plane` will automatically set default values for `port` and `path`.

### Per-Dataplane overrides

#### Universal

```yaml
type: Dataplane
mesh: default
name: web
networking:
  ...
metrics:
  prometheus:
    port: 5670
    path: /metrics
```

#### Kubernetes

Since on Kubernetes `Dataplane` resources are generated automatically (and are not supposed to be edited by users), it should be possible to override `Mesh`-wide Prometheus settings through a use of annotations, namely

* `kuma.io/prometheus-port`
* `kuma.io/prometheus-path`

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: kuma-example
  name: kuma-tcp-echo
spec:
  ...
  template:
    metadata:
      ...
      annotations:
        prometheus.metrics.kuma.io/port: "1234"               # override Mesh-wide default 'port'
        prometheus.metrics.kuma.io/path: "/non-standard-path" # override Mesh-wide default 'path'
    spec:
      containers:
      - name: kuma-tcp-echo
        image: docker.io/kuma-tcp-echo:0.1.0
        imagePullPolicy: Always
        ports:
        - containerPort: 8000
```

## Implementation Notes

* It is assumed, there will be no mTLS between Prometheus and the metrics endpoint exposed by a dataplane (`Envoy`)
