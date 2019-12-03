Kuma Control Plane inside Docker Compose
====================

## Pre-requirements

- `docker-compose`

## Usage

### Build Control Plane images

```bash
make build/example/docker-compose -C ../../../..
```

### Deploy example setup into Docker Compose

```bash
make deploy/example/docker-compose -C ../../../..
```

### Make test requests

```bash
make wait/example/docker-compose -C ../../../..
make curl/example/docker-compose -C ../../../..
```

### Verify Envoy stats

```bash
make verify/example/docker-compose -C ../../../..
```

### Observe Envoy stats

```bash
make stats/example/docker-compose -C ../../../..
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

### Verify traffic routing without mTLS

```bash
make verify/traffic-routing/docker-compose/without-mtls -C ../../../..
```

### Verify traffic routing with mTLS

```bash
make verify/traffic-routing/docker-compose/with-mtls -C ../../../..
```

### Cleanup

```bash
make undeploy/example/docker-compose -C ../../../..
```
