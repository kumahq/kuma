# `MeshGateway` as `targetRef.kind`

- Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/6590

## Context and Problem Statement

With `targetRef` policies, the only way to target a builtin gateway at the moment
is to use the `kuma.io/service` values the `MeshGateway` matches as
`MeshService`.

This may be somewhat unintuitive since a `MeshGateway` consists of not only a
`Dataplane`, where a `kuma.io/service` value is actually set, but also the
`MeshGateway` resource where the listener configuration is held.

It may be more intuitive to allow referring directly to a `MeshGateway` resource
but this isn't yet supported.

Another decision is how a policy can be restricted to a single listener of a
`MeshGateway`.

This MADR does not address narrowing to a specific route resource or route match.

### `spec.to.targetRef`

This MADR also makes clear that currently `spec.to.targetRef.kind: MeshGateway`
to set policy for incoming traffic is not supported in any policy.

Part of the reason for this is that the paradigm that `spec.targetRef` tells us
which sidecars have their config modified by a given policy would no longer apply.

### Policies targetting incoming traffic

We should consider which policies make sense with `spec.targetRef.kind: MeshGateway`
for targetting _incoming traffic_, i.e. not using `spec.to` but rather `spec.from`.
For this kind of policy we can imagine a new `spec.from.targetRef.kind: External`.

## Decision Drivers

- Allow users to refer directly to the `MeshGateway` resource itself

## Considered Options

- Leave out `MeshGateway` as a `kind`
- Support `MeshGateway` as a `kind` but not specific listeners
- Support `MeshGateway` as a `kind` with the option of setting
  `targetRef.listener`
- Support `MeshGateway` as a `kind` but also `MeshGatewayListener` where
  `targetRef.listener` can be set
- Support `MeshGateway` as a `kind` with the option of setting
  `targetRef.sectionName` to match a listener
- Support `MeshGateway` as a `kind` with the option of setting
  `targetRef.tags` to match a listener

## Decision Outcome

Chosen option: `MeshGateway` as `kind` with `targetRef.tags`

### Matching

`targetRef.kind: MeshGateway` applies the policy to all `Dataplanes` matched by
the matchers on the `MeshGateway` object and all listeners configured in the
`MeshGateway`. It has the same semantics as `{ kind: MeshService, name:
<kuma.io/service matched by the MeshGateway> }`

Further specifying `tags` narrows this to specific listeners with matching tags.

Given:

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshGateway
metadata:
  name: edge-gateway
mesh: default
spec:
  selectors:
    - match:
        kuma.io/service: edge-gateway
  conf:
    listeners:
      - port: 8080
        protocol: HTTP
        tags:
          protocol: http
      - port: 8443
        protocol: HTTPS
        tags:
          protocol: https
```

We can use the following policy with `tags` to target only traffic over
the `HTTPS` listener on port `8443`.

```yaml
apiVersion: kuma.io/v1alpha1
kind: SomePolicy
metadata:
  name: https-only
spec:
  targetRef:
    kind: MeshGateway
    name: edge
    tags:
      protocol: https
```

Note that in every policy implementation, we must make sure the Envoy config we
generate is coherent, given that more than one `spec.listeners` can be merged into
a single Envoy listener. There is no 1-1 correspondence guaranteed between Envoy
listeners and `MeshGateway` listeners.

### Specificity

The ordering of `kinds` that can affect builtin gateway resources is the
following:

`Mesh` > `MeshSubset` > `MeshGateway`

In other words, resources that target `Mesh` and `MeshSubset` also target
builtin gateway resources and are less specific than `MeshGateway`. See [the
policy matching MADR](./005-policy-matching.md#Overlapping) for more.

### Positive Consequences

- We can target a specific listener
- Can select one or more listeners
- Allows for flexibility of using existing `tags` field
- Gateway API's `sectionName` can be mapped to `tags`

### Negative Consequences

- More overhead than just naming listeners

## Other option

### `MeshGateway` with `targetRef.sectionName`

### Positive Consequences

- We can target a specific listener
- Allows for flexibility of using `sectionName` for _other_ `kinds` that may
  have the concept of narrowing applicability of policies
- Naming things is hard, `sectionName` is at least known from Gateway API

### Negative Consequences

- Further complicates the `targetRef` schema, we already have `.tags`
