# Including Policy Origin in Inbound Rules for Inspect API

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/14010

## Context and Problem Statement

Currently, the Inspect API for inbound rules looks like this:

```
GET :5681/meshes/{mesh}/dataplanes/{name}/_inbounds/{inbound-kri}/_policies
```
```yaml
policies:
  - kind: MeshTrafficPermission
    rules:
      - conf:
          deny:
            - spiffeId:
                type: Exact
                value: "spiffe://trust-domain.mesh/ns/default/sa/frontend"
      - conf:
          deny:
            - spiffeId:
                type: Exact
                value: "spiffe://trust-domain.mesh/ns/default/sa/api-gateway"
          allowWithShadowDeny:
            - spiffeId:
                type: Prefix
                value: "spiffe://trust-domain.mesh/ns/legacy"
          allow:
            - spiffeId:
                type: Prefix
                value: "spiffe://trust-domain.mesh/"
    origins:
      - kri: kri_mtp_...
      - kri: kri_mtp_...
```

It doesn't show the `origin` for each individual rule, but only displays a combined list of `origins`.

There are several reasons why we now want to see the `origin` for each individual rule:

1. We want to stop merging rules with the same `matches` (or without `matches`)
2. MeshTrafficPermission will translate each rule to Envoy configuration individually, and it needs the `kri` of the original policy

## Design

We will add a new field `policies[].rules[].kri` similar to what we have in the [RouteConf](https://github.com/kumahq/kuma/blob/b219aa49f11c7f7c38957d79a6c0c57d1b5f8d58/api/openapi/specs/common/resource.yaml#L257).

The resulting response will look like this:

```
GET :5681/meshes/{mesh}/dataplanes/{name}/_inbounds/{inbound-kri}/_policies
```
```yaml
policies:
  - kind: MeshTrafficPermission
    rules:
      - conf:
          deny:
            - spiffeId:
                type: Exact
                value: "spiffe://trust-domain.mesh/ns/default/sa/frontend"
        kri: kri_mtp_mesh-1___mesh-operator_
      - conf:
          deny:
            - spiffeId:
                type: Exact
                value: "spiffe://trust-domain.mesh/ns/default/sa/api-gateway"
          allowWithShadowDeny:
            - spiffeId:
                type: Prefix
                value: "spiffe://trust-domain.mesh/ns/legacy"
          allow:
            - spiffeId:
                type: Prefix
                value: "spiffe://trust-domain.mesh/"
        kri: kri_mtp_mesh-1__kuma-demo_service-owner_
    origins:
      - kri: kri_mtp_mesh-1___mesh-operator_
      - kri: kri_mtp_mesh-1__kuma-demo_service-owner_
```
