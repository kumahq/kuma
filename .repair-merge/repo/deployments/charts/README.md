# Kuma Helm Charts

In addition to the `kumactl install` command, Kuma offers a set of Helm charts to install
Kuma on Kubernetes.

These charts are compatible with Helm 3+.


## Testing the helm chart

You can add some input/output in `app/kumactl/cmd/install/testdata/install-cp-helm`.
You create a `values` file and its matching output this ensures no regression will be introduced with future changes.
