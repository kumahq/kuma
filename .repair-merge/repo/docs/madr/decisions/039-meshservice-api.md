# MeshService API

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/6331

## Context and Problem Statement

Currently, our definition of "service" is derived from `kuma.io/service` tag placed in a Dataplane object. It has a several problems.

**Scalability**
* There are a number of objects whose size grows `O(#services)`: ServiceInsights, VIP ConfigMap, ZoneIngress availableServices. At scale storing objects of this size becomes a problem. There are limits for a size of an object in Kube API server.
* Listing available services requires iterating through all Dataplane objects.
* Separating Envoy endpoint generation from cluster generation. Right now a single change of Dataplane object would trigger cluster generation even though the cluster may not have changed.

**Indirection**
* There is no way to add metadata to the service. You can add metadata to a workload (Dataplane).
  While we can compute shared metadata of workloads of same `kuma.io/service` and consider this a service metadata, there is a confusing indirection.

## Considered Options

* Shard objects
* MeshService

## Decision Outcome

Chosen option: "MeshService", because it helps us to provide better foundation to build things on top of it. 

### Positive Consequences

* Solved scalability problems

### Negative Consequences

* A pretty big effort (implementation, a strategy of migration).
* Temporarily increased complexity of a product until the migration is done.

## Pros and Cons of the Options

### Shard objects

We could shard objects that are `O(#services)`, although it's not clear by what key.
While this solves the problem of scalability, it does not solve other "side" problems that can be solved with a dedicated MeshService object.

### MeshService

We introduce a MeshService object.

```yaml
kind: MeshService
metadata:
  name: redis
  namespace: redis-system
  labels: # you could then select this in when you use `MeshService` in `to` section
    team: db-operators
    kuma.io/mesh: default
spec:
  selector:
    dataplaneTags: # tags in Dataplane object, see below
      app: redis
      k8s.kuma.io/namespace: redis-system # added automatically
      kuma.io/zone: east-1 # added automatically
  ports:
  - port: 6739
    targetPort:
      value: 6739
    appProtocol: tcp
  - name: some-port
    port: 16739
    targetPort:
      name: target-port-from-container # name of the inbound
    appProtocol: tcp
status: # managed by CP. Not shared cross zone, but synced to global
  availability: Online # | Offline | NotAvailable | PartialyDegraded
  tls:
    state: Ready
    issuedBackends:
      ca-1: 5
  proxies:
    offline: 3
    online: 5
  addresses:
  - hostname: xyz.com
    status: Available # | NotAvailable
    origin: "universal-generator"
    reason: "not available because of the clash with ..."
  vips:
  - ip: <kube_cluster_ip> # or kuma VIP
    type: Kubernetes # | Kuma
```

Ports in Service can be named, so we can also name ports in MeshService. 
Target port can reference port in container by name. To avoid trying to resolve this port to a real port, we also support `targetPort#name`.
To support this, we'll also enhance our Dataplane model, so it's possible to name the inbound. 

To try to keep this MADR reasonable when it comes to length, next MADRs will cover
* Multizone strategy.
* Autogenerate MeshService on Kubernetes based on Service and on Universal based on Dataplane objects
* VIP and hostname management
* Policy matching
* Referencing MeshService in `Dataplane#outbound` and Reachable Services

This additionally helps us to build other things.

**Flexible `to` targeting of multiple services**
At the moment only targeting one service (`kind: MeshService`) or the entire Mesh is possible.
It’s not possible to target a set of services with `kind: MeshSubset` at the moment because we collect the list of kuma.io/services, each of which represents an outbound (and therefore target of to policies), by iterating over Dataplanes.
There is no way to directly label or tag one of these kuma.io/services, only a Dataplane.
We should support selecting multiple services in to section in policies. This way we can define a policy once and let service owner opt in to use it.

**Virtual Outbound 2.0**
This would eventually lead to a replacement of Virtual Outbound.

How does it solve the scalability problems:
* VIP config map is gone, because we will store VIPs and hostnames in Service objects
* ServiceInsight is gone, because we store information in the status
* `ZoneIngress#availableServices` is gone, because we will sync this cross zone.
* We can generate clusters just by looking at MeshService, we don't need to iterate over all Dataplane objects.
