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
