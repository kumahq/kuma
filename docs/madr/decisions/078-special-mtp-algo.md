# MeshTrafficPermission Is a Special Case of Inbound Policies and Requires a Custom Matching Algorithm

* Status: accepted

## Context and Problem Statement

In the [Inbound Policies MADR](https://docs.google.com/document/d/1tdIOVVYObHbGKX1AFhbPH3ZQKYvpQG-2s6ShHSECaXM) we said 
all inbound policies are going to have similar structure:

```yaml
spec:
  targetRef: {}
  rules:
     - matches: []
       default: {}
```

and we'll create a unified matching algorithm that'll work for all inbound policies and potentially use Envoy Matching API.

But turns out, it's a real challenge to come up with unified algorithm that'd work both for `MeshTrafficPermission` and other policies.

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

This is not just theoretical, it matters for policies like `MeshTrafficPermission`,
where any matched rule might contain a `Deny`, and even one `Deny` must take precedence over multiple `Allows`.

That means we can’t short-circuit or collapse rules based on structure alone, we must keep track of all matched rules per request.

Now imagine mesh with 100 unique SPIFFE IDs, and user adds matches on 5 paths, 5 headers and 5 queries,
we'll get 100 * 5 * 5 * 5 = 12500 leaves!
Sure, we can do pruning and significantly reduce the size of the tree that goes to Envoy, but CP still has to deal with huge trees.

**Intermediate conclusion:** if we're designing a unified matching algorithm for ALL inbound policies, we have to account for `MeshTrafficPermission`,
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

This type of matching is reasonable for every policy except `MeshTrafficPermission` and doesn't cause combinatory explosion.

## Decision 

`MeshTrafficPermission` has fundamentally different requirements compared to other inbound policies.

While both can be implemented using Envoy’s Matching API, their evaluation semantics differ significantly:
`MeshTrafficPermission` must evaluate all matching rules to enforce `Deny` precedence, whereas other policies typically follow first-match-wins behavior.

Because of this, we need a custom matching algorithm for `MeshTrafficPermission` to avoid building a unified match tree with poor growth characteristics, 
solely for the sake of consistency across inbound policies.

In addition, this custom matching logic justifies a distinct API design for `MeshTrafficPermission`, 
so we don't have to introduce ad-hoc exceptions in shared APIs or documentation.
