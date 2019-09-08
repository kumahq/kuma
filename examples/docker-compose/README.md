Kuma Control Plane inside Docker Compose 
====================

WARNING: Docker Compose setup doesn't work at the moment

## Pre-requirements

- `docker-compose`

## Usage

### Build and run the containers

```bash
make run/example/docker-compose -C ../..
```

### Make test requests

```bash
make wait/example/docker-compose -C ../..
make curl/example/docker-compose -C ../..
```

### Verify Envoy stats

```bash
make verify/example/docker-compose -C ../..
```

### Observe Envoy stats

```bash
make stats/example/docker-compose -C ../..
```

E.g.,
```
# TYPE envoy_cluster_upstream_rq_total counter
envoy_cluster_upstream_rq_total{envoy_cluster_name="pass_through"} 11
envoy_cluster_upstream_rq_total{envoy_cluster_name="ads_cluster"} 1
envoy_cluster_upstream_rq_total{envoy_cluster_name="localhost_8080"} 11
```

where

* `envoy_cluster_upstream_rq_total{envoy_cluster_name="localhost_8080"}` is a number of `inbound` requests
* `envoy_cluster_upstream_rq_total{envoy_cluster_name="pass_through"}` is a number of `outbound` requests
