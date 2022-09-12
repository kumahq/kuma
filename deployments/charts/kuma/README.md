[![][kuma-logo]][kuma-url]

A Helm chart for the Kuma Control Plane

![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![Version: 1.8.0](https://img.shields.io/badge/Version-1.8.0-informational?style=flat-square) ![AppVersion: 1.8.0](https://img.shields.io/badge/AppVersion-1.8.0-informational?style=flat-square)

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
| controlPlane.extraLabels | object | `{}` | Labels to add to resources in addition to default labels |
| controlPlane.logLevel | string | `"info"` | Kuma CP log level: one of off,info,debug |
| controlPlane.mode | string | `"standalone"` | Kuma CP modes: one of standalone,zone,global |
| controlPlane.zone | string | `nil` | Kuma CP zone, if running multizone |
| controlPlane.kdsGlobalAddress | string | `""` | Only used in `zone` mode |
| controlPlane.replicas | int | `1` | Number of replicas of the Kuma CP. Ignored when autoscaling is enabled |
| controlPlane.podAnnotations | object | `{}` | Control Plane Pod Annotations |
| controlPlane.autoscaling.enabled | bool | `false` | Whether to enable Horizontal Pod Autoscaling, which requires the [Metrics Server](https://github.com/kubernetes-sigs/metrics-server) in the cluster |
| controlPlane.autoscaling.minReplicas | int | `2` | The minimum CP pods to allow |
| controlPlane.autoscaling.maxReplicas | int | `5` | The max CP pods to scale to |
| controlPlane.autoscaling.targetCPUUtilizationPercentage | int | `80` | For clusters that don't support autoscaling/v2beta, autoscaling/v1 is used |
| controlPlane.autoscaling.metrics | list | `[{"resource":{"name":"cpu","target":{"averageUtilization":80,"type":"Utilization"}},"type":"Resource"}]` | For clusters that do support autoscaling/v2beta, use metrics |
| controlPlane.nodeSelector | object | `{"kubernetes.io/os":"linux"}` | Node selector for the Kuma Control Plane pods |
| controlPlane.podDisruptionBudget.enabled | bool | `false` | Whether to create a pod disruption budget |
| controlPlane.podDisruptionBudget.maxUnavailable | int | `1` | The maximum number of unavailable pods allowed by the budget |
| controlPlane.affinity | object | `{"podAntiAffinity":{"preferredDuringSchedulingIgnoredDuringExecution":[{"podAffinityTerm":{"labelSelector":{"matchExpressions":[{"key":"app.kubernetes.io/name","operator":"In","values":["{{ include \"kuma.name\" . }}"]},{"key":"app.kubernetes.io/instance","operator":"In","values":["{{ .Release.Name }}"]},{"key":"app","operator":"In","values":["{{ include \"kuma.name\" . }}-control-plane"]}]},"topologyKey":"kubernetes.io/hostname"},"weight":100}]}}` | Affinity placement rule for the Kuma Control Plane pods. This is rendered as a template, so you can reference other helm variables or includes. |
| controlPlane.topologySpreadConstraints | string | `nil` | Topology spread constraints rule for the Kuma Control Plane pods. This is rendered as a template, so you can use variables to generate match labels. |
| controlPlane.injectorFailurePolicy | string | `"Fail"` | Failure policy of the mutating webhook implemented by the Kuma Injector component |
| controlPlane.service.enabled | bool | `true` | Whether to create a service resource. |
| controlPlane.service.name | string | `nil` | Optionally override of the Kuma Control Plane Service's name |
| controlPlane.service.type | string | `"ClusterIP"` | Service type of the Kuma Control Plane |
| controlPlane.service.annotations | object | `{}` | Additional annotations to put on the Kuma Control Plane |
| controlPlane.ingress.enabled | bool | `false` | Install K8s Ingress resource that exposes GUI and API |
| controlPlane.ingress.ingressClassName | string | `nil` | IngressClass defines which controller will implement the resource |
| controlPlane.ingress.hostname | string | `nil` | Ingress hostname |
| controlPlane.ingress.annotations | object | `{}` | Map of ingress annotations. |
| controlPlane.ingress.path | string | `"/"` | Ingress path. |
| controlPlane.ingress.pathType | string | `"ImplementationSpecific"` | Each path in an Ingress is required to have a corresponding path type. (ImplementationSpecific/Exact/Prefix) |
| controlPlane.globalZoneSyncService.enabled | bool | `true` | Whether to create a k8s service for the global zone sync service. It will only be created when enabled and deploying the global control plane. |
| controlPlane.globalZoneSyncService.type | string | `"LoadBalancer"` | Service type of the Global-zone sync |
| controlPlane.globalZoneSyncService.loadBalancerIP | string | `nil` | Optionally specify IP to be used by cloud provider when configuring load balancer |
| controlPlane.globalZoneSyncService.annotations | object | `{}` | Additional annotations to put on the Global Zone Sync Service |
| controlPlane.globalZoneSyncService.port | int | `5685` | Port on which Global Zone Sync Service is exposed |
| controlPlane.defaults.skipMeshCreation | bool | `false` | Whether to skip creating the default Mesh |
| controlPlane.automountServiceAccountToken | bool | `true` | Whether to automountServiceAccountToken for cp. Optionally set to false |
| controlPlane.resources | object | the resources will be chosen based on the mode | Optionally override the resource spec |
| controlPlane.lifecycle | object | `{}` | Pod lifecycle settings (useful for adding a preStop hook, when using AWS ALB or NLB) |
| controlPlane.terminationGracePeriodSeconds | int | `30` | Number of seconds to wait before force killing the pod. Make sure to update this if you add a preStop hook. |
| controlPlane.tls.general.secretName | string | `""` | Secret that contains tls.crt, tls.key [and ca.crt when no controlPlane.tls.general.caSecretName specified] for protecting Kuma in-cluster communication |
| controlPlane.tls.general.caSecretName | string | `""` | Secret that contains ca.crt that was used to sign cert for protecting Kuma in-cluster communication (ca.crt present in this secret have precedence over the one provided in the controlPlane.tls.general.secretName) |
| controlPlane.tls.general.caBundle | string | `""` | Base64 encoded CA certificate (the same as in controlPlane.tls.general.secret#ca.crt) |
| controlPlane.tls.apiServer.secretName | string | `""` | Secret that contains tls.crt, tls.key for protecting Kuma API on HTTPS |
| controlPlane.tls.apiServer.clientCertsSecretName | string | `""` | Secret that contains list of .pem certificates that can access admin endpoints of Kuma API on HTTPS |
| controlPlane.tls.kdsGlobalServer.secretName | string | `""` | Name of the K8s TLS Secret resource. If you set this and don't set create=true, you have to create the secret manually. |
| controlPlane.tls.kdsGlobalServer.create | bool | `false` | Whether to create the TLS secret in helm. |
| controlPlane.tls.kdsGlobalServer.cert | string | `""` | The TLS certificate to offer. |
| controlPlane.tls.kdsGlobalServer.key | string | `""` | The TLS key to use. |
| controlPlane.tls.kdsZoneClient.secretName | string | `""` | Name of the K8s Secret resource that contains ca.crt which was used to sign the certificate of KDS Global Server. If you set this and don't set create=true, you have to create the secret manually. |
| controlPlane.tls.kdsZoneClient.create | bool | `false` | Whether to create the TLS secret in helm. |
| controlPlane.tls.kdsZoneClient.cert | string | `""` | CA bundle that was used to sign the certificate of KDS Global Server. |
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
| cni.nodeSelector | object | `{"kubernetes.io/os":"linux"}` | Node Selector for the CNI pods |
| cni.podAnnotations | object | `{}` | Additional pod annotations |
| cni.image.registry | string | `"docker.io/kumahq"` | CNI image registry |
| cni.image.repository | string | `"install-cni"` | CNI image repository |
| cni.image.tag | string | `"0.0.10"` | CNI image tag |
| cni.image.imagePullPolicy | string | `"IfNotPresent"` | CNI image pull policy |
| cni.delayStartupSeconds | int | `0` | it's only useful in tests to trigger a possible race condition |
| cni.experimental | object | `{"image":{"repository":"kuma-cni","tag":null},"imageEbpf":{"registry":"docker.io/merbridge","repository":"merbridge","tag":"0.7.1"}}` | use new CNI image (experimental) |
| cni.experimental.image.repository | string | `"kuma-cni"` | CNI experimental image repository |
| cni.experimental.image.tag | string | `nil` | CNI experimental image tag - defaults to .Chart.AppVersion |
| cni.experimental.imageEbpf.registry | string | `"docker.io/merbridge"` | CNI experimental eBPF image registry |
| cni.experimental.imageEbpf.repository | string | `"merbridge"` | CNI experimental eBPF image repository |
| cni.experimental.imageEbpf.tag | string | `"0.7.1"` | CNI experimental eBPF image tag |
| cni.podSecurityContext | object | `{}` | Security context at the pod level for cni |
| cni.containerSecurityContext | object | `{}` | Security context at the container level for cni |
| dataPlane.image.repository | string | `"kuma-dp"` | The Kuma DP image repository |
| dataPlane.image.pullPolicy | string | `"IfNotPresent"` | Kuma DP ImagePullPolicy |
| dataPlane.image.tag | string | `nil` | Kuma DP Image Tag. When not specified, the value is copied from global.tag |
| dataPlane.initImage.repository | string | `"kuma-init"` | The Kuma DP init image repository |
| dataPlane.initImage.tag | string | `nil` | Kuma DP init image tag When not specified, the value is copied from global.tag |
| ingress.enabled | bool | `false` | If true, it deploys Ingress for cross cluster communication |
| ingress.extraLabels | object | `{}` | Labels to add to resources, in addition to default labels |
| ingress.drainTime | string | `"30s"` | Time for which old listener will still be active as draining |
| ingress.replicas | int | `1` | Number of replicas of the Ingress. Ignored when autoscaling is enabled. |
| ingress.resources | object | `{"limits":{"cpu":"1000m","memory":"512Mi"},"requests":{"cpu":"50m","memory":"64Mi"}}` | Define the resources to allocate to mesh ingress |
| ingress.lifecycle | object | `{}` | Pod lifecycle settings (useful for adding a preStop hook, when using AWS ALB or NLB) |
| ingress.terminationGracePeriodSeconds | int | `30` | Number of seconds to wait before force killing the pod. Make sure to update this if you add a preStop hook. |
| ingress.autoscaling.enabled | bool | `false` | Whether to enable Horizontal Pod Autoscaling, which requires the [Metrics Server](https://github.com/kubernetes-sigs/metrics-server) in the cluster |
| ingress.autoscaling.minReplicas | int | `2` | The minimum CP pods to allow |
| ingress.autoscaling.maxReplicas | int | `5` | The max CP pods to scale to |
| ingress.autoscaling.targetCPUUtilizationPercentage | int | `80` | For clusters that don't support autoscaling/v2beta, autoscaling/v1 is used |
| ingress.autoscaling.metrics | list | `[{"resource":{"name":"cpu","target":{"averageUtilization":80,"type":"Utilization"}},"type":"Resource"}]` | For clusters that do support autoscaling/v2beta, use metrics |
| ingress.service.enabled | bool | `true` | Whether to create a Service resource. |
| ingress.service.type | string | `"LoadBalancer"` | Service type of the Ingress |
| ingress.service.loadBalancerIP | string | `nil` | Optionally specify IP to be used by cloud provider when configuring load balancer |
| ingress.service.annotations | object | `{}` | Additional annotations to put on the Ingress service |
| ingress.service.port | int | `10001` | Port on which Ingress is exposed |
| ingress.service.nodePort | string | `nil` | Port on which service is exposed on Node for service of type NodePort |
| ingress.annotations | object | `{}` | Additional pod annotations (deprecated favor `podAnnotations`) |
| ingress.podAnnotations | object | `{}` | Additional pod annotations |
| ingress.nodeSelector | object | `{"kubernetes.io/os":"linux"}` | Node Selector for the Ingress pods |
| ingress.podDisruptionBudget.enabled | bool | `false` | Whether to create a pod disruption budget |
| ingress.podDisruptionBudget.maxUnavailable | int | `1` | The maximum number of unavailable pods allowed by the budget |
| ingress.affinity | object | `{"podAntiAffinity":{"preferredDuringSchedulingIgnoredDuringExecution":[{"podAffinityTerm":{"labelSelector":{"matchExpressions":[{"key":"app.kubernetes.io/name","operator":"In","values":["{{ include \"kuma.name\" . }}"]},{"key":"app.kubernetes.io/instance","operator":"In","values":["{{ .Release.Name }}"]},{"key":"app","operator":"In","values":["kuma-ingress"]}]},"topologyKey":"kubernetes.io/hostname"},"weight":100}]}}` | Affinity placement rule for the Kuma Ingress pods This is rendered as a template, so you can reference other helm variables or includes. |
| ingress.topologySpreadConstraints | string | `nil` | Topology spread constraints rule for the Kuma Mesh Ingress pods. This is rendered as a template, so you can use variables to generate match labels. |
| ingress.podSecurityContext | object | `{}` | Security context at the pod level for ingress |
| ingress.containerSecurityContext | object | `{}` | Security context at the container level for ingress |
| egress.enabled | bool | `false` | If true, it deploys Egress for cross cluster communication |
| egress.extraLabels | object | `{}` | Labels to add to resources, in addition to the default labels. |
| egress.drainTime | string | `"30s"` | Time for which old listener will still be active as draining |
| egress.replicas | int | `1` | Number of replicas of the Egress. Ignored when autoscaling is enabled. |
| egress.autoscaling.enabled | bool | `false` | Whether to enable Horizontal Pod Autoscaling, which requires the [Metrics Server](https://github.com/kubernetes-sigs/metrics-server) in the cluster |
| egress.autoscaling.minReplicas | int | `2` | The minimum CP pods to allow |
| egress.autoscaling.maxReplicas | int | `5` | The max CP pods to scale to |
| egress.autoscaling.targetCPUUtilizationPercentage | int | `80` | For clusters that don't support autoscaling/v2beta, autoscaling/v1 is used |
| egress.autoscaling.metrics | list | `[{"resource":{"name":"cpu","target":{"averageUtilization":80,"type":"Utilization"}},"type":"Resource"}]` | For clusters that do support autoscaling/v2beta, use metrics |
| egress.service.enabled | bool | `true` | Whether to create the service object |
| egress.service.type | string | `"ClusterIP"` | Service type of the Egress |
| egress.service.loadBalancerIP | string | `nil` | Optionally specify IP to be used by cloud provider when configuring load balancer |
| egress.service.annotations | object | `{}` | Additional annotations to put on the Egress service |
| egress.service.port | int | `10002` | Port on which Egress is exposed |
| egress.service.nodePort | string | `nil` | Port on which service is exposed on Node for service of type NodePort |
| egress.annotations | object | `{}` | Additional pod annotations (deprecated favor `podAnnotations`) |
| egress.podAnnotations | object | `{}` | Additional pod annotations |
| egress.nodeSelector | object | `{"kubernetes.io/os":"linux"}` | Node Selector for the Egress pods |
| egress.podDisruptionBudget.enabled | bool | `false` | Whether to create a pod disruption budget |
| egress.podDisruptionBudget.maxUnavailable | int | `1` | The maximum number of unavailable pods allowed by the budget |
| egress.affinity | object | `{"podAntiAffinity":{"preferredDuringSchedulingIgnoredDuringExecution":[{"podAffinityTerm":{"labelSelector":{"matchExpressions":[{"key":"app.kubernetes.io/name","operator":"In","values":["{{ include \"kuma.name\" . }}"]},{"key":"app.kubernetes.io/instance","operator":"In","values":["{{ .Release.Name }}"]},{"key":"app","operator":"In","values":["kuma-egress"]}]},"topologyKey":"kubernetes.io/hostname"},"weight":100}]}}` | Affinity placement rule for the Kuma Egress pods. This is rendered as a template, so you can reference other helm variables or includes. |
| egress.topologySpreadConstraints | string | `nil` | Topology spread constraints rule for the Kuma Egress pods. This is rendered as a template, so you can use variables to generate match labels. |
| egress.podSecurityContext | object | `{}` | Security context at the pod level for egress |
| egress.containerSecurityContext | object | `{}` | Security context at the container level for egress |
| kumactl.image.repository | string | `"kumactl"` | The kumactl image repository |
| kumactl.image.tag | string | `nil` | The kumactl image tag. When not specified, the value is copied from global.tag |
| kubectl.image.registry | string | `"kumahq"` | The kubectl image registry |
| kubectl.image.repository | string | `"kubectl"` | The kubectl image repository |
| kubectl.image.tag | string | `"v1.20.15"` | The kubectl image tag |
| hooks.nodeSelector | object | `{"kubernetes.io/os":"linux"}` | Node selector for the HELM hooks |
| hooks.podSecurityContext | object | `{}` | Security context at the pod level for crd/webhook/ns |
| hooks.containerSecurityContext | object | `{}` | Security context at the container level for crd/webhook/ns |
| experimental.gatewayAPI | bool | `false` | If true, it installs experimental Gateway API support |
| experimental.cni | bool | `false` | If true, it installs experimental new version of the CNI |
| experimental.transparentProxy | bool | `false` | If true, use the new transparent proxy engine |
| experimental.ebpf.enabled | bool | `false` | If true, ebpf will be used instead of using iptables to install/configure transparent proxy |
| experimental.ebpf.instanceIPEnvVarName | string | `"INSTANCE_IP"` | Name of the environmental variable which will contain the IP address of a pod |
| experimental.ebpf.bpffsPath | string | `"/run/kuma/bpf"` | Path where BPF file system should be mounted |
| experimental.ebpf.programsSourcePath | string | `"/kuma/ebpf"` | Path where compiled eBPF programs which will be installed can be found |

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
