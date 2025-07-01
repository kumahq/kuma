# {short title of solved problem and solution}

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

### Open Questions

1. If Mesh Operator DENIES `exact: spiffe://trust-domain.mesh/ns/default/sa/frontend`,
   can Service Owner ALLOW `exact: spiffe://trust-domain.mesh/ns/default/sa/frontend`?

2. If Mesh Operator DENIES `prefix: spiffe://trust-domain.mesh/`, 
   can Service Owner ALLOW `exact: spiffe://trust-domain.mesh/ns/default/sa/frontend`?

3. If Mesh Operator DENIES `exact: spiffe://trust-domain.mesh/ns/default/sa/frontend`, 
   can Service Owner ALLOW `prefix: spiffe://trust-domain.mesh/`?

4. If Mesh Operator DENIES `prefix: spiffe://trust-domain.mesh/`,
   can Service Owner ALLOW `prefix: spiffe://trust-domain.mesh/`?

### User Stories

#### Mesh Operator

1. I want all requests in the mesh to be denied by default (2.12)

2. I want to allow only services in namespace `observability` to access `metrics-collector`,
so they can scrape metrics without opening access mesh-wide. (2.12)

3. I want to block traffic from deprecated workloads identities,
so that insecure or legacy services can’t talk to anything in the mesh. (2.12)

4. I want to block access to `/admin` for all clients except for internal dashboard,
so that sensitive admin operations aren't exposed to other services.

5. I want to allow only services in namespace `observability` to access `/metrics` on all services,
so that we enforce least-privilege access to telemetry data.

#### Service Owner

1. I want to allow requests from `spiffe://trust-domain.mesh/ns/default/sa/frontend` even if
Mesh Operator denied requests from SPIFFE IDs with `spiffe://trust-domain.mesh/` prefix (2.12)

2. I want to allow GET requests to `/products` from any client, but restrict POST requests,
so that read operations are public but write operations are gated.

3. I want to block malicious/abusive client even if previously it was allowed,
so I can prevent my service from overloading until client's service team reacts to the incident. (2.12)

4. I want to temporarily allow a debug tool to access `/debug/pprof` and `/metrics`,
so that I can investigate a performance issue without opening up broader access.

### Design

```yaml
type: MeshTrafficPermission
mesh: default
spec:
  rules:
    matches:
      - spiffeId:
          type: Prefix
          value: "spiffe://trust-domain.mesh/"
    default:
      action: Deny
---
type: MeshTrafficPermission
mesh: default
spec:
  targetRef:
    kind: Dataplane
    labels:
      app: payments
  rules:
    matches:
      - spiffeId:
          type: Exact
          value: "spiffe://trust-domain.mesh/ns/default/sa/frontend"
    default:
      action: Allow
```

Across all rules specified on applicable Routes, precedence must be given to the match having:
* "Exact" spiffe id
* "Prefix" spiffe id
* "Exact" path match.
* "Prefix" path match with largest number of characters.
* Method match.
* Largest number of header matches.
* Largest number of query param matches.

### Other meshes

#### Istio

1. If there are any CUSTOM policies that match the request, evaluate and deny the request if the evaluation result is deny.
2. If there are any DENY policies that match the request, deny the request.
3. If there are no ALLOW policies for the workload, allow the request.
4. If any of the ALLOW policies match the request, allow the request.
5. Deny the request.

```yaml
apiVersion: security.istio.io/v1
kind: AuthorizationPolicy
metadata:
  name: httpbin
  namespace: foo
spec:
  action: ALLOW
  rules:
  - from:
    - source:
        principals: ["cluster.local/ns/default/sa/sleep"]
    - source:
        namespaces: ["test"]
    to:
    - operation:
        methods: ["GET"]
        paths: ["/info*"]
    - operation:
        methods: ["POST"]
        paths: ["/data"]
    when:
    - key: request.auth.claims[iss]
      values: ["https://accounts.google.com"]
```

#### Linkerd

```yaml
apiVersion: policy.linkerd.io/v1beta1
kind: HTTPRoute
metadata:
  name: authors-get-route
  namespace: booksapp
spec:
  parentRefs:
    - name: authors-server
      kind: Server
      group: policy.linkerd.io
  rules:
    - matches:
      - path:
          value: "/authors.json"
        method: GET
      - path:
          value: "/authors/"
          type: "PathPrefix"
        method: GET
---
apiVersion: policy.linkerd.io/v1alpha1
kind: AuthorizationPolicy
metadata:
  name: authors-get-policy
  namespace: booksapp
spec:
  targetRef:
    group: policy.linkerd.io
    kind: HTTPRoute
    name: authors-get-route
  requiredAuthenticationRefs:
    - name: authors-get-authn
      kind: MeshTLSAuthentication | # they allow putting ServiceAccount directly
      group: policy.linkerd.io
---
apiVersion: policy.linkerd.io/v1alpha1
kind: MeshTLSAuthentication
metadata:
  name: authors-get-authn
  namespace: booksapp
spec:
  identities:
    - "books.booksapp.serviceaccount.identity.linkerd.cluster.local"
    - "webapp.booksapp.serviceaccount.identity.linkerd.cluster.local"
```
