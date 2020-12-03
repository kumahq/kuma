[![][kuma-logo]][kuma-url]

# kuma

<<<<<<< HEAD
The kuma chart supports all the features and options provided by `kumactl install control-plane`.
The chart supports Helm v3+.
=======
![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![Version: 0.4.4](https://img.shields.io/badge/Version-0.4.4-informational?style=flat-square) ![AppVersion: 1.0.3](https://img.shields.io/badge/AppVersion-1.0.3-informational?style=flat-square)

**Homepage:** <https://github.com/kumahq/kuma>
>>>>>>> 86cb20c6... chore(*) helm chart 0.4.4 (#1277)

## Values

| Parameter                                          | Description                                                                                         | Default                                                  |
|----------------------------------------------------|-----------------------------------------------------------------------------------------------------|----------------------------------------------------------|
| `global.image.registry`                            | Default registry for all Kuma images                                                                | `kong-docker-kuma-docker.bintray.io`                     |
| `global.image.tag`                                 | Default tag for all Kuma images                                                                     | nil, defaults to Chart.AppVersion                        |
| `patchSystemNamespace`                             | Patch the release namespace with the Kuma system label                                              | `true`                                                   |
| `controlPlane.logLevel`                            | Kuma CP log level: one of off\|info\|debug                                                          | `info`                                                   |
| `controlPlane.mode`                                | Kuma CP modes: one of standalone\|remote\|global                                                    | `standalone`                                             |
| `controlPlane.zone`                                | Kuma zone name                                                                                      | nil                                                      |
| `controlPlane.kdsGlobalAddress`                    | URL of Global Kuma CP                                                                               |                                                          |
| `controlPlane.nodeSelector`                        | Node Selector for the Kuma Control Plane pods                                                       | `{ kubernetes.io/os: linux, kubernetes.io/arch: amd64 }` |
| `controlPlane.replicas`                            | Number of replicas of the Kuma CP                                                                   | 1                                                        |
| `controlPlane.injectorFailurePolicy`               | Failure policy of the mutating webhook implemented by the Kuma Injector component                   | `Ignore`                                                 |
| `controlPlane.service.name`                        | Service name of the Kuma Control Plane                                                              | nil                                                      |
| `controlPlane.service.type`                        | Service type of the Kuma Control Plane                                                              | ClusterIP                                                |
| `controlPlane.service.annotations`                 | Additional annotations to put on the Kuma Control Plane service                                     | {}                                                       |
| `controlPlane.globalRemoteSyncService.name`        | Service name of the Global-Remote Sync                                                              | nil                                                      |
| `controlPlane.globalRemoteSyncService.type`        | Service type of the Global-Remote Sync                                                              | LoadBalancer                                             |
| `controlPlane.globalRemoteSyncService.port`        | Port on which Global-Remote Sync is exposed                                                         | 5685                                                     |
| `controlPlane.globalRemoteSyncService.annotations` | Additional annotations to put on the Global-Remote Sync service                                     | {}                                                       |
| `controlPlane.defaults.skipMeshCreation`           | Whether or not to skip creating the default Mesh                                                    | `true`                                                   |
| `controlPlane.resources`                           | The K8s resources spec for Kuma CP                                                                  | nil, differs based on mode                               |
| `controlPlane.tls.general.secretName`              | Secret that contains tls.crt, key.crt and ca.crt for protecting Kuma in-cluster communication       | nil, generated and self-signed                           |
| `controlPlane.tls.general.caBundle`                | Base64 encoded CA certificate (the same as in controlPlane.tls.general.secret#ca.crt)               | nil, generated and self-signed                           |
| `controlPlane.tls.apiServer.secretName`            | Secret that contains tls.crt, key.crt for protecting Kuma API on HTTPS                              | nil, autoconfigured from tls.general.secret              |
| `controlPlane.tls.apiServer.clientCertsSecretName` | Secret that contains list of .pem certificates that can access admin endpoints of Kuma API on HTTPS | nil                                                      |
| `controlPlane.tls.kdsGlobalServer.secretName`      | Secret that contains tls.crt, key.crt for protecting cross cluster communication                    | nil, autoconfigured from tls.general.secret              |
| `controlPlane.tls.kdsRemoteClient.secretName`      | Secret that contains ca.crt which was used to sign KDS Global server. Used for CP verification      | nil                                                      |
| `controlPlane.image.pullPolicy`                    | Kuma CP ImagePullPolicy                                                                             | `IfNotPresent`                                           |
| `controlPlane.image.registry`                      | Kuma CP image registry                                                                              | nil, uses global                                         |
| `controlPlane.image.repository`                    | Kuma CP image repository                                                                            | `kuma-cp`                                                |
| `controlPlane.image.tag`                           | Kuma CP image tag                                                                                   | nil, uses global                                         |
| `controlPlane.envVars`                             | Additional environment variables that will be passed to the control plane                           | {}                                                       |
| `controlPlane.config`                              | Config overrides for Kuma CP (YAML encoded as string)                                               |                                                          |
| `cni.enabled`                                      | Install Kuma with CNI instead of proxy init container                                               | `false`                                                  |
| `cni.chained`                                      | Install CNI in chained mode                                                                         | `false`                                                  |
| `cni.netDir`                                       | Set the CNI install directory                                                                       | `/etc/cni/multus/net.d`                                  |
| `cni.binDir`                                       | Set the CNI binary directory                                                                        | `/var/lib/cni/bin`                                       |
| `cni.confName`                                     | Set the CNI configuration name                                                                      | `kuma-cni.conf`                                          |
| `cni.logLevel`                                     | CNI log level: one of off\|info\|debug                                                              | `info`                                                   |
| `cni.nodeSelector`                                 | Node Selector for the CNI pods                                                                      | `{ kubernetes.io/os: linux, kubernetes.io/arch: amd64 }` |
| `cni.image.registry`                               | CNI image registry                                                                                  | `docker.io`                                              |
| `cni.image.repository`                             | CNI image repository                                                                                | `lobkovilya/install-cni`                                 |
| `cni.image.tag`                                    | The CNI image tag                                                                                   | `0.0.2`                                                  |
| `dataPlane.image.registry`                         | The Kuma DP image registry                                                                          | nil, uses global                                         |
| `dataPlane.image.repository`                       | The Kuma DP image repository                                                                        | `kuma-cp`                                                |
| `dataPlane.image.tag`                              | The Kuma DP image tag                                                                               | nil, uses global                                         |
| `dataPlane.initImage.registry`                     | The Kuma DP init image registry                                                                     | nil, uses global                                         |
| `dataPlane.initImage.repository`                   | The Kuma DP init image repository                                                                   | `kuma-init`                                              |
| `dataPlane.initImage.tag`                          | The Kuma DP init image tag                                                                          | nil, uses global                                         |
| `ingress.enabled`                                  | If true, it deploys Ingress for cross cluster communication                                         | false                                                    |
| `ingress.replicas`                                 | Number of replicas of the Ingress                                                                   | 1                                                        |
| `ingress.drainTime`                                | Time for which old listener will still be active as draining                                        | 30s                                                      |
| `ingress.service.name`                             | Service name of the Ingress                                                                         | nil                                                      |
| `ingress.service.type`                             | Service type of the Ingress                                                                         | LoadBalancer                                             |
| `ingress.service.port`                             | Port on which Ingress is exposed                                                                    | 10001                                                    |
| `ingress.service.annotations`                      | Additional annotations to put on the Ingress service                                                | {}                                                       |
| `ingress.mesh`                                     | Mesh to which Dataplane Ingress belongs to                                                          | default                                                  |

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
