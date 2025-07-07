# SPIFFE ID matches for MeshTrafficPermission

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/12374

## Context and Problem Statement

### Problems

#### Problem 1

Current MeshTrafficPermission relies on client's cert with client's DPP tags encoded as URI SANs.
As we're moving towards SPIFFE compliant certs we won't be able to use DPP tags in MeshTrafficPermission anymore.

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

#### Mesh Operator

1. I want all requests in the mesh to be denied by default (2.12)

2. I want to declare a group of identities as explicitly denied,
   so that Service Owner can't override or bypass that decision,
   ensuring enforcement of critical security boundaries across the mesh. (2.12) 

3. I want to allow all clients in the `observability` namespace to access all services by default,
   so that telemetry and monitoring tools function automatically,
   while still allowing Service Owners to explicitly opt out by applying deny policies. (2.12)

4. I want to allow all clients in the `observability` namespace to access the `/metrics` endpoint on all services,
   so that monitoring tools can collect metrics without requiring each Service Owner to configure access individually.

#### Service Owner

1. I want to grant access to my service to any client I choose,
   so that I can support integrations and collaboration with other teams,
   unless the Mesh Operator has explicitly denied that client,
   ensuring I remain in control while respecting mesh-wide security boundaries. (2.12)

2. I want to opt out of mesh-wide `observability` access,
   by denying requests from the `observability` namespace,
   so that my service’s sensitive endpoints remain private unless explicitly allowed. (2.12)

3. I want to block malicious/abusive client even if previously it was allowed,
   so I can prevent my service from overloading until client's service team reacts to the incident. (2.12)

4. I want to allow `GET` requests to my service from any client, but restrict `POST` requests to a group of identities,
   so that read operations are public but write operations are gated.

### Unified matching algorithm for inbound policies

Even though we expect all inbound policies to have similar structure

```yaml
spec:
  targetRef: {}
  rules:
     - matches: []
       default: {}
```

It's a real challenge to come up with unified algorithm that'd work both for MeshTrafficPermissions and other policies.
Technically it's possible to build a matching tree for Envoy that'd be similar for all inbound policies.
But, the size of the tree is getting big really fast.

Let's look at the example:

```yaml
rules:
   - matches:
        - spiffeId:
            type: Prefix
            value: spiffe://trust-domain.mesh/
     default: $conf1
   - matches:
        - path:
            type: Prefix
            value: /foo
     default: $conf2
   - matches:
        - method: GET
     default: $conf3
```

We’re matching on three independent dimensions:

1. Whether the SPIFFE ID starts with `spiffe://trust-domain.mesh/`
2. Whether the request path starts with `/foo`
3. Whether the request method is `GET`

Each condition has two possible outcomes: it either matches or doesn’t.
That gives us 2 * 2 * 2 = 8 possible combinations, and in the worst case, we need to evaluate each of those 8 leaves to determine which rules apply.

This is not just theoretical, it matters for policies like MeshTrafficPermission,
where any matched rule might contain a `Deny`, and even one `Deny` must take precedence over multiple `Allows`.

That means we can’t short-circuit or collapse rules based on structure alone, we must keep track of all matched rules per request.

Now imagine mesh with 100 unique SPIFFE IDs, and user adds matches on 5 paths, 5 headers and 5 queries,
we'll get 100 * 5 * 5 * 5 = 12500 leaves!
Sure, we can do pruning and significantly reduce the size of the tree that goes to Envoy, but CP still has to deal with huge trees.

**Intermediate conclusion:** if we're designing a unified matching algorithm for ALL inbound policies, we have to account for MeshTrafficPermission,
which requires evaluating all matching confs since any one of them might contain a `Deny`.
This leads to a combinatorial explosion in the match tree size.

But the truth is we don't actually need to evaluate all matching confs for other inbound policies.
In Gateway API when user creates

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: http-app-1
spec:
  parentRefs:
  - name: my-gateway
  hostnames:
  - "foo.com"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /bar
    filters:
       - type: RequestHeaderModifier
         requestHeaderModifier:
            add:
               - name: my-foo-header
                 value: foo
  - matches:
    - headers:
      - type: Exact
        name: magic
        value: bar
    filters:
       - type: RequestHeaderModifier
         requestHeaderModifier:
            add:
               - name: my-bar-header
                 value: bar
```

they do **not** expect request with path `/bar` and header `magic: bar` to get both `my-foo-header` and `my-bar-header`.
Rules are evaluated in order, `path` has more priority than `header` so request is getting only `my-foo-header`.

This type of matching is reasonable for every policy except MeshTrafficPermission and doesn't cause combinatory explosion.

**Conclusion:** MeshTrafficPermission has fundamentally different requirements compared to other inbound policies.
While both can be expressed using Envoy’s Matching API, the evaluation semantics diverge:
MeshTrafficPermission requires evaluating all matching rules to enforce Deny precedence,
whereas other policies typically rely on first-match-wins semantics.
Because of this, we need a separate matching algorithm for MeshTrafficPermission
to ensure correctness without causing unnecessary growth in match tree complexity for other policies.
