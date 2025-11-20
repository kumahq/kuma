# Disallow Multiple Meshes per Kubernetes Namespace

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/15049

## Context and Problem Statement

Kuma currently allows users to set the `kuma.io/mesh` label on individual pods.
As a result, pods in the same Kubernetes namespace can belong to different meshes.
While having multiple namespaces in a single mesh is a valid and common pattern,
having multiple meshes in a single namespace leads to several issues:

1. **Workload resource collision**. Workload is a mesh scoped resource generated from the `app.kubernetes.io/name` label.
If pods in the same namespace belong to different meshes,
the same `app.kubernetes.io/name` would need to create Workload resources in multiple meshes, which causes collisions.

2. **Operational confusion**. Splitting meshes inside a single namespace is hard to reason about
and does not match typical Kubernetes usage where a namespace acts as an isolation boundary.

3. **Limited real world usage**. Although this configuration is technically supported,
we believe it is rarely used in production.

In practice the main blocker is Workload resource generation,
which cannot behave correctly when a single namespace is spread across multiple meshes.

## Design

### Warn and Skip Workload Generation

Add detection to `workload_controller.go` for multiple meshes per namespace and:
- Skip Workload resource generation when multiple meshes are detected in a namespace
- Emit a Kubernetes event to inform the user about the configuration issue and log the CP error

### Add Runtime Flag with Future Default Change

Add a configuration flag `runtime.kubernetes.disallowMultipleMeshesPerNamespace` that:
- Disabled by default (current behavior)
- When enabled, prevents pods from using `kuma.io/mesh` label that would result in multiple meshes in a namespace
- Flip the default to enabled in the next major Kuma release to persist this behaviour,
remove the feature flag, document it's a breaking change

### Pros and Cons

- Good, because it avoids breaking changes for existing users.
- Good, because it gradually enforces a clean namespace to mesh relationship.
- Bad, because users who do not check logs or events may be confused why a Workload was not generated.
- Bad, because it still requires a migration step in a future release.

## Implications for Kong Mesh

None

