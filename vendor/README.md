Notice that a tiny subset of `github.com/prometheus/prometheus` has been "vendored"
in order to avoid an issue with transitive dependencies.

We only need a dependency on `github.com/prometheus/prometheus/documentation/examples/custom-sd/adapter`.

However, it depends on `github.com/prometheus/prometheus/discovery`, which brings all `sd` implementations,
which results in a dependency nightmare (e.g., version of `https://github.com/Azure/go-autorest` conflicts with `k8s.io/client-go`).

That's why we opted for "vendoring" + manual editing to avoid transitive dependencies.
