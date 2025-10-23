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
