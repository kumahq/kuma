# MeshExternalService and ZoneEgress

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/10897

## Context and Problem Statement

MeshExternalService traffic can go out directly from the sidecar or through egress.

Configuration (like timeouts, retries, circuit breakers) of MeshExternalService can also be applied both on a sidecar and on an egress.

This MADR will go into the details of:
- limitations / features that occur / are possible when running the traffic directly from sidecar vs through egress
- each policy type and the placement to figure out what is the best option

## Decision Drivers

* flexible enough that we don't limit users in scenarios where it makes sense to configuration in both/one place
* restrictive enough so that it's hard for the user to "shoot themselves in the foot" (e.g. having squared no. of circuit breakers)

## Considered Options

Running traffic through:
* egress only
* sidecar only
* mix of sidecar and egress

Configuration placement:
* for outbound policies on the sidecar, for inbound policies on egress
* configurable where it makes sense on both sidecar and egress, predefined where only one makes sense
* only on the egress
* only on the sidecar

## Decision Outcome

Chosen option:
- egress only
- support only `to` policies
- don't require `MeshTrafficPermission` to communicate with `MeshExternalService` through egress

Chosen option for configuration placement:

| Policy                    | Implementation on                        |
|---------------------------|------------------------------------------|
| MeshAccessLog             | Sidecar (to)                             |
| MeshCircuitBreaker        | Egress (to)                              |
| MeshFaultInjection        | Out of scope (Egress targeting)          |
| MeshHealthCheck           | Egress (to)                              |
| MeshMetric                | Out of scope (single item)               |
| MeshProxyPatch            | Out of scope (single item)               |
| MeshRateLimit             | Out of scope (Egress targeting)          |
| MeshRetry                 | Sidecar (to)                             |
| MeshTimeout               | Sidecar (to)                             |
| MeshTrace                 | Out of scope (single item)               |
| MeshTrafficPermission     | Out of scope (Egress targeting)          |
| MeshLoadBalancingStrategy | Egress (to)                              |
| MeshTCPRoute              | Sidecar (to)                             |
| MeshHTTPRoute             | Sidecar (to)                             |

Targeting Egress is out of the scope of this issue and should be addressed as part of https://github.com/kumahq/kuma/issues/8417

## Pros and Cons of the Options

### Traffic flow

#### Running traffic through egress only

- there is a bit of operational cost to running egress
- it's a bit tricky to set up on universal
- we've got a bit of a gap in other functionalities on egress (configuration, observability)
- mTLS is required to run egress
- there is additional network hop
- with the introduction of MeshPassthrough the barrier for external traffic is lower and some cases can be covered by that

Conclusion: **it seems like we're not cutting any serious use case**.

#### Running traffic through sidecar only

- all the problems of running through egress are not applicable (operational cost, extra networking hop, scalability etc.)
- no single exit point for all the external traffic
- if we put credentials in the ExternalService then these credentials are accessible to the app owners if not delivered via generic_secret
- it's hard to emulate ExternalService because there is no "from" section (there is no destination that is inside the mesh)

#### Running traffic through sidecar and egress

- we can mix and match the pros/cons of both approaches
- it's harder to implement / explain how it's working

### Configuration placement

#### Current state and possible options

##### Table summary

The table below shows a summary of policies and if it makes sense to configure them on:
- sidecar
- egress
- on both

and if it makes sense for the user to pick where.

We assume in this table that it never makes sense to put `from` section of a policy on the sidecar because there is no inbound for it, only an outbound (`to` section).
Both `from` and `to` sections makes sense on the egress.

| Policy                    | On sidecar    | On egress              | On both | User configurable |
|---------------------------|---------------|------------------------|---------|-------------------|
| MeshAccessLog             | Yes           | Yes                    | Yes     | Yes               |
| MeshCircuitBreaker        | Yes           | Maybe                  | No      | No                |
| MeshFaultInjection        | Maybe         | Yes                    | No      | No                |
| MeshHealthCheck           | Yes           | Maybe                  | No      | No                |
| MeshMetric                | Yes           | Yes                    | Yes     | Maybe             |
| MeshProxyPatch            | Yes           | Yes                    | Yes     | Yes               |
| MeshRateLimit             | Probably not  | Yes                    | No      | No                |
| MeshRetry                 | Yes           | Maybe                  | No      | Maybe             |
| MeshTimeout               | Yes           | Maybe                  | No      | No                |
| MeshTrace                 | Yes           | Yes                    | Yes     | Maybe             |
| MeshTrafficPermission     | Maybe         | Yes                    | Maybe   | Maybe             |
| MeshLoadBalancingStrategy | Yes           | Yes                    | No      | Maybe             |
| MeshTCPRoute              | Yes           | Yes                    | No      | No                |
| MeshHTTPRoute             | Yes           | Yes                    | No      | No                |

In paragraphs below I go deeper into each policy.

##### MeshAccessLog

MeshAccessLog is not applied on egress at all, and we don't provide a way to configure it.
I don't see any reason not to do this though.
To define a specific mesh we can use `kuma.io/mesh` label.

##### MeshCircuitBreaker

###### Old ExternalService behaviour

Circuit Breaker is configured on the sidecar always.

###### New behaviour

If left on sidecar:
- a circuit breaker can break whole egress traffic even if only one of the destinations is unhealthy

If on egress
- services from one zone will act as an aggregate - if a CB triggers by service-1 then service-2 will also see that result immediately
  - this makes the configuration more sensitive to drastic changes
  - if a MES starts failing on one endpoint /example1 that is only consumed by service-1 it will also trigger for service-2 that might be happily consuming endpoint /example2
  - this could be mitigated by re-defining a MeshExternalService under a different name (a bit awkward)

The above drawbacks are acceptable. That's the cost of centralized access.

##### MeshFaultInjection

###### Old ExternalService behaviour

FaultInjection and MeshFaultInjection is configured for ExternalService on egress inbound.

###### New behaviour

If on sidecar:
- less traffic in the cluster (it won't even exit the pod)
- you won't be able to validate if your policies (MeshRetry, MeshTimeout, etc.) work correctly because they won't be triggered


##### MeshHealthCheck

###### Old ExternalService behaviour

HealthCheck is configured on the sidecar.

###### New behaviour

If on egress:
- Seems to be the same story as circuit breaker in terms of acting as an aggregate / sensitivity
- Doing it on the egress could cause less traffic to MES (assuming there is fewer instances of egress than services HC-ing)


##### MeshMetric

It's a new policy without a direct ancestor.
MeshMetric is not applied on egress at all, and we don't provide a way to configure it.
I don't see any reason not to allow this in the future.

##### MeshProxyPatch

###### Old ExternalService behaviour

It does not look like it's even possible to modify egress with MeshProxyPatch or ProxyTemplate so this needs work outside the scope of this MADR,
but it makes sense to be able to do this.

##### MeshRateLimit

###### Old ExternalService behaviour

Both RateLimit and MeshRateLimit are configured on the egress.

###### New behaviour

If on sidecar:
- limit is multiplied by the number of instances (which is bad)
- less traffic in the mesh (traffic does )


##### MeshRetry

###### Old ExternalService behaviour

Both Retry and MeshRetry are configured on the sidecar

###### New behaviour

If on egress:
- will behave the same as on sidecar
- we definitely shouldn't configure it both on sidecar and egress, that way we'll have squared retries

##### MeshTimeout

###### Old ExternalService behaviour

Both Timeout and MeshTimeout are configured on the sidecar.

###### New behaviour

If on egress:
- we would have to match the timeouts from the sidecar not to have a clash (that server times out before client)
- I don't see any advantages of having this on egress but nothing rules it out

##### MeshTrace

###### Old ExternalService behaviour

Both TrafficTrace and MeshTrace do not support egress.

###### New behaviour

I think it makes sense both on sidecar and egress.
If we're troubleshooting something it would be great to see spans from sidecar going to ExternalService.

##### MeshTrafficPermission

###### Old ExternalService behaviour

Both MeshTrafficPermission and TrafficPermission only works with egress.

###### New behaviour

It is already implemented on egress for MeshTrafficPermission.

If on the sidecar:
- could you circumvent this limitation by trying to directly call through egress?

##### MeshLoadBalancingStrategy

###### Old ExternalService behaviour

There is no old behaviour because there was no concept of Endpoints in External Service

###### New behaviour

Only the `loadBalancer` field makes sense in this regard.

If on sidecar:
- if you don't need/want egress it still makes sense to pick load balancing mechanism

If on egress:
- makes the same sense on as on sidecar

##### MeshTCPRoute and MeshHTTPRoute

###### Old ExternalService behaviour

It worked for both egress and sidecar.

###### New behaviour

I don't see why we would change that behaviour.

#### For outbound policies on the sidecar, for inbound policies on egress

It naturally fits our targetRef mechanism.
It does not make sense to apply inbound policies on the sidecar since there is no inbound for it.
Egress plays a role of a "sidecar" that is inside the mesh, and we can target it's inbound.
Configuring outbound policies on the sidecar avoids the problem with different clients influencing each other (see [MeshCircuitBreaker](#meshcircuitbreaker)).

##### How to configure it

We decide to implicitly configure sidecar and egress based on the policy.

###### `from` only policy

Policies:
- MeshTrafficPermission
- MeshRateLimit
- MeshFaultInjection

Not supported as a part of this task.

###### `To` policy on sidecar

Policies:
- MeshRetry
- MeshHTTPRoute
- MeshTCPRoute

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshRetry
metadata:
  name: backend
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshExternalService
        name: backend
      default:
        http:
          numRetries: 10
```

###### `To` policy on egress

The following policy we're going to configure on Egerss: 
- `MeshCircuitBreaker` - we don't want to cut out whole egress from traffic only faulty instance of MeshExternalService
- `MeshLoadBalancingStrategy` - (we only care about `default.loadBalancer` portion) we don't want to change the loadbalancing to Egress but to the underlying MeshExternalServices
- `MeshHealthCheck` - we don't want to be health checking Egress but the underlying MeshExternalServices

```yaml
type: MeshCircuitBreaker
mesh: default
name: circuit-breaker
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshExternalService
        name: backend
      default:
        outlierDetection:
          detectors:
            totalFailures:
              consecutive: 10
```

###### `To` and `From` policies

In this case, we will support only `to` policies.

Policies:
- MeshAccessLog
- MeshTimeout

On sidecar
```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshAccessLog
spec:
  targetRef:
    kind: Mesh
  to:
  - targetRef:
      kind: MeshExternalService
      name: backend
    default: ...
```

`From` section not supported as a part of this task.

#### Include default allow rbac configuration for MeshExternalService

When `mTLS` is enabled (which is required with `Egress`), we require users to explicitly create a `MeshTrafficPolicy` to enable traffic. In the case of `MeshExternalService` and the current lack of targeting `Egress`, we have decided to, by default, allow all traffic through `Egress` for `MeshExternalService`.

However, users will be able to disable this behavior on the `Mesh` object.
```proto
// Routing defines configuration for the routing in the mesh
message Routing {
  ...
  // Changes the default behavior to not include the allow RBAC filter.
  bool defaultForbidMeshExternalServiceAccess = 3;
}
```

#### Only configuring policies on the egress

This doesn't work since there are policies that make sense to be in both places and the ones like MeshTimeout that only have drawbacks on egress.

#### Only configuring policies on sidecar

This doesn't work because there are policies that make more sense on egress and in both places.

## Links

* https://github.com/kumahq/kuma/issues/5050
* https://github.com/kumahq/kuma/issues/8417
