# KDS Redesign

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/4354

## Context and Problem Statement

The current implementation of KDS is quite complex, and there may be some bugs, such as the one found at https://github.com/kumahq/kuma/pull/5373, which was very difficult to identify and resolve. The use of a bi-directional stream for configuration exchange between the zone and the global can also be challenging to implement. Additionally, KDS operates as a State of The World, meaning that whenever an item in the resource type changes, the entire state is sent. While this doesn't appear to be a problem when synchronizing data from Global to Zone, it may be less efficient when there are many Dataplane changes. There are several improvements we want to make in the future such as:
* working with different older versions of the control plane on the zone,
* creating separate gRPC services for synchronization between the zone and global,
* supporting multiple control plane tenants,
* prioritizing sync of operational over informational resources,
* changing `default` namespace for sync to global to system one (`kuma-system`), 
* and sending only changes instead of the state of the world when synchronizing data from the zone to the global.

## Considered Options

* create separate RPC for synchronization Zone to Global and Global to Zone
* introduce delta xDS
* keep current KDS

## Decision Outcome

Chosen option: 
* create separate RPC for synchronization Zone to Global and Global to Zone
* introduce delta xDS

## Solution

### Current state
The implementation of bidirectional stream for synchronization of the resources is really complicated and difficult to understand. 
Currently, we are using one bidirection stream for sending resources from Zone to Global and Global to Zone.

```proto
service KumaDiscoveryService {
 rpc StreamKumaResources(stream envoy.service.discovery.v3.DeltaDiscoveryRequest)
     returns (stream envoy.service.discovery.v3.DeltaDiscoveryResponse);
}
```

That is hidden behind `MultiplexService`.

``` proto
service MultiplexService {
  rpc StreamMessage(stream Message) returns (stream Message);
}
```

A client sends DiscoveryRequest to the server and waits for the change. The server stores the state in a state-of-the-world cache and sends the response whenever there is a change. It works well in the case of Global to Zone synchronization
but uses more bandwidth for synchronization from Zone to Global with large deployments. Zone control-plane needs to send all the Dataplane resources to the global whenever there is a change.

### New state

We want to simplify the logic by introducing 2 separate RPCs for synchronization of resources. That separation is going to simplify logic and the readability.

```proto
service GlobalKDSService {
 ...
 rpc ZoneToGlobalSync(stream envoy.service.discovery.v3.DeltaDiscoveryResponse)
     returns (stream envoy.service.discovery.v3.DeltaDiscoveryRequest);

 rpc GlobalToZoneSync(stream envoy.service.discovery.v3.DeltaDiscoveryRequest)
     returns (stream envoy.service.discovery.v3.DeltaDiscoveryResponse);

```

New RPCs are going to come with the new xDS implementation called [`Incremental xDS`](https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol). We don't want to create a breaking changes in the new api that's why we are going to use delta xDS since the begining.

### Prioritize operational informations

Certain resources require more frequent synchronization than others. For example, Policies, ZoneIngress, and ZoneEgress must be updated immediately, whereas Insights and Dataplanes objects may be delayed.

#### **Current state**
Control-plane has a configuration option that allows to specify how often it should refresh the cache state. If newly calculated snapshot hasn't change then informations are not propagated to the clien't.

#### **New state**
We are going to keep that behaviour.


### Change the namespace of synchronized resources from `default` to `kuma-system`

Story: https://github.com/kumahq/kuma/issues/3247

#### **Current state**
Namespace scoped resources synchronized from Zone to Global and Global to Zone are created by default in a `default` namespace(except pluggable policies) and we should change it in the new version.

#### **New state**
All the namespace scope resources are stored in `kuma-system` namespace. It's important to notice that together with the new implementation we need to keep the current behaviour.

#### **Test result**

Current implementation of synchornization allows the smooth transition from the one namespace to another. Control-plane receive the whole state in a KDS response and compare it with the current state of local database. If the namespace has changed it added resources to new one and removed from the old one. There is no problem with rollback, upgrade.

### Handshake

It's not possible for the zone CP to be newer than global CP. This is because we can always guarantee that it's the newest. There is an information in docs about upgrade path that mentions order of upgrades.

### Multitenant

Global control-plane should support multitenant, that means we want to achieve that one global control-plane can handle many independent zone control-planes. Based on the analysis it seems that there are no changes required in KDS to support multitenant.

### xDS communication for KDS synchronization

Story: https://github.com/kumahq/kuma/issues/4926

#### **Current state**

Currently, control-planes uses copy of the snapshot cache from go-control-plane. Also, for the synchronization of resources we are using state of the world which sends all the resources for any change. This works fine but when the number of dataplanes need to be synchronized from zone to global, control-planes sends all dataplanes even if only one changed. 

Flow:

1. Zone sends `DiscoveryRequest` with `typeUrl: Mesh`
2. Global responds with `DiscoveryResponse` with all the mesh resources `resources: ["mesh-1", "mesh-2"]`
3. Zone sends `DiscoveryRequest` with `typeUrl: Mesh` and waits until there is a change
4. `mesh-1` has changed
5. Global responds with `DiscoveryResponse` with all the mesh resources `resources: ["mesh-1", "mesh-2"]`


#### **New state**
We can change the sycnhronization method from state of the world to delta xDS. This allows sending only resources that have changed. The change to the delta xDS requires refreshed implementation of the cache so we need to use the newest go-control-plane code. 

Flow:

1. Zone sends `DeltaDiscoveryRequest` with `typeUrl: Mesh` and `resource_names_subscribe: "*"`
2. Global responds with `DeltaDiscoveryResponse` with all the mesh resources `resources: ["mesh-1", "mesh-2"]`
3.  Zone sends `DeltaDiscoveryRequest` with `typeUrl: Mesh` and `resource_names_subscribe: "*"`
4. `mesh-1` has changed
5. Global responds with `DeltaDiscoveryResponse` with all the mesh resources `resources: ["mesh-1"]`

Implementation of the cache for delta xDS keeps in the map information about the each resource version. Thanks to this control-plane can responds only with the resources that have changed.

xDS Specification
* https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol#how-the-client-specifies-what-resources-to-return
* https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol#incremental-xds
* https://github.com/envoyproxy/go-control-plane/blob/main/pkg/cache/v3/delta.go#L31

## Pros and Cons of the Options

### Create separate RPC for synchronization Zone to Global and Global to Zone

Pros
* improve readability and 
* simplify logic
* separate which RPC is responsible for synchronization

Cons
* before deprication of the old api more code


### Introduce delta xDS

Pros
* more performent way of synchronization of the resources
* together with new model we can synchronize with go-control-plane implementation

Cons
* no client in the go-control-plane requires us to create one
* control-planes needs to do calculation of the delta which makes it more CPU usage

### Keep the current KDS

Pros 

* no code changes needed

Cons
* the code is complicated and requires more time to change or fix something
* each change of a resource requires all the state to be sent

## How we can split the work

We can split the implementation and introduce features independently.

### Stages

1. Migrate the synchronization of the resources from `default` to `kuma-system`
2. Introduce `ZoneToGlobalSync` with incremental xDS
3. Introduce `GlobalToZoneSync` with incremental xDS
