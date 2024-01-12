# Configuring many gateways

- Status: accepted

## Context and Problem Statement

This document aims to clarify how to configure all gateways or a subset of them, especially when adjustments, such as in `MeshTimeout`, are needed compared to regular Mesh traffic.

## Decision Drivers

- When we set up default rules, we want to use different settings for gateways than we do for regular Mesh applications.

## Considered Options

- Introduce new kind `MeshGatewaysSubset`
- Add label to all gateways and target them by kind `MeshSubset`
- Remove requirement of `name` when using kind `MeshGateway`
- Use internal tag `gateways: include/exclude/only` when using with `MeshSubset`

## Decision Outcome

- ?

## Introduce new kind `MeshGatewaysSubset`

We can introduce a new kind `MeshGatewaysSubset` It only picks gateways based on tags, if you provide tags. But, it can't pick a specific gateway by name.

```yaml
targetRef:
  kind: MeshGatewaysSubset
  tags: {} # not required
```

When there is no tags, policy targets all gateways.

Thanks to this we can allow users to target all gateways and apply some good default configuration for them.

> Note:
> This kind might target all listeners of the gateway, ignoring listener tags, except for policies that support targeting listeners by tags.

### Ordering

`MeshGatewaysSubset` is more specific than `MeshSubset` so its priority should be higher. Because `MeshGatewaysSubset` choose subset of Dataplanes that are gateways means is more specific.

Priority:
`Mesh` < `MeshSubset` < `MeshGatewaysSubset` < `MeshGateway` < `MeshService` < `MeshServiceSubset` < `MeshHTTPRoute`

### Positive Consequences

- Users can set things up for all gateways.
- We can use different defaults for gateways than for the whole Mesh.
- More explicit

### Negative Consequences

- The name might be a bit confusing.

## Add label to all gateways and target them by kind `MeshSubset`

We could label all gateways with `kuma.io/gateway` label. Currently, we are using this label only on Kubernetes for gateways with value `enabled` for builtin and `provided` for provided gateways. We could start labeling all gateways with it. That would allow us to create a policy with the kind `MeshSubset` and tag selector matching this label.

```yaml
targetRef:
  kind: MeshSubset
  tags:
    kuma.io/gateway: "enabled"
```

In this case, we need to propagate the label from the pod to the dataplane object.

### Positive Consequences

- Users can set things up for all gateways.
- We can use different defaults for gateways than for the whole Mesh.
- No need for a new Kind.

### Negative Consequences

- Less explicit
- We need to add the tag to all gateways.

### Remove requirement of `name` when using kind `MeshGateway`

We could remove requirement of the name when using `kind: MeshGateway` and when the name is missing policy targets all gateways.

### Positive Consequences

- Users can set things up for all gateways.
- No need for a new Kind.
- Not confusing names

### Negative Consequences

- Users can impact the configuration of all gateways by mistake
- Change of the current api
- Problem with ordering when there is more specific policy with the same kind but with name defined

## Use internal tag `gateways: include/exclude/only` when using with `MeshSubset

We can implement a code change that checks if a policy has the tag `gateway`. If the tag is present, we evaluate one of three possible values: `include`, `exclude`, or `only`. If the value of the tag doesn't match any of these possibilities, we use the actual tag value. If the tag is not defined, we consider it as the default value `include`.

```yaml
targetRef:
  kind: MeshSubset
  tags:
    gateways: "only"
```

### Positive Consequences

- Users can set things up for all gateways.
- No need for a new Kind.
- Not confusing names

### Negative Consequences

- Configuration is not so obvious
- What about merging `MeshSubset`'s
- Can break current behavior if the user has already defined the tag