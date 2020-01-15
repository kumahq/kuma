# kuma-prometheus-sd

## Overview

`kuma-prometheus-sd` is an **adapter** that integrates `Prometheus` service discovery mechanism with [Kuma](https://kuma.io) Service Mesh.

Practically, it means that `Prometheus` will be retrieving a list of scrape targets **directly from** `Kuma` Control Plane rather than out of `Kubernetes`, `Consul`, `AWS`, etc.

**Direct integration** between `Kuma` Control Plane and `Prometheus` has the following **advantages**:
* **consistent ease of use** of `Kuma` Service Mesh **in any environment**, be it `Kubernetes`, legacy datacenter, VMs, bare metal, etc
* inherent **support for hybrid environments**
* **dynamism of reconfiguration** that would not otherwise be possible even on `Kubernetes` (where `Prometheus` configuration would have to remain unchanged throught the entire lifecycle of a `Pod`)

## How it works

`kuma-prometheus-sd` is meant to run alongside `Prometheus` instance.

It is responsible for talking to `Kuma` Control Plane and fetching an up-to-date list of scrape targets from it.

It then transforms that information into a format that `Prometheus` can understand, and saves it into a file on disk.

`Prometheus` watches for changes to that file and updates its scraping configuration accordingly.

## How to use

First, you need to run `kuma-prometheus-sd`, e.g. by using the following command:

```shell
kuma-prometheus-sd run \
  --cp-address=http://kuma-control-plane.internal:5681 \
  --output-file=/var/run/kuma-prometheus-sd/kuma.file_sd.json
```

The above configuration tells `kuma-prometheus-sd` to talk to `Kuma` Control Plane at http://kuma-control-plane.internal:5681 and save the list of scrape targets to `/var/run/kuma-prometheus-sd/kuma.file_sd.json`.

Then, you need to set up `Prometheus` to read from that file, e.g. by using `prometheus.yml` config with the following contents:

```yaml
scrape_configs:
- job_name: 'kuma-dataplanes'
  scrape_interval: 15s
  file_sd_configs:
  - files:
    - /var/run/kuma-prometheus-sd/kuma.file_sd.json
```

and running

```shell
prometheus --config.file=prometheus.yml
```

That's it!

Now, your `Prometheus` instance will always be aware of the most up-to-date list of `Kuma` dataplanes to scrape metrics from.

## How it's implemented

Under the hood, we use [Envoy xDS gRPC protocol](https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol) to deliver a list of scrape targets from `Kuma` Control Plane to `kuma-prometheus-sd`.

By employing bi-directional streaming, we give `Kuma` Control Plane  **full control** over **timing** and **scope of configuration updates**, while `kuma-prometheus-sd` remains logic-less.

This way, `kuma-prometheus-sd` can be deployed into your infrustructure just once and **stay out of** `Kuma` **upgrade path** most of the time.

## Protocol

A unit of configuration for `kuma-prometheus-sd` is called `Monitoring Assignment`, which is an equivalent of [targetgroup](https://github.com/prometheus/prometheus/blob/master/discovery/targetgroup/targetgroup.go) in `Prometheus`.

E.g., a `Monitoring Assignment` that instructs `Prometheus` to scrape metrics from a single `Kuma` dataplane looks the following way:

```yaml
name: /meshes/default/dataplanes/backend-01
targets:
- labels:
    __address__: 192.168.0.1:8080
    __scheme__: http
    __metrics_path__: /metrics
    instance: backend-01
    dataplane: backend-01
    env: prod
labels:
  job: backend
  mesh: default
  service: backend
```

To serve `Monitoring Assignments`, `Kuma` Control Plane implements a gRPC service called `Monitoring Assignment Discovery Service` (MADS).

And `kuma-prometheus-sd` is a client of that service.

Definitions of both `Monitoring Assignment` and `MADS` are available at [kuma/api/observability](https://github.com/Kong/kuma/blob/master/api/observability/v1alpha1/mads.proto).

## References

1. [Traffic Metrics in Kuma Service Mesh](https://kuma.io/docs/latest/policies/#traffic-metrics) @ `kuma.io/docs`
2. [Monitoring Assignment Discovery Service](https://github.com/Kong/kuma/blob/master/api/observability/v1alpha1/mads.proto) @ `github.com/Kong/kuma`
3. [Implementing Custom Service Discovery](https://prometheus.io/blog/2018/07/05/implementing-custom-sd/) @ `prometheus.io/blog`
4. [documentation/examples/custom-sd](https://github.com/prometheus/prometheus/tree/master/documentation/examples/custom-sd) @ `github.com/prometheus/prometheus`
5. [Envoy xDS gRPC protocol](https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol) @ `envoyproxy.io/docs`
