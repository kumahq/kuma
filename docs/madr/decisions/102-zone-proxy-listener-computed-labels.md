# Standardize Computed Labels for Mesh-Scoped Zone Proxy Listeners

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/15567

## Context and Problem Statement

With mesh-scoped zone proxies, zone ingress and zone egress are listeners on a `Dataplane`
([MADR 095](095-mesh-scoped-zone-ingress-egress.md)).

`targetRef.sectionName` is enough for precise listener targeting, but we still need a simple
way to match all Dataplanes that expose a zone ingress or zone egress listener.

## Decision Drivers

- Users need a coarse-grained label-based selector when using `spec.targetRef` to apply policies to all zone ingress or zone egress Dataplanes.
- Inspect API and GUI need to query Dataplanes exposing zone ingress, zone egress, or both, without reconstructing listener presence from the full Dataplane spec.

## Design

### Option 1: Reuse `kuma.io/proxy-type`

Continue using `kuma.io/proxy-type` to identify zone proxies.

* Good, because it already exists
* Bad, because it is too coarse for the listener-based model
* Bad, because a combined proxy can expose both listener types at the same time

### Option 2: Standardize listener-specific computed labels

Add new auto-computed Dataplane labels:

```yaml
kuma.io/dataplane-listener-zoneingress: enabled
kuma.io/dataplane-listener-zoneegress: enabled
```

Rules:

- If a Dataplane has at least one `ZoneIngress` listener, set `kuma.io/dataplane-listener-zoneingress: enabled`
- If a Dataplane has at least one `ZoneEgress` listener, set `kuma.io/dataplane-listener-zoneegress: enabled`
- If a Dataplane has both listener types, set both labels
- If a Dataplane has neither listener type, set neither label

## Security implications and review

None.

## Reliability implications

Standardized labels avoid duplicating listener-detection logic in multiple places.

## Implications for Kong Mesh

None.

## Decision

We standardize two new auto-computed Dataplane labels for mesh-scoped zone proxy listeners:

- `kuma.io/dataplane-listener-zoneingress: enabled`
- `kuma.io/dataplane-listener-zoneegress: enabled`

These labels are set based on listener presence on the Dataplane.
They are the coarse-grained selection mechanism for zone proxy listener type.
For exact listener targeting, use `targetRef.sectionName`.

## Notes

- This complements MADR 095 and does not change its listener model.
