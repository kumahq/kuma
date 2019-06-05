# Developer documentation

## Pre-requirements

1. Install [KIND](https://kind.sigs.k8s.io/docs/user/quick-start) (Kubernetes IN Docker)

## Building

Run:

```bash
make build
```

## Running locally

Run [KIND](https://kind.sigs.k8s.io/docs/user/quick-start) (Kubernetes IN Docker):

```bash
kind create cluster --name konvoy
export KUBECONFIG="$(kind get kubeconfig-path --name="konvoy")"
```

Run Control Plane:

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
