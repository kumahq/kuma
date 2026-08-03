# MeshPassthrough

* Status: accepted

## Context and Problem Statement

During the design process of `MeshExternalService`, we discovered that allowing traffic to specific domains and wildcard domains (e.g., `*.aws.us-east-2.com`) is a great functionality. However, this capability doesn't fully fit within the `MeshExternalService` concept. `MeshExternalService` is a kind that allows adding some external services to work within your mesh, be targeted by policies, and be configured by them. Unfortunately, when a user needs to communicate with services that shouldn't be modified by Envoy, we discovered that we need a separate policy to handle them.

## Current solution

Currently, we don't support passthrough mode. Kuma creates DNS entries for real domains and returns a custom IP. This has a limitation: you cannot have two external services using the same domain. Additionally, we don't support wildcard domains.

## Considered Options

* MeshExternalService resource with passthrough mode support
* New MeshPassthrough policy

## Decision Outcome

New `MeshPassthrough` policy seems like cleaner, more flexible.

### Positive Consequences

* More flexibility: We can target specific proxies.
* Clarity: There is no confusion about which fields are involved in the policy configuration.
* Simplicity: Policies are cleaner and cannot be configured with other policies, unlike MeshExternalService.
* No confusion with Universal: There is no issue with non-transparent proxies.
* Can have separate clusters for each "match" and separate metrics for them.

### Negative Consequences

* Additional complexity: Introduction of another policy.
* Resource management: The need to manage two separate resources.

### New MeshPassthrough policy

We would like to introduce a new policy `MeshPassthrough` which allows exposing domains, IPs, CIDRs through specific sidecars.


```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshPassthrough
metadata:
  name: custom-passthrough
  namespace: kuma-system
  labels:
    kuma.io/mesh: default
spec:
  targetRef:
    kind: MeshSubset # Mesh, MeshSubset
    tags:
      chatgpt.io/access: "true"
  default:
    enabled: false
    appendMatch:
    - type: Domain
      value: api.chatgpt.com
      port: 443
      protocol: tls    
```

`MeshPassthrough` should allow targeting specific subset of proxies and apply configuration only on them. We should support following kinds: `Mesh` and`MeshSubset`.

* **enabled**: defines if sidecar should be in a passthrough mode and allow all outside traffic. If `true`, `matchAppend` is not used. Default: `false`.
* **matchAppend**: list of all domains/ips/cidrs supported through the selected sidecars. In case there is many polcies matching the same sidecar, lists are merged.
* **type**: type of the entry, one of `Domain`, `IP` or `CIDR`
* **value**: value for the entry
* **port**: port on which service can communicate
* **protocol**: defines a protocol of the communication. Possible values:
  * `tls`: should be used when TLS traffic is originated by the client application in the case the `kuma.io/protocol` would be tcp
  * `tcp`: WARNING: can't be used when `match.type == Domain` (at TCP level we are not able to disinguish domain, in this case it is going to hijack whole traffic on this port). This will be validated in the config.
  * `grpc`
  * `http`
  * `http2`

#### Mesh with a passthrough

If there is no `MeshPassthrough` policy targetting specific dataplane we are taking the default value from the `Mesh` object. If there is neither `Mesh` setting nor `MeshPassthrough` policy we are using the default for the control-plane which is `enabled` by default.

#### Security

It is advised that the MeshOperator is responsible for the `MeshPassthrough` policy. This policy can introduce traffic outside of the mesh or even the cluster, and the MeshOperator should be aware of this
If you want to secure access to `MeshPassthrough` to specific services, you must choose them manually. 
If you rely on tags in the top-level `targetRef` you might consider securing them by one of the:
1. Make sure that service owners can't freely set them (using something like [kyverno](https://kyverno.io/policies/other/allowed-label-changes/allowed-label-changes/), OPA)
2. Accept the risk of being able to "impersonate" a passthrough label and rely on auditing

#### Implementation

We can use [Matcher API](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/advanced/matching/matching_api.html#matching-api) which got out of a Alpha phase and is stable. We can use it to support different entries on one listener. Together with [passthrough](https://kuma.io/docs/2.7.x/networking/non-mesh-traffic/#outgoing) mode one the mesh, MeshOperator can disable all outgoing traffic except the one provided by `MeshPassthrough`. It is worth pointing out that when `passthrough.enabled` is set to `true` on the `Mesh`, `MeshPassthrough` policies have no effect because all outgoing traffic is allowed.

From the envoy configuration point of view, we are going to add filter chain matchers under `outbound:passthrough:ipv4"` listener. It's possible to use MeshTrace for tracing, but for the scope of this issue, we decided not to cover it now. We will create an issue and implement it if required.

#### Universal without transparent proxy

This policy won't apply without transparent proxy. We can provide in a [warning field](https://github.com/kumahq/kuma/blob/master/pkg/core/xds/matched_policies.go#L23) information why it wasn't applied.

#### ZoneEgress

##### Static IP and Domains

Static IP and Domain we can achive the same way we are doing it now. By setting SNI on the cluster and later matching in filter chain match.

##### CIDR and Wildcard domains

This case is a bit more problematic. We can use tunneling over [HTTP2 Connect](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/http/upgrades#tunneling-tcp-over-http). Client's application needs to resolve address and later sidecar sends the requests to egress. On the `egress` we need a cluster which supports routing based on the `original_dst_lb_config` and listener supporting `CONNECT`. In this case we cannot match based on SNI because we are sending this traffic over HTTP. [Example Evoy configuration](https://gist.github.com/lukidzi/34cd94528fe6a3d87dd2f2411ff39018).

### Other options

#### MeshExternalService resource with passthrough mode support

As described in `MeshExternalService`(MADR-xxx TBA), we can create a separate `type: Passthrough`. 

```yaml
kind: MeshExternalService
metadata:
  name: example
  namespace: kuma-system
  labels:
    kuma.io/mesh: default
    kuma.io/zone: east-1
spec:
  match:
  - type: Domain
    value: *.aws.us-east-2.com
    port: 80
    protocol: http
  - type: Domain
    value: httpbin.com
    port: 80
    protocol: http
  - type: CIDR
    value: 10.1.1.0/24
    port: 80
    protocol: http
  - type: IP
    value: 192.168.0.1
    port: 80
    protocol: http
  type: Passthrough
```

### Positive Consequences

* One policy handling both cases

### Negative Consequences

* Confusing way of targeting `MeshExternalServices` with policies but not affecting `Passthrough` clusters
* Match with many entries doesn't fit logically with `type: Managed`
* Not possible to apply only on specific sidecar
