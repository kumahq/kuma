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

This approach allows the creation of new default `TargetRef` policies without mandating users to solely adopt these new policies. To implement this, we need to introduce an environment variable called `KUMA_DEFAULTS_POLICY_ENGINE`. This variable will define the default policy engine when no specific one is provided during the creation or loading process.

```yaml
KUMA_DEFAULTS_POLICY_ENGINE: TargetRef # Possible values TargetRef,Legacy. Default: TargetRef
```

#### Existing users behaviour

When a user already possesses a `Mesh` but doesn't define the `policyEngine` field, we consider it as being set to an `Undefined` engine state. The mesh defaulter only assigns an engine based on `KUMA_DEFAULTS_POLICY_ENGINE` when a user creates a new mesh. By default, this variable is set to `TargetRef`, meaning each new mesh will use `policyEngine: TargetRef`.

Why do we need `Undefined`?

Protobuf, by default, takes the first value if there is no value provided. Without `Undefined`, we would default to `Legacy`. The issue arises when a user intends to create a new `Mesh` without specifying the `policyEngine` field, resulting in the selection of the first option due to the lack of an optional `policyEngine` field on the protobuf level.

This situation creates problem in discovering between cases where the user did not provide the definition or explicitly chose `Legacy`.
An additional concern arises when fetching an older mesh from storage and there is no field provided. In such cases, we can interpret `Undefined` as a value determined by `KUMA_DEFAULTS_POLICY_ENGINE`.

When the user updates the Mesh definition to use the `TargetRef` engine, we won't create default policies. Default policies are only generated upon the creation of a new Mesh.

Problem:
What if users update the mesh using CD/Terraform? 

During an upgrade, existing users have the option to set `KUMA_DEFAULTS_POLICY_ENGINE=Legacy`. This setting will result in all existing meshes or newly created ones without the `policyEngine` field being treated as `Legacy`.

Possible results:
| `KUMA_DEFAULTS_POLICY_ENGINE` | `policyEngine` |  Result Policy Engine |
|              ---              |      ---       |          ---          |
|           TargetRef           |   Undefined    |       TargetRef       |
|           TargetRef           |    Legacy      |        Legacy         |
|           TargetRef           |   TargetRef    |       TargetRef       |
|            Legacy             |   Undefined    |        Legacy         |
|            Legacy             |    Legacy      |        Legacy         |
|            Legacy             |   TargetRef    |       TargetRef       |

#### New kuma users behaviour

When the control-plane creates a `default` Mesh during the initial installation, it operates with the new policy engine called `TargetRef`. The control-plane does not generate default `MeshTrafficPermissions`, `MeshHTTPRoute`, or `MeshTCPRoute` because they are not necessary for the traffic to function. Additionally, certain sections of the code need to validate which engine corresponds to a specific Mesh because the flow varies slightly depending on the engine.

#### ExternalServices and new policies

Under legacy policies, `ExternalServices` are filtered by `MeshTrafficPermissions`, potentially restricting user access to some services. However, in the new policies, we intend to change this approach.

This action would enable each dataplane to communicate with all `ExternalServices`, with the option to filter them out using either `reachableServices` or a new mechanism called `autoReachableServices`.

#### Pros

* Greater flexibility in switching between policy engines.
* Users can create new meshes exclusively with new policies while existing meshes continue to use the legacy engine.
* Possibility to extend the mode by introducing support for only new policies in specific Mesh instances.

#### Cons

* Disabling the creation of legacy policies might necessitate an additional call for the `Mesh` object.
