# Common DataSource structure

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/13875

## Context and Problem Statement

While working on SPIFFE compliance, we identified limitations in our existing `DataSource` structure. It currently supports only inline values or secret references:

[Source](https://github.com/kumahq/kuma/blob/master/api/common/v1alpha1/datasource.go)
```golang
// DataSource defines the source of bytes to use.
type DataSource struct {
	// Data source is a secret with given Secret key.
	Secret *string `json:"secret,omitempty"`
	// Data source is inline bytes.
	Inline *[]byte `json:"inline,omitempty"`
	// Data source is inline string`
	InlineString *string `json:"inlineString,omitempty"`
}
```

The current solution has following limitations:
* Users cannot provide values via file path or environment variable.
* Defining private keys inline is insecure, especially since these values are synced to the global control plane.


There is another issue when user wants to explicitly use the `DataSource` in the resource. It's not clear if value provided
by the data source should be accessed by the data plane or control plane. In some resources it makes sense to have it on the control plane
while for others on the data plane.

**Example**

`MeshIdentity` - When a user creates a `MeshIdentity`, they provide a certificate authority (CA) and a private key. These values are read by the control plane and used to sign workload certificates.
```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshIdentity
metadata:
  name: identity
  namespace: kuma-system # only in system namespace
  labels:
    kuma.io/mesh: default
spec:
  ...
  provider:
    type: Bundled
    bundled: # to extend in KM
      ca:
        certificate:
          inline: "my-cert"
        privateKey:
          secret: my-private-key
```

`MeshExternalService` - In contrast, it's unclear how the `DataSource` should behave. Should:
* The control plane read and deliver the data (e.g., via SDS), or
* The Envoy proxy (`ZoneEgress`) directly reference local file paths or environment variables?
```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshExternalService
metadata:
  name: mes-tcp-mtls
  namespace: kuma-system
  labels:
    kuma.io/mesh: default
spec:
  ...
  tls:
    enabled: true
    verification:
      serverName: tcpbin.com
      clientCert:
        inline: "123"
      clientKey:
        secret: my-secret
```

This ambiguity in resolution context leads to confusion and potential misconfiguration, particularly when reusing the same structure across diverse resource types.

**Goal**
* Create a reusable structure applicable across different resources.
* Clearly define where the `DataSource` should be resolved:
  * On the control plane (e.g., secrets)
  * On the data plane (e.g., files or environment variables)

## User stories

### As a Mesh Operator, I want to provide a CA and private key for `MeshIdentity` using a file, environment variable or secret

Mesh Operators should be able to configure a `MeshIdentity` by referencing a file path, environment variable, or `Secret`. This avoids the need to embed sensitive data inline, improving security and aligning with best practices for secret management.

### As a Mesh Operator, I want to configure mTLS for MeshExternalService securely

It should be possible to configure a specific `MeshExternalService` to use a certificate and private key mounted from a file or injected via an environment variable on the `ZoneEgress` - without transmitting them over the network.

### As a Mesh Operator, I want to configure TLS for MeshExternalService

It should be possible to configure only the CA bundle (inline) for verifying the upstream server, as it is safe to expose this information to users.

### As a Mesh Operator, I want to define a MeshExternalService with mTLS settings once and use it across all zones

It should be feasible to define a `MeshExternalService` with all required mTLS configuration in a single location, allowing it to be synced and consumed across multiple zones without additional configuration.

### As a Mesh Operator, I want to prevent users from seeing private keys in the GUI

`Inline` DataSource values can expose sensitive information like private keys in the user interface. For security, `inline` definitions should be discouraged or clearly marked as insecure for use with private keys.

### As a Mesh Operator, I want to configure mTLS for MeshExternalService using a Secret

To maintain a strong security boundary, the control plane should only access `Secrets` from the system namespace (e.g., kuma-system). When a `MeshExternalService` is configured with a `Secret`, it makes sense for the control plane to read that `Secret` and securely deliver it to the data plane via SDS.

## Design

### Option 1: Unified DataSource Structure

```golang
// +kubebuilder:validation:Enum=File;Secret;EnvVar
type Type string

const (
	FileType   Type = "File"
	SecretType Type = "Secret"
	EnvVarType Type = "EnvVar"
)

type DataSource struct {
    Type      Type       `json:"type,omitempty"`
    File      *File      `json:"file,omitempty"`
    SecretRef *SecretRef `json:"secretRef,omitempty"`
    EnvVar    *EnvVar    `json:"envVar,omitempty"`
}

type File struct {
    Path string `json:"path,omitempty"`
}

// +kubebuilder:validation:Enum=Secret
type RefType string

const (
    SecretRefType RefType = "Secret"
)

type SecretRef struct {
    Kind RefType `json:"kind,omitempty"`
    Name string  `json:"name,omitempty"`
}

type EnvVar struct {
    Name string `json:"name,omitempty"`
}
```

This model is extensible and aligns with Kubernetes conventions (e.g., `SecretRef`). However, it introduces ambiguity regarding where the value should be resolved.
To address this, we could introduce a resource-specific `ResourceDescriptor` parameter that defines where the `DataSource` should be resolved:
* Control plane - for example, when signing certificates in `MeshIdentity`.
* Data plane - for example, when performing mutual TLS in `MeshExternalService` via `ZoneEgress`.

#### Pros
* Provides a unified structure.
* Can be reused consistently across different resources.
* Inline secrets are no longer synced to the global control plane, improving security.

#### Cons
* Responsibility for resolving the configuration (control plane vs. data plane) is ambiguous.
* It's unclear where values like file paths or environment variables should be configured - in the control plane or data plane context.
* There's a potential security risk if private keys are delivered by the control plane.
* The API is not self-explanatory: the resolution behavior is only discoverable by reading the code.

### Option 2: Split Structures by Resolution Context

```golang
// +kubebuilder:validation:Enum=File;Secret;EnvVar
type CPType string

const (
    FileType   CPType = "File"
    SecretType CPType = "Secret"
    EnvVarType CPType = "EnvVar"
)

type ControlPlaneDataSource struct {
    Type      CPType     `json:"type,omitempty"`
    File      *File      `json:"file,omitempty"`
    SecretRef *SecretRef `json:"secretRef,omitempty"`
    EnvVar    *EnvVar    `json:"envVar,omitempty"`
}

// +kubebuilder:validation:Enum=File;EnvVar
type DpType string

const (
    FileType   DpType = "File"
    EnvVarType DpType = "EnvVar"
)

type DataplaneDataSource struct {
    Type   DpType  `json:"type,omitempty"`
    File   *File   `json:"file,omitempty"`
    EnvVar *EnvVar `json:"envVar,omitempty"`
}

type File struct {
    Path string `json:"path,omitempty"`
}

// +kubebuilder:validation:Enum=Secret
type RefType string

const (
    SecretRefType RefType = "Secret"
)

type SecretRef struct {
    Kind RefType `json:"kind,omitempty"`
    Name string  `json:"name,omitempty"`
}

type EnvVar struct {
    Name string `json:"name,omitempty"`
}
```

The proposed model addresses the issues of the previous approach by clearly specifying where the `DataSource` should be resolved - either on the control plane or the data plane. It also avoids exposing unsupported options, such as referencing a `Secret` in contexts where it cannot be accessed by the data plane.

#### Pros
* Clear API boundaries — it’s explicitly defined what can be configured for each resource type.
* Improved security — avoids delivering sensitive materials like private keys to the data plane via SDS.
* Inline secrets are no longer synced to the global control plane.

#### Cons
* Requires maintaining two similar but separate structures, which introduces duplication and potential maintenance overhead.
* No explicit support for Secret in the data plane context — users must manually mount secrets as files or environment variables.
* No support for inline values.

### Option 3: Separate Structures for Secrets and Other Data Sources with an Explicit Resolution Context

```golang
// +kubebuilder:validation:Enum=File;Secret;EnvVar;InsecureInline
type SecretDataSourceType string

const (
	SecretDataSourceFile      SecretDataSourceType = "File"
	SecretDataSourceSecretRef SecretDataSourceType = "Secret"
	SecretDataSourceEnvVar    SecretDataSourceType = "EnvVar"
	SecretDataSourceInline    SecretDataSourceType = "InsecureInline"
)

type SecretDataSource struct {
	Type           SecretDataSourceType `json:"type,omitempty"`
	File           *File                `json:"file,omitempty"`
	SecretRef      *SecretRef           `json:"secretRef,omitempty"`
	EnvVar         *EnvVar              `json:"envVar,omitempty"`
	InsecureInline *Inline              `json:"insecureInline,omitempty"`
}

// +kubebuilder:validation:Enum=File;EnvVar;Inline
type DataSourceType string

const (
	DataSourceFile   DataSourceType = "File"
	DataSourceEnvVar DataSourceType = "EnvVar"
	DataSourceInline DataSourceType = "InsecureInline"
)

type DataSource struct {
	Type   DataSourceType `json:"type,omitempty"`
	File   *File          `json:"file,omitempty"`
	EnvVar *EnvVar        `json:"envVar,omitempty"`
	Inline *Inline        `json:"inline,omitempty"`
}

// +kubebuilder:validation:Enum=ControlPlane;DataPlane
// default: ControlPlane
type ResolveOn string

const (
	ResolveOnCP ResolveOn = "ControlPlane"
	ResolveOnDP ResolveOn = "DataPlane"
)

type File struct {
	ResolveOn ResolveOn `json:"resolveOn,omitempty"`
	Path      string    `json:"path,omitempty"`
}

type EnvVar struct {
	ResolveOn ResolveOn `json:"resolveOn,omitempty"`
	Name      string    `json:"name,omitempty"`
}

// +kubebuilder:validation:Enum=Secret
type RefType string

const (
	SecretRefType RefType = "Secret"
)

type SecretRef struct {
	Kind RefType `json:"kind,omitempty"`
	Name string  `json:"name,omitempty"`
}

type Inline struct {
	Value string `json:"value,omitempty"`
}
```

The proposed model introduces clear semantics for resource resolution, allowing users to specify whether a file (or other data source) should be loaded by the control plane or the data plane. Additionally, it separates secret-related data sources from general ones, making it clearer which options are allowed in which contexts.

#### Pros
* Clear API boundaries - users can explicitly configure where the resource should be loaded.
* Inline secrets are clearly marked as insecure, improving visibility and security awareness.

#### Cons
* Two separate models, which may introduce some code duplication and increase maintenance overhead.

### Option 4: Separate Structures for Secrets and Other Data Sources with an Resolution Context directly in the resource

The idea is to avoid embedding the resolution logic (i.e., where the resource should be resolved) into the `DataSource` itself. Instead, this responsibility should remain with the resource that uses it. This way, we keep the `DataSource` object clean and focused.

```golang
// +kubebuilder:validation:Enum=File;Secret;EnvVar;InsecureInline
type SecretDataSourceType string

const (
	SecretDataSourceFile      SecretDataSourceType = "File"
	SecretDataSourceSecretRef SecretDataSourceType = "Secret"
	SecretDataSourceEnvVar    SecretDataSourceType = "EnvVar"
	SecretDataSourceInline    SecretDataSourceType = "InsecureInline"
)

type SecretDataSource struct {
	Type           SecretDataSourceType `json:"type,omitempty"`
	File           *File                `json:"file,omitempty"`
	SecretRef      *SecretRef           `json:"secretRef,omitempty"`
	EnvVar         *EnvVar              `json:"envVar,omitempty"`
	InsecureInline *Inline              `json:"insecureInline,omitempty"`
}

// +kubebuilder:validation:Enum=File;EnvVar;Inline
type DataSourceType string

const (
	DataSourceFile   DataSourceType = "File"
	DataSourceEnvVar DataSourceType = "EnvVar"
	DataSourceInline DataSourceType = "InsecureInline"
)

type DataSource struct {
	Type   DataSourceType `json:"type,omitempty"`
	File   *File          `json:"file,omitempty"`
	EnvVar *EnvVar        `json:"envVar,omitempty"`
	Inline *Inline        `json:"inline,omitempty"`
}

type File struct {
	Path string `json:"path,omitempty"`
}

type EnvVar struct {
	Name string `json:"name,omitempty"`
}

// +kubebuilder:validation:Enum=Secret
type RefType string

const (
	SecretRefType RefType = "Secret"
)

type SecretRef struct {
	Kind RefType `json:"kind,omitempty"`
	Name string  `json:"name,omitempty"`
}

type Inline struct {
	Value string `json:"value,omitempty"`
}
```

#### How would it looks for `MeshExternalService`?

> [!NOTE]
> This is not a finalized design for `MeshExternalService` and should be covered by a separate design document.

For `MeshExternalService`, we could introduce an additional field in the certificate definition, such as `type`, with values like `ZoneEgress` or `ControlPlane`. This field would indicate where the resource should be resolved.

By doing so, we can:

* Keep the DataSource structure small and focused.
* Avoid embedding resolution logic directly inside `DataSource`.
* Allow individual resources (like `MeshExternalService`) to define resolution semantics explicitly and contextually.

This approach promotes cleaner separation of concerns and aligns better with the needs of different resource types.

```yaml
      clientCert:
        type: ControlPlane
        controlPlane:
          type: File | EnvVar | Secret | InsecureInline
          file:
            path: "/cert"
      clientKey:
        zoneEgress:
          type: EnvVar
          envVar:
            name: "PRIVATE_KEY"
```
or
```yaml
    auth:  
      type: ZoneEgress | ControlPlane
      zoneEgress:
        clientCert:
          type: File
          file:
            path: "/cert"
        clientKey:
          type: EnvVar
          envVar:
            name: "PRIVATE_KEY"
      controlPlane:
        clientCert:
          type: File
          file:
            path: "/cert"
        clientKey:
          type: EnvVar
          envVar:
            name: "PRIVATE_KEY"
```

In the case of `MeshIdentity`, we don’t need a `type` field because the data is always accessed by the Control Plane, and there’s no need to control or vary the resolution behavior.

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshIdentity
metadata:
  name: identity
  namespace: kuma-system # only in system namespace
  labels:
    kuma.io/mesh: default
spec:
  ...
  provider:
    type: Bundled
    bundled: # to extend in KM
      ca:
        certificate:
          type: EnvVar
          envVar:
            name: "PRIVATE_KEY"
        privateKey:
          type: EnvVar
          envVar:
            name: "PRIVATE_KEY"
```

#### Pros
* Clear API boundaries - users can explicitly define where the resource should be resolved (control plane or data plane).
* Inline secrets are clearly marked as insecure, improving security visibility and awareness.
* Resolution logic is embedded in the object itself, making it clear what is supported and expected for each resource.

#### Cons
* Two separate models, which may introduce some code duplication and increase maintenance overhead.

## Security implications and review

Following changes increases security:
* Inline secrets are explicitly marked as `insecure`, helping to discourage their use in production environments.
* In Option 2, the private key is not transmitted over the network to the data plane.

## Reliability implications

## Implications for Kong Mesh

`MeshOPA` currently uses the existing API. If we decide to change it, we should either update `MeshOPA` accordingly (introducing a breaking change), or maintain the existing API for backward compatibility.

## Decision

Option 4 offers the most flexibility, enabling users to explicitly (on specific resource) choose whether a value is resolved by the control plane or accessed directly by the data plane. Additionally, by providing two separate models, it becomes easier to understand which options are available and valid in each API context.
