# MeshService migration

* Status: accepted

## Context and Problem Statement

This MADR addresses the switch from not having MeshServices at all to MeshServices
being automatically created, either from `Services` in Kubernetes or from
`Dataplanes` in Universal.

When upgrading from 2.8 to 2.9, nothing at the data plane level should change.
But MeshService behaves differently in several different ways.

### Policy

If for example we have a policy targeted at:

```
spec:
  targetRef:
    kind: MeshService
    name: demo-app_kuma-demo_svc_5000
```

when switching over to MeshService, this policy will no longer apply.

### Local traffic

Because the `MeshServices` we generate have selectors equivalent to the `Service`,
they will select the same `Dataplanes` that previously had the `kuma.io/service`
corresponding to the `Service` and thus, absent any policies, traffic flows to
the same endpoints and Envoy has the same cluster configuration.

### Cross-zone traffic

One of the fundamental changes with `MeshService` is that traffic to a service
is no longer load balanced across all zones.

We need to take measures to prevent this change from happening on upgrade.

In particular with Kubernetes, it's important to keep in mind that after
MeshService is enabled, _any_ requests to the Kubernetes IP/via Kube DNS
go to the same outbound as to MeshService hostnames.

#### Status quo

- On Kubernetes there are two outbound listeners:
  - Traffic to the Kube DNS name/Cluster IP is load-balanced cross-zone
  - Traffic to the `<kuma.io/service>.mesh`/VIP is load-balanced cross-zone

#### `MeshService`

- We now have one outbound listener:
  - Traffic to the generated hostname of the `MeshService` goes through the Cluster IP
  - This cluster IP is no longer cross-zone load balanced

## Considered Options

* Policies applied to `demo-app_kuma-demo_svc_5000` should also apply to
  the `MeshService {name: demo-app, namespace: kuma-demo, spec.ports[port == 5000]}`.
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

MeshServices by default load balance cross-zone, like `kuma.io/service`. Users
would disable this behavior to switch to local zone behavior.

#### Pros

* Users get MeshService, HostnameGenerator advantages without breaking the data
  plane

#### Cons

* Is it coherent?
  - the set of endpoints is shared between MeshServices local
    and synced from other zones
  - HostnameGenerator can't/shouldn't by default generate with a zone suffix
* Need to keep using ZoneIngress availableServices

### Generate MeshMultiZoneService

Generate `MeshMultiZoneService` and generate hostnames equal to the
existing `<name>_<namespace>_svc_<port>` format. Users need to disable this
autogeneration to get the local zone behavior.

#### Pros

- Makes explicit the fact that requests are going cross-zone
- Things look exactly the same as how we want cross-zone services to work

#### Cons

- Would this require all zones to be upgraded so that MeshServices are synced?
- Potentially otherwise the most complex option
