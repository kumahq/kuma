# `MeshService` Kubernetes UX

* Status: accepted

Technical Story: #9704

## Context and Problem Statement

The new `MeshService` resource in Kubernetes zones maps 1-1 to Kubernetes
`Service` resources. Additionally, users should not have to manage
`MeshServices` directly but instead a `MeshService` should be derived directly
from its `Service` with the possibility of users annotating the `Service`
to control specifics.

## Considered Options

* CP manages `MeshService` from `Service`
* Users have to create `MeshService`

## Decision Outcome

We will convert `Services` to `MeshServices` in a Kubernetes controller,
assuming:

1. The namespace is part of the mesh with `kuma.io/sidecar-injection: "true"`
1. The `Mesh` of the `Service` or its namespace exists.
1. The `Service` hasn't opted out via `kuma.io/ignore: "true"`

### Metadata

The `ownerReference` of the `MeshService` points to the `Service`.
Labels are copied from the `Service` object to the `MeshService` object.

### VIP

For `MeshServices` in Kubernetes zones, we no longer create a VIP but instead
reuse the `ClusterIP` field allocated by the Kubernetes control plane and add
this to `status.vips` on `MeshService`.

Note this happens asynchronously.

### `ports`

The `MeshService` ports are derived from the `ports` field on `Service`,
including supporting named `targetPorts`.

Note that we only support `Service.ports[].protocol: TCP`, which is also the
default.

### Positive Consequences

* Users don't have to think about creating `MeshService`
* We don't need to allocate additional VIPs

### Negative Consequences

* None?

## Pros and Cons of the Options

### Manual management

* Bad, because almost always the `MeshService` should be entirely derived from
  the `Service` object
