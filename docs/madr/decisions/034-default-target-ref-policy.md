# TargetRef policies as a default

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/8467

## Context and Problem Statement

At present, Kuma has two policy types:

* Legacy policies with Source and Destination attributes
* New policies based on TargetRef
 
The legacy policies currently serve as the defaults. However, starting from release 2.6, the intention is to switch the default policy to TargetRef.

## Considered Options

1. Prefer avoiding the creation of default policies and utilize plugin code in cases where there are no existing default legacy policy.

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

As of the 2.6 release, we have discontinued the creation of default policies when initializing a new mesh. This change in the plugin's code allows us to verify the existence of an older matching policy. In cases where there is no existing policy, we generate the configuration using the plugin code.

#### Existing users behaviour

If a user already has a `Mesh` with default legacy policies, there will be no change in behavior. Configurations are generated based on these policies, ensuring that users should not observe any differences in behavior.

Problem:
What if users update the mesh using CD/Terraform? 

To maintain a consistent state during an upgrade, users must update their repository with default `Timeout`, `CircuitBreaker`, and `Retry` policies.

#### New kuma users behaviour

When the control-plane initiates a default Mesh during the initial installation, we no longer create default policies. This change enables the control-plane to utilize plugin code for generating configurations.

#### ExternalServices and new policies

Under legacy policies, `ExternalServices` are filtered by `TrafficPermissions`, potentially restricting user access to some services. However, in the new policies, we intend to change this approach.

This action would enable each dataplane to communicate with all `ExternalServices`, with the option to filter them out using either `reachableServices` or a new mechanism called `autoReachableServices`.

#### Pros

* The user is aware of the existence of the default policy.
* Reduce configuration and API adjustments that may face deprecation in the future.
* The default adoption of a new policy engine requires minimal modifications.

#### Cons

* Continuous Delivery (CD) users should incorporate default `Timeout`, `CircuitBreaker`, and `Retry` policies prior to performing an upgrade.
