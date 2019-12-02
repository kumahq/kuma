Kuma Control Plane inside Minikube
====================

## Pre-requirements

- `minikube`
- `kubectl`

## Usage

### Start Minikube

```bash
minikube start
```

### Build Control Plane images

```bash
make build/example/minikube -C ../../../..
```

### Deploy example setup into Minikube

```bash
make deploy/example/minikube -C ../../../..
```

### Make test requests

```bash
make wait/example/minikube -C ../../../..
make curl/example/minikube -C ../../../..
```

### Verify Envoy stats

```bash
make verify/example/minikube -C ../../../..
```

### Observe Envoy stats

```bash
make stats/example/minikube -C ../../../..
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

### Enable mTLS

```bash
make apply/example/minikube/mtls -C ../../../..
```

### Wait until Envoy is configured for mTLS

```bash
make wait/example/minikube/mtls -C ../../../..
```

### Make test requests via Envoy with mTLS

```bash
make curl/example/minikube -C ../../../..
```

### Verify Envoy mTLS stats

```bash
make verify/example/minikube/mtls -C ../../../..
```

### Verify kumactl workflow

```bash
make kumactl/example/minikube -C ../../../..
```

### Undeploy example setup

```bash
make undeploy/example/minikube -C ../../../..
```

### Deploy example setup for traffic routing

```bash
make deploy/traffic-routing/minikube -C ../../../..
```

### Verify traffic routing without mTLS

```bash
make verify/traffic-routing/minikube/without-mtls
```

### Verify traffic routing with mTLS

```bash
make verify/traffic-routing/minikube/with-mtls
```

### Undeploy example setup for traffic routing

```bash
make undeploy/traffic-routing/minikube -C ../../../..
```
