# Hostnames and VIP management in MeshService

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/6331

## Context and Problem Statement

With introduction of MeshService, we need to have a mechanism to assign VIPs and hostnames to MeshServices.
Currently, VIPs are managed in a ConfigMap per Mesh. Hostnames are managed using VirtualOutbound.

## Considered Options

* VIPs management
* Hostname Generator
* Static Hostname template in CP config

## Decision Outcome

Chosen option: "Hostname Generator" and "VIPs management" since it's the most flexible option.

## Pros and Cons of the Options

### VIPs management

We have to assign a VIP for each MeshService to be able to address it using transparent proxy.

MeshService object contains VIPs in status.
```yaml
status: # managed by CP. Not shared cross zone, but synced to global
  vips:
  - ip: <kube_cluster_ip> # or kuma VIP
    type: Kubernetes # | Kuma
```

If MeshService was created based on Kubernetes Service, we can just take ClusterIP, but otherwise we need to assign each MeshService a Kuma VIP.
Kuma VIP is a non-routable IP used only by Envoy to indicate destination (like a host header, but on L4). We currently do this with 240.0.0.0/4 CIDR.

Kuma VIPs are assigned asynchronously with MeshService creation. Assigning Kuma VIPs as sync operation on MeshService creation would create a race condition on VIP allocation.

VIP assigning process (single-threaded):
* List MeshServices
* If a MeshService has no VIP, assign a Kuma VIP. Kubernetes Services would have a VIP already (ClusterIP), so we don’t need to assign an extra VIP.
* To assign a VIP, we build an IPAM state reserving every address we already assigned and allocate the next address.
  To solve the problem of “assigning immediately” a VIP of an old service to a new Service we can also store in memory recently removed VIPs for X seconds and set TTL of DNS responses for X seconds.
  The current value of 30s is too long, especially that the DNS server is run locally in kuma-dp. Leader election should take more than X seconds, so if we were to lose this list, it’s not a problem.

Following this process, we don’t have to store VIPs in any central place like we do currently (VIP ConfigMap for each Mesh).
We also assign VIPs for ExternalServices, but this can work as long as we consider both ExternalServices and MeshServices for this process.

To avoid collision with an existing VIPs mechanism, we can take a completely new CIDR (241.0.0.0/4).

It's not expected for service to have multiple VIPs, but we want expose this possibility for the future (therefore `vips` field is a list).

### Hostname Generator

MeshService has to be addressable with a hostname.
In case of services in a Kubernetes cluster, we don't have to do anything because we can just rely on Kube DNS that resolves to a Kube VIP we placed for MeshService.

In case of services in a Universal cluster, currently we provide two ways:
* `<kuma.io/service>.mesh` address that resolves to Kuma VIP
* VirtualOutbound resource to provide custom addresses

Hostnames should be managed only by MeshOperator. Otherwise, we get into security issues such as one service owner can override a hostname of other service.
Hostnames are created by convention in the majority of cases - Kubernetes hostnames are also such example.
It is still possible for service owner to create unique hostname for their app by asking reaching out to mesh operator.

To manage hostname, we introduce new object called `HostnameGenerator`

```yaml
type: HostnameGenerator
name: k8s-zone-hostnames
spec:
  meshServiceSelector:
    matchLabels:
      kuma.io/zone: east 
  template: {{ label "k8s.kuma.io/service-name" }}.{{ label "k8s.kuma.io/namespace" }}.svc.cluster.{{ label "kuma.io/zone" }}
```

`HostnameGenerator` is a namespaced-scoped object on Kubernetes, but can only be applied in system namespace.
It can be applied on global CP, so it's synced to all zones, or it can be applied on specific zones.

For example, given this mesh service

```yaml
type: MeshService
name: redis.demo-app
labels:
  k8s.kuma.io/service-name: redis
  k8s.kuma.io/namespace: demo-app
  kuma.io/zone: east
spec: ...
```
We would generate such hostname `redis.demo-app.svc.cluster.east`

Possible template functions/keys are:
* `{{ name }}` - name of the MeshService
* `{{ label "x" }}` - value of label `x`.

If the template cannot be resolved (label is missing), the hostname won't be generated.

Generated hostnames will be placed in the status of MeshService
```yaml
status:
  addresses:
    - hostname: redis.demo-app.svc.cluster.east
      status: Available # | NotAvailable if there was a collision
      origin:
        kind: HostnameGenerator
        name: "k8s-zone-hostnames" # name of the HostnameGenerator
      reason: "not available because of the clash with ..." # reason if there was a problem
```

**Precedence and collisions**
While hostnames are managed by Mesh Operator and it's unlikely to hit a hostname collision, it's possible.
HostnameGenerators created in a zone have precedence over those created on global. As a fallback we can order resources by creation time and as a final fallback lexicographically.

**Defaults**
Control plane will ship with default `HostnameGenerator`s, but this will be covered in MeshService multizone MADR and universal UX MADR.

### Static Hostname template in CP config

Instead of creating a new object, we could just have a static config in CP. However:
* In multizone deployment, you need to configure every single CP with it.
  With HostnameGenerator, we define it once in global cp, and it's synced to every zone
* `meshServiceSelector` would be overkill in CP config

Implementing HostnameGenerator does not cancel the idea that we could implement this.
