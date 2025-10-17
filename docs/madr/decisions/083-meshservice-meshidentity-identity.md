# Migration from one SpiffeID to another without a downtime using MeshService

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/14388

## Context and Problem Statement

MeshService stores identities in the object:

```yaml
spec:
  identities:
  - type: ServiceTag
    value: test-server_default_kuma-demo_svc_80
  - type: SpiffeID
    value: spiffe://default.zone.mesh.local/ns/kuma-demo/sa/default
```
`ServiceTag` identity is a value that can be precomputed and set on a `MeshService` without checking whether mTLS is enabled.
In the case of `MeshIdentity`, we set the `SpiffeID` based on the matching resource. We determine which `MeshIdentity` targets a specific `MeshService`, and based on that, we assign the corresponding identity.

Later, this information is used to create the Envoy cluster configuration. Once a `MeshIdentity` is enabled and targets specific dataplanes, Kuma begins delivering identities to each sidecar. At this point, the `MeshService` is recalculated, new identities are added, and these are then translated into SAN matching.

### Where is the issue?

Let’s walk through the flow:

Prerequisites:
* `MeshService` is in `Exclusive` mode
* There is constant traffic from the client to the service

1. Mesh with legacy mTLS (`Mesh.mTLS`) is enabled
2. Client and Server details
* Client presents identity: `client_default_kuma-demo_svc_80`
* Server presents identity: `server_default_kuma-demo_svc_80`

MeshService of Client
```yaml
spec:
  identities:
  - type: ServiceTag
    value: client_default_kuma-demo_svc_80
```

MeshService of Server
```yaml
spec:
  identities:
  - type: ServiceTag
    value: server_default_kuma-demo_svc_80
```
3. Envoy configuration based on MeshService identities
From the `MeshService` definitions, Kuma generates Envoy configuration that enforces identity checks.

**Envoy Cluster for Server**
The cluster expects the server’s identity:
```yaml
              matchTypedSubjectAltNames:
              - matcher:
                  exact: spiffe://default/server_default_kuma-demo_svc_80
                sanType: URI
```

4. We enable `MeshIdentity` with an empty selector.
As a result, Kuma initializes a `MeshIdentity` by creating all required resources: a `Certificate` and a `MeshTrust`.
```yaml
type: MeshIdentity
name: identity
mesh: default
spec:
  selector: {}
  spiffeID:
    trustDomain: "{{ .Mesh }}.{{ .Zone }}.mesh.local"
    path: "/ns/{{ .Namespace }}/sa/{{ .ServiceAccount }}"
  provider:
    type: Bundled
    bundled:
      meshTrustCreation: Enabled
      insecureAllowSelfSigned: true
      certificateParameters:
        expiry: 24h
      autogenerate:
        enabled: true
```
5. `MeshTrust` is propagated, and both the client and the server are able to accept traffic using either the legacy mTLS identities or the new `MeshIdentity`.
6. `MeshIdentity` is now fully enabled but selecting all dataplanes.
```yaml
type: MeshIdentity
name: identity
mesh: default
spec:
  selector:
    dataplane:
      matchLabels: {}
  spiffeID:
    trustDomain: "{{ .Mesh }}.{{ .Zone }}.mesh.local"
    path: "/ns/{{ .Namespace }}/sa/{{ .ServiceAccount }}"
  provider:
    type: Bundled
    bundled:
      meshTrustCreation: Enabled
      insecureAllowSelfSigned: true
      certificateParameters:
        expiry: 24h
      autogenerate:
        enabled: true
```
7. The new identities are delivered to each sidecar.
8. The Client and Server start presenting them:
* Client presents itself as: `spiffe://default.zone.mesh.local/ns/kuma-demo/sa/client`
* Server presents itself as: `spiffe://default.zone.mesh.local/ns/kuma-demo/sa/server`
9. At this point, the `MeshService` **has not yet** been updated and still contains the following configuration:

MeshService of Server
```yaml
spec:
  identities:
  - type: ServiceTag
    value: server_default_kuma-demo_svc_80
```
These translate into Envoy configuration as follows:
Envoy Cluster of Server (client expects):
```yaml
              matchTypedSubjectAltNames:
              - matcher:
                  exact: spiffe://default/server_default_kuma-demo_svc_80
                sanType: URI
```
10.  Traffic breaks because the Server presents itself as:

`spiffe://default.zone.mesh.local/ns/kuma-demo/sa/server` 
but the Client still expects:
`spiffe://default/server_default_kuma-demo_svc_80`

11. After the reconciliation loop (default 5s), the `MeshService` resources are updated:

MeshService of Server
```yaml
spec:
  identities:
  - type: ServiceTag
    value: server_default_kuma-demo_svc_80
  - type: SpiffeID
    value: spiffe://default.zone.mesh.local/ns/kuma-demo/sa/server
```
12.  Envoy configuration is recalculated to accept both ServiceTag and SpiffeID identities:
Envoy Cluster for Client (accepts Server identities):
```yaml
              matchTypedSubjectAltNames:
              - matcher:
                  exact: spiffe://default/client_default_kuma-demo_svc_80
                sanType: URI
              - matcher:
                  exact: spiffe://default.zone.mesh.local/ns/kuma-demo/sa/server
                sanType: URI
```
13.  Traffic resumes successfully

As we can see from this flow diagram, there is a period during which traffic from the new identity is not accepted.

## User stories

### As a Mesh Operator, I want to migrate from legacy mTLS to MeshIdentity without any traffic disruption

This is a real use case: we want to migrate users to a new identity and avoid any problems with traffic during the process.

### As a Mesh Operator, I want to migrate from one trust domain to another without downtime

There might be cases where a user wants to migrate from one trust domain to another, and there should be no problems with traffic during this time.

## Considered Options

- Introduce a new field `mode`
- Generate identity only when there is no provider section
- Use a `kuma.io/effect` label

## Decision Outcome

- Generate identity only when there is no provider section

## Design

### Introduce a new field `mode`

We could introduce a new field into MeshIdentity resource:

```yaml
type: MeshIdentity
name: identity
mesh: default
spec:
  mode: Active (Default) | SAN
  selector:
    dataplane:
      matchLabels: {}
```

In `SAN` mode, we could start generating Identities for `MeshServices`. However, these identities would not yet be delivered to dataplanes. Instead, Envoy clusters would be configured with the SANs corresponding to specific `MeshServices`.

In `Active` mode, identities would be delivered to dataplanes, and they would start using the new certificates.

#### Migration flow

1. The user creates a `MeshIdentity` with `mode: SAN`.
2. `MeshServices` are updated with the corresponding `SpiffeID`.
3. The user switches the `MeshIdentity` mode to `mode: Active`.
4. Traffic continues to work without downtime.

#### Multiple `MeshIdentities` in `Active` mode

This should have no effect, since `MeshIdentities` are selected based on lexicographic order.

#### Pros

* Easy to understand.
* Clearly expressed within the MeshIdentity object.

#### Cons

* Introduces an additional field in the MeshIdentity object.

### Generate identity only when there is no provider section

We could update the validation and reconciler for `MeshIdentity` so that they do not require the provider section.

```yaml
type: MeshIdentity
name: identity
mesh: default
spec:
  selector:
    dataplane:
      matchLabels: {}
  spiffeID:
    trustDomain: "{{ .Mesh }}.{{ .Zone }}.mesh.local"
    path: "/ns/{{ .Namespace }}/sa/{{ .ServiceAccount }}"
```

In this case, we would not create any certificates or issue new identities. The only change would be that the `MeshService` is updated with a new `SpiffeID`.

#### Status field

When a user creates a `MeshIdentity`, a controller sets its status to indicate whether the resource is ready and initializes all related resources.
In this case, since there is no provider, we don’t have an initialization phase. However, we still want to set a status on the resource for better debuggability.

Proposed status:

```golang
        common_api.Condition{
        	Type:    meshidentity_api.SpiffeIDProviderConditionType,
        	Status:  kube_meta.ConditionTrue,
        	Reason:  "SpiffeIDProvided",
        	Message: "Providing only SpiffeIDs for services.",
        },
        common_api.Condition{
        	Type:    meshidentity_api.ReadyConditionType,
        	Status:  kube_meta.ConditionFalse,
        	Reason:  "PartiallyReady",
        	Message: "Running in SpiffeID providing only mode.",
        },
```

#### Migration flow

##### Bundled - autogenerated

1. The user creates a `MeshIdentity` without a provider configuration.
```yaml
type: MeshIdentity
name: identity-spiffe-only
mesh: default
spec:
  selector:
    dataplane:
      matchLabels: {}
  spiffeID:
    trustDomain: "{{ .Mesh }}.{{ .Zone }}.mesh.local"
    path: "/ns/{{ .Namespace }}/sa/{{ .ServiceAccount }}"
```
2. The corresponding `MeshServices` are updated with the new `SpiffeID`.
3. The user creates a new `MeshIdentity` resource with a selector. This automatically generates a corresponding `MeshTrust`, which is then propagated to all services and can accept both legacy mTLS and the new `MeshIdentity` traffic
```yaml
type: MeshIdentity
name: identity
mesh: default
spec:
  selector: {}
  spiffeID:
    trustDomain: "{{ .Mesh }}.{{ .Zone }}.mesh.local"
    path: "/ns/{{ .Namespace }}/sa/{{ .ServiceAccount }}"
  provider:
    type: Bundled
    bundled:
      meshTrustCreation: Enabled
      insecureAllowSelfSigned: true
      certificateParameters:
        expiry: 24h
      autogenerate:
        enabled: true
```
4. Traffic continues to operate without downtime.
5. The user enabled `MeshIdentity` to select all dataplanes, which caused every dataplane to receive a new identity.
```yaml
type: MeshIdentity
name: identity
mesh: default
spec:
  selector:
    dataplane:
      matchLabels: {}
  spiffeID:
    trustDomain: "{{ .Mesh }}.{{ .Zone }}.mesh.local"
    path: "/ns/{{ .Namespace }}/sa/{{ .ServiceAccount }}"
  provider:
    type: Bundled
    bundled:
      meshTrustCreation: Enabled
      insecureAllowSelfSigned: true
      certificateParameters:
        expiry: 24h
      autogenerate:
        enabled: true
```
6. The user can remove the initial `MeshIdentity` (the one without a provider `identity-spiffe-only`), as all dataplanes use a new identity.

##### Bundled - certificate provided by the user

1. The user creates a `MeshIdentity` without a provider configuration.
```yaml
type: MeshIdentity
name: identity-spiffe-only
mesh: default
spec:
  selector:
    dataplane:
      matchLabels: {}
  spiffeID:
    trustDomain: "{{ .Mesh }}.{{ .Zone }}.mesh.local"
    path: "/ns/{{ .Namespace }}/sa/{{ .ServiceAccount }}"
```
2. The corresponding `MeshServices` are updated with the new `SpiffeID`.
3. The user created a `MeshTrust` with the CA of the new identity to allow services to accept new traffic.
```yaml
type: MeshTrust
name: new-trust
mesh: default
spec:
  caBundles:
    - type: Pem
      pem:
        value: |-
          ...
  trustDomain: domain
```
4. The user enabled `MeshIdentity` to select all dataplanes, which caused every dataplane to receive a new identity.
```yaml
type: MeshIdentity
name: identity
mesh: default
spec:
  selector:
    dataplane:
      matchLabels: {}
  spiffeID:
    trustDomain: "{{ .Mesh }}.{{ .Zone }}.mesh.local"
    path: "/ns/{{ .Namespace }}/sa/{{ .ServiceAccount }}"
  provider:
    type: Bundled
    bundled:
      meshTrustCreation: Disabled
      insecureAllowSelfSigned: false
      certificateParameters:
        expiry: 24h
      ca:
        ...
```
1. Traffic continues to operate without downtime.
1. The user can remove the initial `MeshIdentity` (the one without a provider `identity-spiffe-only`), as all dataplanes use a new identity.

#### Multiple `MeshIdentities` in `Active` mode

This should have no effect, since `MeshIdentities` are selected based on lexicographic order.

#### Pros

* Doesn’t require introducing a new field.
* Clean and straightforward to implement.

#### Cons

* Slightly unclear behavior.

### Use a `kuma.io/effect` label

In [MADR](040-transition-to-new-policies.md), we introduced a new label for mesh policies: `kuma.io/effect`. The goal of this label is to help users migrate from the old policies to the new ones. When a user sets `kuma.io/effect: shadow` on a policy, the configuration will not be applied to dataplanes but will allow the user to compare and verify that there are no differences in configuration between the old and new policies.

We could also reuse this label with the value `san`. In this case, identities would be propagated on the `MeshService`, but the workload identity itself would not be provided.

#### Migration flow

1. The user creates a `MeshIdentity` with `kuma.io/effect: san`.
2. `MeshServices` are updated with the corresponding `SpiffeID`.
3. The user removes `kuma.io/effect: san` from the `MeshIdentity` resource.
4. Traffic continues to work without downtime.

#### Pros
* Reuses existing functionality
* No need to introduce a new field

#### Cons
* Less obvious than the previous solution
* Requires proper validation so that the label can only be set on this resource

### Populate spiffeID from all MeshIdentities

It is possible to populate a `MeshService` with all potential `SpiffeID` identities defined in every `MeshIdentity` resource, even if a given resource is not currently enabled. This approach would allow users to migrate from one trust domain to another without requiring additional flags or causing downtime. Instead of looking only for a `MeshIdentity` resource that targets a specific dataplane, the system would resolve all possible values from every `MeshIdentity` and then populate them into the `MeshService`.

#### Pros
* No need to introduce any new mechanisms
* Reuses existing objects

#### Cons
* Potential security risk, as identities not intended for a `MeshService` could still be populated
* Requires more specific code to handle this case
* Adds more SAN matchers in the Envoy cluster configuration
* Increased computation time due to resolving all possible values
