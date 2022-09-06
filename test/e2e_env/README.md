# E2E tests

This package contains E2E tests of Kuma.

We have 3 shared environments each with their own package:
* `kubernetes` - a standalone Kubernetes deployment, deployed using kumactl
* `universal` - a standalone Universal deployment with in-memory DB.
* `multizone` - a multizone deployment with the following topology
  * Global CP on Universal
  * 2 Zone CPs deployed on Kubernetes
  * 2 Zone CPs deployed on Universal
  This setup should cover all scenarios of cross-cluster communication.

All setups come with ZoneEgress by default.

All the test suits (for example `trafficlog`, `trafficpermissions`) are parallelized on multiple processes on one machine using [Ginkgo spec parallelization](https://onsi.github.io/ginkgo/#spec-parallelization). 
Every test suite should run in their own mesh and namespace to not interfere with other tests.


## Recommendations

* (Kubernetes): Use `TriggerDeleteNamespace` instead of `DeleteNamespace`.
  It takes a lot of time to wait for the namespace to be removed.
  We should carry on with other tests while namespace is removed in the background.
* (Kubernetes): Be conservative about resource limits on Kubernetes.
  We are deploying a lot of stuff in parallel, default `DemoClient` and `TestServer` are tuned to have minimal requests/limit.
  If you are bringing new deployment to the test, make sure you set a proper resource limits.
* `KUMA_STORE_UNSAFE_DELETE` is set to true by default, so we can remove the mesh without waiting for all DPPs to be down.
* (Multizone) Remember about calling `WaitForMesh`, otherwise deploying an application in Zone can fail.

## Limitations

* (Universal) Getting logs in the test is currently not supported. https://github.com/kumahq/kuma/issues/4187
* (Universal) UniversalCluster#apps is not shared across parallelized processes.
