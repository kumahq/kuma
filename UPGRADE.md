This document guides you through the process of upgrading `Kuma`.

First, check if a section named `Upgrade to x.y.z` exists,
with `x.y.z` being the version you are planning to upgrade to.

If such a section does not exist, the upgrade you want to perform
does not have any particular instructions.

## Upgrade to `3.0.0`

Sections are grouped by area: [Security and identity](#security-and-identity), [Policies and API](#policies-and-api), [Data plane and traffic](#data-plane-and-traffic), [Multizone, zone proxies and external services](#multizone-zone-proxies-and-external-services), [Gateways](#gateways), [Observability](#observability), [Control plane configuration](#control-plane-configuration). Most sections end with **Action required**; do the steps marked "before upgrading" while the control plane is still on 2.14.

### Security and identity

#### `Mesh.mtls` is not supported

Mesh mTLS comes only from `MeshIdentity` and `MeshTrust`. The legacy mesh CA (`Mesh.spec.mtls`) is removed from the API, issues nothing, and no longer affects any data plane proxy:

- **No mTLS from `mtls`.** A proxy gets an mTLS transport socket, inbound and outbound, only when a `MeshIdentity` matches it. The identity certificate and the trust bundle come from `MeshIdentity` and `MeshTrust`, never from a mesh CA backend. A mesh whose only identity source is `mtls` now serves and accepts plaintext, and `MeshTLS` and `MeshTrafficPermission` no longer apply to its proxies.
- **No trust bridge.** While both systems coexisted, a mesh with `mtls` enabled and at least one `MeshTrust` got the mesh CA injected into the trust bundle under the mesh name, so a `MeshIdentity` proxy still trusted peers presenting a legacy certificate. That bridge is gone: the trust bundle contains only the CAs declared by `MeshTrust` resources.
- **`PERMISSIVE` backends no longer make inbounds permissive.** The inbound mode is resolved from `MeshTLS` alone. A mesh that set `mtls.backends[].mode: PERMISSIVE` and has no `MeshTLS` policy now gets `Strict` inbounds, and plaintext traffic to them is rejected. A mesh using `MeshIdentity` already resolved to `Strict` without a `MeshTLS` policy.
- **No CA provisioning or validation.** The `builtin` and `provided` CA managers are gone. Nothing generates a CA, reads a `provided` certificate or key, or validates backends at admission. Deleting a Secret is no longer blocked when an `mtls` backend references it. `MeshIdentity` and `MeshTrust` secrets were never covered by that check, and the `inter-cp-ca` and `envoy-admin-ca` global secrets are separate PKIs, untouched by this change.
- **Field removed from the API.** `mtls` is gone from `Mesh` together with `CertificateAuthorityBackend` and everything it carried (`enabledBackend`, `backends`, `skipValidation`, `dpCert`, `conf`, `mode`, `rootChain`), and so is the now empty `Mesh.routing` message. The field number is reserved. `mtls` is now an unknown field, and `Mesh` drops unknown fields silently, so an existing `Mesh` still loads and a manifest that still carries `mtls` still applies; the field disappears the next time the resource is written. The Kubernetes `Mesh` CRD is unaffected, its `spec` is `x-kubernetes-preserve-unknown-fields`.
- **Tag-encoded SNIs are gone.** `SNIFromTags`/`TagsFromSNI`-style SNIs go away with the legacy path: outbounds to a destination that is not a real resource are plaintext and carry no SNI. `MeshService.status.tls` no longer gates permissive meshes; it is still computed and still gates outbound TLS for `MeshIdentity` proxies.
- **Insights.** `DataplaneInsight.mTLS` and the `MeshInsight.mTLS` aggregation are populated only from `MeshIdentity` issuance. A proxy with no workload identity reports no certificate.
- **`kumactl`.** `kumactl get meshes` no longer prints the `mTLS` column. `kumactl export --profile federation` no longer rewrites `builtin` backends into `provided` ones.

**Orphaned CA secrets are left in place**

`builtin` CA material was persisted as `<mesh>.ca-builtin-cert-<backend>` and `<mesh>.ca-builtin-key-<backend>` Secrets owned by their `Mesh`. The upgrade does not delete them: they stay, unread, and are still synced from global to zones by KDS and exported by `kumactl export`. Leaving them is deliberate, since the CA private key is unrecoverable once removed. They are removed automatically when their `Mesh` is deleted.

**Action required**

1. Before upgrading, migrate every mesh that still relies on `mtls` to `MeshIdentity` and `MeshTrust`, and complete the migration. This is the point where a partially migrated mesh breaks: once upgraded, proxies still holding a legacy mesh certificate are no longer trusted by their `MeshIdentity` peers, and there is no rolling path back.
2. Before upgrading, author a `MeshTLS` policy with `mode: Permissive` for every mesh that relies on `mtls.backends[].mode: PERMISSIVE`:

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

3. Drop `mtls` from the `Mesh` manifests you keep under source control.
4. Optionally, delete the orphaned `<mesh>.ca-builtin-*` Secrets once no downstream tooling needs the CA material.

#### `MeshIdentity.spec.spiffeID` is immutable and its trust domain no longer follows the zone

A trust domain is an identity namespace, so moving one is a migration rather than an edit. `spec.spiffeID.trustDomain` and `spec.spiffeID.path` are now rejected on update, on Kubernetes by the admission webhook and on Universal by the API server. The control plane also renders the trust domain once, when it first initializes the identity, and records it in `status.trustDomain`. A template such as `{{ .Zone }}` no longer re-renders when the zone is renamed, which used to move issuance into a trust domain no `MeshTrust` published yet.

**Action required**

To move to a different trust domain or SPIFFE ID path, create a second `MeshIdentity` **under a different name** with the new value and delete the old one once every workload has been issued a certificate from it. Both identities publish their own `MeshTrust` and `MeshService.spec.identities` lists the SPIFFE IDs of every matching identity, so peers accept leaves from both for the whole transition. Use `spec.selector.dataplane.matchLabels` or the name to steer which identity a workload picks while both exist.

Deleting the identity and recreating it under the same name is accepted, and the control plane does converge on the new trust domain, but it is not a migration: the `MeshTrust` and the CA are keyed by the identity name, so for as long as convergence takes there is a single trust domain in flight and leaves issued under the old one no longer verify. Two identities never have that gap.

#### `MeshService.spec.identities` accepts SPIFFE IDs only

`MeshService.spec.identities[].type` no longer accepts `ServiceTag`. `MeshService.spec.identities` publishes SPIFFE IDs only.

**Action required**

Before upgrading, update any `MeshService` manifest that still declares a `ServiceTag` identity to use `SpiffeID` entries only. A persisted `ServiceTag` entry is rejected by the updated schema.

#### Outbound mTLS negotiates TLS 1.3

Outbound mesh mTLS connections now allow TLS 1.3. Previously Envoy's client default capped them at TLS 1.2, so mesh traffic negotiated TLS 1.2 even though inbound listeners accepted TLS 1.3. A `MeshTLS` or `MeshExternalService` `tlsVersion.max` that is unset or `TLSAuto` now resolves to TLS 1.3 on the client side, which also fixes `min: TLS13` without `max` failing every connection with `NO_SUPPORTED_VERSIONS_ENABLED`.

**Action required**

None for most meshes. TLS 1.3 cipher suites are not configurable, so `tlsCiphers` no longer restricts connections that negotiate TLS 1.3. To keep outbound traffic on TLS 1.2, set `tlsVersion.max: TLS12` in `MeshTLS` or `MeshExternalService`.

#### Legacy HS256 signing keys no longer accepted

The control plane no longer accepts JWT tokens signed with the symmetric HS256 algorithm, which pre-1.4.x versions of Kuma used to sign dataplane tokens.

**Action required**

If any dataplane tokens issued by a pre-1.4.x control plane are still in use, rotate them before upgrading. They are rejected as using an unsupported algorithm afterward.

#### Control plane TLS certificates are reloaded without a restart

Every control plane server that reads its certificate from disk (API server HTTPS, dp-server, global KDS, MADS, diagnostics) now picks up a rotation performed by an external tool without restarting `kuma-cp`. Nothing has to be configured for this.

**Rotating the trust anchor is still not transparent to proxies**

A rotation is transparent only when the new certificate is issued by the CA the proxies already trust. It is not transparent when the rotation replaces the trust anchor itself, which is what happens with the self-signed certificate the control plane generates on its first run: `kuma-dp` receives that certificate as Envoy's trusted CA when it bootstraps, so a proxy that reconnects after the rotation validates the new certificate against the old one and fails until it bootstraps again. Established xDS streams are unaffected, because the certificate is only verified during the handshake.

**Action required**

If you rotate control plane certificates, issue them from a CA and point proxies at that CA with `kuma-dp run --ca-cert-file=/path/ca.pem` (or `KUMA_CONTROL_PLANE_CA_CERT_FILE`), then rotate leaves under it. This is required for a certificate issued by cert-manager or Vault in any case: such a leaf has `CA:FALSE`, and the control plane rejects a bootstrap request for it with `NotCA` unless the proxy supplies the CA itself.

#### Zone tokens only authenticate zone control planes

A zone token now has one job: authenticating a zone control plane to the global control plane over KDS.

- **Not accepted from zone proxies.** Zone proxies authenticate with a dataplane token, and the DP server no longer accepts zone tokens.
- **No scopes.** `kumactl generate zone-token` no longer accepts `--scope ingress` or `--scope egress`, and `POST /tokens/zone` no longer requires a `scope`. The token carries only the zone name.
- **Validated on global only.** The global control plane stops sending `zone-token-revocations` over KDS and stops deriving `zone-token-signing-public-key-*` from `zone-token-signing-key-*` for the zones. The copies zones already have are deleted on the first KDS sync after the upgrade.
- **Configuration moved.** The issuance settings move from `dpServer.authn.zoneProxy` to the KDS auth configuration, and `dpServer.authn.zoneProxy` is removed.

| Removed | Use instead |
|---|---|
| `dpServer.authn.zoneProxy.zoneToken.enableIssuer` (`KUMA_DP_SERVER_AUTHN_ZONE_PROXY_ZONE_TOKEN_ENABLE_ISSUER`) | `multizone.global.kds.auth.zoneToken.enableIssuer` (`KUMA_MULTIZONE_GLOBAL_KDS_AUTH_ZONE_TOKEN_ENABLE_ISSUER`) |
| `dpServer.authn.zoneProxy.zoneToken.validator.useSecrets` (`KUMA_DP_SERVER_AUTHN_ZONE_PROXY_ZONE_TOKEN_VALIDATOR_USE_SECRETS`) | `multizone.global.kds.auth.zoneToken.validator.useSecrets` (`KUMA_MULTIZONE_GLOBAL_KDS_AUTH_ZONE_TOKEN_VALIDATOR_USE_SECRETS`) |
| `dpServer.authn.zoneProxy.zoneToken.validator.publicKeys` | `multizone.global.kds.auth.zoneToken.validator.publicKeys` |
| `dpServer.authn.zoneProxy.type` (`KUMA_DP_SERVER_AUTHN_ZONE_PROXY_TYPE`) | none, it was read by nothing |

**Action required**

- If you set `enableIssuer` to `false` to mint zone tokens offline, move it to `multizone.global.kds.auth.zoneToken.enableIssuer` on the global control plane. The removed setting is ignored, so the issuer is enabled again until you do. The other removed settings had no effect.
- Drop `--scope ingress`/`--scope egress` from scripts calling `kumactl generate zone-token`.
- Give zone proxies a dataplane token, see [Standalone `ZoneIngress` and `ZoneEgress` are not supported](#standalone-zoneingress-and-zoneegress-are-not-supported).

#### The control plane and hook containers drop all Linux capabilities

`controlPlane.containerSecurityContext` and `hooks.containerSecurityContext` now default to `allowPrivilegeEscalation: false` and `capabilities.drop: [ALL]`, alongside the `readOnlyRootFilesystem: true` they already carried.

Neither the control plane nor the `kubectl` and `kumactl` hook jobs need a capability. Every port the control plane binds is above 1024, and nothing opens a raw or packet socket. The injected sidecar and the zone proxy already shipped with both settings. The transparent proxy init container keeps the `NET_ADMIN` and `NET_RAW` it adds for `iptables`, as does the CNI container, which still runs as root.

**Action required**

Helm merges these values key by key, so overriding one key of `containerSecurityContext` no longer replaces the whole block: an override that sets only `readOnlyRootFilesystem` now also inherits the two new keys. If your control plane needs a capability, or you run a `SecurityContextConstraint` or admission policy that grants one, set it back explicitly:

```yaml
controlPlane:
  containerSecurityContext:
    allowPrivilegeEscalation: true
    capabilities:
      drop: []
```

#### The universal Helm path no longer grants loopback callers admin

A universal control plane installed from the `kuma` Helm chart no longer treats a loopback caller as an administrator, which the Kubernetes path already refused. Until now anyone able to reach the pod over `kubectl port-forward` or `kubectl exec` held full admin without a token.

**Action required**

If you administer such a control plane through `kubectl exec` or `kubectl port-forward` plus `kumactl`, that access stops working. Configure `kumactl` with a user token instead:

```sh
kumactl config control-planes add --name <name> --address <address> --auth-type=tokens --auth-conf token=<token>
```

Getting that first token needs a step the control plane's startup log does not mention. It prints a `curl` command for reading the bootstrapped admin token over loopback, and that command returns 403 once loopback is no longer admin.

On a control plane that already runs, read the token before you upgrade. It stays valid afterwards.

On a new control plane, issue tokens offline instead. Generate a key pair off-cluster and keep the private half there:

```sh
kumactl generate signing-key --format=pem > token-key.pem
kumactl generate public-key --signing-key-path token-key.pem > token-key.pub
```

Give the control plane the public half and nothing else:

```yaml
controlPlane:
  config: |
    apiServer:
      authn:
        tokens:
          bootstrapAdminToken: false
          enableIssuer: false
          validator:
            useSecrets: false
            publicKeys:
            - kid: "1"
              key: |
                -----BEGIN RSA PUBLIC KEY-----
                ...
                -----END RSA PUBLIC KEY-----
```

Then mint an admin token whenever you need one, without reaching the cluster at all:

```sh
kumactl generate user-token --name mesh-system:admin --group mesh-system:admin --valid-for 24h --signing-key-path token-key.pem --kid 1
```

`useSecrets: false` is stricter than the loopback path it replaces: the control plane accepts only tokens signed by a key you hold. Leave it at `true` while tokens the control plane issued are still in use.

To keep the old behaviour instead, set `controlPlane.envVars.KUMA_API_SERVER_AUTHN_LOCALHOST_IS_ADMIN` back to `"true"`. That grants admin to anything that can open a loopback connection to the pod.

#### Control plane RBAC changes on Kubernetes

The Helm-installed control plane RBAC and admission webhooks change as follows. Nothing is needed for Helm installs. If you manage the control plane RBAC or webhook configuration yourself (for example through GitOps), apply the same changes.

**Added**

- `get`, `list`, `watch` on `gateway.networking.k8s.io` `grpcroutes`, `httproutes` and `referencegrants` in the `-control-plane` `ClusterRole`, to translate `GRPCRoute`s and evaluate `ReferenceGrant`s for cross-namespace backends. Without the read permissions the control plane fails to start its route caches and blocks before serving. `patch` on `grpcroutes/status` and `httproutes/status` stays granted.
- `get`, `list`, `patch` on `gateway.networking.k8s.io` `gatewayclasses`, for the legacy `GatewayClass` finalizer cleanup during the upgrade (see [Built-in gateway removed](#built-in-gateway-removed)). You can remove them once no Kuma-managed `GatewayClass` is left, or remove the `gateway-exists-finalizer.gateway.networking.k8s.io` finalizer yourself before upgrading instead.

**Removed**

- All access to `grpcroutes` and `httproutes` from the `-control-plane-workloads` `ClusterRole` and the namespaced `-control-plane` `Role`. These were leftovers from the built-in gateway; the control plane reads routes through the `-control-plane` `ClusterRole` only.
- All access to `gateway.networking.k8s.io` `gateways`.
- `apps` `deployments` entirely; `apps` `replicasets` keeps only `get`, `list` and `watch`.
- `create` and `delete` on core `services`. The control plane only annotates existing Services with `ingress.kubernetes.io/service-upstream`.
- All access to, and webhook rules for, the removed `kuma.io` types: the legacy policies (`proxytemplates`, `ratelimits`, `trafficpermissions`, `trafficroutes`, `timeouts`, `retries`, `circuitbreakers`, `virtualoutbounds`, `faultinjections`, `healthchecks`, `trafficlogs`, `traffictraces`), `externalservices`, `serviceinsights`, `zoneingresses`, `zoneegresses`, `zoneingressinsights`, `zoneegressinsights`, and `meshgateways`, `meshgatewayroutes`, `meshgatewayinstances`, `meshgatewayconfigs` with their `/status` and `/finalizers` subresources.

Leftover permissions and webhook rules for removed types are harmless but unused.

### Policies and API

#### Resources with fields that are not in the schema are rejected

Applying a policy or resource with a field that does not exist in its schema now fails with `400` listing every unknown field, for example `spec.from: unknown field`. Previously such fields were silently dropped, so a policy written for an older version, such as a `MeshTrafficPermission` with `spec.from` instead of `spec.rules`, was stored without it and looked applied while doing nothing.

The check covers policies and resources with a generated schema, and applies to the Kuma API server and `kumactl apply`. Legacy resources without a generated schema, such as `Mesh` and `Dataplane`, still drop unknown fields silently. On Kubernetes, `kubectl apply` behavior is unchanged: the API server prunes unknown fields and prints a warning. Resources stored before the upgrade are not re-validated; a removed field they still carry is ignored. The sections below that remove fields follow these rules.

**Action required**

Fix manifests that stop applying: remove the reported field or use its current equivalent, for example `spec.rules` instead of `spec.from` on inbound policies.

#### Names of `Mesh`, `Zone`, `MeshService`, `MeshExternalService` and `MeshMultiZoneService` must be RFC 1035 labels

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

#### Reserved and control-plane-owned labels are enforced

**Reserved prefixes.** `kuma.io/` and `k8s.kuma.io/` are reserved label prefixes. Every unknown label under these prefixes is rejected on create and update. On Universal this includes the labels of the `Dataplane` passed to `kuma-dp run`: a proxy whose `Dataplane` carries an unknown reserved label, such as a leftover `kuma.io/gateway: "true"`, or an invalid label value fails to register until the label is fixed.

**Computed labels.** The control plane computes `kuma.io/origin`, `kuma.io/zone`, `kuma.io/env`, `kuma.io/mesh`, `kuma.io/display-name`, `kuma.io/policy-role`, `k8s.kuma.io/namespace`, `k8s.kuma.io/service-account`, `kuma.io/listener-zoneingress` and `kuma.io/listener-zoneegress` on every write. Depending on the label and the platform, a value you supplied used to be either rejected or silently replaced. Now, on both Kubernetes and Universal, a supplied value that differs from the computed one is rejected, and an equal or absent one is accepted. On an update, a value equal to the stored one is accepted too, so re-applying a stored resource with `kubectl apply` or `kumactl apply` keeps working; the stored value is then replaced with the computed one, for example when the edit changes the policy role. On Universal, plugin policies now also carry `kuma.io/policy-role: system`.

**Controller-only labels.** `kuma.io/managed-by`, `kuma.io/deletion-grace-period-started-at`, `k8s.kuma.io/service-name` and `k8s.kuma.io/is-headless-service` are set only by the control plane's own controllers, and `k8s.kuma.io/service-account` only by the pod controller. You cannot set any of them, not even to their stored value: an update of a generated `MeshService`, `Workload` or `Dataplane` that carries them is rejected, because accepting it would drop the label.

**`k8s.kuma.io/service-account` on a `Dataplane`.** On Kubernetes the admission webhook rejects any `Dataplane` create or update that carries this label, unless the request comes from the control plane itself or from another user listed in `runtime.kubernetes.allowedUsers`, on both Zone and Global control planes. A proxy is also rejected at xDS authentication when the label on its `Dataplane` does not match the ServiceAccount of its Pod. Resources synced over KDS and resources written by the control plane are unaffected.

**`kuma.io/origin` on a zone.** A zone control plane used to reject a resource applied without `kuma.io/origin: zone` on the system namespace on Kubernetes, or on any federated zone on Universal, unless `multizone.zone.disableOriginLabelValidation` was set. A missing label is now accepted and the zone sets it itself. A label with any value other than `zone` is still rejected, which is what keeps resources synced from the global control plane read-only on the zone. The `disableOriginLabelValidation` setting (`KUMA_MULTIZONE_ZONE_DISABLE_ORIGIN_LABEL_VALIDATION`) is removed.

**Federation export.** `kumactl export --profile federation` and `--profile federation-with-policies` now leave out every control-plane-owned label except `kuma.io/mesh`, so the global control plane accepts the output.

**`Zone` on Kubernetes.** The defaulting webhook selected `zone` where the CRD plural is `zones`, so a `Zone` written straight to the Kubernetes API skipped the webhook. It now matches, so a `Zone` created or updated with `kubectl` gets the same computed labels a `Zone` created through the HTTP API already got: `kuma.io/display-name`, `kuma.io/origin`, and on a zone control plane `kuma.io/zone` and `kuma.io/env`. Existing `Zone` resources get them the next time something writes to them.

**Action required**

- Remove unknown `kuma.io/` and `k8s.kuma.io/` labels from your manifests and from the `Dataplane` files passed to `kuma-dp`.
- Remove control-plane-owned labels from your manifests, or set them to the value the control plane shows on the stored resource. Don't edit a resource the control plane generated; change what it is generated from, such as the `Service` or the `Pod`.
- Remove `k8s.kuma.io/service-account` from any `Dataplane` you apply yourself, including GitOps-managed ones. A `Dataplane` whose label does not match its Pod stops connecting after the upgrade until the label is dropped.
- Remove `multizone.zone.disableOriginLabelValidation` from your configuration. A control plane started with the setting still present ignores it.
- Re-export any federation output made with an older `kumactl` before applying it to the global control plane.

#### `kuma.io/mesh` annotation no longer honored on Kubernetes

The control plane no longer reads the deprecated `kuma.io/mesh` annotation on a Pod, Service, HTTPRoute, or Namespace to assign the resource's mesh. Only the `kuma.io/mesh` label is used: on the resource itself, or, for namespaced resources (Pod, Service, HTTPRoute), on their Namespace. Resources without the label fall back to the `default` mesh.

**Action required**

If you still set `kuma.io/mesh` as an annotation, switch it to a label before upgrading. A resource that only carries the annotation resolves to the `default` mesh after upgrading.

#### A resource's `Mesh` ownerReference must match its mesh

On Kubernetes, a mesh-scoped resource can no longer end up owned by one `Mesh` while belonging to another. The validating webhook now rejects any create or update where a `kuma.io` `ownerReference` of kind `Mesh` names a mesh different from the resource's own `kuma.io/mesh` label (or `mesh` field). This closes both ways such a mismatch could previously happen: editing `kuma.io/mesh` on an existing resource to move it between meshes in place, and hand-writing or hand-editing a `Mesh` ownerReference. Writes from the control plane, `generic-garbage-collector` and the storage version migrator are unaffected, so a `Dataplane` can still be moved between meshes as its Pod is rescheduled.

**Action required**

If a resource in your cluster already has a `Mesh` ownerReference that disagrees with its mesh, delete it and re-apply it in the correct mesh. The webhook rejects any further edit to it until then.

#### `from` removed from inbound policies

The deprecated `spec.from` array is removed from `MeshTrafficPermission`, `MeshTLS`, `MeshTimeout`, `MeshCircuitBreaker`, `MeshAccessLog` and `MeshRateLimit`. Inbound traffic is configured exclusively through `spec.rules`.

A stored policy that still carries `from` keeps loading, but `from` is ignored and has no effect on inbound configuration. On a new write, `kumactl apply` and the Kuma API server reject `from` as an unknown field, and on Kubernetes the API server prunes it with a warning. A spec left with neither `to` nor `rules` is rejected by validation.

Policy-specific consequences:

- **`MeshTrafficPermission`**: `rules` matches clients by `MeshIdentity` (`spiffeID`) or SNI instead of by dataplane tag subsets. A `MeshTrafficPermission` must define at least one entry in `rules`; a spec with only `from` (or with neither) fails validation with `policy must define rules`. A proxy matched only by a `from`-based policy is default-denied once that policy is migrated or removed, unless a `rules`-based policy allows the same traffic.
- **`MeshTLS`**: an un-migrated `from` leaves the proxy on the default inbound mode, which is `Strict` for a proxy with a workload identity. This breaks connectivity if the `from` entry was intentionally `Permissive`.
- **`MeshTimeout`, `MeshCircuitBreaker`, `MeshAccessLog`, `MeshRateLimit`**: inbound listeners fall back to the defaults, or to whatever a `rules`-based policy sets.

**Action required**

Before upgrading, migrate every one of these policies that uses `spec.from` to `spec.rules`. A `from` entry targeting `kind: Mesh` (all clients) maps to a single catch-all rule with the same `default`:

```yaml
# before
spec:
  targetRef:
    kind: Mesh
  from:
    - targetRef:
        kind: Mesh
      default:
        idleTimeout: 1h
# after
spec:
  targetRef:
    kind: Mesh
  rules:
    - default:
        idleTimeout: 1h
```

For `MeshTrafficPermission`, express each `from` entry as `rules[].default.allow`, `deny` or `allowWithShadowDeny` matches on `spiffeID` or `sni`. Audit with:

```bash
for kind in meshtrafficpermissions meshtlses meshtimeouts meshcircuitbreakers meshaccesslogs meshratelimits; do
  kubectl get "$kind" -A -o yaml | grep -B5 'from:'
done
```

#### Empty `MeshTrafficPermission` match entries are rejected

The 2.14 validator accepted empty match objects in `rules[].default.allow`, `rules[].default.deny` and `rules[].default.allowWithShadowDeny`. The 3.0 validator rejects those entries on create and update, including KDS policy writes through the Kubernetes admission webhook. Each entry must contain `spiffeID`, `sni`, or both. An empty action list is allowed only when another action has at least one valid match.

**Action required**

Before upgrading the control plane, remove or replace each empty match object in stored `MeshTrafficPermission` policies. The upgrade does not rewrite stored policies.

#### Legacy `targetRef` kinds removed and real resources selected by `labels` only

**Top-level `targetRef`.** Policies no longer accept `spec.targetRef.kind: MeshSubset`, `MeshService`, `MeshServiceSubset` or `MeshGateway`. Creating or updating a policy with one of these kinds is rejected. This applies to every admission path, including policies replicated from a Global control plane to Kubernetes Zone control planes by KDS, so a policy still using one of them fails to sync into a Zone after the Zone is upgraded.

**`spec.to[].targetRef`.** No policy accepts `kind: MeshSubset`, `MeshServiceSubset` or `MeshGateway` there. The remaining accepted kinds are per policy:

- Every policy that has a `to` array accepts `kind: Mesh`.
- Most policies also accept `MeshService`, `MeshExternalService`, and `MeshMultiZoneService`.
- Only `MeshAccessLog`, `MeshLoadBalancingStrategy`, `MeshRetry`, and `MeshTimeout` accept `kind: MeshHTTPRoute`. For example, `MeshCircuitBreaker` rejects it.
- `MeshRateLimit` and `MeshFaultInjection` accept `kind: Mesh` only.

**Selection by `labels`.** Policies that select a `Dataplane`, `MeshService`, `MeshExternalService`, `MeshMultiZoneService` or `MeshHTTPRoute` through `spec.targetRef` or `spec.to[].targetRef` resolve the target by `labels` only, not by `name` or `namespace`. `sectionName` remains supported for `Dataplane` inbound selection and `MeshService` port selection.

**Action required**

Before upgrading, migrate every affected policy:

- Use `kind: Mesh` when the policy should apply to the whole mesh.
- Use `kind: Dataplane` with `labels` when the policy should apply to a subset of dataplanes previously selected by `MeshSubset`, `MeshService`, or `MeshServiceSubset`.
- Replace `name`/`namespace` selection of real resources with `labels`, for example `kuma.io/display-name` and `k8s.kuma.io/namespace`.

Verify the intended policy coverage in every Zone before upgrading Zone control planes.

#### Legacy policies removed

The 12 legacy policy types superseded by the `Mesh*` `targetRef` policies are removed: `TrafficPermission`, `TrafficRoute`, `TrafficLog`, `TrafficTrace`, `HealthCheck`, `CircuitBreaker`, `Retry`, `Timeout`, `RateLimit`, `FaultInjection`, `VirtualOutbound` and `ProxyTemplate`.

For every one of these types, the resource, its CRD, REST API endpoints (including the generic `_resources`/`_dataplanes` inspect endpoints), `kumactl get`/`inspect` subcommands, KDS sync, `MeshInsight` policy counters, control plane RBAC and admission webhook rules are gone. None of them affects generated Envoy configuration. The `ProxyTemplate` profile mechanism is gone as well; the control plane always generates the standard Envoy configuration for every data plane proxy.

What changes in traffic for meshes that still relied on them:

- **`TrafficPermission`**: `MeshTrafficPermission` is the only policy that controls inbound and egress access. An inbound, external service or gateway route that relied solely on a `TrafficPermission` grant now defaults to deny instead of allow. This is fail-closed: every mTLS inbound listener and egress external-service filter chain still gets a default-deny RBAC filter. External-service outbound clusters are always generated, instead of only when a `TrafficPermission` granted access.
- **`TrafficRoute`**: traffic between services always flows by default, as it already did in meshes that only use `MeshHTTPRoute`/`MeshTCPRoute`. The legacy behavior where just having a `TrafficRoute` disabled the default routing fallback is gone.
- **`TrafficLog`, `TrafficTrace`, `Timeout`, `HealthCheck`, `CircuitBreaker`, `Retry`, `RateLimit`, `FaultInjection`**: their access log, tracing, timeout and connection settings no longer apply; proxies get the defaults unless a `Mesh*` policy sets them.
- **`VirtualOutbound`**: it no longer adds outbounds, DNS domains or zone ingress destinations.
- **`ProxyTemplate`**: neither profile imports nor raw xDS `resources` and `modifications` are applied.

The control plane also no longer creates the default allow-all `TrafficPermission` and route-all `TrafficRoute` when a `Mesh` is created, and the compatibility mode that created them for pre-2.6.0 zones is gone. `defaults.createMeshRoutingResources` (`KUMA_DEFAULTS_CREATE_MESH_ROUTING_RESOURCES`) is removed.

**Action required**

Before upgrading, migrate and then delete every remaining legacy policy. Stored resources of a removed type are not migrated, and the control plane no longer accepts them:

| Removed | Replacement |
|---|---|
| `TrafficPermission` | `MeshTrafficPermission` |
| `TrafficRoute`, `VirtualOutbound` | `MeshHTTPRoute`, `MeshTCPRoute` |
| `TrafficLog` | `MeshAccessLog` |
| `TrafficTrace` | `MeshTrace` |
| `Timeout` | `MeshTimeout` |
| `HealthCheck` | `MeshHealthCheck` |
| `CircuitBreaker` | `MeshCircuitBreaker` |
| `Retry` | `MeshRetry` |
| `RateLimit` | `MeshRateLimit` |
| `FaultInjection` | `MeshFaultInjection` |
| `ProxyTemplate` | `MeshProxyPatch` |

Remove `defaults.createMeshRoutingResources` from your control plane configuration, and remove automation, dashboards and `kumactl` scripts that reference these types, their REST paths or CRDs. On Kubernetes, delete the leftover CRDs once you no longer need the objects in them. If you copied the Kuma `ClusterRole` or webhook configuration into your own manifests, see [Control plane RBAC changes on Kubernetes](#control-plane-rbac-changes-on-kubernetes).

#### Meshes no longer come with default policies

Creating a `Mesh` used to create four policies with it: `mesh-timeout-all-<mesh>`, `mesh-timeout-to-all-<mesh>`, `mesh-circuit-breaker-all-<mesh>` and `mesh-retry-all-<mesh>`. Kuma 3.0 creates none of them, so a new mesh starts with no policies at all, and `skipCreatingInitialPolicies` is removed from the `Mesh` spec.

A mesh that already has these policies keeps them. The control plane neither recreates nor deletes them, so an upgraded mesh behaves as it did before, and `kumactl delete meshtimeout -m <mesh> mesh-timeout-all-<mesh>` removes one when you no longer want it.

Timeouts and circuit breakers are unchanged, because the values those policies carried are the ones the control plane writes anyway when no policy selects a listener, a cluster or a route. Outbound keeps its 5s connect timeout, 1h cluster idle timeout, 15s request timeout and 30m stream idle timeout. Inbound keeps the larger values it is meant to have, so the side that receives a request never cuts it before the side that sent it: a 10s connect timeout, a 2h cluster idle timeout, a 1h stream idle timeout and no request timeout at all. Circuit breakers fall back to Envoy's own limits, which are what the removed policy set: 1024 connections, 1024 pending requests, 1024 requests and 3 retries.

Retries are the one thing that changes. A mesh without a `MeshRetry` does not retry: the removed policy retried an HTTP or gRPC request 5 times with a 16s per-try timeout and a 25ms to 250ms backoff, and made 5 TCP connection attempts.

**Action required**

- Apply a `MeshRetry` of your own to any new mesh that needs requests retried.
- Drop `skipCreatingInitialPolicies` from your `Mesh` manifests. The control plane ignores the field, so a manifest that still sets it applies without an error, and a mesh stored with it loads fine. The first write to such a mesh drops the field, which matters only if you then roll back to 2.14: that mesh gets the default policies created again.

#### `healthyPanicThreshold` removed from `MeshHealthCheck`

The deprecated `to[].default.healthyPanicThreshold` field is removed from `MeshHealthCheck`. Use `to[].default.outlierDetection.healthyPanicThreshold` on `MeshCircuitBreaker` instead.

**Action required**

Before upgrading, move the setting to a `MeshCircuitBreaker` policy. An un-migrated value is ignored after the upgrade, and affected clusters fall back to Envoy's default panic threshold of 50%.

#### `MeshLoadBalancingStrategy` drops nested `hashPolicies` and restricts `crossZone`

- `spec.to[].default.loadBalancer.ringHash.hashPolicies` and `spec.to[].default.loadBalancer.maglev.hashPolicies` are removed. `spec.to[].default.hashPolicies` is the only place to configure hash policies.
- `spec.to[].default.localityAwareness.crossZone` is accepted only when that `to` entry targets a `MeshMultiZoneService`. Validation rejects it on `Mesh`, `MeshService` and `MeshExternalService` targets.

**Action required**

Before upgrading:

- Move `hashPolicies` from `loadBalancer.ringHash` or `loadBalancer.maglev` to `spec.to[].default.hashPolicies`. Un-migrated nested fields no longer affect the generated Envoy config.
- Move `crossZone` configuration under `to` entries whose `targetRef.kind` is `MeshMultiZoneService`. Keep `localityAwareness.localZone`, `loadBalancer` and other settings on the remaining target kinds as needed.

#### `MeshPassthrough` resolves domain matches itself and rejects conflicting matches

**A `Domain` match needs a port.** A `Domain` match used to build an `ORIGINAL_DST` cluster: the sidecar matched the SNI or the `Host` header and then sent the request to the address the client dialed. A workload selected by the policy could therefore dial any address, present an allowed domain, and reach that address through the policy, which is the opposite of what an allowlist in `passthroughMode: Matched` is for. A `Domain` match now resolves the domain in the sidecar and connects to the resolved address. Resolving needs a port, so a `Domain` that is not a wildcard requires `port`, and a policy that has one without a port is rejected on apply. In a policy stored before the upgrade that match stops applying, and when it was the only match in the policy nothing is left to allow, so the sidecar rejects all passthrough traffic until you add the port.

A wildcard `Domain`, for example `*.example.com`, has no address to resolve, so its traffic still goes to the address the client dials and the match only restricts the SNI or `Host`. Keep that in mind when a wildcard entry is part of an allowlist.

**Matches that resolve to the same filter chain are rejected.** Previously such a policy was accepted and Envoy rejected the entire passthrough listener, breaking all passthrough traffic for every proxy the policy matched. Two matches collide when they configure the same port (or both configure no port) with:

- two of `grpc`, `http` and `http2`, which share one filter chain per port
- the same address spelled differently: an IP and a CIDR covering only that IP (`10.0.0.1` and `10.0.0.1/32`), a CIDR with host bits (`10.0.0.1/24` and `10.0.0.0/24`), or another textual form of the same IPv6 address
- `tcp` and `mysql` on the same address, which generate an identical TCP proxy filter chain

`tls` and `http` on the same port, or `http` on an IP next to `grpc` on a domain, generate distinct filter chains and are not reported. A match with a port next to a match without one is not a conflict either: the port-specific match owns its port and the match without a port covers the remaining ports.

An already applied policy with a conflict is not re-validated on upgrade. The control plane keeps the first match of the colliding pair in `appendMatch` order, drops the later one and names it in a debug log of the `MeshPassthrough` component, so proxies previously stuck with a rejected listener recover on their own.

**Action required**

- Add `port` to every `Domain` match that is not a wildcard. Duplicate the match if the domain is used on more than one port:

  ```yaml
  # before
  appendMatch:
  - type: Domain
    value: api.example.com
    protocol: tls
  # after
  appendMatch:
  - type: Domain
    value: api.example.com
    port: 443
    protocol: tls
  ```

- Make sure the sidecar can resolve those domains. A domain the sidecar cannot resolve has no endpoint, so its traffic fails instead of following the original destination.
- Resolve conflicting matches (one protocol per port and one spelling per address), otherwise the next edit of the policy is rejected.

#### `MeshHTTPRoute` follows Gateway API matching and backend semantics

##### A request matching no rule gets a `404`

When a `MeshHTTPRoute` applies to a destination, a request that matches none of its rules is answered with `404` instead of being sent to that destination, as the Gateway API requires of an `HTTPRoute` and the GAMMA conformance suite asserts.

The rules of a `MeshHTTPRoute` apply to every HTTP port of the destination when the `to[].targetRef` names a `MeshService` without a `sectionName`, so the `404` covers those ports too, including ports no rule mentions. Ports whose protocol is not HTTP based are unaffected. A destination with no `MeshHTTPRoute` at all keeps reaching its service.

The case worth checking is a route written only to anchor another policy. A `MeshHTTPRoute` matching `/api` that exists so a `MeshTimeout`, `MeshRetry`, or `MeshAccessLog` can target it now answers `404` on every other path of that destination. On a gRPC destination the same `404` reaches the client as `UNIMPLEMENTED`.

##### A match without a path ranks as `PathPrefix /`

A match that sets only `headers`, `method`, or `queryParams` is now ordered as if it had `path: { type: PathPrefix, value: / }`, as the Gateway API assumes. Before, any match with a path ranked above it. Two orderings change:

- A pathless match next to an explicit `PathPrefix /` match now ties on path, so the method, header count, and query param count decide. A header-only rule (for example a canary on `x-canary: true`) next to a `PathPrefix /` catch-all used to be shadowed by the catch-all; it now receives the requests that carry the header.
- A pathless match now ranks above a `RegularExpression` match. With a rule on `/api/.*` and a rule matching only the header `x-debug`, a request to `/api/foo` with `x-debug` used to go to the regex rule and now goes to the header rule.

Exact matches and `PathPrefix` matches longer than `/` still rank above a pathless match. Routes converted from Gateway API `HTTPRoute`s are unaffected, because their matches always carry a path.

##### A rule whose `backendRefs` all fail to resolve answers `500`

A rule that declares `backendRefs` and resolves none of them no longer falls back to the destination service; it serves `500` to every request it matches, as the Gateway API requires for an invalid backendRef. A rule that resolves at least one of its backendRefs keeps routing to those backends, and a rule with a `RequestRedirect` filter still redirects. This covers every `MeshHTTPRoute`, not only the ones generated from Gateway API. A route pointing at a resource that does not exist yet, such as a `MeshService` that KDS has not synced to the zone, fails its matching requests until the reference resolves.

A `backendRef` whose destination exists but lacks the referenced `sectionName`/port no longer counts as resolved either. Previously it silently dropped out of the split: a missing `Service` port fell through to the parent service on a different port, and a missing `MeshService` port fell through to another port of that `MeshService`. The Gateway API `HTTPRoute` translation reports `ResolvedRefs=False` with reason `BackendNotFound` for a `Service` backend missing the requested port. Because a `MeshService`'s ports can be briefly empty while it converges, a route that references one of its ports can fail matching requests during that window.

**Action required**

- Add an explicit catch-all rule to any `MeshHTTPRoute` that should keep passing unmatched traffic to its destination. Put it on the route itself when the other policies targeting that route should also cover the unmatched traffic, or on a second `MeshHTTPRoute` when they should not. It no longer shadows header-only rules:

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

- Review routes that combine pathless matches with a `PathPrefix /` or `RegularExpression` match. If a pathless rule should only apply under a `RegularExpression` path, add that path to the rule.
- Check that every `backendRefs` entry names a resource and a port that exist, especially in routes that reference a resource from another zone or namespace, or set an explicit port or `sectionName`.

#### Route `backendRef` no longer accepts `MeshServiceSubset`

`MeshHTTPRoute` and `MeshTCPRoute` accept only `MeshService`, `MeshExternalService`, and `MeshMultiZoneService` in `backendRefs[]`, and the `MeshHTTPRoute` `RequestMirror` filter's `backendRef` is limited to the same kinds. A route using `kind: MeshServiceSubset` is rejected with `value 'MeshServiceSubset' is not supported`. The `tags` field is gone from the `backendRef` schema entirely.

The xDS path that turned such a ref into a `kuma.io/service` cluster is gone with it, and a `MeshService` ref selected by `kuma.io/display-name` alone no longer falls back to one when it fails to resolve. A stored route that still carries a `MeshServiceSubset` backendRef keeps loading, but the backendRef is treated as unresolved, so traffic matching those rules loses its destination.

**Action required**

Before upgrading, repoint every route that splits or mirrors to a `MeshServiceSubset` at the real resource. Selecting a subset of endpoints by tag has no equivalent: split the destination into separate `MeshService` resources if you need it.

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

#### Route schema defaults for `type` and `weight` are no longer materialized

A linter bug hid a set of API fields from the shape checks the rest of the API follows. Fixing them drops two declared schema defaults:

- `MeshHTTPRoute` and `MeshRetry` header matches no longer declare a default of `Exact` for `type`. An omitted `type` is still matched as `Exact`.
- `MeshHTTPRoute` and `MeshTCPRoute` backend refs no longer declare a default of `1` for `weight`. An omitted `weight` still counts as `1`.

**Action required**

None. A resource that omits `type` or `weight` no longer comes back from the API with the value filled in.

#### `MeshProxyPatch` can now change circuit breaker thresholds

Since 2.14.0 every generated cluster carries a `DEFAULT`-priority circuit breaker threshold, and a `MeshProxyPatch` patching `circuitBreakers` appended a second one for that same priority. Envoy honours only the first threshold matching a routing priority, so the patched entry was dead config: `/config_dump` showed the requested values while Envoy kept what `MeshCircuitBreaker`, or its own defaults, had set. The patch now merges into the existing threshold of the same priority instead of appending. A threshold whose priority is not yet on the cluster still appends, and a `value` listing the same priority twice keeps the first entry, matching how Envoy resolves them.

**Action required**

Review every `MeshProxyPatch` that patches `circuitBreakers`. Where the same cluster is also covered by a `MeshCircuitBreaker`, the patch now overrides that policy for each field it sets, instead of being ignored: `MeshProxyPatch` runs last, so it wins the fields it names and the policy keeps the rest. Remove patches you wrote before 2.14.0 and no longer rely on, and drop any workaround you put in place because the patch appeared to do nothing.

#### Inline `Mesh` settings removed

The `Mesh` fields that duplicated a policy or a mesh-level toggle are removed:

| Removed `Mesh` field | Replacement |
|---|---|
| `spec.tracing` (`Tracing`, `TracingBackend`, `DatadogTracingBackendConfig`, `ZipkinTracingBackendConfig`) | `MeshTrace` |
| `spec.logging` (`Logging`, `LoggingBackend`, `FileLoggingBackendConfig`, `TcpLoggingBackendConfig`) | `MeshAccessLog` |
| `spec.routing.localityAwareLoadBalancing` | `MeshLoadBalancingStrategy` |
| `spec.networking.outbound.passthrough` | `MeshPassthrough` |
| `spec.routing.zoneEgress` | none, zone egress is used based on the actual zone egress topology plus mTLS |
| `spec.routing.defaultForbidMeshExternalServiceAccess` | `MeshTrafficPermission` on `MeshExternalService` traffic |

`Mesh` drops unknown fields silently, so a `Mesh` that still sets one of them keeps applying, and the field is ignored by the control plane.

`spec.networking.outbound.passthrough` no longer decides passthrough. A proxy that no `MeshPassthrough` selects follows `defaults.restrictOutbound`, see [Outbound traffic is restricted by default](#outbound-traffic-is-restricted-by-default).

**Action required**

Before upgrading, move each setting to its replacement and remove the fields from stored `Mesh` resources and manifests. For `networking.outbound.passthrough: false`, apply a `MeshPassthrough` with `targetRef.kind: Mesh` and `default.passthroughMode: None`.

#### Legacy REST API endpoints removed

These endpoints are removed and answer `404`:

| Removed | Use instead |
|---|---|
| Cross-mesh lists of mesh-scoped resources, such as `GET /meshaccesslogs`, `GET /dataplanes`, `GET /meshservices`, `GET /secrets` and their `/_overview` variants | `GET /meshes/{mesh}/<type>`, for example `GET /meshes/{mesh}/dataplanes`. Global-scoped resources such as `GET /globalsecrets` are unchanged |
| `GET /meshes/{mesh}/dataplanes+insights[/{name}]` | `GET /meshes/{mesh}/dataplanes/_overview`, `GET /meshes/{mesh}/dataplanes/{name}/_overview` |
| `GET /zones+insights[/{name}]` | `GET /zones/_overview`, `GET /zones/{name}/_overview` |
| `GET /meshes/{mesh}/{policyType}/{policyName}/dataplanes` | `GET /meshes/{mesh}/{policyType}/{policyName}/_resources/dataplanes` |
| `GET /meshes/{mesh}/dataplanes/{name}/rules` | `GET /meshes/{mesh}/dataplanes/{name}/_policies`, or the `_inbounds/{inbound_kri}/_policies`, `_outbounds/{outbound_kri}/_policies` and `_outbounds/{outbound_kri}/_routes/{route_kri}/_policies` variants |
| `GET /meshes/{mesh}/meshservices/{name}/_resources/dataplanes` | `GET /meshes/{mesh}/meshservices/{name}/_dataplanes` |
| `GET /meshes/{mesh}/service-insights[/{name}]` | `MeshService` and `MeshExternalService`, see [The `ServiceInsight` resource is removed](#the-serviceinsight-resource-is-removed) |

The endpoints of removed resource types (legacy policies, `ExternalService`, built-in gateway and standalone zone proxy resources) are gone with them.

The inspect `_rules` endpoint no longer returns the legacy `toRules` and `fromRules` fields on each rule entry; `toResourceRules` and `inboundRules` are unaffected. `GET /meshes/{mesh}/dataplanes/{dataplane}/policies` is still served for the bundled GUI but is deprecated; use `_policies`.

`kumactl` follows these changes:

- `kumactl inspect <policy> NAME` uses `_resources/dataplanes` and returns dataplane metadata rather than the legacy attachment details. Results are paginated with `--size` (100 by default) and `--offset`. `--new-api` is still accepted but has no effect.
- `kumactl inspect dataplane --type=policies` calls `_policies` and needs no change.
- `kumactl inspect services` is removed; use `kumactl get meshservices`.

**Action required**

Upgrade `kumactl` together with the control plane, and update scripts, automation and dashboards that call the removed endpoints or read `toRules`/`fromRules`, or that depend on the legacy `kumactl inspect <policy>` attachment output.

#### The `ServiceInsight` resource is removed

`ServiceInsight` is gone: the control plane no longer registers the type, the `serviceinsights` CRD is no longer installed, control plane RBAC no longer grants access to it, and `GET /meshes/{mesh}/service-insights[/{name}]` answers `404`. The control plane stopped writing `ServiceInsight` when services became computed from `MeshService` and `MeshExternalService`, so the endpoints had been returning an empty result for every mesh. The GUI pages backed by them list nothing.

Rows left behind by older versions are no longer cleaned up. They are inert: nothing reads them, they are not served over the API and they are not synced between zones. A stored row carrying a value removed in 3.0 no longer breaks the insight resyncer (issue #18330).

**Action required**

None if you were reading these endpoints, since they no longer returned data. To list services in a mesh, use `MeshService` and `MeshExternalService`.

To reclaim the space, delete the leftover data after upgrading. On Kubernetes, `kubectl delete crd serviceinsights.kuma.io` removes the custom resources along with the definition, which Helm leaves in place on upgrade. On Universal, delete the rows whose resource type is `ServiceInsight` from the `resources` table. Do this only once you are sure you will not roll back, since neither is reversible.

#### `MeshInsight` and global insight fields removed

`MeshInsight` loses the fields that only repeated other totals or were fed by removed features:

- `MeshInsight.dataplanesByType` is removed (`standard`, `gateway`, `gatewayBuiltin`, `gatewayDelegated`). `MeshInsight.dataplanes` carries the same totals across every proxy in the mesh.
- `MeshInsight.services` is removed (field number reserved). Per-service status comes from `MeshService` and `MeshExternalService`; the aggregated `internal`/`external` counts live under `services` in `/global-insight`.
- `MeshInsight.policies` is removed. `MeshInsight.resources` reports a `total` for every resource type, policies included, keyed by the same type name.

The `/global-insight` response changes with it:

- `dataplanes.standard` is replaced by a flat `dataplanes` object with `total`, `online`, `offline` and `partiallyDegraded`.
- `dataplanes.gatewayBuiltin`, `dataplanes.gatewayDelegated`, `services.gatewayBuiltin` and `services.gatewayDelegated` are removed.
- `zones.zoneIngresses` and `zones.zoneEgresses` are removed. They were fed by `ZoneIngressInsight` and `ZoneEgressInsight` and had been reporting zero.

`GET /meshes/{mesh}/dataplanes/_overview?gateway=` is no longer a valid filter and is ignored.

**Action required**

Update automation and dashboards that read the removed fields (through the REST API or `kumactl inspect meshes`). Take dataplane totals from `dataplanes`, policy counts from `MeshInsight.resources`, and service counts from the `/global-insight` `services` object. Zone proxy counts are no longer served by the insight API; list the `Dataplane`s instead.

#### `GET /_resources` reports control-plane writability and drops dead policy fields

- `readOnly` now reports whether generic `PUT` and `DELETE` are disabled for the resource type on the current control plane. On Global control planes, resources provided by Zones report `readOnly: true`. On writable federated Zone control planes, resources provided by the Zone report `readOnly: false`. `GET /policies` already used these semantics and is unchanged.
- `policy.hasFromTargetRef`, `policy.isFromAsRules` and `policy.isTargetRef` are removed. The first two had been hardcoded to `false` since the `from`-style `targetRef` was dropped, and `isTargetRef` was `true` for every policy.

**Action required**

Dynamic clients should treat `readOnly` as the capability of the current control plane, not as an intrinsic property of the resource type. Treat every entry with a `policy` object as a `targetRef` policy, and read `policy.hasRulesTargetRef` to tell whether a policy applies to inbound traffic. No action is needed for Kubernetes installations using the default read-only API configuration.

### Data plane and traffic

#### `meshServices` is always `Exclusive`

The `meshServices` field (and its `mode` enum) no longer has any effect: the control plane behaves as if every mesh were `Exclusive`, and unified resource naming is unconditional. In particular, the control plane generates mesh-scoped zone proxy listeners, initializes `MeshIdentity` and serves the Dataplane inspect `_layout` and policy endpoints in every mode. The Kubernetes warning event `ZoneProxyListenersSkipped` and the `MeshIdentity` status reason `MeshServicesDisabled` are removed, because nothing is skipped anymore.

The field remains in the schema as deprecated so that stored values survive the upgrade and keep syncing to zones over KDS: zones before 3.0 read a missing field as `Disabled`, which makes them delete every generated `MeshService`, skip mesh-scoped zone proxy listeners, and stop serving `MeshService` outbounds and DNS. Zones on 3.0 ignore the field, and writes setting it to any mode other than `Exclusive` are rejected, because that mode would silently behave as `Exclusive`.

**Action required**

- Set `meshServices.mode: Exclusive` on every mesh before upgrading the global control plane, including meshes that never set the field, because a 2.x zone reads a missing field as `Disabled`. The 3.0 global then keeps syncing the stored mode, and 2.x zones keep serving `MeshService`s until they are upgraded. A mesh that still carries another mode when the global is upgraded is rejected by 3.0 zones over KDS until it is set to `Exclusive`.
- A write that carries the field with `Exclusive` still applies and returns a deprecation warning; drop the field from your manifests once every zone is on 3.0.
- Remove alerts on `ZoneProxyListenersSkipped` and automation that expected `MeshServicesDisabled` or treated inspect `_layout` as unavailable outside `Exclusive` mode.

#### `Dataplane` outbounds are defined by `backendRef` only

`networking.outbound[].tags` is removed from the `Dataplane` schema. An outbound is defined exclusively by `backendRef`, pointing at a `MeshService`, `MeshExternalService`, or `MeshMultiZoneService`. Creating or updating a `Dataplane` whose outbound has no `backendRef` is rejected with `backendRef: must be defined`.

The xDS generation path for `kuma.io/service`-tagged outbounds is gone with it. Policies match the referenced resource instead of a `kuma.io/service` value, and locality awareness for cross-zone traffic through zone egress is no longer applied on that path. A stored `Dataplane` that still carries `tags` on an outbound keeps being served, but no outbound listener is generated for it, so the workload loses connectivity to that destination.

Two smaller changes affect outbounds that already use `backendRef`:

- **Port name as section name.** An outbound with `backendRef: {kind: MeshService, name: backend, port: 80}`, where port `80` is named `http`, used to get the port number as its section name. It now gets the port name, so the Envoy listener, cluster and stat prefix change from `..._backend_80` to `..._backend_http`, and policies that target the port with `sectionName: http` now apply to it (before, they were skipped and the service-level or `Mesh` rule applied). Transparent proxy outbounds already used the port name.
- **Address always set.** `networking.outbound[].address` is now written as `127.0.0.1` when an outbound comes in without one, and the OpenAPI schema marks it as required. Data plane behavior is unchanged. A `Dataplane` stored before the upgrade keeps its empty address until it is applied again, and a global control plane serves whatever a zone sent it, so a client reading from a global control plane federated with older zones should still fall back to `127.0.0.1`.

**Action required**

Before upgrading, rewrite every Universal `Dataplane` that declares tag-based outbounds. On Kubernetes, outbounds are generated by the control plane and need no change.

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

Check policies that target a `MeshService` port with `sectionName`, and update dashboards and alerts that match on a numeric outbound stat prefix such as `_80`.

#### Inbound `tags` removed from `Dataplane`

`networking.inbound[].tags` is removed from the `Dataplane` resource. Tags for an inbound must now be set through the `Dataplane`'s own `metadata.labels` (Kubernetes) or the top-level `labels:` field (Universal). On Kubernetes the field was already ignored in favor of Pod labels, so this only affects Universal.

With it, the `kuma.io/protocol` inbound tag no longer sets the protocol. The protocol of an inbound is read only from `networking.inbound[].protocol`. On Kubernetes the field is always derived from the `Service` port.

`Dataplane` drops unknown fields silently, so a `Dataplane` that still submits inbound `tags` upgrades without error, but any policy matching on those tags stops matching, with nothing in the API to signal it.

**Action required**

Before upgrading, on every hand-authored Universal `Dataplane`:

- Move per-inbound tags to `Dataplane` labels.
- Set `networking.inbound[].protocol` on every inbound that declares its protocol only through the `kuma.io/protocol` tag.

**Warning**: an inbound without `protocol` is treated as an unknown protocol and served as plain TCP. Its listener loses the L7 filters that depend on it (HTTP access log fields, `MeshTimeout` HTTP timeouts, `MeshFaultInjection`, `MeshRateLimit` HTTP limits, HTTP-aware routing), and the endpoint stops advertising a protocol, so `MeshService` protocol inference falls back to TCP as well.

#### `dataplaneTags` removed from the `MeshService` selector

`spec.selector.dataplaneTags` matched data plane proxies by their inbound tags. It is removed; `MeshService` selects proxies by `spec.selector.dataplaneRef` or `spec.selector.dataplaneLabels` only.

A stored `MeshService` that still uses `dataplaneTags` keeps its name and ports but matches zero data plane proxies, so it stops producing endpoints and its status goes `Unavailable`. The control plane returns a warning for a `MeshService` left without a selector, but only when the resource is next created or updated.

**Action required**

Before upgrading, migrate every `MeshService` using `dataplaneTags` to `dataplaneLabels`. Audit with:

```bash
kubectl get meshservices -A -o yaml | grep -B5 'dataplaneTags:'
for mesh in $(kumactl get meshes -o json | jq -r '.items[].name'); do
  kumactl get meshservices --mesh "$mesh" -o yaml
done | grep -B5 'dataplaneTags:'
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

Inbound tags are no longer an identity source, so the replacement labels must exist on the `Dataplane` resource itself. On Kubernetes those come from the Pod labels; on Universal, set them under `labels` in the `Dataplane` resource.

#### `advertisedAddress` removed from `Dataplane` networking

`networking.advertisedAddress` is removed from the `Dataplane` resource. Proxies behind NAT or a private network (for example Docker) that relied on it to advertise a routable address to other proxies must be reachable directly via `networking.address`.

**Action required**

Ensure every Universal `Dataplane` is reachable by other proxies on `networking.address` before upgrading.

**Warning**: `Dataplane` drops unknown fields silently, so a `Dataplane` still submitting `advertisedAddress` is accepted and falls back to `networking.address` for xDS endpoints, Envoy admin mTLS SANs and `kumactl get dataplanes` output, which breaks connectivity for proxies not reachable on `networking.address`.

#### Redirect ports and IP family mode removed from `Dataplane`

Kuma 3.0 removes `ipFamilyMode` from `Dataplane.networking.transparentProxying` and deprecates `redirectPortInbound` and `redirectPortOutbound`. `directAccessServices` and `reachableBackends` stay. The control plane reads the redirect ports and the IP family mode from `kuma-dp`, which sends them when it runs with `--transparent-proxy` or `--transparent-proxy-config`.

A proxy that sends none, which means a sidecar injected before 3.0, still gets its redirect ports from the two deprecated fields, so upgrading the control plane does not take transparent proxying away from workloads that have not restarted yet. The IP family comes from what the proxy reports about its machine rather than from the removed `ipFamilyMode` field. Both fields go away in 3.1, so restart your workloads on 3.0 rather than leaving them on a pre-3.0 sidecar.

On Kubernetes, sidecars injected by a 3.0 control plane always pass the transparent proxy configuration to `kuma-dp`, and the control plane keeps writing the redirect ports onto the `Dataplane` of a pod injected by 2.14 for as long as that pod lives.

On Universal, a `Dataplane` that still sets the deprecated fields keeps working on 3.0 and stops working on 3.1.

Before:

```yaml
networking:
  address: 192.168.0.1
  inbound:
    - port: 8080
  transparentProxying:
    redirectPortInbound: 15006
    redirectPortOutbound: 15001
```

```sh
kuma-dp run --dataplane-file=backend.yaml
```

After:

```yaml
networking:
  address: 192.168.0.1
  inbound:
    - port: 8080
```

```sh
kuma-dp run --dataplane-file=backend.yaml --transparent-proxy
```

**Action required**

Nothing on 3.0. To be ready for 3.1, on Universal drop `redirectPortInbound`, `redirectPortOutbound`, and `ipFamilyMode` from your `Dataplane` manifests and `kuma-dp` dataplane files, and start `kuma-dp` with `--transparent-proxy`. If you installed the transparent proxy with a non-default IP family mode, redirect ports, inbound redirection, or virtual networks, pass the same values to `kuma-dp` in a file with `--transparent-proxy-config` instead:

```yaml
ipFamilyMode: ipv4
redirect:
  inbound:
    port: 15006
  outbound:
    port: 15001
```

`kuma-dp` 2.14 already supports both flags, so you can do this before or after the upgrade. On hosts with IPv6 disabled, `kuma-dp` 2.14 cannot start its DNS proxy in the default dual-stack mode, so use `--transparent-proxy-config` with `ipFamilyMode: ipv4` there.

On Kubernetes, restart your workloads at some point on 3.0. Until a pod restarts, its sidecar is the one 2.14 injected.

#### Reachable services replaced by `reachableBackends` with `labels`

`reachableBackends` (`kuma.io/reachable-backends` on Kubernetes) is the only way to trim the outbounds a proxy receives, and its refs select backends by `labels` only:

- **`reachableServices` removed.** `Dataplane.networking.transparentProxying.reachableServices` and the `kuma.io/transparent-proxying-reachable-services` annotation, which listed `kuma.io/service` tag values, are removed.
- **Auto reachable services removed.** The experimental feature that computed reachable services from `MeshTrafficPermission` is removed, with `experimental.autoReachableServices` (`KUMA_EXPERIMENTAL_AUTO_REACHABLE_SERVICES`). Traffic not permitted by a `MeshTrafficPermission` is still denied at the proxy; it is no longer pruned from the proxy configuration.
- **Refs select by `labels`.** `name` and `namespace` are removed from `reachableBackends.refs[]` and from the `kuma.io/reachable-backends` annotation. Every ref requires `kind` and `labels`; `port` stays optional. New writes using `name` or `namespace` fail validation with `labels: must not be empty`, and on Kubernetes the pod converter rejects such an annotation. An existing `Dataplane` is not re-validated: the control plane drops the unknown fields when it reads it, the ref resolves to nothing, and the proxy receives no outbound clusters, even with `KUMA_DEFAULTS_RESTRICT_OUTBOUND=false`, until every ref is rewritten.

**Action required**

Before upgrading, rewrite every `reachableServices` entry and every `reachableBackends` ref. `name` becomes the `kuma.io/display-name` label and `namespace` becomes the `k8s.kuma.io/namespace` label. Remove `experimental.autoReachableServices` from your control plane configuration.

Before:

```yaml
kuma.io/reachable-backends: |
  refs:
  - kind: MeshService
    name: redis
    namespace: redis-system
    port: 6379
```

After:

```yaml
kuma.io/reachable-backends: |
  refs:
  - kind: MeshService
    labels:
      kuma.io/display-name: redis
      k8s.kuma.io/namespace: redis-system
    port: 6379
```

#### Outbound traffic is restricted by default

`defaults.restrictOutbound` (`KUMA_DEFAULTS_RESTRICT_OUTBOUND`) defaults to `true`, which changes two defaults for data plane proxies:

- **No outbounds without `reachableBackends`.** A proxy without `reachableBackends` gets no generated outbounds, so it cannot reach any service through the transparent proxy. This includes pods injected by a 2.14 control plane.
- **Passthrough is `None`.** A transparent proxy that no `MeshPassthrough` policy selects drops traffic to destinations outside the mesh instead of forwarding it to the original destination, as if a `MeshPassthrough` with `passthroughMode: None` targeted it. Proxies without a transparent proxy or with bound outbounds are not affected, and a policy with `passthroughMode: All` keeps passthrough on.

**Action required**

Before upgrading, either:

- define `reachableBackends` on every data plane proxy, and allow the external destinations your workloads use with `MeshExternalService` or a `MeshPassthrough` policy (`passthroughMode: Matched` with `appendMatch`, or `passthroughMode: All`), or
- set `defaults.restrictOutbound` to `false` to restore the previous allow-all behavior.

#### Legacy DNS VIP allocator removed

The control plane no longer computes or persists the `kuma-<mesh>-dns-vips` `Config` resource, the write-only output of a legacy background allocator, and on Kubernetes no longer reconciles the per-namespace `ConfigMap` that exposed it. Data plane DNS records come from the `MeshService`/`MeshExternalService`/`MeshMultiZoneService` VIPs, which are unaffected.

These settings are removed:

- `experimental.useTagFirstVirtualOutboundModel` (`KUMA_EXPERIMENTAL_USE_TAG_FIRST_VIRTUAL_OUTBOUND_MODEL`)
- `dnsServer.CIDR` (`KUMA_DNS_SERVER_CIDR`)
- `dnsServer.serviceVipEnabled` (`KUMA_DNS_SERVER_SERVICE_VIP_ENABLED`)
- `runtime.universal.vipRefreshInterval` (`KUMA_RUNTIME_UNIVERSAL_VIP_REFRESH_INTERVAL`)

**Action required**

Remove the settings above from your control plane configuration. Previously persisted `kuma-<mesh>-dns-vips` resources and their `ConfigMap` mirrors are no longer read and can be deleted.

#### Delta xDS is the only xDS protocol

The control plane delivers configuration to data plane proxies only with incremental (Delta) xDS. The state-of-the-world (SOTW) code path is removed, and the control plane no longer serves `StreamAggregatedResources`.

These settings are removed:

- Control plane: `experimental.deltaXds` (`KUMA_EXPERIMENTAL_DELTA_XDS`) and the Helm value `experimental.deltaXds`.
- Data plane: `dataplaneRuntime.envoyXdsTransportProtocolVariant` (`KUMA_DATAPLANE_RUNTIME_ENVOY_XDS_TRANSPORT_PROTOCOL_VARIANT`) and the `kuma.io/xds-transport-protocol-variant` pod annotation.

The protocol a proxy uses is fixed in its Envoy bootstrap at startup, so a proxy that started against an older control plane keeps using SOTW until it reconnects with a fresh bootstrap. Once the control plane is on 3.0.0, a proxy still trying to use SOTW cannot establish ADS and must be restarted.

**Action required**

For a zero-downtime upgrade:

1. Set `KUMA_EXPERIMENTAL_DELTA_XDS=true` on the 2.14 control plane.
2. Restart all `kuma-dp` instances (or roll the workloads on Kubernetes) so their Envoy bootstraps use Delta xDS.
3. Once every data plane proxy is connected with Delta xDS, upgrade the control plane to 3.0.0 and roll the data plane proxies again as part of the normal upgrade flow.

Remove the settings above from your control plane configuration, Helm values and pod annotations.

#### Transparent proxy configured only through the ConfigMap

The legacy annotation-based transparent proxy injection path is removed. The sidecar injector always builds the transparent proxy configuration from the ConfigMap in the `kuma-system` namespace (merged with pod annotations) and delivers it through the `traffic.kuma.io/transparent-proxy-config` annotation and mounted files. This was previously opt-in with `transparentProxy.configMap.enabled`, which is removed. The per-pod `kuma.io/transparent-proxying-*` annotations are no longer produced by injection.

**Action required**

None for Helm or `kumactl` installs: the control plane always creates the base ConfigMap and points the injector at it. Remove `transparentProxy.configMap.enabled` from custom values files. Pods are reconfigured on their next restart after the upgrade.

**Warning**: this format is not understood by data plane proxies older than the control plane. To downgrade, first roll back the control plane and then restart all workloads so their init and sidecar containers fall back to the previous configuration.

#### eBPF transparent proxy removed

The experimental eBPF transparent proxy, which used prebuilt merbridge binaries for traffic interception instead of iptables, is removed.

This configuration no longer has any effect and is silently ignored:

- Control plane: `runtime.kubernetes.injector.ebpf.*` (`KUMA_RUNTIME_KUBERNETES_INJECTOR_EBPF_*`).
- Helm values: `experimental.ebpf.*`, `hooks.ebpfCleanup`, `cni.experimental.imageEbpf`, and `transparentProxy.configMap.config.ebpf`.
- Pod annotations: `kuma.io/transparent-proxying-ebpf*`.

This CLI surface is removed:

- `kumactl install transparent-proxy --ebpf-*` and `kumactl uninstall transparent-proxy --ebpf-*`.
- `kumactl uninstall ebpf`.

**Action required**

If you use the eBPF transparent proxy (`experimental.ebpf.enabled=true`), migrate to the iptables-based transparent proxy, the default, before upgrading, and remove the eBPF configuration from your Helm values and pod annotations.

If you have eBPF programs pinned on cluster nodes, clean them up **before** upgrading by running `kumactl uninstall ebpf` with the pre-upgrade Kuma version. Otherwise residual BPF state stays on affected nodes (pinned programs under `/sys/fs/bpf`, cgroup connect hooks, and TC filter attachments) and keeps intercepting traffic until removed manually. Consult the merbridge documentation for manual cleanup.

#### CoreDNS removed from the data plane

The bundled CoreDNS binary is removed. The data plane always uses the in-process embedded DNS proxy (previously the default on Kubernetes and opt-in on Universal). CoreDNS and the Envoy DNS filter are no longer used, and `coredns` is no longer shipped in the release tarball or the `kuma-dp` image.

This configuration is removed:

- Data plane (`kuma-dp`): `dns.coreDnsPort`, `dns.envoyDnsPort`, `dns.coreDnsBinaryPath`, `dns.coreDnsConfigTemplatePath`, `dns.configDir`, `dns.prometheusPort` and `dns.coreDNSLogging`, the matching `KUMA_DNS_CORE_DNS_PORT`, `KUMA_DNS_ENVOY_DNS_PORT`, `KUMA_DNS_CORE_DNS_BINARY_PATH`, `KUMA_DNS_CORE_DNS_CONFIG_TEMPLATE_PATH`, `KUMA_DNS_CONFIG_DIR`, `KUMA_DNS_PROMETHEUS_PORT` and `KUMA_DNS_ENABLE_LOGGING` env vars, and the `--dns-envoy-port`, `--dns-coredns-port`, `--dns-coredns-path`, `--dns-coredns-config-template-path`, `--dns-server-config-dir`, `--dns-prometheus-port` and `--dns-enable-logging` flags. The DNS listen port is configured solely with `dns.proxyPort` (`KUMA_DNS_PROXY_PORT`, `--dns-proxy-port`), defaulting to `15053`.
- Control plane: `bootstrapServer.params.corefileTemplatePath` (`KUMA_BOOTSTRAP_SERVER_PARAMS_COREFILE_TEMPLATE_PATH`), `runtime.kubernetes.injector.builtinDNS.experimentalProxy` (`KUMA_RUNTIME_KUBERNETES_INJECTOR_BUILTIN_DNS_EXPERIMENTAL_PROXY`) and `runtime.kubernetes.injector.builtinDNS.logging` (`KUMA_RUNTIME_KUBERNETES_INJECTOR_BUILTIN_DNS_LOGGING`).
- Helm: `dataPlane.dnsLogging`.
- The `kuma.io/builtin-dns-logging` pod annotation.

**Action required**

The control plane no longer emits the Envoy DNS filter configuration that older CoreDNS-based data planes rely on, so upgrade the control plane and all data planes together. Remove the settings listed above from your control plane configuration, Helm values and `kuma-dp` invocations.

#### Strict inbound ports and `SO_REUSEPORT` can no longer be turned off

`kuma-dp` no longer reads `KUMA_DATAPLANE_RUNTIME_STRICT_INBOUND_PORTS_ENABLED` or `KUMA_DATAPLANE_RUNTIME_REUSE_PORT_ENABLED`. Both defaulted to `true`, and the control plane now applies that behavior to every data plane:

- A sidecar with transparent proxy and a workload identity accepts inbound traffic only on the ports of its inbounds, unless a `MeshTLS` policy sets `Permissive` mode for it. Sidecars without a workload identity keep accepting inbound traffic on every port.
- Every inbound Envoy listener sets `enable_reuse_port: true`.

**Action required**

- If a workload relies on `KUMA_DATAPLANE_RUNTIME_STRICT_INBOUND_PORTS_ENABLED=false` to receive traffic on ports it does not declare, declare those ports as inbounds before you upgrade.
- Restart data planes that run with `KUMA_DATAPLANE_RUNTIME_REUSE_PORT_ENABLED=false` after you upgrade the control plane. Envoy cannot change `enable_reuse_port` on a running listener, so it rejects listener updates until the data plane restarts.

#### Sidecar probes always use the readiness port, and virtual probes are removed

**Readiness port.** The injected `kuma-sidecar` container had its liveness, readiness and startup probes pointed at the Envoy admin port (`9901`) when `experimental.envoyAdminUnixSocket` was off, and at the `kuma-dp` readiness port (`9902`) when it was on. Those two settings are decided in different places (the probe port at admission by the injecting webhook, the admin transport at bootstrap by whichever control plane instance answers), so a rolling control plane upgrade, or a change to the flag, left a window where they disagreed and pods went into `CrashLoopBackOff`. Probes now always use the readiness port, which is always excluded from inbound transparent proxy interception. `/ready` on the readiness port is proxied to the Envoy admin `/ready` on both transports, so a Pod is marked ready only once Envoy has its configuration.

Pods injected before the upgrade keep their probes on `9901` and continue to work: the `kuma:envoy:admin` listener still serves `/ready` on that port. They move to the readiness port the next time they are recreated.

The sidecar now receives `KUMA_READINESS_PORT` from `bootstrapServer.params.readinessPort`, so `kuma-dp` listens on the port the probes use. For the same reason `bootstrapServer.params.readinessPort` no longer accepts `0`; a control plane configured that way fails to start.

**Virtual probes removed.** Virtual probes, deprecated since 2.9 in favor of the application probe proxy, are removed:

- Pod annotations `kuma.io/virtual-probes` and `kuma.io/virtual-probes-port`
- Control plane configuration `runtime.kubernetes.injector.virtualProbesEnabled` and `runtime.kubernetes.injector.virtualProbesPort` (`KUMA_RUNTIME_KUBERNETES_VIRTUAL_PROBES_ENABLED`, `KUMA_RUNTIME_KUBERNETES_VIRTUAL_PROBES_PORT`)
- The `probes` field on `Dataplane` resources. `Dataplane` drops unknown fields silently, so `probes:` on a Universal `Dataplane` is accepted and discarded.

The application probe proxy (`kuma.io/application-probe-proxy-port`) is the only probe-rewriting path on Kubernetes and is enabled by default. Setting `kuma.io/application-probe-proxy-port: "0"` no longer falls back to virtual probes; it leaves the pod's original probes untouched.

**Action required**

- Update a `NetworkPolicy`, monitoring check, or `ContainerPatch` that pins the sidecar probe port to `9901` to `9902` (or to `bootstrapServer.params.readinessPort` if you changed it).
- Remove `bootstrapServer.params.readinessPort: 0`, and any `KUMA_READINESS_PORT` you set on the sidecar by hand to a value other than the control plane setting.
- Remove the virtual probes settings and annotations, and re-inject every pod after upgrading: a pod injected by a pre-3.0.0 webhook keeps stale `kuma.io/virtual-probes*` annotations and a probes-listener sidecar configuration until it is redeployed.
- Pods that relied on `kuma.io/application-probe-proxy-port: "0"` falling back to virtual probes must either remove the annotation to use the application probe proxy, or exclude their probe ports from inbound traffic redirection if they need probes served without mTLS.

#### DPP configuration refresh interval default raised to 10s

`xdsServer.dataplaneConfigurationRefreshInterval` (`KUMA_XDS_SERVER_DATAPLANE_CONFIGURATION_REFRESH_INTERVAL`) now defaults to `10s` instead of `1s`. The control plane regenerates the xDS configuration of every connected proxy on this interval, so a 1s default kept the control plane busy and scaled poorly with the number of data plane proxies.

**Action required**

None. Changes to meshes, policies, and services take up to 10 seconds to reach data plane proxies instead of up to 1 second. The same applies to trust bundles, so a CA rotation must leave the old CA in place for at least one refresh interval after the new one is added, otherwise proxies that have not yet been refreshed fail mTLS. If you need faster propagation, set the interval back to the previous value, keeping in mind the control plane CPU cost.

### Multizone, zone proxies and external services

#### `MeshExternalService` requires a `MeshIdentity`, and cross-zone traffic needs both zones upgraded

A client proxy without a workload identity no longer gets a cluster for a `MeshExternalService`. It previously got one that addressed the zone egress by the legacy `zone-egress` service SNI, but such traffic could never be served: a zone egress only generates its egress listener when it has a workload identity, and that listener matches filter chains on the KRI SNI only.

The same SNI change makes the upgrade one-way across zones: once a zone is upgraded, its client proxies send real-resource traffic to zone proxies by the KRI-derived SNI. A zone still on the previous version cannot terminate that traffic, because its zone proxy listeners match the legacy hashed SNI only. An upgraded zone cannot send `MeshService`, `MeshExternalService`, or `MeshMultiZoneService` traffic to a previous-version zone until that destination zone is upgraded too.

**Action required**

Create a `MeshIdentity` that matches the client proxies in any mesh that uses `MeshExternalService`. Without one, only the cluster is dropped: the outbound listener and its endpoints are still generated, so requests fail locally in the client's Envoy with a 503 (`cluster_not_found`) instead of being reset by the zone egress. Keeping the listener is deliberate, removing it would let the request fall through to the transparent proxy passthrough and reach the external service directly, bypassing the egress.

Plan zone upgrades so that cross-zone traffic does not flow from an upgraded zone to one that is not upgraded yet.

#### `MeshExternalService` TLS verification uses the `SecureDataSource` shape

`spec.tls.verification.caCert`, `.clientCert` and `.clientKey` on `MeshExternalService` now use the same `SecureDataSource` type as `MeshIdentity`, instead of the old `DataSource` type. The old type has been removed from the API entirely.

The old type had no discriminator: it was a flat object with `secret`, `inline` or `inlineString`. The new type requires a `type` discriminator and nests the value under a field matching it. Every old field has to be rewritten:

| Old field | New field |
|---|---|
| `inline: <base64>` | `type: InsecureInline`, `insecureInline.value: <plain text>` |
| `inlineString: <text>` | `type: InsecureInline`, `insecureInline.value: <text>` |
| `secret: <name>` | `type: Secret`, `secretRef: {kind: Secret, name: <name>}` |

`inline` was base64-encoded, `insecureInline.value` is plain text, so decode the old value when rewriting it. For example `inline: dGVzdA==` becomes:

```yaml
caCert:
  type: InsecureInline
  insecureInline:
    value: test
```

`File` and `EnvVar`, the two other `SecureDataSource` types, are rejected on `MeshExternalService`, because they read the control plane's own filesystem and environment. This also applies when `spec.extension` is set, even though an extension owns the rest of the `spec.tls` validation.

**Action required**

Rewrite the resources on 2.14 before upgrading the global control plane. 2.14.6 and later accept both shapes (see [Upgrade to `2.14.6`](#upgrade-to-2146)):

1. Upgrade the global control plane and every zone control plane to 2.14.6 or later.
2. Rewrite `caCert`, `clientCert` and `clientKey` on every `MeshExternalService` to the new shape.
3. Upgrade to 3.0, global control plane first.

**Warning**: a `MeshExternalService` written in the old shape after the upgrade is rejected at write time, because the missing `type` discriminator is a validation violation. Resources already stored in the old shape are not rejected: the control plane cannot read their TLS material, so the destination is dropped from the xDS config of every proxy routing to it, with an error logged on the control plane. Rewriting after the global is on 3.0 is too late: it accepts only the new shape, and zones older than 2.14.6 drop the new fields, so their proxies silently fall back to the system CA.

#### Per-zone `MeshExternalService` routing removed

A `MeshExternalService` labeled with `kuma.io/zone` is no longer reached only through that zone's ingress and egress. It is reachable directly through the local zone egress from every zone, the same as an unlabeled `MeshExternalService`.

**Action required**

None for typical usage; `kuma.io/zone` labels on `MeshExternalService` resources no longer gate reachability and can be removed. If you relied on the label to force traffic through a specific zone's egress (for example because only that zone has network access to the external endpoint), make sure each zone's network path to the endpoint is in place before upgrading.

#### Legacy `ExternalService` resource removed

The legacy `ExternalService` resource is removed, together with its CRD, API, validating webhook and RBAC rules.

**Action required**

Migrate any remaining `ExternalService` resources to `MeshExternalService` before upgrading. Applying an `ExternalService` after the upgrade fails because the CRD is gone.

#### Standalone `ZoneIngress` and `ZoneEgress` are not supported

Cross-zone traffic is served only by mesh-scoped zone proxies: ordinary `Dataplane` resources that declare zone proxy listeners (`networking.listeners` of type `ZoneIngress`/`ZoneEgress`), with health reported through `DataplaneInsight`. Everything that existed only for standalone zone proxies is removed.

**Resources and API**

- The `ZoneIngress`, `ZoneEgress`, `ZoneIngressInsight` and `ZoneEgressInsight` resources are gone, together with their CRDs (`zoneingresses.kuma.io`, `zoneegresses.kuma.io`, `zoneingressinsights.kuma.io`, `zoneegressinsights.kuma.io`), REST endpoints, KDS sync, validating webhook and RBAC rules.
- `GET /zoneingresses/{name}/{xds,stats,clusters}`, `GET /zoneegresses/{name}/{xds,stats,clusters}` and the pre-2.6 overview aliases `GET /zoneingresses+insights[/{name}]` and `GET /zoneegressoverviews[/{name}]` return `404`.
- `kumactl get`, `kumactl inspect` and `kumactl delete` lose `zoneingress`, `zoneingresses` (alias `zone-ingresses`), `zoneegress` and `zoneegresses`.
- The GUI's zone ingress and zone egress XDS, stats and clusters tabs depend on the removed endpoints and stop working until the bundled GUI is updated.
- `zones.zoneIngresses` and `zones.zoneEgresses` are removed from `/global-insight`, see [`MeshInsight` and global insight fields removed](#meshinsight-and-global-insight-fields-removed).

**Data plane proxies**

- `dataplane` is the only accepted value of `kuma-dp run --proxy-type` (and `KUMA_DATAPLANE_PROXY_TYPE`); `kuma-dp` exits with `.ProxyType "ingress" is not supported` on startup.
- The control plane no longer generates a bootstrap for `ingress`/`egress` proxy types, and rejects the xDS stream of a pre-upgrade standalone zone proxy that reconnects with `unsupported proxy type "ingress"`.
- `kumactl generate dataplane-token --proxy-type ingress|egress` is rejected, and tokens previously issued with a `type: ingress|egress` claim can no longer be used.

**Authentication**

Zone proxies authenticate like any other data plane proxy, with the method configured under `dpServer.authn.dpProxy` (`serviceAccountToken` on Kubernetes, `dpToken` on Universal). The separate zone proxy authenticator is gone, and zone tokens are no longer accepted by the DP server, see [Zone tokens only authenticate zone control planes](#zone-tokens-only-authenticate-zone-control-planes).

**Kubernetes**

- The Helm chart no longer renders standalone zone proxy Deployments. The top-level `ingress` and `egress` value blocks are gone, along with the Deployment, Service, HorizontalPodAutoscaler, PodDisruptionBudget and RBAC templates they drove. `controlPlane.ingress.*` is unrelated and still configures the Kubernetes Ingress for the control plane GUI and API.
- `kumactl install control-plane` lost `--ingress-enabled`, `--ingress-drain-time`, `--ingress-use-node-port`, `--ingress-node-selector`, `--egress-enabled`, `--egress-drain-time`, `--egress-service-type` and `--egress-node-selector`.
- The Pod controller no longer turns a Pod annotated with `kuma.io/ingress` or `kuma.io/egress` into a standalone zone proxy. `kuma.io/ingress-public-address` and `kuma.io/ingress-public-port` are removed with them; the public address comes from `MeshZoneAddress`. Nothing rejects the annotations: a Pod that keeps them is reconciled as a regular `Dataplane` if it has a sidecar, and ignored otherwise.

**Configuration**

| Removed | Note |
|---|---|
| `experimental.ingressTagFilters` (`KUMA_EXPERIMENTAL_INGRESS_TAG_FILTERS`) | filtered `ZoneIngress.availableServices`, which no longer exists |
| `multizone.zone.ingressUpdateInterval` (`KUMA_MULTIZONE_ZONE_INGRESS_UPDATE_INTERVAL`) | read by nothing |
| `runtime.universal.zoneResourceCleanupAge` (`KUMA_RUNTIME_UNIVERSAL_ZONE_RESOURCE_CLEANUP_AGE`) | offline zone proxies are collected by the Dataplane GC after `runtime.universal.dataplaneCleanupAge` (default 72h) |

The `component_zone_gc` metric is no longer exported; `component_dp_gc` covers zone proxies as well. Leftover values of the removed settings are silently ignored.

**Action required**

Migrate to mesh-scoped zone proxies **before** upgrading the zone control plane. Upgrading the control plane first blackholes cross-zone traffic as soon as the legacy zone proxies restart, and on Kubernetes a `helm upgrade` deletes the legacy zone proxy Deployment, Service, HorizontalPodAutoscaler, PodDisruptionBudget and RBAC objects. Helm does not reject unknown values, so an upgrade with `ingress.enabled=true` left in your values file succeeds without any warning.

1. On Kubernetes, deploy zone proxies through `meshes[].ingress.enabled` / `meshes[].egress.enabled`, or label the `Service` of a hand-rolled zone proxy with `k8s.kuma.io/zone-proxy-type`. On Universal, replace `kuma-dp run --proxy-type=ingress|egress` with a regular `Dataplane` that declares `networking.listeners` of type `ZoneIngress`/`ZoneEgress`, and run it with a dataplane token (`kumactl generate dataplane-token --mesh <mesh> --name <zone-proxy-dp>`, passed with `--dataplane-token-file`).
2. Drop the top-level `ingress` and `egress` blocks from your values files and the removed flags from `kumactl install control-plane`. Most legacy settings map onto `meshes[].ingress` / `meshes[].egress`; `podAnnotations` moves to `deployment.podAnnotations`, next to the new `deployment.podLabels`. These have no equivalent there: `annotations`, `logLevel`, `drainTime`, `lifecycle`, `livenessProbe`, `readinessProbe`, `startupProbe`, `dns.policy`, `dns.config`, `service.enabled` and `service.nodePort`. Drain time and probes are control-plane-wide sidecar injector settings.
3. If you relied on `dpServer.authn.zoneProxy.type: none` to let zone proxies connect without a token while data plane proxies used `dpToken`, zone proxies now need a dataplane token too.
4. Delete any remaining `ZoneIngress`/`ZoneEgress` resources before upgrading. On Universal, remove leftover rows from the resource store directly. On Kubernetes, delete the CRDs after upgrading, since Helm does not remove CRDs: `kubectl delete crd zoneingresses.kuma.io zoneegresses.kuma.io zoneingressinsights.kuma.io zoneegressinsights.kuma.io`.
5. Remove the settings listed above. If you relied on a zone-specific cleanup age, set `runtime.universal.dataplaneCleanupAge`, which now applies to every `Dataplane`.
6. Inspect zone proxies with `kumactl inspect dataplane <name> --mesh <mesh>` and `/meshes/{mesh}/dataplanes/{name}/{xds,stats,clusters}`, and update dashboards and alerts that use the removed endpoints or `component_zone_gc`.
7. Create a `MeshZoneAddress` right after the zone control plane upgrade, see the next section.

#### Cross-zone `MeshService` routing requires `MeshZoneAddress`

The publicly reachable address of a remote zone proxy is now taken solely from the `MeshZoneAddress` resource. The control plane no longer falls back to the `advertisedAddress`/`advertisedPort` of a `ZoneIngress` when building endpoints for `MeshService` and `MeshMultiZoneService` destinations in other zones.

On Kubernetes, the control plane creates `MeshZoneAddress` automatically for every mesh-scoped zone proxy `Service` (`meshes[].ingress.enabled=true`), which carries the `k8s.kuma.io/zone-proxy-type: ingress` label. On Universal it is authored by the user.

`MeshZoneAddress` is mesh-scoped: every mesh whose services are consumed from another zone needs its own resource in the zone that serves them.

**Action required**

`MeshZoneAddress` does not exist before 3.0.0, so it cannot be created ahead of the upgrade. Create it immediately after upgrading the zone control plane: cross-zone `MeshService` and `MeshMultiZoneService` traffic to that zone is down until the resource exists and has propagated over KDS.

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

#### Mesh-scoped zone proxies keep their `kuma-dp` until you restart them

Zone proxy pods deployed through `meshes[]` no longer carry the `helm.sh/chart` and `app.kubernetes.io/version` labels, so upgrading the control plane leaves them running, like any other data plane proxy. Previously every version upgrade rolled them while the old control plane was still injecting sidecars, so they restarted and came back with the old `kuma-dp` anyway.

Upgrading from a 2.14 patch that predates this change still rolls them once, because their pod template changes: the version labels go away and the egress `preStopSleepSeconds` default rises from 15 to 20 seconds. They come back with the 2.14 `kuma-dp`.

**Action required**

After the control plane upgrade, restart the zone proxies to move them to the new `kuma-dp`. Replace `kuma-system` if you installed the control plane in another namespace:

```sh
kubectl rollout restart deployment -n kuma-system -l kuma.io/mesh
```

#### KDS always uses the event-based watchdog, with periodic full resyncs

KDS always uses the event-based snapshot watchdog. The polling fallback and its configuration are removed:

- `experimental.kdsEventBasedWatchdog.enabled` (`KUMA_EXPERIMENTAL_KDS_EVENT_BASED_WATCHDOG_ENABLED`)
- `multizone.global.kds.refreshInterval` (`KUMA_MULTIZONE_GLOBAL_KDS_REFRESH_INTERVAL`)
- `multizone.zone.kds.refreshInterval` (`KUMA_MULTIZONE_ZONE_KDS_REFRESH_INTERVAL`)

The timing is configured under `multizone.global.kds.eventBasedWatchdog` and `multizone.zone.kds.eventBasedWatchdog` (`KUMA_MULTIZONE_{GLOBAL,ZONE}_KDS_EVENT_BASED_WATCHDOG_*`), with these defaults:

- `flushInterval`: `5s`
- `fullResyncInterval`: `1m`
- `delayFullResync`: `true`

Changes reach zones on the event path, coalesced over `flushInterval`. A full resync only recovers events that may have been missed, so a change missed on the event path is repaired within `fullResyncInterval`.

**Action required**

Remove the settings above from your control plane configuration. Set the `eventBasedWatchdog` intervals explicitly if you depend on different timing.

### Gateways

#### Built-in gateway removed

The built-in gateway is removed entirely:

- **Resources.** `MeshGateway`, `MeshGatewayRoute`, `MeshGatewayInstance` and `MeshGatewayConfig`, their CRDs (`meshgateways.kuma.io`, `meshgatewayroutes.kuma.io`, `meshgatewayinstances.kuma.io`, `meshgatewayconfigs.kuma.io`) and their KDS sync no longer exist. `MeshGateway` is no longer a valid `targetRef.kind` for any policy.
- **Kubernetes controllers.** The control plane no longer reconciles `MeshGatewayInstance` or creates the `Service` and `Deployment` for one, no longer converts `Pod`s annotated `kuma.io/gateway: builtin` into a gateway `Dataplane`, and `kumactl inspect meshgateway` is removed.
- **North-south Gateway API.** The control plane no longer reconciles `Gateway`, `GatewayClass` and the `ReferenceGrant`s used to configure the built-in gateway. `HTTPRoute`s whose `parentRefs` target a `Gateway` are ignored; status Kuma already wrote for those parents is left as is. `HTTPRoute`s that target a `Service` (GAMMA) are still translated to `MeshHTTPRoute`. The Helm chart no longer installs the `kuma` `GatewayClass`; on startup the control plane removes the legacy `gateway-exists-finalizer.gateway.networking.k8s.io` finalizer from Kuma-managed `GatewayClass` objects so Helm-pruned objects can finish deleting.
- **Gateway secrets.** `runtime.kubernetes.supportGatewaySecretsInAllNamespaces` (`KUMA_RUNTIME_KUBERNETES_SUPPORT_GATEWAY_SECRETS_IN_ALL_NAMESPACES`) and the Helm value `controlPlane.supportGatewaySecretsInAllNamespaces` are removed. The control plane always scopes its `Secret` watch to its own system namespace. Kuma `Secret`s copied from Gateway API TLS `Secret`s (name prefixed `gapi-`) are orphaned.
- **Legacy gateway policies.** Nothing falls back to `MeshGatewayRoute`, `FaultInjection` or `HealthCheck` anymore, see [Legacy policies removed](#legacy-policies-removed).
- **API.** `MeshInsight.dataplanesByType.gatewayBuiltin`, the `gatewayBuiltin` objects in `/global-insight` and the `gateway=builtin` / `type=gateway_builtin` filters are removed, see [`MeshInsight` and global insight fields removed](#meshinsight-and-global-insight-fields-removed).
- **`BUILTIN` `Dataplane`s.** `Dataplane.networking.gateway` is removed (see [`kuma.io/gateway` is removed](#kumaiogateway-is-removed)), so a stored `Dataplane` with `type: BUILTIN` keeps loading as an ordinary `Dataplane` with no inbounds instead of failing validation.

**Action required**

1. Before upgrading, migrate every built-in gateway to a gateway you run yourself: your own `Deployment`/`Service` fronting a Kuma-injected pod that excludes its listen ports with `traffic.kuma.io/exclude-inbound-ports`, see [`kuma.io/gateway` is removed](#kumaiogateway-is-removed). Migrate gateway routing to `MeshHTTPRoute`/`MeshTCPRoute` and gateway fault injection and health checking to `MeshFaultInjection` and `MeshHealthCheck`.
2. Before upgrading, delete the remaining `MeshGateway`, `MeshGatewayRoute`, `MeshGatewayInstance` and `MeshGatewayConfig` resources. Deleting a `MeshGatewayInstance` cascades to the `Service`, `Deployment` and `Dataplane` it generated through owner references. On Kubernetes, the Helm upgrade deletes the CRDs, which removes any resources still stored under them, so back them up first if you need them.
3. Before upgrading, delete the `Gateway` and `ReferenceGrant` resources you created for the built-in gateway; the control plane leaves them untouched.
4. After upgrading, find `Dataplane`s that still carry `type: BUILTIN` (`kumactl get dataplanes -o yaml` per mesh or `kubectl get dataplanes -A -o yaml`) and delete the ones you no longer serve traffic with. Delete orphaned `gapi-` `Secret`s.
5. Remove `supportGatewaySecretsInAllNamespaces` from your configuration and Helm values.
6. If you manage RBAC yourself, see [Control plane RBAC changes on Kubernetes](#control-plane-rbac-changes-on-kubernetes), including the temporary `gatewayclasses` permissions the finalizer cleanup needs.

Rolling back past this version does not restore the removed controller behavior automatically. A Helm rollback alone does not restore RBAC or webhook rules removed by the upgrade; reapply the previous chart version to restore them.

#### `kuma.io/gateway` is removed

The gateway marking is gone: the `kuma.io/gateway` Pod annotation, the `kuma.io/gateway` Dataplane label, and `Dataplane.networking.gateway` (field number reserved, `type` and `tags` with it) no longer exist. A gateway is an ordinary `Dataplane` that keeps its listen ports out of inbound redirection with the `traffic.kuma.io/exclude-inbound-ports` annotation, which is where the behaviour that mattered, Kuma not proxying the traffic the gateway terminates, actually came from.

Everything that used to branch on the marking is gone with it: policy matching, the validation that forbade inbounds and listeners, `DataplaneOverview` status, the `MeshMetric` `gateway` proxy role, the mesh and global insight gateway counters, and the `?gateway=` filter.

The annotation is now ignored rather than rejected, so a Pod that still carries it is admitted and keeps running. It stops meaning anything, though: the pod is injected like any other workload, gets inbounds generated from its `Service`, and has its inbound traffic redirected through Envoy.

**Action required**

On Kubernetes, replace the annotation with the ports the gateway listens on, before or together with the upgrade:

```yaml
metadata:
  annotations:
    traffic.kuma.io/exclude-inbound-ports: "8000,8443"   # replaces kuma.io/gateway
```

List every port the gateway accepts traffic on. The annotation takes a comma-separated list and has no all-ports spelling. A port left off it is redirected into Envoy and served by an inbound listener, so the traffic the gateway terminates becomes mesh inbound traffic, subject to `MeshTrafficPermission` and mTLS; a client outside the mesh is then rejected rather than reaching the gateway. Restart the pods so the injector rewrites the transparent proxy configuration.

An excluded port is not a `Dataplane` inbound. Envoy does not receive traffic on it, so it is not addressable through the mesh, and the workload accepts connections on it directly. This is also true for a workload that is not a gateway: excluding one of its `Service` target ports takes that port out of the mesh.

If the gateway is fronted by a `Service`, annotate that `Service` too:

```yaml
metadata:
  annotations:
    kuma.io/ignore: "true"
```

The gateway marking used to suppress both the `Dataplane` inbounds and the `MeshService` generated from the gateway's `Service`; `kuma.io/ignore` is what does that now. Without it the `Service` produces a `MeshService`, and in-mesh clients that address the gateway through it get a mesh destination with no endpoints instead of the passthrough traffic they got before.

On Universal, drop `kuma.io/gateway: "true"` from the labels of gateway `Dataplane`s and run `kuma-dp` with `--exclude-inbound-ports` (or `redirect.inbound.excludePorts` in the transparent proxy config) covering the same ports.

Also move any key you use in a `MeshLoadBalancingStrategy` `localZone.affinityTags` from `networking.gateway.tags` to the `Dataplane`'s labels. An affinity key that exists only as a gateway tag stops matching and its locality group is dropped.

**What this changes elsewhere**

- A policy with a proxy-wide top-level `targetRef` (`kind: Mesh`, or `kind: Dataplane` without a `sectionName`) selects a proxy that declares no inbounds. Gateways used to need the marking for this; now nothing does.
- A `Dataplane` may declare both inbounds and zone proxy listeners without the validator rejecting it. The messages `inbound cannot be defined for delegated gateways` and `listeners cannot be defined for delegated gateways` are gone.
- `DataplaneOverview` status for a proxy with no inbounds and no listeners is `Online` while it is connected, which is what gateways reported before. Proxies that do declare inbounds are unaffected.
- `MeshMetric` no longer emits `gateway` for `kuma.proxy_role`; a former gateway reports `sidecar`. Update dashboards and alerts that select on it.
- The gateway counters in `MeshInsight` and `/global-insight` and the `?gateway=` filter are removed, see [`MeshInsight` and global insight fields removed](#meshinsight-and-global-insight-fields-removed). `kumactl inspect dataplanes --gateway` is removed.
- A `Service` or `Pod` that was skipped for carrying the annotation now gets a `MeshService` like any other workload, unless the `Service` carries `kuma.io/ignore: "true"`.
- The injected sidecar of a former gateway pod gets the regular application probe proxy port instead of `0`, so its probes are proxied.
- The control plane no longer writes `kuma.io/zone` into a gateway's tags, and KDS no longer rewrites a zone tag inside a synced `Dataplane` spec. The `kds_zone_attribution_rewrites_total` metric is removed with it; zone attribution on labels is unaffected.
- A `kuma.io/gateway` or `kuma.io/proxy-type` label left on a stored `Dataplane` by an older control plane is no longer deleted on every write. On Kubernetes the control plane drops it on the next reconcile of the Pod; on Universal it goes away when `kuma-dp` registers the proxy again, and a `kuma-dp` `Dataplane` that still carries either label fails to register (see [Reserved and control-plane-owned labels are enforced](#reserved-and-control-plane-owned-labels-are-enforced)). A `targetRef` or `MeshLoadBalancingStrategy` affinity key that selects on them stops matching; move those to a label the control plane does not manage.

#### Gateway API `HTTPRoute` translation changes

These changes apply to east-west (GAMMA) `HTTPRoute`s that Kuma translates to `MeshHTTPRoute`.

**Cross-namespace `backendRefs` require a `ReferenceGrant`.** A backend reference that points across namespaces, to a core `Service` or a `MeshService`, is programmed into the generated `MeshHTTPRoute` only when a matching `ReferenceGrant` exists in the backend namespace. Otherwise the route reports `ResolvedRefs=False` with reason `RefNotPermitted`.

**Producer routes are created in the `HTTPRoute`'s namespace.** A `MeshHTTPRoute` generated from an `HTTPRoute` whose parent (a `Service` or `MeshService`) lives in the `HTTPRoute`'s own namespace is now created in that namespace instead of the Kuma system namespace. That gives it `kuma.io/policy-role: producer`, the same role as a hand-written `MeshHTTPRoute` targeting the same `Mesh` and `to` entry, so the two now tie where the hand-written route used to always win. Its `k8s.kuma.io/namespace` label changes to match, it syncs across zones like any other namespaced policy, and it applies to dataplanes in remote zones. A generated route whose parent is in a different namespace (a consumer route) still lands in the system namespace with its `system` role. The control plane moves each existing generated route the next time its `HTTPRoute` is reconciled.

**Conflicts on the same parent resolve by `creationTimestamp`.** When two `HTTPRoute`s attach to the same parent with equally specific matches, the generated `MeshHTTPRoute`s used to tie-break by name. Each now carries its owning `HTTPRoute`'s `creationTimestamp` as a label, which breaks the tie in favor of the older `HTTPRoute`; name order is used only when the timestamps also match. A hand-written `MeshHTTPRoute` has no such label, so it still takes precedence over any generated route it ties with on an exact conflict. Existing generated routes get the label the next time their `HTTPRoute` is reconciled.

**Action required**

- Create a `ReferenceGrant` in the backend namespace for every cross-namespace `HTTPRoute` backend that should remain valid after upgrading.
- If a mesh has both a generated producer route and a hand-written `MeshHTTPRoute` targeting the same `Mesh` and `to` entry, check which one you expect to win: after the upgrade the two tie and the outcome falls back to resource name.
- Update any policy that selects a generated route by `k8s.kuma.io/namespace: <system namespace>` to the `HTTPRoute`'s namespace.
- If two `HTTPRoute`s previously tied and were resolved by name, check whether the `creationTimestamp` outcome matches what you expect.
- If you manage RBAC yourself, see [Control plane RBAC changes on Kubernetes](#control-plane-rbac-changes-on-kubernetes).

### Observability

#### `MeshMetric` `ActiveMTLSBackend` requires a client certificate

The `ActiveMTLSBackend` TLS mode of a `MeshMetric` Prometheus backend now secures the endpoint with the workload identity of the dataplane. A scraper from a different address, also one on the same node, must send a client certificate that the mesh trusts. The certificate must contain a SPIFFE ID and chain to the CA bundle of that trust domain. With the bundled identity, the bundles come from `MeshTrust`. With SPIRE, they are all the bundles that the SPIRE agent serves, federated bundles included. Scrapes from the dataplane address or from a loopback address stay plaintext. A dataplane without a workload identity accepts only these local scrapes.

Before this change, the mode produced a plaintext endpoint.

**Action required**

Give each remote scraper a client certificate that the mesh trusts before the upgrade. To keep a plaintext endpoint, set the mode to `Disabled`.

MADS sends `https` for this mode. The hint does not supply a client certificate. 3.0 has no direct access mTLS, so a remote scraper must use `https` and its own client certificate, also in the mesh.

A scraper that connects from the dataplane address or from a loopback address must use `http` on the backend port and path, although MADS sends `https`. The endpoint accepts plaintext only from these addresses. This mode has no local TLS.

#### OpenTelemetry backends must be a `MeshOpenTelemetryBackend`

`MeshTrace`, `MeshAccessLog` and `MeshMetric` no longer accept an inline `openTelemetry.endpoint` (in `default.backends[]`, and for `MeshAccessLog` also `rules[].default.backends[]`). `openTelemetry.backendRef`, pointing at a `MeshOpenTelemetryBackend`, is the only way to configure an OpenTelemetry backend. Policies that still set `endpoint` fail validation.

Data for a `backendRef` always goes through `kuma-dp`: Envoy exports to a Unix socket and `kuma-dp` forwards to the collector. `runtime.kubernetes.injector.otelPipeEnabled` (`KUMA_RUNTIME_KUBERNETES_INJECTOR_OTEL_PIPE_ENABLED`) and `KUMA_DATAPLANE_RUNTIME_OTEL_PIPE_ENABLED`, which made Envoy export to the collector directly when set to `false`, are removed and ignored.

**Action required**

Before upgrading, for every `MeshTrace`, `MeshAccessLog` and `MeshMetric` that sets `openTelemetry.endpoint`, create a `MeshOpenTelemetryBackend` with the equivalent endpoint and reference it with `openTelemetry.backendRef`. Remove the two `otelPipeEnabled` settings from your control plane configuration and sidecar environment.

#### MADS restricted to universal deployment mode

The Monitoring Assignment Discovery Service (MADS) server no longer starts on Kubernetes control planes, regardless of `KUMA_MONITORING_ASSIGNMENT_SERVER_ENABLED` or `controlPlane.madsServer.enabled`. It remains supported in universal deployment mode, including universal-on-Kubernetes (`controlPlane.environment: universal`). The Helm chart no longer renders the `mads-server` Service port (5676) when `controlPlane.environment` is `kubernetes`.

**Action required**

Kubernetes users relying on MADS must migrate to `MeshMetric` with Prometheus Kubernetes service discovery before upgrading.

#### `kumactl install observability` removed

The deprecated `kumactl install observability` command is removed. Use separately managed observability components or your platform's observability stack. Kuma still ships first-party Grafana dashboards in the release tarball under `dashboards/grafana/`.

### Control plane configuration

#### Global control plane on Kubernetes is no longer supported

A Kubernetes-native Global control plane is no longer supported. `kuma-cp` rejects `mode=global` with `environment=kubernetes`, and also `mode=global` with `store.type=kubernetes`. A Global control plane must run with `environment=universal` backed by a non-Kubernetes store such as PostgreSQL, even if `kuma-cp` itself is deployed on Kubernetes. The Helm chart no longer renders the `Service`/config needed for the old Kubernetes-native setup. Zone control planes on Kubernetes are unaffected.

**Action required**

If you run the Global control plane on Kubernetes, migrate it to Universal before upgrading: deploy `kuma-cp` in `global` mode on Universal, backed by PostgreSQL, and keep your Kubernetes clusters as Zone control planes connecting to it over KDS.

#### `standalone` mode removed

The deprecated `standalone` control plane mode is removed. `KUMA_MODE`/`controlPlane.mode` no longer accepts `standalone`: `kuma-cp` fails config validation at startup, and the Helm chart fails at template time.

**Action required**

Rename `standalone` to `zone` in `KUMA_MODE`, `controlPlane.mode`, and any other runtime configuration before upgrading. `standalone` and `zone` were already behaviourally identical.

#### System namespace is labeled on every Helm upgrade

The Helm hook that labels the system namespace with `kuma.io/sidecar-injection: "false"` now runs on `pre-upgrade` as well as `pre-install`. Before, a release recovered with `helm upgrade` after a failed first install could leave the namespace unlabeled, so policy admission webhooks didn't select it. The hook's `ServiceAccount`, `ClusterRole` and `ClusterRoleBinding` are now also created during upgrades and deleted once the hook finishes.

**Action required**

None. Set `patchSystemNamespace: false` to keep managing the label yourself.

#### Control plane sets `GOMEMLIMIT` itself to 90% of the cgroup memory limit

The Helm chart used to set `GOMEMLIMIT` to the full `controlPlane.resources.limits.memory`. `GOMEMLIMIT` only covers memory the Go runtime manages, so with no headroom left for the rest, such as the mapped `kuma-cp` binary, the container could be OOM killed before the garbage collector reacted. `kuma-cp` now reads the cgroup memory limit at startup and sets `GOMEMLIMIT` to 90% of it. This also covers limits the chart does not know about, such as a `LimitRange` default or a VPA, and Universal deployments under a cgroup. The chart no longer renders the `GOMEMLIMIT` environment variable, and `controlPlane.runtime.goMemLimit.divisor` is removed.

**Action required**

None. Set the `AUTOMEMLIMIT` environment variable to choose another fraction, for example `AUTOMEMLIMIT=0.85`, or to `off` to disable it. Set `GOMEMLIMIT` explicitly to use an exact value, in which case `kuma-cp` leaves it as is. On Kubernetes use `controlPlane.envVars` for either. Remove `controlPlane.runtime.goMemLimit` from your values if you set it.

#### The resource store cache can no longer be disabled

The resource store cache (`store.cache.enabled`, `KUMA_STORE_CACHE_ENABLED`) was on by default, and disabling it degraded performance so much that the option is removed. The cache is always enabled, on every control plane instance. It is local to an instance and eventually consistent, so reads may briefly observe stale resources right after a write, which is also what happened with the default configuration before.

**Action required**

Remove `store.cache.enabled` / `KUMA_STORE_CACHE_ENABLED` from your configuration if present. The setting is ignored, and only `store.cache.expirationTime` (`KUMA_STORE_CACHE_EXPIRATION_TIME`) remains configurable.

#### Deprecated configuration settings removed

These settings are removed. The control plane and `kuma-dp` ignore leftover values in YAML and environment variables, and Helm accepts unknown values, so nothing fails on startup:

| Removed | Use instead |
|---|---|
| `KUMA_MESH_TRAFFIC_PERMISSION_DISABLE_CLIQUES_ALGORITHM` | none, `MeshTrafficPermission` rule generation always uses the cliques algorithm |
| `KUMA_DATAPLANE_RUNTIME_UNIFIED_RESOURCE_NAMING_ENABLED` (`kuma-dp`), `runtime.kubernetes.injector.unifiedResourceNamingEnabled` (`KUMA_RUNTIME_KUBERNETES_INJECTOR_UNIFIED_RESOURCE_NAMING_ENABLED`), Helm `dataPlane.features.unifiedResourceNaming` | none, unified Envoy resource and stat naming is always on |
| `kuma-dp` `dataplaneRuntime.configDir` / `socketDir` (`KUMA_DATAPLANE_RUNTIME_CONFIG_DIR`, `KUMA_DATAPLANE_RUNTIME_SOCKET_DIR`), `--config-dir` | `workDir` (`KUMA_DATAPLANE_RUNTIME_WORK_DIR`, `--work-dir`) |
| `runtime.kubernetes.injector.sidecarContainer.adminPort` (`KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_ADMIN_PORT`) | `bootstrapServer.params.adminPort` (`KUMA_BOOTSTRAP_SERVER_PARAMS_ADMIN_PORT`), which the injector already read |
| `metrics.mesh.minResyncTimeout` / `maxResyncTimeout` (`KUMA_METRICS_MESH_MIN_RESYNC_TIMEOUT`, `KUMA_METRICS_MESH_MAX_RESYNC_TIMEOUT`) | `metrics.mesh.minResyncInterval` (`KUMA_METRICS_MESH_MIN_RESYNC_INTERVAL`) and `metrics.mesh.fullResyncInterval` (`KUMA_METRICS_MESH_FULL_RESYNC_INTERVAL`) |

**Action required**

Remove these settings. Two of them change behavior:

- If you set a unified naming setting to `false`, Envoy resource and stat names change to the unified names; update automation, dashboards and alerts that depend on the legacy names before upgrading.
- `kuma-dp --config-dir` fails with `unknown flag`. A proxy that relied on `configDir`/`socketDir` in YAML or environment falls back to a generated temporary directory; set `workDir` instead.

## Upgrade to `2.14.6`

### `MeshExternalService` TLS verification accepts the `SecureDataSource` shape

2.14.6 accepts both shapes of `spec.tls.verification.caCert`, `.clientCert` and `.clientKey` on `MeshExternalService`, so resources can be rewritten to the shape 3.0 requires before upgrading. The old `secret`, `inline` and `inlineString` fields keep working on 2.14 and produce a deprecation warning. A single data source cannot mix both shapes, and the `File` and `EnvVar` types are rejected.

**Action required**

Before upgrading to 3.0, upgrade every control plane to 2.14.6 or later, then rewrite each `MeshExternalService` as described in [`MeshExternalService` TLS verification uses the `SecureDataSource` shape](#meshexternalservice-tls-verification-uses-the-securedatasource-shape). A zone on an older 2.14 does not know the new fields: they are dropped (pruned on Kubernetes), and the proxy silently falls back to the system CA.

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

### Unused control plane DNS configuration removed

The remaining `dnsServer` control plane configuration has been removed:

- `dnsServer.domain` (`KUMA_DNS_SERVER_DOMAIN`).
- `dnsServer.serviceVipPort` (`KUMA_DNS_SERVER_SERVICE_VIP_PORT`).

These settings are unused by the 3.0 service discovery implementation. Removing them does not change DNS records or service ports. Use `HostnameGenerator` templates to customize generated service hostnames.

**Action required**

Remove the `dnsServer` section from your control plane configuration and the environment variables above when upgrading to 3.0.

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
