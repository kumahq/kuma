# KDS Redesign

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/4354

## Context and Problem Statement

The current implementation of KDS is quite complex, and there may be some bugs, such as the one found at https://github.com/kumahq/kuma/pull/5373, which was very difficult to identify and resolve. The use of a bi-directional stream for configuration exchange between the zone and the global can also be challenging to implement. Additionally, KDS operates as a State of The World, meaning that whenever an item in the resource type changes, the entire state is sent. While this doesn't appear to be a problem when synchronizing data from Global to Zone, it may be less efficient when there are many Dataplane changes. There are several improvements we want to make in the future such as:
* working with different older versions of the control plane on the zone,
* creating resources on the zone and syncing them to the global, 
* creating separate gRPC services for synchronization between the zone and global, 
* supporting multiple control plane tenants,
* and sending only changes instead of the state of the world when synchronizing data from the zone to the global.

## Considered Options

* Redesign KDS

## Decision Outcome

Chosen option: "Redesign KDS".

## Solution

Currently, we are using one bidirection stream for sending resources from Zone to Global and Global to Zone. 

```proto
service KumaDiscoveryService {
  rpc StreamKumaResources(stream envoy.service.discovery.v3.DeltaDiscoveryRequest)
      returns (stream envoy.service.discovery.v3.DeltaDiscoveryResponse);
}
```

This is hidden behind `MultiplexService`.

``` proto
service MultiplexService {
  rpc StreamMessage(stream Message) returns (stream Message);
}
```

Zone synchronize to Global resources:

* Dataplane
* DataplaneInsight
* ZoneEgress
* ZoneEgressInsight
* ZoneIngress*
* ZoneIngressInsight

`*` ZoneIngress resource is synchronized from Global to Zone and Zone to Global

Global synchronize to Zone resources:
* CircuitBreaker
* ExternalService
* FaultInjection
* HealthCheck
* Mesh
* MeshGateway
* MeshGatewayRoute
* ProxyTemplate
* RateLimit
* Retry
* ServiceInsight
* Timeout
* TrafficLog
* TrafficPermission
* TrafficRoute
* TrafficTrace
* VirtualOutbound
* ZoneIngress*
* Secret
* Config
* MeshAccessLog
* MeshCircuitBreaker
* MeshFaultInjection
* MeshHealthCheck
* MeshHttpRoute
* MeshProxyPatch
* MeshRateLimit
* MeshRetry
* MeshTimeout
* MeshTrace
* MeshTrafficPermission

Client sends DiscoveryRequest to the server and waits for the change. Server store the state in a state of the world cache  and sends the response whenever there is change. It works well in case of Global to Zone synchronization
but might have scalability issues for synchronization Zone to Global with large deployments. Zone control-plane needs to send all the Dataplane resources to the global whenever there is a change. That might be a bottleneck that can happen in the large deployments. We can use [`Incremental xDS`](https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol) which is Envoy's implementation of cache that instead of sending all the resources sends only changes. That could reduce the bandwidth usage.


```
service GlobalKDSService {
  ...

  rpc ZoneToGlobalStream(stream envoy.service.discovery.v3.DeltaDiscoveryRequest)
      returns (stream envoy.service.discovery.v3.DeltaDiscoveryResponse);

  rpc GlobeToZoneStream(stream envoy.service.discovery.v3.DeltaDiscoveryRequest)
      returns (stream envoy.service.discovery.v3.DeltaDiscoveryRequest);
}

```

In the first iteration we can introduce only `ZoneToGlobalStream` with delta xDS, and in the future we could introduce `GlobeToZoneStream` and also use delta xDS. In the code I've noticed that we are using our copy of the Snapshot cache and the code is a bit outdated. It would be nice to check if we are able to use the go-control-plane cache or we might have to copy implemntation of cache for incremental xDS aswell.

### Handshake 

Is it possible for the zone control-plane to be newer than the global control-plane?

If the answer is no, then a handshake is not necessary as the client will use the latest available version.

If the answer is yes, then a handshake API may need to be implemented. The GRPC Reflection API allows services to be discovered on the server side, but not on the client side.

The output of the reflection functionality is:

```
 ["grpc.reflection.v1alpha.ServerReflection", "kuma.mesh.v1alpha1.GlobalKDSService", "kuma.mesh.v1alpha1.MultiplexService"]```
```

How would handshake work?

1. The server exposes the RPC `handshake`.
2. The client initiates a connection to the server.
3. The client sends a request that includes the Kuma version and the supported KDS versions:
```
  kumaVersion: "2.1.0",
  supportedKdsVersions: ["0.1.0", "0.2.0"],
```

4. Server checks the request and responds with the newest version working for both
```
  kdsVersion: "0.2.0"
```

### Multitenant

In order to connect multitenant control-planes to a global one, we must have a means of identifying them. One way to achieve this is by extracting tenant details from an authentication token. This extracted information can then be passed in request to specify the particular tenant and retrieve only their deployment's data. It seems that there is no extra effort required to support multitenancy.

### Create resources on the zone

We intend to introduce a new feature that enables the creation of resources at the zone level. These resources will be exclusively applicable to the specific zone, and we will synchronize them to the global level as read-only. To distinguish the zone resources from global ones, we will prefix them with the name of the zone.

```
zone-1.example-timeout
```

### Resources
The following resources are the ones we should permit creating in the zone:
* CircuitBreaker
* ExternalService
* FaultInjection
* HealthCheck
* MeshGateway
* MeshGatewayRoute
* ProxyTemplate
* RateLimit
* Retry
* Timeout
* TrafficLog
* TrafficPermission
* TrafficRoute
* TrafficTrace
* VirtualOutbound
* MeshAccessLog
* MeshCircuitBreaker
* MeshFaultInjection
* MeshHealthCheck
* MeshHttpRoute
* MeshProxyPatch
* MeshRateLimit
* MeshRetry
* MeshTimeout
* MeshTrace
* MeshTrafficPermission

When there are two policies, one at the zone level and another at the global level, that impact the same dataplanes, the more specific policy takes precedence. In this scenario, the Zone policy would be the one utilized.

Global policy:
```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: timeout-global
  namespace: kuma-system
  labels:
    kuma.io/mesh: default
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: Mesh
      default:
        idleTimeout: 20s
        connectionTimeout: 2s
```

Zone policy:

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: timeout-zone
  namespace: kuma-system
  labels:
    kuma.io/mesh: default
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: Mesh
      default:
        idleTimeout: 10s
        connectionTimeout: 2s
```

The policy `timeout-zone` is going to be applied to all dataplanes in default mesh in the zone.

### Incremental xDS

Incremental xDS in the `DeltaDiscoeryRequest` has a field `resource_names_subscribe` which requires information about resources that we want to subscribe to e.g. policy with name `timeout-1`. If you send `resource_names_subscribe:  "*"` in a `DeltaDiscoveryRequest`, the control plane should respond with all resources of the specified type, not just the ones that have changed since the last update. This is because the `resource_names_subscribe` field is used to filter which resources the control plane should send, and if it is set to `*`, it effectively disables filtering and requests all resources of the specified type.

To receive only the resources that have changed since the last update, include the `version_info` field in `DeltaDiscoveryRequest`. The `version_info` field should contain the version of the last update that the proxy received. When the control plane sends a response, it includes a new version number for each resource, and the proxy can use these version numbers to determine which resources have changed since the last update. The control plane should only include the resources that have changed in the response to a `DeltaDiscoveryRequest` that includes a `version_info`.