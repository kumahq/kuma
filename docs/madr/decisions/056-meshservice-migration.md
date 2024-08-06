# MeshService migration

- Status: accepted

## Context and Problem Statement

This MADR addresses the switch from not having MeshServices at all to MeshServices
being automatically created, either from `Services` in Kubernetes or from
`Dataplanes` in Universal.

When upgrading from 2.8 to 2.9, nothing at the data plane level should change.
But MeshService behaves differently in several different ways.

In general the migration process for MeshService consists of no longer using
`kuma.io/service` so:

- targeting policies to the real `MeshService` resource
- switching over to the `MeshService` i.e. `HostnameGenerator` hostnames, if not
  using Kubernetes DNS

### `MeshService`-targetted policy

If for example we have a policy targeted at:

```
spec:
  to:
  - targetRef:
      kind: MeshService
      name: demo-app_kuma-demo_svc_5000
```

when switching over to MeshService, this policy will no longer apply.

We can always convert the above to an additional entry:

```
  - targetRef:
      kind: MeshService
      labels:
        kuma.io/display-name: demo-app
        kuma.io/namespace: kuma-demo
      port: 5000
```

### Stats

For the Kuma API it's important that we can derive the destination that corresponds
to a given metric label, like `name_namespace_svc_port`.
In particular, it's valuable if we can do this without making additional
requests.

If we have a label like `name_namespace_svc_port` but we need to make a request
in order to figure out what resource it correponds to, that's a disadvantage.

### Traffic

At the moment, we generate a VIP for every `kuma.io/service` and a hostname
`<kuma.io/service>.mesh` that points to this VIP. The VIP leads to an outbound
listener and a cluster being created.

#### Before `MeshService` we have

- `<kuma.io/service>.mesh` -> Kuma generated VIP

On Kubernetes we also have

- `name.namespace.svc.cluster.local` -> ClusterIP

#### Once `MeshService` is created

`MeshService` takes over _the cluster generation_ but does
not affect the VIP generation. So we now have:

- `<kuma.io/service>.mesh` -> Kuma VIP
- `name_namespace_svc_port` -> ClusterIP
- `HostnameGenerator` DNS name for the `MeshService` -> New Kuma VIP
  - note that this includes synced `MeshServices`

`MeshService` load balances only to the local zone, as opposed to cross-zone.
Now all of these listeners point to a cluster generated the MeshService way.

This means all `MeshService` traffic stops being cross-zone load balanced and stays in
the local zone.
We need to take measures to prevent this change from happening immediately
upon upgrade.

In particular with Kubernetes, it's important to keep in mind that after
MeshService is enabled we no longer generate our own VIP.
So any requests to MeshService hostnames go to the same
outbound listener as with requests to Kubernetes DNS names, the ClusterIP.

###### Status quo

- On Kubernetes there are two DNS names/outbound IP addresses:
  - Traffic to `name_namespace_svc_port`/ClusterIP is load-balanced cross-zone
  - Traffic to the `<kuma.io/service>.mesh`/VIP is load-balanced cross-zone

###### `MeshService`

- We now have one outbound IP address:
  - Traffic to the generated hostname of the `MeshService` goes through the Cluster IP
  - This cluster IP is no longer cross-zone load balanced

## Considered Options

- Policies applied to `demo-app_kuma-demo_svc_5000` should also apply to
  the `MeshService labels: {kuma.io/display-name: demo-app, kuma.io/namespace: kuma-demo}, port: 5000`.
  - we can drop this behavior when we drop `kuma.io/service` support
- For traffic:
  - don't generate MeshService by default
  - have a setting to disable using MeshServices
  - change the default behavior of MeshService from single-zone load balancing to cross-zone.
    Users would need to manually switch over to local zone load-balancing after having
    potentially migrated cross-zone services to `MeshMultiZoneService`.
  - instead of changing the behavior of MeshService, instead also generate
    MeshMultiZoneService by default, which would reintroduce the cross-zone load balancing behavior.
  - generate new clusters for MeshService and preserve the old behavior

## Decision Outcome

- Policies applied to `demo-app_kuma-demo_svc_5000` should also apply to
  the `MeshService `{name: demo-app, namespace: kuma-demo, spec.ports[port == 5000]}`.
- Change `MeshService` to do cross-zone load balancing by default and have users
  switch over to local zone plus `MeshMultiZoneServices`

## Pros and Cons of the Options

### Don't generate MeshServices by default

Users would need to enable MeshService generation and MeshService-controlled XDS
generation.

#### Migration

1. Upgrade to new version
1. Create MeshMultiZoneServices for cross-zone traffic.
  - Potentially create MeshService manually to gradually switch over
1. Enable MeshService generation to have generate MeshServices
  and have local zone traffic take over

#### Pros

- One of the easiest, most straightforward option as far as stability and UX goes.

#### Cons

- Will be harder to motivate users to switch at all
- MeshServices need to be deleted if local zone load balancing should be enabled then disabled

### Make cross-zone the default behavior

The cross-zone behavior would be kept for the old, `<kuma.io/service>.mesh` VIP
and the ClusterIP by generating Envoy clusters that match the status quo
generation.
Users would disable this behavior explicitly to switch to local zone behavior.

This means that the clusters generated by Envoy would no longer change on upgrade,
preserving the traffic.

#### Migration

1. Upgrade to new version
1. MeshServices are now generated but all traffic stays cross-zone
1. The user creates MeshMultiZoneServices for traffic that should be cross-zone
1. Disable cross-zone traffic for MeshServices
1. Disable old VIP generation

#### Pros

- Users get MeshService, HostnameGenerator advantages without breaking the data
  plane

#### Cons

- Need to keep using ZoneIngress availableServices

### Generate MeshMultiZoneService

Generate `MeshMultiZoneService` and generate hostnames equal to the
existing `<name>_<namespace>_svc_<port>` format. Users need to disable this
autogeneration to get the local zone behavior.

Policy matching would need to be adjusted so that old policies also target
`MeshMultiZoneService`.

#### Migration

This migration is potentially the trickiest.

1. Upgrade to new version
1. `MeshMultiZoneService` are now generated and they take over generation of
   `name_namespace_svc_port` clusters. `MeshServices` are also generated

It's not clear what the path or even the goal is from here.

#### Pros

- Makes explicit the fact that requests are going cross-zone
- Things look exactly the same as how we want cross-zone services to work

#### Cons

- Ihis would require all zones to be upgraded so that MeshServices are synced
  and can be matched by MeshMultiZoneService
- Otherwise the most complex option

### Generate new clusters

Instead of `MeshService` taking over cluster generation, generate separate clusters so that:

- `<kuma.io/service>.mesh` -> Old Kuma VIP -> old cluster
- `HostnameGenerator` DNS name for the `MeshService` -> New Kuma VIP
  - this includes synced `MeshServices`
- `name.namespace.svc.cluster.local` -> ClusterIP -> depends

Kubernetes has an extra complication because the Kube DNS name/ClusterIP
has to point at one cluster or the other, so we need a switch here:

```
KUMA_RUNTIME_KUBERNETES_DNS_CLUSTER_PRIORITY=ServiceTag|MeshService
```

For example, we could have two formats like `name_namespace_svc_port` and `name_namespace_msvc_port`.

This means increasing the number of synced clusters when **not using reachable
services**, during the migration period where both kinds of VIPs are in use and
thus both clusters are required.

However, with reachable services, things become very straightforward because it's always
very explicit which services are actually in use and need to be synced.

#### Migration

1. Upgrade to new version
1. MeshServices are now generated but no traffic is sent to `MeshService` VIPs
   yet
1. The user starts migrating to either `MeshService` or `MeshMultiZoneService`
   depending on whether they need cross-zone
1. Disable old cluster/VIP generation

#### Pros

- Cleanest separation of behavior
- Separating clusteres is the least likely to introduce unintended behavioral changes
- 1-1 correspondence between format and type of resource, easiest for API

#### Cons

- Increased number of resources to sync to proxies
