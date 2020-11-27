This document guides you through the process of upgrading `Kuma`.

First, check if a section named `Upgrade to x.y.z` exists,
with `x.y.z` being the version you are planning to upgrade to.

If such a section does not exist, the upgrade you want to perform
does not have any particular instructions.

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
