# MeshService migration

* Status: accepted

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

### Policy

If for example we have a policy targeted at:

```
spec:
  to:
  - targetRef:
      kind: MeshService
      name: demo-app_kuma-demo_svc_5000
```

when switching over to MeshService, this policy will no longer apply.

### Traffic

At the moment, we generate a VIP for every `kuma.io/service` and a hostname
`<kuma.io/service>.mesh` that points to this VIP. The VIP leads to an outbound
listener and a cluster being created.

#### Before `MeshService` we have

- `<kuma.io/service>.mesh` -> Kuma generated VIP

On Kubernetes we also have

- `name_namespace_svc_port` -> ClusterIP

#### Once `MeshService` is created

`MeshService` takes over _the cluster generation_ but does
not affect the VIP generation. So we now have:

- `<kuma.io/service>.mesh` -> Kuma VIP
- `name_namespace_svc_port` -> ClusterIP
- `HostnameGenerator` DNS name for the `MeshService` -> New Kuma VIP
  - this name/IP combo is not relevant for this MADR

but all of these point to a cluster generated "the MeshService way".
This affects the data plane as follows.

##### Local traffic

Because the `MeshServices` we generate have selectors equivalent to the `Service`,
they will select the same `Dataplanes` that previously had the `kuma.io/service`
corresponding to the `Service` and thus, ignoring the question of policy, traffic flows to
the same endpoints and Envoy has the same cluster configuration.

##### Cross-zone traffic

One of the fundamental changes with `MeshService` is that traffic to a service
is no longer load balanced across all zones.

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

* Policies applied to `demo-app_kuma-demo_svc_5000` should also apply to
  the `MeshService {name: demo-app, namespace: kuma-demo, spec.ports[port == 5000]}`.
  * we can drop this behavior when we drop `kuma.io/service` support
* For cross-zone:
  * don't generate MeshService by default
  * have a setting to disable using MeshServices
  * change the default behavior of MeshService from single-zone load balancing to cross-zone.
    Users would need to manually switch over to local zone load-balancing after having
    potentially migrated cross-zone services to `MeshMultiZoneService`.
  * instead of changing the behavior of MeshService, instead also generate
    MeshMultiZoneService by default, which would reintroduce the cross-zone load balancing behavior.

## Decision Outcome

* Policies applied to `demo-app_kuma-demo_svc_5000` should also apply to
  the `MeshService `{name: demo-app, namespace: kuma-demo, spec.ports[port == 5000]}`.
* Change `MeshService` to do cross-zone load balancing by default and have users
  switch over to local zone plus `MeshMultiZoneServices`

## Pros and Cons of the Options

### Don't generate MeshServices by default

Users would need to enable MeshService generation.

#### Pros

* One of the easiest, most straightforward option as far as stability and UX goes.

#### Cons

* Will be harder to motivate users to switch at all
* None of the advantages of MeshService & HostnameGenerator "for free"
* MeshServices need to be deleted if generation is enabled then disabled

### Generate but don't enable by default

Users would need to enable usage of MeshService resources.

#### Pros

* One of the easiest, most straightforward option as far as stability and UX goes.

#### Cons

* Will be harder to motivate users to switch at all
* None of the advantages of MeshService & HostnameGenerator "for free"
* Odd that MeshService objects exist but aren't used

### Make cross-zone the default behavior

The cross-zone behavior would be kept for the old, `<kuma.io/service>.mesh` VIP
and the ClusterIP by generating Envoy clusters that match the status quo
generation.
Users would disable this behavior explicitly to switch to local zone behavior.

This means that the clusters generated by Envoy would no longer change on upgrade,
preserving the traffic.

#### Pros

* Users get MeshService, HostnameGenerator advantages without breaking the data
  plane

#### Cons

* Need to keep using ZoneIngress availableServices

### Generate MeshMultiZoneService

Generate `MeshMultiZoneService` and generate hostnames equal to the
existing `<name>_<namespace>_svc_<port>` format. Users need to disable this
autogeneration to get the local zone behavior.

#### Pros

- Makes explicit the fact that requests are going cross-zone
- Things look exactly the same as how we want cross-zone services to work

#### Cons

- Ihis would require all zones to be upgraded so that MeshServices are synced
  and can be matched by MeshMultiZoneService
- Otherwise the most complex option
- Depends on using the new VIP, the old VIP would still break on change
