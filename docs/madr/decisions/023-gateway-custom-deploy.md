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

1. Create a `DeploymentTemplate` object that exposes all the above options (similiar to our support for Services via `serviceTemplate` today).
2. Create a more comprehensive way of customizing the full spec using something like a `ContainerPatch` approach (we should also do this [for Service object customization](https://github.com/kumahq/kuma/issues/4903) as well).
3. Create a `DeploymentTemplate` object that exposes only a subset of the above options (more detail below).

## Decision Outcome

Chosen option: 3. Create a `DeploymentTemplate` object that exposes only a subset of the above options.

We could probably get away with initially exposing only `serviceAccountName`, `readOnlyRootFilesystem` and `fsGroup`. We should set `allowPrivilegeEscalation` to `false` as a matter of good security anyway, so don't need to make that configurable. `runAsGroup` and `runAsUser` are already populated by global config options. There is an argument that setting `readOnlyRootFilesystem` to `true` would also be a good default security setting, however could (is likely to) be a breaking change, as access logs are sometimes (often?) written to local files.

## Decision Drivers

- Minimal work to unblock end-user requirement for phase 1 of this functionality.
- No breaking changes.
- Works in the same way as our current customization for MeshGateway Service objects.

## Solution

### Additional Configuration
Below are the new structs we'd add. We'd also rename `MeshGatewayServiceMetadata` (that currently holds annotations) to `MeshGatewayMetadata` to make it generic (for service and deployment). 

```go
// MeshGatewayCommonConfig represents the configuration in common for both
// Kuma-managed MeshGateways and Gateway API-managed MeshGateways
type MeshGatewayCommonConfig struct {
	//...
	// DeploymentTemplate configures the Deployment owned by this config.
	// +optional
	DeploymentTemplate MeshGatewayDeploymentTemplate `json:"deploymentTemplate,omitempty"`
}

// +k8s:deepcopy-gen=true

// MeshGatewayDeploymentTemplate holds configuration for a Deployment.
type MeshGatewayDeploymentTemplate struct {
	// Metadata holds metadata configuration for a K8s object.
	Metadata MeshGatewayMetadata `json:"metadata,omitempty"`

	// Spec holds some customizable fields of a SerDeploymentvice.
	Spec MeshGatewayDeploymentSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen=true

// MeshGatewayDeploymentSpec holds customizable fields of a Deployment spec.
type MeshGatewayDeploymentSpec struct {
    // ServiceAccountName corresponds to PodSpec.ServiceAccountName.
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
    // FSGroup corresponds to PodSpec.PodSecurityContext.FSGroup.
	// +optional
	FSGroup int64 `json:"fsGroup,omitempty"`
	// ReadOnlyRootFilesystem corresponds to PodSpec.Containers[].SecurityContext.ReadOnlyRootFilesystem.
	// +optional
	ReadOnlyRootFilesystem bool `json:"readOnlyRootFilesystem,omitempty"`
}

```
We should also add `labels` to the `MeshGatewayMetadata` object.

```go
// +k8s:deepcopy-gen=true

// MeshGatewayServiceMetadata holds Service metadata.
type MeshGatewayMetadata struct {
	// Annotations holds annotations to be set on a Service.
	Annotations map[string]string `json:"annotations,omitempty"`
	// Labels holds annotations to be set on a Pod.
	Labels map[string]string `json:"labels,omitempty"`
}
```
