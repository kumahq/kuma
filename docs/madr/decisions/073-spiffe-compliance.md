# SPIFFE Compliance

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/13659

## Context and Problem Statement

In today’s service mesh systems, secure communication between services is essential. To make that possible, each service needs a clear and trusted identity: so we know exactly who is talking to whom.

One of the most trusted standards for defining these service identities is SPIFFE (Secure Production Identity Framework for Everyone). It gives us:

* A standard way to name services: `spiffe://<trust-domain>/<path>`
* A strict format for certificates, especially the Subject Alternative Name (SAN) - which must contain exactly one URI (the SPIFFE ID)
* Tools and rules for issuing, rotating, and checking these identities
* SPIFFE certificates identify workload, not an endpoint or API

Right now, Kuma does something similar by generating SPIFFE-like IDs like `spiffe://<mesh-name>/<service-name>`. But it also adds extra SANs (like `kuma://kuma.io/version/v1`) based on tags. These extra SANs make Kuma's certificates non-compliant with SPIFFE, meaning they can’t be used directly with SPIFFE-based systems like SPIRE, or other tools that expect full compliance. Also, Kuma uses certificates to identify endpoints, whereas SPIFFE-based certificates are used to establish identity via SPIFFE IDs of the workload.

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

In the presence of multiple trust domains. Users should be able to trust each domains.

### As a Mesh operator, I want to be able to define trustDomain at different levels.

It makes sense to allow users to define a trust domain for a specific part of the infrastructure: whether that's a zone or an entire mesh.
By assigning unique trust domains per cluster, we achieve security isolation: if one cluster is compromised, the others remain secure. Additionally, it helps with identity scoping, since we can clearly determine the origin of each identity. Another important aspect is enabling rotation or migration. During the migration, we could use two different trust domains to facilitate the transition from one identity provider to another.

Default trust domain: `spiffe://<mesh>.<clusterId>.kuma.io/`

### As a Mesh operator, I want to be able to migrate some workloads to another trust domain without interrupting traffic.

I have a workloads in a trust domain `example.com` and I want to move them to `example.org`. I should be able to do it and don't break the traffic.

### As a user, I want certificate to be valid for a short period of time

Based on the [specification](https://spiffe.io/docs/latest/spiffe-about/spiffe-concepts/#spiffe-workload-api) certificates should be short-lived.

### As a user, I want to migrate from Kuma non-SPIFFE complient certificates to SPIRE

We should allow users to migrate from Kuma issued certificates, which are not SPIFFE-compliant, to using SPIRE. This may require an intermediate step to first migrate to SPIFFE-compliant certificates, but the details will be specified later.

## Out of scope

### As a user, I want to specify certificates for a specific outbound

You should use MeshExternalService in this case to configure the outbound connection and provide the required certificates.

### As a user, I want my dataplanes to communicate with Istio services without going through a gateway

In this scenario, you should also use MeshExternalService and include the necessary certificates in the configuration to enable direct mTLS communication.

### As a user, I want to use two different sources of CA (Kuma and SPIRE) at the same time

This is a complex use case and can make the configuration error prone and difficult to manage.
We want to add some limitation to avoid issues:
* A dataplane's identity must come from a single source.
* For trust domains, identities must be issued by one system only — either SPIRE or Kuma, never both for the same dataplane.
* It's valid for one dataplane to get its identity from Kuma and another from SPIRE — communication between such dataplanes should still work.

To enable this setup, we’ll need to establish a federation mechanism — potentially by having Kuma expose a federation endpoint.

### As a user, I want to have trust domain per namespace

I think we shouldn't allow higher granularity than zone or mesh.
