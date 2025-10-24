# Transition to new policies

- Status: accepted

## Context and Problem Statement

The current state of the policies as of Kuma 2.6.x:
* new policies are in GA, they're the recommended way to configure new deployments in Kuma
* old policies are not deprecated, they're still in use by majority of Kuma users
* as long as there is at least 1 policy of a new type (e.g. MeshTimeout), the policies of the corresponding old type (e.g. Timeout) are not applied
  (i.e. if there is a MeshTimeout policy, the Timeout policies is not applied)

The goal is to provide a way to transition from old policies to new policies. Having a smooth transition is going to enable 
us to deprecate old policies in the future.

Prerequisites for the transition:
* both Global and Zone CPs are running the latest Kuma version

## Decision Drivers

- Predictability: when replacement policies are created (manually or automatically), user should be able to verify them 
without applying for real
- Short feedback loop: we'd like to deliver small portions of new policies more often, rather than big portions but rarely
- Quick rollback: user should be able to rollback to old policies

## Considered Options

- Introduce a shadow mode for policies on Zone CP
- Introduce a shadow mode for policies on Global CP
  - Show potential changes for a specific DPP
  - Show potential changes for all affected DPPs

## Decision Outcome

- Introduce a shadow mode for policies only on Zone CP

## Shadow mode for policies

Shadow mode is a way to test new policies without affecting the real traffic. When the policy is created with `kuma.io/effect: shadow` label

```yaml
type: MeshTimeout
name: timeout-global
mesh: default
labels:
  kuma.io/effect: shadow
spec:
  targetRef:
    kind: Mesh
  to: [...]
```

it doesn't apply on the real proxies. Instead, by using Inspect API we can inspect what config dump would look like if the shadow policies was applied.

### Changes in the Inspect API

Introduce a new non-physical endpoint `/_config` that returns a config dump of a DPP without the actual `/config_dump` call to Envoy.
If we want to be able to compute diff on the client side we can't use `xds` endpoint as its response slightly diverges from what CP has in the SnapshotCache.

Add a query parameter `shadow=true`:

* `GET /meshes/{mesh}/dataplanes/{name}/_config` -> `GET /meshes/{mesh}/dataplanes/{name}/_config?shadow=true`
* `GET /meshes/{mesh}/dataplanes/{name}/_rules` -> `GET /meshes/{mesh}/dataplanes/{name}/_rules?shadow=true`

when `shadow=true` is set, the response will take policies with `kuma.io/effect: shadow` label into account.

For visualization in GUI it would be nice to include the diff between the current XDS config and the shadow XDS config.
When `include=diff` query parameter is set the response includes the diff in a JSONPatch format.
The config in response `xds` filed is always a result of JSONPatch diff from `diff` field applied to the current XDS config.
So when the query is `shadow=false&include=diff` the server should return an empty JSONPatch diff alongside the current XDS config.

OpenAPI spec for new endpoints:

```yaml
openapi: 3.0.3
info:
  title: Shadow Policies Inspector
  description: API for inspecting policies' effect on the Data Plane Proxy configurations
  version: 1.0.0
paths:
  /meshes/{mesh}/dataplanes/{name}/_config:
    get:
      summary: Get a proxy XDS config on a CP
      description: Returns the configuration of the proxy ([xds](https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol) configuration).
      parameters:
        - in: path
          name: mesh
          required: true
          description: The mesh of the DPP to get the diff for.
          schema:
            type: string
        - in: path
          name: name
          required: true
          description: The name of the DPP within the mesh to get the diff for.
          schema:
            type: string
        - in: query
          name: shadow
          description: |
            When computing XDS config the CP takes into account policies with 'kuma.io/effect: shadow' label
          schema:
            type: boolean
            default: false
        - in: query
          name: include
          description: |
            An array of extra fields to include in the response. When `include=diff` the server computes a diff in JSONPatch format
            between the XDS config returned in 'xds' and the current proxy XDS config.
          schema:
            type: array
            items:
              type: string
              enum: [diff]
      responses:
        '200':
          description: Successfully retrieved proxy XDS config.
          content:
            application/json:
              schema:
                type: object
                required: [xds]
                properties:
                  xds:
                    description: The raw XDS config as an inline JSON object
                    type: object
                  diff:
                    description: |
                      When 'include=diff' query parameter is specified, the field contains a diff in JSONPatch format
                      between the XDS config returned in 'xds' and the current proxy XDS config.
                    type: array
                    items:
                      type: object
                      required:
                        - op
                        - path
                      properties:
                        op:
                          type: string
                          description: Operation to be performed.
                          enum: [add, remove, replace, move, copy, test]
                        path:
                          type: string
                          description: A JSON Pointer path indicating the part of the document to operate on.
                        value:
                          type: object
                          description: The value to be used within the operations.
        '400':
          description: Bad request. Parameters were invalid.
        '404':
          description: The specified mesh/name combination was not found.
```

Endpoint `/_rules` is getting the same `shadow` and `include` query parameters. 
The new `diff` field is going to be placed at the root level of [InspectRules](https://github.com/kumahq/kuma/blob/9d93621900c57c3569a53924f0ec836c88eba559/api/openapi/specs/api.yaml#L148-L163) object:

```yaml
InspectRules:
  type: object
  title: InspectRules
  description: A list of rules for a dataplane
  required: [rules, resource, httpMatches]
  properties:
    resource:
      $ref: './common/resource.yaml#/components/schemas/Meta'
    rules:
      type: array
      items:
        $ref: './common/resource.yaml#/components/schemas/InspectRule'
    httpMatches:
      type: array
      items:
        $ref: './common/resource.yaml#/components/schemas/HttpMatch'
    diff:
      description: |
        When 'include=diff' query parameter is specified, the field contains a diff in JSONPatch format
        between the rules returned in 'rules' field and the current proxy rules.
      type: array
      items: [...]
```

### UX

We can use a CLI tool like [jd](https://github.com/josephburnett/jd?tab=readme-ov-file#examples) to visualize the structural diff.
The goal is to avoid dependencies on `jd` in the codebase, but to provide examples in the documentation that enable users to see diffs conveniently, i.e:

```shell
kumactl inspect dataplane dpp-1 --config-dump --shadow > with-shadow.json && jd <(kumactl inspect dataplane dpp-1 --config-dump) with-shadow.json
```

### Implementation

Implementation wise, shadow mode requires generating 2 Snapshots: one with and without the new policy. 
By comparing these Snapshots we can generate a list of jsonPatches that would be applied to the Envoy configuration. 

Depending on how much time do we want to spend on the implementation, we can choose to implement shadow mode for Zone CP or Global CP.
Global CP support is more time-consuming, but it can provide a better UX, i.e. showing potential changes for all affected DPPs (not only for a specific DPP).
Implementing shadow mode for Global CP should be done by forwarding the request to Zone CPs. 
Computing the Snapshot right on Global CP requires gathering all DPPs from all zones which can be a performance bottleneck.

## Rolling out the new policies

After making sure that list of jsonPatches is correct, we can start applying new policies on the real proxies.
In order to limit the blast radius, we can start with a small portion of proxies, by selecting them with MeshSubset targetRef.

1. Mark some workloads with label like `opt-in-for-new-kuma-policies: true`
2. Set top-level targetRef on policy to: 
```yaml
spec:
  targetRef:
    kind: MeshSubset
    tags:
      opt-in-for-new-kuma-policies: true
```
3. Apply the policy and observe the app's metrics and logs
4. If everything is fine, apply the policy on the rest of the workloads
5. Remove the old policy

## Examples

### Timeout -> MeshTimeout

There are 3 Timeout resources:

```yaml
type: Timeout
mesh: default
name: timeout-global
sources:
  - match:
      kuma.io/service: '*'
destinations:
  - match:
      kuma.io/service: '*'
conf:
  connectTimeout: 21s
  tcp: 
    idleTimeout: 22s
  http:
    requestTimeout: 23s
    idleTimeout: 24s
    streamIdleTimeout: 25s
    maxStreamDuration: 26s
---
type: Timeout
mesh: default
name: timeout-to-backend
sources:
  - match:
      kuma.io/service: '*'
destinations:
  - match:
      kuma.io/service: 'backend_kuma-demo_svc_3001'
conf:
  connectTimeout: 31s
  tcp: 
    idleTimeout: 32s
  http:
    requestTimeout: 33s # leaking when MeshTimeout is applied
    idleTimeout: 34s
    streamIdleTimeout: 35s # leaking when MeshTimeout is applied
    maxStreamDuration: 36s
---
type: Timeout
mesh: default
name: timeout-to-redis
sources:
  - match:
      kuma.io/service: '*'
destinations:
  - match:
      kuma.io/service: 'redis_kuma-demo_svc_6379'
conf:
  connectTimeout: 41s
  tcp: 
    idleTimeout: 42s
  http:
    requestTimeout: 43s
    idleTimeout: 44s
    streamIdleTimeout: 45s
    maxStreamDuration: 46s
```

User creates 3 new MeshTimeout replacements:

```yaml
type: MeshTimeout
name: timeout-global
mesh: default
spec:
  targetRef:
    kind: Mesh
  to:
  - targetRef:
      kind: Mesh
    default:
      connectionTimeout: 21s
      idleTimeout: 22s
      http:
        requestTimeout: 23s
        streamIdleTimeout: 25s
        maxStreamDuration: 26s
        maxConnectionDuration: 27s
  from:
  - targetRef:
      kind: Mesh
    default:
      connectionTimeout: 10s
      idleTimeout: 2h
      http:
        requestTimeout: 0s
        streamIdleTimeout: 1h
---
type: MeshTimeout
name: aaa-timeout-to-backend
mesh: default
spec:
  targetRef:
    kind: Mesh
  to:
  - targetRef:
      kind: MeshService
      name: backend_kuma-demo_svc_3001
    default:
      connectionTimeout: 31s
      idleTimeout: 34s
      http:
        requestTimeout: 33s
        streamIdleTimeout: 35s
        maxStreamDuration: 36s
        maxConnectionDuration: 37s
---
type: MeshTimeout
name: aaa-timeout-to-redis
mesh: default
spec:
  targetRef:
    kind: Mesh
  to:
  - targetRef:
      kind: MeshService
      name: redis_kuma-demo_svc_6379
    default:
      connectionTimeout: 41s
      idleTimeout: 42s
      http:
        requestTimeout: 43s
        streamIdleTimeout: 45s
        maxStreamDuration: 46s
        maxConnectionDuration: 47s
---
# we didn't have special Timeout for frontend, but since we merged 'idleTimeout' and 'http.idleTimeout' now we need
# a separate MeshTimeout policy to have the same '24s' value
type: MeshTimeout
name: aaa-timeout-to-frontend
mesh: default
spec:
  targetRef:
    kind: Mesh
  to:
  - targetRef:
      kind: MeshService
      name: frontend_kuma-demo_svc_8080
    default:
      idleTimeout: 24s
```

Using shadow mode we see 12 changes:

```yaml
- value: 0s
  op: add
  path: envoy.config.cluster.v3.Cluster/backend_kuma-demo_svc_3001/typedExtensionProtocolOptions/envoy.extensions.upstreams.http.v3.HttpProtocolOptions/commonHttpProtocolOptions/maxConnectionDuration
- value: 0s
  op: add
  path: envoy.config.cluster.v3.Cluster/frontend_kuma-demo_svc_8080/typedExtensionProtocolOptions/envoy.extensions.upstreams.http.v3.HttpProtocolOptions/commonHttpProtocolOptions/maxConnectionDuration
- value:
    idleTimeout: 7200s
    maxConnectionDuration: 0s
    maxStreamDuration: 0s
  op: replace
  path: envoy.config.cluster.v3.Cluster/localhost:8080/typedExtensionProtocolOptions/envoy.extensions.upstreams.http.v3.HttpProtocolOptions/commonHttpProtocolOptions
- value: 0s
  op: replace
  path: envoy.config.listener.v3.Listener/inbound:10.42.0.29:8080/filterChains/0/filters/0/typedConfig/commonHttpProtocolOptions/idleTimeout
- value: 0s
  op: add
  path: envoy.config.listener.v3.Listener/inbound:10.42.0.29:8080/filterChains/0/filters/0/typedConfig/requestHeadersTimeout
- value: 3600s
  op: add
  path: envoy.config.listener.v3.Listener/inbound:10.42.0.29:8080/filterChains/0/filters/0/typedConfig/routeConfig/virtualHosts/0/routes/0/route/idleTimeout
- value: 0s
  op: replace
  path: envoy.config.listener.v3.Listener/outbound:10.43.13.71:3001/filterChains/0/filters/0/typedConfig/commonHttpProtocolOptions/idleTimeout
- value: 0s
  op: add
  path: envoy.config.listener.v3.Listener/outbound:10.43.13.71:3001/filterChains/0/filters/0/typedConfig/requestHeadersTimeout
- value: 35s
  op: add
  path: envoy.config.listener.v3.Listener/outbound:10.43.13.71:3001/filterChains/0/filters/0/typedConfig/routeConfig/virtualHosts/0/routes/0/route/idleTimeout
- value: 0s
  op: replace
  path: envoy.config.listener.v3.Listener/outbound:10.43.138.50:8080/filterChains/0/filters/0/typedConfig/commonHttpProtocolOptions/idleTimeout
- value: 0s
  op: add
  path: envoy.config.listener.v3.Listener/outbound:10.43.138.50:8080/filterChains/0/filters/0/typedConfig/requestHeadersTimeout
- value: 25s
  op: add
  path: envoy.config.listener.v3.Listener/outbound:10.43.138.50:8080/filterChains/0/filters/0/typedConfig/routeConfig/virtualHosts/0/routes/0/route/idleTimeout
```

Let's review these changes:

* timeouts `requestHeadersTimeout` and `maxConnectionDuration` are disabled when equal to `0s` or unset. There was no 
way to specify them with old Timeout policy.
* `commonHttpProtocolOptions/idleTimeout` with MeshTimeout is set on cluster and disabled on the listener 
* `route/idleTimeout` is duplicated value of `streamIdleTimeout` but per-route (previously we've set it only per-listener)

### CircuitBreaker -> MeshCircuitBreaker

There is an old CircuitBreaker policy:
```yaml
type: CircuitBreaker
mesh: default
name: circuit-breaker-example
sources:
- match:
    kuma.io/service: "*"
destinations:
- match:
    kuma.io/service: "*"
conf:
  interval: 21s
  baseEjectionTime: 22s
  maxEjectionPercent: 23
  splitExternalAndLocalErrors: false
  thresholds:
    maxConnections: 24
    maxPendingRequests: 25
    maxRequests: 26
    maxRetries: 27
  detectors:
    totalErrors: 
      consecutive: 28
    gatewayErrors: 
      consecutive: 29
    localErrors: 
      consecutive: 30
    standardDeviation:
      requestVolume: 31
      minimumHosts: 32
      factor: 1.33
    failure:
      requestVolume: 34
      minimumHosts: 35
      threshold: 36
```

User creates a new MeshCircuitBreaker replacement:

```yaml
type: MeshCircuitBreaker
name: backend-inbound-outlier-detection
mesh: default
spec:
  targetRef:
    kind: Mesh
  to:
  - targetRef:
      kind: Mesh
    default:
      connectionLimits:
        maxConnections: 24
        maxPendingRequests: 25
        maxRequests: 26
        maxRetries: 27
      outlierDetection:
        interval: 21s
        baseEjectionTime: 22s
        maxEjectionPercent: 23
        splitExternalAndLocalErrors: false
        detectors:
          totalFailures:
            consecutive: 28
          gatewayFailures:
            consecutive: 29
          localOriginFailures:
            consecutive: 30
          successRate:
            requestVolume: 31
            minimumHosts: 32
            standardDeviationFactor: 1.33
          failurePercentage:
            requestVolume: 34
            minimumHosts: 35
            threshold: 36
```

Using shadow mode we see there is an empty jsonPatches list!

### FaultInjection -> MeshFaultInjection

It should be always safe to remove FaultInjection and create new MeshFaultInjection when necessary. 

### TrafficPermission -> MeshTrafficPermission

```yaml
type: TrafficPermission
name: on-redis
mesh: default
sources:
  - match:
      kuma.io/service: frontend_kuma-demo_svc_8080
  - match:
      kuma.io/service: backend_kuma-demo_svc_3001
destinations:
  - match:
      kuma.io/service: redis_kuma-demo_svc_6379
---
type: TrafficPermission
name: on-frontend
mesh: default
sources:
  - match:
      kuma.io/service: *
destinations:
  - match:
      kuma.io/service: frontend_kuma-demo_svc_8080
```

```yaml
type: MeshTrafficPermission
name: on-redis
mesh: default
spec:
  targetRef:
    kind: MeshService
    name: redis_kuma-demo_svc_6379
  from:
    - targetRef:
        kind: MeshService
        name: frontend_kuma-demo_svc_8080
      default:
        action: Allow
    - targetRef:
        kind: MeshService
        name: backend_kuma-demo_svc_3001
      default:
        action: Allow
---
type: MeshTrafficPermission
name: on-frontend
mesh: default
spec:
  targetRef:
    kind: MeshService
    name: frontend_kuma-demo_svc_8080
  from:
    - targetRef:
        kind: Mesh
      default:
        action: Allow
```

Since we rename the policy name in envoy (from having the actual policy name to having a hardcoded value "MeshTrafficPermission")
the diff is not really useful:

```yaml
# on Redis DPPs
- value:
    permissions:
      - any: true
    principals:
      - authenticated:
          principalName:
            exact: spiffe://default/backend_kuma-demo_svc_3001
      - authenticated:
          principalName:
            exact: spiffe://default/frontend_kuma-demo_svc_8080
  op: add
  path: envoy.config.listener.v3.Listener/inbound:10.42.0.28:6379/filterChains/0/filters/0/typedConfig/rules/policies/MeshTrafficPermission
- op: remove
  path: envoy.config.listener.v3.Listener/inbound:10.42.0.28:6379/filterChains/0/filters/0/typedConfig/rules/policies/backend-to-httpbin-4b4zbcc4422zz784
```

It doesn't get any better with HTTP either. Previously TrafficPermission was always applied on TCP level, by applying it on HTTP 
level for HTTP services we're getting a lot of noise in the diff:

```yaml
# on Frontend DPPs
- value:
    - name: envoy.filters.network.http_connection_manager
      typedConfig:
        '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
        commonHttpProtocolOptions:
          idleTimeout: 0s
        forwardClientCertDetails: SANITIZE_SET
        httpFilters:
          - name: envoy.filters.http.rbac
            typedConfig:
              '@type': type.googleapis.com/envoy.extensions.filters.http.rbac.v3.RBAC
              rules:
                policies:
                  MeshTrafficPermission:
                    permissions:
                      - any: true
                    principals:
                      - any: true
          - name: envoy.filters.http.router
            typedConfig:
              '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
        requestHeadersTimeout: 0s
        routeConfig:
          name: inbound:frontend_kuma-demo_svc_8080
          requestHeadersToRemove:
            - x-kuma-tags
          validateClusters: false
          virtualHosts:
            - domains:
                - '*'
              name: frontend_kuma-demo_svc_8080
              routes:
                - match:
                    prefix: /
                  route:
                    cluster: localhost:8080
                    idleTimeout: 3600s
                    timeout: 0s
        setCurrentClientCertDetails:
          uri: true
        statPrefix: localhost_8080
        streamIdleTimeout: 3600s
  op: replace
  path: /type.googleapis.com~1envoy.config.listener.v3.Listener/inbound:10.42.0.16:8080/filterChains/0/filters
```

## TrafficRoute -> MeshHTTPRoute(MeshTCPRoute)

Covered by [046-route-transition.md](046-route-transition.md)

## ProxyTemplate -> MeshProxyPatch

This is where shadow mode shines the most even not in the context of the transition but for a regular usage.

```yaml
type: MeshProxyPatch
mesh: default
name: custom-template-1
labels:
  kuma.io/effect: shadow
spec:
  targetRef:
    kind: MeshService
    name: frontend_kuma-demo_svc_8080
  default:
    appendModifications:
      - cluster:
          operation: Add
          value: |
            name: test-cluster
            connectTimeout: 5s
            type: STATIC
```

```yaml
- value:
    connectTimeout: 5s
    name: test-cluster
    type: STATIC
  op: add
  path: /type.googleapis.com~1envoy.config.cluster.v3.Cluster/test-cluster
```
