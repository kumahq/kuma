# Developer documentation

Hello, and welcome! If you're thinking about contributing to Kuma's code base, you came to the
right place. This document serves as guide/reference for technical
contributors. 

Consult the Table of Contents below, and jump to the desired section.

## Table of Contents

- [Developer documentation](#developer-documentation)
  - [Table of Contents](#table-of-contents)
  - [Dependencies](#dependencies)
    - [Command line tool](#command-line-tool)
    - [Helper commands](#helper-commands)
    - [Installing dev tools](#installing-dev-tools)
  - [Testing](#testing)
  - [Building](#building)
  - [Running Control Plane on local machine](#running-control-plane-on-local-machine)
    - [Universal](#universal)
      - [Universal without any external dependency (in-memory storage)](#universal-without-any-external-dependency-in-memory-storage)
      - [Universal with Postgres as a storage](#universal-with-postgres-as-a-storage)
      - [Running a Dataplane (Envoy)](#running-a-dataplane-envoy)
  - [Kubernetes](#kubernetes)
    - [Local Control Plane working with Kubernetes](#local-control-plane-working-with-kubernetes)
      - [Running Dataplane (Envoy)](#running-dataplane-envoy)
    - [Control Plane on Kubernetes](#control-plane-on-kubernetes)

##  Dependencies

The following dependencies are all necessary. Please follow the installation instructions 
for any tools/libraries that you may be missing.

### Command line tool

- [`curl`](https://curl.haxx.se/)  
- [`git`](https://git-scm.com/) 
- [`make`](https://www.gnu.org/software/make/)
- [`go`](https://golang.org/)
- [`clang-format`](https://clang.llvm.org/docs/ClangFormat.html) # normally included in
  system's clang package with some exceptions, please check with the following command:
  ```bash
  clang-format --version
  ```

For a quick start, use the official `golang` Docker image. It includes all the command line tools pre-installed, with exception for from [`clang-format`](https://clang.llvm.org/docs/ClangFormat.html). 

```bash
docker run --name kuma-build -ti \
  --volume `pwd`:/go/src/github.com/kumahq/kuma \
  --workdir /go/src/github.com/kumahq/kuma \
  --env GO111MODULE=on \
  golang:1.12.12 \
  bash -c 'apt update && apt install -y unzip && export PATH=$HOME/bin:$PATH && bash'
```

If `kuma-build` container already exists, start it with

```bash
docker start --attach --interactive kuma-build
```

To cleanup disk space, run

```bash
docker rm kuma-build
```

### Helper commands

Throughout this guide, we will use the `make` utility to run pre-defined
tasks in the [Makefile](Makefile). Use the following command to list out all the possible commands:

```bash
make help
```

### Installing dev tools

We packaged the remaining dependencies into one command. This is one of many commands
that's listed if you run the `make help` command above. You can also install each tool
individually if you know what you are missing. Run:

```bash
make dev/tools
```

to install all of the following tools at `$HOME/bin`:

1. [Ginkgo](https://github.com/onsi/ginkgo#set-me-up) (BDD-style Go testing framework)
2. [Kubebuilder](https://book.kubebuilder.io/quick-start.html#installation) (Kubernetes API extension framework, comes with `etcd` and `kube-apiserver`)
3. [kustomize](https://book.kubebuilder.io/quick-start.html#installation) (Customization of kubernetes YAML configurations)
4. [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/#install-kubectl-binary-with-curl-on-linux) (Kubernetes API client)
5. [KIND](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) (Kubernetes IN Docker)
6. [Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/#linux) (Kubernetes in VM)

ATTENTION: By default, development tools will be installed at `$HOME/bin`. Remember to include this directory into your `PATH`, 
e.g. by adding `export PATH=$HOME/bin:$PATH` line to the `$HOME/.bashrc` file.

## Testing

We use Ginkgo as our testing framework. To run the existing test suite, you have several options:

For all tests, run:
```bash
make test
```
For integration tests, run:
```bash
make integration
```
And you can run tests that are specific to a part of Kuma by appending the app name as shown below:
```bash
make test/kumactl
```
Please make sure that all tests pass before submitting a pull request in the `master` branch. Thank you!


## Building

After you made some changes, you will need to re-build the binaries. You can build all binaries by running:
```bash
make build
```
And similar to `make test`, you can appending the app name to the build command to build the binary of a specific app. For example, here is how you would build the binary for only `kumactl`:
```bash
make build/kumactl
```
This could help expedite your development process if you only made changes to the `kumactl` files.

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
export PATH=`pwd`/build/artifacts-$(go env GOOS)-$(go env GOARCH)/kumactl:$PATH
```

2. Configure a `kumactl` with running Control Plane

Check current config
```bash
kumactl config control-planes list

ACTIVE   NAME    ADDRESS
*        local   http://localhost:5681
```

If active control plane points to different address than http://localhost:5681, add a new one 
```bash
kumactl config control-planes add --name universal --address http://localhost:5681
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


## Kubernetes

### Local Control Plane working with Kubernetes

The following instructions are for folks who want to run `Control Plane` on their local machine. Local development is a viable path for people who are familiar with Kubernetes and understands trade-offs of out-off cluster deployment. If you want to run the `Control Plane` on Kubernetes instead, jump to the [Control Plane on Kubernetes](#Control-Plane-on-Kubernetes) section below.

1. Run [KIND](https://kind.sigs.k8s.io/docs/user/quick-start) (Kubernetes IN Docker):

```bash
make kind/start

# set KUBECONFIG for use by `kumactl` and `kubectl`
export KUBECONFIG="$(kind get kubeconfig-path --name=kuma)"
```

2. Run `Control Plane` on local machine:

```bash
make run/k8s
```

#### Running Dataplane (Envoy)

1. Run the instructions at "Pointing Envoy at Control Plane" section, so the Dataplane descriptor is available in the Control Plane.

2. Start `Envoy` on local machine (requires `envoy` binary to be on your `PATH`):

```bash
make run/example/envoy/k8s
```

3. Dump effective `Envoy` config:

```bash
make config_dump/example/envoy
```

### Control Plane on Kubernetes

To run the Kuma `Control Plane` on Kubernetes, follow these steps:

1. Run [KIND](https://kind.sigs.k8s.io/docs/user/quick-start) (Kubernetes IN Docker):

```bash
make kind/start

# set KUBECONFIG for use by `kubectl`
export KUBECONFIG="$(kind get kubeconfig-path --name=kuma)"
```

2. Deploy `Control Plane` to [KIND](https://kind.sigs.k8s.io/docs/user/quick-start) (Kubernetes IN Docker):

```bash
make kind/deploy/kuma
```

3. Deploy the demo app (and get Kuma sidecar injected)

```bash
make kind/deploy/example-app
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
kumactl config control-planes add --name k8s --address http://localhost:15681
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
default   demo-app-7cbbd658d5-dj9l6   app=demo-app pod-template-hash=7cbbd658d5 service=demo-app_kuma-demo_svc_80   Online   42m28s               42m28s             8               0
```
