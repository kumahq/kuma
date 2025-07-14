# SPIFFE ID matches for MeshTrafficPermission

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/12374

## Context and Problem Statement

### Problems

#### Problem 1

The current MeshTrafficPermission implementation uses the client’s certificate, with the client’s data plane proxy tags encoded in its URI SANs.
As we're moving towards SPIFFE compliant certificates, we won't be able to use data plane proxy tags in MeshTrafficPermission policy anymore.

#### Problem 2

In the [Inbound Policies MADR](https://docs.google.com/document/d/1tdIOVVYObHbGKX1AFhbPH3ZQKYvpQG-2s6ShHSECaXM) we said:

> #### Merging behaviour
> Same as MeshHTTPRoute merging:
> 1. Pick policies that select the DPP with top-level targetRef
> 2. Sort these policies by top-level targetRef
> 3. Concatenate all “rules” arrays from these policies
> 4. Merge rules items with the same “hash(rules[i].matches)”
     > the new rule position is the highest of two merged rules
> 5. Stable sort rules based on the Extended GAPI matches order
     > we need to split individual matches into separate rules
     But later I discovered [MeshHTTPRoute merging behaviour is not ideal](https://github.com/kumahq/kuma/issues/13440):

1. Targeting routes by policies results in surprising behaviour

   > * `timeout-1` applies to `route-1`
   > * `timeout-2` applies to `route-2`
   >
   > Since the merged configuration is attributed only to `route-2`, we would apply only `timeout-2`, ignoring `timeout-1`.
2. Merging is not aligned with the Gateway API approach that [states](https://gateway-api.sigs.k8s.io/api-types/httproute/#merging):
   > Multiple HTTPRoutes can be attached to a single Gateway resource. Importantly, only one Route rule may match each request.

### User Stories

Version in parentheses indicates which Kuma release is going to support the user story.

#### Mesh Operator

1. I want all requests in the mesh to be denied by default (2.12)

2. I want to be able to declare a group of identities as explicitly denied,
   so that Service Owner can't override or bypass that decision,
   ensuring enforcement of critical security boundaries across the mesh. (2.12)

3. I want to be able to allow all clients in the `observability` namespace to access all services by default,
   so that telemetry and monitoring tools function automatically,
   while still allowing Service Owners to explicitly opt out by applying deny policies. (2.12)

4. I want to be able to allow all clients in the `observability` namespace to access the `/metrics` endpoint on all services,
   so that monitoring tools can collect metrics without requiring each Service Owner to configure access individually.

#### Service Owner

1. I want to be able to grant access to my service to any client I choose,
   so that I can support integrations and collaboration with other teams,
   unless the Mesh Operator has explicitly denied that client,
   ensuring I remain in control while respecting mesh-wide security boundaries. (2.12)

2. I want to be able to opt out of mesh-wide `observability` access,
   by denying requests from the `observability` namespace,
   so that my service’s sensitive endpoints remain private unless explicitly allowed. (2.12)

3. I want to be able to block malicious/abusive client even if previously it was allowed,
   so I can prevent my service from overloading until the client's service team responds to the incident (2.12)

4. I want to be able to allow `GET` requests to my service from any client, but restrict `POST` requests to a group of identities,
   so that read operations are public but write operations are gated.

5. I want to apply a MeshTrafficPermission to just one of my multiple inbounds

### Design

According to [MADR-078](078-special-mtp-algo.md) MeshTrafficPermission's algorithm and API is different from other inbound policies.
Current MADR suggests 2 options to implement the design.
Chosen option is "Option 1: MeshTrafficPermission is a single-item policy".

#### Option 1: MeshTrafficPermission is an inbound policy without 'matches'

Pros:
* Inspect API works with minimal adjustments
* No need to introduce a new type of policy
* `allow`, `deny` lists reuse `rules[].matches[]` schema
* MeshTrafficPermission policy plugin works directly with `conf` which is a merging product of all matched MeshTrafficPermissions.
  No additional data structure (so-called "intermediate representation") is needed.
* `Action` in Envoy can be named as KRI of a MeshTrafficPermission policy that defined the action

Cons:
* Looks verbose without syntactic sugar

#### Schema

If we treat MeshTrafficPermission as an inbound policy without `matches` support,
it naturally follows that its `rules` list can contain only a single item with `default`.

```yaml
type: MeshTrafficPermission
mesh: default
name: by-mesh-operator
spec:
  targetRef: {}
  rules:
     - default:
         deny:
            - spiffeId:
                type: Exact
                value: "spiffe://trust-domain.mesh/ns/default/sa/frontend"
---
type: MeshTrafficPermission
mesh: default
name: by-service-owner
spec:
  targetRef:
     kind: Dataplane
     labels:
        app: backend
     sectionName: http-port
  rules:
    - default:
        deny:
           - spiffeId:
               type: Exact
               value: "spiffe://trust-domain.mesh/ns/default/sa/api-gateway"
        allowWithShadowDeny:
           - spiffeId:
               type: Prefix
               value: "spiffe://trust-domain.mesh/ns/legacy"
        allow:
           - spiffeId:
               type: Prefix
               value: "spiffe://trust-domain.mesh/"
```

To spare the readability we'll add a syntactic sugar for a single rule without `matches`.
For example, the following inbound policies are identical:

```yaml
type: InboundPolicy
mesh: default
name: typical-inbound-policy
spec:
  targetRef: {}
  rules:
     - default: {} # matches all requests
---
type: InboundPolicy
mesh: default
name: same-but-with-syntactic-sugar
spec:
  targetRef: {}
  default: {} # matches all requests
```

With syntactic sugar MeshTrafficPermission will look exactly as we've planned in [Option 2](#option-2-meshtrafficpermission-is-a-single-item-policy)

```yaml
type: MeshTrafficPermission
mesh: default
name: by-service-owner
spec:
  targetRef:
     kind: Dataplane
     labels:
       app: backend
     sectionName: http-port
  default:
    deny:
      - spiffeId:
           type: Exact
           value: "spiffe://trust-domain.mesh/ns/default/sa/api-gateway"
    allowWithShadowDeny:
       - spiffeId:
           type: Prefix
           value: "spiffe://trust-domain.mesh/ns/legacy"
    allow:
       - spiffeId:
            type: Prefix
            value: "spiffe://trust-domain.mesh/"
```

Evaluation rules:
1. If the incoming request matches at least one matcher in `deny` list – the result is DENY.
2. If the incoming request matches at least one matcher in either `allow` or `allowWithShadowDeny` list – the result is ALLOW.
3. If the incoming request doesn't match anything – the result is DENY.

##### Algorithm

1. Collect all MeshTrafficPermissions that target the inbound
2. Concat all `rules` arrays (we don't want to merge `default` confs to solve the [Problem 2](#problem-2))

```yaml
rules:
   - default: # from 'by-mesh-operator' MTP
       deny:
          - spiffeId:
              type: Exact
              value: "spiffe://trust-domain.mesh/ns/default/sa/frontend"
   - default: # from 'by-service-owner' MTP
        deny:
           - spiffeId:
                type: Exact
                value: "spiffe://trust-domain.mesh/ns/default/sa/api-gateway"
        allowWithShadowDeny:
           - spiffeId:
                type: Prefix
                value: "spiffe://trust-domain.mesh/ns/legacy"
        allow:
           - spiffeId:
                type: Prefix
                value: "spiffe://trust-domain.mesh/"
```

3. Starting this step MTP plugin is in charge, it receives the struct from step 2 and generates Envoy config.
For each item in `rules` generate appropriate `matcher_list` and then concat `matcher_list` from old rules prioritizing `Deny`:

```yaml
extensions.filters.network.rbac.v3.RBAC:
  rules: []
  shadow_rules: []
  matcher:
    matcher_list:
       - predicate:
           or_matcher:
              - spiffeId exact spiffe://trust-domain.mesh/ns/default/sa/api-gateway
           on_match:
             action: Deny
             name: kri_mtp_default___by-mesh-operator_
       - predicate:
            or_matcher:
               - spiffeId exact spiffe://trust-domain.mesh/ns/default/sa/api-gateway
            on_match:
               action: Deny
               name: kri_mtp_default___by-service-owner_
       - predicate:
            or_matcher:
               - spiffeId prefix spiffe://trust-domain.mesh/
               - spiffeId prefix spiffe://trust-domain.mesh/ns/legacy # for 'allowWithShadowDeny'
            on_match: Allow
            name: kri_mtp_default___by-service-owner_
    no_match: Deny
  shadow_matcher: # not enforced, just logged
    matcher_list:
     - predicate:
          or_matcher:
             - spiffeId prefix spiffe://trust-domain.mesh/ns/legacy
          on_match: Deny
          name: kri_mtp_default___by-service-owner_
    no_match: Deny
```

##### Inspect API

Inspect API for inbound rules:

```
GET :5681/meshes/{mesh}/dataplanes/{name}/_inbounds/{inbound-kri}/_policies
```
```yaml
policies:
  - kind: MeshTrafficPermission
    rules:
      - conf:
          deny:
             - spiffeId:
                 type: Exact
                 value: "spiffe://trust-domain.mesh/ns/default/sa/frontend"
        origin: kri_mtp_default___by-mesh-operator_ # current Inspect API missing 'origin' inside each 'rule'
      - conf:
          deny:
             - spiffeId:
                  type: Exact
                  value: "spiffe://trust-domain.mesh/ns/default/sa/api-gateway"
          allowWithShadowDeny:
             - spiffeId:
                  type: Prefix
                  value: "spiffe://trust-domain.mesh/ns/legacy"
          allow:
             - spiffeId:
                  type: Prefix
                  value: "spiffe://trust-domain.mesh/"
        origin: kri_mtp_default___by-service-owner_
    origins:
       - kri: kri_mtp_...
       - kri: kri_mtp_...
```

Currently, Inspect API is missing `policies[].rules[].origin`.
As we don't want to merge `rules` to fix the [Problem 2](#problem-2) we can unambiguously match what policy KRI contributed what rule.
This change needs to be covered in a separate MADR.

##### Verify user stories

###### Mesh Operator

1. I want all requests in the mesh to be denied by default (2.12)

No MeshTrafficPermission policies means requests are denied by default.

2. I want to declare a group of identities as explicitly denied,
   so that Service Owner can't override or bypass that decision,
   ensuring enforcement of critical security boundaries across the mesh. (2.12)

```yaml
type: MeshTrafficPermission
mesh: default
name: by-mesh-operator
spec:
   default:
      deny:
         - spiffeId:
              type: Exact
              value: "spiffe://trust-domain.mesh/ns/default/sa/api-gateway"
         - spiffeId:
              type: Exact
              value: "spiffe://trust-domain.mesh/ns/default/sa/legacy-workload"
         - spiffeId:
              type: Prefix
              value: "spiffe://legacy.mesh/"
```

3. I want to allow all clients in the `observability` namespace to access all services by default,
   so that telemetry and monitoring tools function automatically,
   while still allowing Service Owners to explicitly opt out by applying deny policies. (2.12)

```yaml
type: MeshTrafficPermission
mesh: default
name: by-mesh-operator
spec:
   default:
      allow:
         - spiffeId:
              type: Prefix
              value: "spiffe://trust-domain.mesh/ns/observability"
```

4. I want to allow all clients in the `observability` namespace to access the `/metrics` endpoint on all services,
   so that monitoring tools can collect metrics without requiring each Service Owner to configure access individually.

```yaml
type: MeshTrafficPermission
mesh: default
name: by-mesh-operator
spec:
   default:
      allow:
         - spiffeId:
              type: Prefix
              value: "spiffe://trust-domain.mesh/ns/observability"
           path:
             type: Prefix
             value: "/metrics"
```

###### Service Owner

1. I want to grant access to my service to any client I choose,
   so that I can support integrations and collaboration with other teams,
   unless the Mesh Operator has explicitly denied that client,
   ensuring I remain in control while respecting mesh-wide security boundaries. (2.12)

```yaml
type: MeshTrafficPermission
mesh: default
name: by-backend-owner
spec:
   targetRef:
      kind: Dataplane
      labels:
         app: backend
   default:
      allow:
         - spiffeId:
              type: Prefix
              value: "spiffe://trust-domain.mesh/"
```

2. I want to opt out of mesh-wide `observability` access,
   by denying requests from the `observability` namespace,
   so that my service’s sensitive endpoints remain private unless explicitly allowed. (2.12)

```yaml
type: MeshTrafficPermission
mesh: default
name: by-backend-owner
spec:
   targetRef:
      kind: Dataplane
      labels:
         app: backend
   default:
      deny:
         - spiffeId:
              type: Prefix
              value: "spiffe://trust-domain.mesh/ns/observability"
```

3. I want to block malicious/abusive client even if previously it was allowed,
   so I can prevent my service from overloading until client's service team reacts to the incident. (2.12)

```yaml
type: MeshTrafficPermission
mesh: default
name: by-backend-owner
spec:
   targetRef:
      kind: Dataplane
      labels:
         app: backend
   default:
      deny:
         - spiffeId:
              type: Exact
              value: "spiffe://trust-domain.mesh/ns/default/sa/malicious"
```

4. I want to allow `GET` requests to my service from any client, but restrict `POST` requests to a group of identities,
   so that read operations are public but write operations are gated.

```yaml
type: MeshTrafficPermission
mesh: default
name: by-backend-owner
spec:
   targetRef:
      kind: Dataplane
      labels:
         app: backend
   default:
      allow:
         - method: GET
         - method: POST
           spiffeId:
              type: Exact
              value: "spiffe://trust-domain.mesh/ns/default/sa/writer-1"
         - method: POST
           spiffeId:
              type: Exact
              value: "spiffe://trust-domain.mesh/ns/default/sa/writer-2"
         - method: POST
           spiffeId:
              type: Prefix
              value: "spiffe://trust-domain.mesh/ns/writers"
```

5. I want to apply a MeshTrafficPermission to just one of my multiple inbounds

```yaml
type: MeshTrafficPermission
mesh: default
name: by-backend-owner
spec:
   targetRef:
      kind: Dataplane
      labels:
         app: backend
      sectionName: http-port # applies only to the inbound with 'http-port' name
   default: {}
```

##### Extensibility to other authentication methods

Current `spiffeId` authentication method is translated to the following Matching API predicate:

```json
{
  "input": {
    "name": "envoy.matching.inputs.uri_san",
    "typed_config": {
      "@type": "type.googleapis.com/envoy.extensions.matching.common_inputs.ssl.v3.UriSanInput"
    }
  },
  "value_match": {
    "exact": "spiffe://trust-domain.mesh/ns/default/sa/default"
  }
}
```

As long as there is a way to express authentication method as [SinglePredicate](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/common/matcher/v3/matcher.proto#envoy-v3-api-msg-config-common-matcher-v3-matcher-matcherlist-predicate-singlepredicate)
there is no issues with extensibility.

Currently supported `input` extensions could be found on [Envoy Matching API](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/advanced/matching/matching_api) page.

#### Option 2: MeshTrafficPermission is a single-item policy

Pros:
* Inspect API works out-of-the-box!
* No need to introduce a new type of policy
* `allowRules`, `denyRules` reuse `rules[].matches[]` schema
* MeshTrafficPermission policy plugin works directly with `conf` which is a merging product of all matched MeshTrafficPermissions.
No additional data structure (so-called "intermediate representation") is needed.

Cons:
* Action in Envoy can't be correlated with a KRI of a single MeshTrafficPermission policy,
but that's a problem that exists for other single-item policies (i.e. [MeshPassthrough](https://github.com/kumahq/kuma/issues/13886)).
* Although the top-level targetRef supports `sectionName` and allows configuring a specific inbound,
it's not reflected in the Inspect API, since it reuses the proxy policy endpoint, which doesn't expose per-inbound configuration

#### Schema

The idea is to make MeshTrafficPermission a simple single-item policy:

```yaml
type: MeshTrafficPermission
mesh: default
name: by-mesh-operator
spec:
   targetRef: {}
   default:
      deny:
         - spiffeId:
              type: Exact
              value: "spiffe://trust-domain.mesh/ns/default/sa/frontend"
---
type: MeshTrafficPermission
mesh: default
name: by-service-owner
spec:
  targetRef:
     kind: Dataplane
     labels:
       app: backend
     sectionName: http-port
  default:
    deny:
      - spiffeId:
           type: Exact
           value: "spiffe://trust-domain.mesh/ns/default/sa/api-gateway"
    allowWithShadowDeny:
       - spiffeId:
           type: Prefix
           value: "spiffe://trust-domain.mesh/ns/legacy"
    allow:
       - spiffeId:
            type: Prefix
            value: "spiffe://trust-domain.mesh/"
```

Evaluation rules:
1. If the incoming request matches at least one matcher in `deny` list – the result is DENY.
2. If the incoming request matches at least one matcher in either `allow` or `allowWithShadowDeny` list – the result is ALLOW.
3. If the incoming request doesn't match anything – the result is DENY.

##### Merging

Normally in Kuma policies mergeable arrays override each other, for example:

```
name: policy-1
spec:
  default:
    myArray: [1,2,3]
---
name: policy-1
spec:
  default:
    myArray: [4]
```

results in `spec.default.myArray: [4]`.

Switching array merging behaviour from overriding to concatenation normally requires prefixing fields with `append`:

```
name: policy-1
spec:
  default:
    appendArray: [1,2,3]
---
name: policy-1
spec:
  default:
    appendArray: [4]
```

results in `spec.default.appendArray: [1,2,3,4]`.

Instead of naming fields `appendDeny` and `appendAllow` we need to add a new value `concat` for the existing `policyMerge` struct tag:

```go
type Conf struct {
    Allow *[]Matcher `json:"allow,omitempty" policyMerge:"concat"`
}
```

When used on fields prefixed with `append` the struct tag `concat` won't have any additional effect.

Kubernetes has similar approach in their API with [merge-strategy](https://kubernetes.io/docs/reference/using-api/server-side-apply/#merge-strategy).
We don't have to adopt their exact annotations as we won't gain anything from it.
But it shows that the concept is viable,
and potentially we can update `MeshPassthrough`, `MeshProxyPatch` and `MeshOPA` to use struct tags instead of `append` prefix.

##### Algorithm

1. Collect all MeshTrafficPermissions that target the inbound
2. Concat all `deny`, `allowWithShadowDeny` and `allow` lists
3. Convert to the following Matching API structure

```yaml
extensions.filters.network.rbac.v3.RBAC:
  rules: []
  shadow_rules: []
  matcher:
    matcher_list:
       - predicate:
           or_matcher:
              - spiffeId exact spiffe://trust-domain.mesh/ns/default/sa/frontend
              - spiffeId exact spiffe://trust-domain.mesh/ns/default/sa/api-gateway
           on_match:
             action: Deny
       - predicate:
            or_matcher:
               - spiffeId prefix spiffe://trust-domain.mesh/
               - spiffeId prefix spiffe://trust-domain.mesh/ns/legacy # for 'allowWithShadowDeny'
            on_match: Allow
    no_match: Deny
  shadow_matcher: # not enforced, just logged
    matcher_list:
     - predicate:
          or_matcher:
             - spiffeId prefix spiffe://trust-domain.mesh/ns/legacy
          on_match: Deny
    no_match: Deny
```

##### Inspect API

Same as single-item policies (we call them proxy policy in the new Inspect API).

```
GET :5681/meshes/{mesh}/dataplanes/{name}/_policies
```
```yaml
policies:
   - kind: MeshTrafficPermission
     conf:
       deny: [...]
       allowWithShadowDeny: [...]
       allow: [...]
     origins:
        - kri: kri_mtp_...
        - kri: kri_mtp_...
```

##### Verify user stories

###### Mesh Operator

1. I want all requests in the mesh to be denied by default (2.12)

No MeshTrafficPermission policies means requests are denied by default.

2. I want to declare a group of identities as explicitly denied,
   so that Service Owner can't override or bypass that decision,
   ensuring enforcement of critical security boundaries across the mesh. (2.12)

```yaml
type: MeshTrafficPermission
mesh: default
name: by-mesh-operator
spec:
   default:
      deny:
         - spiffeId:
              type: Exact
              value: "spiffe://trust-domain.mesh/ns/default/sa/api-gateway"
         - spiffeId:
              type: Exact
              value: "spiffe://trust-domain.mesh/ns/default/sa/legacy-workload"
         - spiffeId:
              type: Prefix
              value: "spiffe://legacy.mesh/"
```

3. I want to allow all clients in the `observability` namespace to access all services by default,
   so that telemetry and monitoring tools function automatically,
   while still allowing Service Owners to explicitly opt out by applying deny policies. (2.12)

```yaml
type: MeshTrafficPermission
mesh: default
name: by-mesh-operator
spec:
   default:
      allow:
         - spiffeId:
              type: Prefix
              value: "spiffe://trust-domain.mesh/ns/observability"
```

4. I want to allow all clients in the `observability` namespace to access the `/metrics` endpoint on all services,
   so that monitoring tools can collect metrics without requiring each Service Owner to configure access individually.

```yaml
type: MeshTrafficPermission
mesh: default
name: by-mesh-operator
spec:
   default:
      allow:
         - spiffeId:
              type: Prefix
              value: "spiffe://trust-domain.mesh/ns/observability"
           path:
             type: Prefix
             value: "/metrics"
```

###### Service Owner

1. I want to grant access to my service to any client I choose,
   so that I can support integrations and collaboration with other teams,
   unless the Mesh Operator has explicitly denied that client,
   ensuring I remain in control while respecting mesh-wide security boundaries. (2.12)

```yaml
type: MeshTrafficPermission
mesh: default
name: by-backend-owner
spec:
   targetRef:
      kind: Dataplane
      labels:
         app: backend
   default:
      allow:
         - spiffeId:
              type: Prefix
              value: "spiffe://trust-domain.mesh/"
```

2. I want to opt out of mesh-wide `observability` access,
   by denying requests from the `observability` namespace,
   so that my service’s sensitive endpoints remain private unless explicitly allowed. (2.12)

```yaml
type: MeshTrafficPermission
mesh: default
name: by-backend-owner
spec:
   targetRef:
      kind: Dataplane
      labels:
         app: backend
   default:
      deny:
         - spiffeId:
              type: Prefix
              value: "spiffe://trust-domain.mesh/ns/observability"
```

3. I want to block malicious/abusive client even if previously it was allowed,
   so I can prevent my service from overloading until client's service team reacts to the incident. (2.12)

```yaml
type: MeshTrafficPermission
mesh: default
name: by-backend-owner
spec:
   targetRef:
      kind: Dataplane
      labels:
         app: backend
   default:
      deny:
         - spiffeId:
              type: Exact
              value: "spiffe://trust-domain.mesh/ns/default/sa/malicious"
```

4. I want to allow `GET` requests to my service from any client, but restrict `POST` requests to a group of identities,
   so that read operations are public but write operations are gated.

```yaml
type: MeshTrafficPermission
mesh: default
name: by-backend-owner
spec:
   targetRef:
      kind: Dataplane
      labels:
         app: backend
   default:
      allow:
         - method: GET
         - method: POST
           spiffeId:
              type: Exact
              value: "spiffe://trust-domain.mesh/ns/default/sa/writer-1"
         - method: POST
           spiffeId:
              type: Exact
              value: "spiffe://trust-domain.mesh/ns/default/sa/writer-2"
         - method: POST
           spiffeId:
              type: Prefix
              value: "spiffe://trust-domain.mesh/ns/writers"
```

5. I want to apply a MeshTrafficPermission to just one of my multiple inbounds

Even though it's possible to define such policy, it won't be exposed in Inspect API:

```yaml
type: MeshTrafficPermission
mesh: default
name: by-backend-owner
spec:
   targetRef:
      kind: Dataplane
      labels:
         app: backend
      sectionName: http-port
   default: {}
```

#### Option 3: MeshTrafficPermission is a special case of inbound policy with `conditions` instead of `matches`

Pros:
* the concept is extendable to other policies if we'll need similar semantic
* structurally looks similar to other inbound policies with `matches`

Cons:
* it works well only when the number of possible `confs` is limited
* requires IR and as a result Inspect API is not that straightforward
* the algorithm is more complex, still might produce suboptimal envoy structs due to the presence of `AllowWithShadowDeny`

##### Schema

While inbound policies have the following schema

```yaml
type: MeshRateLimit
mesh: default
spec:
   rules:
      - matches: # sort by priorities and first-matched wins (same as Gateway API)
           - spiffeId:
                type: Prefix
                value: "spiffe://trust-domain.mesh/"
             method: GET
        default: $conf
```

MeshTrafficPermission uses `conditions` instead of `matches` to emphasize ALL of them need to be evaluated:

```yaml
type: MeshTrafficPermission
mesh: default
spec:
   rules:
      - conditions:
           - spiffeId:
                type: Exact
                value: "spiffe://trust-domain.mesh/ns/default/sa/frontend"
        default:
           action: Deny
---
type: MeshTrafficPermission
mesh: default
spec:
  rules:
    - conditions:
         - spiffeId:
              type: Exact
              value: "spiffe://trust-domain.mesh/ns/default/sa/api-gateway"
      default:
        action: Deny
    - conditions:
         - spiffeId:
              type: Prefix
              value: "spiffe://trust-domain.mesh/"
      default:
        action: Allow
    - conditions:
         - spiffeId:
              type: Prefix
              value: "spiffe://trust-domain.mesh/ns/legacy"
      default:
         action: AllowWithShadowDeny
```

##### Algorithm

1. Collect all MeshTrafficPermissions that target the inbound
2. Concat all `rules` lists
3. Without peeking into confs, group `conditions` by the same confs (either compare conf's hashes or marshalled content).
With MTP we're getting at most 3 unique confs – `action: Deny`, `action: AllowWithShadowDeny` and `action: Allow`

```yaml
rules:
  - conditions: # condition_1
       - spiffeId:
            type: Exact
            value: "spiffe://trust-domain.mesh/ns/default/sa/api-gateway"
       - spiffeId:
            type: Exact
            value: "spiffe://trust-domain.mesh/ns/default/sa/frontend"
    default:
      action: Deny # conf1
  - conditions: # condition_2
       - spiffeId:
            type: Prefix
            value: "spiffe://trust-domain.mesh/"
    default:
      action: Allow # conf2
  - conditions: # condition_3
       - spiffeId:
            type: Prefix
            value: "spiffe://trust-domain.mesh/ns/legacy"
    default:
       action: AllowWithShadowDeny # conf3
```

4. Build the intermediate representation that'll be used both in Inspect API and Envoy generator.
Roughly it's going to look like:

```yaml
condition_1 ?
  - true
    condition_2 ?
      - true
        condition_3 ?
          - true -> merge(conf1,conf2,conf3) # Deny
          - false -> merge(conf1,conf2) # Deny
      - false
        condition_3 ?
        - true -> merge(conf1,conf3) # Deny
        - false -> merge(conf1) # Deny
  - false
    condition_2 ?
      - true
        condition_3 ?
          - true -> merge(conf2,conf3) # AllowWithShadowDeny
          - false -> merge(conf2) # Allow
      - false
        condition_3 ?
          - true -> merge(conf3) # AllowWithShadowDeny
          - false -> merge() # Deny
```
But final schema for this is yet to be written.

5. Prune the IR structure

```yaml
condition_1 ?
  - true -> Deny
  - false
    condition_2 ?
      - true
        condition_3 ?
          - true -> merge(conf2,conf3) # AllowWithShadowDeny
          - false -> merge(conf2) # Allow
      - false
        condition_3 ?
          - true -> merge(conf3) # AllowWithShadowDeny
          - false -> merge() # Deny
```

6. Based on IR build envoy config

```yaml
extensions.filters.network.rbac.v3.RBAC:
  matcher:
    matcher_list:
       - predicate: condition_1
         on_match:
           action: Deny
    no_match:
      matcher_list:
         - predicate: condition_2
           on_match:
             matcher_list:
                - predicate: condition_3
                  on_match: Allow
             no_match: Allow
      no_match:
        matcher_list:
           - predicate: condition_3
             on_match: Allow
        no_match: Deny
  shadow_matcher:
     matcher_list:
        - predicate: condition_1
          on_match:
             action: Deny
     no_match:
        matcher_list:
           - predicate: condition_2
             on_match:
                matcher_list:
                   - predicate: condition_3
                     on_match: Deny
                no_match: Allow
        no_match:
           matcher_list:
              - predicate: condition_3
                on_match: Deny
           no_match: Deny
```
