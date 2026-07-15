# Propagate MeshService identity into outbound listener tags

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/17381

## Context and Problem Statement

`MeshProxyPatch` (and the legacy `ProxyTemplate`) let users select a generated Envoy listener and patch it. Selection happens on three fields: `origin`, `name`, and `tags`/`listenerTags`.

Enabling `MeshService` silently breaks the `tags`/`listenerTags` selector for outbound listeners, and the `name` selector is not a usable replacement. A user who turns on `MeshService` finds that policies which worked before now apply to nothing, with no error, no warning, and no validation failure.

### Current state: how a listener gets its tags

`io.kuma.tags` is a listener-only concept. Nothing in Envoy consumes it; it exists purely so that Kuma's own patching policies can select listeners. It has exactly one writer in the whole codebase:

```go
// pkg/xds/envoy/listeners/v3/tags_metadata.go:15-27
func (c *TagsMetadataConfigurer) Configure(l *envoy_api.Listener) error {
	// ...
	l.Metadata.FilterMetadata[envoy_metadata.TagsKey] = &structpb.Struct{
		Fields: envoy_metadata.MetadataFields(c.Tags),
	}
	return nil
}
```

with `TagsKey = "io.kuma.tags"` (`pkg/xds/envoy/metadata/v3/metadata.go:99`).

It has six readers, all of them Kuma's own matching code — three in `MeshProxyPatch` (`listener_mod.go:86`, `network_filter_mod.go:136`, `http_filter_mod.go:159`) and three in the legacy `ProxyTemplate` modification stack (`pkg/xds/generator/modifications/v3/{listener,network_filter,http_filter}.go`). All six do the same thing:

```go
// pkg/plugins/policies/meshproxypatch/plugin/v1alpha1/listener_mod.go:84-91
if len(pointer.Deref(l.Match.Tags)) > 0 {
	if listenerProto, ok := listener.Resource.(*envoy_listener.Listener); ok {
		listenerTags := envoy_metadata.ExtractTags(listenerProto.Metadata)
		if !mesh_proto.TagSelector(pointer.Deref(l.Match.Tags)).Matches(listenerTags) {
			return false
		}
	}
}
```

`TagSelector.Matches` (`api/mesh/v1alpha1/dataplane_helpers.go:465-479`) is an AND of exact key/value pairs, with `*` as a match-any-value wildcard. A key that is **absent** from the listener's tags fails the match.

Clusters never carry `io.kuma.tags` — `ClusterMatch` has only `origin` and `name`, no tags field at all. Cluster and endpoint selection use unrelated keys (`envoy.lb`, `io.kuma.labels`, transport-socket matches) written by `EndpointMetadata`. This bounds the blast radius of anything we do to `io.kuma.tags`: it cannot affect load-balancer subsetting, mTLS transport-socket matching, or endpoint selection.

### What we propagated for `kuma.io/service`

In the pre-`MeshService` world, an outbound was declared on the `Dataplane` itself and carried a user-authored tag map:

```yaml
networking:
  outbound:
    - port: 10001
      tags:
        kuma.io/service: backend
        version: v1
        region: us
```

The generators copied that map, verbatim, into the listener metadata — minus `kuma.io/mesh`:

```go
// pkg/xds/generator/outbound_proxy_generator.go:120 (legacy generator)
Configure(envoy_listeners.TagsMetadata(envoy_tags.Tags(outbound.GetTags()).WithoutTags(mesh_proto.MeshTag)))

// pkg/plugins/policies/meshtcproute/plugin/v1alpha1/listeners.go:41 (route plugin)
tags := envoy_tags.Tags(svc.Outbound.TagsOrNil()).WithoutTags(mesh_proto.MeshTag)
```

So the propagated data set was: **the whole outbound tag map, less `kuma.io/mesh`**. In practice that meant `kuma.io/service` plus whatever routing dimensions the user had modelled (`version`, `region`, ...). `kuma.io/mesh` was stripped because cross-mesh outbounds set it and it is not a property of the destination the user is targeting.

This gave the well-known and widely-documented pattern:

```yaml
default:
  appendModifications:
    - networkFilter:
        operation: Patch
        match:
          origin: outbound
          listenerTags:
            kuma.io/service: backend
        value: |
          ...
```

Inbound listeners work the same way and are **not** affected by `MeshService`: inbounds are still described by the `Dataplane` resource, so `iface.GetTags()` is always populated (`pkg/xds/generator/inbound_proxy_generator.go:96`). The inbound/outbound asymmetry is the heart of this problem.

### What breaks with MeshService

With `MeshService`, an outbound is no longer a user-authored tag bag. It is a reference to a real resource:

```go
// pkg/core/xds/types/outbound.go:8-15
type Outbound struct {
	// LegacyOutbound is an old way to define outbounds using 'kuma.io/service' tag
	LegacyOutbound *mesh_proto.Dataplane_Networking_Outbound

	Address  string
	Port     uint32
	Resource kri.Identifier
}
```

`LegacyOutbound != nil` discriminates the old world; a non-empty `Resource` (a KRI to `MeshService`, `MeshMultiZoneService`, or `MeshExternalService`) discriminates the new one. And tags only exist in the old one:

```go
// pkg/core/xds/types/outbound.go:43-49
// TagsOrNil returns tags if Outbound is defined using 'kuma.io/service' tag and so LegacyOutbound field is set.
// Otherwise, it returns nil.
func (o *Outbound) TagsOrNil() map[string]string {
	if o.LegacyOutbound != nil {
		return o.LegacyOutbound.Tags
	}
	return nil
}
```

`WithoutTags` on a nil map returns nil, `TagsMetadata(nil)` is still called, and the configurer writes the key unconditionally. The result is not an absent key but an **empty struct**: `io.kuma.tags: {}`. Every non-empty `TagSelector` therefore fails on the first key lookup, and the policy quietly matches nothing.

### Why `match.name` is not a workaround

```go
// pkg/plugins/policies/meshtcproute/plugin/v1alpha1/listeners.go:30-40
address := svc.Outbound.GetAddressWithFallback("127.0.0.1")
port := svc.Outbound.GetPort()

listenerName := envoy_names.GetOutboundListenerName(address, port)  // "outbound:<address>:<port>"
if id, ok := svc.Outbound.AssociatedServiceResource(); ok && unifiedNaming {
	listenerName = id.String()  // KRI — only when unified naming is enabled
}
```

Without unified resource naming the name is `outbound:<address>:<port>`, where `<address>` is the dynamically-allocated `MeshService` VIP. A user cannot know it ahead of time and it is not stable across restarts, so `match.name` cannot be authored:

```yaml
match:
  name: outbound:10.43.0.17:80   # VIP is dynamic — unusable
```

Only with unified naming (`meshServices: Exclusive` **and** the DP `UnifiedResourceNaming` feature) does the name become a stable KRI. That is a migration users cannot always take today, and it is a heavier switch than "I want to patch one listener".

### Use cases

* **UC1 — Migration.** A user with a working `MeshProxyPatch` matching `listenerTags: {kuma.io/service: backend}` enables `MeshService`. Today their policy silently stops applying, and they need some deterministic selector to move to.
* **UC2 — Target one destination.** Patch the outbound listener to `MeshService` `backend` in namespace `kuma-demo`, without knowing its VIP and without enabling unified resource naming.
* **UC3 — Target one port.** A `MeshService` with ports `80` and `8080` produces two outbound listeners, and they must be distinguishable from one another.
* **UC4 — Target a group.** Patch every outbound listener to any destination the user has labelled `team: payments`, or every destination in zone `east`.
* **UC5 — Parity across destination kinds.** The same mechanism should work for `MeshService`, `MeshMultiZoneService`, and `MeshExternalService`, since all three flow through the same outbound path.

### What identity data is actually available

At generation time the full destination resource — metadata and labels included — is already resolved:

```go
// pkg/plugins/policies/core/xds/meshroute/listeners.go:147-167
if svc = meshCtx.GetServiceByKRI(outbound.Resource); svc == nil {
	continue
}
if port, ok = svc.FindPortByName(outbound.Resource.SectionName); !ok {
	continue
}
result = append(result, DestinationService{
	Outbound:            outbound,
	Protocol:            protocol,
	KumaServiceTagValue: destinationname.MustResolve(false, svc, port),
})
```

`GetServiceByKRI` returns a `core.Destination`, which embeds `core_model.Resource` and therefore exposes `GetMeta().GetLabels()`. `DestinationService` (`meshroute/listeners.go:68-72`) currently drops it, keeping only the KRI and a synthesized name. The data we need is present; we just throw it away.

A `MeshService` carries its identity almost entirely in labels, not in its spec — the spec holds `Selector`, `Ports`, `Identities`, and `State`, while name, zone, namespace, origin, and env live in `meta.labels`. The computed set is centralised in `pkg/core/resources/labels/registry.go:10-23` and applied by `labels.Compute`:

```go
// pkg/core/resources/apis/meshservice/generate/generator.go:272-284 (Universal generator)
out[metadata.KumaMeshLabel] = mesh                        // kuma.io/mesh
out[mesh_proto.DisplayName] = name                        // kuma.io/display-name
out[mesh_proto.ManagedByLabel] = managedByValue           // kuma.io/managed-by
out[mesh_proto.EnvTag] = mesh_proto.UniversalEnvironment  // kuma.io/env
out[mesh_proto.ZoneTag] = zone                            // kuma.io/zone
out[mesh_proto.ResourceOriginLabel] = string(mesh_proto.ZoneResourceOrigin)  // kuma.io/origin
```

plus user labels, when `KUMA_MESH_SERVICE_LABEL_PROPAGATION_ENABLED` is on. On Kubernetes the controller adds `k8s.kuma.io/namespace` and `k8s.kuma.io/is-headless-service` (`pkg/plugins/runtime/k8s/controllers/meshservice_controller.go:373-489`). Note there is no `kuma.io/namespace` label in Kuma — the namespace label is `k8s.kuma.io/namespace` (`mesh_proto.KubeNamespaceTag`).

The KRI is a **lossy projection** of that label set. `kri.FromResourceMetaE` (`pkg/core/kri/kri.go:53-90`) reads exactly `kuma.io/display-name`, `kuma.io/zone`, and `k8s.kuma.io/namespace`, plus `meta.mesh` and the type's short name. `kuma.io/origin`, `kuma.io/env`, `kuma.io/managed-by`, and every user label are **not** encoded in the KRI string. So KRI alone cannot satisfy UC4.

Two precedents are worth naming. `MeshServiceLabelPropagation` (`pkg/config/app/kuma-cp/config.go:593-600`) already propagates allowed user labels from `Dataplane`s into the generated `MeshService`, with reserved keys rejected at validation time. And `LbLabelsKey = "io.kuma.labels"` (`pkg/xds/envoy/metadata/v3/metadata.go:104`) already carries workload labels in Envoy filter metadata — but on **endpoints**, sourced from the `Dataplane`. No `MeshService` label reaches Envoy metadata anywhere today.

## Design

### Option 1: Do nothing — require unified resource naming and `match.name`

Tell users that tag matching on outbound listeners is gone, and that the path forward is unified naming plus `match.name: kri_msvc_default_east_kuma-demo_backend_httpport`.

* Good, because it needs no code change at all.
* Good, because KRI is exact and unambiguous.
* Bad, because it does not solve UC1 — the regression stays silent, and users discover it in production.
* Bad, because it couples a small patching need to a large, mesh-wide migration (`meshServices: Exclusive` plus a DP feature flag).
* Bad, because it cannot express UC4 — a KRI names exactly one resource, so "every destination labelled `team: payments`" needs one policy per destination, regenerated whenever the set changes.
* Bad, because it leaves `io.kuma.tags: {}` on the listener as a permanent trap: the key is present and empty, which reads as "this listener has no tags" rather than "tags do not apply here".

### Option 2: Synthesize a legacy `kuma.io/service` tag

`CollectServices` already computes `KumaServiceTagValue` for real-resource outbounds, so we could write `kuma.io/service: <KumaServiceTagValue>` into `io.kuma.tags` and claim backward compatibility.

* Good, because it is a two-line change.
* Good, because the key `kuma.io/service` is what existing policies already reference, so nothing needs rewriting.
* Bad, because the value does not match and the compatibility is an illusion — for a real resource the value comes from `destinationname.ResolveLegacyFromDestination`, which produces `<mesh>_<name>_<namespace>_<zone>_<shortName>_<port>`, for example `default_backend_kuma-demo_east_msvc_80`, whereas the user's pre-migration tag was `backend` (Universal) or `backend_kuma-demo_svc_80` (Kubernetes), so a policy matching `kuma.io/service: backend` still matches nothing.
* Bad, because it is worse than doing nothing: the key now exists, so the failure moves from "obviously empty" to "present but subtly wrong", and `kuma.io/service: '*'` starts matching while `kuma.io/service: backend` does not.
* Bad, because it re-entrenches the naming scheme that `MeshService` and KRI exist to replace.

### Option 3: New `io.kuma.labels` metadata key on listeners, new match fields

Write the destination's labels under `io.kuma.labels` (reusing the endpoint-side key name) and add `listenerLabels` / `labels` match fields to the `MeshProxyPatch` API.

* Good, because it is semantically honest: labels are labels, tags are tags, and it mirrors the endpoint-side `LbLabelsKey` precedent instead of overloading an existing key.
* Good, because it keeps the two concepts separately evolvable.
* Bad, because it is a public API change: new fields on `ListenerMatch`, `NetworkFilterMatch`, and `HTTPFilterMatch`, regenerated CRDs and OpenAPI, plus the equivalent on the legacy `ProxyTemplate` proto or an explicit decision not to support it there.
* Bad, because users must then know which of two selectors applies to which listener, and the answer depends on whether the destination is legacy or a real resource — leaking an internal distinction straight into the API.
* Bad, because it doubles the matching code path in six places for no behavioural gain over Option 4: the matcher is the same `TagSelector`, over a different map, under a different key.

### Option 4 (chosen): Propagate destination labels and a `kuma.io/kri` tag into `io.kuma.tags`

For outbound listeners backed by a real resource, populate `io.kuma.tags` with the destination resource's labels minus `kuma.io/mesh` (the same strip the legacy path performs, for the same reason), plus a reserved `kuma.io/kri` tag holding the full KRI of the destination including its section name.

No API change. `io.kuma.tags` keeps its meaning — "the identifying key/value set of this listener" — and simply stops being empty in the `MeshService` world.

For a `MeshService` named `backend` in namespace `kuma-demo`, zone `east`, mesh `default`, port `httpport`, the outbound listener metadata becomes:

```yaml
metadata:
  filterMetadata:
    io.kuma.tags:
      kuma.io/kri: kri_msvc_default_east_kuma-demo_backend_httpport
      kuma.io/display-name: backend
      k8s.kuma.io/namespace: kuma-demo
      kuma.io/zone: east
      kuma.io/env: kubernetes
      kuma.io/origin: zone
      kuma.io/managed-by: k8s-controller
      team: payments          # user label, when label propagation is enabled
```

UC2 — target one destination, with no unified naming required:

```yaml
match:
  origin: outbound
  listenerTags:
    kuma.io/display-name: backend
    k8s.kuma.io/namespace: kuma-demo
```

UC3 — target one port. The `kuma.io/kri` tag carries the section name, so the two listeners of a two-port `MeshService` differ:

```yaml
match:
  origin: outbound
  listenerTags:
    kuma.io/kri: kri_msvc_default_east_kuma-demo_backend_httpport
```

This is the deterministic, address-independent selector the issue asks for, and it is available regardless of whether unified resource naming is enabled. That is precisely the gap Option 1 leaves open: unified naming changes the listener's *name*, whereas this changes its *tags*, and the two are independent.

UC4 — target a group:

```yaml
match:
  origin: outbound
  listenerTags:
    team: payments
```

UC5 — parity. `MeshMultiZoneService` and `MeshExternalService` reach `GenerateOutboundListener` through the same `DestinationService`, and `core.Destination` exposes labels for all three, so they get the same treatment for free. Only the `kuma.io/kri` short name differs (`msvc`, `mzsvc`, `extsvc`).

Advantages:

* No public API change, no CRD or OpenAPI regeneration, and no new concept for users to learn.
* Works without unified resource naming, which is the whole point of the issue.
* Purely additive, and provably so — for real-resource outbounds `io.kuma.tags` is empty today, so no currently-matching policy can stop matching; the change can only turn non-matches into matches.
* Fixes the legacy `ProxyTemplate` stack at the same time, since it reads the same key through the same `ExtractTags` helper.
* Reuses `TagSelector`, so `*` wildcards and multi-key AND semantics work unchanged.

Disadvantages:

* The key is named `tags` but now carries labels, which is a real wart — mitigated by the fact that `io.kuma.tags` was never "Dataplane tags" to Envoy or to anything else; it is Kuma's private listener selector bag, with one writer and six readers, all ours.
* Two `MeshService`s in different meshes with identical labels produce identical tags except for `kuma.io/kri` — acceptable, since a proxy's outbounds are mesh-scoped and `kuma.io/kri` disambiguates.
* A pre-existing policy matching, say, `listenerTags: {kuma.io/zone: east}` and intended for inbound or legacy listeners will now also match `MeshService` outbound listeners; see "Reliability implications".

#### Rejected: transparent compatibility for `kuma.io/service`

We considered emitting a `kuma.io/service` tag whose value reproduces the user's pre-migration tag, so that UC1 needs no policy edit. We rejected it. The pre-migration value is a property of the *old* `Dataplane` outbound and is not recoverable from a `MeshService` in the general case: a `MeshService` is generated from a `Service` (Kubernetes) or grouped from `Dataplane` inbounds (Universal), and neither preserves the original `kuma.io/service` string as a label. Any value we synthesize is a guess, and Option 2 explains why a wrong guess is worse than an honest absence. Users migrating to `MeshService` are already rewriting `targetRef`s; rewriting `listenerTags` keys is consistent with that, and it fails loudly at review time rather than silently at runtime.

#### Should this be gated by config?

No. `MeshServiceLabelPropagation` is gated because it writes to a *stored resource* — `MeshService` labels are user-visible, KDS-synced, and can collide with operator-managed labels. Listener metadata is generated, ephemeral, and consumed only by Kuma's own matcher. Combined with the "cannot break an existing match" argument above, a flag would add a knob whose `false` value has no legitimate use.

User labels reaching the listener remain gated *upstream* by `KUMA_MESH_SERVICE_LABEL_PROPAGATION_ENABLED`, since that is what decides whether a `MeshService` carries user labels at all. We do not add a second gate.

#### Evolution of this design

The first draft proposed writing only the destination labels. Review of UC3 showed this could not distinguish the listeners of a multi-port `MeshService` — labels are per-resource, listeners are per-port — so `kuma.io/kri` was added, carrying the section name. That turned out to also close the second half of the issue (the non-deterministic `match.name` fallback), which the labels-only draft did not address: it gives a stable, address-free, per-listener selector without requiring unified resource naming.

A later draft proposed dropping labels entirely and shipping only `kuma.io/kri`, on the grounds that KRI is the canonical identity. That was rejected once we confirmed KRI is a lossy projection of the label set — it encodes only display-name, zone, namespace, mesh, and type, so `kuma.io/origin`, `kuma.io/env`, and all user labels are absent from it, and UC4 becomes inexpressible.

The third draft placed labels under a new `io.kuma.labels` listener key with new API fields (Option 3). It was dropped once it became clear that `io.kuma.tags` is not a user-facing "Dataplane tags" surface but Kuma's own selector bag, making the API split cost real and its benefit cosmetic.

## Security implications and review

Low. The propagated data is `MeshService` metadata that the proxy's owner can already read — the outbound exists precisely because the destination is reachable from this proxy, and its VIP, port, and KRI are already in the config dump.

* Listener metadata is visible in `/config_dump` on the Envoy admin API, so any user label propagated onto a `MeshService` becomes visible to anyone who can reach that endpoint. This is already true for endpoint labels under `io.kuma.labels`, and the existing `AllowedLabelKeys` allowlist is the control point for operators who care.
* `kuma.io/kri` is derived, never user-supplied, so it cannot be spoofed by labelling a resource.
* A user can set an arbitrary label such as `kuma.io/display-name` on a `MeshService` and shadow a reserved tag, so reserved `kuma.io/`-prefixed tags computed by the control plane must be written **last** and win over propagated labels. `IsReservedLabelKey` (`api/mesh/v1alpha1/dataplane_helpers.go:31-33`) already exists and the label-propagation config already rejects reserved keys at validation time, so this is consistent with what we do elsewhere.

## Reliability implications

* **No existing match can break.** Real-resource outbound listeners have empty `io.kuma.tags` today, so every non-empty selector already fails against them; adding tags is strictly monotonic and can only add matches.
* **New matches are possible, and that is the intended behaviour** — but it is a behaviour change, so it needs a changelog entry. A policy written as `listenerTags: {kuma.io/zone: east}` with no `origin` constraint, intended for inbound listeners, will start matching `MeshService` outbound listeners too. The blast radius is limited to policies that match on `kuma.io/`-prefixed keys, do not constrain `origin`, and run on a proxy with real-resource outbounds.
* **No effect on traffic.** `io.kuma.tags` is listener-only and read by nothing except Kuma's matcher; clusters and endpoints use `envoy.lb`, `io.kuma.labels`, and transport-socket matches, which this change does not touch. There is no path from this change to load-balancing, mTLS, or endpoint selection.
* **KDS-synced resources.** `labels.Compute` skips recomputation for non-locally-originated resources (`pkg/core/resources/labels/compute.go:124-128`), so a `MeshService` synced global-to-zone keeps the labels it arrived with. Tags derived at generation time must therefore tolerate labels computed by a *different* control plane — in particular, a `MeshMultiZoneService` has no `kuma.io/zone` label, so its KRI has an empty zone segment. This is semantically correct (it is zone-spanning) but means `kuma.io/zone` is simply absent from its tags rather than empty.
* **Config churn.** Listener metadata now changes when a `MeshService`'s labels change, so a label edit triggers an xDS update for every proxy with an outbound to it. Labels change rarely and the listener is regenerated on any `MeshService` change already, so the delta is bounded.
* **Golden files.** Every outbound-listener golden file for a real-resource destination gains an `io.kuma.tags` block — a large, mechanical diff (`UPDATE_GOLDEN_FILES=true make test`) that is worth reviewing for unintended label leakage rather than rubber-stamping.

## Implications for Kong Mesh

The downstream project ships policies that patch outbound listeners by tag and will inherit both the fix and the `kuma.io/service` to `kuma.io/display-name` rewrite. Its own `MeshProxyPatch`-equivalent policies and any bundled defaults that select outbound listeners by `listenerTags` need an audit before this ships. No API or CRD change is required there, since none is required here. This decision adds no entry to `AllComputedLabels`, so the label registry mirrored downstream (`pkg/core/resources/labels/registry.go:8-9`) does not need to be kept in sync for this change.

## Decision

For outbound listeners backed by a real resource (`MeshService`, `MeshMultiZoneService`, `MeshExternalService`), populate `io.kuma.tags` with:

1. the destination resource's labels, minus `kuma.io/mesh`, matching the strip the legacy path already performs; and
2. a reserved `kuma.io/kri` tag carrying the destination's full KRI including section name — a deterministic, address-independent, per-listener selector that does not require unified resource naming.

Reserved `kuma.io/`-prefixed tags are computed by the control plane and take precedence over propagated user labels. The change is unconditional and not gated by a new config flag; user labels remain gated upstream by the existing `KUMA_MESH_SERVICE_LABEL_PROPAGATION_ENABLED`. No `MeshProxyPatch` or `ProxyTemplate` API change is made — both read `io.kuma.tags` through `ExtractTags` and are fixed by the same change.

Existing policies matching `kuma.io/service: <name>` must be rewritten to match on `kuma.io/display-name: <name>` (or `kuma.io/kri`) once `MeshService` is enabled. We do not synthesize a compatibility `kuma.io/service` tag, because the pre-migration value is not recoverable from a `MeshService` and a wrong value fails more quietly than an absent one.

Implementation notes:

* `DestinationService` (`pkg/plugins/policies/core/xds/meshroute/listeners.go:68-72`) drops the resolved `core.Destination`; it must carry the labels (or the resource) through from `CollectServices`, which already resolves it via `GetServiceByKRI` and needs no new lookup.
* Both `GenerateOutboundListener` implementations — `meshtcproute/plugin/v1alpha1/listeners.go:41` and `meshhttproute/plugin/v1alpha1/listeners.go:78` — must be updated; they currently share the `TagsOrNil().WithoutTags(MeshTag)` line.
* The legacy branch (`LegacyOutbound != nil`) keeps its current behaviour unchanged.

## Notes

* Open topic: whether `kuma.io/kri` should also be added to **inbound** listener tags for symmetry. Out of scope here, since inbounds are not broken — the `Dataplane` still supplies their tags.
* Open topic: `KUMA_EXPERIMENTAL_INBOUND_TAGS_DISABLED` empties **inbound** listener tags (`pkg/plugins/runtime/k8s/controllers/inbound_converter.go:52-57`), producing the same silent-no-match failure on the inbound side. Other plugins have grown workarounds; `MeshProxyPatch` has none. That is a separate issue with the same shape and deserves its own MADR.
* Open topic: whether the empty `io.kuma.tags: {}` written for listeners with no tags should become an omitted key. It is a latent readability wart, but changing it touches golden files across the board and is not needed for this fix.
* Inconsistency spotted while investigating: `cluster_mod.go:65-71` stamps `Origin: metadata.OriginProxyTemplateModifications` while `listener_mod.go:66-72` stamps `Origin: metadata.OriginMeshProxyPatch`. Likely a leftover from the `ProxyTemplate` era; unrelated to this decision but worth a follow-up.
* Follow-up: `MeshProxyPatch` cannot warn when a selector matches nothing. A validation or inspect-time signal for "this policy matched zero resources" would have surfaced this regression at apply time rather than in production, and is worth its own issue. `origin` is likewise an unvalidated free-form string with no enum, so a typo silently matches nothing.
* MeshGateway listeners carry no `io.kuma.tags` at all (no `TagsMetadata` call under `pkg/xds/generator/gateway/`), so `listenerTags` can never match a gateway listener. Unchanged by this decision, but worth documenting.
