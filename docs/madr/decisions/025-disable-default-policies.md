#  Disable Creation of Default Policies

- Status: Accepted

Technical Story: https://github.com/kumahq/kuma/issues/3346

## Context and Problem Statement

Everytime a mesh is created, we also create a set of default policies (TrafficPermission, Retry, Timeout, etc...). It's useful in a number of scenarios 
to be able to disable that default creation of policies selectively.

## Considered Options

1. Overload the `skipDefaultMesh` switch to also mean disabling the creation of all default policies (as suggested 
[here](https://github.com/kumahq/kuma/issues/3346#issuecomment-1006600634)).
2. Add a binary (yes / no) `disableDefaultPolicies` option to the `Mesh` object.
3. Add individual switches (e.g. `disableDefaultTrafficPermission`) for each policy.
4. Add `skipCreatingInitialDefaultPolicies` []string field to allow comma-separated list of policies to skip.

## Decision Outcome

Chosen option: 4. Add `skipCreatingInitialDefaultPolicies` []string field to allow comma-separated list of policies to skip.

## Decision Drivers

### Pros
- More granular control over what's created (good for testing as well)
- Predictable, understandable behavior
- Only adds a single additional field to Mesh configuration

### Cons
- CSV string fields can be a bit messy, but there is precedent in the codebase 
(`kumactl install observability --components`).

## Behaviors

This field is immutable once set, for the following reasons:
- Unpredictable / undesirable behavior (deleting policies) once the Mesh is created / in-use.
- The main probable use-case for this setting is to aid programmatic / automated flows for creating a Mesh with a 'blank slate'.

In light of the above, name of the flag changed to `skipCreatingInitialDefaultPolicies`.

## Proposed Implementation

Below is how the new fields would look on a `Mesh` object:

```protobuf
// Mesh defines configuration of a single mesh.
message Mesh {

    option (kuma.mesh.resource).name = "MeshResource";
    option (kuma.mesh.resource).type = "Mesh";
    
    // ...
    
    // List of policies to skip creating by default when the mesh is created.
    // e.g. TrafficPermission, MeshRetry, etc.
    repeated string skipCreatingInitialDefaultPolicies = 8;
}
```

New configuration would look like:

```yaml
skipCreatingInitialDefaultPolicies:
  - "MeshTrafficPermission"
  - "MeshRetry"
```

The `skipCreatingInitialDefaultPolicies` list can also contain a wildcard `*` entry which will skip all default policies. E.g. 

```yaml
skipCreatingInitialDefaultPolicies:
  - "*"
```
