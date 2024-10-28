# Developer documentation

Hello, and welcome! If you're thinking about contributing to Kuma's code base, you came to the
right place. This document serves as guide/reference for technical
contributors. 

##  Dependencies

The following dependencies are all necessary. Please follow the installation instructions 
for any tools/libraries that you may be missing.

### Command line tool

- [`curl`](https://curl.haxx.se/)  
- [`git`](https://git-scm.com/)
- [`unzip`](http://infozip.sourceforge.net/UnZip.html)
- [`make`](https://www.gnu.org/software/make/)
- [`go`](https://golang.org/)
- [`jq`](https://jqlang.github.io/jq/download/)
- [`yq`](https://mikefarah.gitbook.io/yq)
- [`clang-format`](https://clang.llvm.org/docs/ClangFormat.html) # normally included in
  system's clang package with some exceptions, please check with the following command:
  ```bash
  clang-format --version
  ```

### Helper commands

Throughout this guide, we will use `make` to run pre-defined tasks in the [Makefile](Makefile).
Use the following command to list out all the possible commands:

```bash
make help
```

### Installing dev tools

We packaged the remaining dependencies into one target:

```bash
make dev/tools
```

You can install each commands individually if you prefer.

ATTENTION: By default, development tools will be installed at `$HOME/.kuma-dev/bin`. Remember to include this directory
into your `PATH`, e.g. by adding `export PATH=$HOME/.kuma-dev/bin:$PATH` line to the `$HOME/.bashrc` file or `$HOME/.zshrc` if using zsh.
This can be overridden by setting the env var `CI_TOOLS_DIR`, but it isn't recommended.

## Code checks

To run all code formatting, linting and vetting tools use the target:
```bash
make check
```

## Testing

We use Ginkgo as our testing framework. To run the existing test suite, you have several options:

For all tests, run:
```bash
make test
```
And you can run tests that are specific to a part of Kuma by appending the app name as shown below:
```bash
make test/kumactl
```
For even more specific tests you can specify the package you want to run tests from:
```bash
make test TEST_PKG_LIST=<pkgPath>
```
`pkgPath` is a package list selector for example: `./pkg/xds/...` will run all tests in the `pkg/xds` subtree.

There's a large set of integration tests that can be run with:

```bash
make test/e2e
```

These tests are big and slow, it is recommended to read [e2e-test-tips](docs/guides/e2e-test-tips.md) before running them.

## Building

To build all the binaries run:
```bash
make build
```

Like `make test`, you can append the app name to the target to build a specific binary. For example, here is how you would build the binary for only `kumactl`:
```bash
make build/kumactl
```
This could help expedite your development process if you only made changes to the `kumactl` files.

## Debugging

Like any other go program Kuma can be debugged using [dlv](https://github.com/go-delve/delve).
In this section we'll go into how to trigger a breakpoint both in K8S and Universal.

### K8S

1. Disable k8s leader election (optional)
2. Remove "-w -s" from LDFLAGS [here](https://github.com/kumahq/kuma/blob/7398d8901798d5cf1c2715e036204fc3632ec45d/mk/build.mk#L2) so that debugging symbols are not stripped
3. Always add `EXTRA_GOFLAGS='-gcflags "all=-N -l"'` to build / deploy parameters to make sure debug info is in the binaries
4. Run `make k3d/start`
5. Run `make EXTRA_GOFLAGS='-gcflags "all=-N -l"' -j k3d/deploy/kuma`
6. Change the Kuma deployment:
   1. Remove readiness and Liveness probes (otherwise Kubernetes will kill the container if you stay in a breakpoint long enough)
   2. set runAsNonRoot: false
   3. Double the memory (debugger can make the container OOM)
7. Check go version in `go.mod`, run `kubectl debug --profile=general -n kuma-system -it kuma-control-plane-POD_HASH --image=golang:1.GO_VERSION_FROM_GO_MOD-bookworm --target=control-plane -- bash`
8. Install `dlv` [version that is closes](https://github.com/go-delve/delve/releases) to the `go.mod` version in the container, run: `go install github.com/go-delve/delve/cmd/dlv@vCLOSEST_DLV_VERSION`
9. Run `dlv --listen=:4000 --headless=true --api-version=2 --accept-multiclient attach 1`
10. Setup port forward for `4000`
11. Run goland/vscode debugger with remote target on port `4000`
12. Put a breakpoint where you want
13. Enjoy!

### Universal

1. Add `4000` port in [UniversalApp](https://github.com/kumahq/kuma/blob/201413bdd532e92ff6e1fd017c4970073ba0c09f/test/framework/universal_app.go#L223) so that it's exposed, and the debugger can connect.
2. Remove "-w -s" from LDFLAGS [here](https://github.com/kumahq/kuma/blob/7398d8901798d5cf1c2715e036204fc3632ec45d/mk/build.mk#L2) so that debugging symbols are not stripped
3. Add a `time.Sleep` in a place where you want to debug the test
4. Run the tests `make -j test/e2e/debug EXTRA_GOFLAGS='-gcflags "all=-N -l"' E2E_PKG_LIST=./test/e2e_env/universal/...`
5. Wait to hit the `time.Sleep`
6. Figure out the `kuma-cp` container id by running: `docker ps | grep kuma-cp`
7. Exec into the container: `docker exec -it kuma-3_kuma-cp_3dkYrT bash`
8. Download the same go as in `go.mod` - e.g. `curl -o golang https://dl.google.com/go/go1.23.2.linux-arm64.tar.gz`
9. Extract using `tar xzf golang`
10. Install `dlv` [version that is closes](https://github.com/go-delve/delve/releases) to the `go.mod` version in the container, run: `go install github.com/go-delve/delve/cmd/dlv@vCLOSEST_DLV_VERSION`
11. Run `dlv --listen=:4000 --headless=true --api-version=2 --accept-multiclient attach 1`
12. Figure out the port on the host machine `docker ps | grep kuma-3_kuma-cp_3dkYrT`, look for the port forward for `4000`
13. Run goland/vscode debugger with remote target on port from point 12
14. Enjoy!

## Running

### Kubernetes

Execute
```bash
make -j k3d/restart
```

To stop any existing Kuma K3D cluster, start a new K3D cluster, load images, deploy Kuma and Kuma counter demo. 

### GKE

You can test development versions by first pushing to `gcr.io`:

```bash
gcloud auth configure-docker
make images/push DOCKER_REGISTRY=gcr.io/proj-123456
```

then setting up `kubectl` to connect to your cluster and installing Kuma:

```bash
gcloud container clusters get-credentials cluster-name
kumactl install control-plane --registry=gcr.io/proj-123456 | kubectl apply -f -
```
