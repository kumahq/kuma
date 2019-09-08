# Developer documentation

## Pre-requirements

- `curl`
- `git`
- `go`
- `make`

For a quick start, use the official `golang` Docker image (which has all these tools pre-installed), e.g.

```bash
docker run --rm -ti \
  --user 65534:65534 \
  --volume `pwd`:/go/src/github.com/Kong/konvoy/components/konvoy-control-plane \
  --workdir /go/src/github.com/Kong/konvoy/components/konvoy-control-plane \
  --env HOME=/tmp/home \
  --env GO111MODULE=on \
  golang:1.12.5 bash
export PATH=$HOME/bin:$PATH
```

## Helper commands

```bash
make help
```

## Installing dev tools

Run:

```bash
make dev/tools
```

which will install the following tools at `$HOME/bin`:

1. [Ginkgo](https://github.com/onsi/ginkgo#set-me-up) (BDD testing framework)
2. [Kubebuilder](https://book.kubebuilder.io/quick-start.html#installation) (Kubernetes API extension framework, comes with `etcd` and `kube-apiserver`)
3. [kustomize](https://book.kubebuilder.io/quick-start.html#installation) (Customization of kubernetes YAML configurations)
4. [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/#install-kubectl-binary-with-curl-on-linux) (Kubernetes API client)
5. [KIND](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) (Kubernetes IN Docker)
6. [Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/#linux) (Kubernetes in VM)

ATTENTION: By default, development tools will be installed at `$HOME/bin`. Remember to include this directory into your `PATH`, 
e.g. by adding `export PATH=$HOME/bin:$PATH` line to the `$HOME/.bashrc` file.

## Building

Run:

```bash
make build
```

## Integration tests

Integration tests will run all dependencies (ex. Postgres). Run:

 ```bash
make integration
```

## Running Control Plane on local machine

### Universal

Universal setup does not require Kubernetes to be present. It can be run in two modes.

#### Universal without any external dependency (in-memory storage)

1. Run `Control Plane` on local machine:

```bash
make run/universal/memory
```

#### Universal with Postgres as a storage

1. Run Postgres with initial schema using docker-compose.
It will run on port 15432 with username: `kuma`, password: `kuma` and db name: `kuma`.

```bash
make start/postgres
```

2. Run `Control Plane` on local machine.

```bash
make run/universal/postgres
```

#### Running a Dataplane (Envoy)

1. Build a `kumactl`
```bash
make build/kumactl
export PATH=`pwd`/build/artifacts/kumactl:$PATH
```

2. Configure a `kumactl` with running Control Plane

```bash
kumactl config control-planes add --name universal --api-server-url http://localhost:5681
```

3. Apply a Dataplane descriptor

```bash
kumactl apply -f dev/examples/universal/dataplanes/example.yaml
```

4. Run a Dataplane (requires `envoy` binary to be on your `PATH`)
```bash
make run/example/envoy/universal
```

5. List `Dataplanes` connected to the `Control Plane`:

```bash
kumactl inspect dataplanes

MESH      NAME      TAGS                                     STATUS   LAST CONNECTED AGO   LAST UPDATED AGO   TOTAL UPDATES   TOTAL ERRORS
default   example   env=production service=web version=2.0   Online   32s                  32s                2               0
```

6. Dump effective `Envoy` config:

```bash
make config_dump/example/envoy
```


### Kubernetes

1. Run [KIND](https://kind.sigs.k8s.io/docs/user/quick-start) (Kubernetes IN Docker):

```bash
make start/k8s

# set KUBECONFIG for use by `kumactl` and `kubectl`
export KUBECONFIG="$(kind get kubeconfig-path --name=kuma)"
```

2. Run `Control Plane` on local machine:

```bash
make run/k8s
```

#### Running Dataplane (Envoy)

1. Run the instructions at "Pointing Envoy at Control Plane" section, so the `kuma-injector` is present in the KIND
cluster and Dataplane descriptor is available in the Control Plane.

2. Start `Envoy` on local machine (requires `envoy` binary to be on your `PATH`):

```bash
make run/example/envoy/k8s
```

3. Dump effective `Envoy` config:

```bash
make config_dump/example/envoy
```

## Running Control Plane on Kubernetes

1. Run [KIND](https://kind.sigs.k8s.io/docs/user/quick-start) (Kubernetes IN Docker):

```bash
make start/k8s

# set KUBECONFIG for use by `kubectl`
export KUBECONFIG="$(kind get kubeconfig-path --name=kuma)"
```

2. Deploy `Control Plane` to [KIND](https://kind.sigs.k8s.io/docs/user/quick-start) (Kubernetes IN Docker):

```bash
make start/control-plane/k8s
```

3. Redeploy demo app (to get Kuma sidecar injected)

```bash
kubectl delete -n kuma-demo pod -l app=demo-app
```

4. Build `kumactl`

```bash
make build/kumactl

export PATH=`pwd`/build/artifacts/kumactl:$PATH
```

5. Forward the `Control Plane` port to `localhost`:
```bash
kubectl port-forward -n kuma-system $(kubectl get pods -n kuma-system -l app=kuma-control-plane -o=jsonpath='{.items[0].metadata.name}') 15681:5681
```

6. Add `Control Plane` to your `kumactl` config:

```bash
kumactl config control-planes add --name k8s --api-server-url http://localhost:15681
```

7. Verify that `Control Plane` has been added:

```bash
kumactl config control-planes list

NAME   API SERVER
k8s    http://localhost:15681
```

8. List `Dataplanes` connected to the `Control Plane`:

```bash
kumactl inspect dataplanes

MESH      NAME                        TAGS                                                                          STATUS   LAST CONNECTED AGO   LAST UPDATED AGO   TOTAL UPDATES   TOTAL ERRORS
default   demo-app-7cbbd658d5-dj9l6   app=demo-app pod-template-hash=7cbbd658d5 service=demo-app.kuma-demo.svc:80   Online   42m28s               42m28s             8               0
```
