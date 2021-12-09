# L4 TrafficRoute

## Examples 

### Kubernetes

```yaml
apiVersion: kuma.io/v1alpha1
kind: TrafficRoute
mesh: default
metadata:
  name: rule-name
  namespace: demo-app
spec:
  sources:
  - match:
      service: mobile.mobile-app.svc:8080
      region: us-east-1
      version: 10
  destinations:
  - match:
      # NOTE: only `service` tag can be used here (on `k8s` all TCP connections will have the same Virtual IP as destination => it's not enough info to infer any other destination tags)
      service: backend.backend-app.svc:8080
  conf:
  - weight: 90
    destination:
      service: backend
      region: us-east-1
      version: 2
  - weight: 10
    destination:
      service: backend.backend-app.svc:8080
      region: eu-west-3
      version: 3
```

NOTE:

1. Notice `match:` syntax (which stems from limitations of `Protobuf`)
2. `service` tag get generated automatically on `Kubernetes`. The format is `<Service Name>.<Namespace>.svc:<Service Port>`
3. Read a note above regarding constraints on tags that can be used in `destination`

### Universal

```yaml
type: TrafficRoute
name: rule-name
mesh: default
sources:
- match:
    service: mobile
    region: us-east-1
    version: 10
destinations:
- match:
      # NOTE: only `service` tag can be used here (in `universal` all TCP connections will have `127.0.0.1` as destination => it's not enough info to infer any other destination tags)
    service: backend
conf:
- weight: 90
  destination:
    service: backend
    region: us-east-1
    version: 2
- weight: 10
  destination:
    service: backend
    version: 3
```

NOTE:

1. Notice `match:` syntax (which stems from limitations of `Protobuf`)
2. Read a note above regarding constraints on tags that can be used in `destination`

## Consideration Notes 

* L4 load-balancing in `Envoy` is implemented by `envoy.filters.network.tcp_proxy` network filter
* `envoy.filters.network.tcp_proxy` does support weighted routing to multiple destination clusters, BUT it does NOT support routing to a subset of endpoints within that cluster. E.g.
  ```
    destination:
      service: backend
      region: us-east-1
      version: 2
  ```
  is a subset of endpoints inside the cluster for `service` "*backend*", but `envoy.filters.network.tcp_proxy` can only balance to the while cluster `service` "*backend*"

  * possible solutions:
    1. create an "ad-hoc" cluster that includes only those endpoints that match `destination` (this way we're effectively ignoring subset-based load balancing in Envoy) - `Istio` does that
       * cons:
         * "ad-hoc" cluster will have a unique name different from "*backend*". Which has an impact on `StatsD/Prometheus` metrics emitted by Envoy
         * larger memory footprint (can ignore it for now)
    2. contribute to `Envoy` to implement subset-based load balancing at L4
       * cons:
         * it will take time to merge it upstream and before it will appear in Envoy release
         * it's strange that `Istio` hasn't implemented it yet

## Design

In order to generate `Cluster`s and `ClusterLoadAssignment`s for a given `Dataplane`:
1. List all `TrafficRoute`s in the same `mesh` as `Dataplane`
2. Go through `TrafficRoute`s and select those `rule`s where
   * `sources` selector matches tags on one of the *inbound* interfaces of that `Dataplane`
3. For each *outbound* interface:
   * Go through the `rule`s selected at step `2` and check if `destinations`    selector matches `service` of that *outbound* interface
   * Order matched `rule`s and keep only 1 "best match"
   * if there is no "best match" `rule`
     * => generate `Cluster` for `service` tag of that *outbound* interface
     * => generate `ClusterLoadAssignment` with all `Dataplanes` that have the same `service` tag
     * => generate `envoy.filters.network.tcp_proxy` filter that has only 1 destination `Cluster` (no weight is necessary)
   * if there is a "best match" `rule`
     * => generate `envoy.filters.network.tcp_proxy` filter with weighted `Clusters` filled in as follows:
       * For each *destination* in `rule.conf`:
         * if *destination* has only `service` tag
           * => generate `Cluster` for `service` tag of that *destination*
           * => generate `ClusterLoadAssignment` with all `Dataplanes` that have the same `service` tag
         * if *destination* has tags other than `service`
           * => generate "ad-hoc" `Cluster` for all tags of that *destination*
           * => generate `ClusterLoadAssignment` with all `Dataplanes` that have *inbound* interface that matces all tags of that *destination*
         * add generated `Cluster` to `envoy.filters.network.tcp_proxy` filter with the weight of that *destination*

## Implementation notes

### Kubernetes

* When deciding which `TrafficRoute`s apply to a given `Dataplane`, control plane will ignore the `namespace` a `TrafficRoute` belongs to
  * which means that users constraint by RBAC to `namespace A` will still be able to affect routing to/from service in `namespace B`

## Considered corner cases

1. How to order/prioritize multiple `TrafficRoute`s that match the same *outbound* interface ?

   A. How to order 2 rules with identical selectors ? E.g.,

    ```yaml
    type: TrafficRoute
    name: rule-1
    sources:
    - match:
        service: mobile   # <<< same selector
    destinations:
    - match:
        service: backend
    conf:
    - weight: 100
      destination:
        service: backend
        region: us-east-1 # <<< subset 1
    ```
    
    and
    
    ```yaml
    type: TrafficRoute
    name: rule-2
    sources:
    - match:
        service: mobile   # <<< same selector
    destinations:
    - match:
        service: backend
    conf:
    - weight: 100
      destination:
        service: backend
        region: eu-west-3 # <<< subset 2
    ```

    **Answer:**

    1. the most recently created rule wins
    2. don't allow this to happen by denying a CREATE/UPDATE that introduces a new rule for exactly the same selectors

   B. Does match by 2 tags have higher priority than match by 1 tag ? E.g.,

   ```yaml
   type: TrafficRoute
   name: rule-1
   sources:
   - match:
       service: mobile   # <<< match by 1 tag
   destinations:
   - match:
       service: backend
   conf:
   - weight: 100
     destination:
       service: backend
       region: us-east-1 # <<< subset 1
   ```
   
   and
   
   ```yaml
   type: TrafficRoute
   name: rule-2
   sources:
   - match:
       service: mobile   # <<< match by 2 tags
       version: 2        #
   destinations:
   - match:
       service: backend
   conf:
   - weight: 100
     destination:
       service: backend
       region: eu-west-3 # <<< subset 2
   ```

   **Answer:** yes, a more specifc selector (i.e., by 2 tags instead of just 1) is considered a better match
   
   C. Does match by exact value have higher priority than match by '*' ? E.g.,
 
   ```yaml
   type: TrafficRoute
   name: rule-1
   sources:
   - match:
       service: mobile   # <<< match by 1 tag
   destinations:
   - match:
       service: backend
   conf:
   - weight: 100
     destination:
       service: backend
       region: us-east-1 # <<< subset 1
   ```
   
   and
   
   ```yaml
   type: TrafficRoute
   name: rule-2
   sources:
   - match:
       service: '*'      # <<< match by '*'
   destinations:
   - match:
       service: backend
   conf:
   - weight: 100
     destination:
       service: backend
       region: eu-west-3 # <<< subset 2
   ```

   **Answer:** yes, a match by the exact value has more weight than match by '*'

   D. How to order matches by different tags ? E.g.,

   ```yaml
   type: TrafficRoute
   name: rule-1
   sources:
   - match:
       service: mobile
       version: 2        # <<< match by `version`
   destinations:
   - match:
       service: backend
   conf:
   - weight: 100
     destination:
       service: backend
       region: us-east-1 # <<< subset 1
   ```
   
   and
   
   ```yaml
   type: TrafficRoute
   name: rule-2
   sources:
   - match:
       service: mobile
       region: eu-east-2 # <<< match by `region`
   destinations:
   - match:
       service: backend
   conf:
   - weight: 100
     destination:
       service: backend
       region: eu-west-3 # <<< subset 2
   ```

   **Answer:** the most recently created rule wins

   E. If 2 rules inside a single `TrafficRoute` match, should we take into account their position in the list ? E.g.,

   ```yaml
   type: TrafficRoute
   name: rule-1
   rules:
   - sources:
     - match:
         service: mobile
     destinations:
     - match:
         service: backend
     conf:
     - weight: 100
       destination:
         service: backend
         region: us-east-1 # <<< subset 1
   - sources:
     - match:
         service: mobile
     destinations:
     - match:
         service: backend
     conf:
     - weight: 100
       destination:
         service: backend
         region: eu-west-3 # <<< subset 2
   ```

   **Answer:** prevent this situation from happening by simplifying `TrafficRoute` - there should be only 1 `rule` per `TrafficRoute`

   F. What if a new `TrafficRoute` doesn't cause collisions at the moment when it's created, but it starts causing collisions later when a new `Dataplane` is added or an existing `Dataplane` definition is updated with extra `tags` ?

   **Answer:**
   1. the most recently created rule wins
   2. Control Plane should expose metrics that make this situation noticeable
   3. either incorporate information about collisions into the existing `DataplaneInsights` resource or introduce a new `Alert` resource
   4. distinguish different levels of alerts: `warnings` vs `errors`

2. Naming of "ad-hoc" clusters
   * What naming scheme to use ?

     **Answer:** concatenate all labeles

   * How will it affect per `Cluster` metrics ?

     **Answer:** it is possible to aggregate (inside `Envoy`) metrics of an "ad-hoc" cluster into the metrics of the cluster per `service`

## Out of scope

* circuit breakers
* health checking
* zone-aware load-balancing
* priority-based load-balancing
