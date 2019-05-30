Konvoy Control Plane inside Minikube
====================

## Usage

### Start Minikube

```bash
minikube start
```

### Build Control Plane image

```bash
# point at Docker daemon inside Minikube
eval $(minikube docker-env)

# build Docker image with Control Plane
make image -C ../..
```

### Deploy demo setup into Minikube

```bash
kubectl apply -f konvoy-demo.yaml
```

### Make test requests

```bash
kubectl -n konvoy-demo run --rm -it busybox --image=busybox --restart=Never -- sh -c 'while true ; do wget -qO- demo-app:8000/request ; sleep 1 ; done'
```

### Observe Envoy stats

```bash
kubectl -n konvoy-demo exec $(kubectl -n konvoy-demo get pods -l app=demo-app -o=jsonpath='{.items[0].metadata.name}') -c envoy-sidecar -- wget -qO- http://localhost:9901/stats | grep upstream_rq_total
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
