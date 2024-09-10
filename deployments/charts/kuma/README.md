[![][kuma-logo]][kuma-url]

A Helm chart for the Kuma Control Plane

![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![Version: 0.0.0-preview.vlocal-build](https://img.shields.io/badge/Version-0.0.0--preview.vlocal--build-informational?style=flat-square) ![AppVersion: 0.0.0-preview.vlocal-build](https://img.shields.io/badge/AppVersion-0.0.0--preview.vlocal--build-informational?style=flat-square)

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
| noHelmHooks | bool | `false` | Whether to disable all helm hooks |
| restartOnSecretChange | bool | `true` | Whether to restart control-plane by calculating a new checksum for the secret |
| controlPlane.environment | string | `"kubernetes"` | Environment that control plane is run in, useful when running universal global control plane on k8s |
| controlPlane.extraLabels | object | `{}` | Labels to add to resources in addition to default labels |
| controlPlane.logLevel | string | `"info"` | Kuma CP log level: one of off,info,debug |
| controlPlane.logOutputPath | string | `""` | Kuma CP log output path: Defaults to /dev/stdout |
| controlPlane.mode | string | `"zone"` | Kuma CP modes: one of zone,global |
| controlPlane.zone | string | `nil` | Kuma CP zone, if running multizone |
| controlPlane.kdsGlobalAddress | string | `""` | Only used in `zone` mode |
| controlPlane.replicas | int | `1` | Number of replicas of the Kuma CP. Ignored when autoscaling is enabled |
| controlPlane.minReadySeconds | int | `0` | Minimum number of seconds for which a newly created pod should be ready for it to be considered available. |
| controlPlane.deploymentAnnotations | object | `{}` | Annotations applied only to the `Deployment` resource |
| controlPlane.podAnnotations | object | `{}` | Annotations applied only to the `Pod` resource |
| controlPlane.autoscaling.enabled | bool | `false` | Whether to enable Horizontal Pod Autoscaling, which requires the [Metrics Server](https://github.com/kubernetes-sigs/metrics-server) in the cluster |
| controlPlane.autoscaling.minReplicas | int | `2` | The minimum CP pods to allow |
| controlPlane.autoscaling.maxReplicas | int | `5` | The max CP pods to scale to |
| controlPlane.autoscaling.targetCPUUtilizationPercentage | int | `80` | For clusters that don't support autoscaling/v2, autoscaling/v1 is used |
| controlPlane.autoscaling.metrics | list | `[{"resource":{"name":"cpu","target":{"averageUtilization":80,"type":"Utilization"}},"type":"Resource"}]` | For clusters that do support autoscaling/v2, use metrics |
| controlPlane.nodeSelector | object | `{"kubernetes.io/os":"linux"}` | Node selector for the Kuma Control Plane pods |
| controlPlane.tolerations | list | `[]` | Tolerations for the Kuma Control Plane pods |
| controlPlane.podDisruptionBudget.enabled | bool | `false` | Whether to create a pod disruption budget |
| controlPlane.podDisruptionBudget.maxUnavailable | int | `1` | The maximum number of unavailable pods allowed by the budget |
| controlPlane.affinity | object | `{"podAntiAffinity":{"preferredDuringSchedulingIgnoredDuringExecution":[{"podAffinityTerm":{"labelSelector":{"matchExpressions":[{"key":"app.kubernetes.io/name","operator":"In","values":["{{ include \"kuma.name\" . }}"]},{"key":"app.kubernetes.io/instance","operator":"In","values":["{{ .Release.Name }}"]},{"key":"app","operator":"In","values":["{{ include \"kuma.name\" . }}-control-plane"]}]},"topologyKey":"kubernetes.io/hostname"},"weight":100}]}}` | Affinity placement rule for the Kuma Control Plane pods. This is rendered as a template, so you can reference other helm variables or includes. |
| controlPlane.topologySpreadConstraints | string | `nil` | Topology spread constraints rule for the Kuma Control Plane pods. This is rendered as a template, so you can use variables to generate match labels. |
| controlPlane.injectorFailurePolicy | string | `"Fail"` | Failure policy of the mutating webhook implemented by the Kuma Injector component |
| controlPlane.service.apiServer.http.nodePort | int | `30681` | Port on which Http api server Service is exposed on Node for service of type NodePort |
| controlPlane.service.apiServer.https.nodePort | int | `30682` | Port on which Https api server Service is exposed on Node for service of type NodePort |
| controlPlane.service.enabled | bool | `true` | Whether to create a service resource. |
| controlPlane.service.name | string | `nil` | Optionally override of the Kuma Control Plane Service's name |
| controlPlane.service.type | string | `"ClusterIP"` | Service type of the Kuma Control Plane |
| controlPlane.service.annotations | object | `{"prometheus.io/port":"5680","prometheus.io/scrape":"true"}` | Annotations to put on the Kuma Control Plane |
| controlPlane.ingress.enabled | bool | `false` | Install K8s Ingress resource that exposes GUI and API |
| controlPlane.ingress.ingressClassName | string | `nil` | IngressClass defines which controller will implement the resource |
| controlPlane.ingress.hostname | string | `nil` | Ingress hostname |
| controlPlane.ingress.annotations | object | `{}` | Map of ingress annotations. |
| controlPlane.ingress.path | string | `"/"` | Ingress path. |
| controlPlane.ingress.pathType | string | `"ImplementationSpecific"` | Each path in an Ingress is required to have a corresponding path type. (ImplementationSpecific/Exact/Prefix) |
| controlPlane.ingress.servicePort | int | `5681` | Port from kuma-cp to use to expose API and GUI. Switch to 5682 to expose TLS port |
| controlPlane.globalZoneSyncService.enabled | bool | `true` | Whether to create a k8s service for the global zone sync service. It will only be created when enabled and deploying the global control plane. |
| controlPlane.globalZoneSyncService.type | string | `"LoadBalancer"` | Service type of the Global-zone sync |
| controlPlane.globalZoneSyncService.loadBalancerIP | string | `nil` | Optionally specify IP to be used by cloud provider when configuring load balancer |
| controlPlane.globalZoneSyncService.loadBalancerSourceRanges | list | `[]` | Optionally specify allowed source ranges that can access the load balancer |
| controlPlane.globalZoneSyncService.annotations | object | `{}` | Additional annotations to put on the Global Zone Sync Service |
| controlPlane.globalZoneSyncService.nodePort | int | `30685` | Port on which Global Zone Sync Service is exposed on Node for service of type NodePort |
| controlPlane.globalZoneSyncService.port | int | `5685` | Port on which Global Zone Sync Service is exposed |
| controlPlane.globalZoneSyncService.protocol | string | `"grpc"` | Protocol of the Global Zone Sync service port |
| controlPlane.defaults.skipMeshCreation | bool | `false` | Whether to skip creating the default Mesh |
| controlPlane.automountServiceAccountToken | bool | `true` | Whether to automountServiceAccountToken for cp. Optionally set to false |
| controlPlane.resources | object | `{"limits":{"memory":"256Mi"},"requests":{"cpu":"500m","memory":"256Mi"}}` | Optionally override the resource spec |
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
| controlPlane.tls.kdsZoneClient.skipVerify | bool | `false` | If true, TLS cert of the server is not verified. |
| controlPlane.serviceAccountAnnotations | object | `{}` | Annotations to add for Control Plane's Service Account |
| controlPlane.image.pullPolicy | string | `"IfNotPresent"` | Kuma CP ImagePullPolicy |
| controlPlane.image.repository | string | `"kuma-cp"` | Kuma CP image repository |
| controlPlane.image.tag | string | `nil` | Kuma CP Image tag. When not specified, the value is copied from global.tag |
| controlPlane.secrets | object with { Env: string, Secret: string, Key: string } | `nil` | Secrets to add as environment variables, where `Env` is the name of the env variable, `Secret` is the name of the Secret, and `Key` is the key of the Secret value to use |
| controlPlane.envVars | object | `{}` | Additional environment variables that will be passed to the control plane |
| controlPlane.envVarEntries | string | `nil` | Additional environment variables that will be passed to the control plane. Can be used with Kubernetes downward API |
| controlPlane.extraConfigMaps | list | `[]` | Additional config maps to mount into the control plane, with optional inline values |
| controlPlane.extraSecrets | object with { name: string, mountPath: string, readOnly: string } | `nil` | Additional secrets to mount into the control plane, where `Env` is the name of the env variable, `Secret` is the name of the Secret, and `Key` is the key of the Secret value to use |
| controlPlane.webhooks.validator.additionalRules | string | `""` | Additional rules to apply on Kuma validator webhook. Useful when building custom policy on top of Kuma. |
| controlPlane.webhooks.ownerReference.additionalRules | string | `""` | Additional rules to apply on Kuma owner reference webhook. Useful when building custom policy on top of Kuma. |
| controlPlane.hostNetwork | bool | `false` | Specifies if the deployment should be started in hostNetwork mode. |
| controlPlane.admissionServerPort | int | `5443` | Define a new server port for the admission controller. Recommended to set in combination with hostNetwork to prevent multiple port bindings on the same port (like Calico in AWS EKS). |
| controlPlane.podSecurityContext | object | `{"runAsNonRoot":true}` | Security context at the pod level for control plane. |
| controlPlane.containerSecurityContext | object | `{"readOnlyRootFilesystem":true}` | Security context at the container level for control plane. |
| controlPlane.supportGatewaySecretsInAllNamespaces | bool | `false` | If true, then control plane can support TLS secrets for builtin gateway outside of mesh system namespace. The downside is that control plane requires permission to read Secrets in all namespaces. |
| controlPlane.dns | object | `{"config":{"nameservers":[],"searches":[]},"policy":""}` | DNS configuration for the control-plane pod. This is equivalent to the [Kubernetes DNS policy](https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#pod-s-dns-policy). |
| controlPlane.dns.policy | string | `""` | Defines how DNS resolution is configured for that Pod. |
| controlPlane.dns.config | object | `{"nameservers":[],"searches":[]}` | Optional dns configuration, required when policy is 'None' |
| controlPlane.dns.config.nameservers | list | `[]` | A list of IP addresses that will be used as DNS servers for the Pod. There can be at most 3 IP addresses specified. |
| controlPlane.dns.config.searches | list | `[]` | A list of DNS search domains for hostname lookup in the Pod. |
| cni.enabled | bool | `false` | Install Kuma with CNI instead of proxy init container |
| cni.chained | bool | `false` | Install CNI in chained mode |
| cni.netDir | string | `"/etc/cni/multus/net.d"` | Set the CNI install directory |
| cni.binDir | string | `"/var/lib/cni/bin"` | Set the CNI bin directory |
| cni.confName | string | `"kuma-cni.conf"` | Set the CNI configuration name |
| cni.logLevel | string | `"info"` | CNI log level: one of off,info,debug |
| cni.nodeSelector | object | `{"kubernetes.io/os":"linux"}` | Node Selector for the CNI pods |
| cni.tolerations | list | `[]` | Tolerations for the CNI pods |
| cni.podAnnotations | object | `{}` | Additional pod annotations |
| cni.namespace | string | `"kube-system"` | Set the CNI namespace |
| cni.image.repository | string | `"kuma-cni"` | CNI image repository |
| cni.image.tag | string | `nil` | CNI image tag - defaults to .Chart.AppVersion |
| cni.image.imagePullPolicy | string | `"IfNotPresent"` | CNI image pull policy |
| cni.delayStartupSeconds | int | `0` | it's only useful in tests to trigger a possible race condition |
| cni.experimental | object | `{"imageEbpf":{"registry":"docker.io/kumahq","repository":"merbridge","tag":"0.8.5"}}` | use new CNI (experimental) |
| cni.experimental.imageEbpf.registry | string | `"docker.io/kumahq"` | CNI experimental eBPF image registry |
| cni.experimental.imageEbpf.repository | string | `"merbridge"` | CNI experimental eBPF image repository |
| cni.experimental.imageEbpf.tag | string | `"0.8.5"` | CNI experimental eBPF image tag |
| cni.resources.requests.cpu | string | `"100m"` |  |
| cni.resources.requests.memory | string | `"100Mi"` |  |
| cni.resources.limits.memory | string | `"100Mi"` |  |
| cni.podSecurityContext | object | `{}` | Security context at the pod level for cni |
| cni.containerSecurityContext | object | `{"readOnlyRootFilesystem":true,"runAsGroup":0,"runAsNonRoot":false,"runAsUser":0}` | Security context at the container level for cni |
| dataPlane.dnsLogging | bool | `false` | If true, then turn on CoreDNS query logging |
| dataPlane.image.repository | string | `"kuma-dp"` | The Kuma DP image repository |
| dataPlane.image.pullPolicy | string | `"IfNotPresent"` | Kuma DP ImagePullPolicy |
| dataPlane.image.tag | string | `nil` | Kuma DP Image Tag. When not specified, the value is copied from global.tag |
| dataPlane.initImage.repository | string | `"kuma-init"` | The Kuma DP init image repository |
| dataPlane.initImage.tag | string | `nil` | Kuma DP init image tag When not specified, the value is copied from global.tag |
| ingress.enabled | bool | `false` | If true, it deploys Ingress for cross cluster communication |
| ingress.extraLabels | object | `{}` | Labels to add to resources, in addition to default labels |
| ingress.drainTime | string | `"30s"` | Time for which old listener will still be active as draining |
| ingress.replicas | int | `1` | Number of replicas of the Ingress. Ignored when autoscaling is enabled. |
| ingress.logLevel | string | `"info"` | Log level for ingress (available values: off|info|debug) |
| ingress.resources | object | `{"limits":{"cpu":"1000m","memory":"512Mi"},"requests":{"cpu":"50m","memory":"64Mi"}}` | Define the resources to allocate to mesh ingress |
| ingress.lifecycle | object | `{}` | Pod lifecycle settings (useful for adding a preStop hook, when using AWS ALB or NLB) |
| ingress.terminationGracePeriodSeconds | int | `40` | Number of seconds to wait before force killing the pod. Make sure to update this if you add a preStop hook. |
| ingress.autoscaling.enabled | bool | `false` | Whether to enable Horizontal Pod Autoscaling, which requires the [Metrics Server](https://github.com/kubernetes-sigs/metrics-server) in the cluster |
| ingress.autoscaling.minReplicas | int | `2` | The minimum CP pods to allow |
| ingress.autoscaling.maxReplicas | int | `5` | The max CP pods to scale to |
| ingress.autoscaling.targetCPUUtilizationPercentage | int | `80` | For clusters that don't support autoscaling/v2, autoscaling/v1 is used |
| ingress.autoscaling.metrics | list | `[{"resource":{"name":"cpu","target":{"averageUtilization":80,"type":"Utilization"}},"type":"Resource"}]` | For clusters that do support autoscaling/v2, use metrics |
| ingress.service.enabled | bool | `true` | Whether to create a Service resource. |
| ingress.service.type | string | `"LoadBalancer"` | Service type of the Ingress |
| ingress.service.loadBalancerIP | string | `nil` | Optionally specify IP to be used by cloud provider when configuring load balancer |
| ingress.service.annotations | object | `{}` | Additional annotations to put on the Ingress service |
| ingress.service.port | int | `10001` | Port on which Ingress is exposed |
| ingress.service.nodePort | string | `nil` | Port on which service is exposed on Node for service of type NodePort |
| ingress.annotations | object | `{}` | Additional pod annotations (deprecated favor `podAnnotations`) |
| ingress.podAnnotations | object | `{}` | Additional pod annotations |
| ingress.nodeSelector | object | `{"kubernetes.io/os":"linux"}` | Node Selector for the Ingress pods |
| ingress.tolerations | list | `[]` | Tolerations for the Ingress pods |
| ingress.podDisruptionBudget.enabled | bool | `false` | Whether to create a pod disruption budget |
| ingress.podDisruptionBudget.maxUnavailable | int | `1` | The maximum number of unavailable pods allowed by the budget |
| ingress.affinity | object | `{"podAntiAffinity":{"preferredDuringSchedulingIgnoredDuringExecution":[{"podAffinityTerm":{"labelSelector":{"matchExpressions":[{"key":"app.kubernetes.io/name","operator":"In","values":["{{ include \"kuma.name\" . }}"]},{"key":"app.kubernetes.io/instance","operator":"In","values":["{{ .Release.Name }}"]},{"key":"app","operator":"In","values":["kuma-ingress"]}]},"topologyKey":"kubernetes.io/hostname"},"weight":100}]}}` | Affinity placement rule for the Kuma Ingress pods This is rendered as a template, so you can reference other helm variables or includes. |
| ingress.topologySpreadConstraints | string | `nil` | Topology spread constraints rule for the Kuma Mesh Ingress pods. This is rendered as a template, so you can use variables to generate match labels. |
| ingress.podSecurityContext | object | `{"runAsGroup":5678,"runAsNonRoot":true,"runAsUser":5678}` | Security context at the pod level for ingress |
| ingress.containerSecurityContext | object | `{"readOnlyRootFilesystem":true}` | Security context at the container level for ingress |
| ingress.serviceAccountAnnotations | object | `{}` | Annotations to add for Control Plane's Service Account |
| ingress.automountServiceAccountToken | bool | `true` | Whether to automountServiceAccountToken for cp. Optionally set to false |
| ingress.dns | object | `{"config":{"nameservers":[],"searches":[]},"policy":""}` | DNS configuration for the ingress pod. This is equivalent to the [Kubernetes DNS policy](https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#pod-s-dns-policy). |
| ingress.dns.policy | string | `""` | Defines how DNS resolution is configured for that Pod. |
| ingress.dns.config | object | `{"nameservers":[],"searches":[]}` | Optional dns configuration, required when policy is 'None' |
| ingress.dns.config.nameservers | list | `[]` | A list of IP addresses that will be used as DNS servers for the Pod. There can be at most 3 IP addresses specified. |
| ingress.dns.config.searches | list | `[]` | A list of DNS search domains for hostname lookup in the Pod. |
| egress.enabled | bool | `false` | If true, it deploys Egress for cross cluster communication |
| egress.extraLabels | object | `{}` | Labels to add to resources, in addition to the default labels. |
| egress.drainTime | string | `"30s"` | Time for which old listener will still be active as draining |
| egress.replicas | int | `1` | Number of replicas of the Egress. Ignored when autoscaling is enabled. |
| egress.logLevel | string | `"info"` | Log level for egress (available values: off|info|debug) |
| egress.autoscaling.enabled | bool | `false` | Whether to enable Horizontal Pod Autoscaling, which requires the [Metrics Server](https://github.com/kubernetes-sigs/metrics-server) in the cluster |
| egress.autoscaling.minReplicas | int | `2` | The minimum CP pods to allow |
| egress.autoscaling.maxReplicas | int | `5` | The max CP pods to scale to |
| egress.autoscaling.targetCPUUtilizationPercentage | int | `80` | For clusters that don't support autoscaling/v2, autoscaling/v1 is used |
| egress.autoscaling.metrics | list | `[{"resource":{"name":"cpu","target":{"averageUtilization":80,"type":"Utilization"}},"type":"Resource"}]` | For clusters that do support autoscaling/v2, use metrics |
| egress.resources.requests.cpu | string | `"50m"` |  |
| egress.resources.requests.memory | string | `"64Mi"` |  |
| egress.resources.limits.cpu | string | `"1000m"` |  |
| egress.resources.limits.memory | string | `"512Mi"` |  |
| egress.service.enabled | bool | `true` | Whether to create the service object |
| egress.service.type | string | `"ClusterIP"` | Service type of the Egress |
| egress.service.loadBalancerIP | string | `nil` | Optionally specify IP to be used by cloud provider when configuring load balancer |
| egress.service.annotations | object | `{}` | Additional annotations to put on the Egress service |
| egress.service.port | int | `10002` | Port on which Egress is exposed |
| egress.service.nodePort | string | `nil` | Port on which service is exposed on Node for service of type NodePort |
| egress.annotations | object | `{}` | Additional pod annotations (deprecated favor `podAnnotations`) |
| egress.podAnnotations | object | `{}` | Additional pod annotations |
| egress.nodeSelector | object | `{"kubernetes.io/os":"linux"}` | Node Selector for the Egress pods |
| egress.tolerations | list | `[]` | Tolerations for the Egress pods |
| egress.podDisruptionBudget.enabled | bool | `false` | Whether to create a pod disruption budget |
| egress.podDisruptionBudget.maxUnavailable | int | `1` | The maximum number of unavailable pods allowed by the budget |
| egress.affinity | object | `{"podAntiAffinity":{"preferredDuringSchedulingIgnoredDuringExecution":[{"podAffinityTerm":{"labelSelector":{"matchExpressions":[{"key":"app.kubernetes.io/name","operator":"In","values":["{{ include \"kuma.name\" . }}"]},{"key":"app.kubernetes.io/instance","operator":"In","values":["{{ .Release.Name }}"]},{"key":"app","operator":"In","values":["kuma-egress"]}]},"topologyKey":"kubernetes.io/hostname"},"weight":100}]}}` | Affinity placement rule for the Kuma Egress pods. This is rendered as a template, so you can reference other helm variables or includes. |
| egress.topologySpreadConstraints | string | `nil` | Topology spread constraints rule for the Kuma Egress pods. This is rendered as a template, so you can use variables to generate match labels. |
| egress.podSecurityContext | object | `{"runAsGroup":5678,"runAsNonRoot":true,"runAsUser":5678}` | Security context at the pod level for egress |
| egress.containerSecurityContext | object | `{"readOnlyRootFilesystem":true}` | Security context at the container level for egress |
| egress.serviceAccountAnnotations | object | `{}` | Annotations to add for Control Plane's Service Account |
| egress.automountServiceAccountToken | bool | `true` | Whether to automountServiceAccountToken for cp. Optionally set to false |
| egress.dns | object | `{"config":{"nameservers":[],"searches":[]},"policy":""}` | DNS configuration for the egress pod. This is equivalent to the [Kubernetes DNS policy](https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#pod-s-dns-policy). |
| egress.dns.policy | string | `""` | Defines how DNS resolution is configured for that Pod. |
| egress.dns.config | object | `{"nameservers":[],"searches":[]}` | Optional dns configuration, required when policy is 'None' |
| egress.dns.config.nameservers | list | `[]` | A list of IP addresses that will be used as DNS servers for the Pod. There can be at most 3 IP addresses specified. |
| egress.dns.config.searches | list | `[]` | A list of DNS search domains for hostname lookup in the Pod. |
| kumactl.image.repository | string | `"kumactl"` | The kumactl image repository |
| kumactl.image.tag | string | `nil` | The kumactl image tag. When not specified, the value is copied from global.tag |
| kubectl.image.registry | string | `"docker.io"` | The kubectl image registry |
| kubectl.image.repository | string | `"bitnami/kubectl"` | The kubectl image repository |
| kubectl.image.tag | string | `"1.27.5"` | The kubectl image tag |
| hooks.nodeSelector | object | `{"kubernetes.io/os":"linux"}` | Node selector for the HELM hooks |
| hooks.tolerations | list | `[]` | Tolerations for the HELM hooks |
| hooks.podSecurityContext | object | `{"runAsNonRoot":true}` | Security context at the pod level for crd/webhook/ns |
| hooks.containerSecurityContext | object | `{"readOnlyRootFilesystem":true}` | Security context at the container level for crd/webhook/ns |
| hooks.ebpfCleanup | object | `{"containerSecurityContext":{"readOnlyRootFilesystem":false},"podSecurityContext":{"runAsNonRoot":false}}` | ebpf-cleanup hook needs write access to the root filesystem to clean ebpf programs Changing below values will potentially break ebpf cleanup completely, so be cautious when doing so. |
| hooks.ebpfCleanup.podSecurityContext | object | `{"runAsNonRoot":false}` | Security context at the pod level for crd/webhook/cleanup-ebpf |
| hooks.ebpfCleanup.containerSecurityContext | object | `{"readOnlyRootFilesystem":false}` | Security context at the container level for crd/webhook/cleanup-ebpf |
| experimental.ebpf.enabled | bool | `false` | If true, ebpf will be used instead of using iptables to install/configure transparent proxy |
| experimental.ebpf.instanceIPEnvVarName | string | `"INSTANCE_IP"` | Name of the environmental variable which will contain the IP address of a pod |
| experimental.ebpf.bpffsPath | string | `"/sys/fs/bpf"` | Path where BPF file system should be mounted |
| experimental.ebpf.cgroupPath | string | `"/sys/fs/cgroup"` | Host's cgroup2 path |
| experimental.ebpf.tcAttachIface | string | `""` | Name of the network interface which TC programs should be attached to, we'll try to automatically determine it if empty |
| experimental.ebpf.programsSourcePath | string | `"/tmp/kuma-ebpf"` | Path where compiled eBPF programs which will be installed can be found |
| experimental.sidecarContainers | bool | `false` | If true, enable native Kubernetes sidecars. This requires at least Kubernetes v1.29 |
| postgres.port | string | `"5432"` | Postgres port, password should be provided as a secret reference in "controlPlane.secrets" with the Env value "KUMA_STORE_POSTGRES_PASSWORD". Example: controlPlane:   secrets:     - Secret: postgres-postgresql       Key: postgresql-password       Env: KUMA_STORE_POSTGRES_PASSWORD |
| postgres.tls.mode | string | `"disable"` | Mode of TLS connection. Available values are: "disable", "verifyNone", "verifyCa", "verifyFull" |
| postgres.tls.disableSSLSNI | bool | `false` | Whether to disable SNI the postgres `sslsni` option. |
| postgres.tls.caSecretName | string | `nil` | Secret name that contains the ca.crt |
| postgres.tls.secretName | string | `nil` | Secret name that contains the client tls.crt, tls.key |

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
