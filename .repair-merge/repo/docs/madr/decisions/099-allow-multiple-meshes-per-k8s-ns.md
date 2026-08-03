# Allow Multiple Meshes per Kubernetes Namespace

* Status: accepted

* Supersedes: [MADR 093](093-disallow-multiple-meshes-per-k8s-ns.md)

Technical Story: https://github.com/kumahq/kuma/issues/15631

## Context and Problem Statement

[MADR 093](093-disallow-multiple-meshes-per-k8s-ns.md) introduced restrictions on having multiple meshes in a single Kubernetes namespace to prevent Workload resource collisions.

However, [MADR 094](094-zone-proxy-deployment-model.md) introduces mesh-scoped zone proxies,
which requires deploying zone proxies for multiple meshes into the `kuma-system` namespace.
Requiring separate namespaces for each mesh's infrastructure components adds operational complexity with no benefit.

This MADR supersedes MADR 093 to allow multiple meshes per namespace
while handling Workload collisions through controller-level error handling.

## Design

### Allow Multiple Meshes per Namespace

Multiple meshes are allowed in a single Kubernetes namespace.
This is a general change — not scoped only to `kuma-system`.

### Handle Workload Collisions in Controller

Instead of preventing multiple meshes per namespace,
the Workload controller will fail with a clear error message if a Workload name collision occurs across meshes.
This provides explicit feedback to users while not blocking valid use cases like mesh-scoped zone proxies.

For zone proxies specifically,
collisions are inherently avoided by the naming pattern `zone-proxy-<mesh>-<role>`,
which guarantees unique Workload names per mesh.

### Remove Runtime Flag

The `runtime.kubernetes.disallowMultipleMeshesPerNamespace` flag introduced in MADR 093 is removed.
The flag was disabled by default and never flipped to enabled.

## Security implications and review

None.

## Reliability implications

None.

## Implications for Kong Mesh

None.

## Decision

1. Allow multiple meshes per Kubernetes namespace (reverting MADR 093).
2. Handle Workload name collisions in the Workload controller with clear error messages.
3. Remove the `runtime.kubernetes.disallowMultipleMeshesPerNamespace` runtime flag.

