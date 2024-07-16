# MeshExternalService policies and ZoneEgress

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/10897

## Context and Problem Statement

Configuration (like timeouts, retries, circuit breakers) of MeshExternalService can be applied both on a sidecar and on an egress.

This MADR will go into the details of each policy type and the placement to figure out what is the best option.

## Decision Drivers <!-- optional -->

* {driver 1, e.g., a force, facing concern, …}
* {driver 2, e.g., a force, facing concern, …}
* … <!-- numbers of drivers can vary -->

## Considered Options

* configurable where it makes sense on both sidecar and egress, predefined where only one makes sense
* predefined mixed approach on a case-by-case basis
* only on the egress
* only on the sidecar

## Decision Outcome

Chosen option: "{option 1}", because {justification. e.g., only option, which meets k.o. criterion decision driver | which resolves force {force} | … | comes out best (see below)}.

### Positive Consequences <!-- optional -->

* {e.g., improvement of quality attribute satisfaction, follow-up decisions required, …}
* …

### Negative Consequences <!-- optional -->

* {e.g., compromising quality attribute, follow-up decisions required, …}
* …

## Pros and Cons of the Options

### Mixed approach on a case-by-case basis

#### MeshAccessLog

MeshAccessLog is not applied on egress at all, and we don't provide a way to configure it.
I don't see any reason not to do this though.

##### New behaviour

#### MeshCircuitBreaker

##### Old ExternalService behaviour

Circuit Breaker is configured on the sidecar always.

##### New behaviour

If on egress
- services from one zone will act as an aggregate - if a CB triggers by service-1 then service-2 will also see that result immediately
  - this makes the configuration more sensitive to drastic changes
  - if a MES starts failing on one endpoint /example1 that is only consumed by service-1 it will also trigger for service-2 that might be happily consuming endpoint /example2
  - this could be mitigated by re-defining a MeshExternalService under a different name (a bit awkward)

##### Verdict

There might be situations where it makes sense on egress but seems like it's more natural on the sidecar

#### MeshFaultInjection

##### Old ExternalService behaviour

FaultInjection and MeshFaultInjection is configured for ExternalService on egress inbound.

##### New behaviour

If on sidecar:
- less traffic in the cluster (it won't even exit the pod)

##### Verdict

Probably should stay on egress.

#### MeshHealthCheck

##### Old ExternalService behaviour

HealthCheck is configured on the sidecar.

##### New behaviour

If on egress:
- Seems to be the same story as circuit breaker in terms of acting as an aggregate / sensitivity
- Doing it on the egress could cause less traffic to MES (assuming there is fewer instances of egress than services HC-ing)

##### Verdict

To me, it makes a little bit more sense (than circuit breaker) to have HC be configurable on where it's placed

#### MeshMetric

It's a new policy without a direct ancestor.
MeshMetric is not applied on egress at all, and we don't provide a way to configure it.
I don't see any reason not to allow this in the future.

#### MeshProxyPatch

##### Old ExternalService behaviour

It does not look like it's even possible to modify egress with MeshProxyPatch or ProxyTemplate so this needs work outside the scope of this MADR.

#### MeshRateLimit

##### Old ExternalService behaviour

Both RateLimit and MeshRateLimit are configured on the egress.

##### New behaviour

If on sidecar:
- limit is multiplied by the number of instances (which is bad)
- less traffic in the mesh (traffic does )

##### Verdict

It seems that it only makes sense on egress, we don't want the ratelimit to be changing with number of instances.

#### MeshRetry

##### Old ExternalService behaviour

Both Retry and MeshRetry are configured on the sidecar

##### New behaviour

If on egress:
- when an ExternalService is unavailable a retry will only occur `numRetries` tries - not `numRetries` * instances (in sidecar case) - which is good because we're not overwhelming the service
- same story as circuit breaker in terms of acting as an aggregate / sensitivity
- we definitely shouldn't configure it both on sidecar and egress, that way we'll have squared retries

#### MeshTimeout

##### Old ExternalService behaviour

Both Timeout and MeshTimeout are configured on the sidecar.

##### New behaviour

If on egress:
- we would have to match the timeouts from the sidecar not to have a clash (that server times out before client)
- I don't see any advantages of having this on egress

#### MeshTrace

##### Old ExternalService behaviour

Both TrafficTrace and MeshTrace do not support egress.

##### New behaviour

I think it makes sense both on sidecar and egress.
If we're troubleshooting something it would be great to see spans from sidecar going to ExternalService.

#### MeshTrafficPermission

##### Old ExternalService behaviour

Both MeshTrafficPermission and TrafficPermission only works with egress.

##### New behaviour

It is already implemented on egress for MeshTrafficPermission.

If on the sidecar:
- could you circumvent this limitation by trying to directly call through egress?

#### MeshLoadBalancingStrategy

##### Old ExternalService behaviour

There is no old behaviour because there was no concept of Endpoints in External Service

##### New behaviour

Only the `loadBalancer` field makes sense in this regard.

If on sidecar:
- if you don't need/want egress it still makes sense to pick load balancing mechanism

If on egress:
- makes the same sense on as on sidecar

#### MeshTCPRoute and MeshHTTPRoute

##### Old ExternalService behaviour

It worked for both egress and sidecar.

##### New behaviour

I don't see why we would change that behaviour.

### only on the egress

Are there any policies that only make sense on sidecar?


### {option 3}

{example | description | pointer to more information | …} <!-- optional -->

* Good, because {argument a}
* Good, because {argument b}
* Bad, because {argument c}
* … <!-- numbers of pros and cons can vary -->

## Links <!-- optional -->

* https://github.com/kumahq/kuma/issues/5050
