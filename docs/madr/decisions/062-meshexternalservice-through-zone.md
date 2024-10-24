# MeshExternalService Accessible Only Through a Specific Zone

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/11071

## Context and Problem Statement

In certain cloud environments, mesh-enabled applications need to communicate with applications running in a Datacenter (DC) that are not part of the mesh. These DC applications are not exposed to the internet. To address this, we installed a ZoneCP in the DC. This setup allows DC applications to be resolved from the Kuma Egress running in the DC.

For example, by registering the DC apps as MeshExternalService with a zone tag matching the DC zone name, traffic originating from other zones can be routed to the DC zone. Once the traffic reaches the DC, the egress can resolve and direct it to the appropriate DC applications.

This scenario provides a possibility of seamless communication between mesh and non-mesh applications. However, we have additional use cases that also need consideration, where traffic flow and resolution rules may vary across zones and require tailored configurations.
Scenario:

* Zone 2 has a Zone Egress but cannot route traffic outside the organization.
* Zone 1 has a Zone Egress that can route traffic outside the organization.

## Considered Options

* Create `MeshExternalService` on a Zone. 
  The `MeshExternalService` would be created in a zone and exposed similarly to how `MeshService` is synced across zones.

* Add a Label
  Introduce a label like `kuma.io/accessible-zone` (to determine the best name) to restrict external access to a specific zone.

## Decision Outcome

Chosen option: "`MeshExternalService` can be created on the zone and exposed in the same way as `MeshService` synced to other zones."
This approach aligns with how `MeshService` works and should reduce confusion for users.

## Pros and Cons of the Options

### Create `MeshExternalService` on a Zone

A user can create a `MeshExternalService` in a zone that allows outgoing traffic via Zone Egress. In this case, the `MeshExternalService` will be synced to the global control plane and all other zones (except the originating zone). The zone where the `MeshExternalService` was created will expose the service on the ingress, and other zones will route traffic to this Zone Ingress from their ZoneEgress.


Example:

        ---> Zone 1
Global
        ---> Zone 2


Let's create a `MeshExternalService` on Zone 1

```yaml
type: MeshExternalService
name: mes-through-zone-1
mesh: default
labels:
  kuma.io/origin: zone
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: http
  endpoints:
    - address: httpbin.org
      port: 80
```

A user in Zone 2 would make a request to `mes-through-zone-1` like this:

```bash
curl mes-through-zone-1.extsvc.zone-1.mesh.local
```

The traffic flow looks like this:

`curl mes-through-zone-1.extsvc.zone-1.mesh.local` -> Zone Egress (Zone 2) -> Zone Ingress (Zone 1) -> Zone Egress (Zone 1)

When a user in Zone 1 calls `mes-through-zone-1`, the call is handled differently, as the service is local. The hostname generated for a local `MeshExternalService` follows a different pattern:

```bash
curl mes-through-zone-1.extsvc.mesh.local
```

The traffic flow in this case is simpler:

`curl mes-through-zone-1.extsvc.mesh.local` -> Zone Egress (Zone 1)

#### Advantages:

* Works the same way as a `MeshService` synced to other zones (accessible via a specific hostname).
* No need for handling additional labels logic on the global control plane or adding extra fields.

#### Disadvantages:
* Adds complexity to the flow.
* Every `MeshExternalService` created in a zone is exposed via Zone Ingress.

### Add a label

In this approach, the user must explicitly provide information during resource creation. The resource can only be created on the global control-plane.

```yaml
type: MeshExternalService
name: mes-through-zone-1
mesh: default
labels:
  kuma.io/accessible-zone: zone-1
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: http
  endpoints:
    - address: httpbin.org
      port: 80
```

Once the resource is synced, the control-plane uses the `kuma.io/accessible-zone` label to generate configurations. This configuration directs all zones—except the one specified by the label's value—to use the ingress of the specified zone.

It's not possible to validate the zone name during resource creation, but we could log a message indicating there is no zone with the provided name. However, this is not an ideal solution.

#### Advantages:
* There's no need to expose all `MeshExternalServices` created in the zone on the ingress.

#### Disadvantages:
* Requires managing an additional label.
* Additional logic is needed to handle the label.
