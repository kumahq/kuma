# MeshService migration

- Status: accepted

## Context and Problem Statement

This MADR addresses the switch from not having MeshServices at all to MeshServices
being automatically created, either from `Services` in Kubernetes or from
`Dataplanes` in Universal.

When upgrading from 2.8 to 2.9, nothing at the data plane level should change.
But MeshService behaves differently in several ways.

In general the migration process for MeshService consists of no longer using
`kuma.io/service` so:

- targeting policies to the real `MeshService` resource
- switching over to the `MeshService` i.e. `HostnameGenerator` hostnames and
  local-zone load balancing from cross-zone load balancing

Ideally users can try out `MeshService` behavior for individual consumers.

Note that we in general can't be smart about which behavior should be the
default because we can't differentiate between a new user, who doesn't care
about `kuma.io/service`-style behavior, and an existing user who's just
redeploying on a new `Mesh` or control plane.

We have to make a choice between:

- new behavior by default, `UPGRADE.md` tells users to set a variable to keep
  old behavior
- old behavior by default, new users have to be convinced to set a variable to
  get new behavior

### `MeshService`-targeted policy

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
      sectionName: 5000
```

Potentially we could offer a tool to convert these refs from old-style to
new-style `MeshService`.

#### Require `sectionName`

Because the original refers to instances of `demo-app_kuma-demo_svc_5000` from
any zone, we have to use `labels` to match instances outside of the local
zone.

Note that there's an ambiguity here with universal. We can't tell whether:

```
spec:
  to:
  - targetRef:
      kind: MeshService
      name: backend
```

is meant to target `kuma.io/service` or a `MeshService` named `backend` and thus
we can't determine whether all instances of `kuma.io/service: backend` are meant
or only the instances matched by `MeshService`.

Thus we have to require `sectionName` to a port, in order to refer to
`MeshServices` unambiguously.

#### Port name

We can't require users to add a `name` field to their existing `Services` or
annotate their `Dataplanes` for universal, so we should
set a canonical name for every port assuming one isn't set.

Because the `port` field must be unique anyway, this could be a suitable,
automatic `name`.

#### Do not apply _routes_ targeted at `kuma.io/service` to corresponding `MeshServices`

An issue arises with `MeshHTTPRoute`/`MeshTCPRoute` because these contain,
inside each `spec.to[].rules` list, further references to `kuma.io/service` and
`MeshService`. Because the behavior of routes is fundamentally linked to the the
endpoint selection, which `MeshService` changes, it's important in any case that
every route resource is examined for correctness when migrating to
`MeshService`.

For this reason, `MeshHTTPRoutes` and `MeshTCPRoutes` will **not** be targetted
to corresponding real `MeshService` objects when they target `kuma.io/service`.

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

- One cross-zone cluster for a `kuma.io/service`

- A `<kuma.io/service>.mesh` listener for Kuma generated VIP -> cross-zone cluster

On Kubernetes we also have

- `name.namespace.svc.cluster.local` -> ClusterIP -> same cross-zone cluster

#### Once `MeshService` is created

`MeshService` takes over _the cluster generation_ but does
not affect the VIP generation. So we now have:

- `<kuma.io/service>.mesh` -> Kuma VIP -> local-zone cluster
- `name.namespace.svc.cluster.local` -> ClusterIP -> local-zone cluster
- `HostnameGenerator` DNS name for the `MeshService` -> New Kuma VIP -> local-zone cluster
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
outbound listener as with requests to Kubernetes DNS names,
the listener for the ClusterIP.

## Considered Options

- Policies applied to `demo-app_kuma-demo_svc_5000` should also apply to
  the `MeshService labels: {kuma.io/display-name: demo-app, kuma.io/namespace: kuma-demo}, sectionName: "5000"`.
  - we can drop this behavior when we drop `kuma.io/service` support
  - require `sectionName` to disambiguate
  - automatically set `MeshService.spec.ports[].name` when generating these
    resources
    - on both k8s and universal, it should be the `port` field
- For traffic:
  - don't generate MeshService by default
  - have a setting to disable using MeshServices
  - change the default behavior of MeshService from single-zone load balancing to cross-zone.
    Users would need to manually switch over to local zone load-balancing after having
    potentially migrated cross-zone services to `MeshMultiZoneService`.
  - instead of changing the behavior of MeshService, also generate
    MeshMultiZoneService by default, which would reintroduce the cross-zone load balancing behavior.
  - generate new clusters for MeshService and preserve the old behavior

## Decision Outcome

- Policies applied to `demo-app_kuma-demo_svc_5000` should also apply to
  the `MeshService `{name: demo-app, namespace: kuma-demo, spec.ports[port == 5000]}`.
  - **except** for `MeshHTTPRoute`/`MeshTCPRoute`
  - in order to target `MeshService` only, `sectionName` must be set
  - For `MeshLoadBalancingStrategy`, only the `localZone` portion applies
- When the feature is enabled, generate new clusters for `MeshServices`
  - the new behavior stays off by default, users must opt in
  - they can gradually switch consumers over with per-Dataplane annotations

### Migration

1. Upgrade to new version

- If the user doesn't change any config, nothing changes. No `MeshServices` are generated.
- On `Mesh`, users have the option `meshServices` with `Disabled` being the
  default value

2. The user starts migrating to either `MeshService` or `MeshMultiZoneService`
   depending on whether they need cross-zone

   - requires updating `MeshHTTPRoute`/`MeshTCPRoute`

3. Users can enable behavior by setting
   - `meshServices: Everywhere` syncs `MeshService` outbounds to all data plane proxies
   - `meshServices: ReachableBackendRefs` to enable on a case-by-case basis by setting
     `reachableBackendRefs`:
     `reachableBackendRefs: { kind: MeshService, labels: {} }`

- **NOTE**: On k8s if the consumer can reach a given local `MeshService`,
  the Kubernetes IP has the behavior of `MeshService`
- **NOTE**: This change makes the _experimental_ options `GENERATE_MESH_SERVICES` & `SKIP_PERSISTED_VIPS`
  _obsolete_, these options _now do nothing_

4. Set `meshServices: Exclusive` to disable old behavior and config generation

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

It also means that policies targeted to real `MeshServices` also target `.mesh` traffic, which is probably fine.

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

The new format is:

```
<name>_<namespace>_<zone>_<mesh>_msvc_<port>
```

where `<namespace>` is empty for universal.

This means increasing the number of synced clusters when **not using reachable
services**, during the migration period where both kinds of VIPs are in use and
thus both clusters are required.

However, with reachable services, things become very straightforward because it's always
very explicit which services are actually in use and need to be synced.

#### Kubernetes DNS ClusterIP

Kubernetes has an extra complication because the Kube DNS name/ClusterIP
has to point at one cluster or the other.

We either need a switch:

```
KUMA_RUNTIME_KUBERNETES_DNS_CLUSTER_PRIORITY=ServiceTag|MeshService
```

or maybe it's OK to switch to local-zone load balancing:

- we currently only add an outbound for a service on k8s if there's a local
  instance of it
- with `localityAwareLoadBalancing: false`
  - only if local instances becomes unhealthy does traffic flow to other zones
- with `localityAwareLoadBalancing: true`
  - traffic never flows cross-zone
- what about `MeshLoadBalancingStrategy`?
  - only `localZone` applies

#### Migration

1. Upgrade to new version
   If the user doesn't change any config, nothing changes. No `MeshServices` are generated.
1. On `Mesh`, users have the option `meshServices`:

- `Disabled` is the default
- `Everywhere` syncs `MeshService` outbounds to all data plane proxies
- `Exclusive` disables `kuma.io/service` outbounds and enables `MeshService`
- `ReachableBackendRefs` to generate all `MeshServices` but sync
  on a case-by-case basis by setting `reachableBackendRefs`:
  ```
  reachableBackendRefs: { kind: MeshService, labels: {} }
  ```

1. On K8s: if the consumer can reach a given local `MeshService`,
   the Kubernetes IP has the behavior of `MeshService`
1. The user starts migrating to either `MeshService` or `MeshMultiZoneService`
   depending on whether they need cross-zone
1. Disable old cluster/VIP generation

#### Pros

- Cleanest separation of behavior
- Separating clusteres is the least likely to introduce unintended behavioral changes
- 1-1 correspondence between format and type of resource, easiest for API

#### Cons

- Increased number of resources to sync to proxies
