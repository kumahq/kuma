[![][kuma-logo]][kuma-url]

A Helm chart for the Kuma Control Plane

![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![Version: 1.6.0](https://img.shields.io/badge/Version-1.6.0-informational?style=flat-square) ![AppVersion: 1.6.0](https://img.shields.io/badge/AppVersion-1.6.0-informational?style=flat-square)

**Homepage:** <https://github.com/kumahq/kuma>

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| global.image.registry | string | `"docker.io/kumahq"` | Default registry for all Kuma Images |
| global.image.tag | string | `nil` | The default tag for all Kuma images, which itself defaults to .Chart.AppVersion |
| global.imagePullSecrets | list | `[]` | Add `imagePullSecrets` to all the service accounts used for Kuma components |
| patchSystemNamespace | bool | `true` | Whether to patch the target namespace with the system label |
| installCrdsOnUpgrade.enabled | bool | `true` | Whether install new CRDs before upgrade (if any were introduced with the new version of Kuma) |
| installCrdsOnUpgrade.imagePullSecrets | list | `[]` | The `imagePullSecrets` to attach to the Service Account running CRD installation. This field will be deprecated in a future release, please use .global.imagePullSecrets |
| controlPlane.logLevel | string | `"info"` | Kuma CP log level: one of off,info,debug |
| controlPlane.mode | string | `"standalone"` | Kuma CP modes: one of standalone,zone,global |
| controlPlane.zone | string | `nil` | Kuma CP zone, if running multizone |
| controlPlane.kdsGlobalAddress | string | `""` | Only used in `zone` mode |
| controlPlane.replicas | int | `1` | Number of replicas of the Kuma CP. Ignored when autoscaling is enabled |
| controlPlane.autoscaling.enabled | bool | `false` | Whether to enable Horizontal Pod Autoscaling, which requires the [Metrics Server](https://github.com/kubernetes-sigs/metrics-server) in the cluster |
| controlPlane.autoscaling.minReplicas | int | `2` | The minimum CP pods to allow |
| controlPlane.autoscaling.maxReplicas | int | `5` | The max CP pods to scale to |
| controlPlane.autoscaling.targetCPUUtilizationPercentage | int | `80` | For clusters that don't support autoscaling/v2beta, autoscaling/v1 is used |
| controlPlane.autoscaling.metrics | list | `[{"resource":{"name":"cpu","target":{"averageUtilization":80,"type":"Utilization"}},"type":"Resource"}]` | For clusters that do support autoscaling/v2beta, use metrics |
| controlPlane.nodeSelector | object | `{"kubernetes.io/arch":"amd64","kubernetes.io/os":"linux"}` | Node selector for the Kuma Control Plane pods |
| controlPlane.affinity | object | `{}` | Affinity placement rule for the Kuma Control Plane pods |
| controlPlane.injectorFailurePolicy | string | `"Fail"` | Failure policy of the mutating webhook implemented by the Kuma Injector component |
| controlPlane.service.name | string | `nil` | Optionally override of the Kuma Control Plane Service's name |
| controlPlane.service.type | string | `"ClusterIP"` | Service type of the Kuma Control Plane |
| controlPlane.service.annotations | object | `{}` | Additional annotations to put on the Kuma Control Plane |
| controlPlane.ingress.enabled | bool | `false` | Install K8s Ingress resource that exposes GUI and API |
| controlPlane.ingress.ingressClassName | string | `nil` | IngressClass defines which controller will implement the resource |
| controlPlane.ingress.hostname | string | `nil` | Ingress hostname |
| controlPlane.ingress.annotations | object | `{}` | Map of ingress annotations. |
| controlPlane.ingress.path | string | `"/"` | Ingress path. |
| controlPlane.ingress.pathType | string | `"ImplementationSpecific"` | Each path in an Ingress is required to have a corresponding path type. (ImplementationSpecific/Exact/Prefix) |
| controlPlane.globalZoneSyncService | object | `{"annotations":{},"loadBalancerIP":null,"port":5685,"type":"LoadBalancer"}` | URL of Global Kuma CP |
| controlPlane.globalZoneSyncService.type | string | `"LoadBalancer"` | Service type of the Global-zone sync |
| controlPlane.globalZoneSyncService.loadBalancerIP | string | `nil` | Optionally specify IP to be used by cloud provider when configuring load balancer |
| controlPlane.globalZoneSyncService.annotations | object | `{}` | Additional annotations to put on the Global Zone Sync Service |
| controlPlane.globalZoneSyncService.port | int | `5685` | Port on which Global Zone Sync Service is exposed |
| controlPlane.defaults.skipMeshCreation | bool | `false` | Whether to skip creating the default Mesh |
| controlPlane.resources | object | the resources will be chosen based on the mode | Optionally override the resource spec |
| controlPlane.tls.general.secretName | string | `""` | Secret that contains tls.crt, tls.key [and ca.crt when no controlPlane.tls.general.caSecretName specified] for protecting Kuma in-cluster communication |
| controlPlane.tls.general.caSecretName | string | `""` | Secret that contains ca.crt that was used to sign cert for protecting Kuma in-cluster communication (ca.crt present in this secret have precedence over the one provided in the controlPlane.tls.general.secretName) |
| controlPlane.tls.general.caBundle | string | `""` | Base64 encoded CA certificate (the same as in controlPlane.tls.general.secret#ca.crt) |
| controlPlane.tls.apiServer.secretName | string | `""` | Secret that contains tls.crt, tls.key for protecting Kuma API on HTTPS |
| controlPlane.tls.apiServer.clientCertsSecretName | string | `""` | Secret that contains list of .pem certificates that can access admin endpoints of Kuma API on HTTPS |
| controlPlane.tls.kdsGlobalServer.secretName | string | `""` | Secret that contains tls.crt, tls.key for protecting cross cluster communication |
| controlPlane.tls.kdsZoneClient.secretName | string | `""` | Secret that contains ca.crt which was used to sign KDS Global server. Used for CP verification |
| controlPlane.image.pullPolicy | string | `"IfNotPresent"` | Kuma CP ImagePullPolicy |
| controlPlane.image.repository | string | `"kuma-cp"` | Kuma CP image repository |
| controlPlane.image.tag | string | `nil` | Kuma CP Image tag. When not specified, the value is copied from global.tag |
| controlPlane.secrets | list of { Env: string, Secret: string, Key: string } | `nil` | Secrets to add as environment variables, where `Env` is the name of the env variable, `Secret` is the name of the Secret, and `Key` is the key of the Secret value to use |
| controlPlane.envVars | object | `{}` | Additional environment variables that will be passed to the control plane |
| controlPlane.extraConfigMaps | list | `[]` | Additional config maps to mount into the control plane, with optional inline values |
| controlPlane.extraSecrets | list | `[]` | Additional secrets to mount into the control plane |
| controlPlane.webhooks.validator.additionalRules | string | `""` | Additional rules to apply on Kuma validator webhook. Useful when building custom policy on top of Kuma. |
| controlPlane.webhooks.ownerReference.additionalRules | string | `""` | Additional rules to apply on Kuma owner reference webhook. Useful when building custom policy on top of Kuma. |
| controlPlane.hostNetwork | bool | `false` | Specifies if the deployment should be started in hostNetwork mode. |
| controlPlane.podSecurityContext | object | `{}` | Security context at the pod level for control plane. |
| controlPlane.containerSecurityContext | object | `{}` | Security context at the container level for control plane. |
| cni.enabled | bool | `false` | Install Kuma with CNI instead of proxy init container |
| cni.chained | bool | `false` | Install CNI in chained mode |
| cni.netDir | string | `"/etc/cni/multus/net.d"` | Set the CNI install directory |
| cni.binDir | string | `"/var/lib/cni/bin"` | Set the CNI bin directory |
| cni.confName | string | `"kuma-cni.conf"` | Set the CNI configuration name |
| cni.logLevel | string | `"info"` | CNI log level: one of off,info,debug |
| cni.nodeSelector | object | `{"kubernetes.io/arch":"amd64","kubernetes.io/os":"linux"}` | Node Selector for the CNI pods |
| cni.image.registry | string | `"docker.io"` | CNI image registry |
| cni.image.repository | string | `"kumahq/install-cni"` | CNI image repository |
| cni.image.tag | string | `"0.0.9"` | CNI image tag |
| cni.podSecurityContext | object | `{}` | Security context at the pod level for cni |
| cni.containerSecurityContext | object | `{}` | Security context at the container level for cni |
| dataPlane.image.repository | string | `"kuma-dp"` | The Kuma DP image repository |
| dataPlane.image.pullPolicy | string | `"IfNotPresent"` | Kuma DP ImagePullPolicy |
| dataPlane.image.tag | string | `nil` | Kuma DP Image Tag. When not specified, the value is copied from global.tag |
| dataPlane.initImage.repository | string | `"kuma-init"` | The Kuma DP init image repository |
| dataPlane.initImage.tag | string | `nil` | Kuma DP init image tag When not specified, the value is copied from global.tag |
| ingress.enabled | bool | `false` | If true, it deploys Ingress for cross cluster communication |
| ingress.drainTime | string | `"30s"` | Time for which old listener will still be active as draining |
| ingress.replicas | int | `1` | Number of replicas of the Ingress |
| ingress.service.type | string | `"LoadBalancer"` | Service type of the Ingress |
| ingress.service.loadBalancerIP | string | `nil` | Optionally specify IP to be used by cloud provider when configuring load balancer |
| ingress.service.annotations | object | `{}` | Additional annotations to put on the Ingress service |
| ingress.service.port | int | `10001` | Port on which Ingress is exposed |
| ingress.service.nodePort | string | `nil` | Port on which service is exposed on Node for service of type NodePort |
| ingress.annotations | object | `{}` | Additional deployment annotation |
| ingress.nodeSelector | object | `{"kubernetes.io/arch":"amd64","kubernetes.io/os":"linux"}` | Node Selector for the Ingress pods |
| ingress.affinity | object | `{}` | Affinity placement rule for the Kuma Ingress pods |
| ingress.podSecurityContext | object | `{}` | Security context at the pod level for ingress |
| ingress.containerSecurityContext | object | `{}` | Security context at the container level for ingress |
| egress.enabled | bool | `false` | If true, it deploys Egress for cross cluster communication |
| egress.drainTime | string | `"30s"` | Time for which old listener will still be active as draining |
| egress.replicas | int | `1` | Number of replicas of the Egress |
| egress.service.type | string | `"ClusterIP"` | Service type of the Egress |
| egress.service.loadBalancerIP | string | `nil` | Optionally specify IP to be used by cloud provider when configuring load balancer |
| egress.service.annotations | object | `{}` | Additional annotations to put on the Egress service |
| egress.service.port | int | `10002` | Port on which Egress is exposed |
| egress.service.nodePort | string | `nil` | Port on which service is exposed on Node for service of type NodePort |
| egress.annotations | object | `{}` | Additional deployment annotation |
| egress.nodeSelector | object | `{"kubernetes.io/arch":"amd64","kubernetes.io/os":"linux"}` | Node Selector for the Egress pods |
| egress.affinity | object | `{}` | Affinity placement rule for the Kuma Ingress pods |
| egress.podSecurityContext | object | `{}` | Security context at the pod level for egress |
| egress.containerSecurityContext | object | `{}` | Security context at the container level for egress |
| kumactl.image.repository | string | `"kumactl"` | The kumactl image repository |
| kumactl.image.tag | string | `nil` | The kumactl image tag. When not specified, the value is copied from global.tag |
| kubectl.image.registry | string | `"bitnami"` | The kubectl image registry |
| kubectl.image.repository | string | `"kubectl"` | The kubectl image repository |
| kubectl.image.tag | string | `"1.20"` | The kubectl image tag |
| hooks.nodeSelector | object | `{"kubernetes.io/arch":"amd64","kubernetes.io/os":"linux"}` | Node selector for the HELM hooks |
| hooks.podSecurityContext | object | `{}` | Security context at the pod level for crd/webhook/ns |
| hooks.containerSecurityContext | object | `{}` | Security context at the container level for crd/webhook/ns |
| experimental.meshGateway | bool | `false` | If true, it installs experimental built-in Gateway support |
| experimental.gatewayAPI | bool | `false` | If true, it installs experimental Gateway API support |

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

## Autoscaling

In production, it is advisable to enable Control Plane autoscaling for High Availability. Autoscaling uses the
`HorizontalPodAutoscaler` resource to add redundancy and scale the CP pods based on CPU utilization, which requires
the [k8s metrics-server][kube-metrics-server] to be running on the cluster.

## Development

The charts are used internally in `kumactl install`, therefore the following rules apply when developing new chat features:
 * all templates that start with `pre-` and `post-` are omitted when processing in `kumactl install`

### Installing Metrics Server for Autoscaling

If running on kind, or on a cluster with a similarly self-signed cert, the metrics server must be configured to allow
insecure kubelet TLS. The make task `kind/deploy/metrics-server` installs this patched version of the server.

[kuma-url]: https://kuma.io/
[kuma-logo]: https://kuma-public-assets.s3.amazonaws.com/kuma-logo-v2.png
[helm-crd]: https://helm.sh/docs/chart_best_practices/custom_resource_definitions/
[helm-crd-limitations]: https://helm.sh/docs/topics/charts/#limitations-on-crds
[kube-metrics-server]: https://github.com/kubernetes-sigs/metrics-server
