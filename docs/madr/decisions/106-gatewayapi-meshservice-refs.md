# Accept MeshService as a Gateway API parentRef and backendRef

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/17985

## Context and Problem Statement

The GAMMA `HTTPRoute` translator (`pkg/plugins/runtime/k8s/controllers/gatewayapi/`) only understands a Kubernetes `Service` as a parent or a backend. Every `MeshService` that a `Service` generates is reachable that way, indirectly, through the `Service` that produced it. But a `MeshService` that no `Service` produced — a serviceless workload, or a hand-written destination such as a VM — has no `Service` to route through, so an `HTTPRoute` cannot reach it at all.

Gateway API does not define a `MeshService` reference, so this adds a Kuma-specific extension: `group: kuma.io, kind: MeshService` in `parentRefs` and `backendRefs`. Because it is a public interface, this MADR settles the shape before code, per the task brief.

## Design

The design mirrors the existing `Service` path in the same translator, rather than generalizing it into one code path for both kinds. `Service` and `MeshService` differ enough — a `Service` always carries `spec.ports[].port`, a `MeshService` can have zero ports, and a `MeshService` reference resolves by label rather than by living behind a `Service` object — that sharing one function would cost more in conditionals than it would save in duplication.

### Reference shape

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: my-route
  namespace: kuma-demo
spec:
  parentRefs:
    - group: kuma.io
      kind: MeshService
      name: backend
  rules:
    - backendRefs:
        - group: kuma.io
          kind: MeshService
          name: backend
          port: 80 # optional
```

The group is spelled out as `kuma.io` rather than left to default, because Gateway API's own default group for `backendRefs`/`parentRefs` is the core group (`""`, meaning `Service`) or `gateway.networking.k8s.io` for other Gateway API kinds. `MeshService` is neither, so an explicit group is required to disambiguate it from a typo'd `Service` reference, and it matches the group Kuma's own CRD (`meshservice_k8s.GroupVersion.Group`) already uses.

### Port-optional backendRefs

Gateway API requires `port` on a `Service` `backendRef`, because a `Service` can expose several ports and nothing else disambiguates them. A `MeshService` `backendRef` makes `port` optional instead:

* When `port` is set, it must match one of `spec.ports[].port` on the referenced `MeshService`. The matching port's name becomes the destination `sectionName`, so traffic is aimed at that single port.
* When `port` is omitted, the resulting `MeshHTTPRoute`'s `to[].targetRef` carries no `sectionName`, so it targets every port the `MeshService` exposes. This is a deliberate difference from `Service`: a `MeshService` is Kuma's own resource, and unlike a `Service`'s `spec.ports`, its ports are not always the point of the reference — some `MeshService`s exist to select a set of endpoints where every port shares the same route.

An unresolvable reference — the `MeshService` does not exist, or `port` does not match any `spec.ports[].port` — reports `ResolvedRefs=False` with reason `BackendNotFound`, the same behavior an unresolvable `Service` `backendRef` already has today.

### parentRef sectionName

Gateway API interprets `sectionName` on a `Service` parentRef as a port name, and requires an implementation that accepts another parent kind to state how it reads `sectionName` for that kind. On a `MeshService` parentRef it names a port too: it matches `Port.GetName()`, which is the port's `name` when it has one and the stringified `port` when it does not, so `sectionName: http` and `sectionName: "80"` both select a port. `sectionName` is Core support in Gateway API and `port` is experimental, so a `MeshService` parentRef honours the name, not only the number.

When a parentRef sets both, a port has to match both, which is what Gateway API asks for: the name and the port of the selected port must match both specified values. Setting neither targets every port the `MeshService` exposes.

The `Service` parentRef path still reads `port` alone. #17987 covers it, and it should adopt the same rule.

### Zero-port MeshServices

A `MeshService` can be created with no `spec.ports` at all (e.g. before its controller has observed any endpoints). A parentRef that names no port resolves successfully against such a `MeshService` — it is not treated as `NoMatchingParent` — and produces zero `to[]` entries, i.e. a route that matches but sends traffic nowhere. That keeps the route's status still while the `MeshService`'s port list is briefly empty during reconciliation.

A parentRef that does name a port, by `port` or by `sectionName`, is a different case. Nothing can match it, there is no flapping to protect against, and the reference is a plain user error, so the ref reports `Accepted=False` with reason `NoMatchingParent` and generates no route.

### parentRef routing and naming

A `MeshService` parentRef produces a `MeshHTTPRoute.kuma.io`, exactly like a `Service` parentRef does: a producer route (`targetRef.kind: Mesh`) when the `HTTPRoute` and the `MeshService` share a namespace, a consumer route (`targetRef.kind: Dataplane`, labeled by the route's namespace) otherwise. One `to[]` entry is generated per `MeshService` port (filtered to the referenced port when `parentRef.port` or `parentRef.sectionName` is set), each entry's `targetRef.kind` is `MeshService`, addressed by `kuma.io/display-name` and `k8s.kuma.io/namespace` labels — the same labels `KubernetesMetaAdapter.GetLabels` computes, so a display-name annotation override or a headless `Service`'s hashed `MeshService` name resolve identically whether the `MeshService` was hand-written or `Service`-generated.

The generated route's Kubernetes object name is
`<route>-<route-ns>-meshservice-<parent>.<parent-ns>`, with an explicit `meshservice` infix, unlike the `Service` path's `<route>-<route-ns>-<parent>.<parent-ns>`. This is required, not cosmetic: a `Service` auto-generates a `MeshService` of the same name and namespace, so an `HTTPRoute` that names both the `Service` and its generated `MeshService` as parents would produce the same sub-route key for both without the infix, and the second `Reconcile` write would silently overwrite the first in `common.ReconcileLabelledObject`'s owned-object map. The `Service` path's names are left unchanged to avoid regenerating every existing `MeshHTTPRoute` on upgrade.

### Missing parent handling

When a `MeshService` parentRef cannot be resolved, the ref's status gets `Accepted=False` (reason `NoMatchingParent`) and no route is generated for that ref — instead of silently dropping it, which is what the pre-existing `Service` parentRef path still does (a stop condition in this MADR's originating plan: the `Service` path's behavior is left as-is, TODO and all, to keep this change minimal). `ResolvedRefs` stays untouched by the parent lookup, because it reports on the route's own `backendRefs` and filters: a missing parent is not a backend problem, and reporting `BackendNotFound` there would tell a user whose `backendRefs` all resolve that they do not.

### Re-reconciling on MeshService changes

A new field index, `.metadata.meshservices`, and a `Watches(&meshservice_k8s.MeshService{}, ...)` registration make the `HTTPRoute` controller re-reconcile whenever a referenced `MeshService` is created, changed, or deleted — covering both `parentRefs` and `backendRefs` (including request-mirror backendRefs). This is a separate index from the existing `.metadata.services` one, kept that way because the `Service` index's coverage gap — it does not index `parentRefs` at all — is a pre-existing narrower bug this change does not fix.

## Security implications and review

None beyond what the `Service` path already exposes. A `MeshService` backendRef resolves to the same kind of destination (`targetRef.kind: MeshService`) a `Service` backendRef already resolves to; this change adds a second way to name that destination, not a new destination kind or a new trust boundary.

## Reliability implications

* Adds one field index and one `Watches` registration to the `HTTPRoute` controller; both are bounded by the number of `HTTPRoute`s that reference `MeshService`s.
* A `MeshService` update that changes only unrelated fields (e.g. status) still re-triggers reconciliation of routes that reference it, the same way `Service` updates do today — no new churn pattern is introduced.
* The `Service` path's condition set, sub-route names, and NotFound (silent-drop) behavior are unchanged, verified by the existing `gapiServiceToMeshRoute`/`uncheckedGapiToKumaRef` Service-path tests passing unmodified.

## Implications for Kong Mesh

None known; the enterprise fork does not currently define its own Gateway API MeshService handling. If it later adds MeshMultiZoneService or MeshExternalService support, it should follow the same `group: kuma.io` convention and the same asymmetric-naming rule this decision establishes, to avoid a second collision class with `Service`-generated `MeshService`s.

## Decision

`group: kuma.io, kind: MeshService` is accepted as a Gateway API `parentRef` and `backendRef` kind, alongside the existing `Service` kind, in the GAMMA `HTTPRoute` translator. `port` is optional on a `MeshService` `backendRef`; when set it must match a `MeshService` port, and when omitted the reference targets every port. A `backendRef` to a `MeshService` that does not exist, or with a `port` that does not match any of its ports, reports `ResolvedRefs=False`; an unresolvable `parentRef` reports `Accepted=False` rather than dropping the route silently. A `MeshService` `parentRef` reads `sectionName` as a port name and `port` as a port number, and a parentRef that names a port the `MeshService` does not have reports `Accepted=False`. `HTTPRoute`s re-reconcile when a referenced `MeshService` changes. The `Service` path — its conditions, sub-route names, and NotFound handling — is left byte-identical.

Out of scope for this decision: `MeshMultiZoneService` and `MeshExternalService` references, which are deferred to follow-up work once this shape has proven out; and any change to the pre-existing gaps in the `Service` path (its missing parentRef index, its silent drop on a missing parent).

## Notes

The public documentation site (`kumahq/kuma-website`) update for this reference shape is tracked as a follow-up in that repository, not in this PR.
