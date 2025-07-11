# SPIFFE ID matches for MeshTrafficPermission

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/12374

## Context and Problem Statement

The current MeshTrafficPermission implementation uses the client’s certificate, with the client’s data plane proxy tags encoded in its URI SANs.
As we're moving towards SPIFFE compliant certificates, we won't be able to use data plane proxy tags in MeshTrafficPermission policy anymore.

### User Stories

Version in parenthesis indicates which Kuma release is going to support the user story.

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

### Design

According to [MADR-078](078-special-mtp-algo.md) MeshTrafficPermission's algorithm and API is different from other inbound policies.
Current MADR suggests 2 options to implement the design.
Chosen option is "Option 1: MeshTrafficPermission is a single-item policy".

#### Option 1: MeshTrafficPermission is a single-item policy

Pros:
* Inspect API works out-of-the-box!
* No need to introduce a new type of policy
* `allowRules`, `denyRules` reuse `rules[].matches[]` schema
* No intermediate representation is required, `rbac_configurer.go` generates Envoy configuration directly from `conf`

Cons:
* Action in Envoy can't be correlated with a KRI of a single MeshTrafficPermission policy,
but that's a problem that exists for other single-item policies (i.e. [MeshPassthrough](https://github.com/kumahq/kuma/issues/13886)).

##### Schema

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
  targetRef: {}
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
But it shows that the concept is viable
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
             name: kri_mtp_
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

Same as singe-item policies (we call them proxy policy in the new Inspect API).

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

##### Extensibility with DPP labels selector

Today, MeshTrafficPermission allows specifying DPP tags of the workloads we want to allow or deny requests from.
This works because DPP tags are currently encoded as URI SANs in the workload certificate.
However, making workload certificates SPIFFE-compliant requires us to retain only a single URI SAN, which must be the SPIFFE ID.

Ability to grant access to workloads based on the labels rather than identity provides nicer UX.
That's why potentially, we're thinking on introducing workload selector alongside `spiffeId`:

```yaml
type: MeshTrafficPermission
mesh: default
name: by-service-owner
spec:
   targetRef: {}
   default:
      allow:
         - spiffeId:
              type: Exact
              value: "spiffe://trust-domain.mesh/ns/default/sa/api-gateway"
         - dataplaneLabels:
             app: frontend
```

The `dataplaneLabels` will be resolved on CP and replaced with a list of `spiffeId` corresponding to matched DPPs.

Though, without a label-enforcement mechanism, the feature isn't truly secure:
if clients can add or remove any labels on their own workloads, they can simply insert themselves into the allow list.

Additionally, what if several pods share the same `spiffeId`, but only one of them has `app: frontend`?
Service owner might either potentially "open the door" too wide or unintentionally cut someones access.

Without answering these security questions it seems like we won't be able to introduce DPP labels selector.

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

#### Option 2: MeshTrafficPermission is a special case of inbound policy

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
