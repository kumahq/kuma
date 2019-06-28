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

## Running locally

Run [KIND](https://kind.sigs.k8s.io/docs/user/quick-start) (Kubernetes IN Docker):

```bash
make start/deps
export KUBECONFIG="$(kind get kubeconfig-path --name=konvoy)"
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
