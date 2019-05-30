# Konvoy Control Plane

Universal Control Plane for Envoy-based Service Mesh.

## Building

Run:

```bash
make build
```

## Running locally

Run:

```bash
make run
```

Make a test `Discovery` request to `LDS`:

```bash
make curl/listeners
```

Make a test `Discovery` request to `CDS`:

```bash
make curl/clusters
```

## Pointing Envoy at Control Plane

Start `Control Plane`:

```bash
make run
```

Assuming `envoy` binary is on your `PATH`, run:

```bash
make run/example/envoy
```

Dump effective Envoy config:

```bash
make config_dump/example/envoy
```

## Running demo setup inside Docker Compose

Start example setup (`Control Plane` + `Envoy` + app):

```bash
make run/example/docker-compose
```

Make test requests (`Envoy` must intercept both `inbound` and `outbound` connections):

```bash
make curl/example/docker-compose
```

Observe `Envoy` stats:

```bash
make stats/example/docker-compose
```

E.g.,
```
cluster.ads_cluster.upstream_rq_total: 1
cluster.localhost_8080.upstream_rq_total: 7
cluster.pass_through.upstream_rq_total: 7
```

where

* `cluster.localhost_8080.upstream_rq_total` is a number of `inbound` requests
* `cluster.pass_through.upstream_rq_total` is a number of `outbound` requests

## Running demo setup inside Minikube

Follow instructions in [Minikube](examples/minikube/README.md) example.
