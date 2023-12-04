# TargetRef policies as a default

* Status: accepted

Technical Story: none

## Context and Problem Statement

At present, Kuma has two policy types:

* Legacy policies with Source and Destination attributes
* New policies based on TargetRef
 
The legacy policies currently serve as the defaults. However, starting from release 2.6, the intention is to switch the default policy to TargetRef.

## Considered Options

1. Introduce a `policyEngine` field within the `Mesh` object, allowing the selection of the default policy version.

## Decision Outcome

The decision is to introduce a `policyEngine` field within the Mesh object to select the default policy version.

### Positive Consequences

* New deployments will start with the new default TargetRef policies.
* Kuma remains capable of working with and applying old policies.
* Existing meshes remain unaffected by the change.

### Negative Consequences

* Possibility of an additional, potentially unused field in the mesh object in the future.

## Pros and Cons of the Options

### Introduce a `policyEngine` field within the `Mesh` object, allowing the selection of the default policy version.


The proposed addition to the Mesh object would appear as follows in the protobuf definition:

```protobuf
// Mesh defines configuration of a single mesh.
message Mesh {

  // ...
    
  enum PolicyEngine {
    // An Undefined mode means that the policy engine wasn't defined.
    // When creating, the default is set to TargetRef.
    // During retrieval of old settings, distinguishing it as Legacy.
    Undefined = 0;
    // A Legacy mode creates old default policies.
    Legacy = 0;
    // A TargetRef mode creates new default targetRef policies.
    TargetRef = 1;
  }

  // policyEngine implies which policy engine should be used as a default.
  PolicyEngine policyEngine = 9;
}
```

New configuration would look like:

```yaml
policyEngine: TargetRef
```

This approach enables the creation of new default target ref policies without enforcing users to only use new polcies. The control-plane will support both old and new policies, with the default `Mesh` being created using the new policies. New policies will not require `MeshTrafficPermissions`, `MeshHTTPRoute`, or `MeshTCPRoute` to enable traffic across the cluster.

Potential future enhancements may include introducing another mode, such as `TargetRefOnly`, which disables old policies. This could be implemented in `MeshContext` by excluding existing legacy policies from the context. On the API level, checks could be performed to verify if the `Mesh` permits the creation of legacy resources.

#### Existing users behaviour

When a user already has a `Mesh` and doesn't define the `policyEngine` field, we treat it as an `Undefined` engine. Only when a user creates a new mesh does the mesh defaulter set an engine to `TargetRef`.

Why do we need `Undefined`?

Protobuf, by default, takes the first value if there is no value provided. Without `Undefined`, we would default to `Legacy`. The issue arises when a user intends to create a new `Mesh` without specifying the `policyEngine` field, resulting in the selection of the first option due to the lack of an optional `policyEngine` field on the protobuf level.

This situation creates problem in discovering between cases where the user did not provide the definition or explicitly chose `Legacy`.
Another issue is when retrieving the old mesh from the storage and there is no field provided. In this case we can treat `Undefined` as a `Legacy`.

#### New kuma users behaviour

When a new user creates a `Mesh` and does not define the `policyEngine` field, it defaults to a `TargetRef` engine.

The control-plane does not generate default `MeshTrafficPermissions`, `MeshHTTPRoute`, or `MeshTCPRoute` because they are not necessary for the traffic to function.

#### ExternalServices and new policies

Under legacy policies, `ExternalServices` are filtered by `MeshTrafficPermissions`, potentially restricting user access to some services. However, in the new policies, we intend to change this approach.

This action would enable each dataplane to communicate with all `ExternalServices`, with the option to filter them out using either `reachableServices` or a new mechanism called `autoReachableServices`.

#### Pros

* Greater flexibility in switching between policy engines.
* Users can create new meshes exclusively with new policies while existing meshes continue to use the legacy engine.
* Possibility to extend the mode by introducing support for only new policies in specific Mesh instances.

#### Cons

* Disabling the creation of legacy policies might necessitate an additional call for the `Mesh` object.
