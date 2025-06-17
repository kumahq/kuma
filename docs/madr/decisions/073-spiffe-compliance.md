# SPIFFE Compliance

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/13659

## Context and Problem Statement

In today’s service mesh systems, secure communication between services is essential. To make that possible, each service needs a clear and trusted identity: so we know exactly who is talking to whom and who we can trust.

We should define two crucial concepts:
* **Identity** - (Who a workload is) — A workload's identity is the name encoded in its certificate, and this identity is considered valid only if the certificate is signed by a Trust.
* **Trust** - (Who to believe) - Trust defines which identities you accept as valid, and is established through trusted certificate authorities (CAs) that issue those identities. Trust is attached to trust domain, and there can be mutliple Trusts in the cluster.

One of the most trusted standards for defining these service identities is SPIFFE (Secure Production Identity Framework for Everyone). It gives us:

* A standard way to name services: `spiffe://<trust-domain>/<path>`
* A strict format for certificates, especially the Subject Alternative Name (SAN) - which must contain exactly one URI (the SPIFFE ID)
* Tools and rules for issuing, rotating, and checking these identities
* SPIFFE certificates identify workload, not an endpoint or API

Right now, Kuma does something similar by generating SPIFFE-like IDs like `spiffe://<mesh-name>/<service-name>`. But it also adds extra SANs (like `kuma://kuma.io/version/v1`) based on tags. These extra SANs make Kuma's certificates non-compliant with SPIFFE, meaning they can’t be used directly with SPIFFE-based systems like SPIRE, or other tools that expect full compliance. Also, Kuma uses certificates to identify endpoints, whereas SPIFFE-based certificates are used to establish identity via SPIFFE IDs of the workload.

In SPIFFE, a Trust Domain is a logical boundary for identity and trust, similar to a DNS domain. Services within the same trust domain can securely identify and trust each other. Communication across different trust domains (such as between organizations or clusters) requires federation, which establishes trust relationships so that identities from one trust domain are recognized and trusted by another.

Kuma currently uses the mesh name as the trust domain. While this works inside Kuma, it doesn’t clearly separate trust boundaries, and it limits how easily Kuma can connect with external services or other applications in the same trust domain.

This setup causes problems when:

* You want to connect Kuma with other SPIFFE-compliant systems outside the mesh
* You run multiple clusters that need to trust each other
* You work in environments that require strict identity standards for security or compliance reasons
* You have identity not coming from Kuma (.e.g SPIRE)

To fully support SPIFFE and modern security practices, Kuma should:

* Issue certificates that include only one SPIFFE-compliant URI SAN (currently we generate `spiffe://<mesh>/<service>` for each inbound and `kuma://kuma.io/<tag_name>/<tag_value>` for each tag)
* Let users set the trust domain explicitly, not tie it to the mesh name
* Allow using other systems (like SPIRE) to manage identities
* Kuma should separate issuing identity from validating trust.

## User Stories

### As a user, I want a Dataplane (e.g., Service A) inside the mesh to accept mTLS connections based on Trusts.

A user should be able to use certificates signed by the same Certificate Authority (CA) across different services. This capability simplifies migration paths and unlocks use cases such as enabling mTLS communication between services inside the mesh and components outside the mesh (e.g., a gateway without a sidecar proxy). It's also worth mentioning that the CA might be managed by another component e.g.: SPIRE, which can issue certificates that are SPIFFE compliant.

### As a user, I want to be able to change the way I issue Identity without interrupting existing connections or causing service downtime.

During the entire process, communication between services must remain secure to ensure uninterrupted, encrypted traffic. This is essential for maintaining zero-trust security guarantees and ensuring operational reliability in production environments.

Example:
I want to move from using certificates issued by Kuma to SPIRE in different trust domain. 

### As a user, it should not become harder to enable or manage mTLS in either single-zone or multi-zone deployments than it is today.

Currently, users can enable mTLS in both single-zone and multi-zone deployments just by configuring it in the Mesh resource. As we introduce features like trust domains, external identities, or federation, we must preserve this simplicity and avoid adding unnecessary complexity to the mTLS setup process.

Release `2.12.x`: We are going to sync only Trust from global to zones, but not from zone to other zones.

### As a user, I want to be able to communicate with applications in different trust domains.

In the presence of multiple trust domains. Users should be able to trust each domains.

### As a Mesh operator, I want to be able to define Identity issuers for subset of Dataplanes.

It makes sense to allow users to define a trust domain for specific parts of the infrastructure — whether that's a single zone or an entire mesh. Additionally, it should be possible to select a specific subset of dataplanes to use a particular identity provider.

One important use case is identity rotation or migration. During migration, multiple trust domains can be used simultaneously to support a smooth transition from one identity provider to another without service disruption.

From a security perspective, assigning separate identities or trust domains per zone provides better isolation. If one trust domain is compromised, it does not impact the others — enabling a stronger and more granular trust model.

Suggestions trust domain on global: `spiffe://<mesh>.<KumaClusterId>.kuma.io/` - is generated on Global, to ensure uniquness

### As a Mesh operator, I want to be able to migrate some workloads to another trust domain without interrupting traffic.

I have a workloads in a trust domain `example.com` and I want to move them to `example.org`. I should be able to do it and don't break the traffic.

### As a user, I want certificate to be valid for a short period of time

Based on the [specification](https://spiffe.io/docs/latest/spiffe-about/spiffe-concepts/#spiffe-workload-api) certificates should be short-lived.

## Future User Stories (things that are going to be implemented in future releases)

### As a user, I want to have a SPIFFE-compliant SAN in a non-Kubernetes environment

It would be beneficial to support connectivity between different environments. Unfortunately, our services currently lack proper identity on Universal, which makes this challenging.

### As a user, I want to federate different trust domains using the Federation API exposed by the control plane

It should be possible to federate different trust domains without creating additional resources, only by using the federation endpoint exposed by the control plane. This would be a valuable feature, enabling interoperability between different SPIFFE-compliant systems.

### As a user, I want to use an intermediate CA instead of a root CA on the zone

When a user provides a root CA on the global control plane, we could generate a separate intermediate CA for each connecting zone. This would offer a stronger security model, ensuring that if a certificate is compromised in one zone, it does not impact other zones.

### As a user, I want to Trust from my zone to be synced to other zones automatically.

Trust could be automatically generated by the control plane and propagated to other zones. While this would improve automation and reduce configuration overhead, it is not critical at this stage and can be addressed in a future iteration.

### As a user, I want to use all available CA providers from legacy mTLS in new mTLS setup

Kuma currently supports various CA (Certificate Authority) backends for identity management. The new mTLS system should at least support the same set of CA providers, to ensure a smooth transition and preserve compatibility with existing deployments.

### As a user, I want to migrate from Kuma non-SPIFFE complient certificates to SPIRE

We should allow users to migrate from Kuma issued certificates, which are not SPIFFE-compliant, to using SPIRE. This may require an intermediate step to first migrate to SPIFFE-compliant certificates, but the details will be specified later.

## Out of scope

### As a user, I want to specify certificates for a specific outbound or inbound

You should use MeshExternalService in this case to configure the outbound connection and provide the required certificates. Dataplane has at the time only one Identity.

### As a user, I want my dataplanes to communicate with Istio services without going through a gateway

In this scenario, you should also use MeshExternalService and include the necessary certificates in the configuration to enable direct mTLS communication.

### As a user, I want to use two different sources of CA (Kuma non-SPIFFE compliant and SPIRE) at the same time

This is a complex use case and can make the configuration error prone and difficult to manage.
We want to add some limitation to avoid issues:
* We should be add a Trust to SPIRE service while using Kuma certificate.
* Identities must be issued by one system only — either SPIRE or Kuma, never both for the same dataplane.
* It's valid for one dataplane to get its identity from Kuma and another from SPIRE — communication between such dataplanes should still work.

To enable this setup, we’ll need to establish a federation mechanism — potentially by having Kuma expose a federation endpoint.

### As a user, I want to provide MeshTrafficPermission for SPIFFEID

This should be covered by a separate MADR in a separate [issue](https://github.com/kumahq/kuma/issues/12374).

## Summary

Release 2.12
1. The system is SPIFFE-compliant on Kubernetes.
2. Certificate configuration is defined outside of the Mesh object.
3. It is possible to manually federate different trust domains.
4. The system can accept mTLS traffic from services outside the mesh.
5. The mTLS setup is no more complex than it is today.
6. It is possible to run Kuma with SPIRE — users can register entries manually or by using the [spire-controller-manager](https://github.com/spiffe/spire-controller-manager).
7. Migration from old mTLS to SPIFFE-compliant mTLS should be possible
8. Let's support: builtin, provided and SPIRE as Identity providers

Future releases
1. The dataplane sends a CSR (Certificate Signing Request) to the control plane for signing.
2. Each zone uses its own intermediate CA.
3. Universal (non-Kubernetes) workflows are SPIFFE-compliant.
4. Support other Identity providers

Nice to have:
1. Kuma control plane can register entries into SPIRE (How can we attest and trust control-plane?)
2. Kuma exposes a federation endpoint.
