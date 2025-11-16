# MeshService multizone UX

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/9706

## Context and Problem Statement

With the introduction of `MeshService`, we need to decide how does it work with multizone deployments.

There are two main use cases for multicluster connectivity:

**Access service from another cluster**
Service A is deployed only in zone X and the service in zone Y wants to consume it.

**Replicate service over multiple clusters**
Service A is deployed across multiple clusters for high availability.
This strategy is used with "lift and shift" - migrating a workload from on prem to kubernetes.

The current state is that a service is represented by `kuma.io/service` tag and as long as the tag is the same between zones, we consider this the same service.
It means that a Kube service with the same name and namespace in two clusters is considered the same. 
By default, if you don't have `MeshLoadbalancingStrategy` in the cluster, we load balance between local and remote zones. 

This has a security disadvantage that if you take over the zone, you can advertise any service.

## Considered Options

* Access service from another cluster use case
  * MeshService synced to other zones
  * Service export / import
    * Export field
    * MeshTrafficPermission
* Replicated services use case
  * MeshService with dataplaneTags selector applied on global CP
  * MeshMultiZoneService
  * MeshService with meshServiceSelector selector applied on global CP

## Decision Outcome

Chosen options:
* Access service from another cluster use case
  * MeshService synced to other zones, because it's the only way
  * Export field, because it's more explicit and more managable
* Replicated services use case
  * MeshMultiZoneService, with potentially diverging to "MeshService with meshServiceSelector selector applied on global CP" based on the future policy matching MADR.

## Pros and Cons of the Options

### Access service from another cluster use case

#### MeshService synced to other zones

MeshService is an object created on the zone that matches a set of proxies in this specific zone.
This object is the synced to the global control plane.

To solve the first use case we can then sync MeshService to other zones.
Let's say we have the following MeshService in zone `east`
```yaml
kind: MeshService
metadata:
  name: redis
  namespace: redis-system
labels:
  kuma.io/display-name: redis
  k8s.kuma.io/namespace: redis-system
  kuma.io/zone: east
spec:
  selector:
    dataplane:
      tags:
        app: redis
        k8s.kuma.io/namespace: redis-system
  ports: ...
status: ...
```

This will be synced to global and look like this
```yaml
kind: MeshService
metadata:
  name: redis-xyz12345 # we add hash suffix of (mesh, zone, namespace)
  namespace: kuma-system # note that it's synced to system namespace
  labels:
    kuma.io/display-name: redis # original name of the object on the Kubernetes
    k8s.kuma.io/namespace: redis-system
    kuma.io/zone: east
    kuma.io/origin: zone
spec:
  selector:
    dataplane:
      tags:
        app: redis
        k8s.kuma.io/namespace: redis-system
        kuma.io/zone: east-1
  ports: ...
status: ...
```

This will be then synced to other zones in this form, which is the same syncing strategy that we already have with ZoneIngress.

Because of the hashes and syncing only to system namespace, we avoid clashes of synced services and existing services in the cluster.

**Hostname and VIP management**

Advertising the service is only part of the solution. We also need to provide a way to consume this service.

We cannot reuse:
* existing hostnames like `redis.redis-system.svc.cluster.local`, because there might be already a redis in cluster `east.
* ClusterIP, because each cluster network is isolated. We might clash with IPs.

Therefore `status` object should not be synced from zone `east` to zone `west`.

To provide hostname, we can use `HostnameGenerator` described in MADR #046.
```yaml
---
type: HostnameGenerator
name: k8s-zone-hostnames
spec:
  targetRef:
    kind: MeshService
  template: {{ label "kuma.io/display-name" }}.{{ label "k8s.kuma.io/namespace" }}.svc.mesh.{{ label "kuma.io/zone" }}
---
type: HostnameGenerator
name: uni-zone-hostnames
spec:
  targetRef:
    kind: MeshService
  template: {{ label "kuma.io/service" }}.svc.mesh.{{ label "kuma.io/zone" }}
```
so that the synced service in zone `west` gets this hostname `redis.redis-system.svc.mesh.east`.

To assign Kuma VIP, we follow the process described in MADR #046.

When generating the XDS configuration for the client in zone `west`, we check if MeshService has `kuma.io/zone` label that is not equal to local zone.
If so, then we need to point Envoy Cluster to endpoint for ZoneIngress of `west`.
This way we get rid of `ZoneIngress#availableServices`, which was one of the initial motivation of introducing MeshService (see MADR #039).

#### Service export / import

One thing to consider is whether we should sync all MeshServices to other clusters by default.
It's reasonable to consider that only a subset of services is consumed cross zone.

Trimming it could benefit in:
* performance. Fewer objects to store and sync around.
  However, changes to MeshServices spec are rather rare. We don't consider this to be a significant perf issue.
* security. Unexported Mesh Service would not even be exposed with ZoneIngress.

Explicit import seems to have very little value, because clients needs to explicitly use service from a specific zone.

##### Export field

We can introduce `spec.export.mode` field to MeshService with values `All`, `None`.
We want to have nested `mode` in `export`, so we can later extend api if needed. For example `spec.export.mode: Subset` and `spec.export.list: zone-a,zone-b`.
On Kubernetes, this would be controlled via `kuma.io/export-mode` annotation on `Service` object.
On Universal, where MeshService is synthesized from Dataplane objects, this would be controlled via `kuma.io/export-mode` tag on a Dataplane object.

If missing, then default behavior from kuma-cp configuration will be applied. The default CP configuration is `spec.export.mode: All`.
Then it's up to a mesh operator if they want to sync everything by default or opt-in exported services.

##### MeshTrafficPermission

Similarly to what we do with reachable services (see [MADR 031](https://github.com/kumahq/kuma/blob/master/docs/madr/decisions/031-automatic-rechable-services.md)), 
We could implicitly check if service can be consumed by proxy in other zone and only then sync the service.
This however is difficult, we either need to require `kuma.io/zone` for it to work or check which proxies are in other zones which is expensive.

### Replicated services use case

As opposed to the current solution with `kuma.io/service`, redis in zone `east` is a different service than redis in zone `west`.
There are two ways we could go about aggregating those services.

#### MeshService with dataplaneTags selector applied on global CP

To achieve replicated service, we define a MeshService on global CP.
This MeshService is then synced to all zones.

```yaml
kind: MeshService
metadata:
  name: redis-xzh64d
  namespace: kuma-system
  labels:
    team: db-operators
    kuma.io/mesh: default
    kuma.io/origin: global
spec:
  selector:
    dataplane:
      tags: # it has no kuma.io/zone, so it aggregates all redis from redis-system from all zones
        app: redis
        k8s.kuma.io/namespace: redis-system
status:
  ports:
  - port: 1234
    protocol: tcp
  addresses:
  - hostname: redis.redis-system.svc.mesh.global
    status: Available
    origin: global-generator
  vips:
  - ip: 241.0.0.1
    type: Kuma
  zones: # computed on global. The rest of status fields are computed on the zone
  - name: east
  - name: west
```

The disadvantages:
* We need to do a computation of list of zones on global cp, because we don't sync all Dataplanes across all zones.
  This can be expensive because global has list of all data plane proxies.
* We wouldn't have a capability of apply "multizone" service on specific zone, because we don't have all DPPs in all zones.

**Hostname and VIP management**
We can provide hostnames by selecting by `kuma.io/origin` label.

```yaml
type: HostnameGenerator
name: global-generator
spec:
  targetRef:
    kind: MeshServices
    labels:
      kuma.io/origin: global
  template: {{ label "kuma.io/display-name" }}.{{ label "k8s.kuma.io/namespace" }}.svc.mesh.global
```

#### MeshMultiZoneService

If we assume that replicated service is aggregation of multiple services, we can define a specific object for it.
`MeshMultiZoneService` selects services instead of proxies. 

```yaml
kind: MeshMultiZoneService
metadata:
  name: redis-xzh64d
  namespace: kuma-system
  labels:
    team: db-operators
    kuma.io/mesh: default
spec:
  selector:
    meshService:
      labels:
        k8s.kuma.io/service-name: redis
        k8s.kuma.io/namespace: redis-system
status: # computed on the zone
  ports:
  - port: 1234
    protocol: tcp
  addresses:
  - hostname: redis.redis-system.svc.mesh.global
    status: Available
    origin: global-generator
  vips:
  - ip: 242.0.0.1 # separate CIDR
    type: Kuma
  zones:
  - name: east
  - name: west
```

Considered names alternatives:
* MeshGlobalService - we want to avoid using Global as it may indicate non-mesh specific object.
* GlobalMeshService

The advantages:
* We don't need to compute anything on global cp. In the example above, both redis MeshServices have been synced to the zone.
* We don't need to traverse over all data plane proxies
* Users don't need to redefine ports. In case there is difference. We take only common ports.
* Policy matching easier to implement (will be covered more in policy matching MADR)
* We can expose a capability to apply this locally on one zone.

The disadvantages:
* A separate resource
* Cannot paginate both services on one page in the GUI

**Hostname and VIP management**
HostnameGenerator has to select `MeshMultiZoneService`, so we need `targetRef` that can accept both `MeshService` and `MeshMultiZoneService` 

```yaml
type: HostnameGenerator
name: global-generator
spec:
  targetRef:
    kind: MeshMultiZoneService
  template: {{ label "kuma.io/display-name" }}.{{ label "k8s.kuma.io/namespace" }}.svc.mesh.global
```

#### MeshService with serviceSelector selector applied on global CP

Alternatively we can combine both solutions to have `MeshService`, but with `selector.meshService.labels`

```yaml
kind: MeshService
metadata:
  name: redis-xzh32a
  namespace: kuma-system
  labels:
    team: db-operators
    kuma.io/mesh: default
    kuma.io/origin: global
spec:
  selector:
    meshService: # we select services by labels, not dataplane objects
      labels:
        k8s.kuma.io/service-name: redis
        k8s.kuma.io/namespace: redis-system
```

The advantages:
* Same as `MeshMultiZoneService`

The disadvantages:
* It might be confusing that MeshService may select both data plane proxies or services
* Policy matching might be more convoluted (will be covered more in policy matching MADR).

**Hostname and VIP management**
Same as with `MeshService with dataplaneTags selector applied on global CP`.

### Replicated service between Kubernetes and Universal deployments

To treat the service on Kubernetes and Universal as the same service, we need to have common tags on them. We can either:
* Add k8s specific labels for Universal Dataplane
* Add common label like (`myorg.com/service`) to both of them.
  In this case we need to trust that it cannot be freely modified by service owners, otherwise it's easy to impersonate for a service.

### Explicit namespace sameness

Whether we choose `MeshMultiZoneService` or `MeshService` for replicated services use case it will require extra step than the existing implementation.
This extra step is defining an object that represent global service.
It is a disadvantage if user only wants to use global services, they would need to create every single one of them.
If we see demand for it, we can create a controller on global CP that would generate global services that has the same set of defined tags (for example: `kuma.io/display-name` + `k8s.kuma.io/namespace`).

## Multicluster API

[Multicluster API](https://github.com/kubernetes/enhancements/blob/master/keps/sig-multicluster/1645-multi-cluster-services-api/README.md) solves a way to manage services across multiple clusters. The way it works is by introducing two objects - `ServiceImport` and `ServiceExport`.
To expose a service, a user needs to apply ServiceExport

```yaml
apiVersion: multicluster.k8s.io/v1alpha1
kind: ServiceExport
metadata:
  name: redis
  namespace: kuma-demo
```

and then in every cluster that wants to consume this we need to apply `ServiceImport`
```yaml
apiVersion: multicluster.k8s.io/v1alpha1
kind: ServiceImport
metadata:
  name: redis
  namespace: kuma-demo
spec:
  ips:
  - 42.42.42.42
  type: "ClusterSetIP"
  ports:
  - protocol: TCP
    port: 6379
status:
  clusters:
  - cluster: east
```

Service then can be addressed by `redis.kuma-demo.svc.clusterset.local`

ServiceImport is managed by MCS-controller which means that it has similar problem that we had - if someone take over one cluster they can advertise their service.

* Export field controlled by annotation described in MADR is an equivalent of ServiceExport.
  The difference is that we can do it by default, whereas in Multicluster API, it's explicit.
* Synced MeshService is similar to ServiceImport.
  The difference is that the object is synced to the mesh system namespace, and synced service is bound to a specific zone.
  We get access to services of a **specific** cluster without any extra action of exporting.
  For example `redis-kuma-demo.svc.cluster.east`
  There is no way to do this in MC API.
* MeshMultiZoneService is an equivalent of `ServiceImport`.
  It is also "zone agnostic" - the client do not care in which zone the Service is deployed.
  Just like ServiceImport/ServiceExport it requires an extra manual step - creating an object.
  You don't need to create this in every cluster, you can create MeshMultiZoneService on Global CP and it's synced to every cluster.
  When applied on Global CP, it's synced to the system namespace.

MeshService + MeshMultiZoneService seems to be more flexible approach.

This approach does not block us from adopting multicluster API in the future if it gains traction. We could either:
* Convert ServiceImport to MeshMultiZoneService.
  Write the controller that listens on ServiceExport to annotate Service with `kuma.io/export`
* Natively switch to it as a primary objects. This might be tricky when we consider supporting Universal deployments.
