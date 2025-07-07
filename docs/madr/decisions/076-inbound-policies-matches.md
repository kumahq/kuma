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
