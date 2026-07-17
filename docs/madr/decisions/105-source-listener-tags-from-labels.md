# Source Envoy listener tags from resource labels

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/17381

## Context and Problem Statement

`MeshProxyPatch` lets users pick a generated Envoy listener and patch it. You can pick a listener by `origin`, by `name`, or by `tags`/`listenerTags`. The last one matches against a set of key/value pairs that Kuma puts on every listener as Envoy filter metadata, under the key `io.kuma.tags`.

On an inbound listener those pairs come from the `Dataplane`'s inbound tags. On an outbound listener there is only one pair, `kuma.io/service`, and the control plane hard-codes it from the destination service. Kuma is moving identity out of tags and into labels, but the listener metadata still reads tags. So tag matching stops working, and it stops working quietly: the policy is accepted, it validates, and it matches nothing.

### User cases

#### `MeshProxyPatch` that matches `listenerTags: {kuma.io/service: backend}` stops applying once `MeshService` is on. 

The user needs something else to match on.

#### Patch the outbound listener to `MeshService` `backend` in namespace `kuma-demo`, without turning on unified resource naming.

The user may want to enable unified naming separately, and without tags there is no way to apply the patch.

#### Patch a proxy's own inbound listener when `InboundTagsDisabled` has emptied its tags.

The user needs something else to match on.

#### The same thing should work for `MeshService`, `MeshMultiZoneService` and `MeshExternalService`, which share one outbound code path.

### Current state

Listener tags are Envoy metadata under `io.kuma.tags`. One configurer writes them, at a few inbound and outbound spots, and what lands in that map depends on the setup. Three setups matter. The call sites and the configurer are in the deep dive.

Identity metadata lives in one more place: endpoints. Each endpoint carries the inbound tags under `envoy.lb` and the `Dataplane` labels under `io.kuma.labels`, and `MeshLoadBalancingStrategy` reads them, tags first and labels as a fallback once the flag has stripped the tags.

| Case                              | Inbound listener tags, what gets written | Outbound listener tags, what gets written            | Endpoint metadata, what gets written                         |
|:----------------------------------|:-----------------------------------------|:-----------------------------------------------------|:-------------------------------------------------------------|
| **1. `kuma.io/service`** (legacy) | the proxy's own `Dataplane` inbound tags | single `kuma.io/service`, hard-coded by the VIP allocator | `envoy.lb` holds inbound tags, `io.kuma.labels` holds `Dataplane` labels |
| **2. `MeshService` enabled**      | the proxy's own `Dataplane` inbound tags | nothing, the reference carries no tags so the map is `{}` | same as case 1 |
| **3. `InboundTagsDisabled`**      | nothing, the flag zeroes them so the map is `{}` | unchanged by this flag                        | `envoy.lb` is now empty, `io.kuma.labels` still holds `Dataplane` labels |

Cases 2 and 3 are two separate switches. Turn both on and both listener directions write an empty map. Endpoints keep working through it: they still carry the `Dataplane` labels under `io.kuma.labels`, which is why `MeshLoadBalancingStrategy` does not break when the tags go.

**Case 1, `kuma.io/service`.** The destination is a string, copied from proxy to proxy. It starts as the `kuma.io/service` tag on the destination's inbound, goes into the VIP view, then into the consumer `Dataplane`'s outbound tags, and ends up on the outbound listener. Two things to note. The tags describe the destination but live on the consumer. And on the generated path, which is all of Kubernetes and every transparent-proxy setup, there is only ever one key, because the control plane hard-codes it. Outbounds with more tags, like `version` or `region`, only came from hand-written Universal `Dataplane`s.

**Case 2, `MeshService`.** The destination becomes a real resource, addressed by KRI. There is no tag set to copy, so the outbound listener gets an empty one. `kuma.io/service` is gone from the model now: the consumer has no outbound tags, and the VIP path no longer feeds outbounds. Nothing is left to read, so `listenerTags: {kuma.io/service: backend}` cannot work, whatever we do. Those policies have to be rewritten. The only open question is what we rewrite them to.

**Case 3, `InboundTagsDisabled`.** The flag strips inbound tags, so inbound listeners break the same way, for the opposite reason. The flag saves memory: on Kubernetes the tags mostly repeat what the labels already hold, once on every inbound of every proxy. So the tags are not coming back, and we have to read labels instead. For Pod labels the data is still there, because they are copied onto the `Dataplane`'s labels either way, and the flag only drops the copy that lived in tags. Node labels are the one exception, covered in the deep dive: they only ever lived in the tags, so the label fallback cannot bring them back.

Cases 2 and 3 are part of a migration most of the codebase has already made. Identity is read from labels now, and listener tags are one of the last places still on tags. The move was done piece by piece, and listener metadata got skipped. It is also the piece that breaks a user-facing policy API.

We have fixed this same bug once before, on another policy. `MeshLoadBalancingStrategy` affinity tags broke under case 3 in the same quiet way, and the fix was to read Pod labels instead, sending them to Envoy on the endpoints under `io.kuma.labels`. That work covered the endpoint path and stopped there. The deep dive has the rest, including why that fix needed a second key and this one does not.

What reads the listener map is our own policy matching. `MeshProxyPatch`'s `listenerTags`, the case this MADR is about, matches `io.kuma.tags`, so an empty map matches nothing. The deprecated `ProxyTemplate` `modifications` path reads the same key. Endpoint metadata is a separate channel: only `MeshLoadBalancingStrategy` reads it, and `MeshProxyPatch` does not, so the two never overlap.

## Design

Every option answers the same question: when a listener has no tags, what identifies the thing it stands for? An outbound listener stands for the destination service, which is the `Mesh*Service` resource it was generated from. An inbound listener stands for the proxy itself. For both of those, identity lives in labels now.

An outbound listener does not stand for the individual workloads behind the destination. Those are its endpoints. They are picked by the service, and their own labels already go to Envoy separately, on the endpoints, under `io.kuma.labels`. So the outbound listener wants the service's labels, and the inbound listener wants the `Dataplane`'s. Clusters are not part of this. They never carry `io.kuma.tags`, so there is nothing on them to fill.

### Option 1: Do nothing, require unified resource naming and `match.name`

* Good, because it needs no code change, and a KRI-based name is exact and can tell individual ports apart.
* Bad, because User case 1 stays broken and silent, so users find out in production.
* Bad, because without unified naming the listener is called `outbound:<address>:<port>`, and that address is a VIP we allocate at runtime. You cannot know it up front and it changes across restarts, so there is no way to write the policy.
* Bad, because getting a stable name means turning on unified naming across the whole mesh (`meshServices: Exclusive` plus a DP feature flag). That is a lot to ask of someone who wants to patch one listener.
* Bad, because a name matches a single listener, so User case 3 needs one policy per destination and has to be regenerated whenever the set changes.
* Bad, because it does nothing for User case 4. `InboundTagsDisabled` has no alternative based on names.
* Bad, because every affected listener keeps an empty `io.kuma.tags`, which looks like a listener with no tags.

### Option 2: Build a `kuma.io/service` tag

Rebuild a `kuma.io/service` value from the destination resource and keep writing it, so old policies look like they still work.

* Good, because it is a small change and nobody has to rewrite a policy. It is smaller than it looks: the value is already computed and carried as `DestinationService.KumaServiceTagValue` (`meshroute/listeners.go:167`), so we would only have to write it.
* Bad, because `kuma.io/service` has been deliberately removed, and bringing it back on the listener works against the whole point of `MeshService`.
* Bad, because the value would not match anyway, and we cannot make one that does. What we can build is the legacy resolver's `<mesh>_<name>_<namespace>_<zone>_<shortName>_<port>`, and the policy says `backend`. We could invent a shorter string, but `backend` on its own does not say which namespace or which of the three destination kinds is meant, so any string we invent is a guess at what the user meant and they would find out at runtime.

### Option 3: A separate `io.kuma.labels` key on listeners, with new match fields

Write labels under a new listener key and add `listenerLabels`/`labels` to the `MeshProxyPatch` API. We already have an `io.kuma.labels` key on endpoints, added for this exact problem, so the obvious move is to copy it onto listeners.

* Good, because it keeps labels and tags cleanly apart.
* Good, because it looks consistent with what endpoints already do, and consistency is worth something.
* Bad, because the endpoint split exists for a reason that doesn't apply here, and not the reason it looks like. On endpoints, tags live under `envoy.lb` and `envoy.transport_socket_match`, which drive LB subset matching and TLS selection. Both match against a fixed set of declared keys rather than the whole metadata map, so Envoy tolerates extra keys and arbitrary labels are inert unless a label key collides with a declared split or TLS key. The real reason for the second key is that labels have to survive when `InboundTagsDisabled` strips the tags out of `envoy.lb`, and a separate key keeps that collision class away too; the deep dive works through it. `io.kuma.tags` is read by nothing except our own matchers and is filled only when tags are already gone, so neither constraint applies. The split on endpoints was a workaround for a problem that listeners do not have.
* Bad, because `io.kuma.labels` is an internal hand-off rather than a selector surface. It exists so `MeshLoadBalancingStrategy` can read labels back out of endpoints another generator already built, it has one consumer, it is undocumented for users, and nobody ever writes it in a policy. Reusing the name on listeners would give it a second, unrelated meaning.
* Bad, because it changes a public API across three match types, and means regenerating CRDs and OpenAPI.
* Bad, because users then have to know which of the two selectors applies to a given listener, and the answer depends on whether the destination is legacy or a real resource. That is an internal detail they should not have to care about.
* Bad, because the matcher and the behaviour come out the same as Option 4, and only the key name differs. We would do a full API migration only to get a better key name.
* Bad, because we have to backport this fix, and we don't want to ship new API fields and CRD schemas in a patch release.
* Bad, because `MeshProxyPatch` is not the only reader of `io.kuma.tags`. The deprecated `ProxyTemplate` `modifications` path matches on the same key and would not get the new fields, so the same listener would be selectable one way and not the other.

### Option 4: Fill `io.kuma.tags` from labels when tags are gone

Keep one key and one selector. When a listener has no tags, take them from the labels of whatever the listener stands for: the destination resource for an outbound, the proxy's own `Dataplane` for an inbound. Strip `kuma.io/mesh`, which the outbound path already does today.

The labels are there in both environments, which is what makes this work. A `MeshService` has no tags to lose in the first place, so its computed labels always identify it: `kuma.io/display-name`, `kuma.io/zone`, `kuma.io/env`, `kuma.io/origin` plus `k8s.kuma.io/namespace` on Kubernetes. A `Dataplane` has them too, from Pod labels on Kubernetes and from label computation on Universal. Which of those keys we actually copy is the one thing still open, and the deep dive works through the candidates.

This reuses the substitution Kuma has already applied elsewhere and brings the last component in line with it.

A policy that used to say `kuma.io/service: backend` becomes:

```yaml
match:
  origin: outbound
  listenerTags:
    kuma.io/display-name: backend
    k8s.kuma.io/namespace: kuma-demo
```

And User case 3, which no name-based selector can do at all, becomes:

```yaml
match:
  origin: outbound
  listenerTags:
    team: payments
```

* Good, because there is no API change. The selector people already use starts working again, and it is safe to ship as a patch.
* Good, because it can only add matches. The affected listeners have empty tags today, so nothing that matches now can stop matching.
* Good, because labels give you more to match on than the single `kuma.io/service` string ever did. Zone, origin, env and user labels all become usable, which is what makes User case 3 possible.
* Good, because one rule handles both directions, so User case 4 falls out of the same change.
* Bad, because the key is called `tags` and now holds labels. That is confusing, though `io.kuma.tags` was never a user-facing "Dataplane tags" surface. It is our own listener selector map, written in one place and read only by our policy matching.
* Bad, because labels belong to a resource while listeners belong to a port. The two listeners of a two-port `MeshService` get identical tags and `listenerTags` cannot tell them apart. You still need `match.name` with unified naming for that, and this MADR does not fix it.
* Bad, because labels don't say what kind of resource a listener stands for. A `MeshMultiZoneService` groups the `MeshService`s of one service across zones, and it is normal to give it the same name as the `MeshService`s it groups. Its labels are then a subset of theirs, and a selector cannot say "this key must be absent". So a policy that matches the group also matches the single service behind it, and the group cannot be patched on its own. Sub-option B below addresses this.

Option 4 says where the values come from. It does not say what we write. There are three ways to fill the tag set, and A and B can also be taken together.

### What we write into the tags

| Direction    | Sub-option A, the resource's labels   | Sub-option B, a synthesized `kuma.io/kri`                   | Sub-option C, KRI outbound, labels inbound          |
|:-------------|:--------------------------------------|:------------------------------------------------------------|:-----------------------------------------------------|
| **Outbound** | labels of the destination `Mesh*Service` | KRI of `outbound.Resource`, section name is the port      | KRI of `outbound.Resource`, section name is the port |
| **Inbound**  | labels of the proxy's own `Dataplane` | KRI of that `Dataplane`, section name is the inbound's port | labels of the proxy's own `Dataplane`                |

#### Sub-option A: copy the resource's labels

The outbound listener takes the labels of the `MeshService`, `MeshMultiZoneService` or `MeshExternalService` it was generated from; the inbound listener takes the labels of the proxy's own `Dataplane`. This is the reading of Option 4 everything above assumes.

```yaml
match:
  origin: outbound
  listenerTags:
    kuma.io/display-name: backend
    k8s.kuma.io/namespace: kuma-demo
```

* Good, because the keys are the ones people already read in `kubectl get` output and already write in `targetRef.labels`, so nothing new has to be learned.
* Good, because it is the only one of the three that does User case 3. A selector can match a set, so `team: payments` or `kuma.io/zone: east` patches a group without naming its members.
* Good, because it degrades gracefully. A missing key just doesn't match, and every listener still carries something.
* Bad, because it cannot express kind or port, which is what the two collisions below are about.
* Bad, because we copy whatever labels the resource happens to have. On Kubernetes that is every label on the `Service`. Which keys we should copy is left open, in the implementation notes.
* Bad, because of potential listener/cluster change on label update if we not use fixed map.

#### Sub-option B: synthesize a `kuma.io/kri` tag

Add one key Kuma computes rather than copies. A KRI names a resource exactly. It can also name one part of that resource, a single port, in a field called the section name. That is what lets a KRI tell the two listeners of a two-port service apart, which labels cannot.

For an outbound, we already have it: `outbound.Resource` holds the destination's KRI with the port as its section name (`core/xds/types/outbound.go:8-24`). For an inbound we would build it, with one line we already have elsewhere: `kri.WithSectionName(kri.FromResourceMeta(dp.GetMeta(), DataplaneType), portName)`. `InboundIdentifyingName` (`dataplane_helpers.go:179-189`) does exactly this, but it is gated on `InboundTagsDisabled` and falls back to a plain service name, so we would call the two `kri` helpers directly rather than reuse it.

```yaml
match:
  origin: outbound
  listenerTags:
    kuma.io/kri: kri_msvc_default_east_kuma-demo_backend_http
```

The format is `kri_<shortName>_<mesh>_<zone>_<namespace>_<name>_<sectionName>`, always seven fields, with absent parts left as empty segments (`kri.go:28-39`). The short names are `msvc`, `mzsvc` and `extsvc`; the section name is the port, by name or by number when unnamed (`meshservice/api/v1alpha1/helpers.go:89`, `:118-120`). The same `MeshMultiZoneService` from the collision example is `kri_mzsvc_default__kuma-system_backend_http`: it is created on global, so it has no zone, and the doubled underscore is that empty field rather than a typo.

Note this is a tag we synthesize, not a label we store. A `kuma.io/kri` label would be a derived copy of labels the resource already carries, since `kri.FromResourceMeta` (`kri.go:86-92`) builds the identifier out of `kuma.io/zone`, `k8s.kuma.io/namespace` and the display name. It would need recomputing and syncing, and it still could not carry a section name, because a label belongs to a resource while a listener belongs to a port.

* Good, because it is exact and unambiguous. `ResourceType` is part of the KRI (`kri.go:38`), so it separates the destination kinds that labels cannot.
* Good, because it answers per-port targeting. The two listeners of a two-port `MeshService` finally differ.
* Good, because it is one bounded, stable key that moves only when identity moves, so it does not add to the churn described in the reliability section.
* Good, because it needs no API change and backports cleanly: `kri.Identifier` and its format are identical on release-2.13 and release-2.14.
* Bad, because the value is opaque and positional, and a typo matches nothing silently, which is the failure mode this MADR exists to fix.
* Bad, because it overlaps `match.name` wherever unified naming is on, since the listener name is already the KRI there.
* Bad, because legacy outbounds have no KRI, so the key is absent in case 1 and users have to know which world they are in.

#### Sub-option C: KRI on the outbound, labels on the inbound

Treat the two directions differently, because the two problems are different. Only the outbound has the collision problem: three kinds of destination share one code path and can share a name. An inbound listener stands for the proxy's own `Dataplane`, and there is only one of those and only one kind, so nothing can collide with it. Give the outbound the exact key and leave the inbound with the labels it can already carry.

* Good, because it puts exactness where the ambiguity actually is, and readable keys where they are enough.
* Good, because inbound tags used to hold Pod labels, so a user matching `version: v1` on their own inbound keeps writing what they wrote before.
* Bad, because the two directions answer the same selector differently, so users have to remember which side they are on.
* Bad, because the inbound keeps the per-port ambiguity that the outbound just lost, so a two-port proxy still cannot patch one of its own inbounds.


## Security implications and review

Low for the fact that a destination appears at all. The outbound exists because the destination is reachable, and its VIP and port are already in the config dump, so we reveal no new reachability.

## Reliability implications

* Nothing that matches today can break, because the affected listeners have empty `io.kuma.tags` and every non-empty selector already fails against them.
* Traffic is unaffected. `io.kuma.tags` only exists on listeners and only our matcher reads it. Clusters and endpoints use other keys that we don't touch, so there is no path from here to load balancing, mTLS or endpoint selection.
* Labels are not recomputed for resources that came from elsewhere, so a `MeshService` synced from global keeps the labels it arrived with. Tags built at generation time have to cope with labels computed by another control plane. A `MeshMultiZoneService`, for instance, has no `kuma.io/zone` label at all, so that key just won't be there.
* Listener metadata now changes when labels change, so editing a label on a `MeshService` updates the outbound listener of every proxy that talks to it. The trigger is not new. We hash a `MeshService` by its version, so any write to it already regenerates every proxy in the mesh. What is new is that the regeneration now produces different bytes. Today that output is thrown away, because `autoVersion` hashes each resource type and compares it against the previous snapshot, and the reconciler returns early when nothing changed (`pkg/xds/server/v3/reconcile.go:69-91`). Writes that leave labels alone, status updates in particular, stay free for the same reason.
* Copying labels wholesale would tie listener churn to how often people deploy. Keys like `app.kubernetes.io/version` and `argocd.argoproj.io/instance` change on every rollout, and each change would replace a listener on every proxy reaching that service, for labels nobody is matching on. A fixed key list removes this, since only keys we picked can move the bytes.
* Memory on the proxy is a small addition to something larger that already ships. Outbound listeners are one per destination and port, while endpoints are one per pod, and every local endpoint already carries the full unfiltered `Dataplane` label set (`pkg/xds/topology/outbound.go:384`, `:421`). A typical `MeshService` carries around 500 bytes of keys and values, closer to a kilobyte once it is a `google.protobuf.Struct`. A proxy reaching 200 services would hold roughly 200 KB of listener metadata against the few megabytes of endpoint labels it already holds. The control plane keeps a snapshot per proxy, so it scales the same way and the same ratio applies there.
* The typical case is not the worst one. Nothing caps how many labels a `Service` can have, in Kuma or in Kubernetes, so a heavily labelled `Service` reaches a few kilobytes per listener and there is no upper bound. This is another reason to bound the key set rather than to rely on the ratio above staying true.
* This adds to a config we are otherwise trying to shrink. `InboundTagsDisabled` exists to cut memory, and #11065 wants to strip tags from endpoints for the same reason. The addition here is modest and lands on listeners rather than endpoints, but it adds rather than removes, so we should write no more metadata than the feature needs.
* Every affected golden file already has an `io.kuma.tags` block, because we write the key even when it is empty. What changes is that the empty ones gain content. It is a big mechanical diff, and it is worth actually reading it to catch labels we didn't mean to expose.

## Backport

This is a silent regression in released versions, so we backport the fix to release-2.13 and release-2.14. That shapes the design, and it is the main reason we pick Option 4 over Option 3. Option 4 only changes how we fill generated metadata, so there is no API, no CRD schema and no stored resource to change in a patch. And since it can only add matches, it can't break a policy that works today.

The two halves don't backport the same way, because the features they react to landed at different times:

* The outbound half goes to 2.13 and 2.14. Everything it needs (`Outbound.TagsOrNil`, `DestinationService`, `GetServiceByKRI`) is on both.
* The inbound half only goes to 2.14 and master. `InboundTagsDisabled` doesn't exist on 2.13, so inbound tags are always populated there and there is nothing to fix.

One practical note for whoever does the cherry-pick: label computation moved packages between 2.13 and 2.14, from `core_model.ComputeLabels` in `pkg/core/resources/model/resource.go` to `pkg/core/resources/labels`. The 2.13 pick will conflict there. The computed labels themselves are the same, so it is import paths and call shape to fix up, and the behaviour stays put.

Because the tags a listener carries depend on which labels that version computes, a policy written for 2.14 may not match on 2.13. That comes with backporting a fix whose payload is "whatever labels this version has", and it belongs in the release notes so people don't hit it cluster by cluster.

## Implications for Kong Mesh

None

## Decision

Listener `io.kuma.tags` is filled from labels wherever tags no longer exist.

1. Outbound listeners will get `kuma.io/kri` label with value of the destination `Mesh*Service` to avoid more frequent configuration changes and potential higher memory usage.
2. Inbound listeners whose tags are empty take the labels of the proxy's own `Dataplane`. `InboundTagsDisabled` is how that happens on Kubernetes, but the rule is keyed on the tags being empty, not on the flag, because a hand-written Universal `Dataplane` can have tagless inbounds with the flag off.

Point 2 is worth stating explicitly, because the flag's name suggests the opposite. Turning off inbound tags should not turn off listener tags. The flag decides what goes on the `Dataplane`'s inbounds, and the listener still has to say what it is. An inbound listener always carries a filled `io.kuma.tags`: from tags when they exist, from the `Dataplane`'s labels when they don't. We never leave it empty.

Both cases strip `kuma.io/mesh`, which is what the outbound path already does. Reserved `kuma.io/` tags come from the control plane and win over user labels. Legacy paths keep working as they do today, meaning a `Dataplane` with real outbound tags, or inbounds with tags enabled.

There is no `MeshProxyPatch` API change. The fix is entirely in how we fill the listener metadata. It is unconditional and has no new flag.

Policies matching `kuma.io/service: <name>` have to be rewritten against labels, usually `kuma.io/kri: <kri>`.


## Deep dive

Everything below is the supporting detail for the sections above.

### The `io.kuma.tags` channel

`TagsKey = "io.kuma.tags"` is defined at `pkg/xds/envoy/metadata/v3/metadata.go:99`. One thing writes it, `TagsMetadataConfigurer` (`pkg/xds/envoy/listeners/v3/tags_metadata.go:15-27`):

```go
l.Metadata.FilterMetadata[envoy_metadata.TagsKey] = &structpb.Struct{
	Fields: envoy_metadata.MetadataFields(c.Tags),
}
```

Six places read it, and they are all ours. Three are `MeshProxyPatch`, at `listener_mod.go:86`, `network_filter_mod.go:136` and `http_filter_mod.go:159`. The other three are the deprecated `ProxyTemplate` `modifications` path, which does the same match on the same key: `modifications/v3/listener.go:68`, `network_filter.go:117` and `http_filter.go:140`. Each one does:

```go
listenerTags := envoy_metadata.ExtractTags(listenerProto.Metadata)
if !mesh_proto.TagSelector(pointer.Deref(l.Match.Tags)).Matches(listenerTags) {
	return false
}
```

`TagSelector.Matches` (`api/mesh/v1alpha1/dataplane_helpers.go:465-479`) ANDs exact key/value pairs, with `*` matching any value. If a key is missing, the match fails.

The configurer has no empty guard. `EndpointMetadata` returns `nil` for empty tags, but this one always writes the key. Empty tags therefore give you `io.kuma.tags: {}`, an empty struct rather than a missing key. That is why the failure is quiet: the tag set is there and it is empty, so every non-empty selector fails on its first lookup.

The scope of any change here is small. Clusters never carry `io.kuma.tags` (`ClusterMatch` only has `origin` and `name`), and endpoint and LB selection use `envoy.lb`, `io.kuma.labels` and transport-socket matches, all written by `EndpointMetadata`, which this decision leaves alone.

Five call sites stamp listener tags: `inbound_proxy_generator.go:96` and `meshtls/plugin.go:376` for inbound, `outbound_proxy_generator.go:120` for legacy outbound, and `meshtcproute/listeners.go:52` and `meshhttproute/listeners.go:78` for route-plugin outbound. There is no `TagsMetadata` call under `pkg/xds/generator/gateway/`.

### Case 1, where `kuma.io/service` outbound tags come from

The path, end to end:

destination `Dataplane` inbound `kuma.io/service` tag, then VIP outbound entry `TagSet`, then consumer `Dataplane`'s `networking.outbound[].tags`, then listener `io.kuma.tags`.

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

The control plane then builds the consumer's outbounds from that view (`pkg/xds/topology/dns.go:36-41`), and the generators copy the set onto the listener, stripping `kuma.io/mesh`:

```go
// pkg/xds/generator/outbound_proxy_generator.go:120
Configure(envoy_listeners.TagsMetadata(envoy_tags.Tags(outbound.GetTags()).WithoutTags(mesh_proto.MeshTag)))

// pkg/plugins/policies/meshtcproute/plugin/v1alpha1/listeners.go:41
tags := envoy_tags.Tags(svc.Outbound.TagsOrNil()).WithoutTags(mesh_proto.MeshTag)
```

Outbounds with more tags only exist in hand-written Universal `Dataplane`s and in `VirtualOutbound` and cross-mesh `MeshGateway` entries (`vips_allocator.go:262`, `:489`).

Inbound listeners are fed straight from the proxy's own `Dataplane` (`inbound_proxy_generator.go:96`, `TagsMetadata(iface.GetTags())`). On Kubernetes, `InboundTagsForService` (`inbound_converter.go:241-271`) builds those from Pod labels, skipping `kuma.io/` keys, plus namespace, service, port, zone and node labels. This path does not strip `kuma.io/mesh`.

### Case 2, what `MeshService` changes

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

`LegacyOutbound != nil` means the old world, a non-empty `Resource` means the new one. Only the old one has tags (`:43-49`):

```go
func (o *Outbound) TagsOrNil() map[string]string {
	if o.LegacyOutbound != nil {
		return o.LegacyOutbound.Tags
	}
	return nil
}
```

`WithoutTags(nil)` returns nil, `TagsMetadata(nil)` still runs, and the unguarded configurer writes `io.kuma.tags: {}`.

Why `kuma.io/service` is gone, link by link: the consumer's outbound tags no longer exist (`outbound.go:8-15`); the VIP path no longer feeds outbounds, since destinations are addressed by KRI; and the destination's inbounds carry no tag under `meshServices: Exclusive` with `InboundTagsDisabled` (`inbound_converter.go:52-57`, path chosen at `pod_converter.go:325`).

The `match.name` fallback (`meshtcproute/plugin/v1alpha1/listeners.go:30-40`):

```go
listenerName := envoy_names.GetOutboundListenerName(address, port)  // "outbound:<address>:<port>"
if id, ok := svc.Outbound.AssociatedServiceResource(); ok && unifiedNaming {
	listenerName = id.String()  // KRI — only when unified naming is enabled
}
```

### Case 3, what `InboundTagsDisabled` changes

There is one choke point (`pkg/plugins/runtime/k8s/controllers/inbound_converter.go:52-57`):

```go
func (ic *InboundConverter) tagsOrEmpty(tagsFn func() map[string]string) map[string]string {
	if ic.InboundTagsDisabled {
		return map[string]string{}
	}
	return tagsFn()
}
```

Two call sites use it, `:74` for services and `:116` for serviceless, both wrapping the tag builders in a closure that never runs when the flag is set. Inbounds lose `kuma.io/service`, `kuma.io/zone`, `k8s.kuma.io/namespace` and all the copied Pod and node labels in one go.

The flag is a Kubernetes-side stripper only. `tagsOrEmpty` lives in the pod converter, so it never touches a Universal `Dataplane`, whose inbounds keep whatever tags their author wrote. Universal reaches the same empty state in a different way: `validateNetworking` (`dataplane_validator.go:105-111`) only requires `kuma.io/service` on an inbound that has tags at all, and that rule is not gated on the flag. A hand-written `Dataplane` with tagless inbounds is therefore valid on any control plane, in either environment, with or without the flag, and its inbound listeners get `io.kuma.tags: {}` today. So the fallback keys on "the tags are empty" rather than on the flag, which is also all `inbound_proxy_generator.go:96` can see.

What we fall back to is thinner on Universal. There are no Pod labels to copy, so the `Dataplane`'s metadata labels are whatever `Compute` puts there at write time (`dataplane_manager.go:66`, `:108`): `kuma.io/display-name`, `kuma.io/mesh`, `kuma.io/origin`, `kuma.io/zone`, `kuma.io/env: universal`, `kuma.io/proxy-type`, plus `kuma.io/listener-zoneingress`/`kuma.io/listener-zoneegress` on gateways. Alongside those sit any labels the author wrote under `metadata.labels`, kept verbatim. `kuma.io/workload` is not among them, because `WithWorkload` is only passed by the Kubernetes pod converter (`pod_converter.go:74`). A Universal inbound listener still ends up with a non-empty tag set, and there `kuma.io/display-name` is the `Dataplane`'s own name rather than a service's.

On Kubernetes the data survives, because Pod labels are copied onto the `Dataplane`'s metadata labels either way (`pod_converter.go:62-84`):

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

`mergeLabels` (`:511-518`) is a plain clone and copy. The flag adds no label mechanism, it only drops the copy that lived in tags.

Node labels are the exception the fallback cannot cover. `getNodeLabelsToCopy` (`inbound_converter.go:220-235`) copies an allow-listed set of node labels straight into the inbound tags (`maps.Copy` at `:262`), keyed by `KUMA_RUNTIME_KUBERNETES_INJECTOR_NODE_LABELS_TO_COPY` and defaulting to `topology.kubernetes.io/zone`, `topology.kubernetes.io/region` and `kubernetes.io/hostname` (`config.go:113`, `:254`). Those never reach the `Dataplane`'s metadata labels: `mergeLabels` merges the existing labels with `pod.Labels` only, and the node labels sit in a separate map that the pod converter never folds in. So a selector on a node key works today through the inbound tags, and once the flag empties them the `Dataplane`-label fallback has nothing to substitute. Recovering it would mean also copying those node labels onto the `Dataplane`'s labels, which is a change to the pod converter rather than to listener metadata, so it is out of scope here.

### The tags-to-labels substitution inventory

Four places actually fall back to labels:

* `pkg/core/resources/apis/mesh/dataplane_helpers.go:193-204`, `IdentifyingName` returns the `kuma.io/workload` label when tags are disabled, otherwise the `kuma.io/service` tag. Used by `mads/v1/generator/assignments.go:50`, `meshtrace/plugin.go:266`, `meshaccesslog/plugin.go:108` and `meshmetric/plugin.go:146`.
* `pkg/core/resources/apis/mesh/dataplane_helpers.go:181-190`, `InboundIdentifyingName` returns the `Dataplane` KRI with the port as section name.
* `pkg/plugins/runtime/k8s/controllers/meshservice_controller.go:374-385`, the selector swaps `DataplaneTags` for `DataplaneLabels` using the same Service selector keys.
* `meshloadbalancingstrategy`, in `priority.go:123-137` (`resolveAffinityValues`) and `locality_aware.go:151-154`, prefers inbound tags and falls back to Pod labels, filtered down to affinity keys so unrelated labels don't leak.

Everywhere else the data is dropped or the rule is relaxed:

* `pkg/insights/resyncer.go:443`, `:668` skip `ServiceInsight` entries for dataplanes with no tags.
* `pkg/core/resources/apis/mesh/dataplane_validator.go:105-111` turns the `kuma.io/service` requirement into a no-op.
* `pkg/xds/topology/outbound.go:386`, `:423` still derive `Locality` from `getZone(inboundTags)`, tags only.
* `pkg/core/xds/types.go:138`, `:422`, `:435`: `Protocol()`, `ContainsTags` and selector matching are tags only.
* `pkg/xds/generator/inbound_proxy_generator.go:96`, the inbound listener's tags, which is the gap this MADR closes.

### Endpoint `Locality` under `InboundTagsDisabled`

This is the same bug as the listener one, one component over, and it is worth spelling out because it is a correctness bug on the traffic path rather than on a selector.

A local endpoint gets its zone from `getZone(inboundTags)` (`outbound.go:1002-1007`), which reads `kuma.io/zone` off the endpoint's tag map and returns `nil` when the key is absent:

```go
func getZone(tags map[string]string) *string {
	if zone, ok := tags[mesh_proto.ZoneTag]; ok {
		return &zone
	}
	return nil
}
```

`GetLocality` (`outbound.go:976-999`) returns `nil` the moment the zone is `nil`, so the endpoint carries no `Locality` at all: no `Zone`, no `Priority`.

```go
func GetLocality(localZone string, otherZone *string, localityAwareness bool) *core_xds.Locality {
	if otherZone == nil {
		return nil
	}
	// ...
}
```

Under `InboundTagsDisabled` the destination `Dataplane`'s inbound tags are empty, so `getZone` returns `nil` at both local call sites: the legacy path (`outbound.go:386`) and the `MeshService` local path (`outbound.go:423`). The effect is silent and it is a traffic effect, not a matching one. With the flag off, every local endpoint gets `Priority: local` and cross-zone endpoints `Priority: remote`, which is how Kuma keeps traffic in-zone; with the flag on, the local endpoints have no priority, the local/remote split collapses, and zone-aware routing quietly stops. Nothing errors, and no policy is involved.

The data is still there, and it is already in scope. The zone lives on the destination `Dataplane`'s `kuma.io/zone` label, and both call sites already hold that label set: `dataplane.GetMeta().GetLabels()` at `outbound.go:384` and `dpp.GetMeta().GetLabels()` at `:421`, one line above each `GetLocality` call. So the fix is the same substitution this MADR applies to listeners: `getZone` should fall back to the `kuma.io/zone` label when the tag is absent. It is a small, self-contained change, and it does not touch the KDS or cross-zone builders (`:263`, `:303`), which already pass a real zone rather than reading it from tags.

This one stays out of scope. It is a traffic-path fix, not a listener-metadata one, so folding it into this MADR would break the "traffic is unaffected" property the listener fix relies on and would need its own reliability and backport story. It is tracked as a separate follow-up; this section exists so whoever picks it up has the diagnosis and the fix already written down.

### `io.kuma.labels`, and why we don't extend it

We already ship labels to Envoy under a key called `io.kuma.labels`, and it was added for the same problem this MADR is about, one layer down.

It arrived in `1c62231766`, "fix(mlbs): migrate AffinityTags to use pod labels when inbound tags are disabled" (https://github.com/kumahq/kuma/pull/16030, May 2026). That PR's motivation is worth quoting, because it is our problem with different nouns:

> When `KUMA_EXPERIMENTAL_INBOUND_TAGS_DISABLED=true` is set, Kuma strips inbound tags from Dataplane resources to reduce memory overhead. However, `MeshLoadBalancingStrategy.LocalityAwareness.LocalZone.AffinityTags` relies on inbound tags [...] This results in locality-aware load balancing via `AffinityTags` silently failing when `KUMA_EXPERIMENTAL_INBOUND_TAGS_DISABLED` is enabled. Pod labels, however, are always available from Kubernetes pod metadata regardless of this flag.

Same flag, same silent failure, same answer: read labels, which are there either way. `MeshLoadBalancingStrategy` got that fix; `MeshProxyPatch` did not. It also tells us why the flag exists at all, which is to cut memory overhead, so the tags are not coming back.

The doc comment on the writer records the design (`pkg/xds/envoy/metadata/v3/metadata.go:44-49`):

```go
// EndpointMetadataWithLabels builds Envoy endpoint filter metadata that includes
// inbound tags (under the "envoy.lb" key) and pod/workload labels (under the
// "io.kuma.labels" key). Labels are encoded under a separate key so they remain
// available even when KUMA_EXPERIMENTAL_INBOUND_TAGS_DISABLED strips inbound
// tags. The "prefer tags, fall back to labels" semantics are applied by
// consumers (see resolveAffinityValues / resolveEndpointAffinityValue).
```

So the endpoint path hit case 3 first and solved it by shipping labels alongside tags under a second key, leaving consumers to prefer tags and fall back to labels. Given that, the obvious question is why the listener path doesn't just do the same thing. Three reasons.

First, the endpoint path had a real reason for the second key, but not the one it looks like. Endpoint tags go under `envoy.lb` and `envoy.transport_socket_match` (`metadata.go:11-28`), and Envoy does read both: `envoy.lb` drives LB subset matching, which is what makes `TrafficRoute` traffic splitting work (`pkg/xds/envoy/clusters/configurers.go:183`: "LbSubset is required for MetadataMatch in Weighted Cluster in TCP Proxy to work"), and `envoy.transport_socket_match` selects the TLS context. But both match against a fixed set of declared keys, not against whatever the endpoint happens to carry. Subset selectors are built only from the split clusters' own tag keys (`lb_subset_configurer.go:16-29`, `zoneproxy/generator.go:37-50`), and the transport-socket match from the cluster's declared tags (`client_side_mtls_configurer.go:66-72`); Envoy tolerates any extra key on the endpoint and ignores it. So arbitrary Pod labels in `envoy.lb` are inert unless a label key collides with one of those declared keys and carries a different value, which is plausible because the tags are themselves derived from Pod labels. The reason the second key is right is in the writer's own comment (`metadata.go:44-49`): labels have to stay available when `InboundTagsDisabled` strips the tags out of `envoy.lb`, and a separate key keeps the "prefer tags, fall back to labels" resolution clean while sidestepping that collision class. `io.kuma.tags` has neither problem: nothing in Envoy reads it — `grep` finds no reference to it in any cluster or listener config we generate — so it feeds no matcher to collide with, and the listener fill only runs once tags are gone. It is ours to fill.

Second, `io.kuma.labels` is inert to Envoy too, and that is the point of it. It exists so one Kuma component can hand data to another through an xDS resource. `MeshLoadBalancingStrategy` receives endpoints another generator already built, so the labels are no longer in scope as Go values, and it parses them back out of the proto (`locality_aware.go:33-36, 60-64`):

```go
for _, lbEndpoint := range localityLbEndpoint.LbEndpoints {
	ed := createEndpoint(lbEndpoint, localZone)
	// ...
}

func createEndpoint(lbEndpoint *envoy_endpoint.LbEndpoint, localZone string) core_xds.Endpoint {
	endpoint := core_xds.Endpoint{}
	endpoint.Tags = envoy_metadata.ExtractLbTags(lbEndpoint.Metadata)
	endpoint.Labels = envoy_metadata.ExtractLbLabels(lbEndpoint.Metadata)
```

That is a stage-to-stage channel between our own components. `io.kuma.tags` on a listener is a surface users write policies against. The two keys look alike and do different jobs.

Envoy never sees the labels as labels, either. `MeshLoadBalancingStrategy` turns a match into `locality.sub_zone` and a load balancing weight (`locality_aware.go:114-127`), and that is the part Envoy acts on. `io.kuma.labels` only exists so the later stage can see what the earlier one knew.

The two keys are symmetric: `io.kuma.labels` is to endpoints what `io.kuma.tags` is to listeners. Both are `io.kuma.` namespaced, both are invisible to Envoy, and both exist only as Kuma-to-Kuma channels through xDS. In practice the `io.kuma.` prefix is our marker for "Envoy ignores this".

Third, it is narrower than its name suggests, and messier. It has one consumer, `meshloadbalancingstrategy/locality_aware.go:64`, and no user-facing documentation. It only appears on local-zone endpoints: `Labels` is set at `pkg/xds/topology/outbound.go:384` and `:421`, while the cross-zone builders (`fillRemoteMeshServices` at `:258` and `:298`, `fillIngressOutbounds` at `:636` and `:652`) never set it. Since labels come from Pod metadata, it is effectively Kubernetes-only too. And on the base CLA it carries every Pod label of every endpoint, straight from `dataplane.GetMeta().GetLabels()` with no filtering. `MeshLoadBalancingStrategy` trims them to the affinity keys its policy names (`affinityTagPodLabels`, `priority.go:109-121`), but only on the path where it rebuilds the assignment. With no MLBS policy in play, the unfiltered set stays in the config.

So it is a local-zone, Kubernetes-only, unfiltered, single-consumer channel that reads like a general identity surface. Worth knowing before assuming it carries labels broadly, or that a listener could just read from it.

Note also what `io.kuma.labels` confirms about the shape of our fix. The endpoint path did not invent a new idea; it shipped labels next to tags and let consumers fall back. We are doing the same thing for listeners, and merging rather than adding a key only because the constraint that forced the split on endpoints doesn't exist on listeners.

### Where the labels come from

The destination resource is already resolved at generation time and then thrown away (`pkg/plugins/policies/core/xds/meshroute/listeners.go:147-167`):

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

`GetServiceByKRI` returns a `core.Destination`, which embeds `core_model.Resource` and so exposes `GetMeta().GetLabels()`. `DestinationService` (`:68-72`) keeps only the KRI and a synthesized name, so the labels are discarded. Threading them through needs no extra lookup.

A `MeshService` keeps its identity in labels rather than its spec. The computed set lives in `pkg/core/resources/labels/registry.go:10-23`, and only six of its twelve keys ever land on a `MeshService`, since the rest are for proxies and policies. The Universal generator does not go through `Compute` at all: `desiredLabels` (`meshservice/generate/generator.go:272-284`) writes `kuma.io/mesh`, `kuma.io/display-name`, `kuma.io/env` hardcoded to `universal`, `kuma.io/zone` and `kuma.io/origin` directly, plus `kuma.io/managed-by`, which is not a computed label and is not in the registry. User labels join them when `KUMA_MESH_SERVICE_LABEL_PROPAGATION_ENABLED` is on. The Kubernetes controller adds `k8s.kuma.io/namespace` and `k8s.kuma.io/is-headless-service` (`meshservice_controller.go:373-489`). There is no `kuma.io/namespace` label in Kuma; the namespace label is `k8s.kuma.io/namespace` (`mesh_proto.KubeNamespaceTag`).

The two environments differ in how user labels get there, which matters for how much we would be copying. The Universal generator filters them through `AllowedLabelKeys` and only runs when propagation is enabled, which is off by default. The Kubernetes controller does not: `ms.Labels = maps.Clone(svc.GetLabels())` at `meshservice_controller.go:483` takes every label off the `Service` unconditionally, and `MeshServiceLabelPropagation` never reaches that path. So on Kubernetes, which is where most of this runs, a `MeshService` carries whatever labels its `Service` carries.

### A Universal `MeshService` with no tags

Worth separating two things that sound the same. A `MeshService` never has tags, in either environment: it has labels only, never tags, and the tags in question live on the `Dataplane`s behind it. So "the destination has no tags" cannot empty a Universal outbound listener, because none of the six computed labels is derived from a tag. `desiredLabels` (`generate/generator.go:272-284`) sets them unconditionally, and after `kuma.io/mesh` is stripped the listener carries `kuma.io/display-name`, `kuma.io/zone`, `kuma.io/env: universal`, `kuma.io/origin: zone` and `kuma.io/managed-by: meshservice-generator`. There is no `k8s.kuma.io/namespace`, since Universal has no namespaces, so the User case 2 pair we recommend on Kubernetes reduces to `kuma.io/display-name` alone.

What tagless `Dataplane`s do cost on Universal is user labels, and only when propagation is on. `dpContribution` (`generate/labels.go:80-183`) reads two sources: the non-reserved inbound tags that every inbound of the service agrees on, with conflicts dropped, overlaid with the `Dataplane`'s own metadata labels, both filtered through `AllowedLabelKeys`. `mergeAcrossDataplanes` (`:193-244`) then settles disagreements between `Dataplane`s by majority. Empty inbound tags simply contribute nothing to step one; the metadata labels still land, and the six computed labels are untouched. User case 3 on Universal therefore depends on operators labelling `Dataplane`s rather than on tags surviving.

A hand-written Universal `MeshService` skips the generator entirely. It goes through the API server, which calls `resource_labels.Compute` (`resource_endpoints.go:481-499`), and comes out with `kuma.io/display-name`, `kuma.io/mesh`, `kuma.io/origin: zone`, `kuma.io/zone` and `kuma.io/env: universal`. `kuma.io/managed-by` is absent, and that absence is intentional rather than an oversight: it is how the generator knows to leave the resource alone (`generator.go:319-321`). Labels the author wrote under `metadata.labels` are kept as they are, with no propagation flag and no allow list in the way, because those only gate what the generator copies off a `Dataplane`. A `MeshService` that arrived over KDS is not recomputed at all (`compute.go:126-128`), so it keeps the labels the other control plane gave it.

One Universal combination makes the outbound half moot. With `InboundTagsDisabled` on, the generator switches to one `MeshService` per `kuma.io/workload` label (`generator.go:109-114`, `:178-188`) and skips any `Dataplane` without it. Nothing on the Universal write path sets that label, so no `MeshService` is generated at all unless the operator writes `kuma.io/workload` by hand on every `Dataplane`, and with no destination resource there is no outbound listener to patch. That is a gap in the generator rather than in this fix, and it is out of scope here.

### Name collisions across destination kinds

A `MeshService` and a `MeshMultiZoneService` sharing a name is not an edge case, it is the intended shape of a multi-zone service. Our own e2e fixtures do it: `test/e2e_env/multizone/localityawarelb/meshmultizoneservice.go:24-37` creates a `MeshMultiZoneService` named `test-server` whose selector is `meshService.matchLabels: {kuma.io/display-name: test-server}`, and `test/e2e_env/multizone/meshservice/sync.go:29-38` does the same with `backend`. The aggregate is named after what it aggregates, and its selector matches on the very label we are proposing people select listeners with. A proxy in that mesh has an outbound to the local `MeshService` and an outbound to the `MeshMultiZoneService`, so `listenerTags: {kuma.io/display-name: test-server}` matches both listeners.

Nothing in labels records a resource type. There is no kind or type label anywhere in the registry, so disambiguation has to come from the computed keys, and it does not get far. Two properties make it worse than it first looks. A `MeshService` synced from another zone arrives with `kuma.io/origin: global`, not `zone`, because the global-to-zone mapper overwrites the label on the way out (`kds/context/context.go:79`, `util/meta.go:32-36`). And `TagSelector` has no way to say "this key is absent": it requires every selector key to exist on the listener (`dataplane_helpers.go:465-479`), so a resource whose labels are a subset of another's can never be selected on its own.

Put together, per pair:

* `MeshMultiZoneService` against a `MeshService` synced from another zone: the aggregate cannot be selected at all. Both carry `kuma.io/display-name: test-server`, `kuma.io/origin: global` and the same mesh and namespace; the synced `MeshService` additionally has `kuma.io/zone` and `kuma.io/env`, which the `MeshMultiZoneService` never gets (`compute.go:163` gates both on `ProvidedByZoneFlag`, which its descriptor lacks). The aggregate's label set is a strict subset of the `MeshService`'s, so every selector matching the aggregate's listener also matches the `MeshService`'s.
* `MeshMultiZoneService` against a locally created `MeshService`: separable, but only because of a difference that has nothing to do with kind. The local one keeps `kuma.io/origin: zone`, so adding `kuma.io/origin: global` excludes it, while still matching the synced `MeshService` above.
* `MeshMultiZoneService` against a `MeshExternalService`, both created on global: identical in both environments. Same display name, `kuma.io/origin: global`, no zone, no env, and on Kubernetes both sit in the global system namespace. They even land on the zone under the same hashed name, since `HashedName` omits the type (`kds/hash/hash.go:12-14`).
* `MeshExternalService` against a `MeshService` on the same zone: identical on Universal, same origin, zone and env, with no namespace label to fall back on. Kubernetes separates them only because a `MeshExternalService` is system-namespace-only.

So the canonical multi-zone setup already has a destination that `listenerTags` cannot address, and two other pairs collide outright. This is the argument for sub-option B.

### Which labels reach the listener

The computed labels are the bounded half. Six of the twelve keys in `AllComputedLabels` (`labels/registry.go:10-23`) can land on a `Mesh*Service`, the rest being for proxies and policies: `kuma.io/display-name`, `kuma.io/mesh`, `kuma.io/origin`, `kuma.io/zone`, `kuma.io/env` and `k8s.kuma.io/namespace`. Two more are written outside `Compute` and so are not in the registry at all: `kuma.io/managed-by` and, on Kubernetes, `k8s.kuma.io/is-headless-service`. Taking the candidates one at a time:

* `kuma.io/display-name`, the service's name and the direct replacement for `kuma.io/service: backend`. It carries User case 1 and User case 2, and it is the one key set unconditionally by every creation path in both environments (`compute.go:150`, `generator.go:278`).
* `k8s.kuma.io/namespace`, the other half of a Kubernetes identity, since display names only disambiguate within a namespace. Kubernetes only: `Compute` gates it on `IsK8s` (`compute.go:173`) and the controller sets it directly. Never present on Universal, which has no namespaces.
* `kuma.io/zone`, which makes zone-scoped patches possible. It is set only for a resource created on a zone, and only when its descriptor has `ProvidedByZoneFlag` (`compute.go:156-170`, where the flag check sits inside the zone branch). A `MeshMultiZoneService` fails both tests: it is created on global, and it lacks the flag. Anything created on global lacks the key, including a `MeshService` created there.
* `kuma.io/env`, `kubernetes` or `universal`. Set in the same branch under the same two conditions, so it is present and absent exactly where `kuma.io/zone` is. It is the natural key for "patch every Universal destination".
* `kuma.io/origin`, `zone` or `global`, meaning which control plane created the resource rather than where the workload runs. A `MeshMultiZoneService` is always `global`.
* `k8s.kuma.io/is-headless-service`, `true` or `false` (`meshservice_controller.go:373`, `:421`). Kubernetes only, and it describes the `Service` rather than the traffic, so it is a plausible cut too.
* `k8s.kuma.io/service-name`, the name of the `Service` the `MeshService` was built from (`meshservice_controller.go:488`, constant at `k8s/metadata/annotations.go:134`). Kubernetes only, written outside `Compute` like the two above. It deserves a place in this list: it is the closest thing to what a user meant by `kuma.io/service: backend`, and it is not always the same string as `kuma.io/display-name`.

None of these moves during a rollout, which is the property the churn argument above depends on.

We strip `kuma.io/mesh`, as the outbound path already does. A `MeshProxyPatch` is already scoped to a mesh, so it tells a policy nothing.

User labels are the unbounded half, and their source decides how much we would be copying. On Kubernetes nothing filters them: `maps.Clone(svc.GetLabels())` takes every label off the `Service`. On Universal the generator's propagation path is off by default and trimmed by `AllowedLabelKeys` when on, so the same set arrives pre-filtered, though a hand-written `MeshService` still carries its author's labels unfiltered. The same key list therefore means "a handful of keys" on Universal and "whatever the cluster stamps on its `Service`s" on Kubernetes. That asymmetry, rather than the typical size, is the argument for choosing the list ourselves.

The resulting outbound listener for `MeshService` `backend` in namespace `kuma-demo`, zone `east`:

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
      team: payments          # user label, copied from the Service on Kubernetes
```

### Implementation notes

* `DestinationService` (`meshroute/listeners.go:68-72`) has to carry the resolved `core.Destination`, or its labels, through from `CollectServices`.
* Both outbound generators share the `TagsOrNil().WithoutTags(MeshTag)` line and both need updating: `meshtcproute/plugin/v1alpha1/listeners.go:41` and `meshhttproute/plugin/v1alpha1/listeners.go:78`.
* The inbound half is two call sites, not one: `inbound_proxy_generator.go:96` and `meshtls/plugin.go:376`, which rebuilds the inbound listener when a `MeshTLS` policy applies. Both need the `Dataplane`'s labels when `GetTags()` is empty. If only the first is changed, "an inbound listener always carries a filled `io.kuma.tags`" is false whenever `MeshTLS` is in play, and the tags a proxy gets depend on which generator produced its listener. `pkg/xds/context/context.go:40` already carries `InboundTagsDisabled` into the xDS context.
* Legacy branches keep their current behaviour: `LegacyOutbound != nil`, and inbounds with tags enabled.
* Out of scope, noted because this fix rewrites the same metadata and should stay consistent with it. Two neighbours are left alone on purpose. Node labels (see "Case 3" in the deep dive) only ever lived in the inbound tags, so a selector on `topology.kubernetes.io/zone` or `kubernetes.io/hostname` breaks under the flag and the `Dataplane`-label fallback cannot restore it without also propagating those node labels onto the `Dataplane`'s labels. And the endpoint path fixed the same class of bug with a second key (`io.kuma.labels`) rather than by filling `envoy.lb` from labels; the reasoning for why listeners merge and endpoints split is in the `io.kuma.labels` section. Both would add metadata we are otherwise trying to keep minimal, so this MADR takes neither on and each is a separate follow-up.
* Open question: we may want to copy only labels whose keys are on a list we choose ourselves, rather than every label the resource happens to carry. The candidates and what each one buys are in "Which labels reach the listener" above: the six computed keys are bounded, stable and already the ones we tell people to match on, while user labels are neither bounded nor stable and on Kubernetes arrive wholesale from the `Service`. A fixed list keeps the metadata small, keeps rollouts from churning listeners, and keeps the config dump free of labels nobody selects on. It needs settling before implementation, because it is easier to add keys later than to take them away.
* A list we choose ourselves and User case 3 pull against each other. Group targeting by a user label such as `team: payments` only works if that key reaches the listener, and we cannot know those keys up front. Either the list holds computed labels only and User case 3 goes away, which takes with it the main advantage Option 4 has over Option 1, or operators name the extra keys themselves and User case 3 survives at the cost of configuration. The middle option is to reuse `MeshServiceLabelPropagation.AllowedLabelKeys`, which already exists, already rejects reserved keys and is already validated. It does not gate the Kubernetes path today, so wiring it in would mean extending it rather than adding a flag. Note its current default is "empty means allow all", which is the opposite of what we want here.
