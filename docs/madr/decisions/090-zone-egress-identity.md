# Zone Egress Identity

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/14178

## Context and Problem Statement

Kuma leverages ZoneEgress as a single point of exit for outgoing requests to destinations outside the local zone network.
This pattern enhances security by allowing users to restrict egress traffic - enabling only a specific subset of addresses to communicate with external services. While this approach significantly improves overall security, it also requires that all outbound traffic from different meshes be handled by a single workload: ZoneEgress.

Communication between ZoneEgress and application sidecars is secured using mutual TLS (mTLS).
Currently, Kuma configures ZoneEgress with SNI-based matching for specific targets, where each filter chain is assigned a separate identity unique to each mesh. This approach introduces a problem — a single ZoneEgress instance ends up having multiple identities, which breaks the intended trust and identity model.

## User Stories

### As a user, I want to communicate with MeshExternalServices through ZoneEgress over a secure connection

As this is the current behavior that works correctly with the legacy mutual TLS defined on the Mesh object, we should continue to support it.

### As a user, I want to configure all traffic leaving the local cluster to route through ZoneEgress over a secure connection

When using the `kuma.io/service` tag, it is currently possible to route traffic through ZoneEgress, but this is not yet supported for MeshService.
We should keep in mind that this functionality might be required in the future, and ideally, it should be supported out of the box.

https://github.com/kumahq/kuma/issues/10728

### As a user, I want to integrate SPIRE with ZoneEgress, so that I can leverage SPIRE-issued identities

When using SPIRE, all identities and trust bundles are delivered by the SPIRE SDS server.
We need to ensure that communication to ZoneEgress from different meshes functions correctly under this setup.

### As a user, I want to enable traffic through ZoneEgress without any additional configuration

It should be just as easy to enable traffic through ZoneEgress as it is today.
The process should not require any additional configuration or manual setup beyond what is already needed in the current implementation.

## Out of scope

### As a user, I want to support SPIRE and other identity providers simultaneously when using ZoneEgress

Since integrating Kuma with SPIRE requires a federation API, we might need to postpone this feature for now.

## Options Considered

1. Use the `MeshIdentity` of the default `Mesh` to provide identity for the `ZoneEgress`.
2. `ZoneEgress` as a `Mesh`-scoped resource

### Use the `MeshIdentity` of the default `Mesh` to provide identity for the `ZoneEgress`.

In most cases, users deploy a single `Mesh`. Therefore, we can use the `MeshIdentity` resource from that Mesh to generate the `ZoneEgress` identity.
This approach would cover the majority of deployments.

For environments with multiple Meshes, we would introduce a configuration option that allows the user to explicitly specify which Mesh’s `MeshIdentity` should be used to provide the `ZoneEgress` identity.

#### SPIRE

> [!WARNING] 
> Currently, it’s not possible to use MeshTrust when SPIRE is enabled, since SPIRE itself is responsible for providing the validation context.

If any of the `MeshIdentity` resources use SPIRE, we should select that one to provide the identity for the `ZoneEgress`.
Because SPIRE manages the validation context, it is currently the only viable option for enabling secure communication with `ZoneEgress`.

#### Multiple Meshes, Each with a Different Identity Provider

If there is a default `Mesh` specified with a `MeshIdentity`, we will use it to provide the `ZoneEgress` identity.
If there is no default `Mesh`, the user must manually specify which `Mesh` should be used for `ZoneEgress` identity provisioning using the following configuration option:

```yaml
KUMA_DEFAULTS_MESH: "my-mesh" # used for MeshIdentity; any dataplane started without kuma.io/mesh joins this mesh
```

Based on this configuration, the `ZoneEgress` Envoy will receive an identity issued for its specific `Mesh`. Using `MeshTrust`, we configure the `ZoneEgress` to trust identities from all `Meshes`, while dataplanes are configured to trust the `ZoneEgress` identity.

##### How do we choose which MeshIdentity configures the ZoneEgress?

We choose the policy in lexicographical order among those that select all dataplanes, from the default `Mesh`.

##### SPIRE

When SPIRE is involved, the setup becomes more complex. There are two possible scenarios:

1. All `Meshes` use SPIRE

This case is simpler since all `Meshes` share a common identity provider. We don’t need to handle mixed configurations or custom trust setups, as SPIRE already manages the validation context consistently.

2. Some `Meshes` use SPIRE and others use a different `MeshIdentity` provider

This scenario is more complicated, as MeshTrust cannot be used together with a SPIRE-provided validation context.
To address this, we could first resolve [issue](https://github.com/kumahq/kuma/issues/14786).

This approach would allow SPIRE to interoperate with other `MeshIdentity` providers and enable `MeshTrust` support while SPIRE is in use.

#### What if there is no default Mesh?

In this case, the `kuma-cp` logs will contain a message indicating that the default Mesh is not defined, and since the system has multiple Meshes, the user must explicitly configure which one to use.
An alternative option would be to retrieve the list of all Meshes and select the first one in lexicographical order that has a `MeshIdentity` policy defined.

#### What if there is one Mesh using legacy mutual TLS and another using MeshIdentity?

If the default `Mesh` is not explicitly defined by the user to point to the `Mesh` with `MeshIdentity`, the control plane will use the `Mesh` that relies on mutual TLS.

#### What if a single Mesh is configured with both mutual TLS and MeshIdentity?

If the default `Mesh` is not explicitly defined by the user to point to the `Mesh` with `MeshIdentity`, the control plane will use the `Mesh` that relies on mutual TLS.

#### Validation Context

Dataplanes that communicate with `ZoneEgress` need to set a validation context referencing a secret that contains the CA used to validate the `Egress` identity.
Since we already have a secret named `system_trust_bundle` for validating traffic between services, we need to introduce a new secret specifically for Egress identity validation.

Proposed name
`system_trust_bundle_egress`

In the case of Egress, we need to provide one of the following:

* Aggregated MeshTrust from all Meshes: `system_trust_bundle`
* Aggregated MeshTrust from a specific Mesh: `system_trust_bundle_<mesh>`

#### Pros

* No need to add support for targeting ZoneEgress
* In most cases, it doesn’t require any additional configuration

#### Cons

* Not intuitive and difficult to understand
* Requires some manual user configuration
* It’s not obvious which identity provider to use when multiple meshes and MeshIdentity resources exist
* Might be tricky from the validation context configuration as it requires specific MeshTrust which might be independent from MeshIdentity

### ZoneEgress as a Mesh-scoped resource

Another option is to make the `ZoneEgress` resource Mesh-scoped. In this case, users could deploy a `ZoneEgress` for a specific `Mesh` by labeling it with the `kuma.io/mesh` label. This would be a completely new feature, as it would require a dedicated design and should be covered separately.

Making ZoneEgress Mesh-scoped could simplify several use cases, such as:
* Targeting `ZoneEgress` instances with policies
* Determining which identity should be applied to a `ZoneEgress`

#### Deployment of ZoneEgress

In this case we couldn't deploy a ZoneEgress as a separate resource since it's per mesh. On the kubernetes we could have a controller which based on the mesh configuration `zoneEgress` could deploy automatically mesh specific zone egress. On universal it would have to be handled by the user as it's now.

#### Pros

* Easy to understand which identity is used for `ZoneEgress`
* Policy targeting for `ZoneEgress` would be easier to implement and reason about

#### Cons

* More maintenance for the user, as it requires multiple `ZoneEgress` instances in a multi-mesh environment
* Requires a new design
* In a multi-mesh environment with access to `MeshExternalService`, resource utilization increases since we need as many `ZoneEgress` instances as there are `Meshes`
