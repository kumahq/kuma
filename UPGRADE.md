This document guides you through the process of upgrading `Kuma`.

First, check if a section named `Upgrade to x.y.z` exists,
with `x.y.z` being the version you are planning to upgrade to.

If such a section does not exist, the upgrade you want to perform
does not have any particular instructions.

## Upcoming release

### `kuma-prometheus-sd`

This component has been removed
after [a long period of deprecation](https://github.com/kumahq/kuma/issues/2851).

### Helm

`ingress.annotations` and `egress.annotations` are deprecated in favour of `ingress.podAnnotations` and `egress.podAnnotations` which is a better name and aligne with the existing `controlPlane.podAnnoations`.

## Upgrade to `1.8.x`

### Kumactl

* `kumactl inspect dataplane --config-dump` was deprecated in favour of `kumactl inspect dataplane --type config-dump`. The behaviour of the new flag is unchanged but you should migrate.
* `kumactl install transparent-proxy --skip-resolv-conf` was deprecated as there's no reason for us to update the `/etc/resolv.conf` of the user.
* `kumactl install transparent-proxy --kuma-cp-ip` was removed as it's not possible to run a DNS server on the cp. 

### Helm

* Under `cni.image`, the default values for `repository` and `registry` have been
changed to agree with the other `image` values.

### CP

* The `/versions` endpoint was removed. This is not something that was reliable enough and version compatibility
is checked inside the DP
* We are deprecating `kuma.io/builtindns` and `kuma.io/builtindnsport` annotations in favour of the clearer `kuma.io/builtin-dns` and `kuma.io/builtin-dns-port`. The behavior of the new annotations is unchanged but you should migrate (a warning is present on the log if you are using the deprecated version).
* Applications that are binding to `localhost` won't be reachable anymore. We are changing the default inbound cluster that was always pointing to `localhost` to `DataplaneIP`. Before upgrade check if your applications are listening on `localhost` and should be exposed. In this case change address on which application listen to `0.0.0.0`. Other option is to define `dataplane.networking.inbound[].serviceAddress` to the address which service is binding. Another way is to disable the new behavior by setting `kuma-cp` configuration `KUMA_DEFAULTS_ENABLE_LOCALHOST_INBOUND_CLUSTERS` to `true` or `defaults.enableLocalhostInboundClusters` to `true`. The last option is going to be removed in further versions.

## Upgrade to `1.7.x`

### Kumactl

* We're deprecating `kumactl install metrics/tracing/logging`, please use `kumactl install observability` instead

### DNS

The `control-plane` no longer hosts a builtin DNS server. You should always rely on the embedded DNS in the dataplane proxy and VIPs can't be used without transparent proxy.

### Timeout policy

'grpc' section is deprecated.
Timeouts for HTTP, HTTP2 and GRPC should be set in 'http' section:

```yaml
tcp: 
  idleTimeout: 1h 
http: # http, http2, grpc
  requestTimeout: 15s 
  idleTimeout: 1h
  streamIdleTimeout: 30m
  maxStreamDuration: 0s
grpc: # DEPRECATED
  streamIdleTimeout: 30m # DEPRECATED, use 'http.streamIdleTimeout'
  maxStreamDuration: 0s # DEPRECATED, use 'http.maxStreamDuration'
```

## Upgrade to `1.6.x`

### Helm

* the Helm chart for this release requires at least Helm version `3.8.0`.
* `controlPlane.resources` is now on object instead of a string. Any existing value should be adapted accordingly.

### Zone egress and ExternalService

When an `ExternalService` has the tag `kuma.io/zone` and `ZoneEgress` is enabled then the request flow will be different after upgrading Kuma to the newest version.
Previously, the request to the `ExternalService` goes through the `ZoneEgress` in the current zone. The newest version flow is different, and when `ExternalService` is defined in a different zone then the request will go through local `ZoneEgress` to `ZoneIngress` in zone where `ExternalService` is defined and leave the cluster through `ZoneEgress` in this cluster. To keep previous behavior, remove the `kuma.io/zone` tag from the `ExternalService` definition.

### Zone egress

Previously, when mTLS was configured and `ZoneEgress` deployed, requests were routed automatically through `ZoneEgress`. Now it's required to
explicitly set that traffic should be routed through `ZoneEgress` by setting `Mesh` configuration property `routing.zoneEgress: true`. The
default value of the property is set to `false` so in case your network policies don't allow you to reach other external services/zone without
using `ZoneEgress`, set `routing.zoneEgress: true`.

```yaml
type: Mesh
name: default
mtls: # mTLS is required for zoneEgress
 [...]
routing:
 zoneEgress: true
```

The new approach changes the flow of requests to external services. Previously when there was no instance of `ZoneEgress` traffic was routed
directly to the destination, now it won't reach the destination.

### Gateway (experimental)

Previously, a `MeshGatewayInstance` generated a `Deployment` and `Service` whose
names ended with a unique suffix. With this release, those objects will have the
same name as the `MeshGatewayInstance`.

### Inspect API

In connection with the changes around `MeshGateway` and `MeshGatewayRoute`, the output
schema of the `<policy-type>/<policy>/dataplanes` has changed. Every policy can
now affect both normal `Dataplane`s and `Dataplane`s configured as builtin gateways.
The configuration for the latter type is done via `MeshGateway` resources.

Every item in the `items` array now has a `kind` property of either:

* `SidecarDataplane`: a normal `Dataplane` with outbounds, inbounds,
  etc.
* `MeshGatewayDataplane`: a `MeshGateway`-configured `Dataplane` with a new
  structure representing the `MeshGateway` it serves.

Some examples can be found in the [Inspect API
docs](https://kuma.io/docs/1.6.x/documentation/http-api/#inspect-api).

## Upgrade to `1.5.x`

### Any type

The `kuma.metrics.dataplane.enabled` and `kuma.metrics.zone.enabled` configurations have been removed.

Kuma always generate the corresponding metrics.

### Kubernetes

- Please migrate your `kuma.io/sidecar-injection` annotations to labels.
  The new version still supports annotation, but to have a guarantee that applications can only start with sidecar, you must use label instead of annotation.
- Configuration parameter `kuma.runtime.kubernetes.injector.sidecarContainer.adminPort` and environment variable `KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_ADMIN_PORT`
  have been deprecated in favor of `kuma.bootstrapServer.params.adminPort` and `KUMA_BOOTSTRAP_SERVER_PARAMS_ADMIN_PORT`.

### Universal

- We removed support for old Ingress (`Dataplane#networking.ingress`) from pre 1.2 days.
  If you are still using it, please migrate to `ZoneIngress` first (see `Upgrade to 1.2.0` section).
- You can't use 0.0.0.0 or :: in `networking.address` most of the time using loopback is what people intended.
- Kuma DP flag `--admin-port` and environment variable `KUMA_DATAPLANE_ADMIN_PORT` have been deprecated, 
  admin port should be specified in Dataplane or ZoneIngress resources.

## Upgrade to `1.4.0`

Starting with this version, the default API server authentication method is user
tokens. In order to continue using client certificates (the previous default
method), you'll need to explicitly set the authentication method to client
certificates. This can be done by setting the `KUMA_API_SERVER_AUTHN_TYPE` variable to
`"clientCerts"`.

See [Configuration - Control plane](https://kuma.io/docs/1.3.1/documentation/configuration/#control-plane)
for how to set this variable.

## Upgrade to `1.3.0`

Starting with this version `Mesh` resource will limit the maximal number of mtls backends to 1, so please make sure your `Mesh` has correct backend applied before the upgrade.

Outbound generated internally are no longer listed in `dataplane.network.outbound[]`. For Kubernetes, they will automatically disappear. For universal to remove them you should recreate your dataplane resources (either with `kumactl apply` or by restarting your services if the dataplanes lifecycle is managed by Kuma).

Kuma 1.3.0 has additional mechanism for tracking data plane proxies and zone statuses in a more reliable way. This mechanism works as a heartbeat and periodically increments the `generation` counter for the Insights. If the overall time for upgrading all Kuma CP instances is more than 5 minutes, then some data plane proxies or zones may become Offline in the GUI, but this doesn't affect real connectivity, only view. This unwanted effect will disappear as soon as all Kuma CP instances will be upgraded to 1.3.0.

## Upgrade to `1.2.1`

When Global is upgraded to `1.2.1` and Zone CP is still `1.2.0`, ZoneIngresses will always be listed as offline.
After Zone CPs are upgraded to `1.2.1`, the status will work again. ZoneIngress status does not affect cross-zone traffic.

## Upgrade to `1.2.0`

One of the changes introduced by Kuma 1.2.0 is renaming `Remote Control Planes` to `Zone Control Planes` and `Dataplane Ingress` to `Zone Ingress`. 
We think this change makes the naming more consistent with the rest of the application and also removes some of unnecessary confusion.

As a result of this renaming, some values and arguments in multizone/kubernetes environment changed. You can read below more.

### Upgrading with `kumactl` on Kubernetes

1. Changes in arguments/flags for `kumactl install control-plane`

   * `--mode` accepts now values: `standalone`, `zone` and `global` (`remote` changed to `zone`)

   * `--tls-kds-remote-client-secret` flag was renamed to `--tls-kds-zone-client-secret`

2. Service `kuma-global-remote-sync` changed to `kuma-global-zone-sync` so after upgrading `global` control plane you have to manually remote old service. For example:

   ```sh
   kubectl delete -n kuma-system service/kuma-global-remote-sync 
   ```

    Hint: It's worth to remember that often at this point the IP address/hostname which is used as a KDS address when installing Kuma Zone Control Planes will change. Make sure that you update the address when upgrading the Remote CPs to the newest version.

### Upgrading with `helm` on Kubernetes

Changes in values in Kuma's HELM chart

* `controlPlane.mode` accepts now values: `standalone`, `zone` and `global` (`remote` changed to `zone`)

* `controlPlane.globalRemoteSyncService` was renamed to `controlPlane.globalZoneSyncService`

* `controlPlane.tls.kdsRemoteClient` was renamed to `controlPlane.tls.kdsZoneClient`

### Suggested Upgrade Path on Universal

1. Zone Control Planes should be started using new environment variables

   * `KUMA_MODE` accepts now values: `standalone`, `zone` and `global` (`remote` changed to `zone`)

     Old:
     ```sh
     KUMA_MODE="remote" [...] kuma-cp run
     ```

     New:
     ```sh
     KUMA_MODE="zone" [...] kuma-cp run
     ```

   * `KUMA_MULTIZONE_REMOTE_ZONE` was renamed to `KUMA_MULTIZONE_ZONE_NAME`

     Old:
     ```sh
     KUMA_MULTIZONE_REMOTE_ZONE="remote-1" [...] kuma-cp run
     ```

     New:
     ```sh
     KUMA_MULTIZONE_ZONE_NAME="remote-1" [...] kuma-cp run
     ```

   * `KUMA_MULTIZONE_REMOTE_GLOBAL_ADDRESS` was renamed to `KUMA_MULTIZONE_ZONE_GLOBAL_ADDRESS`

     Old:
     ```sh
     KUMA_MULTIZONE_REMOTE_GLOBAL_ADDRESS="grpcs://localhost:5685" [...] kuma-cp run
     ```

     New:
     ```sh
     KUMA_MULTIZONE_ZONE_GLOBAL_ADDRESS="grpcs://localhost:5685" [...]  kuma-cp run
     ```

   * `KUMA_MULTIZONE_REMOTE_KDS_ROOT_CA_FILE` was renamed to `KUMA_MULTIZONE_ZONE_KDS_ROOT_CA_FILE`

     Old:
     ```sh
     KUMA_MULTIZONE_REMOTE_KDS_ROOT_CA_FILE="/rootCa" [...] kuma-cp run
     ```

     New:
     ```sh
     KUMA_MULTIZONE_ZONE_KDS_ROOT_CA_FILE="/rootCa" [...] kuma-cp run
     ```

   * `KUMA_MULTIZONE_REMOTE_KDS_ROOT_CA_FILE` was renamed to `KUMA_MULTIZONE_ZONE_KDS_ROOT_CA_FILE`

     Old:
     ```sh
     KUMA_MULTIZONE_REMOTE_KDS_REFRESH_INTERVAL="9s" [...] kuma-cp run
     ```

     New:
     ```sh
     KUMA_MULTIZONE_ZONE_KDS_REFRESH_INTERVAL="9s" [...] kuma-cp run
     ```

2. Dataplane Ingress resource should be replaced with ZoneIngress resource:

    Old:
    ```yaml
    type: Dataplane
    name: dp-ingress
    mesh: default
    networking:
      address: <ADDRESS>
      ingress:
        publicAddress: <PUBLIC_ADDRESS>
        publicPort: <PUBLIC_PORT>
      inbound:
      - port: <PORT>
        tags:
          kuma.io/service: ingress
    ```

    New:
    ```yaml
    type: ZoneIngress
    name: zone-ingress
    networking:
      address: <ADDRESS>
      port: <PORT>
      advertisedAddress: <PUBLIC_ADDRESS>
      advertisedPort: <PUBLIC_PORT>
    ```

    NOTE: ZoneIngress resource is a global scoped resource, it's not bound to a Mesh
    The old Dataplane resource is still supported but it's considered deprecated and will be removed in the next major version of Kuma


3. Since ZoneIngress resource is not bound to a Mesh, it requires another token type that is bound to a Zone:
   
    ```shell
    kumactl generate zone-ingress-token --zone=zone-1 > /tmp/zone-ingress-token
    ```

4. `kuma-dp run` command should be updated with a new flag `--proxy-type=ingress`:

    ```sh
    kuma-dp run \
      --proxy-type=ingress \
      --dataplane-token-file=/tmp/zone-ingress-token \
      --dataplane-file=zone-ingress.yaml
    ```


## Upgrade to `1.1.0`

The major change in this release is the migration to XDSv3 for the `kuma-cp` to `envoy` data plane proxy communication. The
previous XDSv2 is still available and will continue working. All the existing data plane proxies will still use XDSv2 until
being restarted. The newly deployed `kuma-dp` instances will automatically get bootstrapped to XDSv3. In case that needs to be
changed, `kuma-cp` needs to be started with `KUMA_BOOTSTRAP_SERVER_API_VERSION=v2`.

With Kuma 1.1.0, the `kuma-cp` will installs default [retry](https://kuma.io/docs/1.1.0/policies/retry/) and [timeout](https://kuma.io/docs/1.1.0/policies/timeout/) policies for each new
created Mesh object. The pre-existing meshes will not automatically get these default policies. If needed, they should be created accordingly.

This version removes the deprecated `--dataplane` flag in `kumactl generate dataplane-token`, please consider migrating to use `--name` instead.

## Upgrade to `1.0.0`

This release introduces a number of breaking changes. If Kuma is being deployed in production we strongly suggest to backup the current configuration, tear down the whole cluster and zones, and install in a clean setup. However, we enumerate the details of these changes below.

### Suggested Upgrade Path on Kubernetes
 * Drop k8s 1.13 support

    Take this into account if you run Kuma on an old Kubernetes version.

 * `kumactl` merged `install ingress` into `install control-plane`

    This change impacts any deployment pipelines that are based on `kumactl` and are used for multi-zone deployments.

 * Change policies on K8S to scope global

    All the CRDs are now in the global scope, therefore all policies need to be backed up. The relevant CRDs need to be deleted, which will clear all the policies. After the upgrade, you can apply the policies again. We do recommend to keep all the Kuma Control Planes down while doing these operations.

 * Autoconfigure single cert for all services

    Deployment flags for providing TLS certificates in Helm and `kumactl` have changed, refer to the relevant [documentation](https://github.com/kumahq/kuma/blob/release-1.0/deployments/charts/kuma/README.md#values) to verify the new naming.

 * Create default resources for Mesh

    The following default resources will be created upon the first start of Kuma Control Plane
        - default signing key
        - default [Allow All traffic permission](https://kuma.io/docs/1.0.0/policies/traffic-permissions/#traffic-permissions) policy `allow-all-<mesh name>`
        - Default [Allow All traffic route](https://kuma.io/docs/1.0.0/policies/traffic-route/#default-trafficroute) policy `allow-all-<mesh name>`
    
    Please verify if this conflicts with your deployment and expected policies.

 * New Multizone deployment flow

    Deploying Multizone clusters is now simplified, please refer to the deployment [documentation](https://kuma.io/docs/1.0.0/documentation/deployments/#multi-zone-mode) of the updated procedure.
   
 * Improved control plane communication security
   
    Kuma Control Plane exposed ports are reduced, please revise the [documentation](https://kuma.io/docs/1.0.0/documentation/networking/#kuma-cp-ports) for detailed list.
    Consider reinstalling the metrics due to the port changes in Kuma Prometheus SD.
 
 * Traffic route format
 
    The format of the [TrafficRoute](https://kuma.io/docs/1.0.0/policies/traffic-route) has changed. Please check the documentation and adapt your resources. 

### Suggested Upgrade Path on Universal
 * Get rid of advertised hostname
    `KUMA_GENERAL_ADVERTISED_HOSTNAME` was removed and not needed now.
 
 * Autoconfigure single cert for all services
    Deployment flags for providing TLS certificates in Helm and `kumactl` have changed, refer to the [documentation](https://github.com/kumahq/kuma/blob/release-1.0/pkg/config/app/kuma-cp/kuma-cp.defaults.yaml) to verify the new naming.

 * Create default resources for Mesh
    
    The following default resources will be created upon the first start of Kuma Control Plane
        - default signing key
        - default [Allow All traffic permission](https://kuma.io/docs/1.0.0/policies/traffic-permissions/#traffic-permissions) policy `allow-all-<mesh name>`
        - Default [Allow All traffic route](https://kuma.io/docs/1.0.0/policies/traffic-route/#default-trafficroute) policy `allow-all-<mesh name>`
    
    Please verify if this conflicts with your deployment and expected policies.

* New Multizone deployment flow

    Deploying Multizone clusters is now simplified, please refer to the deployment [documentation](https://kuma.io/docs/1.0.0/documentation/deployments/#multi-zone-mode) of the updated procedure.
   
 * Improved control plane communication security
   
    `kuma-dp` invocation has changed and now [allows](https://kuma.io/docs/1.0.1/documentation/dps-and-data-model/#dataplane-entity) for a more flexible usage leveraging automated, template based Dataplane resource creation, customizable data-plane token boundaries and additional CA ceritficate validation for the Kuma Control plane boostrap server.
    Kuma Control Plane exposed ports are reduced, please revise the [documentation](https://kuma.io/docs/1.0.0/documentation/networking/#kuma-cp-ports) for detailed list.
 
  * Traffic route format
  
     The format of the [TrafficRoute](https://kuma.io/docs/1.0.0/policies/traffic-route) has changed. Please check the documentation and adapt your resources. 

 
## Upgrade to `0.7.0`
Support for `kuma.io/sidecar-injection` annotation. On Kubernetes change the namespace resources that host Kuma mesh services with the aforementioned annotation and delete the label. 

Prefix the Kuma built-in tags with `kuma.io/` as follows: `kuma.io/service`, `kuma.io/protocol`, `kuma.io/zone`.

### Suggested Upgrade Path on Kubernetes

Update the applied policy tag selector to include the `kuma.io/` prefix. A sample traffic resource follows:

```yaml
apiVersion: kuma.io/v1alpha1
kind: TrafficPermission
mesh: default
metadata:
  namespace: default
  name: allow-all-traffic
spec:
  sources:
    - match:
        kuma.io/service: '*'
  destinations:
    - match:
        kuma.io/service: '*'
```

The Kuma Control Plane will update the relevant Dataplane resources accordingly

### Suggested Upgrade Path on Universal

Update the applied policy tag selector to include the `kuma.io/` prefix. A sample traffic resource follows:

```yaml
type: TrafficPermission
name: allow-all-traffic
mesh: default
sources:
  - match:
      kuma.io/service: '*'
destinations:
  - match:
      kuma.io/service: '*'
```

Update the dataplane resources with the new tag format as well. Example:

```bash
echo "type: Dataplane
mesh: default
name: redis-1
networking:
  address: 192.168.0.1
  inbound:
  - port: 9000
    servicePort: 6379
    tags:
      kuma.io/service: redis" | kumactl apply -f -
```

This release changes the way that Distributed and Hybrid Kuma Control planes are deployed. Please refer to the [documentation](https://kuma.io/docs/0.7.0/documentation/deployments/#usage) for more details.

## Upgrade to `0.6.0`

[Passive Health Check](https://kuma.io/docs/0.5.1/policies/health-check/) were removed in favor of [Circuit Breaking](https://kuma.io/docs/0.6.0/policies/circuit-breaker/).

Format of Active Health Check changed from :
```yaml
apiVersion: kuma.io/v1alpha1
kind: HealthCheck
mesh: default
metadata:
  namespace: default
  name: web-to-backend-check
mesh: default
spec:
  sources:
  - match:
      service: web
  destinations:
  - match:
      service: backend
  conf:
    activeChecks:
      interval: 10s
      timeout: 2s
      unhealthyThreshold: 3
      healthyThreshold: 1
    passiveChecks:
      unhealthyThreshold: 3
      penaltyInterval: 5s
```
to 
```yaml
apiVersion: kuma.io/v1alpha1
kind: HealthCheck
mesh: default
metadata:
  namespace: default
  name: web-to-backend-check
mesh: default
spec:
  sources:
  - match:
      service: web
  destinations:
  - match:
      service: backend
  conf:
    interval: 10s
    timeout: 2s
    unhealthyThreshold: 3
    healthyThreshold: 1
```

### Suggested Upgrade Path on Kubernetes

In the new Kuma version serivce tag format has been changed. Instead of `backend.kuma-demo.svc:5678` service tag will look like this `backend_kuma-demo_svc_5678`. This is a breaking change and Policies should be updated to be compatible with the new Kuma version.

Please re-install Prometheus via `kubectl install metrics` and make sure that `skipMTLS` is set to `false` or omitted.
```yaml
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: default
spec:
  metrics:
    enabledBackend: prometheus-1
    backends:
    - name: prometheus-1
      type: prometheus
      conf:
        skipMTLS: false
```

### Suggested Upgrade Path on Universal

Make sure that `skipMTLS` is set to `true`.

```yaml
type: Mesh
name: default
metrics:
  enabledBackend: prometheus-1
  backends:
  - name: prometheus-1
    type: prometheus
    conf:
      skipMTLS: true
```


## Upgrade to `0.5.0`
### Suggested Upgrade Path on Kubernetes

#### Mesh resource format changes

The Mesh resource format in Kubernetes changed from
```yaml
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: default
spec:
  mtls:
    enabled: true
    ca:
      builtin: {}
  metrics:
    prometheus: {}
  logging:
    backends:
    - name: file-1
      file:
        path: /var/log/access.log
  tracing:
    backends:
    - name: zipkin-1
      zipkin:
        url: http://zipkin.local:9411/api/v1/spans
```
to
```yaml
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: default
spec:
  mtls:
    enabledBackend: ca-1
    backends:
    - name: ca-1
      type: builtin
  metrics:
    enabledBackend: prom-1
    backends:
    - name: prom-1
      type: prometheus
  logging:
    backends:
    - name: file-1
      type: file
      conf:
        path: /var/log/access.log
  tracing:
    backends:
    - name: zipkin-1
      type: zipkin
      conf:
        url: http://zipkin.local:9411/api/v1/spans
```

#### Removing `kuma-injector`

Kuma 0.5.0 ships with `kuma-injector` embedded into the `kuma-cp`, which makes its previously created resources obsolete and potentially
 can cause problems with the deployments. Before deploying the new version, it is strongly advised to run a cleanup script [kuma-0.5.0-k8s-remove_injector_resources.sh](tools/migrations/0.5.0/kuma-0.5.0-k8s-remove_injector_resources.sh).
 
 NOTE: if Kuma was deployed in a namespace other than `kuma-system`, please run `export KUMA_SYSTEM=<othernamespace` before running the cleanup script.

#### Kuma resources `ownerReferences` 
Kuma 0.5.0 introduce webhook for setting `ownerReferences` to the Kuma resources. If you have some 
Kuma resources in your k8s cluster, then you can use our script [kuma-0.5.0-k8s-set_owner_references.sh](tools/migrations/0.5.0/kuma-0.5.0-k8s-set_owner_references.sh) 
in order to properly set `ownerReferences` .

### Suggested Upgrade Path on Universal

#### Mesh resource format changes
Mesh format on Universal changed from
```yaml
type: Mesh
name: default
mtls:
  enabled: true
  ca:
    builtin: {}
metrics:
  prometheus: {}
logging:
  backends:
  - name: file-1
    file:
      path: /var/log/access.log
tracing:
  backends:
  - name: zipkin-1
    zipkin:
      url: http://zipkin.local:9411/api/v1/spans
```
to
```yaml
type: Mesh
name: default
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
metrics:
  enabledBackend: prom-1
  backends:
  - name: prom-1
    type: prometheus
logging:
  backends:
  - name: file-1
    type: file
    conf:
      path: /var/log/access.log
tracing:
  backends:
  - name: zipkin-1
    type: zipkin
    conf:
      url: http://zipkin.local:9411/api/v1/spans
```

## Upgrade to `0.4.0`

### Suggested Upgrade Path on Kubernetes

No additional steps are needed.

### Suggested Upgrade Path on Universal

#### Migrations

Kuma 0.4.0 introduces DB Migrations for Postgres therefore before running the new version of Kuma, run the kuma-cp migration command.
```
kuma-cp migrate up
```
Remember to provide config just like in `kuma-cp run` command.
All existing data will be preserved.

#### New Dataplane Entity format

Kuma 0.4.0 introduces new Dataplane entity format to improve readability as well as add support for scraping metrics of Gateway Dataplanes. 

Here is example of migration to the new format.

**Dataplane**

Old format
```yaml
type: Dataplane
mesh: default
name: web-01
networking:
  inbound:
  - interface: 192.168.0.1:21011:21012
    tags:
      service: web
  outbound:
  - interface: :3000
    service: backend
```

New format
```yaml
type: Dataplane
mesh: default
name: web-01
networking:
  address: 192.168.0.1
  inbound:
  - port: 21011
    servicePort: 21012
    tags:
      service: web
  outbound:
  - port: 3000
    service: backend
```

**Gateway Dataplane**

Old format
```yaml
type: Dataplane
mesh: default
name: kong-01
networking:
  gateway:
    tags:
      service: kong
```

New format
```yaml
type: Dataplane
mesh: default
name: kong-01
networking:
  address: 192.168.0.1
  gateway:
    tags:
      service: kong
```

Although the old format is still supported, it is recommended to migrate since the support for it will be dropped in the next major version of Kuma.

## Upgrade to `0.3.1`

### List of breaking changes

`kuma policies`:
* `Mesh` CRD on Kubernetes is now Cluster-scoped
* `TrafficLog` policy is applied differently now: instead of applying all `TrafficLog` policies that match to a given `outbound` interface of a `Dataplane`, only a single the most specific `TrafficLog` policy is applied

`kumactl`:
* a few options in `kumactl config control-planes add` command have been renamed:
  * `--dataplane-token-client-cert` has been renamed into `--admin-client-cert`
  * `--dataplane-token-client-key` has been renamed into `--admin-client-key`

### Suggested Upgrade Path on Kubernetes

* Users on Kubernetes will have to re-install `Kuma`:

  1. Export all `Kuma` resources
     ```shell
     kubectl get meshes,trafficpermissions,trafficroutes,trafficlogs,proxytemplates --all-namespaces -oyaml > backup.yaml
     ```
  2. Uninstall previous version of `Kuma Control Plane`
     ```shell
     # using previous version of `kumactl`

     kumactl install control-plane | kubectl delete -f -
     ```
  3. Install new version of `Kuma Control Plane`
     ```shell
     # using new version of `kumactl`

     kumactl install control-plane | kubectl apply -f -
     ```
  4. Re-apply `Kuma` resources back again
     ```shell
     kubectl apply -f backup.yaml
     ```

### Suggested Upgrade Path on Universal

* Those users who used `--dataplane-token-client-cert` and `--dataplane-token-client-key` command line options in the past will have to re-run

   ```
   kumactl config control-planes add
   ```

   this time with

    ```shell
    --admin-client-cert <CERT> --admin-client-cert <KEY> --overwrite
    ```
* all components of `Kuma Control Plane` - `kuma-cp`, `kuma-dp`, `envoy` - have to be re-deployed
