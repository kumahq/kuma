# Source Envoy listener tags from resource labels

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/17381

## Context and Problem Statement

`MeshProxyPatch` (and the legacy `ProxyTemplate`) let users select a generated Envoy listener and patch it. A listener can be selected by `origin`, by `name`, or by `tags`/`listenerTags` — the last of which matches against a key/value bag Kuma stamps onto every listener as Envoy filter metadata under `io.kuma.tags`.

That bag has always been filled from `Dataplane` **tags**. Kuma is moving identity out of tags and into **labels**, and the listener metadata never followed. The result is that tag-based listener matching silently stops working — the policy still applies, still validates, and matches nothing.

### Three cases

The same listener-tag mechanism behaves differently in three configurations. Only the first works today.

| Case                              | Inbound listener tags                               | Outbound listener tags                                             |
|:----------------------------------|:----------------------------------------------------|:-------------------------------------------------------------------|
| **1. `kuma.io/service`** (legacy) | from the proxy's own `Dataplane` inbounds — works   | from the consuming `Dataplane`'s outbounds — works                 |
| **2. `MeshService` enabled**      | unchanged — still works                             | **empty** — outbounds became resource references and carry no tags |
| **3. `InboundTagsDisabled`**      | **empty** — inbound tags are stripped at the source | unchanged by this flag                                             |

Cases 2 and 3 are independent switches, but they are the same migration and are commonly enabled together. Enabling both leaves a proxy with no usable tag-based listener selector in either direction.

**Case 1 — `kuma.io/service`.** The destination's identity is a string, relayed from proxy to proxy. Its journey is: the destination `Dataplane`'s inbound `kuma.io/service` tag, into the VIP view, into the *consuming* `Dataplane`'s outbound tags, and finally onto the outbound listener. Two things about it matter. The tags describe the destination but live on the consumer. And in the generated path — all of Kubernetes, every transparent-proxy deployment — the bag holds exactly **one** key, `kuma.io/service`, because the control plane hard-codes it. Multi-tag outbounds (`version`, `region`) only ever existed in hand-written Universal `Dataplane`s.

**Case 2 — `MeshService`.** The destination stops being a relayed string and becomes a real resource, referenced by KRI. The outbound tag bag has nowhere to come from, so it is empty. Critically, `kuma.io/service` is not merely absent from the listener — it is **removed from the model**. The consumer's outbound tags are gone, the VIP relay no longer feeds outbounds, and the destination's inbounds do not carry the tag either. There is no value anywhere to fetch, so no design can keep `listenerTags: {kuma.io/service: backend}` working. Those policies must be rewritten; the only question is what they are rewritten *to*.

**Case 3 — `InboundTagsDisabled`.** Inbound tags are emptied at the source, so inbound listeners hit the identical wall from the other direction. The important detail is that the underlying data was never lost: on Kubernetes, Pod labels are copied onto the `Dataplane`'s **metadata labels** unconditionally, flag or no flag. The flag does not remove information — it removes the *duplicate copy* held in tags, forcing everything onto labels.

### The tags-to-labels move is already the established direction

Case 3 is not a regression Kuma stumbled into; it is a deliberate migration that most of the codebase has already made. Identity now reads from labels:

* The service name comes from the `kuma.io/workload` **label** via `IdentifyingName`, used by the monitoring assignment service, `MeshTrace`, `MeshAccessLog`, and `MeshMetric`.
* An inbound's identity comes from the `Dataplane`'s **KRI**, which is itself derived from labels.
* The Kubernetes `MeshService` selector switches from `DataplaneTags` to `DataplaneLabels`.
* Endpoint metadata carries `Dataplane` labels under `io.kuma.labels`.

The substitution is real but **partial and opportunistic** — only four call sites were taught to read labels. Elsewhere the data is simply dropped: `ServiceInsight`s vanish for tagless dataplanes, and endpoint locality still reads tags only. Listener metadata is one of the places that was missed, and it is the one that breaks a user-facing policy API.

So the problem is not "MeshService broke `MeshProxyPatch`". It is that **listener tags are the last identity channel still sourced from tags, in a system whose identity now lives in labels.**

### Why `match.name` is not a workaround

With tags gone, `match.name` is the only selector left. Without unified resource naming a listener is named `outbound:<address>:<port>`, where the address is a dynamically-allocated VIP — not knowable ahead of time, not stable across restarts, so the policy cannot be written. Only with unified naming (`meshServices: Exclusive` plus a DP feature flag) does the name become a stable KRI. That is a mesh-wide migration, which is a heavy price for "I want to patch one listener". And a name selects exactly one listener, so it cannot express "every destination owned by team X".

### Use cases

* **UC1 — Migration.** A `MeshProxyPatch` matching `listenerTags: {kuma.io/service: backend}` silently stops applying once `MeshService` is on, and the user needs an authorable selector to move to.
* **UC2 — Target one destination.** Patch the outbound listener to `MeshService` `backend` in namespace `kuma-demo`, without knowing its VIP and without enabling unified resource naming.
* **UC3 — Target a group.** Patch every outbound listener to any destination labelled `team: payments`, or every destination in zone `east`.
* **UC4 — Target an inbound.** Patch a proxy's own inbound listener when `InboundTagsDisabled` has emptied its tags.
* **UC5 — Parity.** The same mechanism should work for `MeshService`, `MeshMultiZoneService`, and `MeshExternalService`, which share one outbound code path.

## Design

Whatever we choose has to answer one question: when a listener's tags are gone, what identifies the thing that listener represents? For an outbound, that is the **destination resource**. For an inbound, it is the **proxy's own `Dataplane`**. In both cases the answer is now labels.

### Option 1: Do nothing — require unified resource naming and `match.name`

* Good, because it needs no code change, and a KRI-based name is exact and distinguishes individual ports.
* Bad, because it does not address UC1 — the regression stays silent and users find it in production.
* Bad, because it couples a small patching need to a mesh-wide migration.
* Bad, because a name matches one listener, so UC3 needs one policy per destination, regenerated whenever the set changes.
* Bad, because it does nothing for UC4 — `InboundTagsDisabled` has no naming escape hatch.
* Bad, because it leaves an empty `io.kuma.tags` on every affected listener, which reads as "no tags here" rather than "tags do not apply here".

### Option 2: Synthesize a `kuma.io/service` tag

Rebuild a `kuma.io/service` value from the destination resource and keep writing it, so existing policies appear to work.

* Good, because it is a small change and needs no policy rewrites — at first glance.
* Bad, because `kuma.io/service` is a removed concept, not a missing value; reviving it on the listener contradicts the direction `MeshService` exists to take.
* Bad, because the synthesized value would not match anyway — the legacy resolver produces `<mesh>_<name>_<namespace>_<zone>_<shortName>_<port>`, while the user's policy says `backend`.
* Bad, because that makes it *worse* than doing nothing: the key exists again, so `kuma.io/service: '*'` starts matching while `kuma.io/service: backend` still does not. The compatibility is an illusion that fails at runtime instead of at review time.

### Option 3: A separate `io.kuma.labels` key on listeners, with new match fields

Write labels under a new listener key and add `listenerLabels`/`labels` to the `MeshProxyPatch` API.

* Good, because it is semantically tidy — labels and tags stay distinct.
* Bad, because it is a public API change across three match types, plus CRDs, OpenAPI, and a decision about the legacy `ProxyTemplate`.
* Bad, because users must then know which selector applies to which listener, and the answer depends on whether the destination is legacy or a real resource — leaking an internal distinction into the API.
* Bad, because the matcher, the semantics, and the outcome are identical to Option 4; only the key name differs. It buys tidiness and charges an API migration for it.

### Option 4 (chosen): Fill `io.kuma.tags` from labels when tags are gone

Keep one key and one selector. When a listener's tags are unavailable, source them from the labels of whatever the listener represents — the destination resource for an outbound, the proxy's own `Dataplane` for an inbound. `kuma.io/mesh` is stripped, as the outbound path already does today.

This is not a new mechanism; it applies the substitution Kuma has already made everywhere else to the last component that missed it.

A policy that used to say `kuma.io/service: backend` is rewritten as:

```yaml
match:
  origin: outbound
  listenerTags:
    kuma.io/display-name: backend
    k8s.kuma.io/namespace: kuma-demo
```

and UC3, which no name-based selector can express at all, becomes:

```yaml
match:
  origin: outbound
  listenerTags:
    team: payments
```

* Good, because there is no API change — the selector users already know starts working again, and the legacy `ProxyTemplate` is fixed by the same change since it reads the same key.
* Good, because it works without unified resource naming, which is the point of the issue.
* Good, because it is provably additive: the affected listeners have empty tags today, so no currently-matching policy can stop matching. It can only turn non-matches into matches.
* Good, because labels are a richer selector than the single `kuma.io/service` string ever was — zone, origin, env, and user labels become matchable, which is what makes UC3 possible.
* Good, because one rule covers both directions, closing UC4 with the same reasoning rather than a second mechanism.
* Bad, because the key is named `tags` but carries labels. A real wart — mitigated by the fact that `io.kuma.tags` is not a user-facing "Dataplane tags" surface but Kuma's private listener selector bag, with one writer and six readers, all of them ours.
* Bad, because **labels are per-resource while listeners are per-port**, so the two listeners of a two-port `MeshService` get identical tags and cannot be told apart by `listenerTags` alone. Narrowing to one port still needs `match.name` under unified naming. This MADR does not close that gap.

#### Why not gate this behind a flag

`MeshServiceLabelPropagation` is gated because it writes to a stored, user-visible, KDS-synced resource. Listener metadata is generated, ephemeral, and read only by Kuma's own matcher. Combined with the additive-by-construction argument above, a flag would have no legitimate `false` value. User labels remain gated *upstream* by `KUMA_MESH_SERVICE_LABEL_PROPAGATION_ENABLED`, which decides whether a `MeshService` carries user labels at all; we do not add a second gate.

#### How this design evolved

The first draft tried to preserve `listenerTags: {kuma.io/service: backend}` verbatim by synthesizing the tag. Tracing the provenance killed it: `kuma.io/service` is not mislaid but deleted, along with every link that fed it. Once it was clear no design could keep those policies untouched, the goal changed from "preserve the selector" to "provide the best selector to migrate to".

The second draft proposed a `kuma.io/kri` tag carrying the full KRI, to solve per-port targeting. Dropped: it duplicated the listener name under unified naming, and a KRI names one resource, so it did nothing for group targeting.

The third draft scoped the decision to outbound only, treating the `InboundTagsDisabled` gap as unrelated. Investigating case 3 showed it is the same defect reached by a different switch, with the same fix. Scoping to one direction would have meant deciding the same question twice.

## Security implications and review

Low. The propagated data is metadata the proxy's owner can already read — the outbound exists precisely because the destination is reachable, and its VIP and port are already in the config dump.

* Listener metadata is visible in `/config_dump`, so any user label propagated onto a `MeshService` is visible to anyone who can reach the Envoy admin API. This is already true of endpoint labels, and the existing `AllowedLabelKeys` allowlist is the control point for operators who care.
* A user could set a label such as `kuma.io/display-name` and shadow a control-plane-computed tag, so reserved `kuma.io/`-prefixed tags must be written last and win over user labels. Reserved-key rejection already exists in the label-propagation config, so this is consistent with what we do elsewhere.

## Reliability implications

* **No existing match can break.** The affected listeners have empty `io.kuma.tags` today, so every non-empty selector already fails against them. Adding tags is strictly monotonic.
* **New matches are possible, and that is intended** — but it is a behaviour change and needs a changelog entry. A policy matching `listenerTags: {kuma.io/zone: east}` with no `origin` constraint will start matching listeners it previously missed. The blast radius is policies that match `kuma.io/`-prefixed keys, do not constrain `origin`, and run on an affected proxy.
* **No effect on traffic.** `io.kuma.tags` is listener-only and read by nothing but Kuma's matcher. Clusters and endpoints use separate keys, untouched here, so there is no path to load-balancing, mTLS, or endpoint selection.
* **KDS-synced resources.** Labels are not recomputed for non-locally-originated resources, so a `MeshService` synced global-to-zone keeps the labels it arrived with. Tags derived at generation time must tolerate labels computed by another control plane — notably a `MeshMultiZoneService` has no `kuma.io/zone` label, so that key is absent rather than empty.
* **Config churn.** Listener metadata now changes when labels change, triggering an xDS update for every proxy with an outbound to that destination. Labels change rarely and the listener is already regenerated on any `MeshService` change, so the delta is bounded.
* **Golden files.** Every affected listener golden file gains an `io.kuma.tags` block — a large mechanical diff worth reviewing for unintended label leakage rather than rubber-stamping.

## Implications for Kong Mesh

The downstream project ships policies that patch listeners by tag and inherits both the fix and the mandatory `kuma.io/service` rewrite. Its `MeshProxyPatch`-equivalent policies and bundled defaults need an audit before this ships — they are already silently broken wherever `MeshService` or `InboundTagsDisabled` is enabled. No API or CRD change is required there, since none is required here, and this decision adds no computed label, so the mirrored label registry stays in sync untouched.

## Decision

Listener `io.kuma.tags` is sourced from labels wherever tags no longer exist.

1. **Outbound listeners backed by a real resource** (`MeshService`, `MeshMultiZoneService`, `MeshExternalService`) take the **destination resource's labels**.
2. **Inbound listeners whose tags have been emptied** by `InboundTagsDisabled` take the **proxy's own `Dataplane` labels**.

In both cases `kuma.io/mesh` is stripped, matching what the outbound path already does. Reserved `kuma.io/`-prefixed tags are computed by the control plane and take precedence over user labels. Legacy paths — a `Dataplane` with real outbound tags, or inbounds with tags enabled — keep their current behaviour untouched.

No `MeshProxyPatch` or `ProxyTemplate` API change is made; both read `io.kuma.tags` through the same helper and are fixed by the same change. The change is unconditional and not gated by a new flag.

Policies matching `kuma.io/service: <name>` must be rewritten against labels, typically `kuma.io/display-name: <name>` plus `k8s.kuma.io/namespace: <ns>`. This is unavoidable rather than a choice we are making: `kuma.io/service` does not exist once `MeshService` is in use. We do not synthesize a replacement, because every candidate value is a guess that fails at runtime instead of at review time.

## Notes

* Open topic: per-port targeting. Labels are per-resource and listeners are per-port, so the listeners of a multi-port destination are indistinguishable by `listenerTags`; only `match.name` under unified naming separates them today. Left open rather than guessed at.
* Open topic: whether the empty `io.kuma.tags: {}` written for tagless listeners should become an omitted key. A latent readability wart, but changing it churns golden files repo-wide and is not needed here.
* Follow-up: `MeshProxyPatch` cannot warn when a selector matches nothing, which is why this shipped as a silent regression in the first place. A validation or inspect-time "this policy matched zero resources" signal would have caught it at apply time. `origin` is likewise an unvalidated free-form string, so a typo also silently matches nothing.
* Follow-up: endpoint `Locality` still reads tags only, with no label fallback, so zone-based locality is silently lost under `InboundTagsDisabled`. Same disease, different component, out of scope here.
* Follow-up: `ServiceInsight`s are dropped entirely for tagless dataplanes rather than falling back to labels.
* `MeshGateway` listeners carry no `io.kuma.tags` at all, so `listenerTags` can never match one. Unchanged by this decision, but worth documenting.

## Deep dive

Everything below is supporting detail for the narrative above. It is the evidence, not the decision.

### The `io.kuma.tags` channel

`TagsKey = "io.kuma.tags"` is defined at `pkg/xds/envoy/metadata/v3/metadata.go:99`. It has exactly **one writer** — `TagsMetadataConfigurer` (`pkg/xds/envoy/listeners/v3/tags_metadata.go:15-27`):

```go
l.Metadata.FilterMetadata[envoy_metadata.TagsKey] = &structpb.Struct{
	Fields: envoy_metadata.MetadataFields(c.Tags),
}
```

It has **six readers**, all Kuma's own matching code: `MeshProxyPatch` (`listener_mod.go:86`, `network_filter_mod.go:136`, `http_filter_mod.go:159`) and the legacy `ProxyTemplate` stack (`pkg/xds/generator/modifications/v3/{listener,network_filter,http_filter}.go`). All six do:

```go
listenerTags := envoy_metadata.ExtractTags(listenerProto.Metadata)
if !mesh_proto.TagSelector(pointer.Deref(l.Match.Tags)).Matches(listenerTags) {
	return false
}
```

`TagSelector.Matches` (`api/mesh/v1alpha1/dataplane_helpers.go:465-479`) is an AND of exact key/value pairs with `*` as a match-any-value wildcard; an **absent** key fails the match.

The configurer has **no empty guard** — unlike `EndpointMetadata`, which returns `nil` for empty tags, it writes the key unconditionally. Empty tags therefore produce `io.kuma.tags: {}` (an empty struct), not an absent key. That is why the failure is silent: the bag exists and is empty, so every non-empty selector fails on its first lookup.

Blast radius is bounded: clusters never carry `io.kuma.tags` (`ClusterMatch` has only `origin` and `name`), and endpoint/LB selection uses `envoy.lb`, `io.kuma.labels`, and transport-socket matches — all written by `EndpointMetadata`, untouched by this decision.

Only five call sites stamp listener tags: `inbound_proxy_generator.go:96`, `meshtls/plugin.go:376` (inbound), `outbound_proxy_generator.go:120` (legacy outbound), and `meshtcproute/listeners.go:52` / `meshhttproute/listeners.go:78` (route-plugin outbound). There is no `TagsMetadata` call under `pkg/xds/generator/gateway/`.

### Case 1 — provenance of `kuma.io/service` outbound tags

The chain, end to end:

**destination `Dataplane` inbound `kuma.io/service` tag → VIP outbound entry `TagSet` → consumer `Dataplane`'s `networking.outbound[].tags` → listener `io.kuma.tags`**

The VIPs allocator walks every other `Dataplane`'s inbounds (`pkg/dns/vips_allocator.go:360-361`) and hard-codes a single-key tag set (`:497-501`):

```go
func addDefault(outboundSet *vips.VirtualOutboundMeshView, service string, port uint32) error {
	return outboundSet.Add(vips.NewServiceEntry(service), vips.OutboundEntry{
		TagSet: map[string]string{mesh_proto.ServiceTag: service},
		Origin: string(metadata.OriginService),
		Port:   port,
	})
}
```

The control plane then synthesizes the consumer's outbounds from that view (`pkg/xds/topology/dns.go:36-41`), and the generators copy the bag onto the listener, stripping `kuma.io/mesh`:

```go
// pkg/xds/generator/outbound_proxy_generator.go:120
Configure(envoy_listeners.TagsMetadata(envoy_tags.Tags(outbound.GetTags()).WithoutTags(mesh_proto.MeshTag)))

// pkg/plugins/policies/meshtcproute/plugin/v1alpha1/listeners.go:41
tags := envoy_tags.Tags(svc.Outbound.TagsOrNil()).WithoutTags(mesh_proto.MeshTag)
```

Richer multi-tag outbounds exist only in hand-written Universal `Dataplane`s and in `VirtualOutbound`/cross-mesh `MeshGateway` entries (`vips_allocator.go:262`, `:489`).

Inbound listeners are fed directly from the proxy's own `Dataplane` (`inbound_proxy_generator.go:96`, `TagsMetadata(iface.GetTags())`), where on Kubernetes `InboundTagsForService` (`inbound_converter.go:241-271`) builds them from Pod labels (excluding `kuma.io/` keys) plus namespace, service, port, zone, and node labels. `kuma.io/mesh` is not stripped on this path.

### Case 2 — what `MeshService` changes

An outbound becomes a reference (`pkg/core/xds/types/outbound.go:8-15`):

```go
type Outbound struct {
	// LegacyOutbound is an old way to define outbounds using 'kuma.io/service' tag
	LegacyOutbound *mesh_proto.Dataplane_Networking_Outbound

	Address  string
	Port     uint32
	Resource kri.Identifier
}
```

`LegacyOutbound != nil` discriminates the old world; a non-empty `Resource` the new one. Tags exist only in the old one (`:43-49`):

```go
func (o *Outbound) TagsOrNil() map[string]string {
	if o.LegacyOutbound != nil {
		return o.LegacyOutbound.Tags
	}
	return nil
}
```

`WithoutTags(nil)` returns nil, `TagsMetadata(nil)` is still called, and the unguarded configurer writes `io.kuma.tags: {}`.

The claim that `kuma.io/service` is *removed*, link by link: the consumer's outbound tags are gone (`outbound.go:8-15`); the VIP relay no longer feeds outbounds, since destinations are addressed by KRI; and the destination's inbounds carry no tag either under `meshServices: Exclusive` with `InboundTagsDisabled` (`inbound_converter.go:52-57`, path selected at `pod_converter.go:325`).

`match.name` fallback (`meshtcproute/plugin/v1alpha1/listeners.go:30-40`):

```go
listenerName := envoy_names.GetOutboundListenerName(address, port)  // "outbound:<address>:<port>"
if id, ok := svc.Outbound.AssociatedServiceResource(); ok && unifiedNaming {
	listenerName = id.String()  // KRI — only when unified naming is enabled
}
```

### Case 3 — what `InboundTagsDisabled` changes

One choke point (`pkg/plugins/runtime/k8s/controllers/inbound_converter.go:52-57`):

```go
func (ic *InboundConverter) tagsOrEmpty(tagsFn func() map[string]string) map[string]string {
	if ic.InboundTagsDisabled {
		return map[string]string{}
	}
	return tagsFn()
}
```

Two call sites (`:74` for services, `:116` for serviceless) wrap the tag builders in a closure that is never evaluated when disabled. Inbounds lose `kuma.io/service`, `kuma.io/zone`, `k8s.kuma.io/namespace`, and all copied Pod/node labels at once.

The data is not lost, because Pod labels are copied onto the `Dataplane`'s metadata labels **unconditionally** (`pod_converter.go:62-84`):

```go
labels, err := resource_labels.Compute(
	core_mesh.DataplaneResourceTypeDescriptor,
	currentSpec,
	mergeLabels(dataplane.GetLabels(), pod.Labels),   // pod labels → Dataplane meta labels
	...
	resource_labels.WithWorkload(workloadName),        // kuma.io/workload
)
dataplane.SetLabels(labels)
```

`mergeLabels` (`:455-464`) is a plain clone-and-copy. The flag adds no label mechanism; it removes the duplicate copy held in tags.

### The tags-to-labels substitution inventory

**Real fallback to labels — four sites:**

* `pkg/core/resources/apis/mesh/dataplane_helpers.go:193-204` — `IdentifyingName` returns the `kuma.io/workload` label when tags are disabled, else the `kuma.io/service` tag. Consumers: `mads/v1/generator/assignments.go:50`, `meshtrace/plugin.go:266`, `meshaccesslog/plugin.go:108`, `meshmetric/plugin.go:146`.
* `pkg/core/resources/apis/mesh/dataplane_helpers.go:181-190` — `InboundIdentifyingName` returns the `Dataplane` KRI with the port as section name.
* `pkg/plugins/runtime/k8s/controllers/meshservice_controller.go:374-385` — selector swaps `DataplaneTags` for `DataplaneLabels`, using the same Service selector keys.
* `meshloadbalancingstrategy` — `priority.go:123-137` (`resolveAffinityValues`) and `locality_aware.go:151-154`, preferring inbound tags and falling back to Pod labels, pre-filtered to affinity keys so unrelated labels do not leak.

**No fallback — data simply dropped or relaxed:**

* `pkg/insights/resyncer.go:443`, `:668` — `ServiceInsight` entries skipped for tagless dataplanes.
* `pkg/core/resources/apis/mesh/dataplane_validator.go:105-111` — the `kuma.io/service` requirement becomes a no-op.
* `pkg/xds/topology/outbound.go:386`, `:422` — `Locality` still derives from `getZone(inboundTags)`, tags only.
* `pkg/core/xds/types.go:138`, `:422`, `:435` — `Protocol()`, `ContainsTags`, selector matching: tags only.
* `pkg/xds/generator/inbound_proxy_generator.go:96` — the inbound listener's tags, the gap this MADR closes.

`io.kuma.labels` (`metadata.go:104`) is worth a caveat: despite the generic name it is a closed three-site loop — written at `endpoints/v3/endpoints.go:41`, encoded/decoded in `metadata.go:50`/`:71`, and read by exactly **one** consumer, `meshloadbalancingstrategy/locality_aware.go:64`. It is not a general identity channel.

### Where the labels come from

The destination resource is already resolved at generation time and then discarded (`pkg/plugins/policies/core/xds/meshroute/listeners.go:147-167`):

```go
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

`GetServiceByKRI` returns a `core.Destination`, which embeds `core_model.Resource` and exposes `GetMeta().GetLabels()`. `DestinationService` (`:68-72`) keeps only the KRI and a synthesized name, so the labels are thrown away — threading them through needs no new lookup.

A `MeshService` carries identity in labels, not its spec. The computed set is centralised in `pkg/core/resources/labels/registry.go:10-23`; the Universal generator (`meshservice/generate/generator.go:272-284`) sets `kuma.io/mesh`, `kuma.io/display-name`, `kuma.io/managed-by`, `kuma.io/env`, `kuma.io/zone`, and `kuma.io/origin`, plus user labels when `KUMA_MESH_SERVICE_LABEL_PROPAGATION_ENABLED` is on. The Kubernetes controller adds `k8s.kuma.io/namespace` and `k8s.kuma.io/is-headless-service` (`meshservice_controller.go:373-489`). Note there is no `kuma.io/namespace` label — it is `k8s.kuma.io/namespace` (`mesh_proto.KubeNamespaceTag`).

A resulting outbound listener for `MeshService` `backend` in namespace `kuma-demo`, zone `east`:

```yaml
metadata:
  filterMetadata:
    io.kuma.tags:
      kuma.io/display-name: backend
      k8s.kuma.io/namespace: kuma-demo
      kuma.io/zone: east
      kuma.io/env: kubernetes
      kuma.io/origin: zone
      kuma.io/managed-by: k8s-controller
      team: payments          # user label, when label propagation is enabled
```

### Implementation notes

* `DestinationService` (`meshroute/listeners.go:68-72`) must carry the resolved `core.Destination` (or its labels) through from `CollectServices`.
* Both outbound generators share the `TagsOrNil().WithoutTags(MeshTag)` line and must both be updated: `meshtcproute/plugin/v1alpha1/listeners.go:41` and `meshhttproute/plugin/v1alpha1/listeners.go:78`.
* The inbound half is `inbound_proxy_generator.go:96`, which needs the `Dataplane`'s labels when `iface.GetTags()` is empty. `pkg/xds/context/context.go:40` already carries `InboundTagsDisabled` into the xDS context.
* Legacy branches — `LegacyOutbound != nil`, and inbounds with tags enabled — keep their current behaviour unchanged.
