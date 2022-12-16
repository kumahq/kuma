# MeshProxyTemplate

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/4737

## Context and Problem Statement

We need to revamp ProxyTemplate for the new policy matching.
Proxy Template lets you adjust Envoy Config when Kuma policies do not provide a native way to configure thing you want.
In a secured setup, it's only safe for mesh operators to operate Proxy Template. Otherwise, service owners could bypass any policy.

There are a couple of problems with existing ProxyTemplate policy 

### No aggregation

Let say that there is a policy that selects all DP proxies and tweaks one value.
Then, if there is another ProxyTemplate that matches specific DPP, the global one is not applied anymore.

### Profile imports section

The policy contains `imports` section which is not really useful.
Currently, you need to specify a profile in `imports`, otherwise the default config won't be generated.
Available values are `default-profile` or `gateway-profile`.

One use case is to skip default config generation and build your own config by scratch.
However, we haven't seen this in the wild. You can always apply modifications that delete all the clusters, listeners etc.

Additionally, the requirement of profile for a default config is a blocker when you want to apply mesh wide proxy template (`kuma.io/service: '*'`), but you have both sidecar and gateway proxies.

When we created we were predicting that we may have user-created profiles.
This never happened in the end. We only used profiles internally.
If someone would like to build their own config I presume they would rather implement this as a plugin. Maintaing whole Envoy config in ProxyTemplate is not really maintainable solution.
If we ever see a need to skip the default config, we can do this by introducing a new flag like `skipDefaultConfigGeneration` in `defaults` section.

### No modifications on zone proxies

Since ProxyTemplate is a mesh-wide policy, we cannot modify zone proxies.

### Providing parts of config as a secret

Out of scope for this PR. This should be implemented as something "universal" for all policies https://github.com/kumahq/kuma/issues/4010

### Observability

Config Dump in the GUI definitely improved finding out if modification was applied, but we could do a better job anyway.
While applying modifications we can gather statistics of what modification was applied on which resource.
Then we could either log it or add it to Dataplane Insight, so we can show it in the GUI.

This is not required to deliver a new policy, but it would improve the UX. We can implement this as an improvement.

### Other resources

We could support TransportSocketMatch https://github.com/kumahq/kuma/issues/4948
We also do not support any endpoint manipulation, but there was no demand for this so far.

Out of scope for the initial implementation.

## Considered Options

* MeshProxyTemplate

## Decision Outcome

Chosen option: "MeshProxyTemplate".

### MeshProxyTemplate

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshProxyTemplate
metadata:
  name: custom-template-1
  namespace: kuma-system
  labels:
    kuma.io/mesh: default
spec:
  targetRef:
    - kind: Mesh
  defaults:
    appendModifications:
      - cluster:
          operation: add
          value: |
            name: test-cluster
            connectTimeout: 5s
            type: STATIC
      - cluster:
          operation: patch
          match: # optional: if absent, all clusters will be patched
            name: test-cluster # optional: if absent, all clusters regardless of name will be patched
            origin: inbound # optional: if absent, all clusters regardless of its origin will be patched
          value: | # you can specify only part of cluster definition that will be merged into existing cluster
            connectTimeout: 5s
      - cluster:
          operation: remove
          match: # optional: if absent, all clusters will be removed
            name: test-cluster # optional: if absent, all clusters regardless of name will be removed
            origin: inbound # optional: if absent, all clusters regardless of its origin will be removed
```

A couple of things to note when you compare this to the old policy:
* `imports` is gone, solving _Profile imports section_ problem.
* `appendModifications` allows to append modifications. This solves _No aggregation_ problem.
* Because list are appended, there is no way to override or exclude modifications.
  This should be fine for this time being, since proxy templates are managed by the mesh operator.
  We could introduce `modifications` next to `appendModifications` or provide more general mechanism for excluding/overriding.
* `modifications` section stays the same as with the old policy

#### Name alternatives

We could argue that `Proxy Template` may not be the most accurate name. After all, we are not providing template, but we are altering Envoy Config.

A couple of name alternatives:
* MeshEnvoyConfig
* MeshXDSConfig
* MeshEnvoyMods (but the policy can potentially include other stuff than modifications in the future)
* MeshProxyConfig (too generic? sounds like configuration of Kuma DP)

While `MeshEnvoyConfig` seems to be the most accurate name, I'm hesitant to make this change. I think that changing the name may confuse existing users. 
