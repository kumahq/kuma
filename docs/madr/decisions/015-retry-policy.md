# Retry policy

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/4738

## Context and Problem Statement

We want to create a new policy for retries that uses TargetRef [matching](https://github.com/kumahq/kuma/blob/22c157d4adac7f518b1b49939c7e9ea4d2a1876c/docs/madr/decisions/005-policy-matching.md).

### Specification 

#### Name

MeshRetry

#### From level

Even though it's possible to place a retry policy on the inbound routes, 
it doesn't seem like there is a value in it.
We won't add inbound retries mostly because retries create additional load on the service.
Most of the time retrying requests is a client's concern.

#### Top level

Top-level targetRef can have the following kinds:

```yaml
targetRef:
  kind: Mesh|MeshSubset|MeshService|MeshServiceSubset
  name: ...
```

#### To level

To-level targetRef can have the following kinds:

```yaml
targetRef:
  kind: Mesh|MeshService
  name: ...
```

#### Conf

Existing Retry policy:

```yaml
conf:
  http:
    numRetries: 5
    perTryTimeout: 200ms
    backOff:
      baseInterval: 20ms
      maxInterval: 1s
    retriableStatusCodes:
      - 500
      - 504
    retriableMethods:
      - GET
    retryOn:
      - all_5xx
      - gateway_error
      - reset
      - connect_failure
      - envoy_ratelimited
      - retriable_4xx
      - refused_stream
      - retriable_status_codes
      - retriable_headers
      - http3_post_connect_failure
  grpc:
    numRetries: 5
    perTryTimeout: 200ms
    backOff:
      baseInterval: 20ms
      maxInterval: 1s
    retryOn:
      - cancelled
      - deadline_exceeded
      - internal
      - resource_exhausted
      - unavailable
  tcp:
    maxConnectAttempts: 3
```

We can slightly change the conf for better compatibility with merging:

```yaml
default:
  http:
    numRetries: 5
    perTryTimeout: 200ms
    backOff:
      baseInterval: 20ms
      maxInterval: 1s
    retriableMethods:
      - GET
    retryOn:
      - 5xx
      - gateway_error
      - reset
      - connect_failure
      - envoy_ratelimited
      - retriable_4xx
      - refused_stream
      - http3_post_connect_failure
      - 400
      - 504
    retriableResponseHeaders:
      - name: header-name
        exact: value
    retriableRequestHeaders:
      - name: should-retry-this
        exact: yes
  grpc:
    numRetries: 5
    perTryTimeout: 200ms
    backOff:
      baseInterval: 20ms
      maxInterval: 1s
    retryOn:
      - cancelled
      - deadline_exceeded
      - internal
      - resource_exhausted
      - unavailable
  tcp:
    maxConnectAttempts: 3
```

Changes:

1. Rename `http.retryOn.all_5xx` to `http.retryOn.5xx`. Apparently it was a limitation of proto enums.
2. Merge `retriableStatusCodes` into `retryOn` list. 
Kuma can automatically generate `retriable_status_codes` list for Envoy based on `retryOn`.
3. New field `retriableResponseHeaders`. Even though we have already had `retryOn.retriable_headers` 
there was no way to provide a list of headers. 
If this field is set we automatically add `retriable_headers` to `retryOn`.
4. New field `retriableRequestHeaders`. Purely for the sake of symmetry, 
if we allow retrying on response headers why not allow that on request headers. 

#### Other retry-related Envoy parameters

So far we didn't get requests from users to extend the policy with other Envoy parameters.
But there are few additional things we can potentially support.

##### per_try_idle_timeout

Same as `idle_timeout` but enforced on each individual attempt.

##### Host selection

Envoy has "host predicates" and "priority predicates" to enforce an additional logic 
when selecting a host for the attempt:

* don't retry on previously retried hosts
* don't retry on canary hosts (the host should be preliminarily marked with `canary: true`)
* don't retry on hosts with specified metadata key/value
* don't retry on previous priorities

These predicates can be combined:

```yaml
hostSelection:
  - predicate: OMIT_PREVIOUS_HOSTS
  - predicate: OMIT_CANARY_HOSTS
  - predicate: OMIT_HOSTS_WITH_TAGS
    tags:
      env: dev
  - predicate: OMIT_PREVIOUS_PRIORITIES
    updateFrequency: 2 # how often the priority load should be updated based on previously attempted priorities                                                                                             
```

Having configurable `hostSelection` implies exposing the `host_selection_retry_max_attempts` parameter as well.
It allows configuring the maximum number of times host selection will be reattempted before giving up.

##### Retry budget

This parameter allows specifying a limit on concurrent retries in relation to the number of active requests.
The problem is that `retry budget` is set on Envoy cluster 
while other parameters are set on a "per route" or "per virtual host" basis.
Even though today Kuma generates an Envoy cluster "per route" it might not be the case in the future.
So I don't think having `retry budget` in MeshRetry policy is a good idea.

## Considered Options

* Implement MeshRetry without additional parameters
* Implement MeshRetry with `per_try_idle_timeout` and `hostSelection`

### Implement MeshRetry without additional parameters

#### Pros

* easy to implement
* policy is not overloaded with parameters

#### Cons

* MeshRetry could be considered by some users as not flexible enough 

### Implement MeshRetry with `per_try_idle_timeout` and `hostSelection`

#### Pros

* users can configure complex host selection strategies for retries

#### Cons

* more difficult implementation
* policy is overloaded with parameters

## Decision Outcome

Chosen option: "Implement MeshRetry without additional parameters". 
Let's start simple. 
We can always extend the policy in the future if we see interest in the community. 