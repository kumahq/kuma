This document guides you through the process of upgrading `Kuma`.

First, check if a section named `Upgrade to x.y.z` exists,
with `x.y.z` being the version you are planning to upgrade to.

If such a section does not exist, the upgrade you want to perform
does not have any particular instructions.

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
