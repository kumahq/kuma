# MeshProxyPatch

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
Solving this would require duplicating ProxyTemplate policy with global scope. This won't be covered by this MADR.

### Providing parts of config as a secret

Out of scope for this PR. This should be implemented as something "universal" for all policies https://github.com/kumahq/kuma/issues/4010

### Observability

Config Dump in the GUI definitely improved finding out if modification was applied, but we could do a better job anyway.
While applying modifications we can gather statistics of what modification was applied on which resource.
Then we could either log it or add it to Dataplane Insight, so we can show it in the GUI.

This is not required to deliver a new policy, but it would improve the UX. We can implement this as an improvement.

### Routes are hard to modify

We allow to modify VirtualHost, in which you can replace list of routes. You cannot modify a single route.
You also cannot modify routes if routes are delivered through RDS.

### Merging of TypedConfig

Some parts in Envoy Config are encoded as TypedConfig (marshaled proto in proto).
Merging such configs results in just replacing. That's why we cannot currently patch TransportSocketMatch.

### Appending to lists

The default strategy for list is replace. Sometimes we may want to add something to the list. There is no such option right now.

### Other resources

We could support TransportSocketMatch https://github.com/kumahq/kuma/issues/4948

## Considered Options

* MeshProxyPatch

## Decision Outcome

Chosen option: "MeshProxyPatch".

### MeshProxyPatch

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshProxyPatch
metadata:
  name: custom-template-1
  namespace: kuma-system
  labels:
    kuma.io/mesh: default
spec:
  targetRef:
    kind: Mesh
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
          jsonPatches: # either jsonPatches or value. Only possible to use with operation: patch. See https://jsonpatch.com/ for more.
            - op: add
              from: ""
              path: "/connectTimeout"
              value: "5s"
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
  This should be fine for the time being, since proxy templates are managed by the mesh operator.
  We could introduce `modifications` next to `appendModifications` or provide more general mechanism for excluding/overriding.
* `modifications` section stays almost the same as with the old policy except `jsonPatches` field.
* `jsonPatches` is the alternative way of patching config which we also use in `ContainerPatch`.
  It is more expensive to execute, because before applying the patch we need to marshal the config to JSON.
  It should not be a default way of patching things, but this way we can overcome some limitations of a default proto merge. 
  You can modify TypedConfig, because TypedConfig is serialized to JSON. This solves _Merging of TypedConfig_
  You can modify parts of lists (add element, etc.). This solves _Appending to lists_

#### Name

We could argue that `Proxy Template` may not be the most accurate name. After all, we are not providing template, but we are altering Envoy Config.

A couple of name alternatives:
* MeshEnvoyConfig
* MeshXDSConfig
* MeshEnvoyMods (but the policy can potentially include other stuff than modifications in the future)
* MeshProxyConfig (too generic? sounds like configuration of Kuma DP)
* MeshProxyPatch

We decided to name the new resource `MeshProxyPatch` 

#### Other resources

##### Transport Socket modifications

We could introduce "native" transport_socket modification, but if we introduce `jsonPatches` we can also rely on it.

```yaml
appendModifications:
  - transport_socket:
      operation: add # (or patch or remove)
      match:
        name: "envoy.transport_sockets.tls" # optional: name of the transport socket match on which to apply modification (can be used only with patch and remove) 
        origin: outbound # optional: origin of the resource on which to apply the modification (ex. inbound, outbound)
        resource: cluster # optional: resource type on which to apply the modification (available values: listener, cluster)
        resourceName: "xyz" # optional: exact name of the resource on which to apply the modification
      value: |
        name: envoy.transport_sockets.tls
        typedConfig:
          '@type': 'type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext'
          commonTlsContext:
            tlsParams:
              tlsMinimumProtocolVersion: TLSv1_2
            validationContextSdsSecretConfig:
              name: cp_validation_ctx
          sni: kuma-control-plane.kuma-system
```

Example of overriding TLS version for mesh mTLS

```yaml
appendModifications:
  - transport_socket:
      operation: patch
      match:
        name: "envoy.transport_sockets.tls" 
        origin: outbound
        resource: cluster
      value: |
        name: envoy.transport_sockets.tls
        typedConfig:
          '@type': 'type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext'
          commonTlsContext:
            tlsParams:
              tlsMinimumProtocolVersion: TLSv1_3
  - transport_socket:
      operation: patch
      match:
        name: "envoy.transport_sockets.tls"
        origin: inbound
        resource: listener
      value: |
        name: envoy.transport_sockets.tls
        typedConfig:
          '@type': 'type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.DownstreamTlsContext'
          commonTlsContext:
            tlsParams:
              tlsMinimumProtocolVersion: TLSv1_3
```

##### Route modifications

Route modifications will support ordered modifications by allowing the same operations as `httpFilter` - `addFirst`, `addBefore`, `addAfter`, `addLast`, `patch`, `remove`.
Route modifications will support static routes (in-cluster communication) and dynamic routes through RDS (gateway).
To support name matching, we also need to name all the routes from now.

```yaml
appendModifications:
  - route:
      operation: addFirst # it will be added as a first route in the virtual host.  
      match:
        virtualHostName: vh-1
      value: |
        match:
          prefix: /x
        route:
          cluster: demo-app_kuma-demo_svc_5000
  - route:
      operation: addBefore # it will be added before route named `catch-all` in virtual host named vh-1 
      match:
        name: catch-all # requires name so we know before what route we need to add it. Same with addAfter
        virtualHostName: vh-1
      value: |
        match:
          prefix: /x
        route:
          cluster: demo-app_kuma-demo_svc_5000
  - route:
      operation: remove 
      match:
        name: catch-all # removes route of 'catch-all' name. If not specified, then all routes in virtual host vh-1 are removed  
        virtualHostName: vh-1
  - route:
      operation: patch
      match:
        name: catch-all # optional: if not specified, patch will be executed on all routes in virtual host  
        virtualHostName: vh-1
        origin: outbound # optional: it enables a user to patch route in all virtual hosts of all outbound listeners.
      value: |
        timeout: 10s
```
