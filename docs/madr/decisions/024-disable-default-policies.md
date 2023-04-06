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
4. Add `skipDefaultPolicies` []string field to allow comma-separated list of policies to skip.

## Decision Outcome

Chosen option: 4. Add `skipDefaultPolicies` []string field to allow comma-separated list of policies to skip.

## Decision Drivers

### Pros
- More granular control over what's created (good for testing as well)
- Predictable, understandable behavior
- Only adds a single additional field to Mesh configuration

### Cons
- CSV string fields can be a bit messy, but there is precedent in the codebase 
(`kumactl install observability --components`).

## Open Questions

Should this be `disableDefault<Policy>` or `enableDefault<Policy>` (as suggested 
[here](https://github.com/kumahq/kuma/issues/3346#issuecomment-1209360627))? The former allows us to not introduce a 
breaking change, but the latter makes more sense as new user maybe? 

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
    repeated string skipDefaultPolicies = 8;
}
```