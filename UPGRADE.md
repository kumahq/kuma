This document guides you through the process of upgrading `Kuma`.

First, check if a section named `Upgrade to x.y.z` exists,
with `x.y.z` being the version you are planning to upgrade to.

If such a section does not exist, the upgrade you want to perform
does not have any particular instructions.

## Upgrade to `3.0.0`

### The legacy overview paths `dataplanes+insights` and `zones+insights` are removed

`GET /meshes/{mesh}/dataplanes+insights` and `GET /zones+insights`, along with their `/{name}` forms, were kept as aliases when overviews moved to `_overview`. They are now removed and answer `404`.

**Action required**

Use the replacements, which have been available for several releases and return the same payload:

- `/meshes/{mesh}/dataplanes+insights` becomes `/meshes/{mesh}/dataplanes/_overview`
- `/meshes/{mesh}/dataplanes+insights/{name}` becomes `/meshes/{mesh}/dataplanes/{name}/_overview`
- `/zones+insights` becomes `/zones/_overview`
- `/zones+insights/{name}` becomes `/zones/{name}/_overview`

### Kubernetes probes on the sidecar always use the readiness port

The injected `kuma-sidecar` container had its liveness, readiness and startup probes pointed at the Envoy admin port (`9901`) when `experimental.envoyAdminUnixSocket` was off, and at the kuma-dp readiness port (`9902`) when it was on. Those two settings are decided in different places - the probe port is written into the Pod by the injecting webhook at admission, the admin transport is chosen at bootstrap by whichever control plane instance answers - so a rolling control plane upgrade, or a change to the flag, left a window where they disagreed and pods went into `CrashLoopBackOff`.

Probes now always use the readiness port, whatever the admin transport, and the readiness port is always excluded from inbound transparent proxy interception. To keep the same meaning on both transports, `/ready` on the readiness port is now proxied to the Envoy admin `/ready` in both cases, so a Pod is marked ready only once Envoy has its configuration. Previously the readiness port answered `READY` as soon as kuma-dp was up when admin ran over TCP; the probes did not use that port in that mode, so no probe changes meaning.

Pods injected before the upgrade keep their existing probes on `9901` and continue to work: the `kuma:envoy:admin` listener still serves `/ready` on that port. They move to the readiness port the next time they are recreated.

The sidecar now also receives `KUMA_READINESS_PORT` from `bootstrapServer.params.readinessPort`, so kuma-dp listens on the port the probes use. Previously the injector took the probe port from the control plane setting while kuma-dp took its listen port from its own `KUMA_READINESS_PORT`, and the two agreed only because both default to `9902`. For the same reason `bootstrapServer.params.readinessPort` no longer accepts `0`; a control plane configured that way now fails to start instead of injecting probes for port `0`.

**Action required**

None for most users. If you have a `NetworkPolicy`, a monitoring check, or a `ContainerPatch` that pins the sidecar probe port to `9901`, update it to `9902` (or to `bootstrapServer.params.readinessPort` if you changed it). If you set `bootstrapServer.params.readinessPort` to `0`, or set `KUMA_READINESS_PORT` on the sidecar by hand to a value other than the control plane setting, remove it.

### The `k8s.kuma.io/service-account` label on a `Dataplane` is managed by the control plane

On Kubernetes this label is computed by the control plane from the Pod's ServiceAccount and is not meant to be set by hand. The admission webhook now rejects any `Dataplane` create or update that carries `k8s.kuma.io/service-account`, unless the request comes from the control plane itself or from another user listed in `runtime.kubernetes.allowedUsers`, on both Zone and Global control planes. A proxy is also rejected at xDS authentication when the label on its `Dataplane` does not match the ServiceAccount of its Pod. Other resource types are unaffected, as are resources synced over KDS and resources written by the control plane.

**Action required**

Remove `k8s.kuma.io/service-account` from any `Dataplane` you apply yourself, including GitOps-managed ones. A `Dataplane` whose label does not match its Pod stops connecting after the upgrade until the label is dropped.

### A request matching no `MeshHTTPRoute` rule now gets a `404`

When a `MeshHTTPRoute` applies to a destination, a request that matches none of its rules is answered with `404` instead of being sent to that destination. This is what the Gateway API requires of an `HTTPRoute`, and it is what the GAMMA conformance suite asserts. Before this change the unmatched request fell through to the destination service as if no route existed.

The rules of a `MeshHTTPRoute` apply to every HTTP port of the destination when the `to[].targetRef` names a `MeshService` without a `sectionName`, so the `404` covers those ports too, including ports no rule mentions. Ports whose protocol is not HTTP based are unaffected. A destination with no `MeshHTTPRoute` at all is also unaffected and keeps reaching its service.

The case worth checking is a route written only to anchor another policy. A `MeshHTTPRoute` matching `/api` that exists so a `MeshTimeout`, `MeshRetry`, or `MeshAccessLog` can target it now answers `404` on every other path of that destination. On a gRPC destination the same `404` reaches the client as `UNIMPLEMENTED`.

**Action required**

Add an explicit catch-all rule to any `MeshHTTPRoute` that should keep passing unmatched traffic to its destination:

```yaml
rules:
  - matches:
      - path:
          type: PathPrefix
          value: /
    default:
      backendRefs:
        - kind: MeshService
          labels:
            kuma.io/display-name: backend
          port: 80
```

Put it on the route itself when the other policies targeting that route should also cover the unmatched traffic, or on a second `MeshHTTPRoute` when they should not.

### RBAC: control plane now reads Gateway API `GRPCRoute`s

The Helm-installed control plane `ClusterRole` now grants `get`, `list`, and
`watch` on `gateway.networking.k8s.io` `grpcroutes`, alongside the existing
`httproutes` and `referencegrants` permissions. This is required for Kuma to
watch and translate Gateway API `GRPCRoute` resources into mesh routing
configuration.

**Action required**

If you manage the control plane RBAC outside of Helm (for example via GitOps or
manual manifests), add the same `grpcroutes` read permissions to the control
plane `ClusterRole`.

### `MeshLoadBalancingStrategy` cross-zone settings now require a `MeshMultiZoneService` `to` target

`MeshLoadBalancingStrategy.spec.to[].default.localityAwareness.crossZone` is now accepted only when that `to` entry targets a `MeshMultiZoneService`. Create and update validation now rejects the same `crossZone` block on `Mesh`, `MeshService`, and `MeshExternalService` targets.

**Action required**

Move any existing `crossZone` configuration under `to` entries whose `targetRef.kind` is `MeshMultiZoneService` before upgrading. Keep `localityAwareness.localZone`, `loadBalancer`, and other supported settings on the remaining target kinds as needed.

### Generated Gateway API producer routes now win the same ties as hand-written ones

A `MeshHTTPRoute` generated from a Gateway API `HTTPRoute` whose parent (a `Service` or `MeshService`) lives in the `HTTPRoute`'s own namespace is now created in that namespace, instead of always landing in the Kuma system namespace. This is what the policy role model calls a producer route, and putting it in the right namespace gives it `kuma.io/policy-role=producer`, the same role a hand-written `MeshHTTPRoute` targeting the same `Mesh` and `to` entry gets. Before this change, every generated route landed in the system namespace and was ranked `system`, so it always lost to an equivalent hand-written producer route even when the two expressed the same intent.

A generated route whose parent is in a different namespace (a consumer route) still lands in the Kuma system namespace and keeps its `system` role and precedence unchanged.

Because the route now lives in the `HTTPRoute`'s namespace, its `k8s.kuma.io/namespace` label changes to match, it now syncs across zones like any other namespaced policy, and it applies to dataplanes in remote zones the same way a hand-written route in that namespace would.

**Action required**

If a mesh has both a generated producer route and a hand-written `MeshHTTPRoute` targeting the same `Mesh` and `to` entry, check which one you expect to win: before this upgrade the hand-written route always won, after it the two tie and the outcome falls back to resource name. Remove or adjust one of them if you relied on the previous, implicit precedence.

If any policy selects the generated route by its old `k8s.kuma.io/namespace: <kuma-system>` label (or your system namespace), update it to the `HTTPRoute`'s namespace instead. The control plane moves each existing generated route to its new namespace the next time its `HTTPRoute` is reconciled, so no manual migration of the `MeshHTTPRoute` object itself is needed.

### Gateway API HTTPRoute conflicts on the same parent now resolve by creationTimestamp

When two Gateway API `HTTPRoute`s attach to the same parent with equally specific matches, the generated `MeshHTTPRoute`s used to tie-break by resource name alone, so which route won depended on name rather than which `HTTPRoute` was applied first. Each generated `MeshHTTPRoute` now carries its owning `HTTPRoute`'s `creationTimestamp` as a label, and that label breaks the tie in favor of the older `HTTPRoute`; name order is used only when the timestamps also match.

A hand-written `MeshHTTPRoute` has no such label, so it still takes precedence over any generated route it ties with on an exact conflict, same as before this change.

**Action required**

None. Existing generated `MeshHTTPRoute`s are backfilled with the timestamp label the next time their `HTTPRoute` is reconciled. If two `HTTPRoute`s previously tied and were resolved by name, check whether the new creationTimestamp-based outcome matches what you expect.

### The `BUILTIN` gateway type and its statistics are removed from the API

The built-in gateway implementation was removed over the previous releases, and the Dataplane validator has been rejecting `networking.gateway.type: BUILTIN` since then. The remaining API surface is now gone too:

- `Dataplane.networking.gateway.type` no longer has a `BUILTIN` value. `DELEGATED` is the only gateway type. A `Dataplane` carrying `type: BUILTIN` is now rejected while it is parsed, as an unknown enum value, instead of producing the previous `BUILTIN gateways are no longer supported, use DELEGATED instead` validation error.
- `MeshInsight.dataplanesByType.gatewayBuiltin` and the `gateway_builtin` `ServiceInsight` service type are removed. `dataplanesByType.gateway` now reports the delegated gateway totals only.
- `GET /meshes/{mesh}/dataplanes+insights?gateway=builtin` is no longer a valid filter. Use `gateway=delegated`, or `gateway=true` for any gateway.
- `GET /meshes/{mesh}/service-insights?type=gateway_builtin` is no longer a valid filter and returns `400`. Use `type=gateway_delegated`.
- The `gatewayBuiltin` object disappears from the `/global-insight` response, in both `dataplanes` and `services`.

All three protobuf ordinals are reserved, so they can never be reused for something else, and a Zone control plane on an older version still syncs to a Global control plane on this one.

**Action required**

Delete every `Dataplane` with `networking.gateway.type: BUILTIN` before upgrading. `2.14.x` still accepts them, and after the upgrade the control plane fails to parse them as an unknown enum value, which breaks listing the `Dataplane`s of that mesh. Find them with `kumactl get dataplanes -o yaml` per mesh, or `kubectl get dataplanes -A -o yaml`, and grep for `BUILTIN`.

Also drop `type: BUILTIN` from any manifest you still keep under source control, and stop consuming the `gatewayBuiltin` fields and the `gateway=builtin` / `type=gateway_builtin` filters if you query the API directly.

### Control plane RBAC is narrowed on Kubernetes

Two `ClusterRole` rules that existed only for the built-in gateway are reduced to what the control plane actually uses:

- `apps`: `deployments` is dropped entirely and `replicasets` keeps only `get`, `list` and `watch`. The control plane reads ReplicaSets to resolve the workload behind a serviceless Pod and never writes either resource.
- `core`: `services` loses `create` and `delete`. The control plane only annotates existing Services with `ingress.kubernetes.io/service-upstream`.

**Action required**

If you manage the control plane RBAC yourself, apply the same narrowing.

### Names of `Mesh`, `Zone`, `MeshService`, `MeshExternalService` and `MeshMultiZoneService` must be RFC 1035 labels

Names of these resources are rendered into DNS hostnames by `HostnameGenerator`, so a name that is not a valid [RFC 1035](https://www.rfc-editor.org/rfc/rfc1035.html) label produces an invalid hostname. Since 2.10.x such names were accepted with a deprecation warning, now they are rejected:

> A lowercase RFC 1035 label must have at most 63 characters, consist of lower case alphanumeric characters or '-', start with an alphabetic character and end with an alphanumeric character.

This also covers `MeshMultiZoneService` names longer than 63 characters, which were deprecated in 2.14.x.

On Kubernetes the name of the underlying object is validated, not the `<name>.<namespace>` core name, so namespaced resources are unaffected by the namespace part.

The `KUMA_MULTIZONE_ZONE_NAME` control plane setting is validated with the same rule instead of the RFC 1123 label rule, so a Zone control plane with a non-conforming name fails to start instead of failing later when the `Zone` resource is created on the Global control plane.

**Action required**

Existing resources are not rewritten, but any create or update of a resource with a non-conforming name is rejected. This includes resources written by the control plane itself, so a `Mesh` or `MeshService` with a non-conforming name stops being synced from the Global to the Zone control planes, and a Zone control plane whose name is not a valid label fails to register. Rename them before upgrading:

```sh
kumactl get meshes
kumactl get zones
kumactl get meshservices
kumactl get meshexternalservices
kumactl get meshmultizoneservices
```

### `Mesh.mtls` is removed from the API

The `mtls` field is gone from the `Mesh` resource, together with `CertificateAuthorityBackend` and everything it carried: `enabledBackend`, `backends`, `skipValidation`, `dpCert.rotation.expiration`, `dpCert.requestTimeout`, `conf`, `mode` and `rootChain.requestTimeout`. The field number is reserved, so it can never be reused for something else. The empty `Mesh.routing` message goes with it — every field it held had already been removed.

`mtls` had been inert since the legacy SDS path and the CA plugin subsystem were removed: it issued nothing, was provisioned by nothing and validated by nothing. Removing the field only removes the ability to write it down.

Nothing breaks on the way in: `mtls` is now an unknown field, and the control plane accepts unknown fields, so an existing `Mesh` still loads and a manifest that still carries `mtls` still applies. The field is silently dropped the next time the resource is written, so reading a `Mesh` back (`kumactl get mesh`, `kubectl get mesh -o yaml` after a rewrite) no longer shows `mtls`. The Kubernetes `Mesh` CRD is unaffected — its `spec` is `x-kubernetes-preserve-unknown-fields`, so `mtls` was never part of the CRD schema and a stored `Mesh` keeps it in etcd until it is rewritten.

Two user-visible surfaces change with it. `kumactl get meshes` no longer prints the `mTLS` column, so the table is now `NAME` and `AGE`. `kumactl export --profile federation` no longer rewrites `builtin` backends into `provided` ones pointing at the CA secrets, because there is no backend left to rewrite; the orphaned `<mesh>.ca-builtin-*` Secrets are still exported like any other secret.

**Action required**

None to keep the control plane running. Drop `mtls` from the manifests you keep under source control so they describe what is actually stored. If you have not migrated to `MeshIdentity` yet, do that first: see the sections below for what that migration involves.

### `Mesh.mtls` no longer produces mTLS

The transport socket builders no longer read `Mesh.spec.mtls`. A proxy gets an mTLS transport socket — inbound and outbound — only when a `MeshIdentity` matches it; the identity certificate and the trust bundle come from `MeshIdentity` and `MeshTrust`, never from the mesh CA backend. A mesh whose only identity source is `mtls` now serves and accepts plaintext, and `MeshTLS` and `MeshTrafficPermission` no longer apply to its proxies.

The `mtls` field itself is still accepted, but it no longer issues anything: the legacy SDS path is gone, so no certificate is ever generated for it.

**Action required**

Migrate every mesh that still relies on `mtls` to `MeshIdentity` before upgrading. A mesh already covered by a `MeshIdentity` is unaffected.

Two things go away with the legacy path: `SNIFromTags`/`TagsFromSNI`-style tag-encoded SNIs (outbounds to a destination that is not a real resource are plaintext, so they carry no SNI at all), and `MeshService.status.tls` gating of *permissive* meshes — the status is still computed and still gates outbound TLS for `MeshIdentity` proxies.

### The legacy mesh CA is no longer added to the `MeshTrust` bundle

While both systems coexisted, a mesh that had `mtls` enabled *and* at least one `MeshTrust` got the mesh CA injected into the trust bundle under the mesh name, so a `MeshIdentity` proxy still trusted peers presenting a legacy mesh certificate. That bridge is removed together with the legacy SDS server: the trust bundle now contains only the CAs declared by `MeshTrust` resources.

`DataplaneInsight.mTLS` (and the `MeshInsight.mTLS` aggregation built from it) is now populated exclusively from `MeshIdentity` issuance. A proxy with no workload identity reports no certificate at all instead of the legacy mesh CA's.

**Action required**

Complete the migration to `MeshIdentity` *before* upgrading — this is the point where a partially migrated mesh breaks. Once upgraded, proxies still holding a legacy mesh certificate are no longer trusted by their `MeshIdentity` peers, and there is no rolling path back.

### `Mesh.mtls` backends are no longer provisioned or validated

The CA plugin subsystem is gone — both the `builtin` and the `provided` CA managers, and the code that drove them on every Mesh create, update and Kubernetes reconcile. `mtls.backends` is now inert configuration: nothing generates a CA for a `builtin` backend, nothing reads the certificate and key of a `provided` one, and nothing checks either at admission time.

Three validations disappear with it. An unknown `mtls.backends[].type` is accepted instead of being rejected with `could not find installed plugin for this type`. A `provided` backend missing `conf.cert` or `conf.key` is accepted. Changing `mtls.enabledBackend` while mTLS is enabled is accepted instead of being rejected with `Changing CA when mTLS is enabled is forbidden`.

Deleting a Secret is no longer blocked when an `mtls` backend still references it, on both Universal and Kubernetes — the check dispatched into the CA managers, which no longer exist. `MeshIdentity` and `MeshTrust` secrets were never covered by it, and the `inter-cp-ca` and `envoy-admin-ca` global secrets are separate PKIs, untouched by this change.

**Orphaned CA secrets are left in place**

`builtin` CA material was persisted as `<mesh>.ca-builtin-cert-<backend>` and `<mesh>.ca-builtin-key-<backend>` Secrets owned by their Mesh. The upgrade does not delete them: they stay, unread by any code, and are still synced global→zone by KDS. Leaving them is deliberate — the CA private key is unrecoverable once removed, and a dead row costs nothing.

**Action required**

None. Delete the orphaned Secrets yourself once you are certain no downstream tooling needs the CA material; they are removed automatically when their Mesh is deleted.

### Control plane TLS certificates are reloaded without a restart

Every control plane server that reads its certificate from disk (API server HTTPS, dp-server, global KDS, MADS, diagnostics) now picks up a rotation performed by an external tool without restarting `kuma-cp`. Nothing has to be configured for this, and no action is required to keep the previous behaviour, which was to serve the certificate loaded at startup until the process was restarted.

**Rotating the trust anchor is still not transparent to proxies**

A rotation is transparent only when the new certificate is issued by the CA the proxies already trust. It is not transparent when the rotation replaces the trust anchor itself, which is what happens with the self-signed certificate the control plane generates on its first run: `kuma-dp` receives that certificate as Envoy's trusted CA when it bootstraps, so a proxy that reconnects after the rotation validates the new certificate against the old one and fails until it bootstraps again. Established xDS streams are unaffected, because the certificate is only verified during the handshake.

**Action required**

If you rotate control plane certificates, issue them from a CA and point proxies at that CA with `kuma-dp run --ca-cert-file=/path/ca.pem` (or `KUMA_CONTROL_PLANE_CA_CERT_FILE`), then rotate leaves under it. This is required for a certificate issued by cert-manager or Vault in any case: such a leaf has `CA:FALSE`, and the control plane rejects a bootstrap request for it with `NotCA` unless the proxy supplies the CA itself.

### `mtls.backends[].mode: PERMISSIVE` no longer makes inbounds permissive

`MeshTLS` is now the only thing that decides whether an inbound accepts plaintext. The inbound listener is built once, and the mode is resolved from the `MeshTLS` policy alone — the `mode` of the enabled `Mesh` CA backend is no longer consulted. A mesh that sets `mtls.backends[].mode: PERMISSIVE` and has no `MeshTLS` policy now gets `Strict` inbounds, and plaintext traffic to those inbounds is rejected.

**Action required**

Before upgrading, author a `MeshTLS` policy with `default.mode: Permissive` for every mesh that relies on `mtls.backends[].mode: PERMISSIVE`:

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshTLS
metadata:
  name: permissive
  namespace: kuma-system
  labels:
    kuma.io/mesh: default
spec:
  rules:
    - default:
        mode: Permissive
```

Meshes on `mtls.backends[].mode: STRICT`, meshes with mTLS disabled, and meshes that already have a `MeshTLS` policy are unaffected. Note that a mesh using `MeshIdentity` already resolved to `Strict` without a `MeshTLS` policy.

### Zone Token `ingress` and `egress` scopes removed

`kumactl generate zone-token` no longer accepts `--scope ingress` or `--scope egress`, and the `POST /tokens/zone` endpoint no longer requires a `scope`. A zone proxy is an ordinary `Dataplane` and authenticates with a dataplane token, so these scopes identified components that no longer exist. Kuma itself defines no zone token scopes now; the token carries only the zone name unless a distribution registers its own.

**Action required**

Drop `--scope ingress`/`--scope egress` from any script calling `kumactl generate zone-token`. Zone proxies need a dataplane token — generate one with `kumactl generate dataplane-token`.

### `MeshExternalService` clusters require a `MeshIdentity`

A client proxy without a workload identity no longer gets a cluster for a `MeshExternalService`. It previously got one that addressed the zone egress by the legacy `zone-egress` service SNI, but such traffic could never be served: a zone egress only generates its egress listener when it has a workload identity, and that listener matches filter chains on the KRI SNI only, so the legacy SNI matched nothing.

The same SNI change makes the upgrade one-way across zones: once a zone is upgraded, its client proxies send real-resource traffic to zone proxies by the KRI-derived SNI. A zone that is still on the previous version cannot terminate that traffic because its zone proxy listeners still match the legacy hashed SNI only. In practice, an upgraded zone cannot send `MeshService`, `MeshExternalService`, or `MeshMultiZoneService` traffic to a previous-version zone until that destination zone is upgraded too.

**Action required**

Create a `MeshIdentity` that matches the client proxies in any mesh that uses `MeshExternalService`. Without one, only the cluster is dropped — the outbound listener and its endpoints are still generated, so requests now fail locally in the client's Envoy with a 503 (`cluster_not_found`) instead of being reset by the zone egress. Keeping the listener is deliberate: removing it would let the request fall through to the transparent proxy passthrough and reach the external service directly, bypassing the egress.

### `ZoneIngress` and `ZoneEgress` resources removed

The standalone `ZoneIngress`, `ZoneEgress`, `ZoneIngressInsight` and `ZoneEgressInsight` resources are gone, together with their CRDs (`zoneingresses.kuma.io`, `zoneegresses.kuma.io`, `zoneingressinsights.kuma.io`, `zoneegressinsights.kuma.io`), their REST endpoints, their `kumactl get`/`kumactl inspect` subcommands and their KDS sync. A zone proxy is an ordinary `Dataplane` carrying zone proxy listeners, and its health is reported through `DataplaneInsight`.

The control plane ClusterRole no longer grants access to these four CRDs, and the validating webhook no longer intercepts them.

`globalInsight.zones.zoneIngresses` and `globalInsight.zones.zoneEgresses` are still present in the API response but report `0` until zone proxy counts are surfaced from `MeshInsight`.

**Action required**

Delete any remaining `ZoneIngress`/`ZoneEgress` resources before upgrading. On Kubernetes, delete the four CRDs after upgrading — Helm does not remove CRDs on upgrade, so the stale objects would otherwise stay in etcd:

```sh
kubectl delete crd zoneingresses.kuma.io zoneegresses.kuma.io zoneingressinsights.kuma.io zoneegressinsights.kuma.io
```

On Universal, `kumactl delete zone-ingress <name>` and `kumactl delete zone-egress <name>` are no longer available; remove the rows from the resource store directly if any are left.

### `experimental.ingressTagFilters` removed

The `experimental.ingressTagFilters` configuration field and its `KUMA_EXPERIMENTAL_INGRESS_TAG_FILTERS` environment variable are removed. They filtered tags out of `ZoneIngress.availableServices`, which no longer exists. Config loading is non-strict, so a leftover value is silently ignored rather than rejected.

**Action required**

Remove `experimental.ingressTagFilters` from `kuma-cp.yaml` and `KUMA_EXPERIMENTAL_INGRESS_TAG_FILTERS` from control plane deployments and Helm values.

### `multizone.zone.ingressUpdateInterval` removed

The `multizone.zone.ingressUpdateInterval` configuration field and its `KUMA_MULTIZONE_ZONE_INGRESS_UPDATE_INTERVAL` environment variable are removed. Nothing has read them since the control plane stopped maintaining `ZoneIngress.availableServices`. Config loading is non-strict, so a leftover value is silently ignored rather than rejected.

**Action required**

Remove `multizone.zone.ingressUpdateInterval` from `kuma-cp.yaml` and `KUMA_MULTIZONE_ZONE_INGRESS_UPDATE_INTERVAL` from control plane deployments and Helm values.

### `kuma.io/ingress` and `kuma.io/egress` Pod annotations removed

The Pod controller no longer turns a Pod annotated with `kuma.io/ingress` or `kuma.io/egress` into a standalone `ZoneIngress`/`ZoneEgress`. Zone proxies are ordinary `Dataplane` resources carrying zone proxy listeners, built from a `Service` labelled `k8s.kuma.io/zone-proxy-type`. The Helm chart stopped emitting these annotations in 3.0, so only hand-rolled Pod manifests are affected. The `kuma.io/ingress-public-address` and `kuma.io/ingress-public-port` annotations are removed with them — the public address now comes from `MeshZoneAddress`.

**Action required**

Replace any hand-rolled ingress or egress Pod with the zone proxy Deployment shipped by the chart (`mesh-zoneproxy-*`), or label its `Service` with `k8s.kuma.io/zone-proxy-type`. A Pod that keeps the old annotations is reconciled as a regular `Dataplane` if it has a sidecar, and ignored otherwise — nothing rejects the annotation, so the change is silent.
### Zone proxy garbage collection folded into the Dataplane collector

Universal zone proxies are ordinary `Dataplane` resources carrying zone proxy listeners, so they are collected by the existing Dataplane GC. The separate collector for `ZoneIngress`/`ZoneEgress` has been removed, along with the `runtime.universal.zoneResourceCleanupAge` configuration field and its `KUMA_RUNTIME_UNIVERSAL_ZONE_RESOURCE_CLEANUP_AGE` environment variable. Offline zone proxies are now cleaned up after `runtime.universal.dataplaneCleanupAge` (default 72h) instead.

The `component_zone_gc` histogram is no longer exported by the control plane. `component_dp_gc` covers zone proxies as well.

**Action required**

Remove `runtime.universal.zoneResourceCleanupAge` from `kuma-cp.yaml` and `KUMA_RUNTIME_UNIVERSAL_ZONE_RESOURCE_CLEANUP_AGE` from control plane deployments and Helm values. Both are silently ignored after upgrading rather than rejected, so a stale value will not fail startup but will have no effect. If you relied on a zone-specific cleanup age, set `runtime.universal.dataplaneCleanupAge` to that value — it now applies to all Dataplanes, zone proxies included. Update any dashboard or alert that queries `component_zone_gc` to use `component_dp_gc`.

### `kuma.io/protocol` inbound tag no longer sets the protocol

The protocol of a `Dataplane` inbound is now read only from `networking.inbound[].protocol`. The `kuma.io/protocol` tag stays a regular tag — policies keep matching on it — but it is no longer used as a fallback when the field is unset. This only affects Universal: on Kubernetes the field is always derived from the `Service` port during conversion.

**Action required**

Set `networking.inbound[].protocol` on every Universal `Dataplane` that currently declares its protocol only through the `kuma.io/protocol` tag, before upgrading.

**Warning**: an inbound without `protocol` is treated as an unknown protocol and served as plain TCP. Its listener loses the `kuma.io/protocol` tag and the L7 filters that depend on it — HTTP access log fields, `MeshTimeout` HTTP timeouts, `MeshFaultInjection`, `MeshRateLimit` HTTP limits, HTTP-aware routing — and the endpoint stops advertising a protocol, so `MeshService`-level protocol inference falls back to TCP as well. Nothing rejects the resource, so the change is silent.

### `KUMA_MESH_TRAFFIC_PERMISSION_DISABLE_CLIQUES_ALGORITHM` removed

The `KUMA_MESH_TRAFFIC_PERMISSION_DISABLE_CLIQUES_ALGORITHM` environment variable
has been removed. MeshTrafficPermission rule generation now always uses the
cliques-based grouping algorithm.

**Action required**

Remove `KUMA_MESH_TRAFFIC_PERMISSION_DISABLE_CLIQUES_ALGORITHM` from control
plane deployments, Helm values, and any other runtime configuration before or
after upgrading. Leaving it set no longer has any effect in Kuma 3.0.0.

### Unified resource naming opt-out removed

The `KUMA_DATAPLANE_RUNTIME_UNIFIED_RESOURCE_NAMING_ENABLED` `kuma-dp`
environment variable, the `KUMA_RUNTIME_KUBERNETES_INJECTOR_UNIFIED_RESOURCE_NAMING_ENABLED`
control plane environment variable / `runtime.kubernetes.injector.unifiedResourceNamingEnabled`
`kuma-cp.yaml` key, and the `dataPlane.features.unifiedResourceNaming` Helm value have
all been removed. Unified Envoy resource and stat naming is now always enabled,
and the sidecar injector no longer stamps the corresponding env var onto injected
`kuma-sidecar` containers.

**Action required**

If you previously set any of these to `false` to opt out, unified naming is now
always on and the control plane generates unified Envoy resource and stat names
regardless of any leftover config. Update automation, dashboards, or alerting that
depend on the legacy names before upgrading. The leftover config values themselves
are silently ignored rather than rejected: `kuma-cp` does not use strict YAML
parsing outside of tests, `envconfig` ignores unknown environment variables, and
Helm accepts unknown `--set` paths.

### MADS restricted to universal deployment mode

The Monitoring Assignment Discovery Service (MADS) server no longer starts on
Kubernetes control planes, regardless of `KUMA_MONITORING_ASSIGNMENT_SERVER_ENABLED`
or `controlPlane.madsServer.enabled`. It remains fully supported in universal
deployment mode, including universal-on-Kubernetes
(`controlPlane.environment: universal`). The Helm chart no longer renders the
`mads-server` Service port (5676) when `controlPlane.environment` is
`kubernetes`.

**Action required**

Kubernetes users relying on MADS must migrate to `MeshMetric` with Prometheus
Kubernetes service discovery before upgrading. `controlPlane.madsServer.enabled`
now only applies when `controlPlane.environment` is `universal`.

### `advertisedAddress` removed from `Dataplane` networking

The `networking.advertisedAddress` field has been removed from the `Dataplane` resource. Proxies behind NAT or a private network (e.g. Docker) that relied on it to advertise a routable address to other proxies must now be reachable directly via `networking.address`.

**Action required**

Ensure every Universal `Dataplane` is reachable by other proxies on `networking.address` before upgrading.

**Warning**: `networking.advertisedAddress` is silently dropped on deserialization — protos are unmarshalled with `AllowUnknownFields`, so the field is simply ignored rather than rejected. Dataplanes still submitting it will fall back to `networking.address` for xDS endpoints, Envoy admin mTLS SANs, and `kumactl get dataplanes` output, which may break connectivity for proxies that are not reachable on `networking.address`.

### Real-resource policy selection now uses `labels` only

Policies that select real resources through `spec.targetRef` or `spec.to[].targetRef` now resolve those targets by `labels` only. This applies to `Dataplane`, `MeshService`, `MeshExternalService`, `MeshMultiZoneService`, and `MeshHTTPRoute`.

**Action required**

Migrate any policy that still selects those resources by `name` and/or `namespace` to use `labels` instead before upgrading. `sectionName` remains supported for `Dataplane` inbound selection and `MeshService` port selection.

### A `MeshHTTPRoute` rule whose backendRefs all fail to resolve answers 500

A rule that declares `backendRefs` and resolves none of them no longer falls back to the destination service. The rule now serves `500` to every request it matches, which is what the Gateway API requires of an invalid backendRef. A rule that resolves at least one of its backendRefs keeps routing to those backends, and a rule with a `RequestRedirect` filter still redirects.

This covers every `MeshHTTPRoute`, not only the ones the Gateway API translation generates. A route pointing at a resource that does not exist yet — a `MeshService` that KDS has not synced to the zone, for example — fails its matching requests instead of sending them to the destination service, until the reference resolves.

**Action required**

None, as long as every `backendRefs` entry names a resource that exists. Check the routes that reference a resource from another zone or another namespace before upgrading: those are the ones whose masked misconfiguration becomes visible traffic loss.

### A `MeshHTTPRoute` `backendRef` naming a port the destination lacks now fails closed

A `backendRef` whose destination exists but does not carry the referenced
`sectionName`/port no longer counts as resolved. Previously it silently
dropped out of the split: a missing `Service` port fell through to the
parent service on a different port than the one requested, and a missing
`MeshService` port fell through to another port of that same `MeshService`.
Both now count against `AllBackendRefsUnresolved` the same way a
nonexistent destination does, so a rule whose backendRefs all name a
missing port answers `500` instead of routing traffic to a port nobody
asked for.

The Gateway API `HTTPRoute` translation reports `ResolvedRefs=False` with
reason `BackendNotFound` for a `Service` backend missing the requested
port; it already did for `MeshService`. Because a `MeshService`'s `Spec.Ports`
can be briefly empty while it converges, a route that references one of its
ports can fail matching requests during that window.

**Action required**

None, as long as every `backendRefs` entry names a port the destination
actually has. Check routes with an explicit port or `sectionName` before
upgrading: those are the ones that were silently reaching the wrong
destination.

### Gateway API cross-namespace `backendRefs` now require a `ReferenceGrant`

The Gateway API `HTTPRoute` translation now enforces `ReferenceGrant` for any
backend reference that points across namespaces, for both core `Service`
backends and Kuma `MeshService` backends. A cross-namespace backend without a
matching grant is no longer programmed into the generated `MeshHTTPRoute`; the
route reports `ResolvedRefs=False` with reason `RefNotPermitted` instead.

To evaluate those grants, the control plane ClusterRole now includes `get`,
`list`, and `watch` on `gateway.networking.k8s.io/referencegrants`.

**Action required**

Create a `ReferenceGrant` in the backend namespace for every cross-namespace
`HTTPRoute` backend that should remain valid after upgrading. If you manage
RBAC outside Helm (for example via GitOps or manual manifests), also add
`get`, `list`, and `watch` on `referencegrants` in the
`gateway.networking.k8s.io` API group to the control plane ClusterRole.

### `MeshService.spec.identities` now accepts SPIFFE IDs only

`MeshService.spec.identities[].type` no longer accepts `ServiceTag`. `MeshService.spec.identities` now publishes SPIFFE IDs only, while service-tag-based routing keeps using the `kuma.io/service` label or the MeshService resource name as its fallback naming signal.

**Action required**

Update any MeshService manifest that still declares a `ServiceTag` identity to use `SpiffeID` entries only before upgrading. A persisted `ServiceTag` entry is rejected by the updated schema once the new CRD or API validation is in place.

### Transparent proxy configured only through the ConfigMap

The legacy annotation-based transparent proxy injection path has been removed.
The sidecar injector now always builds the transparent proxy configuration from
the ConfigMap in the `kuma-system` namespace (merged with pod annotations) and
delivers it through the `traffic.kuma.io/transparent-proxy-config` annotation and
mounted files. This was previously an opt-in feature gated by
`transparentProxy.configMap.enabled`.

**Action required**

No action is required for Helm or `kumactl` installs — the control plane always
creates the base ConfigMap and points the injector at it. The
`transparentProxy.configMap.enabled` Helm value has been removed; remove it from
any custom values files (leaving it set is harmless but has no effect).

The per-pod `kuma.io/transparent-proxying-*` annotations are no longer produced
by injection. Pods are reconfigured automatically on their next restart after the
upgrade.

**Warning**: this format is not understood by data plane proxies older than the
control plane. To downgrade, first roll back the control plane and then restart
all workloads so their init and sidecar containers fall back to the previous
configuration.

### `from` removed from `MeshTLS`

The deprecated `from` field has been removed from the `MeshTLS` policy. Use the `rules` field instead.

**Action required**

Migrate any `MeshTLS` resources using `from` to use `rules` before upgrading.

**Warning**: Un-migrated `from` configurations are silently ignored after upgrade — the `from` field no longer exists in the schema and the data is discarded during deserialization. The impact depends on your CA backend configuration:
- **CA backend mode `PERMISSIVE`**: workloads fall back to permissive TLS when you intended strict
- **CA backend mode `STRICT` or workload has identity**: workloads default to strict mTLS, potentially breaking connectivity if the `from` policy was intentionally permissive

Before upgrading, audit your cluster for affected resources:
```bash
kubectl get meshtls -A -o yaml | grep -B5 'from:'
```

```yaml
# Before (deprecated)
spec:
  targetRef:
    kind: Mesh
  from:
    - targetRef:
        kind: Mesh
      default:
        mode: Strict

# After
spec:
  targetRef:
    kind: Mesh
  rules:
    - default:
        mode: Strict
```

### `healthyPanicThreshold` removed from `MeshHealthCheck`

The deprecated `to[].default.healthyPanicThreshold` field has been removed from the `MeshHealthCheck` policy. Use `to[].default.outlierDetection.healthyPanicThreshold` on the `MeshCircuitBreaker` policy instead.

**Action required**

Migrate any `MeshHealthCheck` resources using `to[].default.healthyPanicThreshold` to a `MeshCircuitBreaker` policy with `to[].default.outlierDetection.healthyPanicThreshold` before upgrading.

**Warning**: Un-migrated `healthyPanicThreshold` settings are silently dropped after upgrade — the field no longer exists in the schema, so it is pruned by CRD validation on Kubernetes and discarded during deserialization on Universal. Affected clusters fall back to Envoy's default panic threshold of 50%.

### `Dataplane` outbounds must use `backendRef`

The `tags` field on `networking.outbound[]` has been removed from the `Dataplane` schema. An outbound is now defined exclusively by `backendRef`, pointing at a `MeshService`, `MeshExternalService`, or `MeshMultiZoneService`. Creating or updating a `Dataplane` whose outbound has no `backendRef` is rejected with `backendRef: must be defined`.

The xDS generation path that consumed `kuma.io/service`-tagged outbounds is gone with it. Policies no longer match those outbounds through the deprecated subset rules (`spec.to[].targetRef` resolved against a `kuma.io/service` value); they match the referenced resource instead. Locality awareness for cross-zone traffic through ZoneEgress is no longer applied to outbound clusters on this path.

**Action required**

Rewrite every `Dataplane` that declares tag-based outbounds before upgrading. On Kubernetes, outbounds are generated by the control plane and need no change.

```yaml
# Before (removed)
networking:
  outbound:
    - port: 3000
      tags:
        kuma.io/service: postgres

# After
networking:
  outbound:
    - port: 3000
      backendRef:
        kind: MeshService
        name: postgres
        port: 5432
```

**Warning**: Stored `Dataplane` resources that still carry `tags` on an outbound keep being served, but the tags are discarded during deserialization and no outbound listener is generated for them, so the workload loses connectivity to that destination.

### Route `backendRef` no longer accepts `MeshServiceSubset`

`MeshHTTPRoute` and `MeshTCPRoute` validators now accept only `MeshService`, `MeshExternalService`, and `MeshMultiZoneService` in `backendRefs[]`; the `MeshHTTPRoute` `RequestMirror` filter's `backendRef` is limited to the same kinds. A route using `kind: MeshServiceSubset` is rejected with `value 'MeshServiceSubset' is not supported`.

The xDS path that turned such a ref into a `kuma.io/service` cluster is gone with it, and a `MeshService` ref selected by `kuma.io/display-name` alone no longer falls back to one when it fails to resolve. A backendRef that resolves to no real resource now produces no split at all, rather than a split towards a cluster that is never generated.

**Action required**

Repoint every route that splits or mirrors to a `MeshServiceSubset` at the real resource before upgrading. Selecting a subset of endpoints by tag has no equivalent: split the destination into separate `MeshService` resources if you need it.

```yaml
# Before (rejected)
backendRefs:
  - kind: MeshServiceSubset
    tags:
      kuma.io/service: payments
      version: v1

# After
backendRefs:
  - kind: MeshService
    labels:
      kuma.io/display-name: payments
    port: 8080
```

**Warning**: Stored routes that still carry a `MeshServiceSubset` backendRef keep being served, but the current control plane no longer resolves that kind into a real backend. During xDS generation the backendRef is therefore treated as unresolved, so traffic matching those rules loses its destination.

The `tags` field is also gone from the `backendRef` schema entirely, not just disallowed for `MeshServiceSubset`. On Kubernetes, a `tags` key on a `backendRef` is pruned by the CRD structural schema before it reaches the control plane. On Universal, it is silently ignored when the resource is unmarshalled.

### Legacy `ExternalService` resource removed

The legacy `ExternalService` resource has been removed. Its CRD, API
definition, and validating webhook no longer exist, and the control plane
RBAC and admission webhooks no longer reference `externalservices`.

**Action required**

Migrate any remaining `ExternalService` resources to `MeshExternalService`
before upgrading. Applying an `ExternalService` after the upgrade will fail
because the CRD is gone.

### Global control plane on Kubernetes is no longer supported

A Kubernetes-native Global control plane is no longer supported. `kuma-cp` now
rejects `mode=global` with `environment=kubernetes`, and it also rejects
`mode=global` with `store.type=kubernetes`. A Global control plane must run
with `environment=universal` backed by a non-Kubernetes store such as
PostgreSQL, even if `kuma-cp` itself is deployed on Kubernetes. The Helm chart
no longer renders the `Service`/config needed for the old Kubernetes-native
setup. Zone control planes on Kubernetes (`mode` `zone`) are unaffected.

**Action required**

If you currently run the Global control plane on Kubernetes, migrate it to
Universal (non-Kubernetes) infrastructure before upgrading: deploy `kuma-cp`
in `global` mode on Universal, backed by PostgreSQL, and keep your Kubernetes
clusters as Zone control planes connecting to that Global control plane over
KDS. Kubernetes clusters running `zone` mode require no changes.

### `standalone` mode removed

The deprecated `standalone` control plane mode has been removed. `KUMA_MODE`/
`controlPlane.mode` no longer accepts `standalone`: `kuma-cp` fails config
validation at startup, and the Helm chart fails at template time.

**Action required**

Rename `standalone` to `zone` in `KUMA_MODE`, `controlPlane.mode`, and any
other runtime configuration before upgrading. `standalone` and `zone` were
already behaviourally identical, so no other changes are required.

### `meshServices` removed from the `Mesh` schema

The `meshServices` field (and its `mode` enum) has been removed from the
`Mesh` resource spec. Unified resource naming is now unconditional,
regardless of what the mesh's former `meshServices.mode` was set to.

**Action required**

None. A `Mesh` spec that still sets `meshServices` continues to apply
successfully; the field is silently ignored by the control plane.

### `routing.zoneEgress` removed from the `Mesh` schema

The `routing.zoneEgress` boolean has been removed from the `Mesh` resource
spec. Cross-zone and `MeshExternalService` traffic now uses ZoneEgress based on
actual zone egress topology plus mTLS, rather than an explicit mesh-level
toggle.

**Action required**

Remove any `routing.zoneEgress` entries from stored `Mesh` resources and
manifests before upgrading. After the upgrade, that field no longer exists in
the schema and is ignored during deserialization.

### `routing.defaultForbidMeshExternalServiceAccess` removed from the `Mesh` schema

The `routing.defaultForbidMeshExternalServiceAccess` boolean has been removed
from the `Mesh` resource spec. Zone egress no longer reads a mesh-level toggle
to block `MeshExternalService` traffic by default.

**Action required**

Remove any `routing.defaultForbidMeshExternalServiceAccess` entries from stored
`Mesh` resources and manifests before upgrading. After the upgrade, that field
no longer exists in the schema and is ignored during deserialization. Use
`MeshTrafficPermission` to express explicit access control for
`MeshExternalService` traffic instead of relying on the removed mesh-wide
default-deny toggle.

### MeshService mode no longer disables zone proxy listeners, inspect endpoints, or MeshIdentity initialization

The control plane now generates mesh-scoped zone proxy listeners and serves
Dataplane inspect `_layout` and policy endpoints regardless of
`meshServices.mode`.

The Kubernetes warning event reason `ZoneProxyListenersSkipped`, previously
emitted when a zone proxy Service was ignored outside `Exclusive` mode, has been
removed because these listeners are no longer skipped.

The `MeshIdentity` status reason `MeshServicesDisabled`, previously reported
when identity initialization was skipped outside `Exclusive` mode, has also
been removed because MeshIdentity initialization now proceeds in every
MeshService mode.

**Action required**

If you alert on `ZoneProxyListenersSkipped`, remove or update that alert before
upgrading. Zone proxy Pods that previously depended on this skipped-listener
behavior will now receive listener configuration in every MeshService mode.
Also update any automation that expected the `MeshServicesDisabled`
`MeshIdentity` status reason or treated inspect `_layout` as unavailable
outside `Exclusive` mode.

### Standalone `ZoneIngress`/`ZoneEgress` proxies are no longer supported

The control plane no longer serves data plane proxies started as standalone zone
proxies, i.e. `kuma-dp run --proxy-type=ingress|egress` on Universal and the
`ingress.enabled` / `egress.enabled` Helm deployments on Kubernetes. Cross-zone
traffic is now served by mesh-scoped zone proxies, which are regular `Dataplane`
resources.

`dataplane` is the only accepted value of `--proxy-type` (and of the
`KUMA_DATAPLANE_PROXY_TYPE` environment variable); `kuma-dp` exits with
`.ProxyType "ingress" is not supported` on startup. The control plane no longer
generates a bootstrap for those proxy types, no longer registers or deregisters
their `ZoneIngress`/`ZoneEgress` resources, and no longer writes
`ZoneIngressInsight`/`ZoneEgressInsight` from an xDS stream. A pre-upgrade zone
proxy that reconnects has its xDS stream rejected with `unsupported proxy type
"ingress"`. `kumactl generate dataplane-token --proxy-type ingress|egress` is
rejected, and tokens previously issued with a `type: ingress|egress` claim can no
longer be used.

**Action required**

Migrate to mesh-scoped zone proxies **before** upgrading the zone control plane.
On Kubernetes deploy them through `meshes[].ingress.enabled` /
`meshes[].egress.enabled`. On Universal, replace
`kuma-dp run --proxy-type=ingress|egress` with a regular `Dataplane` that
declares `networking.listeners` of type `ZoneIngress`/`ZoneEgress`, and reissue
its token without `--proxy-type`. Upgrading the control plane first blackholes
cross-zone traffic as soon as the legacy zone proxy Pods restart.

### Legacy `ingress`/`egress` Helm values and `kumactl install` flags removed

The chart no longer renders standalone `ZoneIngress`/`ZoneEgress` Deployments.
The top-level `ingress` and `egress` value blocks are gone, along with the
Deployment, Service, HorizontalPodAutoscaler, PodDisruptionBudget and RBAC
templates they drove. `kumactl install control-plane` lost the matching flags:
`--ingress-enabled`, `--ingress-drain-time`, `--ingress-use-node-port`,
`--ingress-node-selector`, `--egress-enabled`, `--egress-drain-time`,
`--egress-service-type` and `--egress-node-selector`. `controlPlane.ingress.*`
is unrelated and still configures the Kubernetes Ingress for the control plane
GUI and API.

Helm does not reject unknown values, so an upgrade with `ingress.enabled=true`
left in your values file succeeds without any warning. Because the templates are
gone, the upgrade deletes the legacy zone proxy Deployment, Service,
HorizontalPodAutoscaler, PodDisruptionBudget and RBAC objects that the previous
release owned, which drops all cross-zone traffic still flowing through them.

**Action required**

Complete the migration described in the previous section, then drop the
top-level `ingress` and `egress` blocks from your values files and the removed
flags from any `kumactl install control-plane` invocation.

Most legacy settings map onto `meshes[].ingress` / `meshes[].egress`. These have
no equivalent there: `podAnnotations`, `annotations`, `logLevel`, `drainTime`,
`lifecycle`, `livenessProbe`, `readinessProbe`, `startupProbe`, `dns.policy`,
`dns.config`, `service.enabled` and `service.nodePort`. Drain time and probes
are now control-plane-wide sidecar injector settings.

### Standalone zone proxy inspect endpoints and `kumactl inspect` commands removed

The Envoy admin inspect endpoints for standalone zone proxies are gone:
`GET /zoneingresses/{name}/{xds,stats,clusters}` and
`GET /zoneegresses/{name}/{xds,stats,clusters}` now return 404. So do the
pre-2.6 overview aliases `GET /zoneingresses+insights[/{name}]` and
`GET /zoneegressoverviews[/{name}]`, which have been redundant with
`/zoneingresses[/{name}]/_overview` and `/zoneegresses[/{name}]/_overview` since
2.6. Reading and listing the `ZoneIngress`/`ZoneEgress` resources themselves is
gone as well — the resources no longer exist, see the section above.

`kumactl inspect` loses `zoneingress`, `zoneingresses` (alias `zone-ingresses`),
`zoneegress` and `zoneegresses`.

**Action required**

Mesh-scoped zone proxies are regular `Dataplane` resources, so inspect them with
`kumactl inspect dataplane <name> --mesh <mesh>` and the
`/meshes/{mesh}/dataplanes/{name}/{xds,stats,clusters}` endpoints. Update any
automation or dashboard that still calls the removed paths.

The GUI's zone ingress and zone egress XDS, stats and clusters tabs depend on the
removed endpoints and stop working until the bundled GUI is updated. Overview and
resource views are unaffected.

### ServiceInsight, MeshInsight, and inspect `_rules` no longer report kuma.io/service based data

With `meshServices.mode` always `Exclusive`, `kuma.io/service`-tagged services and
legacy `ExternalService` resources are represented by `MeshService` and
`MeshExternalService` instead, so the control plane no longer computes their
legacy statistics:

- `ServiceInsight` is no longer computed at all. The control plane never writes
  the resource, and it deletes any `ServiceInsight` left over by the previous
  version on every insight resync (every
  `KUMA_METRICS_MESH_FULL_RESYNC_INTERVAL`, 20s by default). During a rolling
  upgrade, an old replica can still write the resource between resync ticks on
  an upgraded replica, so the legacy REST endpoints may briefly serve stale
  data until the next resync deletes it again. Once every replica is upgraded
  and a resync interval has elapsed, `GET /meshes/{mesh}/service-insights`
  returns an empty list and `GET /meshes/{mesh}/service-insights/{name}`
  returns `404`. This also covers delegated gateways, which used to be the
  last services reported there, along with their per-service `zones` list.
  The GUI pages backed by that endpoint list nothing. `kumactl inspect
  services` is removed; use `kumactl get meshservices` instead.
- `MeshInsight.services` is removed from the API. Field number 6 is reserved
  and will not be reused. Mesh-scoped service status now comes from
  `MeshService` and `MeshExternalService`; the aggregated
  `internal`/`external`/`gatewayDelegated` counts live under `services` in the
  global insight endpoint, not in `MeshInsight`.
- The Dataplane/MeshGateway inspect `_rules` endpoint no longer returns the
  legacy `toRules` and `fromRules` fields on each rule entry; both fields are
  removed from the response. `toResourceRules` and `inboundRules` are
  unaffected.
- The legacy `GET /meshes/{mesh}/meshservices/{name}/_resources/dataplanes`
  endpoint is removed. Use `GET /meshes/{mesh}/meshservices/{name}/_dataplanes`
  instead.

**Action required**

Update any automation or dashboards that read the legacy `ServiceInsight`
resource or `/meshes/{mesh}/service-insights` endpoints, `MeshInsight.services`,
or the `_rules` `toRules`/`fromRules` fields. Use `MeshService`/
`MeshExternalService` status for per-resource service state, `_rules`
`toResourceRules`/`inboundRules` for inspect output, and the global insight
endpoint's `services` object for aggregated `internal`/`external`/
`gatewayDelegated` counts. For delegated gateways, which are never turned into
a `MeshService`, use the `Dataplane`/`DataplaneOverview` endpoints filtered by
gateway type when you need per-gateway detail.

### Zone proxies authenticate with a dataplane token

A zone proxy is now a `Dataplane` with zone proxy listeners, so the DP server
authenticates it exactly like any other data plane proxy. The separate zone
proxy authenticator is gone: every proxy is authenticated with the method
configured under `dpServer.authn.dpProxy` (`serviceAccountToken` on Kubernetes,
`dpToken` on Universal), and zone tokens are no longer validated.

`dpServer.authn.zoneProxy.type` and
`dpServer.authn.zoneProxy.zoneToken.validator` no longer affect anything.

**Action required**

On Universal, issue a dataplane token for each zone proxy
(`kumactl generate dataplane-token --mesh <mesh> --name <zone-proxy-dp>`)
instead of a zone token, and pass it to `kuma-dp` with `--dataplane-token-file`.
Tokens generated with `kumactl generate zone-token` are no longer accepted by
the DP server. If you relied on `dpServer.authn.zoneProxy.type: none` to let
zone proxies connect without a token while data plane proxies used `dpToken`,
zone proxies now need a dataplane token too.

### `dataplaneTags` removed from the `MeshService` selector

`spec.selector.dataplaneTags` matched data plane proxies by their inbound tags.
The field has been removed; `MeshService` selects proxies by
`spec.selector.dataplaneRef` or `spec.selector.dataplaneLabels` only.

**Warning**: un-migrated selectors are silently dropped during deserialization.
An affected `MeshService` keeps its name and ports but matches zero data plane
proxies, so it stops producing endpoints and its status goes `Unavailable`.
The control plane returns a warning for any `MeshService` left without a
selector, but only when the resource is next created or updated.

**Action required**

Migrate any `MeshService` using `dataplaneTags` to `dataplaneLabels` before
upgrading. Audit with:

```bash
kubectl get meshservices -A -o yaml | grep -B5 'dataplaneTags:'
kumactl get meshservices -o yaml --all-meshes | grep -B5 'dataplaneTags:'
```

```yaml
# Before (removed)
spec:
  selector:
    dataplaneTags:
      app: redis

# After
spec:
  selector:
    dataplaneLabels:
      matchLabels:
        app: redis
```

Inbound tags are no longer an identity source, so the replacement labels must
exist on the `Dataplane` resource itself. On Kubernetes those come from the Pod
labels; on Universal, set them under `labels` in the `Dataplane` resource.

### CoreDNS removed from the data plane

The bundled CoreDNS binary has been removed. The data plane now always uses the in-process embedded DNS proxy (previously the default on Kubernetes and opt-in on Universal). CoreDNS and the Envoy DNS filter are no longer used, and the `coredns` binary is no longer shipped in the release tarball or the `kuma-dp` image.

The following configuration has been removed:

- Data plane (`kuma-dp`): `dns.coreDnsPort`, `dns.envoyDnsPort`, `dns.coreDnsBinaryPath`, `dns.coreDnsConfigTemplatePath`, `dns.configDir`, `dns.prometheusPort` and `dns.coreDNSLogging`, together with the matching `KUMA_DNS_CORE_DNS_PORT`, `KUMA_DNS_ENVOY_DNS_PORT`, `KUMA_DNS_CORE_DNS_BINARY_PATH`, `KUMA_DNS_CORE_DNS_CONFIG_TEMPLATE_PATH`, `KUMA_DNS_CONFIG_DIR`, `KUMA_DNS_PROMETHEUS_PORT` and `KUMA_DNS_ENABLE_LOGGING` env vars, and the `--dns-envoy-port`, `--dns-coredns-port`, `--dns-coredns-path`, `--dns-coredns-config-template-path`, `--dns-server-config-dir`, `--dns-prometheus-port` and `--dns-enable-logging` flags. The DNS listen port is now configured solely via `dns.proxyPort` (`KUMA_DNS_PROXY_PORT`, `--dns-proxy-port`), defaulting to `15053`.
- Control plane: `bootstrapServer.params.corefileTemplatePath` (`KUMA_BOOTSTRAP_SERVER_PARAMS_COREFILE_TEMPLATE_PATH`), `runtime.kubernetes.injector.builtinDNS.experimentalProxy` (`KUMA_RUNTIME_KUBERNETES_INJECTOR_BUILTIN_DNS_EXPERIMENTAL_PROXY`) and `runtime.kubernetes.injector.builtinDNS.logging` (`KUMA_RUNTIME_KUBERNETES_INJECTOR_BUILTIN_DNS_LOGGING`).
- Helm: `dataPlane.dnsLogging`.
- The `kuma.io/builtin-dns-logging` pod annotation.

**Action required**

Because the control plane no longer emits the Envoy DNS filter configuration that older CoreDNS-based data planes rely on, upgrade the control plane and all data planes together. Remove any of the settings listed above from your control plane config, Helm values and `kuma-dp` invocations.

### eBPF transparent proxy removed

The experimental eBPF transparent proxy feature has been removed. This feature
used prebuilt merbridge binaries for traffic interception instead of iptables.

The following configuration no longer has any effect and is silently ignored:

- Control plane: `runtime.kubernetes.injector.ebpf.*` environment variables
  (`KUMA_RUNTIME_KUBERNETES_INJECTOR_EBPF_*`).
- Helm values: `experimental.ebpf.*`, `hooks.ebpfCleanup`,
  `cni.experimental.imageEbpf`, and `transparentProxy.configMap.config.ebpf`.
- Pod annotations: `kuma.io/transparent-proxying-ebpf*`.

The following CLI surface has been removed:

- CLI flags: `kumactl install transparent-proxy --ebpf-*` and
  `kumactl uninstall transparent-proxy --ebpf-*`.
- CLI command: `kumactl uninstall ebpf`.

**Action required**

If you are using eBPF transparent proxy (`experimental.ebpf.enabled=true`), you
must migrate to the iptables-based transparent proxy before upgrading. Remove
the eBPF configuration from your Helm values and pod annotations. The
iptables-based transparent proxy is the default and does not require additional
configuration.

If you have eBPF programs pinned on cluster nodes, clean them up **before**
upgrading by running `kumactl uninstall ebpf` with the pre-upgrade Kuma version.
After upgrading, this command is no longer available. If you skip this step,
residual BPF state remains on affected nodes: pinned programs under bpffs
(`/sys/fs/bpf`), cgroup connect hooks, and TC filter attachments. This state
continues intercepting traffic until manually removed. Consult the merbridge
documentation for manual cleanup procedures if needed.

### North-south Gateway API (built-in gateway) support removed

The control plane no longer reconciles the north-south Gateway API resources
`Gateway`, `GatewayClass`, and `ReferenceGrant` used to configure the
built-in gateway via the Kubernetes Gateway API. `HTTPRoute` is still
reconciled, but only for its east-west (GAMMA) use — `HTTPRoute`s whose
`parentRefs` target a `Service` and get translated to `MeshHTTPRoute`.
`HTTPRoute`s that target a `Gateway` are now ignored by the control plane.
Any status that Kuma already wrote for those `Gateway` parent references is
left unchanged, but it is no longer updated.

The built-in gateway itself (`MeshGateway`, `MeshGatewayInstance`) is
unaffected by this change. See "Built-in gateway Kubernetes controllers
removed" below for the separate removal of `MeshGatewayInstance` management.

**Action required**

- Before upgrading, delete any `Gateway` and `ReferenceGrant` resources you
  created for the built-in gateway; the control plane no longer manages them
  and leaves them untouched on disk.
- The `kuma` `GatewayClass` that the Helm chart used to install is no longer
  installed. On startup, the control plane removes the legacy
  `gateway-exists-finalizer.gateway.networking.k8s.io` finalizer from
  Kuma-managed `GatewayClass` objects so Helm-pruned objects can finish
  deleting. If you manage RBAC manually, grant the control plane `get`, `list`,
  `patch`, and `update` on `gatewayclasses` and `get`, `patch`, and `update`
  on `gatewayclasses/finalizers` during the upgrade, or remove that finalizer
  before upgrading.
- Any Kuma `Secret` that was copied from a Gateway API TLS `Secret` (name
  prefixed `gapi-`) is now orphaned and must be deleted manually.
- The default Helm-installed cluster RBAC no longer grants access to
  `gateways` and `referencegrants`. It still grants access to `httproutes`
  for the GAMMA path, plus narrow `gatewayclasses` access for the finalizer
  cleanup described above. If you manage RBAC manually, add the
  `gatewayclasses` permissions called out above for the upgrade window.
- The config field `runtime.kubernetes.supportGatewaySecretsInAllNamespaces`
  (env `KUMA_RUNTIME_KUBERNETES_SUPPORT_GATEWAY_SECRETS_IN_ALL_NAMESPACES`)
  and the Helm value `controlPlane.supportGatewaySecretsInAllNamespaces` have
  been removed. The control plane now always scopes its `Secret` watch to its
  own system namespace. Remove this key from your config file and Helm
  values — leaving it in place is harmless but has no effect.

### `kuma.io/gateway` is a boolean annotation

The `kuma.io/gateway` Pod annotation is now read the same way on both sides of the injection: as a boolean, accepting `enabled`, `true`, `yes` to mark a delegated gateway and `disabled`, `false`, `no` to opt out. Anything else is rejected with `annotation "kuma.io/gateway" has wrong value "<value>"`.

The `provided` value is gone. It has not worked since 2.10: the injector parses this annotation as a boolean, so a Pod annotated `kuma.io/gateway: provided` fails admission before any `Dataplane` is created. Its only effect on the `Dataplane` was a `kuma.io/service-name` gateway tag that nothing has read since `kuma.io/service` was removed.

Two values that used to be inconsistent now behave as the injector always intended. `kuma.io/gateway: "true"` was injected as a gateway but then failed conversion with `invalid delegated gateway type 'true'`; it now produces a gateway `Dataplane`. `kuma.io/gateway: disabled` was injected as a regular Pod but also failed conversion, so the Pod never got a `Dataplane` at all; it now produces a regular `Dataplane` with inbounds.

The consumers that keyed off the annotation being present rather than its value follow the same rule now. A Pod or `Service` annotated `kuma.io/gateway: disabled` gets a `MeshService` like any other workload instead of being skipped, a change of the annotation between enabled and disabled triggers a `MeshService` reconcile, and the injected sidecar of such a Pod gets the regular application probe proxy port instead of `0`, which previously left its probes pointing at a port nothing served.

**Action required**

Replace `kuma.io/gateway: provided` with `kuma.io/gateway: enabled` on any Pod, Deployment template, or Helm value that still sets it. Such a Pod is currently failing admission, so this is a fix rather than a regression, but the annotation has to change before the Pod can start.

### Built-in gateway Kubernetes controllers removed

The control plane no longer reconciles `MeshGatewayInstance` resources on
Kubernetes. It no longer creates or manages the `Service` and `Deployment`
generated for a `MeshGatewayInstance`, no longer converts `Pod`s annotated
`kuma.io/gateway: builtin` into a built-in gateway `Dataplane`, and no longer
runs the `MeshGatewayInstance` admission validator. `kumactl inspect
meshgateway` has been removed along with its client. Delegated gateways
(`kuma.io/gateway: enabled`) and the Gateway API `HTTPRoute` GAMMA path are
unaffected.

The `MeshGatewayInstance` CRD, its API types, and the `MeshGateway`/
`MeshGatewayRoute` resources themselves are not removed by this change.

**Action required**

- Existing `MeshGatewayInstance` resources become inert: the control plane
  stops reconciling them, so any `Service`, `Deployment`, and `BUILTIN`
  `Dataplane` it previously generated for them is not updated, recreated, or
  cleaned up automatically. If you still rely on a built-in gateway, migrate
  it to a delegated gateway (bring your own `Deployment`/`Service` fronting a
  Kuma-injected pod annotated `kuma.io/gateway: enabled`) before upgrading.
- Before upgrading, or as part of your migration, manually delete the
  `MeshGatewayInstance` resources you no longer need, along with the
  `Service`, `Deployment`, and `Dataplane` objects they previously generated
  (they are owned by the `MeshGatewayInstance`, so deleting it will cascade
  via owner references, but only for objects created before this upgrade).
- The default Helm-installed cluster RBAC no longer grants access to
  `meshgatewayinstances`, `meshgatewayinstances/status`, or
  `meshgatewayinstances/finalizers`, and the validating webhook configuration
  no longer includes `meshgatewayinstances`. If you manage RBAC manually,
  remove these permissions and webhook rules; if you keep them, they are
  harmless but unused.
- Rolling back a control plane upgraded past this version does not restore the
  removed gateway controller behavior automatically: the objects generated for
  any `MeshGatewayInstance` created
  or changed while running the new control plane are not retroactively
  reconciled by an older control plane's cache until it re-lists them, and a
  Helm rollback alone does not restore RBAC/webhook rules removed by a
  `helm upgrade` that already ran with the new chart — reapply the previous
  chart version to restore them.

### Default `TrafficPermission`/`TrafficRoute` creation removed

The control plane no longer creates the default allow-all `TrafficPermission`
and route-all `TrafficRoute` resources when a new `Mesh` is created. Traffic
routing and permissions for a new `Mesh` are now expected to come from the
default `MeshTrafficPermission` and implicit routing instead.

The following configuration has been removed:

- Control plane: `defaults.createMeshRoutingResources`
  (`KUMA_DEFAULTS_CREATE_MESH_ROUTING_RESOURCES`).

**Action required**

Remove `defaults.createMeshRoutingResources` from your control plane YAML
config and `KUMA_DEFAULTS_CREATE_MESH_ROUTING_RESOURCES` from your
environment if set. Both settings no longer have any effect in Kuma 3.0.0.
A stale YAML key is rejected only by strict config parsing; regular control
plane startup and the environment variable loader silently ignore the removed
setting.

The compatibility mode for creating legacy defaults for pre-2.6.0 zones is no
longer available in Kuma 3.0.0. Existing `TrafficPermission` and `TrafficRoute`
resources remain fully valid and supported; this change only affects automatic
creation of new defaults.

### Legacy top-level policy `targetRef` selector kinds removed

Policies no longer accept `spec.targetRef.kind: MeshSubset`,
`spec.targetRef.kind: MeshService`, or
`spec.targetRef.kind: MeshServiceSubset` at the top level. Creating or updating
a policy with one of these top-level kinds is rejected during validation.

This restriction applies to all admission paths, including policies replicated
from a Global control plane to Kubernetes Zone control planes by KDS. A policy
that still uses one of these legacy top-level kinds can fail to sync into a
Zone after the Zone is upgraded.

**Action required**

Before upgrading, migrate every policy that uses one of these top-level
`targetRef` kinds:

- Use `kind: Mesh` when the policy should apply to the whole mesh.
- Use `kind: Dataplane` with `labels` when the policy should apply to a subset
  of dataplanes previously selected by `MeshSubset`, `MeshService`, or
  `MeshServiceSubset`.

After the migration, verify the intended policy coverage in every Zone before
upgrading Zone control planes.

The `spec.to[].targetRef` field of a policy is a separate, narrower selector
and is affected too: no policy accepts `kind: MeshSubset`,
`kind: MeshServiceSubset`, or `kind: MeshGateway` there. The remaining
accepted kinds are per policy, so check the policy documentation before you
migrate:

- Every policy that has a `to` array accepts `kind: Mesh`.
- Most policies also accept `MeshService`, `MeshExternalService`, and
  `MeshMultiZoneService`.
- Only `MeshAccessLog`, `MeshLoadBalancingStrategy`, `MeshRetry`, and
  `MeshTimeout` accept `kind: MeshHTTPRoute`. For example,
  `MeshCircuitBreaker` rejects it.
- `MeshRateLimit` and `MeshFaultInjection` accept `kind: Mesh` only, and only
  when the top-level `targetRef` selects a gateway.

`MeshServiceSubset` remains valid only as a route `backendRefs[].kind`, not as
a top-level or `to[]` `targetRef.kind`.

### `from` removed from `MeshTimeout`

The deprecated `spec.from` array has been removed from `MeshTimeout`. Timeouts
for incoming traffic are now configured exclusively through `spec.rules`.
`spec.from` is silently dropped on create/update: if `spec.rules` or `spec.to`
is also set, the resource is accepted but `from` has no effect on inbound
configuration; if `from` was the only field set, the resulting spec has
neither `to` nor `rules`, so the request is rejected by validation.

**Action required**

Before upgrading, migrate every `MeshTimeout` that uses `spec.from` to
`spec.rules`. A `from` entry targeting `kind: Mesh` (all clients) maps to a
single catch-all rule:

```yaml
# before
spec:
  from:
    - targetRef:
        kind: Mesh
      default:
        idleTimeout: 1h
# after
spec:
  rules:
    - default:
        idleTimeout: 1h
```

### `from` removed from `MeshCircuitBreaker`

The deprecated `spec.from` array has been removed from `MeshCircuitBreaker`.
Circuit breaking for incoming traffic is now configured exclusively through
`spec.rules`. `spec.from` is silently dropped on create/update: if
`spec.rules` or `spec.to` is also set, the resource is accepted but `from` has
no effect on inbound configuration; if `from` was the only field set, the
resulting spec has neither `to` nor `rules`, so the request is rejected by
validation.

**Action required**

Before upgrading, migrate every `MeshCircuitBreaker` that uses `spec.from` to
`spec.rules`. A `from` entry targeting `kind: Mesh` (all clients) maps to a
single catch-all rule:

```yaml
# before
spec:
  from:
    - targetRef:
        kind: Mesh
      default:
        connectionLimits:
          maxConnections: 1024
# after
spec:
  rules:
    - default:
        connectionLimits:
          maxConnections: 1024
```

### `from` removed from `MeshAccessLog`

The deprecated `spec.from` array has been removed from `MeshAccessLog`. Access logging for incoming traffic is now configured exclusively through `spec.rules`. `spec.from` is silently dropped on create/update: if `spec.rules` or `spec.to` is also set, the resource is accepted but `from` has no effect on inbound configuration; if `from` was the only field set, the resulting spec has neither `to` nor `rules`, so the request is rejected by validation.

**Action required**

Before upgrading, migrate every `MeshAccessLog` that uses `spec.from` to `spec.rules`. A `from` entry targeting `kind: Mesh` (all clients) maps to a single catch-all rule:

```yaml
# before
spec:
  from:
    - targetRef:
        kind: Mesh
      default:
        backends:
          - type: File
            file:
              path: /tmp/access.log
# after
spec:
  rules:
    - default:
        backends:
          - type: File
            file:
              path: /tmp/access.log
```

### `from` removed from `MeshRateLimit`

The deprecated `spec.from` array has been removed from `MeshRateLimit`. Rate limiting for incoming traffic is now configured exclusively through `spec.rules`. `spec.from` is silently dropped on create/update: if `spec.rules` or `spec.to` is also set, the resource is accepted but `from` has no effect on inbound configuration; if `from` was the only field set, the resulting spec has neither `to` nor `rules`, so the request is rejected by validation.

**Action required**

Before upgrading, migrate every `MeshRateLimit` that uses `spec.from` to `spec.rules`. A `from` entry targeting `kind: Mesh` (all clients) maps to a single catch-all rule:

```yaml
# before
spec:
  from:
    - targetRef:
        kind: Mesh
      default:
        local:
          http:
            requestRate:
              num: 100
              interval: 10s
# after
spec:
  rules:
    - default:
        local:
          http:
            requestRate:
              num: 100
              interval: 10s
```

### Auto reachable services removed

The experimental auto reachable services feature has been removed. The control
plane no longer computes the set of services a data plane proxy can reach from
`MeshTrafficPermission` to prune its Envoy configuration.

The following configuration has been removed:

- Control plane: `experimental.autoReachableServices`
  (`KUMA_EXPERIMENTAL_AUTO_REACHABLE_SERVICES`).

**Action required**

Remove the setting above from your control plane config. Setting
`KUMA_EXPERIMENTAL_AUTO_REACHABLE_SERVICES` no longer has any effect in Kuma
3.0.0.

To trim the outbound clusters a proxy receives, configure reachable backends
explicitly on the `Dataplane` (the `kuma.io/reachable-backends` annotation on
Kubernetes). Traffic that is not permitted by a `MeshTrafficPermission` is
still denied at the proxy; it is simply no longer pruned from the proxy
configuration.

### `reachableServices` / `kuma.io/transparent-proxying-reachable-services` removed

The legacy `kuma.io/service`-based reachable services mechanism has been
removed in favor of `reachableBackends` (`kuma.io/reachable-backends` on
Kubernetes), which targets `MeshService`/`MeshExternalService`/
`MeshMultiZoneService` resources instead of the `kuma.io/service` tag.

The following have been removed:

- `Dataplane.spec.networking.transparentProxying.reachableServices`
- The `kuma.io/transparent-proxying-reachable-services` annotation on
  Kubernetes.

**Action required**

Migrate any usage of the annotation or field above to `reachableBackends` /
`kuma.io/reachable-backends`, referencing the target `MeshService`,
`MeshExternalService`, or `MeshMultiZoneService` by name/namespace/port
instead of the `kuma.io/service` tag value.

### `kumactl install observability` removed

The deprecated `kumactl install observability` command has been removed for Kuma 3.0.
Use separately managed observability components or your platform's preferred observability stack instead.
Kuma still ships first-party Grafana dashboards in the release tarball under `dashboards/grafana/`.

### `TrafficLog` no longer affects generated Envoy config

The legacy `TrafficLog` policy is no longer consumed when generating Envoy
configuration. Applying, updating, or removing a `TrafficLog` resource no
longer changes the access log configuration of any listener.

The `TrafficLog` resource, API, and KDS sync are still in place for this
release; existing resources are still accepted and stored.

**Action required**

Migrate access logging to `MeshAccessLog`, which replaces `TrafficLog`.

### KDS snapshot watchdog config removed

KDS now always uses the event-based snapshot watchdog. The previous polling
fallback and the configuration that controlled it have been removed.

The following configuration has been removed:

- Control plane: `experimental.kdsEventBasedWatchdog.enabled`
  (`KUMA_EXPERIMENTAL_KDS_EVENT_BASED_WATCHDOG_ENABLED`).
- Global control plane: `multizone.global.kds.refreshInterval`
  (`KUMA_MULTIZONE_GLOBAL_KDS_REFRESH_INTERVAL`).
- Zone control plane: `multizone.zone.kds.refreshInterval`
  (`KUMA_MULTIZONE_ZONE_KDS_REFRESH_INTERVAL`).

**Action required**

Remove the settings above from your control plane config and environment if
set. KDS snapshot generation remains event-driven, but the timing config moved
to the stable multizone KDS config:

- Global control plane: `multizone.global.kds.eventBasedWatchdog.flushInterval`
  (`KUMA_MULTIZONE_GLOBAL_KDS_EVENT_BASED_WATCHDOG_FLUSH_INTERVAL`),
  `multizone.global.kds.eventBasedWatchdog.fullResyncInterval`
  (`KUMA_MULTIZONE_GLOBAL_KDS_EVENT_BASED_WATCHDOG_FULL_RESYNC_INTERVAL`), and
  `multizone.global.kds.eventBasedWatchdog.delayFullResync`
  (`KUMA_MULTIZONE_GLOBAL_KDS_EVENT_BASED_WATCHDOG_DELAY_FULL_RESYNC`).
- Zone control plane: `multizone.zone.kds.eventBasedWatchdog.flushInterval`
  (`KUMA_MULTIZONE_ZONE_KDS_EVENT_BASED_WATCHDOG_FLUSH_INTERVAL`),
  `multizone.zone.kds.eventBasedWatchdog.fullResyncInterval`
  (`KUMA_MULTIZONE_ZONE_KDS_EVENT_BASED_WATCHDOG_FULL_RESYNC_INTERVAL`), and
  `multizone.zone.kds.eventBasedWatchdog.delayFullResync`
  (`KUMA_MULTIZONE_ZONE_KDS_EVENT_BASED_WATCHDOG_DELAY_FULL_RESYNC`).

### `TrafficTrace` no longer affects generated Envoy config

The legacy `TrafficTrace` policy is no longer consumed when generating Envoy
configuration. Applying, updating, or removing a `TrafficTrace` resource no
longer changes the tracing configuration of any listener.

The `TrafficTrace` resource, API, and KDS sync are still in place for this
release; existing resources are still accepted and stored.

**Action required**

Migrate tracing to `MeshTrace`, which replaces `TrafficTrace`.

### `Timeout` no longer affects generated Envoy config

The legacy `Timeout` policy is no longer consumed when generating Envoy
configuration. Applying, updating, or removing a `Timeout` resource no
longer changes the connection, request, or idle timeout configuration of
any listener, route, or cluster.

The `Timeout` resource, API, and KDS sync are still in place for this
release; existing resources are still accepted and stored.

**Action required**

Migrate timeouts to `MeshTimeout`, which replaces `Timeout`.

### `VirtualOutbound` no longer affects generated Envoy config

The legacy `VirtualOutbound` policy is no longer consumed when generating
Envoy configuration. Applying, updating, or removing a `VirtualOutbound`
resource no longer changes the outbounds, DNS domains, or zone ingress
destinations of any dataplane proxy.

The `VirtualOutbound` resource, API, and KDS sync are still in place for
this release; existing resources are still accepted and stored.

**Action required**

Migrate to `MeshHTTPRoute`/`MeshTCPRoute`, which replace `VirtualOutbound`.

### Legacy DNS VIP allocator and persisted VIP config removed

The control plane no longer computes or persists the `kuma-<mesh>-dns-vips`
`Config` resource. That resource was the write-only output of a legacy
background allocator; dataplane DNS records are generated directly from the
`MeshService`/`MeshExternalService`/`MeshMultiZoneService` DNS/VIP path, which
this change does not affect.

On Kubernetes, the control plane no longer reconciles the per-namespace
`ConfigMap` used to expose the allocator's output either.

The following configuration has been removed:

- `experimental.useTagFirstVirtualOutboundModel`
  (`KUMA_EXPERIMENTAL_USE_TAG_FIRST_VIRTUAL_OUTBOUND_MODEL`).
- `dnsServer.CIDR` (`KUMA_DNS_SERVER_CIDR`).
- `dnsServer.serviceVipEnabled` (`KUMA_DNS_SERVER_SERVICE_VIP_ENABLED`).
- `runtime.universal.vipRefreshInterval`
  (`KUMA_RUNTIME_UNIVERSAL_VIP_REFRESH_INTERVAL`).

**Action required**

Remove the settings above from your control plane config and environment if
set. Any previously persisted `kuma-<mesh>-dns-vips` `Config` resources (and
their Kubernetes `ConfigMap` mirrors) are no longer written or read and can be
deleted.

### `TrafficRoute` no longer affects generated Envoy config

The legacy `TrafficRoute` policy is no longer consumed when generating Envoy
configuration. Applying, updating, or removing a `TrafficRoute` resource no
longer changes outbound routing, load balancing, or reachable destinations
for any dataplane, zone ingress, or zone egress.

Traffic between services now always flows by default, the same way it already
did in meshes that only use `MeshHTTPRoute`/`MeshTCPRoute`. A user-authored
`TrafficRoute` no longer has any effect, including the legacy behavior where
just having a `TrafficRoute` disabled the default routing fallback when no
`MeshHTTPRoute`/`MeshTCPRoute` matched instead.

The `TrafficRoute` resource, API, and KDS sync are still in place for this
release; existing resources are still accepted and stored.

**Action required**

Migrate routing to `MeshHTTPRoute`/`MeshTCPRoute`, which replace
`TrafficRoute`.

### `TrafficPermission` no longer affects generated Envoy config

The legacy `TrafficPermission` policy is no longer consumed when generating
Envoy configuration. Applying, updating, or removing a `TrafficPermission`
resource no longer changes the RBAC rules of any dataplane inbound listener,
zone egress external-service filter chain, or gateway external-service
routing. `MeshTrafficPermission` is the only policy that now controls
inbound and egress access.

For dataplanes, zone egresses, and gateways already using
`MeshTrafficPermission`, this is a no-op: `MeshTrafficPermission` already
took precedence over `TrafficPermission` and provably re-owns every mTLS
inbound RBAC filter. For any inbound, external service, or gateway route
that still relied solely on a `TrafficPermission` grant with no equivalent
`MeshTrafficPermission`, access now defaults to deny instead of allow. No
mTLS inbound listener or egress external-service filter chain loses its RBAC
filter — every one still gets a default-deny filter, so this is a fail-closed
change rather than fail-open. External-service outbound clusters are also now
always generated, instead of only when a `TrafficPermission` granted access.

The `TrafficPermission` resource, API, and KDS sync are still in place for
this release; existing resources are still accepted and stored.

**Action required**

Migrate access control to `MeshTrafficPermission`, which replaces
`TrafficPermission`.

### `ProxyTemplate` no longer affects generated Envoy config

The `ProxyTemplate` resource is no longer consumed when generating Envoy
configuration. Applying, updating, or removing a `ProxyTemplate` resource no
longer changes the Envoy configuration of any dataplane: neither the
user-selected profile imports nor the raw xDS resources and modifications
(`Conf.Resources`, `Conf.Modifications`) are applied anymore. The
default-profile mechanism that `ProxyTemplate` used internally is unaffected
and still generates the standard Envoy configuration for every dataplane.

The `ProxyTemplate` resource, API, and KDS sync are still in place for this
release; existing resources are still accepted and stored.

**Action required**

Migrate raw Envoy resource modifications to `MeshProxyPatch`, which replaces
the `ProxyTemplate.Conf.Resources` and `ProxyTemplate.Conf.Modifications`
raw-Envoy paths.

### Built-in gateway no longer falls back to `MeshGatewayRoute` or legacy connection policies

The built-in gateway (`MeshGateway`) no longer consumes the legacy
`MeshGatewayRoute` policy when generating Envoy configuration. Previously, a
gateway listener/hostname without any matching `MeshHTTPRoute`/`MeshTCPRoute`
fell back to routes derived from `MeshGatewayRoute` resources; that fallback
is gone, so a gateway host with no `MeshHTTPRoute`/`MeshTCPRoute` now serves a
`404` for every request instead.

The built-in gateway also no longer binds the legacy `FaultInjection` and
`HealthCheck` connection policies to routes it generates; those policies
never had any effect on `MeshHTTPRoute`/`MeshTCPRoute`-derived routes and are
now also ignored for `MeshGatewayRoute`-derived ones.

The `MeshGatewayRoute`, `FaultInjection`, and `HealthCheck` resources, APIs,
and KDS sync are still in place for this release; existing resources are
still accepted and stored.

**Action required**

Migrate gateway routing to `MeshHTTPRoute`/`MeshTCPRoute`, which replace
`MeshGatewayRoute`. Migrate fault injection and health checking on gateway
routes to `MeshFaultInjection` and `MeshHealthCheck`.

### Delta xDS is now the only xDS protocol

The control plane previously delivered configuration to data plane proxies using
state-of-the-world (SOTW) xDS by default, with incremental (Delta) xDS available
behind an experimental flag. Delta xDS is now always used and the SOTW code path
has been removed. The control plane no longer serves `StreamAggregatedResources`;
only `DeltaAggregatedResources` is implemented.

The following configuration has been removed:

- Control plane: `experimental.deltaXds` (`KUMA_EXPERIMENTAL_DELTA_XDS`) and the
  Helm value `experimental.deltaXds`.
- Data plane: `dataplaneRuntime.envoyXdsTransportProtocolVariant`
  (`KUMA_DATAPLANE_RUNTIME_ENVOY_XDS_TRANSPORT_PROTOCOL_VARIANT`) and the
  `kuma.io/xds-transport-protocol-variant` pod annotation.

**Action required**

Remove the settings above from your control plane config, Helm values, and pod
annotations. Setting `KUMA_EXPERIMENTAL_DELTA_XDS` no longer has any effect in
Kuma 3.0.0.

For zero-downtime upgrades, first enable Delta xDS with
`KUMA_EXPERIMENTAL_DELTA_XDS=true` on the old control plane version, then restart
all `kuma-dp` instances (or roll the workloads on Kubernetes) so their Envoy
bootstraps use Delta xDS. After every data plane proxy is connected with Delta
xDS, upgrade the control plane to Kuma 3.0.0 and roll the data plane proxies
again as part of the normal upgrade flow.

The protocol a proxy uses is fixed in its Envoy bootstrap at startup, so a proxy
that started against an older control plane keeps using SOTW until it reconnects
with a fresh bootstrap. Once the control plane is upgraded to Kuma 3.0.0, any
proxy still trying to use the removed SOTW stream cannot establish ADS and must
be restarted with a Delta xDS bootstrap.

### Legacy policy resources removed

The 12 legacy policy resources superseded by the `Mesh*` targetRef policies
have been removed from the resource registry: `TrafficPermission`,
`TrafficRoute`, `TrafficLog`, `TrafficTrace`, `HealthCheck`,
`CircuitBreaker`, `Retry`, `Timeout`, `RateLimit`, `FaultInjection`,
`VirtualOutbound`, and `ProxyTemplate`. Each of these had
already stopped affecting generated Envoy configuration in earlier releases
(see the entries above); this change removes the resources themselves.

For every one of these types: the REST API endpoints (including the generic
`_resources`/`_dataplanes` inspect endpoints), `kumactl get`/`inspect`
subcommands, KDS sync, and `MeshInsight`/`ServiceInsight` policy counters are
gone, and the corresponding CRD is no longer installed on Kubernetes.

The `ProxyTemplate` proto message and the template/profile indirection it fed
(`ProxyTemplateResolver`, profile imports, `RegisterProfile`) are gone as well.
The control plane now always generates the standard Envoy configuration for
every data plane proxy.

**Action required**

Delete any remaining resources of these types before upgrading — the control
plane no longer accepts create/update requests for them, and stored resources
of a removed type are not migrated. Remove any automation, dashboards, or
kumactl scripts that reference these resource types, REST paths, or CRDs.

### Legacy policy resources dropped from control plane RBAC and webhooks

The control plane `ClusterRole` no longer grants access to the legacy policy
CRDs removed above (`proxytemplates`, `ratelimits`, `trafficpermissions`,
`trafficroutes`, `timeouts`, `retries`, `circuitbreakers`, `virtualoutbounds`,
`faultinjections`, `healthchecks`, `trafficlogs`, `traffictraces`), and the
`kuma.io` validating and owner-reference admission webhooks no longer register
rules for them. Those CRDs are no longer installed, so the rules matched
nothing.

**Action required**

None. If you copied the Kuma `ClusterRole` or webhook configuration into your
own manifests, drop the same resource names from your copy.

### Built-in gateway API and CRDs removed

The built-in gateway API has been removed entirely. The `MeshGateway`,
`MeshGatewayRoute`, `MeshGatewayInstance`, and `MeshGatewayConfig` resources,
their Go/proto types, their Kubernetes CRDs
(`meshgateways.kuma.io`, `meshgatewayroutes.kuma.io`,
`meshgatewayinstances.kuma.io`, `meshgatewayconfigs.kuma.io`), and KDS sync
registration for these types no longer exist. `MeshGateway` is also no longer
a valid `targetRef.kind` for any policy.

A `Dataplane` with `networking.gateway.type: BUILTIN` is now rejected at
admission and update. The `Dataplane.networking.gateway` message and the
`DELEGATED` gateway type are unaffected — delegated gateways (bring your own
`Deployment`/`Service` fronting a Kuma-injected pod annotated
`kuma.io/gateway: enabled`) continue to work exactly as before.

**Action required**

- Before upgrading, delete any remaining `MeshGateway`, `MeshGatewayRoute`,
  `MeshGatewayInstance`, and `MeshGatewayConfig` resources. On Kubernetes,
  deleting their CRDs (as this Helm chart upgrade does) removes any resources
  still stored under them; delete them explicitly first if you need to inspect
  or back them up beforehand.
- Migrate any remaining `BUILTIN` gateway `Dataplane`s to `DELEGATED` before
  upgrading (see "Built-in gateway Kubernetes controllers removed" above) —
  after upgrading, creating or updating a `Dataplane` with
  `networking.gateway.type: BUILTIN` fails validation.
- The default Helm-installed cluster RBAC no longer grants access to
  `meshgateways`, `meshgatewayroutes`, `meshgatewayinstances`,
  `meshgatewayconfigs`, or their `/status` and `/finalizers` subresources, and
  the validating webhook configuration no longer includes these types. If you
  manage RBAC or webhooks manually, remove these rules; if you keep them, they
  are harmless but unused.

### Legacy HS256 signing keys no longer accepted

The control plane no longer accepts JWT tokens signed with the symmetric
HS256 algorithm. This algorithm was used to sign Dataplane Tokens in
pre-1.4.x versions of Kuma; support for verifying such tokens was kept
around for backwards compatibility long after RS256 became the default.
`SigningKeyAccessor.GetLegacyKey` and the HS256 branch of token validation
have been removed.

**Action required**

If any Dataplane Tokens issued by a pre-1.4.x control plane are still in
use, rotate them (generate new tokens with the current control plane)
before upgrading — they will be rejected as using an unsupported
algorithm afterward.

### `kuma.io/mesh` annotation no longer honored on Kubernetes

The control plane no longer reads the deprecated `kuma.io/mesh` annotation on
a Pod, Service, HTTPRoute, or Namespace to assign the resource's mesh. Only
the `kuma.io/mesh` label is used: on the resource itself, or — for namespaced
resources (Pod, Service, HTTPRoute) — on their Namespace. Resources without
the label fall back to the `default` mesh.

**Action required**

If you still set `kuma.io/mesh` as an annotation, switch it to a label before
upgrading. A resource that only carries the annotation will resolve to the
`default` mesh after upgrading.

### `MeshMetric` OpenTelemetry backend no longer accepts an inline `endpoint`

The deprecated `default.backends[].openTelemetry.endpoint` field has been
removed from the `MeshMetric` policy. `backendRef`, pointing at a
`MeshOpenTelemetryBackend` resource, is now the only way to configure an
OpenTelemetry metrics backend.

**Action required**

If any `MeshMetric` policy still sets `openTelemetry.endpoint`, create a
`MeshOpenTelemetryBackend` resource with the equivalent endpoint and update
the policy to reference it via `openTelemetry.backendRef` before upgrading.
Policies that still set `openTelemetry.endpoint` will fail validation.

### `MeshAccessLog` OpenTelemetry backend no longer accepts an inline `endpoint`

The deprecated `default.backends[].openTelemetry.endpoint` /
`rules[].default.backends[].openTelemetry.endpoint` field has been removed
from the `MeshAccessLog` policy. `backendRef`, pointing at a
`MeshOpenTelemetryBackend` resource, is now the only way to configure an
OpenTelemetry access log backend.

**Action required**

If any `MeshAccessLog` policy still sets `openTelemetry.endpoint`, create a
`MeshOpenTelemetryBackend` resource with the equivalent endpoint and update
the policy to reference it via `openTelemetry.backendRef` before upgrading.
Policies that still set `openTelemetry.endpoint` will fail validation.

### `MeshLoadBalancingStrategy` load-balancer-specific `hashPolicies` removed

The deprecated `spec.to[].default.loadBalancer.ringHash.hashPolicies` and
`spec.to[].default.loadBalancer.maglev.hashPolicies` fields have been
removed. `spec.to[].default.hashPolicies` is now the only place to configure
hash policies.

**Action required**

Move any `hashPolicies` still configured under `loadBalancer.ringHash` or
`loadBalancer.maglev` to `spec.to[].default.hashPolicies` before upgrading.
After upgrading, policies that still set the removed nested fields may be
rejected by validation or have those fields pruned by the API server, and they
no longer affect the generated Envoy config.

### `MeshTrace` OpenTelemetry backend no longer accepts an inline `endpoint`

The deprecated `default.backends[].openTelemetry.endpoint` field has been
removed from the `MeshTrace` policy. `backendRef`, pointing at a
`MeshOpenTelemetryBackend` resource, is now the only way to configure an
OpenTelemetry tracing backend.

**Action required**

If any `MeshTrace` policy still sets `openTelemetry.endpoint`, create a
`MeshOpenTelemetryBackend` resource with the equivalent endpoint and update
the policy to reference it via `openTelemetry.backendRef` before upgrading.
Policies that still set `openTelemetry.endpoint` will fail validation.

### `Mesh.spec.tracing` removed

The inline `tracing` field (and its `Tracing`/`TracingBackend`/
`DatadogTracingBackendConfig`/`ZipkinTracingBackendConfig` types) has been
removed from the `Mesh` resource spec. The `MeshTrace` policy has been the GA
replacement for configuring tracing and is unaffected by this change.

**Action required**

Migrate any `Mesh` resources that still configure `spec.tracing` to a
`MeshTrace` policy before upgrading. A `Mesh` spec that still sets `tracing`
continues to apply successfully; the field is silently ignored by the control
plane.

### `Mesh.spec.routing.localityAwareLoadBalancing` removed

The inline `routing.localityAwareLoadBalancing` field has been removed from
the `Mesh` resource spec. The `MeshLoadBalancingStrategy` policy has been the
GA replacement for configuring locality-aware load balancing and is
unaffected by this change.

**Action required**

Migrate any `Mesh` resources that still configure
`spec.routing.localityAwareLoadBalancing` to a `MeshLoadBalancingStrategy`
policy before upgrading. A `Mesh` spec that still sets
`routing.localityAwareLoadBalancing` continues to apply successfully; the
field is silently ignored by the control plane.

### `Mesh.spec.logging` removed

The inline `logging` field (and its `Logging`/`LoggingBackend`/
`FileLoggingBackendConfig`/`TcpLoggingBackendConfig` types) has been removed
from the `Mesh` resource spec. The `MeshAccessLog` policy has been the GA
replacement for configuring access logging and is unaffected by this change.

**Action required**

Migrate any `Mesh` resources that still configure `spec.logging` to a
`MeshAccessLog` policy before upgrading. A `Mesh` spec that still sets
`logging` continues to apply successfully; the field is silently ignored by
the control plane.

### Per-zone MeshExternalService routing removed

A `MeshExternalService` labeled with `kuma.io/zone` is no longer restricted to
being reached only from that zone via a routing path through the remote
zone's ingress and egress. It is now reachable directly through the local
zone egress from every zone, the same as an unlabeled `MeshExternalService`.

**Action required**

None for typical usage; existing `kuma.io/zone` labels on `MeshExternalService`
resources are no longer used to gate reachability and can be removed. If you
relied on the label to force all traffic through a specific zone's egress
(for example, because only that zone has network-level access to the
external endpoint), that forwarding no longer happens: every zone's local
egress now dials the external endpoint directly, so make sure each zone's
network path to the endpoint is in place before upgrading.

### `MeshInsight.policies` removed in favor of `resources`

The deprecated `policies` field (a map of policy type to a `total` count) has
been removed from `MeshInsight`. The `resources` field, which reports a
`total` count for every resource type (policies included), has been the
replacement since it was introduced and is unaffected by this change.

**Action required**

Update any automation or dashboards that read `MeshInsight.policies` (via the
REST API or `kumactl inspect meshes`) to read the equivalent entry from
`MeshInsight.resources` instead, keyed by the same resource type name.

### `Mesh.spec.networking.outbound.passthrough` removed

The inline `networking.outbound.passthrough` field has been removed from the
`Mesh` resource spec. The `MeshPassthrough` policy is the replacement for
controlling the default outbound passthrough cluster and is unaffected by
this change. After upgrading, the control plane always behaves as if
`passthrough` was `true` (its previous default) unless a `MeshPassthrough`
policy says otherwise.

**Action required**

Migrate any `Mesh` resources that still set `networking.outbound.passthrough`
to `false` to a `MeshPassthrough` policy with `targetRef.kind: Mesh` and
`default.passthroughMode: None` before upgrading. A `Mesh` spec that still
sets `networking.outbound.passthrough` continues to apply successfully; the
field is silently ignored by the control plane.

### `MeshTrafficPermission.spec.from` removed

The `from` field (and its legacy client-targetRef-based `Allow`/`Deny`/
`AllowWithShadowDeny` matching) has been removed from the
`MeshTrafficPermission` resource spec. The `rules` field, which matches
clients by `MeshIdentity` (`spiffeID`) or SNI instead of by dataplane tag
subsets, is now the only supported way to configure traffic permissions and
is unaffected by this change. A `MeshTrafficPermission` must now define at
least one entry in `rules`; a spec with only `from` (or with neither `from`
nor `rules`) fails validation with `policy must define rules`.

**Action required**

Migrate any `MeshTrafficPermission` resources that still configure `from` to
use `rules` with `MeshIdentity` (`spiffeID`) matches before upgrading.
Dataplane proxies still on legacy mTLS (no `MeshIdentity`/SPIFFE identity)
that are matched only by a `from`-based `MeshTrafficPermission` will
default-deny once that policy is migrated or removed, unless a `rules`-based
policy is added to allow the same traffic.

### Legacy dataplane inspect rules endpoint removed

The legacy `GET /meshes/{mesh}/dataplanes/{dataplane}/rules` endpoint has
been removed.

**Action required**

Use `GET /meshes/{mesh}/dataplanes/{name}/_policies` instead (or the
per-inbound/outbound scoped variants — see below).

`kumactl inspect dataplane --type=policies` now calls
`GET /meshes/{mesh}/dataplanes/{name}/_policies` (and the
`_inbounds/{inbound_kri}/_policies`, `_outbounds/{outbound_kri}/_policies`,
and `_outbounds/{outbound_kri}/_routes/{route_kri}/_policies` variants for
per-inbound/outbound scoping) and requires no changes to invocation. The
underlying `GET /meshes/{mesh}/dataplanes/{dataplane}/policies` HTTP
endpoint it used to call is still registered — the vendored GUI bundle
(`app/kuma-ui`) still calls it directly and is re-vendored on its own
release cadence — but it is deprecated; new integrations should call
`_policies` instead.

### Cross-zone `MeshService` routing requires `MeshZoneAddress`

The publicly reachable address of a remote zone proxy is now taken solely from
the `MeshZoneAddress` resource. The control plane no longer falls back to the
`advertisedAddress`/`advertisedPort` of a `ZoneIngress` when building endpoints
for `MeshService` and `MeshMultiZoneService` destinations in other zones.

On Kubernetes, the control plane creates `MeshZoneAddress` automatically for every mesh-scoped zone proxy `Service` (`meshes[].ingress.enabled=true`), which carries the `k8s.kuma.io/zone-proxy-type: ingress` label. On Universal it is authored by the user.

`MeshZoneAddress` is mesh-scoped: every mesh whose services are consumed from another zone needs its own resource in the zone that serves them.

**Action required**

`MeshZoneAddress` does not exist before 3.0.0, so it cannot be created ahead of the upgrade. Create it immediately after upgrading the zone control plane — cross-zone `MeshService` and `MeshMultiZoneService` traffic to that zone is down until the resource exists and has propagated over KDS.

On Universal, author one `MeshZoneAddress` per mesh on the zone control plane, pointing at the public address and port of that zone's ingress listener:

```yaml
type: MeshZoneAddress
name: zone-proxy-ingress
mesh: default
labels:
  kuma.io/origin: zone
spec:
  address: 10.0.0.1
  port: 10001
```

`kuma.io/origin: zone` is required on a zone control plane federated to a global control plane. `kuma.io/zone` is stamped by the control plane; if you set it explicitly it must match the local zone name. `spec.address` accepts an IP or a DNS name; a DNS name is resolved by the control plane.

Zones without a `MeshZoneAddress` are not reachable cross-zone: their `MeshService` destinations get no endpoints in other zones. The control plane logs `no MeshZoneAddress found for zone` when this happens.

### kuma-dp `configDir` / `socketDir` removed

The deprecated `configDir` and `socketDir` `dataplaneRuntime` config fields (and their
`KUMA_DATAPLANE_RUNTIME_CONFIG_DIR` / `KUMA_DATAPLANE_RUNTIME_SOCKET_DIR` environment
variables) have been removed from `kuma-dp`. The `--config-dir` flag has also been removed.

**Action required**

Use `workDir` (`KUMA_DATAPLANE_RUNTIME_WORK_DIR` / `--work-dir`) instead. `--config-dir` now
fails with `unknown flag`, so any script or deployment passing it will error immediately.
`configDir`/`socketDir` in YAML config and `KUMA_DATAPLANE_RUNTIME_CONFIG_DIR` /
`KUMA_DATAPLANE_RUNTIME_SOCKET_DIR` are silently ignored, since the config loader does not
reject unknown fields — proxies still relying on them will silently fall back to a
generated temporary directory instead of erroring.

### Virtual probes removed

The legacy Virtual Probes feature, deprecated since 2.9 in favor of Application Probe Proxy, has been removed. The following are gone:

- Pod annotations `kuma.io/virtual-probes` and `kuma.io/virtual-probes-port`
- Control plane configuration keys `runtime.kubernetes.injector.virtualProbesEnabled` and `runtime.kubernetes.injector.virtualProbesPort`
- Environment variables `KUMA_RUNTIME_KUBERNETES_VIRTUAL_PROBES_ENABLED` and `KUMA_RUNTIME_KUBERNETES_VIRTUAL_PROBES_PORT`
- The `probes` field on `Dataplane` resources (Kubernetes and Universal)

**Action required**

Re-inject every pod after upgrading — the injector no longer rewrites pod probes to a virtual probes listener, and any pod injected by a pre-3.0.0 webhook keeps stale `kuma.io/virtual-probes*` annotations and a probes-listener sidecar config until it is redeployed. Application Probe Proxy (`kuma.io/application-probe-proxy-port`) is now the only supported probe-rewriting path on Kubernetes and is enabled by default.

**Warning**: `probes:` submitted on a Universal `Dataplane` is now silently dropped — the field no longer exists on the resource, so it is accepted and discarded rather than rejected. Setting `kuma.io/application-probe-proxy-port: "0"` no longer falls back to virtual probes; it now leaves the pod's original probes untouched. Pods relying on the old fallback must either remove the annotation to keep using Application Probe Proxy, or exclude their probe ports from inbound traffic redirection if they need probes served without mTLS.

### Injector sidecar container `adminPort` removed

The deprecated `kuma.runtime.kubernetes.injector.sidecarContainer.adminPort`
config field and `KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_ADMIN_PORT`
environment variable have been removed. The field was already dead — the
injector always read the Envoy admin port from `bootstrapServer.params.adminPort`
(`KUMA_BOOTSTRAP_SERVER_PARAMS_ADMIN_PORT`).

**Action required**

Use `kuma.bootstrapServer.params.adminPort` /
`KUMA_BOOTSTRAP_SERVER_PARAMS_ADMIN_PORT` instead. Any deployment still setting
the removed field or environment variable will have it silently ignored, since
the config loader does not reject unknown fields.

### `Metrics.Mesh.MinResyncTimeout` / `MaxResyncTimeout` removed

The deprecated `Metrics.Mesh.MinResyncTimeout` (`KUMA_METRICS_MESH_MIN_RESYNC_TIMEOUT`) and
`Metrics.Mesh.MaxResyncTimeout` (`KUMA_METRICS_MESH_MAX_RESYNC_TIMEOUT`) config fields have
been removed.

**Action required**

Use `Metrics.Mesh.MinResyncInterval` (`KUMA_METRICS_MESH_MIN_RESYNC_INTERVAL`) and
`Metrics.Mesh.FullResyncInterval` (`KUMA_METRICS_MESH_FULL_RESYNC_INTERVAL`) instead.
`MinResyncTimeout`/`MaxResyncTimeout` in YAML config and their environment variables are
now silently ignored, since the config loader does not reject unknown fields.

### `MeshExternalService` TLS verification uses the `SecureDataSource` shape

`spec.tls.verification.caCert`, `.clientCert` and `.clientKey` on `MeshExternalService` now
use the same `SecureDataSource` type as `MeshIdentity`, instead of the old `DataSource` type.
The old type has been removed from the API entirely.

The old type had no discriminator: it was a flat object with `secret`, `inline` or
`inlineString`. The new type requires a `type` discriminator and nests the value under a
field matching it. Every old field has to be rewritten:

| Old field | New field |
|---|---|
| `inline: <base64>` | `type: InsecureInline`, `insecureInline.value: <plain text>` |
| `inlineString: <text>` | `type: InsecureInline`, `insecureInline.value: <text>` |
| `secret: <name>` | `type: Secret`, `secretRef: {kind: Secret, name: <name>}` |

`inline` was base64-encoded, `insecureInline.value` is plain text, so decode the old value
when rewriting it. For example `inline: dGVzdA==` becomes:

```yaml
caCert:
  type: InsecureInline
  insecureInline:
    value: test
```

`File` and `EnvVar`, the two other `SecureDataSource` types, are rejected on
`MeshExternalService` — they read the control plane's own filesystem and environment. This
also applies when `spec.extension` is set, even though an extension owns the rest of the
`spec.tls` validation.

**Action required**

Rewrite `caCert`, `clientCert` and `clientKey` on every `MeshExternalService` to the new
shape as part of the upgrade.

**Warning**: a `MeshExternalService` written in the old shape after the upgrade is rejected
at write time, because the missing `type` discriminator is a validation violation. Resources
already stored in the old shape are not rejected — the control plane cannot read their TLS
material, so the destination is dropped from the xDS config of every proxy routing to it,
with an error logged on the control plane. Plan the rewrite together with the upgrade to
avoid an outage on those destinations.

### Inbound `tags` removed from `Dataplane`

The `networking.inbound[].tags` field has been removed from the `Dataplane` resource. Tags for an inbound must now be set through the `Dataplane`'s own `metadata.labels` (Kubernetes) or the `Dataplane`'s top-level `labels:` field (Universal), not per-inbound. This only affects Universal: on Kubernetes the field was already ignored in favor of pod labels.

**Action required**

Move any per-inbound tags declared in hand-authored Universal `Dataplane` resources to `Dataplane` labels before upgrading.

**Warning**: `networking.inbound[].tags` is silently dropped on deserialization, not rejected — the field is `reserved` in the proto and protos are unmarshalled with `AllowUnknownFields`, so it is simply ignored. Dataplanes still submitting it will upgrade without error, but any policy matching on those inbound tags stops matching, with nothing in the API to signal it.

### A resource's `Mesh` ownerReference must match its mesh

On Kubernetes, a mesh-scoped resource can no longer end up owned by one `Mesh` while belonging to another. The validating webhook now rejects any create or update where a `kuma.io` `ownerReference` of kind `Mesh` names a mesh different from the resource's own `kuma.io/mesh` label (or `mesh` field). This closes both ways such a mismatch could previously happen: editing `kuma.io/mesh` on an existing resource to move it between meshes in place, and hand-writing or hand-editing a `Mesh` ownerReference. Writes from the control plane, `generic-garbage-collector` and the storage version migrator are unaffected, so a `Dataplane` can still be moved between meshes as its Pod is rescheduled.

**Action required**

If a resource in your cluster already has a `Mesh` ownerReference that disagrees with its mesh, it must be deleted and re-applied in the correct mesh — the webhook rejects any further edit to it until then.

### The control plane and hook containers drop all Linux capabilities

`controlPlane.containerSecurityContext` and `hooks.containerSecurityContext` now default to `allowPrivilegeEscalation: false` and `capabilities.drop: [ALL]`, alongside the `readOnlyRootFilesystem: true` they already carried.

Neither the control plane nor the `kubectl` and `kumactl` hook jobs need a capability. Every port the control plane binds is above 1024, so it never needed `NET_BIND_SERVICE`, and nothing in the codebase opens a raw or packet socket. The injected sidecar and the zoneproxy have shipped with both settings for some time; this brings the control plane in line with them. The transparent proxy init container is unaffected and keeps the `NET_ADMIN` and `NET_RAW` it adds for `iptables`, as does the CNI container, which still runs as root.

**Action required**

Helm merges these values key by key, so overriding one key of `containerSecurityContext` no longer replaces the whole block: an override that sets only `readOnlyRootFilesystem` now also inherits the two new keys. If your control plane needs a capability, or you run a `SecurityContextConstraint` or admission policy that grants one, set it back explicitly:

```yaml
controlPlane:
  containerSecurityContext:
    allowPrivilegeEscalation: true
    capabilities:
      drop: []
```

### `MeshProxyPatch` can now change circuit breaker thresholds

Since 2.14.0 every generated cluster carries a `DEFAULT`-priority circuit breaker threshold, and a `MeshProxyPatch` patching `circuitBreakers` appended a second one for that same priority. Envoy honours only the first threshold matching a routing priority, so the patched entry was dead config: `/config_dump` showed the requested values while Envoy kept what `MeshCircuitBreaker`, or its own defaults, had set. The patch now merges into the existing threshold of the same priority instead of appending. A threshold whose priority is not yet on the cluster still appends, and a `value` listing the same priority twice keeps the first entry, matching how Envoy resolves them.

**Action required**

Review every `MeshProxyPatch` that patches `circuitBreakers`. Where the same cluster is also covered by a `MeshCircuitBreaker`, the patch now overrides that policy for each field it sets, instead of being ignored — `MeshProxyPatch` runs last, so it wins the fields it names and the policy keeps the rest. Remove patches you wrote before 2.14.0 and no longer rely on, and drop any workaround you put in place because the patch appeared to do nothing.

## Upgrade to `2.13.7`

Patch releases normally do not require upgrade instructions. The entry below is included because the underlying change is a security fix that alters TLS verification behavior in a way some deployments may notice.

### Insecure TLS fallback removed when no CA cert is provided

Previously, when no CA certificate was configured:

- `kuma-dp` connecting to `kuma-cp` over HTTPS silently set `InsecureSkipVerify=true` and only printed a warning.
- `kumactl` connecting to the API server silently set `InsecureSkipVerify=true`.

These clients now use the host system's trust store (Go's default) and verify the server certificate. Skipping verification is opt-in.

**Impact**

Connections to a control plane whose certificate cannot be verified against the system trust store (typically self-signed or signed by a private CA that is not installed on the host) will now fail instead of silently succeeding without verification. The inter-CP `grpcs` client returns an error if called without a TLS config.

Deployments that already provide a CA via `--ca-cert-file` / `KUMA_CONTROL_PLANE_CA_CERT` / `KUMA_CONTROL_PLANE_CA_CERT_FILE`, or that use a publicly trusted certificate, or that run the default Helm-installed setup, are not affected.

**Action required**

If your control plane uses a self-signed certificate or a private CA, provide that CA explicitly.

For `kuma-dp`:

```sh
kuma-dp run --ca-cert-file=/path/to/ca.pem ...
# or
KUMA_CONTROL_PLANE_CA_CERT_FILE=/path/to/ca.pem kuma-dp run ...
# or pass the PEM directly
KUMA_CONTROL_PLANE_CA_CERT="$(cat /path/to/ca.pem)" kuma-dp run ...
```

For `kumactl`:

```sh
kumactl config control-planes add --ca-cert-file=/path/to/ca.pem ...
```

**Opt-out (insecure, development/testing only)**

A new explicit flag preserves the previous insecure behavior:

- `kuma-dp run --skip-verify` (env var `KUMA_CONTROL_PLANE_TLS_SKIP_VERIFY=true`)
- `kumactl config control-planes add --skip-verify`

Do not use this in production.

### Rolling-upgrade note: inbound listeners use SO_REUSEPORT by default

See [Inbound listeners now use SO_REUSEPORT by default](#inbound-listeners-now-use-so_reuseport-by-default).

## Upgrade to `2.12.11`

See [Insecure TLS fallback removed when no CA cert is provided](#insecure-tls-fallback-removed-when-no-ca-cert-is-provided).

## Upgrade to `2.11.14`

See [Insecure TLS fallback removed when no CA cert is provided](#insecure-tls-fallback-removed-when-no-ca-cert-is-provided).

## Upgrade to `2.9.16`

See [Insecure TLS fallback removed when no CA cert is provided](#insecure-tls-fallback-removed-when-no-ca-cert-is-provided).

## Upgrade to `2.7.26`

See [Insecure TLS fallback removed when no CA cert is provided](#insecure-tls-fallback-removed-when-no-ca-cert-is-provided).

## Upgrade to `2.14.x`

### Inbound listeners now use SO_REUSEPORT by default

> **Affected versions:** this change is in `2.13.7`+ and all `2.14.x`, not just `2.14`. The rolling-upgrade note below applies any time you upgrade from a version without it (before `2.13.7`) to one with it — including a `2.13.x` upgrade that crosses `2.13.7`.

The data plane now advertises the `feature-reuse-port` capability to the control plane, which causes inbound Envoy listeners to be generated with `enable_reuse_port: true`. This lets each Envoy worker thread own its own listen socket, improving connection distribution under load.

**Note:** `enable_reuse_port` cannot be changed on a running Envoy listener. If a data plane is upgraded and the flag later toggled, the listener will not pick up the change until the data plane restarts.

**Action required:**

If your environment has known issues with `SO_REUSEPORT` (e.g. certain Linux kernel versions or network configurations), disable the feature before upgrading using the instructions below.

In a rolling CP upgrade, **disable reuse port for ZoneIngress/ZoneEgress before upgrading**.

During the upgrade, a ZoneIngress/ZoneEgress can first receive `enable_reuse_port: false` from an old CP,
then `enable_reuse_port: true` from a new CP.
Envoy cannot change this setting on a live listener, so it NACKs the update and keeps serving the stale listener.

**Kubernetes — injected sidecars**

Create a `ContainerPatch`:

```yaml
apiVersion: kuma.io/v1alpha1
kind: ContainerPatch
metadata:
  name: disable-reuse-port
  namespace: kuma-system
spec:
  sidecarPatch:
  - op: add
    path: /env/-
    value: '{
      \"name\": \"KUMA_DATAPLANE_RUNTIME_REUSE_PORT_ENABLED\",
      \"value\": \"false\"
    }'
```

Then set the annotation `kuma.io/container-patches` on deployments where it should be disabled:

```yaml
"kuma.io/container-patches": "disable-reuse-port"
```

or globally for all injected sidecars via control-plane configuration:

```
KUMA_RUNTIME_KUBERNETES_INJECTOR_CONTAINER_PATCHES="disable-reuse-port"
```

**Kubernetes — ZoneIngress and ZoneEgress**

`ContainerPatch` only applies to sidecars injected into user pods. The Helm chart does not expose an env-var override for the `kuma-ingress`/`kuma-egress` Deployments, so patch them directly:

```bash
kubectl -n kuma-system set env deployment/kuma-ingress KUMA_DATAPLANE_RUNTIME_REUSE_PORT_ENABLED=false
kubectl -n kuma-system set env deployment/kuma-egress  KUMA_DATAPLANE_RUNTIME_REUSE_PORT_ENABLED=false
```

If you manage Helm releases declaratively, add the env var via a kustomize patch or post-render step targeting the same Deployments.

**Universal**

Set the environment variable when running `kuma-dp` (data plane, zone ingress, or zone egress):

```bash
KUMA_DATAPLANE_RUNTIME_REUSE_PORT_ENABLED=false kuma-dp run ...
```

### MeshService propagation tracking switched to hashed keys

Auto-generated `MeshService` resources track which non-system labels were copied from a `Dataplane` so that the next reconcile can remove labels whose source has gone away. Previously the tracking entry stored the raw key name as a Kubernetes label *value*, which silently skipped any qualified-name key containing `/` or `.` (e.g. `app.example.com/tier`). Such labels were copied onto the `MeshService` but never tracked, so they persisted after the carrier `Dataplane` was removed.

The control plane now stores a SHA-256 hash of the label key in the tracking label key (`kuma.io/pkey-<16hex>`). The old plain-value format is still recognised on read for one reconcile so the upgrade is seamless for keys the previous version was already able to track.

**Action required:**

None for valid label keys. Labels with `/` or `.` in the key that were leaked onto an auto-generated `MeshService` by the previous code cannot be reattributed retroactively — they have no tracking record at all and the new code preserves them as operator-managed. Remove any such leaked labels by hand if needed:

```sh
kubectl -n kuma-system label meshservice <name> app.example.com/tier-
```

### MeshAccessLog OpenTelemetry attribute keys are now validated

`MeshAccessLog` now validates `openTelemetry.attributes[].key` against the access-log attribute-key grammar used by this policy. Keys must start with a lowercase letter, use only lowercase letters, digits, `_` or `.`, avoid consecutive delimiters, end with a letter or digit, and must not use the reserved `otel.` prefix. `%...%` placeholders remain supported in attribute values, but are no longer accepted in keys.

Existing policies with invalid keys keep their current runtime behavior until they are updated or reapplied. In particular, placeholder-based keys continue to emit the same interpolated key after a control-plane upgrade. GitOps or other reconcilers that re-apply `MeshAccessLog` resources after the upgrade hit the same validation path immediately, so invalid keys must be fixed before the next reconcile. Any create or update using an invalid key is rejected until the key is renamed to a static value.

**Action required:**

Before the next apply, review the `MeshAccessLog` resources you manage using the tooling and workflow standard for your environment. On Kubernetes, audit `MeshAccessLog` resources across all namespaces. On Universal, audit the resources for each mesh you manage. If GitOps or another reconciler is the source of truth, review the manifests that will be re-applied as well.

When auditing `openTelemetry.attributes[].key`, flag any key that:

- starts with the reserved `otel.` prefix
- contains `%...%` placeholders
- does not start with a lowercase letter
- contains characters other than lowercase letters, digits, `_`, or `.`
- contains consecutive delimiters
- ends with a delimiter

Capture enough context to update each invalid policy before the next reconcile, for example the Kubernetes namespace, mesh, and resource name.

Then rename invalid keys such as `%KUMA_ZONE%`, `request-id`, `Service.Version`, or `otel.attribute`, and keep the dynamic content in the value instead:

```yaml
attributes:
  - key: service.zone
    value: "%KUMA_ZONE%"
```

### Readiness reporter is now TCP-only

The kuma-dp readiness reporter no longer listens on a Unix domain socket. `/ready` is served exclusively on TCP `KUMA_READINESS_PORT` (default `9902`) in both Kubernetes and Universal mode.

**What changed:**
- Removed config field `dataplane.readinessUnixSocketDisabled` and env var `KUMA_READINESS_UNIX_SOCKET_DISABLED`.
- New DPs no longer advertise the `feature-readiness-unix-socket` flag. The CP still honors it for older DPs during CP-first upgrades; the flag will be removed in a future release.
- The K8s injector no longer injects `KUMA_READINESS_UNIX_SOCKET_DISABLED=true` (the env var is now a no-op).
- The Helm ingress/egress chart templates no longer set these env vars.

**Action required:**

- Universal-mode operators who probed readiness via the Unix socket must switch to TCP loopback: `curl http://localhost:9902/ready`.
- Universal-mode hosts running more than one `kuma-dp` instance must assign a distinct `KUMA_READINESS_PORT` per instance. Each instance previously used its own Unix socket; they now all default to TCP `9902` and will fail to bind on conflict:

  ```sh
  KUMA_READINESS_PORT=9902 kuma-dp run ...   # instance 1
  KUMA_READINESS_PORT=9903 kuma-dp run ...   # instance 2
  ```
- Custom manifests that still set `KUMA_READINESS_UNIX_SOCKET_DISABLED` can leave the env var in place — it is ignored — or remove it.

### dp-server graceful shutdown is now time-bounded

The dp-server's graceful shutdown is now bounded by a configurable timeout. Previously the HTTP server would wait indefinitely for xDS streams to drain, which could keep the pod from exiting within its `terminationGracePeriodSeconds` and surface as a non-zero exit.

**What changed:**
- New config: `dpServer.gracefulShutdownTimeout` (default `10s`).
- Env var: `KUMA_DP_SERVER_GRACEFUL_SHUTDOWN_TIMEOUT`.
- Once the timeout expires, dp-server force-stops the gRPC server (aborting any stream whose handler ignored cancellation) and lets the pod exit cleanly.

**Action required:**

None for default deployments. The 10s default sits inside controller-runtime's 30s shutdown budget and the chart's `controlPlane.terminationGracePeriodSeconds=30`.

If you previously raised `terminationGracePeriodSeconds` to absorb long xDS drains, raise this timeout in lockstep, e.g.:

```sh
KUMA_DP_SERVER_GRACEFUL_SHUTDOWN_TIMEOUT=60s
```

Keep it strictly smaller than `terminationGracePeriodSeconds` so the bounded shutdown can run before the pod is killed.

### `MeshMultiZoneService` names longer than 63 characters are deprecated

`MeshService` and `MeshExternalService` already reject names longer than 63 characters.
`MeshMultiZoneService` (MMZS) historically allowed up to 253, but its name is used to render DNS hostnames through `HostnameGenerator` (e.g. `{{ .DisplayName }}.mzsvc.mesh.local`), so any name longer than 63 produces an invalid RFC 1035 label.

The control plane now emits a deprecation warning in the API response when an MMZS is created or updated with a name longer than 63 characters.
This will become a hard validation error in 3.0.

**Action required:**

Rename any `MeshMultiZoneService` whose name exceeds 63 characters.

### Kubernetes native sidecar containers are now the only injection mode

The `experimental.sidecarContainers` opt-out (Helm value and `KUMA_EXPERIMENTAL_SIDECAR_CONTAINERS` env var) has been removed. Kubernetes native sidecar containers (init containers with `restartPolicy: Always`, GA since Kubernetes 1.29) are now unconditionally used.

**What changed:**
- The Helm value `experimental.sidecarContainers` and the `KUMA_EXPERIMENTAL_SIDECAR_CONTAINERS` env var no longer exist; setting either has no effect.
- The control plane still detects the Kubernetes server version and automatically falls back to legacy (non-native) injection on clusters older than 1.29, but this is no longer configurable.

**Action required:**

None for Kubernetes 1.29+, which is already below Kuma's own minimum tested/supported Kubernetes version. If you were relying on `sidecarContainers: false` to opt out, that is no longer possible; remove the setting from your Helm values or environment.

### DNS server domain must not start with a dot

The control plane now rejects a DNS server domain configuration that begins with a `.` (e.g., `.mesh`). Previously such a value was silently accepted and produced broken hostname generation.

**Action required:**

If you have `KUMA_DNS_SERVER_DOMAIN` or `dnsServer.domain` set to a value starting with `.`, remove the leading dot:

```yaml
# kuma-cp config
dnsServer:
  domain: mesh   # was: .mesh
```

Or via environment variable: `KUMA_DNS_SERVER_DOMAIN=mesh`

### HostnameGenerator templates that produce invalid DNS names now fail explicitly

`HostnameGenerator` templates that evaluate to an invalid DNS subdomain (e.g., starting with a dot, containing uppercase letters, or having consecutive dots) now return an error instead of silently generating a broken hostname.

**Action required:**

Review your `HostnameGenerator` resources and ensure their `spec.template` values produce valid [RFC 1123](https://tools.ietf.org/html/rfc1123) DNS subdomains for all inputs.

### localhost-admin is restricted to direct loopback; CORS is now opt-in

The defaults have been tightened:

- `LocalhostIsAdmin` still defaults to `true`, but only direct loopback requests are promoted to admin.
- `CorsAllowedDomains` now defaults to `[]` / empty (was `[".*"]`).

The `LocalhostIsAdmin` restriction also applies to release branches as a security fix: the authenticator now only grants admin when the request is a **direct** loopback call (loopback `RemoteAddr`, loopback `Host`, no proxy-hop headers, and a matching `Origin` if present). Browsers connecting over loopback from a non-localhost page are no longer promoted to admin.

**Action required:**

_Local bootstrap / development (Universal mode)_

If you rely on `LocalhostIsAdmin` for initial kumactl setup, keep using direct loopback access or switch to token-based authentication with `kumactl config control-planes add --auth-type=tokens`.

_Reverse-proxy / CORS users_

If your deployment relies on cross-origin API access (e.g., a custom GUI on a different port), set the allowed domains explicitly:

```yaml
# kuma-cp config
apiServer:
  corsAllowedDomains:
    - "https://my-gui.example.com"
```

or via environment variable:

```sh
KUMA_API_SERVER_CORS_ALLOWED_DOMAINS=https://my-gui.example.com
```

The Helm chart already sets `KUMA_API_SERVER_AUTHN_LOCALHOST_IS_ADMIN=false` and is not affected by the `LocalhostIsAdmin` default.

### CPU limits removed from `kuma-init` and `kuma-sidecar` containers

The default CPU limit for injected `kuma-init`, `kuma-sidecar` and `kuma-validation` containers has been removed (set to `0`, meaning no limit). Previously the defaults were `100m` and `1000m` respectively.

**Why:** 

CPU limits cause throttling even when CPU is available, which increases latency under load. Removing the limit allows the containers to burst during startup and high-traffic periods.

**Action required:** 

None for most users. If your cluster enforces CPU limits, either relax those policies or set explicit limits:

**Kubernetes (Helm)**
```yaml
dataPlane:
  initContainer:
    resources:
      limits:
        cpu: 100m
  sidecarContainer:
    resources:
      limits:
        cpu: 1000m
  validationContainer:
    resources:
      limits:
        cpu: 100m
```

**Control plane environment variables**
```sh
KUMA_INJECTOR_INIT_CONTAINER_RESOURCES_LIMITS_CPU=100m
KUMA_INJECTOR_SIDECAR_CONTAINER_RESOURCES_LIMITS_CPU=1000m
KUMA_INJECTOR_VALIDATION_CONTAINER_RESOURCES_LIMITS_CPU=100m
```

### Envoy admin API now uses Unix domain socket by default

The Envoy admin API (`localhost:9901`) now binds to a Unix domain socket instead of TCP by default. This eliminates the shared-network-namespace attack vector where a compromised app container could reach the admin API to kill the sidecar, dump config, or modify runtime behavior.

**What changed:**
- `KUMA_BOOTSTRAP_SERVER_PARAMS_ENVOY_ADMIN_UNIX_SOCKET` defaults to `true`
- Helm value: `experimental.envoyAdminUnixSocket` (default `true`)
- A readiness reporter on TCP port 9902 handles K8s probes and lifecycle hooks
- Port 9902 must not conflict with application ports in sidecar-injected pods

**Action required:**

If you have tooling that directly accesses the Envoy admin API on `localhost:9901` (e.g., scripts calling `/config_dump`, `/stats`, `/clusters`), you need to either:
- Use the readiness reporter proxy on port 9902 instead, OR
- Use `curl --unix-socket /tmp/kuma-dp-*/kuma-envoy-admin.sock http://localhost/...`

**How to disable (revert to TCP admin):**

**Kubernetes (Helm)**
```yaml
experimental:
  envoyAdminUnixSocket: false
```

**Universal**
```sh
KUMA_BOOTSTRAP_SERVER_PARAMS_ENVOY_ADMIN_UNIX_SOCKET=false kuma-cp run
```

### Observability: CP metrics OTLP push enabled by default

When `OTEL_EXPORTER_OTLP_ENDPOINT` is set, the control plane now automatically pushes all CP Prometheus metrics to the configured OTLP collector. This is enabled by default.

**What changed:**
- A new config field `metrics.openTelemetry.enabled` (default `true`) gates CP metrics OTLP push.
- If you already set `OTEL_EXPORTER_OTLP_ENDPOINT` for tracing, metrics will now also be pushed to that endpoint.

**Action required:**

If you use `OTEL_EXPORTER_OTLP_ENDPOINT` for tracing only and do not want CP metrics pushed via OTLP, disable it:

```yaml
# kuma-cp config
metrics:
  openTelemetry:
    enabled: false
```

Or via environment variable: `KUMA_METRICS_OPENTELEMETRY_ENABLED=false`

### Observability: Prometheus metrics migration from Summary to Histogram

Internal Kuma Prometheus metrics changed from `Summary` to `Histogram` types to fix stale values that accumulated over long windows.

**What changed:**
- All metrics previously exported as `prometheus.Summary` or `prometheus.SummaryVec` are now `prometheus.Histogram` or `prometheus.HistogramVec`.
- Metrics that previously exposed quantiles (`0.5`, `0.9`, `0.99`) now use histogram buckets (`_bucket` with `le` labels).
- `DefaultObjectives` is replaced by `DefaultBuckets`.

**Action required:**
- Update any dashboards, alerts, or queries that use the old quantile representations to use histogram buckets and the `histogram_quantile` function.

### RBAC: Added `events.k8s.io` API group

The control plane ClusterRole and the namespaced Role used by the control plane for `events` now include the `events.k8s.io` API group alongside the core (`""`) API group for `events` resources. This aligns with the Kubernetes `events.k8s.io/v1` API which replaced the deprecated core `v1` Events API.

**Action required:**

If you manage RBAC resources outside of Helm (e.g., via GitOps or manual manifests), update your RBAC rules for events in both ClusterRole and Role definitions to include the `events.k8s.io` API group for events resources.

### RBAC: Added Role for mesh-scoped zone proxy rollout job

When deploying mesh-scoped zone proxies via the `meshes` list in `values.yaml`, a new post-install hook job is created to rollout restart zone proxy deployments after the control plane becomes available. This job requires a new Role (namespaced to the release namespace) with permissions to `get`, `list`, and `patch` deployments in the `apps` API group.

**Action required:**

If you manage RBAC resources outside of Helm (e.g., via GitOps or manual manifests) and use mesh-scoped zone proxies, add the following Role in your release namespace:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: <chart-name>-rollout-zoneproxy-job
  namespace: <release-namespace>
rules:
  - apiGroups:
      - "apps"
    resources:
      - deployments
    verbs:
      - get
      - list
      - patch
```

Along with the corresponding ServiceAccount and RoleBinding in the same namespace.

## Upgrade to `2.13.x`

### Strict Inbound Port Filtering Enabled by Default

Strict inbound port filtering is now enabled by default to improve security. When transparent proxy is enabled, the sidecar will only accept inbound traffic on ports that are explicitly defined in the dataplane configuration. This prevents unauthorized access to services through undeclared ports.

**What changed:**
- `strictInboundPortsEnabled` now defaults to `true` in `kuma-dp` configuration
- Inbound passthrough listeners now filter traffic to only allow explicitly configured ports
- This applies to both IPv4 and IPv6 traffic

**Action required:**

No action is required for most users. However, if you have services that:
1. Accept traffic on ports not explicitly declared in your dataplane configuration, AND
2. Rely on transparent proxy to route this traffic

You will need to either:
- Add the missing ports to your dataplane inbound configuration, OR
- Disable strict inbound port filtering (see below)

**How to disable strict inbound port filtering:**

If you need to disable this feature and revert to the previous behavior:

**Kubernetes**

Create a `ContainerPatch`:

```yaml
apiVersion: kuma.io/v1alpha1
kind: ContainerPatch
metadata:
  name: disable-strict-inbounds-ports
  namespace: kuma-system
spec:
  sidecarPatch:
  - op: add
    path: /env/-
    value: '{
      \"name\": \"KUMA_DATAPLANE_RUNTIME_STRICT_INBOUND_PORTS_ENABLED\",
      \"value\": \"false\"
    }'
```

Then set the annotation `kuma.io/container-patches` on deployment where it should be disabled

```yaml
"kuma.io/container-patches":"disable-strict-inbounds-ports"
```

or for all Deployments by setting control-plane configuration:

```
KUMA_RUNTIME_KUBERNETES_INJECTOR_CONTAINER_PATCHES="disable-strict-inbounds-ports"
```

**Universal**

Set the environment variable when running `kuma-dp`:

```bash
KUMA_DATAPLANE_RUNTIME_STRICT_INBOUND_PORTS_ENABLED=false kuma-dp run ...
```

**Security recommendation:**

We strongly recommend keeping strict inbound port filtering enabled. If you need to disable it temporarily, please audit your dataplane configurations to ensure all required inbound ports are explicitly declared, then re-enable the feature.

### Virtual Probes Disabled by Default

Virtual Probes are now disabled by default. This feature was deprecated in version 2.9.x in favor of Application Probe Proxy, which provides broader support for different probe types (HTTPGet, TCPSocket, and gRPC).

**What changed:**
- `virtualProbesEnabled` now defaults to `false` (previously `true`)
- Application Probe Proxy remains enabled by default on port 9001

**Action required:**

If you still rely on Virtual Probes and want to keep using them:

1. Enable Virtual Probes explicitly in control plane configuration:

   ```yaml
   runtime:
     kubernetes:
       injector:
         virtualProbesEnabled: true
   ```

2. Or via environment variable:

   ```bash
   KUMA_RUNTIME_KUBERNETES_VIRTUAL_PROBES_ENABLED=true
   ```

**How to disable Application Probe Proxy:**

If you need to disable Application Probe Proxy entirely:

1. Set `kuma.io/virtual-probes: disabled` annotation on your pods
2. Gateway mode automatically disables it

When both Virtual Probes and Application Probe Proxy are not explicitly configured, Application Probe Proxy is enabled by default.

**Migration recommendation:**

We strongly recommend migrating to Application Probe Proxy, which is the supported solution going forward. Virtual Probes will be removed in a future release. Application Probe Proxy works automatically with no configuration changes required.

### Go Module Path Migration

**Breaking Change for Library Users**

If you use Kuma as a Go library or have custom extensions, the module path has changed from `github.com/kumahq/kuma` to `github.com/kumahq/kuma/v2`.

**What changed:**
- Go module path now includes `/v2` suffix following Go modules semantic versioning conventions
- All package imports must be updated

**Action required:**

Update all import statements in your code:

**Before:**
```go
import "github.com/kumahq/kuma/pkg/core/resources/model"
import "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
```

**After:**
```go
import "github.com/kumahq/kuma/v2/pkg/core/resources/model"
import "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
```

Run `go mod tidy` after updating imports.

**Who is affected:**
- Anyone building custom Kuma plugins or extensions
- Downstream projects that import Kuma packages
- Custom operators or controllers integrating with Kuma

### Removal of ServiceAccountName Configuration

The `ServiceAccountName` configuration option has been removed after deprecation in version 2.6.x.

**What changed:**
- `KUMA_RUNTIME_KUBERNETES_SERVICE_ACCOUNT_NAME` environment variable is no longer supported
- `runtime.kubernetes.serviceAccountName` config option removed

**Action required:**

If you were using custom ServiceAccountName configuration, migrate to `AllowedUsers` instead:

```yaml
runtime:
  kubernetes:
    allowedUsers:
      - "system:serviceaccount:kuma-system:kuma-control-plane"
```

Or via environment variable:
```bash
KUMA_RUNTIME_KUBERNETES_ALLOWED_USERS="system:serviceaccount:kuma-system:kuma-control-plane"
```

This is already configured by default via Helm template with `KUMA_RUNTIME_KUBERNETES_ALLOWED_USERS`.

### MeshTrust spec.origin Field Removed

The `spec.origin` field in MeshTrust resources has been removed. The origin is
now reported through `status.origin`.

**What changed:**
- `spec.origin` is no longer part of the MeshTrust API or Kubernetes CRD schema
- Manifests and API requests that still set `spec.origin` can be rejected as unknown-field input
- The field is automatically populated in `status.origin` by the MeshIdentity status updater

**Action required:**

Before upgrading, remove `spec.origin` from MeshTrust manifests and update any
automation or tooling that reads it to use `status.origin` instead.

**Example:**

**Before:**
```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshTrust
spec:
  origin: MeshIdentity
  # ...
```

**After (read from status instead):**
```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshTrust
spec:
  # origin no longer set in spec
  # ...
status:
  origin: MeshIdentity  # automatically populated
```

## Upgrade to `2.12.x`

### Helm: Configurable hook job TTL and annotations

Hook job templates (`pre-install`, `pre-upgrade`, `pre-delete`, `post-delete`, `post-install`) were modified to make `ttlSecondsAfterFinished` and job annotations configurable. This fixes ArgoCD deployments getting stuck during upgrades: with `ttlSecondsAfterFinished: 0`, Kubernetes deleted hook Jobs immediately upon completion before ArgoCD could read their status, causing syncs to hang indefinitely.

New values:
- `hooks.ttlSecondsAfterFinished` (default: `0`) — set to `null` to disable TTL deletion
- `hooks.annotations` (default: `{}`) — extra annotations merged into all hook Job metadata

**Recommended ArgoCD configuration:**
```yaml
hooks:
  ttlSecondsAfterFinished: null
  annotations:
    argocd.argoproj.io/hook-delete-policy: BeforeHookCreation
```

**Action required:** None for non-ArgoCD users. Default behaviour is preserved.

### Removal of `/status/zones` endpoints

These endpoints were deprecated, and are now removed. You can achieve the same functionality with `/zones/_overview`.

### Deprecation of readiness reporter TCP port in favor of Unix socket

> **Note:** This deprecation was reversed in `2.14.x`. The readiness reporter is now TCP-only and the Unix socket has been removed. See the `2.14.x` notes above.

It is no longer possible to disable the readiness reporter, which means TCP port 0 is not allowed to be used.

## Upgrade to `2.11.x`

### Helm upgrades to 2.11.8 require explicit `namespaceAllowList: []` in values.yaml

If a user upgrades to 2.11.8 (or earlier 2.11.x patch versions) with `--reuse-values` Helm flag, the upgrade fails with:

```
Error: UPGRADE FAILED: template: kuma/templates/cp-webhooks-and-secrets.yaml:69:17: executing "kuma/templates/cp-webhooks-and-secrets.yaml" at <len .Values.namespaceAllowList>: error calling len: len of nil pointer
```

As a workaround, add `namespaceAllowList: []` to `values.yaml`. This behaviour is fixed starting from version 2.11.9.

### Embedded Proxy DNS is Enabled by Default

In version `2.11.x`, we reimplemented how mesh DNS queries are resolved by replacing CoreDNS with our Embedded DNS Server. This server is built into `kuma-dp`. After a pod restart (in Kubernetes) or an upgrade of `kuma-dp` (in Universal mode) to `2.11.x`, the embedded DNS proxy is enabled by default.

If you encounter any issues, you can disable it as follows:

**Kubernetes**
Disable it by setting the environment variable when deploying `kuma-cp`:
```bash
KUMA_RUNTIME_KUBERNETES_INJECTOR_BUILTIN_DNS_EXPERIMENTAL_PROXY=false
```

Or via configuration:

```yaml
runtime:
  kubernetes:
    injector:
      builtinDns:
        experimentalProxy: false
```

**Universal**

Disable it by running `kuma-dp` with the following environment variable:

```bash
KUMA_DNS_PROXY_PORT=0
```

### `kuma-sidecar` container has `allowPrivilegeEscalation` set to `false`

In previous versions, Kuma did not explicitly set `allowPrivilegeEscalation`. Starting with this version, it is now explicitly set to `false`.

Before upgrading, ensure that your configuration does not override this setting.

### System namespace requires the `kuma.io/sidecar-injection: false` label

To simplify the namespace selector logic in webhooks, we now require the `kuma.io/sidecar-injection: false` label to be set on Kuma's system namespace (`kuma-system` by default).

Since Kubernetes v1.22, the API server automatically adds the `kubernetes.io/metadata.name` label to all namespaces. As a result, we’ve replaced the use of the custom `kuma.io/system-namespace` label in the secret webhook selector with this standard label.

If you are running helm with `noHelmHooks` please set label on the system namespace:

```bash
kubectl label namespace SYSTEM_NAMESPACE kuma.io/sidecar-injection=disabled
```

### More Restricted `ClusterRole` for Control Plane and CNI

We have split the `ClusterRole` for the control plane into two parts:

* A cluster-scoped `ClusterRole` with read access to namespaced resources.
* A `ClusterRole` with write permissions, now scoped more narrowly.

By default, a `ClusterRoleBinding` is used to grant write permissions to the control plane, and no action is required from the user. However, if you want the control plane to have access only in specific namespaces, you can use the `namespaceAllowList` configuration to define where it should have write permissions.

### Namespaces that are part of the mesh requires `kuma.io/sidecar-injection` label to exist

Since version 2.11.x, to improve performance and security, each namespace participating in the mesh is required to have the `kuma.io/sidecar-injection` label set.

Before upgrading, check whether any deployments are using the `kuma.io/sidecar-injection: true` or `enabled` label in namespaces that do not have the `kuma.io/sidecar-injection` label set. If so, add `kuma.io/sidecar-injection: disabled` to those namespaces.

As a one-time fix, you can use this script to detect such namespaces by looking for running mesh-enabled pods:

```bash
for ns in $(kubectl get ns -o jsonpath='{.items[*].metadata.name}'); do
  ns_label=$(kubectl get ns "$ns" -o jsonpath='{.metadata.labels.kuma\.io/sidecar-injection}' 2>/dev/null)
  if [ -z "$ns_label" ]; then
    kubectl get pods -n "$ns" --show-labels --no-headers 2>/dev/null | \
      grep 'kuma.io/sidecar-injection' | \
      awk -v ns="$ns" '{print ns "/" $1}'
  fi
done
```

You can later patch namespaces with the following command:

```bash
kubectl label namespace NAMESPACE_NAME kuma.io/sidecar-injection=disabled
```

It's recommended to update your workflow that creates namespaces to include this label.

### Fixed: Extra Newlines in MeshAccessLog with TCP Backends

This fix affects users of MeshAccessLog policies that send logs to a TCP backend using the `plain` format 
(or where the format type was not explicitly set):

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshAccessLog
spec:
  to:
    - targetRef:
        kind: Mesh
      default:
        backends:
          - type: Tcp # <--- backend type is 'Tcp'
            tcp:
              format:
                type: Plain 
                plain: '...' # <--- the format is 'plain' or unspecified
              address: "%s:9999"
```

In previous versions of Kuma, logs sent with this setup included **an unintended double newline** between entries, producing output like:

```
[2025-05-20T14:03:17.123Z] "GET /api/v1/users HTTP/1.1" 200 - "-" "curl/8.1.2" 0 123 45 44 "192.168.1.10" "service-backend" "cluster-backend" "10.0.0.15:8080" "envoy-router" - default

[2025-05-20T14:03:18.456Z] "POST /auth/login HTTP/1.1" 401 - "-" "Mozilla/5.0" 0 98 56 55 "192.168.1.11" "auth-service" "cluster-auth" "10.0.0.20:9090" "envoy-router" - default

[2025-05-20T14:03:19.789Z] "GET /metrics HTTP/1.1" 200 - "-" "Prometheus/2.47.0" 0 6789 12 12 "127.0.0.1" "metrics-service" "cluster-metrics" "127.0.0.1:9100" "envoy-router" - default

```
This has now been corrected—each log line will end with a single newline `\n`, as intended.

> **Note:** If your logging backend or tooling relied on the previous double-newline behavior (e.g. for framing or parsing), 
> you can preserve it by manually adding `\n` at the end of your `plain` format string.

## Upgrade to `2.10.x`

### API Server behaviour changes

The response of successful `DELETE` or `PUT` requests without warnings will now include "content-type: application/json" in the header and return an empty JSON object as the body.

### Stricter validation rules for resource names

In Kuma 2.7.x we deprecated usage of non [RFC 1123](https://www.rfc-editor.org/rfc/rfc1123.html) characters, and start from 2.10.x it's no longer possible to create or update non RFC compliant resource names. In order to be compatible with Kubernetes naming policy we updated the validation rules. 

Old rule:

> Valid characters are numbers, lowercase latin letters and '-', '_' symbols.

New rule:

> A lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character

Before the upgrade ensure that your resources don't use unsupported characters.

#### Add RFC-1035-Label constraints for specific resources names

Starting from version 2.10.x, we are deprecating the usage of non [RFC 1035](https://www.rfc-editor.org/rfc/rfc1035.html) characters for names of `Mesh`, `Zone`, `MeshService`, `MeshExternalService`, `MeshMultizoneService` resources. These names will be rejected in the future.

Old rule:

> Valid characters are numbers, lowercase latin letters and '-', '_' symbols.

New rule:

> * Characters contain at most 63 characters.
> * Characters contain only lowercase alphanumeric characters or '-'.
> * Characters start with an alphabetic character.
> * Characters end with an alphanumeric character.

### On Universal, a MeshService is no longer generated for any dataplane that has an invalid `kuma.io/service` tag

When using `MeshService`, we were automatically generating `MeshService` resources in Universal mode. However, due to stricter resource validation, we have decided not to generate the `MeshService` resource when a dataplane has a `kuma.io/service` that is not [RFC 1035](https://www.rfc-editor.org/rfc/rfc1035.html) compliant.

If you want the control plane on Universal to autogenerate `MeshService` resources, update the `kuma.io/service` tags to valid names. Otherwise, you must create them manually.

Old rule:

> Valid characters are numbers, lowercase latin letters and '-', '_' symbols.

New rule:

> A lowercase RFC 1035 Label Names must have at most 63 characters and consist of lower case alphanumeric characters or '-', and must start with an alphabetic character, end with an alphanumeric character.

### New kind Dataplane for targetRef in policies

This version introduced new `Dataplane` kind for top level `targetRef` in policies. Dataplane will replace `MeshSubset` kind.
Making `MeshSubset` deprecated. Dataplane selects Dataplane resources by its labels and adds possibility to select single inbound
by using `sectionName` field, as opposed to old MeshSubset which was selecting proxies by inbound tags. More detailed info in docs.

### New Rules API

This version adds new `rules` api that replaces `from` section in policies. Making `from` deprecated. Support was added for
policies:
- MeshAccessLog
- MeshCircuitBreaker
- MeshRateLimit
- MeshTimeout
- MeshTls

You cannot combine inbound configuration with outbound traffic configuration in policies when using new Rules API.
If you have old policies with both `from` and `to` you need to split them into separate policies before migrating to `rules`.

### MeshHTTPRoute

#### Unifying defaults for `statusCode`

Due to misconfiguration, a default for `statusCode` for http route on Universal could have been missing.
If you're using Universal mode, and you did not specify `default.filters[].requestRedirect.statusCode` value in your `MeshHTTPRoute` resource, you have to explicitly set it to 302.

### MeshTrace

#### Unifying defaults for `apiVersion`

Due to misconfiguration, a default for `apiVersion` for traces on Universal could have been missing.
If you're using Universal mode, and you did not specify `tracing.backends[].conf.apiVersion` value in your `MeshTrace` resource, you have to explicitly set it to "httpJson".

#### Unifying defaults for `sharedSpanContext`

Due to misconfiguration, a default `sharedSpanContext` for traces on Universal ("false") was different from on Kubernetes ("true").
If you're using Universal mode, and you did not specify `tracing.backends[].conf.sharedSpanContext` value in your `MeshTrace` resource, you have to explicitly set it to "false" to continue using that value.

### MeshMetric

#### Unifying defaults for `path`

Due to misconfiguration, a default `path` for metrics on Universal ("/metrics") was different from on Kubernetes ("/metrics/prometheus").
If you're using Universal mode, and you did not specify `default.applications[].path` value in your `MeshMetric` resource, you have to explicitly set it to "/metrics" to continue using that value.

### MeshPassthrough

#### Unifying defaults for `passthroughMode`

Due to misconfiguration, a default `passthroughMode` for `MeshPasstrhough` on Universal ("Matched") was different from on Kubernetes ("None").
If you're using Kubernetes mode, and you did not specify `default.passthroughMode` value in your `MeshPasstrhough` resource, you have to explicitly set it to "None" to continue using that value.

### MeshLoadBalancingStrategy

#### Removal of `hashPolicies.type: SourceIP`

The `SourceIP` hash policy type has been removed. If you are using `type: SourceIP` in your `MeshLoadBalancingStrategy` policy, update it to use
`type: Connection` with `connection.sourceIP: true` instead.

### Built-in MeshGateway policy targeting

Policies no longer attach directly to built-in `MeshGateway` listeners through `spec.targetRef.kind: MeshGateway` for `MeshAccessLog`, `MeshCircuitBreaker`, `MeshFaultInjection`, `MeshHealthCheck`, `MeshRateLimit`, `MeshRetry`, `MeshTimeout`, `MeshTLS`, `MeshTrace`, or `MeshLoadBalancingStrategy`.
Use `MeshHTTPRoute` or `MeshTCPRoute` resources for built-in gateway routing behavior, and move the affected policy configuration to supported target kinds before upgrading.

### Legacy `FaultInjection` policy no longer generates Envoy configuration

The legacy `FaultInjection` policy (`kind: FaultInjection`, superseded by `MeshFaultInjection`) no longer produces any Envoy fault-injection filters. Existing `FaultInjection` resources remain creatable and readable through the API, but they have no effect on data plane proxies.

**Action required**

Migrate any remaining `FaultInjection` resources to `MeshFaultInjection` before upgrading. See the [MeshFaultInjection documentation](https://kuma.io/docs/latest/policies/meshfaultinjection/) for the equivalent configuration.

### MeshHealthCheck

#### Deprecation of `healthyPanicThreshold` for `MeshHealthCheck`

The `healthyPanicThreshold` field in the `MeshHealthCheck` policy is deprecated in favor of the `MeshCircuitBreaker` policy. It has since been removed — see [`healthyPanicThreshold` removed from `MeshHealthCheck`](#healthypanicthreshold-removed-from-meshhealthcheck).

### Changes on revoking dataplane tokens

Authentication between the control plane and dataplanes is now only checked at connection start. This means if a token expires or is revoked after the dataplane connects, the connection won't stop. The recommended action on token revocation is to restart either the control plane or the concerned dataplanes.

## Upgrade to `2.9.x`

### MeshAccessLog

Policies targeting `spec.targetRef.kind: MeshGateway` can now only target `kind: Mesh` in
`to[].targetRef`. Previously MeshService, MeshExternalService, MeshMultiZoneService were allowed but the resulting configuration
was ambiguous and nondeterministic.

### MeshLoadBalancingStrategy

Policies targeting `spec.targetRef.kind: MeshGateway` and setting the `spec.loadBalancer` field can now only target `kind: Mesh` in
`to[].targetRef`. Previously MeshService, MeshExternalService, MeshMultiZoneService were allowed but the resulting configuration
was ambiguous and nondeterministic.

### MeshTimeout

The default inbound request timeout is now 15s instead of being unbounded.
If you need longer inbound timeouts make sure to create a policy to override it.

For example:

```yaml
type: MeshTimeout
name: mt-higher-inbound
mesh: mesh-1
spec:
  targetRef:
    kind: Mesh
  from:
    - targetRef:
        kind: Mesh
      default:
        http:
          requestTimeout: 60s
```

### MeshExternalService

#### Removal of unix sockets support

It's no longer possible to define a unix domain socket on the `address` field.

### Upgrading Transparent Proxy Configuration

#### Removal of Deprecated IPv6 Redirection Flag and Annotation

In this release, the following deprecated options for configuring IPv6 transparent proxy redirection have been removed:

- The `--redirect-inbound-port-ipv6` flag in `kumactl install transparent-proxy`.
- The `kuma.io/transparent-proxying-inbound-v6-port` annotation.

Previously, disabling IPv6 transparent proxy redirection could be achieved by setting these options to `0`. This method is no longer supported.

To disable IPv6 transparent proxy redirection, you should now use the `--ip-family-mode` flag or the `kuma.io/transparent-proxying-ip-family-mode` annotation and set their value to `ipv4`. The default value for these options is `dualstack`.

**Example:**

In Universal mode, to install a transparent proxy:

```sh
kumactl install transparent-proxy --ip-family-mode ipv4 ...
```

In the definition of the `Dataplane` resource:

```yaml
type: Dataplane
mesh: default
name: dp-1
networking:
  # ...
  transparentProxying:
    redirectPortInbound: 15006
    redirectPortOutbound: 15001
    ipFamilyMode: ipv4
```

To set the configuration for Kubernetes workloads:

```sh
kumactl install control-plane --set controlPlane.envVars.KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_IP_FAMILY_MODE=ipv4 ...
```

or

```sh
helm install --set controlPlane.envVars.KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_IP_FAMILY_MODE=ipv4 ... kuma kuma/kuma
```

For more information about disabling IPv6 in transparent proxy redirection, visit our documentation: [Disabling IPv6](https://kuma.io/docs/2.8.x/production/dp-config/ipv6/#disabling-ipv6).

Please update your configurations accordingly to ensure a smooth transition and avoid any disruptions in your service.

#### Removal of `redirectPortInboundV6` Field from Dataplane Resource

The `Dataplane` resource no longer includes the `redirectPortInboundV6` field. Any configuration containing this field will fail validation. Update your `Dataplane` resources as shown below:

**Previous configuration:**

```yaml
type: Dataplane
mesh: default
name: dp-1
networking:
  # ...
  transparentProxying:
    redirectPortInbound: 15006
    redirectPortInboundV6: 15006
    redirectPortOutbound: 15001
```

**Updated configuration:**

```yaml
type: Dataplane
mesh: default
name: dp-1
networking:
  # ...
  transparentProxying:
    redirectPortInbound: 15006
    redirectPortOutbound: 15001
```

Ensure to update your Dataplane resources to the new format to avoid any validation errors.

#### Removal of Deprecated Exclude Outbound TCP/UDP Ports for UIDs Flags

The flags `--exclude-outbound-tcp-ports-for-uids` and `--exclude-outbound-udp-ports-for-uids` have been removed from the `kumactl install transparent-proxy` command. Users should now use the consolidated flag `--exclude-outbound-ports-for-uids <protocol:>?<ports:>?<uids>` instead.

##### Examples:

- To disable redirection of outbound TCP traffic on port 22 for users with UID 1000:
  ```sh
  kumactl install transparent-proxy --exclude-outbound-ports-for-uids tcp:22:1000 ...
  ```

- To disable redirection of outbound UDP traffic on port 53 for users with UID 1000:
  ```sh
  kumactl install transparent-proxy --exclude-outbound-ports-for-uids udp:53:1000 ...
  ```

#### Removal of Deprecated Exclude Outbound TCP/UDP Ports for UIDs Annotations

The annotations `traffic.kuma.io/exclude-outbound-tcp-ports-for-uids` and `traffic.kuma.io/exclude-outbound-udp-ports-for-uids` have also been removed. Use the annotation `traffic.kuma.io/exclude-outbound-ports-for-uids` instead.

##### Examples:

- To disable redirection of outbound TCP traffic on port 22 for users with UID 1000:
  ```yaml
  traffic.kuma.io/exclude-outbound-ports-for-uids: tcp:22:1000
  ```

- To disable redirection of outbound UDP traffic on port 53 for users with UID 1000:
  ```yaml
  traffic.kuma.io/exclude-outbound-ports-for-uids: udp:53:1000
  ```

Make sure to update your configuration files and scripts accordingly to accommodate these changes.

#### Deprecation of `--kuma-dp-uid` Flag

In this release, the `--kuma-dp-uid` flag used in the `kumactl install transparent-proxy` command has been deprecated. The functionality of specifying a user by UID is now included in the `--kuma-dp-user` flag, which accepts both usernames and UIDs.

**New Usage Example:**

Instead of using:
```sh
kumactl install transparent-proxy --kuma-dp-uid 1234
```

You should now use:
```sh
kumactl install transparent-proxy --kuma-dp-user 1234
```

If the `--kuma-dp-user` flag is not provided, the system will attempt to use the default UID (`5678`) or the default username (`kuma-dp`).

Please update your scripts and configurations accordingly to accommodate this change.

### Setting `kuma.io/service` in tags of `MeshGatewayInstance` had been forbidden

To increase security, in version 2.7.x, setting a `kuma.io/service` tag for the `MeshGatewayInstance` was deprecated and since 2.9.x is not supported. We generate the `kuma.io/service` tag based on the `MeshGatewayInstance` resource. The service name is constructed as `{MeshGatewayInstance name}_{MeshGatewayInstance namespace}_svc`.

E.g.:

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshGatewayInstance
metadata:
  name: demo-app
  namespace: kuma-demo
  labels:
    kuma.io/mesh: default
```

The generated `kuma.io/service` value is `demo-app_kuma-demo_svc`.

#### Migration

The migration process requires updating all policies and `MeshGateway` resources using the old `kuma.io/service` value to adopt the new one.

Migration step:
1. Create a copy of policies using the new `kuma.io/service` and the new resource name to avoid overwriting previous policies.
2. Duplicate the `MeshGateway` resource with a selector using the new `kuma.io/service` value.
3. Deploy the gateway and verify if traffic works correctly.
4. Remove the old resources.

### Introduction to Application Probe Proxy and deprecation of Virtual Probes

To support more types of application probes on Kubernetes, in version 2.9, we introduced a new feature named "Application Probe Proxy" which supports `HTTPGet`, `TCPSocket` and `gRPC` application probes. Starting from `2.9.x`, Virtual Probes is deprecated, and Application Probe Proxy is enabled by default.

Application workloads using Virtual Probes will be migrated to Application Probe Proxy automatically on next restart/redeploy on Kubernetes, without other operations. 

Application Probe Proxy will by default listen on port `9001`. To prevent potential conflicts with applications, you may customize this port using one of these methods:

1. Configuring on the control plane to apply on all dataplanes: set the port onto configuration key `runtime.kubernetes.injector.applicationProbeProxyPort` 
1. Configuring on the control plane to apply on all dataplanes: set the port using environment variable `KUMA_RUNTIME_KUBERNETES_APPLICATION_PROBE_PROXY_PORT` 
1. Configuring for certain dataplanes: set the port using pod annotation `kuma.io/application-probe-proxy-port`

By setting the port to `0`, Application Probe Proxy feature will be disabled, and when it's disabled, Virtual Probes still works as usual until the deprecation period ends.

Because of deprecation of Virtual Probes, the following items are considered deprecated:

- Pod annotation `kuma.io/virtual-probes`
- Pod annotation `kuma.io/virtual-probes-port`
- Control plane configuration key `runtime.kubernetes.injector.virtualProbesEnabled`
- Control plane configuration key `runtime.kubernetes.injector.virtualProbesPort`
- Control plane environment variable `KUMA_RUNTIME_KUBERNETES_VIRTUAL_PROBES_ENABLED`
- Control plane environment variable `KUMA_RUNTIME_KUBERNETES_VIRTUAL_PROBES_PORT`
- Data field `probes` on `Dataplane` objects

### kumactl

#### Default prometheus scrape config removes `service`

If you rely on a scrape config from previous version it's advised to remove the relabel config that was adding `service`.
Indeed `service` is a very common label and metrics were sometimes coliding with Kuma metrics. If you want the label `kuma_io_service` is always the same as `service`.

### Removal of KDS `KUMA_EXPERIMENTAL_KDS_DELTA_ENABLED` configuration option

In this release, KDS Delta is used by default and the CP environment variable `KUMA_EXPERIMENTAL_KDS_DELTA_ENABLED` doesn't exist anymore.

### Deprecation of `yes/no` values for annotation switches

The values `yes` and `no` are deprecated for specifying boolean values in switches based on pod annotations, and support for these values will be removed in a future release. Since these values were undocumented, they are not expected to be widely used.

Please use `true` and `false` as replacements; some boolean switches also support `enabled` and `disabled`. [Check the documentation](https://kuma.io/docs/latest/reference/kubernetes-annotations/) for the specific annotation to confirm the correct replacements.

#### Deprecation of `kuma.io/mesh` annotation

It was previously possible to create a resource in a `Mesh` by providing the `Mesh` name as an annotation, but this support has been deprecated and will be removed in the future.

Please use the `kuma.io/mesh` label instead.

## Upgrade to `2.8.x`

### MeshFaultInjection responseBandwidth.limit

With [#10371](https://github.com/kumahq/kuma/pull/10371) we have tightened the validation of the `responseBandwidth.limit` field in `MeshFaultInjection` policy. Policies with invalid values, such as `-10kbps`, will be rejected.

### MeshRetry tcp.MaxConnectAttempt

With [#10250](https://github.com/kumahq/kuma/pull/10250) `MeshRetry` policies with `spec.tcp.MaxConnectAttempt=0` will be rejected.
Prior to 2.8.x these were semantically valid but would create invalid Envoy configuration and would cause issues on the dataplane.
Now this is rejected sooner to avoid service disruption.

### Removal of legacy tokens

Tokens issued from versions before 2.1.x needs to renewed before upgrading.

If you observe following log in control-plane logs, please rotate your tokens before upgrade.
```yaml
[WARNING] Using token with KID header, you should rotate this token as it will not be valid in future versions of Kuma
```
* [User token](https://kuma.io/docs/2.7.x/production/secure-deployment/api-server-auth/)
* [Dataplane token](https://kuma.io/docs/2.7.x/production/secure-deployment/dp-auth/)
* [Zone token](https://kuma.io/docs/2.7.x/production/cp-deployment/zoneproxy-auth/#zone-token)

## Upgrade to `2.7.x`

### MeshMetric and cluster stats merging

For MeshMetric we disabled cluster [stats merging](https://github.com/kumahq/kuma/pull/9768) so that metrics are generated per [traffic split](https://kuma.io/docs/2.6.x/policies/meshhttproute/#traffic-split).
This means that in Grafana there will be at least two entries under "Destination service" - one for the service without a hash (e.g. `backend_kuma-demo_svc_3001`) and one per each split ending with a hash (e.g. `backend_kuma-demo_svc_3001-de1397ec09e96dfb`).
If you want to see combined metrics you can run queries with a prefix instead of exact match, e.g.:

```
... envoy_cluster_name=~"$destination_cluster.*" ...
```

instead of

```
... envoy_cluster_name="$destination_cluster" ...
```

To correlate between a hash and a particular pod you have to click on the outbound, and then click on "clusters" and associate pod ip with cluster ip.
This will be improved in the future by having the tags next to the outbound.
[This issue](https://github.com/kumahq/kuma-gui/issues/2412) tracks the progress of that as well as contains screenshots of the steps.

### MeshMetric `sidecar.regex` is replaced by `sidecar.profiles.exclude`

If you're using `sidecar.regex` field it is getting replaced by `sidecar.profiles.exclude`.
Replace usages of:

```yaml
...
  sidecar:
    regex: "my_match.*"
...
```

with:

```yaml
  sidecar:
    profiles:
      exclude:
        - type: Regex
          match: "my_match.*"
```

### Setting `kuma.io/service` in tags of `MeshGatewayInstance` is deprecated

To increase security, since version 2.7.x, setting a `kuma.io/service` tag for the `MeshGatewayInstance` is deprecated. If the tag is not provided, we generate the `kuma.io/service` tag based on the `MeshGatewayInstance` resource. The service name is constructed as `{MeshGatewayInstance name}_{MeshGatewayInstance namespace}_svc`.

E.g.:

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshGatewayInstance
metadata:
  name: demo-app
  namespace: kuma-demo
  labels:
    kuma.io/mesh: default
```

The generated `kuma.io/service` value is `demo-app_kuma-demo_svc`.

#### Migration

The migration process requires updating all policies and `MeshGateway` resources using the old `kuma.io/service` value to adopt the new one.

Migration step:
1. Create a copy of policies using the new `kuma.io/service` and the new resource name to avoid overwriting previous policies.
2. Duplicate the `MeshGateway` resource with a selector using the new `kuma.io/service` value.
3. Deploy the gateway and verify if traffic works correctly.
4. Remove the old resources.

### ZoneIngress Token support removed

The control-plane does not support tokens generated with `kumactl generate zone-ingress-token`. If you are running Kuma ingress with a zone ingress token generated using the deprecated method, before upgrading, verify if you are still using the old token.

#### How to validate if I am using `zone-ingress-token`?

1. Obtain the ingress token value
2. Run the following command
```bash
jq -R 'split(".") | .[0],.[1] | @base64d | fromjson' <<< $YOUR_TOKEN
```

Example output of a zone token:
```json
{
  "alg": "RS256",
  "kid": "1",
  "typ": "JWT"
}
{
  "Zone": "test",
  "Scope": [
    "ingress",
    "egress",
    "cp",
    "ratelimit"
  ],
  "exp": 1712414035,
  "nbf": 1709821735,
  "iat": 1709822035,
  "jti": "efeb8cca-2341-47a4-b4f2-daf49290e481"
}
```

Example output of a zone ingress token:
```json
{
  "alg": "RS256",
  "kid": "1",
  "typ": "JWT"
}
{
  "Zone": "test",
  "exp": 1709822002,
  "nbf": 1709821702,
  "iat": 1709822002,
  "jti": "c4cf30c5-ca30-42ec-b08d-de56fba75e7b"
}
```
3. If the output does not have the `Scope` field, you need to generate a new zone token using `kumactl generate zone-token` for your ingress before upgrading.
4. Restart the Ingress with the new token.
5. Now, you can safely upgrade the control-plane.

### Configuration option `KUMA_DP_SERVER_AUTH_*`, `dpServer.auth.*` was removed

The option to configure authentication was deprecated and has been removed in release `2.7.x`. If you are still using `KUMA_DP_SERVER_AUTH_*`
environment variables or `dpServer.auth.*` configuration, please migrate your configuration to use `dpServer.authn` before upgrade.

### Deprecation of `--redirect-inbound-port-v6` flag and `runtime.kubernetes.injector.sidecarContainer.redirectPortInboundV6` configuration option.

The `--redirect-inbound-port-v6` flag and the corresponding configuration option `runtime.kubernetes.injector.sidecarContainer.redirectPortInboundV6` are deprecated and will be removed in a future release of Kuma. These flags and configuration options were used to configure the port used for redirecting IPv6 traffic to Kuma.

In the upcoming release, Kuma will redirect IPv6 traffic to the same port as IPv4 traffic (15006). This means that you no longer need to configure a separate port for IPv6 traffic. If you want to disable traffic redirection for IPv6 traffic, you can set `--ip-family-mode ipv4`. We have also added a new configuration option `runtime.kubernetes.injector.sidecarContainer.ipFamilyMode` to switch traffic redirection for IP families.

We recommend that you update your configurations to use the new defaults for IPv6 traffic redirection. If you need to retain separate ports for IPv4 and IPv6 traffic, you can continue to use the deprecated flags and configuration options until they are removed.

### Deprecation of 'from[].targetRef.kind: MeshService'

At this moment only MeshTrafficPermission and MeshFaultInjection allowed `MeshService` in the `from[].targetRef.kind`.
Starting `2.7` this value is deprecated, instead the `MeshSubset` with `kuma.io/service` tag should be used. For example, instead of:

```yaml
type: MeshTrafficPermission
name: allow-orders
mesh: default
spec:
  targetRef:
    kind: Mesh
  from:
    - targetRef:
        kind: MeshService
        name: orders
      default:
        action: Allow
```

we should have:

```yaml
type: MeshTrafficPermission
name: allow-orders
mesh: default
spec:
  targetRef:
    kind: Mesh
  from:
    - targetRef:
        kind: MeshSubset
        tags:
          kuma.io/service: orders
      default:
        action: Allow
```

### Change in internal resources with Kubernetes Gateway API

This section describes changes to internal resources used by Kuma when configuring the built-in gateway using the Kubernetes Gateway API.

#### Prior Behavior (Before Kuma 2.7.0):

  * Applying a `Gateway` resource resulted in the creation of corresponding `MeshGateway` and `MeshGatewayInstance` resources.
  * An applied `HTTPRoute` resource was converted to a `MeshGatewayRoute` resource.

#### Changes Introduced in Kuma 2.7.0:

  * `HTTPRoute` resources are now converted to `MeshHTTPRoute` resources instead of `MeshGatewayRoute` resources.

#### Upgrade Impact:

  * Existing `MeshGatewayRoute` resources automatically created from `HTTPRoute` definitions will be deleted during the upgrade.
  * New `MeshHTTPRoute` resources will be created to replace the deleted ones.

#### Important Note:

This change is transparent with regard to the generated Envoy configuration. There should be no impact on existing traffic routing.

### Gateway API Promotion to GA

The Gateway API functionality within Kuma is now considered Generally Available (GA). This means the `--experimental-gatewayapi` flag and the `experimental.gatewayAPI` setting are no longer required for installation.

> [!WARNING]
> If you previously used the `--experimental-gatewayapi` flag with `kumactl install control-plane` in your workflows, it's important to note that this flag has been removed and is no longer supported. Using it will now result in an error.

#### Removed Flags and Settings

Previously, these flags were necessary for using the Gateway API feature:

- `--experimental-gatewayapi` flag for `kumactl install control-plane` and `kumactl install crds`
- `experimental.gatewayAPI=true` setting in both `kumactl install control-plane` and Helm charts

### TLS Secrets with Gateway API in namespace other than mesh system namespace

If you use TLS secrets with Gateway API for a builtin gateway deployed in any other namespace than mesh system namespace, set `controlPlane.supportGatewaySecretsInAllNamespaces` HELM value to true.
This change was introduced so that control plane does not have capability to read content of secrets in all namespaces by default.

## Upgrade to `2.6.x`

### Policy

#### Sorting

This change relates only to the new targetRef policies. When 2 policies have a tie on the targetRef kind we compare their names lexicographically.
Policy merging now gives precedence to policies that lexicographically "less" than other policies, i.e. policy "aaa" takes precedence over "bbb" because "aaa" < "bbb".
Previously, before 2.6.0 the order was the opposite.

#### `targetRef.kind: MeshGateway`

Note that when targeting `MeshGateways` you should be using `targetRef.kind:
MeshGateway`. Previously `targetRef.kind: MeshService` was necessary but this
left the control plane unable to fully validate policies for builtin gateway
usage.

##### `to` instead of `from`

With `MeshFaultInjection` and `MeshRateLimit`, `spec.to` with `kind:
MeshGateway` is now required instead of `spec.from` and `kind: MeshService`.

### `MeshGateway`

A new maximum length of 253 characters for listener hostnames has been introduced in order to ensure they are valid DNS names.

### Unifying Default Connection Timeout Values

To simplify configuration and provide a more consistent user experience, we've unified the default connection timeout values. When no `MeshTimeout` or `Timeout` policy is specified, the connection timeout will now be the same as the default `connectTimeout` values for `MeshTimeout` and `Timeout` policies. This value is now `5s`, which is a decrease from the previous default of `10s`.

The connection timeout specifies the amount of time Envoy will wait for an upstream TCP connection to be established.

The only users who need to take action are those who are explicitly relying on the previous default connection timeout value of `10s`. These users will need to create a new `MeshTimeout` policy with the appropriate `connectTimeout` value to maintain their desired behavior.

We encourage all users to review their configuration, but we do not anticipate that this change will require any action for most users.

### Default `TrafficRoute` and `TrafficPermission` resources are not created when creating a new `Mesh`

We decided to remove default `TrafficRoute` and `TrafficPermission` policies that were created during a new mesh creation. Since this release your applications can communicate without need to apply any policy by default.
If you want to keep the previous behaviour set `KUMA_DEFAULTS_CREATE_MESH_ROUTING_RESOURCES` to `true`.

**The following policies will no longer be created automatically**:
  
  * `CircuitBreaker`
  * `Retry`
  * `Timeout`
  * `TrafficPermission`
  * `TrafficRoute`

**The following policies will be created by default**:

  * `MeshCircuitBreaker`
  * `MeshRetry`
  * `MeshTimeout`

> [!CAUTION]
> Before enabling `mTLS`, remember to add `MeshTrafficPermission.`

Previously, Kuma would automatically create the default `TrafficPermission` policy for traffic routing. However, starting from version `2.6.0`, this is no longer the case.

If you are using `mTLS`, you will need to manually create the `MeshTrafficPermission` policy before enabling `mTLS`.

The `MeshTrafficPermission` policy allows you to specify which services can communicate with each other. This is necessary in a `mTLS` environment because `mTLS` requires that all communication between services be authenticated and authorized.

#### When is it appropriate to set the `KUMA_DEFAULTS_CREATE_MESH_ROUTING_RESOURCES` environment variable to `true`?

* When zones connecting to the global control plane may be running an older version than `2.6.0`.
* When recreating an environment using continuous delivery (CD) with legacy policies, missing the `TrafficRoute` policy will prevent legacy policies from being applied.

### Change of underlying envoy RBAC plugin for MeshTrafficPermission policies targeting HTTP services

With the release of Kuma 2.6.0, we've made some changes to the implementation of `MeshTrafficPermission` policies targeting HTTP services. These changes primarily revolve around the use of the `envoy.filters.http.rbac` envoy filter instead of the `envoy.filters.network.rbac` filter. This migration entails the following adjustments:

1. **Denied Request Response**: Rejected requests will now receive a 403 response code with the message `RBAC: access denied` instead of the previous 503 code. This aligns with the typical HTTP response code for authorization failures.

2. **RBAC-Related Envoy Stats**: The prefix for RBAC-related Envoy stats has been updated from `<inbound|outbound>:<stat_prefix>.rbac.` to `http.<stat_prefix>.rbac.`. This reflects the use of the HTTP filter for RBAC enforcement. For instance, the stat `inbound:127.0.0.1:21011.rbac.allowed` will now become `http.127.0.0.1:21011.rbac.allowed.` If you're utilizing these stats in your observability stack, you'll need to update your configuration to reflect the change.

To ensure a smooth transition to Kuma 2.6.0, carefully review your existing configuration files and make necessary adjustments related to denied request responses and RBAC-related Envoy stats.

### Make SI format valid for bandwidth in MeshFaultInjection policy

Prior to this upgrade `mbps` and `gbps` were used for units for parameter `conf.responseBandwidth.percentage`.
These are not valid units according to the [International System of Units](https://en.wikipedia.org/wiki/International_System_of_Units) they are respectively corrected to `Gbps` and `Mbps` if using
these invalid units convert them into `kbps` prior to upgrade to avoid invalid format.

### Deprecation of postgres driverName=postgres (lib/pq)

The postgres driver `postgres` (lib/pq) is deprecated and will be removed in the future.
Please migrate to the new postgres driver `pgx` by setting `DriverName=pgx` configuration option or `KUMA_STORE_POSTGRES_DRIVER_NAME=pgx` env variable.

## Upgrade to `2.5.x`

### Transparent-proxy and CNI v1 removal

v2 has been default since 2.2.x. We are therefore removing v1.

### Deprecated argument to transparent-proxy

Parameters `--exclude-outbound-tcp-ports-for-uids` and `--exclude-outbound-udp-ports-for-uids` are now merged into `--exclude-outbound-ports-for-uids` for `kumactl install transparent-proxy`.
We've also added the matching Kubernetes annotation: `traffic.kuma.io/exclude-outbound-ports-for-uids`.
The previous versions will still work but will be removed in the future.

### More strict validation rules for resource names

In order to be compatible with Kubernetes naming policy we updated the validation rules. Old rule:

> Valid characters are numbers, lowercase latin letters and '-', '_' symbols.

New rule:

> A lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character

New rule is applied for CREATE operations. The old rule is still applied for UPDATE, but this is going to change in Kuma 2.7.x or later.

### API

#### overview API coherency

These endpoints are getting replaced to achieve more coherency on the API:

- `/meshes/{mesh}/zoneegressoverviews` moves to `/meshes/{mesh}/zoneegresses/_overview`
- `/meshes/{mesh}/zoneingresses+insights` moves to `/meshes/{mesh}/zone-ingresses/_overview`
- `/meshes/{mesh}/dataplanes+insights` moves to `/meshes/{mesh}/dataplanes/_overview`
- `/zones+insights` moves to `/zones/_overview`

While you can use the old API they will be removed in a future version

### Prometheus inbound listener is not secured by TrafficPermission anymore

Due to the shadowing [issue](https://github.com/kumahq/kuma/issues/2417) with old TrafficPermission it was quite impossible to protect Prometheus inbound listener as expected.
RBAC rules on the Prometheus inbound listener were blocking users from fully migrate to the new MeshTrafficPermission policy. 
That's why we decided to discontinue TrafficPermission support on the Prometheus inbound listener starting 2.5.x.

### Gateway API

We support `v1` resources and `v1.0.0` of `gateway-api`. `v1beta1` resources are
still supported but support for these WILL be removed in a future release.

### KDS Delta enabled by default

KDS Delta is enabled by default. You can fallback to SOTW KDS by setting `KUMA_EXPERIMENTAL_KDS_DELTA_ENABLED=false`.
As a side effect, on kubernetes policies synced will be persisted in the `kuma-system` namespace instead of `default`.

## Upgrade to `2.4.x`

### Configuration change

The configuration: `Metrics.Mesh.MinResyncTimeout` and `Metrics.Mesh.MaxResyncTimeout` are replaced by `Metrics.Mesh.MinResyncInterval` and `Metrics.Mesh.FullResyncInterval`.
You can still use the current configs but it will be removed in the future.

### **Breaking changes**

#### Removal of service field in Dataplane outbound

After a period of depreciation, the service field in now removed. The service name is only defined by the value of  `kuma.io/service` in the outbound tags field.

## Upgrade to `2.3.x`

### **Breaking changes**

#### `MeshHTTPRoute`

* Changed path match `type` from `Prefix` to `PathPrefix`

#### `MeshAccessLog`

* Added a new field `Type` for `Backend` as a [Discriminator Field](https://github.com/kubernetes/enhancements/blob/master/keps/sig-api-machinery/1027-api-unions/README.md#discriminator-field)
* Added a new field `Type` for `Format` as a [Discriminator Field](https://github.com/kubernetes/enhancements/blob/master/keps/sig-api-machinery/1027-api-unions/README.md#discriminator-field)

#### `MeshTrace`

* Added a new field `Type` for `Backend` as a [Discriminator Field](https://github.com/kubernetes/enhancements/blob/master/keps/sig-api-machinery/1027-api-unions/README.md#discriminator-field)

#### `kumactl` container image

* Changed image's entrypoint to `/usr/bin/kumactl`

This change was introduced to be consistent with `kuma-cp` and `kuma-dp` images,
where names of images refer to binaries set in entrypoint. 

Example valid before:
```sh
docker run kumahq/kumactl:2.2.1 kumactl install transparent-proxy --help
```

Equivalent example valid now:
```sh
docker run kumahq/kumactl:2.3.0 install transparent-proxy --help
```

#### TLS verification between Zone CP and Global CP

If the CA used to sign the Global CP sync server is not provided to a Zone CP (HELM `controlPlane.tls.kdsZoneClient`, ENV: `KUMA_MULTIZONE_ZONE_KDS_ROOT_CA_FILE`), and the certificate is signed by a CA that is not included in the system's CA bundle on the Zone CP machine, you must do one of the following:
* Provide the CA to the Zone CP, see https://kuma.io/docs/2.2.x/production/secure-deployment/certificates/#control-plane-to-control-plane-multizone .
* Configure Zone CP. Set `KUMA_MULTIZONE_ZONE_KDS_TLS_SKIP_VERIFY` or HELM value of `controlPlane.tls.kdsZoneClient.skipVerify` to `true`. 

#### Removal of Common Name from generated certificates

This only affects users who rely on generated certificates having a common name set.

* `kumactl generate tls-certificate` generates certificates without CN
* autogenerated TLS certificate for kuma-cp (when `general.tlsCertFile` is not provided) won't have CN

## Upgrade to `2.2.x`

### Universal

#### CentOS 7

We are dropping support for running Envoy on CentOS 7 with this release and will
not release CentOS 7 compatible Envoy builds.

#### Changed default postgres driver to pgx

- If you encounter any problems with the persistence layer please [submit an issue](https://github.com/kumahq/kuma/issues/new) and temporarily switch to the previous driver (`lib/pq`) by setting
`DriverName=postgres` configuration option or `KUMA_STORE_POSTGRES_DRIVER_NAME='postgres'` env variable.
- Several configuration settings are not supported by the new driver right now, if used to configure them please try running with new defaults or [submit an issue](https://github.com/kumahq/kuma/issues/new).
List of unsupported configuration options:
  - MaxIdleConnections (used in store)
  - MinReconnectInterval (used in events listener)
  - MaxReconnectInterval (used in events listener)

#### Longer name of the resource in postgres

Kuma now permits the creation of a resource with a name of up to 253 characters, which is an increase from the previous limit of 100 characters. This adjustment brings our system in line with the naming convention supported by Kubernetes.
This change requires to run `kuma-cp migrate up` to apply changes to the postgres database.

### K8s

#### Removed deprecated annotations

- `kuma.io/builtindns` and `kuma.io/builtindnsport` are removed in favour of `kuma.io/builtin-dns` and `kuma.io/builtin-dns-port` introduced in 1.8.0. If you are using the legacy CNI you main need to set these old annotations manually in your pod definition.
- `kuma.io/sidecar-injection` is no longer supported as an annotation, you should use it as a label.

#### Helm

All containers now have defaults for `resources.requests.{cpu,memory}` and `resources.limits.{memory}`.
There are new default values for `*.podSecurityContext` and `*.containerSecurityContext`, see `values.yaml`.

#### Gateway API

We now support version `v0.6.0` of the Gateway API. See the [upstream API
changes](https://github.com/kubernetes-sigs/gateway-api/releases/tag/v0.6.0) for
more info.

### Auth configuration of DP server in Kuma CP

`dpServer.auth` configuration of Kuma CP was deprecated. You can still set config in this section, but it will be removed in the future.
It's recommended to migrate to `dpServer.authn` if you explicitly set any of the configuration in this config section.
* `dpServer.auth.type` is now split into two: `dpServer.authn.dpProxy.type` and `dpServer.authn.zoneProxy.type` and is still autoconfigured based on the environment.
* `dpServer.auth.useTokenPath` is now `dpServer.authn.enableReloadableTokens`

### Transparent Proxy Engine v2 and CNI v2 as default

As they matured, in the upcoming release Kuma will by default use transparent
proxy engine v2 and CNI v2.

If you want to still use v1 versions of these components, you will have to install 
Kuma with provided `legacy.transparentProxy=true` or `legacy.cni.enabled=true`
options.

#### Examples

##### CNI

*Helm*

```sh
helm upgrade --install --create-namespace --namespace kuma-system \
  --set "legacy.cni.enabled=true" \
  --set "cni.enabled=true" \
  --set "cni.chained=true" \
  --set "cni.netDir=/etc/cni/net.d" \
  --set "cni.binDir=/opt/cni/bin" \
  --set "cni.confName=10-calico.conflist"
  kuma kuma/kuma
```

*kumactl*

```sh
kumactl install control-plane \
  --set "legacy.cni.enabled=true" \
  --set "cni.enabled=true" \
  --set "cni.chained=true" \
  --set "cni.netDir=/etc/cni/net.d" \
  --set "cni.binDir=/opt/cni/bin" \
  --set "cni.confName=10-calico.conflist" \
  | kubectl apply -f-
```

##### Transparent Proxy Engine

*Helm*

```sh
helm upgrade --install --create-namespace --namespace kuma-system \
  --set "legacy.transparentProxy=true" kuma kuma/kuma
```

*kumactl*

```sh
kumactl install control-plane --set "legacy.transparentProxy=true" | kubectl apply -f-
```

### Removal of deprecated options to reach applications bound to `localhost`

The deprecated options `KUMA_DEFAULTS_ENABLE_LOCALHOST_INBOUND_CLUSTERS` and
`defaults.enableLocalhostInboundClusters` were removed.

This change affects only applications using transparent proxy.

Applications that are binding to `localhost` won't be reachable anymore.
This is the default behaviour from Kuma 1.8.0. Until now, it was possible to set
a deprecated kuma-cp configurations `KUMA_DEFAULTS_ENABLE_LOCALHOST_INBOUND_CLUSTERS`
or `defaults.enableLocalhostInboundClusters` to `true`, which was allowing to
still reach these applications.

One of the options to upgrade change address which the application is
listening on, to `0.0.0.0`.
Other option is to define `dataplane.networking.inbound[].serviceAddress`
to the address which service is binding to.

## Upgrade to `2.1.x`

### **Breaking changes**

#### **Naming Serviceless dataplanes has changed**

Currently, the `kuma.io/service` value of the inbound of a `Dataplane` generated for a `Pod` without a `Service` is based on the `Pod` name. The Kuma CP takes the pod's name and removes 2 last elements after splitting by `-`. This behavior is correct when the `Pod` is owned by a `Deployment` or `CronJob` but not for other owner kinds. Kuma will now use the name of the owner resource as the `kuma.io/service` value.
Before upgrade:
1. Identify all `Service`less `Pods` that are not managed by a `Deployment` or `CronJob`.
2. Create copies of policies that were created for the services corresponding to these `Pods`. The `kuma.io/service` value is the name of the owner resource. If there is no owner, `Kuma` uses the `Pod`'s name.

This breaking change is required to provide correct naming. The previous behavior could produce the same `kuma.io/service` value of the inbound of a `Dataplane` for many different serviceless Dataplanes.

#### MeshTrafficPermission

Action value have switched to PascalCase. ALLOW is Allow, DENY is Deny and ALLOW_WITH_SHADOW_DENY is AllowWithShadowDeny.

### HTTP api

We've removed the deprecated endpoint `POST /tokens`, use the `POST /tokens/dataplane` endpoint instead (same request and response).
Make sure you are using a recent `kumactl` or that you use the right path if using the API directly to upgrade with no issues.

### Kubernetes

The sidecar container is always injected first (since [#5436](https://github.com/kumahq/kuma/pull/5436)). This should only impact you when modifying the sidecar container with a container-patch. If you do so, upgrade Kuma and then change your container patch to modify the right container.

This version changes the leader election mechanism from leader for life to the more robust leader with lease.
As the result, during the upgrade you may have two leaders in the cluster.
This should not impact the system in any significant way other than logs like `resource was already updated`.

### Kumactl

`--valid-for` must be set for all token types, before it was defaulting to 10 years.

## Upgrade to `2.0.x`

### Built-in gateway

If you're using the `PREFIX` path match for `MeshGatewayRoute`,
note that validation is now stricter.
If you try to update an existing `MeshGatewayRoute` or create a new one,
make sure your `PREFIX` matching `value` does not include a trailing slash.
All prefix matches are checked path-separated,
meaning that `/prefix` only matches
if the request's path is `/prefix` or begins with `/prefix/`.
This has always been the case,
so no behavior has been changed
and existing resources with a trailing slash are not affected.

### Universal

A `lib/pq` change enables SNI by default when connecting to Postgres over TLS.
Either make sure your certificates contain a valid CN or SANs for the hostname
you're using
or update to `2.0.1` and disable `sslsni` by setting the
`KUMA_STORE_POSTGRES_TLS_DISABLE_SSLSNI` environment variable or
`store.postgres.tls.disableSSLSNI` in the config to `true`.

### `kuma-prometheus-sd`

This component has been removed
after [a long period of deprecation](https://github.com/kumahq/kuma/issues/2851).

### Zone Ingress Token migration

This is only relevant to Multizone deployment with Universal zones.
Zone Token that was previously used for authenticating Zone Egress, can now be used to authenticate Zone Ingress.
Please regenerate Zone Ingress token using `kumactl generate zone-token --scope=ingress`.
For the time being you can still use the old Zone Ingress token and Zone Token with scope ingress.
However, Zone Ingress Token is now deprecated and will be removed in the future.

### Helm

`ingress.annotations` and `egress.annotations` are deprecated in favour of `ingress.podAnnotations` and `egress.podAnnotations` which is a better name and aligne with the existing `controlPlane.podAnnoations`.


### Kuma-cp

- By default, the minimum TLS version allowed on servers is TLSv1.2. If you require using TLS < 1.2 you can set `KUMA_GENERAL_TLS_MIN_VERSION`.
- `KUMA_MONITORING_ASSIGNMENT_SERVER_GRPC_PORT` was removed after a long deprecation period use `KUMA_MONITORING_ASSIGNMENT_SERVER_PORT` instead.

### gRPC metrics

With this release, emitting separate statistics for every gRPC method is disabled.
gRPC metrics from different methods are now aggregated under `envoy_cluster_grpc_request_message_count`.
It will be re-enabled again in the future once Envoy with [`replace_dots_in_grpc_service_name`](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/grpc_stats/v3/config.proto#envoy-v3-api-field-extensions-filters-http-grpc-stats-v3-filterconfig-stats-for-all-methods) feature is released.
If you need to enable this setting, you can use ProxyTemplate to patch `envoy.filters.http.grpc_stats` http filter.

## Upgrade to `1.8.x`

### Kumactl

* `kumactl inspect dataplane --config-dump` was deprecated in favour of `kumactl inspect dataplane --type config-dump`. The behaviour of the new flag is unchanged but you should migrate.
* `kumactl install transparent-proxy --skip-resolv-conf` was deprecated as there's no reason for us to update the `/etc/resolv.conf` of the user.
* `kumactl install transparent-proxy --kuma-cp-ip` was removed as it's not possible to run a DNS server on the cp. 

### Helm

* Under `cni.image`, the default values for `repository` and `registry` have been
changed to agree with the other `image` values.

### CP

* The `/versions` endpoint was removed. This is not something that was reliable enough and version compatibility
is checked inside the DP
* We are deprecating `kuma.io/builtindns` and `kuma.io/builtindnsport` annotations in favour of the clearer `kuma.io/builtin-dns` and `kuma.io/builtin-dns-port`. The behavior of the new annotations is unchanged but you should migrate (a warning is present on the log if you are using the deprecated version).
* By default, applications binding to `localhost` are not reachable anymore. A `Dataplane` inbound's default `serviceAddress` is now the inbound's `address`. Before upgrade, if you have applications listening on `localhost` that you want to expose on:
  * Kubernetes: listen on `0.0.0.0` instead
  * Universal: listen on `inbound.address` instead or set `dataplane.networking.inbound[].serviceAddress: "127.0.0.1"`
To make migration easier you can temporarily disable this new behavior by setting `KUMA_DEFAULTS_ENABLE_LOCALHOST_INBOUND_CLUSTERS=true` on `kuma-cp`, this option will be removed in a future version.

## Upgrade to `1.7.x`

### Kumactl

* We're deprecating `kumactl install metrics/tracing/logging`, please use `kumactl install observability` instead

### DNS

The `control-plane` no longer hosts a builtin DNS server. You should always rely on the embedded DNS in the dataplane proxy and VIPs can't be used without transparent proxy.

### Timeout policy

'grpc' section is deprecated.
Timeouts for HTTP, HTTP2 and GRPC should be set in 'http' section:

```yaml
tcp: 
  idleTimeout: 1h 
http: # http, http2, grpc
  requestTimeout: 15s 
  idleTimeout: 1h
  streamIdleTimeout: 30m
  maxStreamDuration: 0s
grpc: # DEPRECATED
  streamIdleTimeout: 30m # DEPRECATED, use 'http.streamIdleTimeout'
  maxStreamDuration: 0s # DEPRECATED, use 'http.maxStreamDuration'
```

## Upgrade to `1.6.x`

### Helm

* the Helm chart for this release requires at least Helm version `3.8.0`.
* `controlPlane.resources` is now on object instead of a string. Any existing value should be adapted accordingly.

### Zone egress and ExternalService

When an `ExternalService` has the tag `kuma.io/zone` and `ZoneEgress` is enabled then the request flow will be different after upgrading Kuma to the newest version.
Previously, the request to the `ExternalService` goes through the `ZoneEgress` in the current zone. The newest version flow is different, and when `ExternalService` is defined in a different zone then the request will go through local `ZoneEgress` to `ZoneIngress` in zone where `ExternalService` is defined and leave the cluster through `ZoneEgress` in this cluster. To keep previous behavior, remove the `kuma.io/zone` tag from the `ExternalService` definition.

### Zone egress

Previously, when mTLS was configured and `ZoneEgress` deployed, requests were routed automatically through `ZoneEgress`. Now it's required to
explicitly set that traffic should be routed through `ZoneEgress` by setting `Mesh` configuration property `routing.zoneEgress: true`. The
default value of the property is set to `false` so in case your network policies don't allow you to reach other external services/zone without
using `ZoneEgress`, set `routing.zoneEgress: true`.

```yaml
type: Mesh
name: default
mtls: # mTLS is required for zoneEgress
 [...]
routing:
 zoneEgress: true
```

The new approach changes the flow of requests to external services. Previously when there was no instance of `ZoneEgress` traffic was routed
directly to the destination, now it won't reach the destination.

### Gateway (experimental)

Previously, a `MeshGatewayInstance` generated a `Deployment` and `Service` whose
names ended with a unique suffix. With this release, those objects will have the
same name as the `MeshGatewayInstance`.

### Inspect API

In connection with the changes around `MeshGateway` and `MeshGatewayRoute`, the output
schema of the `<policy-type>/<policy>/dataplanes` has changed. Every policy can
now affect both normal `Dataplane`s and `Dataplane`s configured as builtin gateways.
The configuration for the latter type is done via `MeshGateway` resources.

Every item in the `items` array now has a `kind` property of either:

* `SidecarDataplane`: a normal `Dataplane` with outbounds, inbounds,
  etc.
* `MeshGatewayDataplane`: a `MeshGateway`-configured `Dataplane` with a new
  structure representing the `MeshGateway` it serves.

Some examples can be found in the Inspect API docs.

## Upgrade to `1.5.x`

### Any type

The `kuma.metrics.dataplane.enabled` and `kuma.metrics.zone.enabled` configurations have been removed.

Kuma always generate the corresponding metrics.

### Kubernetes

- Please migrate your `kuma.io/sidecar-injection` annotations to labels.
  The new version still supports annotation, but to have a guarantee that applications can only start with sidecar, you must use label instead of annotation.
- Configuration parameter `kuma.runtime.kubernetes.injector.sidecarContainer.adminPort` and environment variable `KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_ADMIN_PORT`
  have been deprecated in favor of `kuma.bootstrapServer.params.adminPort` and `KUMA_BOOTSTRAP_SERVER_PARAMS_ADMIN_PORT`.

### Universal

- We removed support for old Ingress (`Dataplane#networking.ingress`) from pre 1.2 days.
  If you are still using it, please migrate to `ZoneIngress` first (see `Upgrade to 1.2.0` section).
- You can't use 0.0.0.0 or :: in `networking.address` most of the time using loopback is what people intended.
- Kuma DP flag `--admin-port` and environment variable `KUMA_DATAPLANE_ADMIN_PORT` have been deprecated, 
  admin port should be specified in Dataplane or ZoneIngress resources.

## Upgrade to `1.4.0`

Starting with this version, the default API server authentication method is user
tokens. In order to continue using client certificates (the previous default
method), you'll need to explicitly set the authentication method to client
certificates. This can be done by setting the `KUMA_API_SERVER_AUTHN_TYPE` variable to
`"clientCerts"`.

## Upgrade to `1.3.0`

Starting with this version `Mesh` resource will limit the maximal number of mtls backends to 1, so please make sure your `Mesh` has correct backend applied before the upgrade.

Outbound generated internally are no longer listed in `dataplane.network.outbound[]`. For Kubernetes, they will automatically disappear. For universal to remove them you should recreate your dataplane resources (either with `kumactl apply` or by restarting your services if the dataplanes lifecycle is managed by Kuma).

Kuma 1.3.0 has additional mechanism for tracking data plane proxies and zone statuses in a more reliable way. This mechanism works as a heartbeat and periodically increments the `generation` counter for the Insights. If the overall time for upgrading all Kuma CP instances is more than 5 minutes, then some data plane proxies or zones may become Offline in the GUI, but this doesn't affect real connectivity, only view. This unwanted effect will disappear as soon as all Kuma CP instances will be upgraded to 1.3.0.

## Upgrade to `1.2.1`

When Global is upgraded to `1.2.1` and Zone CP is still `1.2.0`, ZoneIngresses will always be listed as offline.
After Zone CPs are upgraded to `1.2.1`, the status will work again. ZoneIngress status does not affect cross-zone traffic.

## Upgrade to `1.2.0`

One of the changes introduced by Kuma 1.2.0 is renaming `Remote Control Planes` to `Zone Control Planes` and `Dataplane Ingress` to `Zone Ingress`. 
We think this change makes the naming more consistent with the rest of the application and also removes some of unnecessary confusion.

As a result of this renaming, some values and arguments in multizone/kubernetes environment changed. You can read below more.

### Upgrading with `kumactl` on Kubernetes

1. Changes in arguments/flags for `kumactl install control-plane`

   * `--mode` accepts now values: `standalone`, `zone` and `global` (`remote` changed to `zone`)

   * `--tls-kds-remote-client-secret` flag was renamed to `--tls-kds-zone-client-secret`

2. Service `kuma-global-remote-sync` changed to `kuma-global-zone-sync` so after upgrading `global` control plane you have to manually remote old service. For example:

   ```sh
   kubectl delete -n kuma-system service/kuma-global-remote-sync 
   ```

    Hint: It's worth to remember that often at this point the IP address/hostname which is used as a KDS address when installing Kuma Zone Control Planes will change. Make sure that you update the address when upgrading the Remote CPs to the newest version.

### Upgrading with `helm` on Kubernetes

Changes in values in Kuma's HELM chart

* `controlPlane.mode` accepts now values: `standalone`, `zone` and `global` (`remote` changed to `zone`)

* `controlPlane.globalRemoteSyncService` was renamed to `controlPlane.globalZoneSyncService`

* `controlPlane.tls.kdsRemoteClient` was renamed to `controlPlane.tls.kdsZoneClient`

### Suggested Upgrade Path on Universal

1. Zone Control Planes should be started using new environment variables

   * `KUMA_MODE` accepts now values: `standalone`, `zone` and `global` (`remote` changed to `zone`)

     Old:
     ```sh
     KUMA_MODE="remote" [...] kuma-cp run
     ```

     New:
     ```sh
     KUMA_MODE="zone" [...] kuma-cp run
     ```

   * `KUMA_MULTIZONE_REMOTE_ZONE` was renamed to `KUMA_MULTIZONE_ZONE_NAME`

     Old:
     ```sh
     KUMA_MULTIZONE_REMOTE_ZONE="remote-1" [...] kuma-cp run
     ```

     New:
     ```sh
     KUMA_MULTIZONE_ZONE_NAME="remote-1" [...] kuma-cp run
     ```

   * `KUMA_MULTIZONE_REMOTE_GLOBAL_ADDRESS` was renamed to `KUMA_MULTIZONE_ZONE_GLOBAL_ADDRESS`

     Old:
     ```sh
     KUMA_MULTIZONE_REMOTE_GLOBAL_ADDRESS="grpcs://localhost:5685" [...] kuma-cp run
     ```

     New:
     ```sh
     KUMA_MULTIZONE_ZONE_GLOBAL_ADDRESS="grpcs://localhost:5685" [...]  kuma-cp run
     ```

   * `KUMA_MULTIZONE_REMOTE_KDS_ROOT_CA_FILE` was renamed to `KUMA_MULTIZONE_ZONE_KDS_ROOT_CA_FILE`

     Old:
     ```sh
     KUMA_MULTIZONE_REMOTE_KDS_ROOT_CA_FILE="/rootCa" [...] kuma-cp run
     ```

     New:
     ```sh
     KUMA_MULTIZONE_ZONE_KDS_ROOT_CA_FILE="/rootCa" [...] kuma-cp run
     ```

   * `KUMA_MULTIZONE_REMOTE_KDS_ROOT_CA_FILE` was renamed to `KUMA_MULTIZONE_ZONE_KDS_ROOT_CA_FILE`

     Old:
     ```sh
     KUMA_MULTIZONE_REMOTE_KDS_REFRESH_INTERVAL="9s" [...] kuma-cp run
     ```

     New:
     ```sh
     KUMA_MULTIZONE_ZONE_KDS_REFRESH_INTERVAL="9s" [...] kuma-cp run
     ```

2. Dataplane Ingress resource should be replaced with ZoneIngress resource:

    Old:
    ```yaml
    type: Dataplane
    name: dp-ingress
    mesh: default
    networking:
      address: <ADDRESS>
      ingress:
        publicAddress: <PUBLIC_ADDRESS>
        publicPort: <PUBLIC_PORT>
      inbound:
      - port: <PORT>
        tags:
          kuma.io/service: ingress
    ```

    New:
    ```yaml
    type: ZoneIngress
    name: zone-ingress
    networking:
      address: <ADDRESS>
      port: <PORT>
      advertisedAddress: <PUBLIC_ADDRESS>
      advertisedPort: <PUBLIC_PORT>
    ```

    NOTE: ZoneIngress resource is a global scoped resource, it's not bound to a Mesh
    The old Dataplane resource is still supported but it's considered deprecated and will be removed in the next major version of Kuma


3. Since ZoneIngress resource is not bound to a Mesh, it requires another token type that is bound to a Zone:
   
    ```shell
    kumactl generate zone-ingress-token --zone=zone-1 > /tmp/zone-ingress-token
    ```

4. `kuma-dp run` command should be updated with a new flag `--proxy-type=ingress`:

    ```sh
    kuma-dp run \
      --proxy-type=ingress \
      --dataplane-token-file=/tmp/zone-ingress-token \
      --dataplane-file=zone-ingress.yaml
    ```


## Upgrade to `1.1.0`

The major change in this release is the migration to XDSv3 for the `kuma-cp` to `envoy` data plane proxy communication. The
previous XDSv2 is still available and will continue working. All the existing data plane proxies will still use XDSv2 until
being restarted. The newly deployed `kuma-dp` instances will automatically get bootstrapped to XDSv3. In case that needs to be
changed, `kuma-cp` needs to be started with `KUMA_BOOTSTRAP_SERVER_API_VERSION=v2`.

With Kuma 1.1.0, the `kuma-cp` will installs default retry and timeout policies for each new
created Mesh object. The pre-existing meshes will not automatically get these default policies. If needed, they should be created accordingly.

This version removes the deprecated `--dataplane` flag in `kumactl generate dataplane-token`, please consider migrating to use `--name` instead.

## Upgrade to `1.0.0`

This release introduces a number of breaking changes. If Kuma is being deployed in production we strongly suggest to backup the current configuration, tear down the whole cluster and zones, and install in a clean setup. However, we enumerate the details of these changes below.

### Suggested Upgrade Path on Kubernetes
 * Drop k8s 1.13 support

    Take this into account if you run Kuma on an old Kubernetes version.

 * `kumactl` merged `install ingress` into `install control-plane`

    This change impacts any deployment pipelines that are based on `kumactl` and are used for multi-zone deployments.

 * Change policies on K8S to scope global

    All the CRDs are now in the global scope, therefore all policies need to be backed up. The relevant CRDs need to be deleted, which will clear all the policies. After the upgrade, you can apply the policies again. We do recommend to keep all the Kuma Control Planes down while doing these operations.

 * Autoconfigure single cert for all services

    Deployment flags for providing TLS certificates in Helm and `kumactl` have changed, refer to the relevant [documentation](https://github.com/kumahq/kuma/blob/release-1.0/deployments/charts/kuma/README.md#values) to verify the new naming.

 * Create default resources for Mesh

    The following default resources will be created upon the first start of Kuma Control Plane
        - default signing key
        - default Allow All traffic permission policy `allow-all-<mesh name>`
        - Default Allow All traffic route policy `allow-all-<mesh name>`
    
    Please verify if this conflicts with your deployment and expected policies.

 * New Multizone deployment flow

    Deploying Multizone clusters is now simplified, please refer to the deployment documentation of the updated procedure.
   
 * Improved control plane communication security
   
    Kuma Control Plane exposed ports are reduced, please revise the documentation for detailed list.
    Consider reinstalling the metrics due to the port changes in Kuma Prometheus SD.
 
 * Traffic route format
 
    The format of the TrafficRoute has changed. Please check the documentation and adapt your resources. 

### Suggested Upgrade Path on Universal
 * Get rid of advertised hostname
    `KUMA_GENERAL_ADVERTISED_HOSTNAME` was removed and not needed now.
 
 * Autoconfigure single cert for all services
    Deployment flags for providing TLS certificates in Helm and `kumactl` have changed, refer to the documentation](https://github.com/kumahq/kuma/blob/release-1.0/pkg/config/app/kuma-cp/kuma-cp.defaults.yaml) to verify the new naming.

 * Create default resources for Mesh
    
    The following default resources will be created upon the first start of Kuma Control Plane
        - default signing key
        - default Allow All traffic permission policy `allow-all-<mesh name>`
        - Default Allow All traffic route policy `allow-all-<mesh name>`
    
    Please verify if this conflicts with your deployment and expected policies.

* New Multizone deployment flow

    Deploying Multizone clusters is now simplified, please refer to the deployment documentation of the updated procedure.
   
 * Improved control plane communication security
   
    `kuma-dp` invocation has changed and now allows for a more flexible usage leveraging automated, template based Dataplane resource creation, customizable data-plane token boundaries and additional CA ceritficate validation for the Kuma Control plane boostrap server.
    Kuma Control Plane exposed ports are reduced, please revise the documentation for detailed list.
 
  * Traffic route format
  
     The format of the TrafficRoute has changed. Please check the documentation and adapt your resources. 

 
## Upgrade to `0.7.0`
Support for `kuma.io/sidecar-injection` annotation. On Kubernetes change the namespace resources that host Kuma mesh services with the aforementioned annotation and delete the label. 

Prefix the Kuma built-in tags with `kuma.io/` as follows: `kuma.io/service`, `kuma.io/protocol`, `kuma.io/zone`.

### Suggested Upgrade Path on Kubernetes

Update the applied policy tag selector to include the `kuma.io/` prefix. A sample traffic resource follows:

```yaml
apiVersion: kuma.io/v1alpha1
kind: TrafficPermission
mesh: default
metadata:
  namespace: default
  name: allow-all-traffic
spec:
  sources:
    - match:
        kuma.io/service: '*'
  destinations:
    - match:
        kuma.io/service: '*'
```

The Kuma Control Plane will update the relevant Dataplane resources accordingly

### Suggested Upgrade Path on Universal

Update the applied policy tag selector to include the `kuma.io/` prefix. A sample traffic resource follows:

```yaml
type: TrafficPermission
name: allow-all-traffic
mesh: default
sources:
  - match:
      kuma.io/service: '*'
destinations:
  - match:
      kuma.io/service: '*'
```

Update the dataplane resources with the new tag format as well. Example:

```bash
echo "type: Dataplane
mesh: default
name: redis-1
networking:
  address: 192.168.0.1
  inbound:
  - port: 9000
    servicePort: 6379
    tags:
      kuma.io/service: redis" | kumactl apply -f -
```

This release changes the way that Distributed and Hybrid Kuma Control planes are deployed. Please refer to the documentation for more details.

## Upgrade to `0.6.0`

Passive Health Check were removed in favor of Circuit Breaking.

Format of Active Health Check changed from :
```yaml
apiVersion: kuma.io/v1alpha1
kind: HealthCheck
mesh: default
metadata:
  namespace: default
  name: web-to-backend-check
mesh: default
spec:
  sources:
  - match:
      service: web
  destinations:
  - match:
      service: backend
  conf:
    activeChecks:
      interval: 10s
      timeout: 2s
      unhealthyThreshold: 3
      healthyThreshold: 1
    passiveChecks:
      unhealthyThreshold: 3
      penaltyInterval: 5s
```
to 
```yaml
apiVersion: kuma.io/v1alpha1
kind: HealthCheck
mesh: default
metadata:
  namespace: default
  name: web-to-backend-check
mesh: default
spec:
  sources:
  - match:
      service: web
  destinations:
  - match:
      service: backend
  conf:
    interval: 10s
    timeout: 2s
    unhealthyThreshold: 3
    healthyThreshold: 1
```

### Suggested Upgrade Path on Kubernetes

In the new Kuma version serivce tag format has been changed. Instead of `backend.kuma-demo.svc:5678` service tag will look like this `backend_kuma-demo_svc_5678`. This is a breaking change and Policies should be updated to be compatible with the new Kuma version.

Please re-install Prometheus via `kubectl install metrics` and make sure that `skipMTLS` is set to `false` or omitted.
```yaml
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: default
spec:
  metrics:
    enabledBackend: prometheus-1
    backends:
    - name: prometheus-1
      type: prometheus
      conf:
        skipMTLS: false
```

### Suggested Upgrade Path on Universal

Make sure that `skipMTLS` is set to `true`.

```yaml
type: Mesh
name: default
metrics:
  enabledBackend: prometheus-1
  backends:
  - name: prometheus-1
    type: prometheus
    conf:
      skipMTLS: true
```


## Upgrade to `0.5.0`
### Suggested Upgrade Path on Kubernetes

#### Mesh resource format changes

The Mesh resource format in Kubernetes changed from
```yaml
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: default
spec:
  mtls:
    enabled: true
    ca:
      builtin: {}
  metrics:
    prometheus: {}
  logging:
    backends:
    - name: file-1
      file:
        path: /var/log/access.log
  tracing:
    backends:
    - name: zipkin-1
      zipkin:
        url: http://zipkin.local:9411/api/v1/spans
```
to
```yaml
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: default
spec:
  mtls:
    enabledBackend: ca-1
    backends:
    - name: ca-1
      type: builtin
  metrics:
    enabledBackend: prom-1
    backends:
    - name: prom-1
      type: prometheus
  logging:
    backends:
    - name: file-1
      type: file
      conf:
        path: /var/log/access.log
  tracing:
    backends:
    - name: zipkin-1
      type: zipkin
      conf:
        url: http://zipkin.local:9411/api/v1/spans
```

#### Removing `kuma-injector`

Kuma 0.5.0 ships with `kuma-injector` embedded into the `kuma-cp`, which makes its previously created resources obsolete and potentially
 can cause problems with the deployments. Before deploying the new version, it is strongly advised to run a cleanup script kuma-0.5.0-k8s-remove_injector_resources.sh.
 
 NOTE: if Kuma was deployed in a namespace other than `kuma-system`, please run `export KUMA_SYSTEM=<othernamespace` before running the cleanup script.

#### Kuma resources `ownerReferences` 
Kuma 0.5.0 introduce webhook for setting `ownerReferences` to the Kuma resources. If you have some 
Kuma resources in your k8s cluster, then you can use our script kuma-0.5.0-k8s-set_owner_references.sh 
in order to properly set `ownerReferences` .

### Suggested Upgrade Path on Universal

#### Mesh resource format changes
Mesh format on Universal changed from
```yaml
type: Mesh
name: default
mtls:
  enabled: true
  ca:
    builtin: {}
metrics:
  prometheus: {}
logging:
  backends:
  - name: file-1
    file:
      path: /var/log/access.log
tracing:
  backends:
  - name: zipkin-1
    zipkin:
      url: http://zipkin.local:9411/api/v1/spans
```
to
```yaml
type: Mesh
name: default
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
metrics:
  enabledBackend: prom-1
  backends:
  - name: prom-1
    type: prometheus
logging:
  backends:
  - name: file-1
    type: file
    conf:
      path: /var/log/access.log
tracing:
  backends:
  - name: zipkin-1
    type: zipkin
    conf:
      url: http://zipkin.local:9411/api/v1/spans
```

## Upgrade to `0.4.0`

### Suggested Upgrade Path on Kubernetes

No additional steps are needed.

### Suggested Upgrade Path on Universal

#### Migrations

Kuma 0.4.0 introduces DB Migrations for Postgres therefore before running the new version of Kuma, run the kuma-cp migration command.
```
kuma-cp migrate up
```
Remember to provide config just like in `kuma-cp run` command.
All existing data will be preserved.

#### New Dataplane Entity format

Kuma 0.4.0 introduces new Dataplane entity format to improve readability as well as add support for scraping metrics of Gateway Dataplanes. 

Here is example of migration to the new format.

**Dataplane**

Old format
```yaml
type: Dataplane
mesh: default
name: web-01
networking:
  inbound:
  - interface: 192.168.0.1:21011:21012
    tags:
      service: web
  outbound:
  - interface: :3000
    service: backend
```

New format
```yaml
type: Dataplane
mesh: default
name: web-01
networking:
  address: 192.168.0.1
  inbound:
  - port: 21011
    servicePort: 21012
    tags:
      service: web
  outbound:
  - port: 3000
    service: backend
```

**Gateway Dataplane**

Old format
```yaml
type: Dataplane
mesh: default
name: kong-01
networking:
  gateway:
    tags:
      service: kong
```

New format
```yaml
type: Dataplane
mesh: default
name: kong-01
networking:
  address: 192.168.0.1
  gateway:
    tags:
      service: kong
```

Although the old format is still supported, it is recommended to migrate since the support for it will be dropped in the next major version of Kuma.

## Upgrade to `0.3.1`

### List of breaking changes

`kuma policies`:
* `Mesh` CRD on Kubernetes is now Cluster-scoped
* `TrafficLog` policy is applied differently now: instead of applying all `TrafficLog` policies that match to a given `outbound` interface of a `Dataplane`, only a single the most specific `TrafficLog` policy is applied

`kumactl`:
* a few options in `kumactl config control-planes add` command have been renamed:
  * `--dataplane-token-client-cert` has been renamed into `--admin-client-cert`
  * `--dataplane-token-client-key` has been renamed into `--admin-client-key`

### Suggested Upgrade Path on Kubernetes

* Users on Kubernetes will have to re-install `Kuma`:

  1. Export all `Kuma` resources
     ```shell
     kubectl get meshes,trafficpermissions,trafficroutes,trafficlogs,proxytemplates --all-namespaces -oyaml > backup.yaml
     ```
  2. Uninstall previous version of `Kuma Control Plane`
     ```shell
     # using previous version of `kumactl`

     kumactl install control-plane | kubectl delete -f -
     ```
  3. Install new version of `Kuma Control Plane`
     ```shell
     # using new version of `kumactl`

     kumactl install control-plane | kubectl apply -f -
     ```
  4. Re-apply `Kuma` resources back again
     ```shell
     kubectl apply -f backup.yaml
     ```

### Suggested Upgrade Path on Universal

* Those users who used `--dataplane-token-client-cert` and `--dataplane-token-client-key` command line options in the past will have to re-run

   ```
   kumactl config control-planes add
   ```

   this time with

    ```shell
    --admin-client-cert <CERT> --admin-client-cert <KEY> --overwrite
    ```
* all components of `Kuma Control Plane` - `kuma-cp`, `kuma-dp`, `envoy` - have to be re-deployed
