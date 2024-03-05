# Moving to routes 2.0

- Status: accepted

## Context and Problem Statement

This document is meant to make explicit how we will transition
from `TrafficRoute` to `MeshTCPRoute` and `MeshHTTPRoute` and also how
these policies interact with each other.

### Status quo

With the new paradigm "no default creation for 2.0 policies" it becomes necessary
for traffic to flow when there are no `TrafficRoutes` and no
`MeshTCPRoutes`/`MeshHTTPRoute`. This break from the behavior of `TrafficRoutes` and the
ensuing migration need some careful consideration.

## Decision Drivers

- When installing Kuma, traffic should continue to flow assuming the user takes no
  explicit action to configure routing.
- Avoid creation of resources by the CP when a `Mesh` is created
- Provide a smooth transition for `TrafficRoutes` to `Mesh*Route`

## Decision Outcome

- Depend on the existence of _any_ `TrafficRoute` resources in the `Mesh` to determine which behavior wins.

### Scenarios

So when would we generate Envoy listeners and routes allowing traffic to flow from
service A to B?

|                                       | Any `TrafficRoutes` resource exists | No `TrafficRoutes`                        |
| ------------------------------------- | ----------------------------------- | ----------------------------------------- |
| Some `Mesh*Route.spec.to` targets `B` | configured by `Mesh*Route`          | configured by `Mesh*Route`                |
| No `Mesh*Route.spec.to` targets `B`   | configured by `TrafficRoutes`       | traffic flows, configured by `Mesh*Route` |

Note that currently, no `TrafficRoutes`, no `Mesh*Routes` would mean a broken
cluster and is therefore in some sense an invalid state. This state is the only
one for which behavior changes in a breaking way. Right now, we
require a `TrafficRoute` to exist in order to route traffic if there are _no_
`Mesh*Routes` matching a given data plane proxy. This will no longer be
necessary.

#### Destination is `HTTP`

|        | Both exist                    | Some `MeshHTTPRoute`, no `MeshTCPRoute` | No `MeshHTTPRoute`, no `MeshTCPRoute` | No `MeshHTTPRoute`, some `MeshTCPRoute` |
| ------ | ----------------------------- | --------------------------------------- | ------------------------------------- | --------------------------------------- |
| `HTTP` | configured by `MeshHTTPRoute` | configured by `MeshHTTPRoute`           | configured by `MeshHTTPRoute`         | configured by `MeshTCPRoute`, as TCP    |

#### Destination is `TCP`

Routes to `TCP` protocol services are always configured by `MeshTCPRoute`.

### Positive Consequences <!-- optional -->

- Users can transition to the intended behavior of `Mesh*Route` where no
  resource is necessary

### Negative Consequences <!-- optional -->

None?
