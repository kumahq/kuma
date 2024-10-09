# E2E tests tips

E2E tests are less stable and slower than unit tests therefore it's useful to know some tricks

## Using a VM

If you are developing on a laptop e2e tests might be almost impossible to run.
Most core devs use a VM to run e2e tests quickly ([Hetzner](https://hetzner.com) and [Scaleway](https://scaleway.com) have very cheap ones).

If you are doing small patches relying on PR checks might be enough.

## Faster environment setup

When e2e tests are executed here are the steps

1) Build containers
2) Start Kubernetes Cluster 1
3) Start Kubernetes Cluster 2
4) Execute tests

## Use KIND instead of K3D

K3D is faster than KIND and it is a default tool to run E2E tests. However, not all tests run with it.
To use KIND in E2E tests add `K8S_CLUSTER_TOOL=kind`

```bash
make test/e2e K8S_CLUSTER_TOOL=kind
```

## Execute single test

Regular `make test/e2e` will execute all the tests in the project.

If you want to execute single test you can change the code from `It()` to `FIt()` or `Describe()` to `FDescribe()` and then use `E2E_PKG_LIST`

Example:
```bash
FIt("should access service locally and remotely", func() {
...

make test/e2e E2E_PKG_LIST=./test/e2e/deploy/... 
```

## Running e2e tests one at a time

When you run `make test/e2e`, the Docker infrastructure will get torn down if any tests fail.
In the case of failing tests, it can be useful to run one suite at a time, while leaving the Docker containers in place.

To do this, first create the test environment:
```bash
make images test/e2e/k8s/start
```

Now you can run each test suite (exiting on any failures):
```bash
(
    set -e
    for t in $(go list ./test/e2e/...); do
        make test/e2e/test E2E_PKG_LIST="$t"
    done
)
```

## Debug the test - do not clean up environment

Even if you execute one test and it fails, our framework clean up the environment (delete Kubernetes clusters, containers etc.)

When you use `make test/e2e/debug` and the test fails, execution will immediately stop and environment is not cleaned up.

`test/e2e/debug` works with the same envs as `test/e2e` like `E2E_PKG_LIST`

## Decide which Kubernetes clusters will be created

If you know you are running only universal tests you can skip creating Kubernetes clusters by setting
```bash
make test/e2e/debug K8SCLUSTERS= E2E_PKG_LIST=./test/e2e/trafficpermission/universal/...
```

or if the test need only 1 Kuberenetes cluster do this

```bash
make test/e2e/debug K8SCLUSTERS=kuma-1 E2E_PKG_LIST=./test/e2e/trafficpermission/universal/...
```

## Useful commands

### Cleanup environment

Running `make test/e2e/debug` can intentionally leave resources if test fails. Clean them up with   

```bash
make k3d/stop/all && docker stop $(docker ps -aq) # omit $ for fish
```

### Tests failing because of disk pressure

From time to time when running tests your docker environment can run out of space. You will see k8s events like this:
```bash
Warning  FailedScheduling  63s (x1 over 2m15s)  default-scheduler  0/1 nodes are available: 
1 node(s) had taint {node.kubernetes.io/disk-pressure:}, that the pod didn't tolerate.
```

To fix this issue you need to clean up your docker environment:

```bash
docker system prune --volumes --all
```

### Integration with direnv

[direnv](https://direnv.net/) is a useful tool that can populate environment variables in your shell as you change directories.
The Kuma build has an optional `dev/envrc` target that generates a `.envrc` file to set the `$CI_TOOLS_DIR` and `$KUBECONFIG` environment variables.
This is useful to keeping the Kuma CI tools installation tidy in your Kuma workspace, and for conveniently accessing the Kind clusters that are provisioned by the e2e tests.

```bash
$ make dev/envrc
direnv: loading ~/upstream/konghq/kuma/.envrc
direnv: export +CI_TOOLS_DIR +KUBECONFIG
```
