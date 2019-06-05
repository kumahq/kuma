Konvoy Control Plane inside Minikube
====================

## Pre-requirements

- `minikube`
- `kubectl`

## Usage

### Start Minikube

```bash
minikube start
```

### Build Control Plane image

```bash
make build/example/minikube -C ../..
```

### Deploy demo setup into Minikube

```bash
make deploy/example/minikube -C ../..
```

### Make test requests

```bash
make wait/example/minikube -C ../..
make curl/example/minikube -C ../..
```

### Verify Envoy stats

```bash
make verify/example/minikube -C ../..
```

### Observe Envoy stats

```bash
make stats/example/minikube -C ../..
```

E.g.,
```
# TYPE envoy_cluster_upstream_rq_total counter
envoy_cluster_upstream_rq_total{envoy_cluster_name="localhost_8000"} 11
envoy_cluster_upstream_rq_total{envoy_cluster_name="ads_cluster"} 1
envoy_cluster_upstream_rq_total{envoy_cluster_name="pass_through"} 3
```

where

* `envoy_cluster_upstream_rq_total{envoy_cluster_name="localhost_8000"}` is a number of `inbound` requests
* `envoy_cluster_upstream_rq_total{envoy_cluster_name="pass_through"}` is a number of `outbound` requests
