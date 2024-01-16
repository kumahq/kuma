# TargetRef policies as a default

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/8467

## Context and Problem Statement

At present, Kuma has two policy types:

* Legacy policies with Source and Destination attributes
* New policies based on TargetRef
 
The legacy policies currently serve as the defaults. However, starting from release 2.6, the intention is to switch the default policy to TargetRef.

## Considered Options

1. Prefer avoiding the creation of default routing policies and utilize plugin code in cases where there are no existing default legacy policy.
2. Introduce a `policyEngine` field within the `Mesh` object, allowing the selection of the default policy version.

## Decision Outcome

The decision is to stop creating default policies and utilize plugin code in cases where there are no existing default legacy policy.

### Positive Consequences

* New deployments will start with the new policy engine.
* Existing meshes remain unaffected by the change.
* No additional configuration.

### Negative Consequences

* Clients who are recreating their deployment using continuous delivery (CD) need to manually add default `Timeout`, `CircuitBreaker`, and `Retry` policies

## Pros and Cons of the Options

### Prefer avoiding the creation of default policies and utilize plugin code in cases where there are no existing default legacy policy.

As of the 2.6 release, we have discontinued the creation of default policies for `MeshTrafficPermisson` and `Mesh*Route` when initializing a new mesh. This change allows plugins' code to verify the existence of an older matching policy. In cases when there is no existing legacy policy, we generate the configuration using the code for targetRef policies. We are going to create default `MeshTimeout`, `MeshRetry` and `MeshCircuitBreaker` as a replacement of legacy policies to provide basic defaults for the communication.

To support older versions of zones we are going to introduce `KUMA_DEFAULTS_CREATE_MESH_ROUTING_RESOURCES` which by default is disabled. The user can enable `KUMA_DEFAULTS_CREATE_MESH_ROUTING_RESOURCES` and that allows new meshes to be created with legacy policies required for routing: `TrafficRoute` and `TrafficPermission`.

Why do we need need `KUMA_DEFAULTS_CREATE_MESH_ROUTING_RESOURCES`?

When global is newer than a zone, and a new mesh is created it won't have default `TrafficPermission` and `TrafficRoute` policies, without which the older zone won't be able to generate configuration required to route between services. In this case the user can deploy global or federated zone with `KUMA_DEFAULTS_CREATE_MESH_ROUTING_RESOURCES` and create the new mesh that will work with the old zone.

#### Existing users behaviour

If a user already has a `Mesh` with default legacy policies, there will be no change in behavior. Configurations are generated based on these policies, ensuring that users should not observe any differences in behavior.

Problem:
What if users update the mesh using CD/Terraform? 

To maintain a consistent state during an upgrade, users must update their repository with default `TrafficPermission` and `TrafficRoute` or enable `KUMA_DEFAULTS_CREATE_MESH_ROUTING_RESOURCES`.

#### New kuma users behaviour

When the control-plane initiates a default Mesh during the initial installation, we no longer create default policies for `MeshTrafficPermisson` and `Mesh*Route`. This change enables the control-plane to utilize plugin code for generating configurations.

#### ExternalServices and new policies

Under legacy policies, `ExternalServices` are filtered by `TrafficPermissions`, potentially restricting user access to some services. However, in the new policies, we intend to change this approach.

This action would enable each dataplane to communicate with all `ExternalServices`, with the option to filter them out using either `reachableServices` or a new mechanism called `autoReachableServices`.

#### Pros

* The user is aware of the existence of the default policy.
* Reduce configuration and API adjustments that may face deprecation in the future.
* The default adoption of a new policy engine requires minimal modifications.

#### Cons

* New env `KUMA_DEFAULTS_CREATE_MESH_ROUTING_RESOURCES`

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

This approach allows the creation of new default `TargetRef` policies without mandating users to solely adopt these new policies. To implement this, we need to introduce an environment variable called `KUMA_DEFAULTS_USE_LEGACY_POLICY_ENGINE`. This variable will define the default policy engine when no specific one is provided during the creation or loading process.

#### Existing users behaviour

When a user already possesses a `Mesh` but doesn't define the `policyEngine` field, we consider it as being set to an `Undefined` engine state. The mesh defaulter only assigns an engine based on `KUMA_DEFAULTS_USE_LEGACY_POLICY_ENGINE` when a user creates a new mesh. By default, this variable is set to `false`, meaning each new mesh will use `policyEngine: TargetRef`.

Why do we need `Undefined`?

Protobuf, by default, takes the first value if there is no value provided. Without `Undefined`, we would default to `Legacy`. The issue arises when a user intends to create a new `Mesh` without specifying the `policyEngine` field, resulting in the selection of the first option due to the lack of an optional `policyEngine` field on the protobuf level.

This situation creates problem in discovering between cases where the user did not provide the definition or explicitly chose `Legacy`.

An additional concern arises when fetching an older mesh from storage and there is no field provided. In such cases, we can interpret `Undefined` as a value determined by `KUMA_DEFAULTS_USE_LEGACY_POLICY_ENGINE`.

When the user updates the Mesh definition to use the `TargetRef` engine, we won't create default policies. Default policies are only generated upon the creation of a new Mesh.

Problem:
What if users update the mesh using CD/Terraform? 

During an upgrade, existing users have the option to set `KUMA_DEFAULTS_USE_LEGACY_POLICY_ENGINE=true`. This setting will result in all existing meshes or newly created ones without the `policyEngine` field being treated as `Legacy`.

#### New kuma users behaviour

When a new user creates a `Mesh` and does not define the `policyEngine` field, it defaults to a `TargetRef` engine.
The control-plane does not generate default `MeshTrafficPermissions`, `MeshHTTPRoute`, or `MeshTCPRoute` because they are not necessary for the traffic to function.

#### Pros
* Greater flexibility in switching between policy engines.
* Users can create new meshes exclusively with new policies while existing meshes continue to use the legacy engine.
* Possibility to extend the mode by introducing support for only new policies in specific Mesh instances.
#### Cons
* Disabling the creation of legacy policies might necessitate an additional call for the `Mesh` object.
* Problem when global is newer than a zone