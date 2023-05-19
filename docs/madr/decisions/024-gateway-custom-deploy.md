#  Customizing Gateway Deployment / Pod Objects

- Status: Accepted

Technical Story: https://github.com/kumahq/kuma/issues/5027

## Context and Problem Statement

We want the ability to customize the Deployment spec (and be extension, Pod spec) of the objects created as part of a MeshGateway. Initially there is a end-user requirement for adding:

- serviceAccountName
- podSecurityContext.fsGroup
- securityContext.allowPrivilegeEscalation (for containers)
- securityContext.readOnlyRootFilesystem (for containers)
- securityContext.runAsGroup (for containers)
- securityContext.runAsUser (for containers)

We should probably also add the ability to include metadata (as we do with the equivalent Service customization).

## Considered Options

1. Create a `PodTemplate` object that exposes all the above options (similiar to our support for Services via `serviceTemplate` today).
2. Create a more comprehensive way of customizing the full spec using something like a `ContainerPatch` approach (we should also do this [for Service object customization](https://github.com/kumahq/kuma/issues/4903) as well).
3. Create a `PodTemplate` object that exposes only a subset of the above options (more detail below).

## Decision Outcome

Chosen option: 3. Create a `PodTemplate` object that exposes only a subset of the above options.

We could probably get away with initially exposing only `serviceAccountName`, `readOnlyRootFilesystem` and `fsGroup`. We should set `allowPrivilegeEscalation` to `false` as a matter of good security anyway, so don't need to make that configurable. `runAsGroup` and `runAsUser` are already populated by global config options. There is an argument that setting `readOnlyRootFilesystem` to `true` would also be a good default security setting, however could (is likely to) be a breaking change, as access logs are sometimes (often?) written to local files. In case customer want to set `readOnlyRootFilesystem`, we will also add `/tmp` as an `emptyDir` volume by default.

## Decision Drivers

- Minimal work to unblock end-user requirement for phase 1 of this functionality.
- No breaking changes.
- Works in the same way as our current customization for MeshGateway Service objects.

## Solution

### Additional Configuration
Below are the new structs we'd add. We'd also rename `MeshGatewayServiceMetadata` (that currently holds annotations) to `MeshGatewayObjectMetadata` to make it generic (for service and deployment). 

```go

// MeshGatewayCommonConfig represents the configuration in common for both
// Kuma-managed MeshGateways and Gateway API-managed MeshGateways
type MeshGatewayCommonConfig struct {
	//...

	// PodTemplate configures the Pod owned by this config.
	//
	// +optional
	PodTemplate MeshGatewayPodTemplate `json:"podTemplate,omitempty"`
}

// MeshGatewayPodTemplate holds configuration for a Service.
type MeshGatewayPodTemplate struct {
	// Metadata holds metadata configuration for a Service.
	Metadata MeshGatewayObjectMetadata `json:"metadata,omitempty"`

	// Spec holds some customizable fields of a Pod.
	Spec MeshGatewayPodSpec `json:"spec,omitempty"`
}

// MeshGatewayPodSpec holds customizable fields of a Service spec.
type MeshGatewayPodSpec struct {
	// ServiceAccountName corresponds to PodSpec.ServiceAccountName.
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// PodSecurityContext corresponds to PodSpec.SecurityContext
	// +optional
	PodSecurityContext PodSecurityContext `json:"securityContext,omitempty"`

	// Container corresponds to PodSpec.Container
	// +optional
	Container Container `json:"container,omitempty"`
}

// PodSecurityContext corresponds to PodSpec.SecurityContext
type PodSecurityContext struct {
	// FSGroup corresponds to PodSpec.SecurityContext.FSGroup
	// +optional
	FSGroup int64 `json:"fsGroup,omitempty"`
}

// Container corresponds to PodSpec.Container
type Container struct {
	// ContainerSecurityContext corresponds to PodSpec.Container.SecurityContext
	SecurityContext SecurityContext `json:"securityContext,omitempty"`
}

// SecurityContext corresponds to PodSpec.Container.SecurityContext
type SecurityContext struct {
	// ReadOnlyRootFilesystem corresponds to PodSpec.Container.SecurityContext.ReadOnlyRootFilesystem
	// +optional
	ReadOnlyRootFilesystem bool `json:"readOnlyRootFilesystem,omitempty"`
}

```
We should also add `labels` to the `MeshGatewayMetadata` object.

```go
// MeshGatewayObjectMetadata holds Service metadata.
type MeshGatewayObjectMetadata struct {
	// Annotations holds annotations to be set on an object.
	Annotations map[string]string `json:"annotations,omitempty"`

	// Labels holds labels to be set on an objects.
	Labels map[string]string `json:"labels,omitempty"`
}
```
