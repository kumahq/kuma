# SPIFFE Compliance

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/13659

## Context and Problem Statement

In today’s service mesh systems, secure communication between services is essential. To make that possible, each service needs a clear and trusted identity: so we know exactly who is talking to whom.

One of the most trusted standards for defining these service identities is SPIFFE (Secure Production Identity Framework for Everyone). It gives us:

* A standard way to name services: `spiffe://<trust-domain>/<path>`
* A strict format for certificates, especially the Subject Alternative Name (SAN) - which must contain exactly one URI (the SPIFFE ID)
* Tools and rules for issuing, rotating, and checking these identities

Right now, Kuma does something similar by generating SPIFFE-like IDs like `spiffe://<mesh-name>/<service-name>`. But it also adds extra SANs (like `kuma://kuma.io/version/v1`) based on tags. These extra SANs make Kuma's certificates non-compliant with SPIFFE, meaning they can’t be used directly with SPIFFE-based systems like SPIRE, or other tools that expect full compliance.

In SPIFFE, a Trust Domain defines which services are allowed to trust each other. It’s like a boundary — services inside the same trust domain can safely communicate. To talk across trust domains (for example, between two companies or clusters), you need federation, which lets them trust each other's identities.

Kuma currently uses the mesh name as the trust domain. While this works inside Kuma, it doesn’t clearly separate trust boundaries, and it limits how easily Kuma can connect with external services or other applications in the same trust domain.

This setup causes problems when:

* You want to connect Kuma with other SPIFFE-compliant systems outside the mesh
* You run multiple clusters that need to trust each other
* You work in environments that require strict identity standards for security or compliance reasons

To fully support SPIFFE and modern security practices, Kuma should:

* Issue certificates that include only one SPIFFE-compliant URI SAN (currently we generate `spiffe://<mesh>/<service>` for each inbound and `kuma://kuma.io/<tag_name>/<tag_value>` for each tag)
* Let users set the trust domain explicitly, not tie it to the mesh name
* Allow using other systems (like SPIRE) to manage identities

## User Stories

### As a user, I want a Dataplane (e.g., Service A) inside the mesh to accept mTLS connections from a service (e.g., Service B) outside the mesh.

A user should be able to use certificates signed by the same Certificate Authority (CA) across different services. This capability simplifies migration paths and unlocks use cases such as enabling mTLS communication between services inside the mesh and components outside the mesh (e.g., a gateway without a sidecar proxy). It's also worth mentioning that the CA might be managed by another component e.g.: SPIRE, which can issue certificates that are SPIFFE compliant.

### As a user, I want to be able to rotate or change the Certificate Authority (CA) used for issuing mTLS identities without interrupting existing connections or causing service downtime.

During the entire process, communication between services must remain secure to ensure uninterrupted, encrypted traffic. This is essential for maintaining zero-trust security guarantees and ensuring operational reliability in production environments.

### As a user, it should not become harder to enable or manage mTLS in either single-zone or multi-zone deployments than it is today.

Currently, users can enable mTLS in both single-zone and multi-zone deployments just by configuring it in the Mesh resource. As we introduce features like trust domains, external identities, or federation, we must preserve this simplicity and avoid adding unnecessary complexity to the mTLS setup process.

### As a user, I want to be able to communicate with applications in the different trust domain.

If a user has a different trust domain in each zone, it should be possible to allow communication between these zones.

### As a Mesh operator, I want to be able to define a trustDomain for an entire mesh or zone

It makes sense to allow users to define a trust domain for a specific part of the infrastructure: whether that's a zone or an entire mesh. However, this becomes tricky when balanced against the need for stricter security guarantees in other areas.

### As a Mesh operator, I should be prevented from configuring the same trust domain in multiple zones when not using SPIRE as the identity provider.

This point is a bit tricky for me, because on one hand, it makes sense to allow users to define a trust domain for an entire mesh (which may span multiple clusters). On the other hand, preventing users from using the same trust domain across different clusters improves security compliance.

By assigning unique trust domains per cluster, we achieve security isolation: if one cluster is compromised, the others remain secure. Additionally, it helps with identity scoping, since we can clearly determine the origin of each identity.

Default trust domain: Mesh name

### As a Mesh operator, I want to be able to migrate some workloads to another trust domain without interrupting traffic.

I have a workloads in a trust domain `example.com` and I want to move them to `example.org`. I should be able to do it and don't break the traffic.

### As a user, I want certificate to be valid for a short period of time

Based on the [specification](https://spiffe.io/docs/latest/spiffe-about/spiffe-concepts/#spiffe-workload-api) certificates should be short-lived.

## Out of scope

### As a user, I want to specify certificates for a specific outbound

You should use MeshExternalService in this case to configure the outbound connection and provide the required certificates.

### As a user, I want my dataplanes to communicate with Istio services without going through a gateway

In this scenario, you should also use MeshExternalService and include the necessary certificates in the configuration to enable direct mTLS communication.

### As a user, I want to use two different sources of CA (Kuma and SPIRE) at the same time

This is a complex use case and can make the configuration error prone and difficult to manage.
Instead, we recommend choosing one of the following approaches:

* Use Kuma as the CA and configure other zones or services to trust its certificates (i.e., share the same root/intermediate CA).
* Use SPIRE as the CA and federate it across zones and meshes to provide a unified identity source.

### As a user, I want to have trust domain per namespace

I think we shouldn't allow higher granularity than zone or mesh.
