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
  or the `Service` is labelled with `kuma.io/mesh`.
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

### Headless Service with selectors

In Kubernetes, a headless Service with selectors is used to create a DNS record
for every Pod selected by the Service that points directly to the Pod's IP.

To support this with Kuma, we will create a `MeshService` per Pod, each
represented by the hostname allocated by the headless Service and the Pod
IP as the "VIP" and single endpoint.

In order to do this we need to have a list of all the Pods selected by the
Service, which we can get by looking at `EndpointSlices`. These resources hold a
list of endpoints, each of which has a `targetRef`. If the `targetRef` is `kind:
Pod`, we can rely on the naming of `Dataplane` objects and directly select a
given `Dataplane` by setting `spec.selector.dataplaneName` to the name of the
`Pod`.

```
kind: MeshService
spec:
  selector:
    dataplaneName: pod-1
    # dataplaneTags: ...
```

#### Policy matching

Note that this prevents using `kind: MeshService` to select all Pods of a
StatefulSet. In another MADR, we will cover this use case.

### Positive Consequences

* Users don't have to think about creating `MeshService`
* We don't need to allocate additional VIPs

### Negative Consequences

* None?

## Pros and Cons of the Options

### Manual management

* Bad, because almost always the `MeshService` should be entirely derived from
  the `Service` object
