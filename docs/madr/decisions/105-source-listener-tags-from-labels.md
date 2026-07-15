# Source Envoy listener tags from resource labels

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/17381

## Context and Problem Statement

`MeshProxyPatch` lets users pick a generated Envoy listener and patch it. You can pick a listener by `origin`, by `name`, or by `tags`/`listenerTags`. The last one matches against a set of key/value pairs that Kuma puts on every listener as Envoy filter metadata, under the key `io.kuma.tags`.

Those pairs have always come from `Dataplane` tags. Kuma is now moving identity out of tags and into labels, but the listener metadata still reads tags. So tag matching quietly stops working. The policy is accepted, it validates, and it matches nothing.

### Three cases

The same mechanism behaves differently in three setups. Only the first one works today.

| Case                              | Inbound listener tags                    | Outbound listener tags                               |
|:----------------------------------|:-----------------------------------------|:-----------------------------------------------------|
| **1. `kuma.io/service`** (legacy) | come from the proxy's own `Dataplane`    | come from the consuming `Dataplane`. Works           |
| **2. `MeshService` enabled**      | still work                               | empty, because outbounds are now resource references |
| **3. `InboundTagsDisabled`**      | empty, because inbound tags are stripped | not affected by this flag                            |

Cases 2 and 3 are separate switches, but they belong to the same migration and people usually turn them on together. With both on, a proxy has no working tag selector in either direction.

**Case 1, `kuma.io/service`.** The destination is identified by a string that gets passed along from proxy to proxy. It starts as the `kuma.io/service` tag on the destination's inbound, goes into the VIP view, then into the consuming `Dataplane`'s outbound tags, and ends up on the outbound listener. Two things are worth knowing. The tags describe the destination but live on the consumer. And in the generated path, which covers all of Kubernetes and every transparent-proxy setup, there is only one key in there: `kuma.io/service`, because the control plane hard-codes it. Outbounds with several tags (`version`, `region`) only ever existed in hand-written Universal `Dataplane`s.

**Case 2, `MeshService`.** The destination becomes a real resource, referenced by KRI, so there is no tag set to copy and the listener ends up with an empty one. `kuma.io/service` is gone from the model entirely at this point. The consumer's outbound tags are gone, the VIP path no longer feeds outbounds, and the destination's inbounds don't carry the tag either. There is no value left to read anywhere, so we cannot keep `listenerTags: {kuma.io/service: backend}` working no matter what we do. Those policies have to be rewritten. The only open question is what we give people to rewrite them to.

**Case 3, `InboundTagsDisabled`.** Inbound tags are emptied at the source, so inbound listeners hit the same wall from the other side. The data itself is still around: on Kubernetes, Pod labels are copied onto the `Dataplane`'s metadata labels whether the flag is set or not. The flag only drops the second copy that lived in tags, so everything has to read labels instead.

Cases 2 and 3 are both part of a deliberate migration that most of the codebase has already gone through. Identity is read from labels now, and listener tags are the last place still reading tags. The rest of the migration was done case by case, and listener metadata was one of the spots that got skipped. It also happens to be the one that breaks a user-facing policy API.

### Use cases

* UC1, migration: a `MeshProxyPatch` that matches `listenerTags: {kuma.io/service: backend}` stops applying once `MeshService` is on, and the user needs something else to match on.
* UC2, one destination: patch the outbound listener to `MeshService` `backend` in namespace `kuma-demo`, without knowing its VIP and without turning on unified resource naming.
* UC3, a group: patch every outbound listener to any destination labelled `team: payments`, or every destination in zone `east`.
* UC4, an inbound: patch a proxy's own inbound listener when `InboundTagsDisabled` has emptied its tags.
* UC5, parity: the same thing should work for `MeshService`, `MeshMultiZoneService` and `MeshExternalService`, which share one outbound code path.

## Design

Every option has to answer the same question: when a listener has no tags, what identifies the thing it stands for? An outbound listener stands for the destination service, which is the `Mesh*Service` resource it was generated from. An inbound listener stands for the proxy itself. For both of those, identity lives in labels now.

An outbound listener does not stand for the individual workloads behind the destination. Those are its endpoints. They are picked by the service, and their own labels already go to Envoy separately, on the endpoints, under `io.kuma.labels`. So for the listener we want the service's labels, even though the `Dataplane` labels are the right choice on the endpoint path.

### Option 1: Do nothing, require unified resource naming and `match.name`

* Good, because it needs no code change, and a KRI-based name is exact and can tell individual ports apart.
* Bad, because UC1 stays broken and silent, so users find out in production.
* Bad, because without unified naming the listener is called `outbound:<address>:<port>`, and that address is a VIP we allocate at runtime. You cannot know it up front and it changes across restarts, so there is no way to write the policy.
* Bad, because getting a stable name means turning on unified naming across the whole mesh (`meshServices: Exclusive` plus a DP feature flag). That is a lot to ask of someone who wants to patch one listener.
* Bad, because a name matches a single listener, so UC3 needs one policy per destination and has to be regenerated whenever the set changes.
* Bad, because it does nothing for UC4. `InboundTagsDisabled` has no naming escape hatch.
* Bad, because every affected listener keeps an empty `io.kuma.tags`, which looks like a listener with no tags.

### Option 2: Build a `kuma.io/service` tag

Rebuild a `kuma.io/service` value from the destination resource and keep writing it, so old policies look like they still work.

* Good, because it is a small change and nobody has to rewrite a policy, at least at first glance.
* Bad, because `kuma.io/service` has been deliberately removed, and bringing it back on the listener works against the whole point of `MeshService`.
* Bad, because the value would not match anyway. The legacy resolver gives us `<mesh>_<name>_<namespace>_<zone>_<shortName>_<port>`, and the policy says `backend`.
* Bad, because that leaves us worse off than doing nothing. The key is back, so `kuma.io/service: '*'` starts matching while `kuma.io/service: backend` still doesn't. Users find out at runtime instead of at review time.

### Option 3: A separate `io.kuma.labels` key on listeners, with new match fields

Write labels under a new listener key and add `listenerLabels`/`labels` to the `MeshProxyPatch` API.

* Good, because it keeps labels and tags cleanly apart.
* Bad, because it changes a public API across three match types, and means regenerating CRDs and OpenAPI.
* Bad, because users then have to know which of the two selectors applies to a given listener, and the answer depends on whether the destination is legacy or a real resource. That is an internal detail they should not have to care about.
* Bad, because the matcher and the behaviour come out the same as Option 4, and only the key name differs. We would be paying for an API migration to get tidier naming.
* Bad, because we have to backport this fix, and we don't want to ship new API fields and CRD schemas in a patch release.

### Option 4 (chosen): Fill `io.kuma.tags` from labels when tags are gone

Keep one key and one selector. When a listener has no tags, take them from the labels of whatever the listener stands for: the destination resource for an outbound, the proxy's own `Dataplane` for an inbound. Strip `kuma.io/mesh`, which the outbound path already does today.

This reuses the substitution Kuma has already applied elsewhere and brings the last component in line with it.

A policy that used to say `kuma.io/service: backend` becomes:

```yaml
match:
  origin: outbound
  listenerTags:
    kuma.io/display-name: backend
    k8s.kuma.io/namespace: kuma-demo
```

And UC3, which no name-based selector can do at all, becomes:

```yaml
match:
  origin: outbound
  listenerTags:
    team: payments
```

* Good, because there is no API change. The selector people already use starts working again, and it is safe to ship as a patch.
* Good, because it works without unified resource naming, which is what the issue asks for.
* Good, because it can only add matches. The affected listeners have empty tags today, so nothing that matches now can stop matching.
* Good, because labels give you more to match on than the single `kuma.io/service` string ever did. Zone, origin, env and user labels all become usable, which is what makes UC3 possible.
* Good, because one rule handles both directions, so UC4 falls out of the same change.
* Bad, because the key is called `tags` and now holds labels. That is confusing, though `io.kuma.tags` was never a user-facing "Dataplane tags" surface. It is our own listener selector bag, written in one place and read only by our policy matching.
* Bad, because labels belong to a resource while listeners belong to a port. The two listeners of a two-port `MeshService` get identical tags and `listenerTags` cannot tell them apart. You still need `match.name` with unified naming for that, and this MADR does not fix it.

#### Why we don't gate this behind a flag

`MeshServiceLabelPropagation` is gated because it writes to a stored resource that users can see and that KDS syncs around. Listener metadata is generated, short-lived, and only our own matcher reads it. Since the change can only add matches, a flag set to `false` would have no real use. User labels are already gated further up by `KUMA_MESH_SERVICE_LABEL_PROPAGATION_ENABLED`, which decides whether a `MeshService` carries user labels at all, and we don't need a second switch for that.

#### How this design evolved

The first draft tried to keep `listenerTags: {kuma.io/service: backend}` working by rebuilding the tag. Following where the tag actually came from ruled that out, because `kuma.io/service` has been deleted along with everything that fed it. Once we accepted that no design keeps those policies untouched, the question changed from "how do we save the selector" to "what do we give people instead".

The second draft added a `kuma.io/kri` tag with the full KRI, to solve per-port targeting. We dropped it. It repeated the listener name under unified naming, and a KRI points at one resource, so it did nothing for group targeting.

The third draft only covered outbound and treated the `InboundTagsDisabled` gap as a separate problem. Looking into case 3 showed it is the same bug reached through a different switch, with the same fix, so covering one direction would have meant answering the same question twice.

## Security implications and review

Low. The data we propagate is metadata the proxy's owner can already see. The outbound exists because the destination is reachable, and its VIP and port are already in the config dump.

* Listener metadata shows up in `/config_dump`, so any user label on a `MeshService` is visible to anyone who can reach the Envoy admin API. That is already true for endpoint labels, and operators who care can use the existing `AllowedLabelKeys` allowlist.
* Someone could set a label like `kuma.io/display-name` and shadow a tag the control plane computes, so reserved `kuma.io/` tags have to be written last and win over user labels. We already reject reserved keys in the label-propagation config, so this matches what we do elsewhere.

## Reliability implications

* Nothing that matches today can break, because the affected listeners have empty `io.kuma.tags` and every non-empty selector already fails against them.
* New matches will happen, which is the point, but it is still a behaviour change and needs a changelog entry. A policy matching `listenerTags: {kuma.io/zone: east}` with no `origin` set will start matching listeners it used to miss. This only affects policies that match `kuma.io/` keys, leave `origin` unset, and run on an affected proxy.
* Traffic is unaffected. `io.kuma.tags` only exists on listeners and only our matcher reads it. Clusters and endpoints use other keys that we don't touch, so there is no path from here to load balancing, mTLS or endpoint selection.
* Labels are not recomputed for resources that came from elsewhere, so a `MeshService` synced from global keeps the labels it arrived with. Tags built at generation time have to cope with labels computed by another control plane. A `MeshMultiZoneService`, for instance, has no `kuma.io/zone` label at all, so that key just won't be there.
* Listener metadata now changes when labels change, so editing a label triggers an xDS update for every proxy with an outbound to that destination. Labels rarely change and we already regenerate the listener on any `MeshService` change, so this is small.
* Every affected golden file gains an `io.kuma.tags` block. It is a big mechanical diff, and it is worth actually reading it to catch labels we didn't mean to expose.

## Backport

This is a silent regression in released versions, so we backport the fix to release-2.13 and release-2.14. That shapes the design, and it is the main reason we pick Option 4 over Option 3. Option 4 only changes how we fill generated metadata, so there is no API, no CRD schema and no stored resource to change in a patch. And since it can only add matches, it can't break a policy that works today.

The two halves don't backport the same way, because the features they react to landed at different times:

* The outbound half goes to 2.13 and 2.14. Everything it needs (`Outbound.TagsOrNil`, `DestinationService`, `GetServiceByKRI`) is on both.
* The inbound half only goes to 2.14 and master. `InboundTagsDisabled` doesn't exist on 2.13, so inbound tags are always populated there and there is nothing to fix.
* Group targeting by user label (UC3) also only works from 2.14. `MeshServiceLabelPropagation` doesn't exist on 2.13, so a `MeshService` there only has labels the control plane computed. `kuma.io/display-name` and `kuma.io/zone` work on 2.13; `team: payments` does not.

One practical note for whoever does the cherry-pick: label computation moved packages between 2.13 and 2.14, from `core_model.ComputeLabels` in `pkg/core/resources/model/resource.go` to `pkg/core/resources/labels`. The 2.13 pick will conflict there. The computed labels themselves are the same, so it is import paths and call shape to fix up, and the behaviour stays put.

Because the tags a listener carries depend on which labels that version computes, a policy written for 2.14 may not match on 2.13. That comes with backporting a fix whose payload is "whatever labels this version has", and it belongs in the release notes so people don't hit it cluster by cluster.

## Implications for Kong Mesh

The downstream project ships policies that patch listeners by tag, so it gets both the fix and the forced `kuma.io/service` rewrite. Its `MeshProxyPatch`-equivalent policies and any bundled defaults need a look before this ships, since they are already quietly broken anywhere `MeshService` or `InboundTagsDisabled` is on. Nothing there needs an API or CRD change, because nothing here does, and we add no computed label, so the mirrored label registry stays as it is.

## Decision

Listener `io.kuma.tags` is filled from labels wherever tags no longer exist.

1. Outbound listeners backed by a real resource take the labels of the destination `Mesh*Service` resource itself, meaning the `MeshService`, `MeshMultiZoneService` or `MeshExternalService` the listener was generated from. We do not use the labels of the `Dataplane`s behind it, because the listener stands for the service. Per-workload labels already reach Envoy on the endpoints, under `io.kuma.labels`.
2. Inbound listeners whose tags have been emptied by `InboundTagsDisabled` take the labels of the proxy's own `Dataplane`.

Point 2 needs saying out loud, because the flag's name suggests the opposite. Turning off inbound tags should not turn off listener tags. The flag decides what goes on the `Dataplane`'s inbounds, and the listener still has to say what it is. An inbound listener always carries a filled `io.kuma.tags`: from tags when they exist, from the `Dataplane`'s labels when they don't. We never leave it empty.

Both cases strip `kuma.io/mesh`, which is what the outbound path already does. Reserved `kuma.io/` tags come from the control plane and win over user labels. Legacy paths keep working as they do today, meaning a `Dataplane` with real outbound tags, or inbounds with tags enabled.

There is no `MeshProxyPatch` API change. The fix is entirely in how we fill the listener metadata. It is unconditional and has no new flag.

Policies matching `kuma.io/service: <name>` have to be rewritten against labels, usually `kuma.io/display-name: <name>` plus `k8s.kuma.io/namespace: <ns>`. We are not choosing to break them: `kuma.io/service` simply doesn't exist once `MeshService` is in use. We don't invent a replacement value, because any value we make up is a guess, and users would find out at runtime.

## Notes

* Open topic: per-port targeting. Labels belong to a resource and listeners belong to a port, so the listeners of a multi-port destination look identical to `listenerTags`. Only `match.name` with unified naming separates them today. We are leaving this open instead of guessing at an API for it.
* Open topic: whether the empty `io.kuma.tags: {}` we write for listeners with no tags should just be left out. It reads badly, but changing it churns golden files across the repo and this fix doesn't need it.
* Follow-up: `MeshProxyPatch` can't warn when a selector matches nothing, which is why this shipped quietly in the first place. A validation or inspect-time signal for "this policy matched zero resources" would have caught it at apply time. `origin` has the same problem: it is a free-form string with no enum, so a typo also matches nothing.
* Follow-up: endpoint `Locality` still reads tags only, with no label fallback, so zone-based locality is silently lost under `InboundTagsDisabled`. Same bug, different component, out of scope here.
* Follow-up: `ServiceInsight`s are dropped for dataplanes with no tags instead of falling back to labels.
* `MeshGateway` listeners have no `io.kuma.tags` at all, so `listenerTags` can never match one. This decision doesn't change that, but it is worth writing down.

## Deep dive

Everything below is the supporting detail for the sections above.

### The `io.kuma.tags` channel

`TagsKey = "io.kuma.tags"` is defined at `pkg/xds/envoy/metadata/v3/metadata.go:99`. One thing writes it, `TagsMetadataConfigurer` (`pkg/xds/envoy/listeners/v3/tags_metadata.go:15-27`):

```go
l.Metadata.FilterMetadata[envoy_metadata.TagsKey] = &structpb.Struct{
	Fields: envoy_metadata.MetadataFields(c.Tags),
}
```

Only `MeshProxyPatch` reads it, at `listener_mod.go:86`, `network_filter_mod.go:136` and `http_filter_mod.go:159`. Each one does:

```go
listenerTags := envoy_metadata.ExtractTags(listenerProto.Metadata)
if !mesh_proto.TagSelector(pointer.Deref(l.Match.Tags)).Matches(listenerTags) {
	return false
}
```

`TagSelector.Matches` (`api/mesh/v1alpha1/dataplane_helpers.go:465-479`) ANDs exact key/value pairs, with `*` matching any value. If a key is missing, the match fails.

The configurer has no empty guard. `EndpointMetadata` returns `nil` for empty tags, but this one always writes the key. Empty tags therefore give you `io.kuma.tags: {}`, an empty struct rather than a missing key. That is why the failure is quiet: the bag is there and it is empty, so every non-empty selector fails on its first lookup.

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

The data survives, because Pod labels are copied onto the `Dataplane`'s metadata labels either way (`pod_converter.go:62-84`):

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

`mergeLabels` (`:455-464`) is a plain clone and copy. The flag adds no label mechanism, it only drops the copy that lived in tags.

### The tags-to-labels substitution inventory

Four places actually fall back to labels:

* `pkg/core/resources/apis/mesh/dataplane_helpers.go:193-204`, `IdentifyingName` returns the `kuma.io/workload` label when tags are disabled, otherwise the `kuma.io/service` tag. Used by `mads/v1/generator/assignments.go:50`, `meshtrace/plugin.go:266`, `meshaccesslog/plugin.go:108` and `meshmetric/plugin.go:146`.
* `pkg/core/resources/apis/mesh/dataplane_helpers.go:181-190`, `InboundIdentifyingName` returns the `Dataplane` KRI with the port as section name.
* `pkg/plugins/runtime/k8s/controllers/meshservice_controller.go:374-385`, the selector swaps `DataplaneTags` for `DataplaneLabels` using the same Service selector keys.
* `meshloadbalancingstrategy`, in `priority.go:123-137` (`resolveAffinityValues`) and `locality_aware.go:151-154`, prefers inbound tags and falls back to Pod labels, filtered down to affinity keys so unrelated labels don't leak.

Everywhere else the data is dropped or the rule is relaxed:

* `pkg/insights/resyncer.go:443`, `:668` skip `ServiceInsight` entries for dataplanes with no tags.
* `pkg/core/resources/apis/mesh/dataplane_validator.go:105-111` turns the `kuma.io/service` requirement into a no-op.
* `pkg/xds/topology/outbound.go:386`, `:422` still derive `Locality` from `getZone(inboundTags)`, tags only.
* `pkg/core/xds/types.go:138`, `:422`, `:435`: `Protocol()`, `ContainsTags` and selector matching are tags only.
* `pkg/xds/generator/inbound_proxy_generator.go:96`, the inbound listener's tags, which is the gap this MADR closes.

One caveat on `io.kuma.labels` (`metadata.go:104`): the name sounds general, but it is a closed loop of three sites. It is written at `endpoints/v3/endpoints.go:41`, encoded and decoded in `metadata.go:50` and `:71`, and read by a single consumer, `meshloadbalancingstrategy/locality_aware.go:64`. Don't treat it as a general identity channel.

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

A `MeshService` keeps its identity in labels rather than its spec. The computed set lives in `pkg/core/resources/labels/registry.go:10-23`. The Universal generator (`meshservice/generate/generator.go:272-284`) sets `kuma.io/mesh`, `kuma.io/display-name`, `kuma.io/managed-by`, `kuma.io/env`, `kuma.io/zone` and `kuma.io/origin`, plus user labels when `KUMA_MESH_SERVICE_LABEL_PROPAGATION_ENABLED` is on. The Kubernetes controller adds `k8s.kuma.io/namespace` and `k8s.kuma.io/is-headless-service` (`meshservice_controller.go:373-489`). There is no `kuma.io/namespace` label in Kuma; the namespace label is `k8s.kuma.io/namespace` (`mesh_proto.KubeNamespaceTag`).

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
      team: payments          # user label, when label propagation is enabled
```

### Implementation notes

* `DestinationService` (`meshroute/listeners.go:68-72`) has to carry the resolved `core.Destination`, or its labels, through from `CollectServices`.
* Both outbound generators share the `TagsOrNil().WithoutTags(MeshTag)` line and both need updating: `meshtcproute/plugin/v1alpha1/listeners.go:41` and `meshhttproute/plugin/v1alpha1/listeners.go:78`.
* The inbound half is `inbound_proxy_generator.go:96`, which needs the `Dataplane`'s labels when `iface.GetTags()` is empty. `pkg/xds/context/context.go:40` already carries `InboundTagsDisabled` into the xDS context.
* Legacy branches keep their current behaviour: `LegacyOutbound != nil`, and inbounds with tags enabled.
