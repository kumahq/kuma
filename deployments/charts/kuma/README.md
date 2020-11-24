[![][kuma-logo]][kuma-url]

A Helm chart for the Kuma Control Plane

![Version: 0.4.1](https://img.shields.io/badge/Version-0.4.1-informational?style=flat-square)
![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square)
![AppVersion: 1.0.1](https://img.shields.io/badge/AppVersion-1.0.1-informational?style=flat-square)

**Homepage:** <https://github.com/kumahq/kuma>

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| cni.binDir | string | `"/var/lib/cni/bin"` |  |
| cni.chained | bool | `false` |  |
| cni.confName | string | `"kuma-cni.conf"` |  |
| cni.enabled | bool | `false` |  |
| cni.image.registry | string | `"docker.io"` |  |
| cni.image.repository | string | `"lobkovilya/install-cni"` | Docker image name for the cni instance |
| cni.image.tag | string | `"0.0.2"` | Docker tag for the cni instance |
| cni.logLevel | string | `"info"` |  |
| cni.netDir | string | `"/etc/cni/multus/net.d"` |  |
| cni.nodeSelector."kubernetes.io/arch" | string | `"amd64"` |  |
| cni.nodeSelector."kubernetes.io/os" | string | `"linux"` |  |
| controlPlane.defaults.skipMeshCreation | bool | `false` |  |
| controlPlane.envVars | object | `{}` |  |
| controlPlane.globalRemoteSyncService.annotations | object | `{}` |  |
| controlPlane.globalRemoteSyncService.port | int | `5685` |  |
| controlPlane.globalRemoteSyncService.type | string | `"LoadBalancer"` |  |
| controlPlane.image.pullPolicy | string | `"IfNotPresent"` |  |
| controlPlane.image.repository | string | `"kuma-cp"` |  |
| controlPlane.injectorFailurePolicy | string | `"Ignore"` |  |
| controlPlane.kdsGlobalAddress | string | `""` |  |
| controlPlane.logLevel | string | `"info"` |  |
| controlPlane.mode | string | `"standalone"` |  |
| controlPlane.nodeSelector."kubernetes.io/arch" | string | `"amd64"` |  |
| controlPlane.nodeSelector."kubernetes.io/os" | string | `"linux"` |  |
| controlPlane.service.annotations | object | `{}` |  |
| controlPlane.service.type | string | `"ClusterIP"` |  |
| controlPlane.tls.apiServer.clientCertsSecretName | string | `""` |  |
| controlPlane.tls.apiServer.secretName | string | `""` |  |
| controlPlane.tls.general.caBundle | string | `""` |  |
| controlPlane.tls.general.secretName | string | `""` |  |
| controlPlane.tls.kdsGlobalServer.secretName | string | `""` |  |
| controlPlane.tls.kdsRemoteClient.secretName | string | `""` |  |
| dataPlane.image.pullPolicy | string | `"IfNotPresent"` |  |
| dataPlane.image.repository | string | `"kuma-dp"` |  |
| dataPlane.initImage.repository | string | `"kuma-init"` |  |
| global.image.registry | string | `"kong-docker-kuma-docker.bintray.io"` | Default registry for all Kuma Images |
| global.namespace | string | `"kuma-system"` |  |
| ingress.drainTime | string | `"30s"` |  |
| ingress.enabled | bool | `false` |  |
| ingress.mesh | string | `"default"` |  |
| ingress.service.annotations | object | `{}` |  |
| ingress.service.port | int | `10001` |  |
| ingress.service.type | string | `"LoadBalancer"` |  |
| patchSystemNamespace | bool | `true` |  |

## Custom Resource Definitions

All Kuma CRDs are loaded via the [`crds`](crds) directory. For more detailed information on CRDs and Helm,
please refer to [the Helm documentation][helm-crd].

## Deleting

As part of [Helm's limitations][helm-crd-limitations], CRDs will not be deleted when the `kuma` chart is deleted and
must be deleted manually. When a CRD is deleted Kubernetes deletes all resources of that kind as well, so this should
be done carefully.

To do this with `kubectl` on *nix platforms, run:

```shell
kubectl get crds | grep kuma.io | tr -s " " | cut -d " " -f1 | xargs kubectl delete crd

# or with jq
kubectl get crds -o json | jq '.items[].metadata.name | select(.|test(".*kuma\\.io"))' | xargs kubectl delete crd
```

## Note to Chart developers

The charts are used internally in `kumactl install`, therefore the following rules apply when developing new chat features:
 * use `make generate/kumactl/install/k8s/control-plane` to sync the Helm Chart and `kumactl install` templates
 * all templates that start with `pre-` and `post-` are omitted when processing in `kumactl install`

[kuma-url]: https://kuma.io/
[kuma-logo]: https://kuma-public-assets.s3.amazonaws.com/kuma-logo-v2.png
[helm-crd]: https://helm.sh/docs/chart_best_practices/custom_resource_definitions/
[helm-crd-limitations]: https://helm.sh/docs/topics/charts/#limitations-on-crds