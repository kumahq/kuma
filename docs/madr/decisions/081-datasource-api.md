# Common DataSource structure

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/13875

## Context and Problem Statement

While working on SPIFFE Compliance we noticed that we have a `DataSource` common structure which doesn't cover all possible ways of providing external informations and might have security implications. Currently our model looks:

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

It's easy for the user to configure, but there are some limitations:
* The user cannot provide the information via a file path or environment variable.
* Defining the private key as an inline value does not seem secure. Since the resource is synced to the Global control-plane, the value is also presented on the global control-plane.

Additionally, we want to design a structure that is reusable across multiple resources, rather than something tailored to a single specific use case.

## Design

### Create a one DataSource structure

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

This model is extensible and allows defining one of several specific sources. It also follows the Kubernetes convention by referencing other resources using `SecretRef`.

> Do we need a namespace?
Since we only list secrets from the system namespace (default: `kuma-system`), we likely don't need the namespace field for now.

**Problem**

While attempting to apply this `DataSource` structure, I’ve noticed some ambiguities in the API's behavior.
Let’s consider two examples: `MeshIdentity` and `MeshExternalService`.

`MeshIdentity`
When a user creates a `MeshIdentity`, they are expected to provide a certificate authority (CA) and a private key. These values are read by the control plane and used to sign workload certificates.
This behavior is control-plane centric and secure by design.

`MeshExternalService`
In contrast, for `MeshExternalService`, it's unclear whether:
* The control plane should resolve the DataSource (e.g., read `Secret`, `File`, or `EnvVar`) and pass it to Envoy via SDS, or
* The generated Envoy config should directly reference local paths or environment variables (which must be pre-mounted or pre-set on the ZoneEgress proxy).

This inconsistency makes the API harder to reason about, especially when using a shared structure.

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
          type: File
          file:
            path: "/etc/ssl/certs/my-cert.crt"
        privateKey:
          type: Secret
          secretRef:
            kind: Secret
            name: kuma-private-key
```

In case of `MeshExternalServie` it doesn't seem to be clear if control-plane should read the DataSource or maybe it should be handled by the Envoy.

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
        type: File
        file:
          path: "/etc/ssl/certs/my-cert.crt"
      clientKey:
        type: Secret
        secretRef:
          kind: Secret
          name: kuma-private-key
```

The API is not fully clear in this case. Specifically, it’s unclear whether the resource needs to be provided. Should it be:
* Read by the control plane and delivered to `ZoneEgress` via `SDS`, or
* Used to generate Envoy configuration where:
  * EnvVar and File result in a cluster referencing the path or environment variable on the `ZoneEgress` (pod/VM),
  * Secret is read by the control plane and passed via SDS.

This results in a divergence in behavior across resources. For example:
* In MeshIdentity, the control plane always reads the source and delivers it.
* In MeshExternalService, it’s unclear whether the control plane or the data plane (`ZoneEgress`) is responsible.

**How can we solve it?**

To resolve the ambiguity around how and where a `DataSource` is accessed (i.e., control plane vs data plane), we could introduce a resource specific ResourceDescriptor parameter. This parameter would define whether the `DataSource` should be:

* Resolved by the control plane (e.g., for signing certificates in `MeshIdentity`)
* Accessed at runtime by the data plane (e.g., for mutual TLS in `MeshExternalService` on `ZoneEgress`)

#### Pros
* One unified structure
* Can be reused consistently across different resources
* Inline secrets are not synced to the global anymore

#### Cons
* Unclear who is responsible for delivering the configuration (control plane or ZoneEgress)
* Not obvious where values like file paths or environment variables need to be set (control plane or data plane context)
* Potential security risk: private keys might be delivered by the control plane


### Create different DataSources based on where should be accessed

```golang
// +kubebuilder:validation:Enum=File;Secret;EnvVar
type CPType string

const (
	FileType   CPType = "File"
	SecretType CPType = "Secret"
	EnvVarType CPType = "EnvVar"
)

type ControlPlaneDataSource struct {
	Type      CPType       `json:"type,omitempty"`
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
	Type      DpType       `json:"type,omitempty"`
	File      *File      `json:"file,omitempty"`
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

The proposed model addresses the issues found in the previous approach, as it clearly defines where the DataSource should be resolved (control plane or data plane). It also avoids exposing irrelevant options (e.g., using a Secret where it can’t be accessed by the data plane).

#### Pros
* Clear API boundaries — it’s explicitly defined what can be configured for each resource type.
* Improved security — avoids delivering sensitive material like private keys to the data plane via SDS.
* Inline secrets are not synced to the global anymore

#### Cons
* Two similar but separate structures — which adds some duplication and potential maintenance overhead.
* No explicit support for Secret in data plane context — users would need to manually mount secrets as files or environment variables.
* Lack of inline

## Security implications and review

Following changes increses security:
* secrets are not more visiable in global control-plane
* In case of 2nd option, private key is not delivered by the network to the dataplane

## Reliability implications

## Implications for Kong Mesh

`MeshOPA` uses current API, if we decide to change current we should probably update it (breaking change), or just maintain current.

## Decision

Personally, I feel that Option 2 is clearer and easier to understand. The API is explicit, and it prevents users from specifying configuration options that aren't actually supported or available in the target context (e.g., referencing a secret in a data plane resolved setting).

By separating control-plane and data plane resolved data sources, we make the model more predictable and user-friendly, while avoiding accidental misconfiguration or security issues.