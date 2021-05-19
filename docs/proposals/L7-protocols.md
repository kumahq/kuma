# Support for L7 Protocols

## Context

* In order to support `L7` features (such as _HTTP Routing_, _HTTP Metrics_, _HTTP Access Logs_, _Distributed Tracing_, etc),
  `Control Plane` needs to know the protocol of a service that is being proxied since
  configuration for proxying _HTTP_ traffic is very different from configuration for _MySQL_, _MongoDB_, _Kafka_, etc
* Technically, `Envoy` allows to define multiple configurations for a single listener and choose one of them at runtime based on actual traffic
  * E.g., a listener for `192.168.0.1:8080` can have 2 configurations:
    * 1st configuration that treats traffic as opaque TCP
    * 2nd configuration that treats traffic as HTTP
  * To be able to choose a proper alternative at runtime,
    `Envoy` configuration must also include a `Listener Filter` that is responsible for inference of application protocol from the actual traffic
  * `Listener Filter` will be buffering connection data until it has seen enough to make a verdict
  * Out-of-the-box, `Envoy` comes with a `Listener Filter` that is capable to detect `HTTP 1.x` and `HTTP 2`
* Although a dataplane could automatically infer application protocol by introspecting actual traffic,
  it would imply some overhead on every TCP connection
* To give users a choice between convenience of auto-configuration and guaranteed performance, we need to support a use case where users prefer to define application protocol explicitly

## Requirements

* Provide a way for users to define application `protocol` explicitly
* Take into account possible future look of `Kuma` configuration (e.g., `ServiceProfile`)

## Alternative configuration models

* At first, it seems natural to include information about application `protocol` into `Dataplane` definition, e.g. as a `protocol` tag
* However, in the future, `Kuma` might introduce a notion of `ServiceProfile` to let users describe their services at the level of APIs rather than at the level of networking
  * e.g., this `service` is a _REST API_, _gRPC service_, _Postgres database_, etc

Comparison table:

| Source of configuration | Pros                                                                                                    | Cons  |
| ---------------- | ------------------------------------------------------------------------------------------------------- | ----- |
| `Dataplane`      |                                                                                                         | different `Dataplanes` might specify different `protocol` for the same `service` |
| `ServiceProfile` | users describe their high-level services, `Kuma` infers low-level technical aspects, such as `protocol` | |

### Configuration model #1

* User explicitly defines application `protocol` as a tag in `Dataplane` resource

#### Universal mode

 * To define application protocol explicitly, a user should add tag `protocol` to the `inbound` interface of a `Dataplane`
 * Tag `protocol` is optional
 * If tag `protocol` is missing, `TCP` will be assumed (current behaviour)

E.g.,

```yaml
type: Dataplane
mesh: default
name: backend
networking:
  inbound:
  - interface: 192.168.0.1:18080:8080
    tags:
      service: backend
      protocol: http   # notice `protocol` tag
```

#### Kubernetes mode

* Since `Dataplane` resource is auto-generated on `k8s`, a user should edit other `k8s` resources instead
* To define application protocol explicitly, a user should edit `Service` resource (rather than `Deployment` / `Pod`)
  * The reason for that is that `k8s` does not require `Deployment` / `Pod` to explicitly mention all ports exposed by the app
  * On the other hand, `k8s` does require `Service` to explicitly mention all its ports
* Annotations might be too verbose and redundant if a user already sticks to minimal conventions when naming service ports
  * E.g., if a user already gives to service ports names like `http`, `http-metrics`, `grpc`, `postgres`, that information is (optimistically) enough for `Kuma`
* If a user prefers explicit annotations, e.g. to make it very clear why this particular piece of configuration is important, he should use `<port>.service.kuma.io/protocol`
 * In case if there is no explicit annotation and port naming conventions are not followed, `TCP` protocol will be assumed (current behaviour)

E.g.,

* A user might opt for minimal port naming conventions (`protocol` can appear anywhere inside the port `name`)

  ```yaml
  apiVersion: v1
  kind: Service
  metadata:
    name: backend
    namespace: example
  spec:
    ports:
    - port: 8080
      name: http          # `http` protocol in the `name`
    - port: 7070
      name: http-metrics  # `http` protocol in the `name`
    - port: 6060
      name: somethinggrpc # `grpc` protocol in the `name`
    selector:
      app: backend
  ```

* A user might opt for explicit annotations in the form `<port>.service.kuma.io/protocol`

  ```yaml
  apiVersion: v1
  kind: Service
  metadata:
    name: backend
    namespace: example
    annotations:
      8080.service.kuma.io/protocol: http # explicit `http` protocol
      7070.service.kuma.io/protocol: http # explicit `http` protocol
      6060.service.kuma.io/protocol: grpc # explicit `grpc` protocol
  spec:
    ports:
    - port: 8080
      name: a
    - port: 7070
      name: b
    - port: 6060
      name: tcp  # `protocol` from annotation will take precedence
    selector:
      app: backend
  ```

### Configuration model #2

* User describes a high-level service, `Kuma` infers low-level technical aspects, such as `protocol`

#### Universal mode

* REST API
  ```yaml
  type: ServiceProfile
  mesh: default
  name: backend # the same value as in `service` tag
  profiles:
  - api:
      rest:
        spec:
          url: https://generator.swagger.io/api/swagger.json
  ```
* gRPC service
  ```yaml
  type: ServiceProfile
  mesh: default
  name: backend # the same value as in `service` tag
  profiles:
  - api:
      grpc:
        spec:
          url: https://github.com/envoyproxy/envoy/blob/master/api/envoy/rvice/accesslog/v3/als.proto
  ```
* Postgres database
  ```yaml
  type: ServiceProfile
  mesh: default
  name: db      # the same value as in `service` tag
  profiles:
  - database:
      postgres: {}
  ```

#### Kubernetes mode

```yaml
apiVersion: kuma.io/v1alpha1
kind: ServiceProfile
metadata:
  namespace: example
  name: backend # the same name as `Service` resource has
mesh: default
spec:
  profiles:
  - api:
      rest:
        spec:
          url: https://generator.swagger.io/api/swagger.json
```

## Other Considerations

### About other ways to infer application protocol automatically

* port numbers (e.g., `80`, `443`, `8080`, `5432`) and `k8s` health checks (`HTTP` and `TCP`) are another good strategy to infer application protocol 
* however, they are not a part of this proposal
* we could add support for it in the future, probably in the form of recommendations

### About "protocol" tag in case of ServiceProfile

* If it is important to have `protocol` tag for the sake of `Kuma` policy matching (e.g., to be able to apply a policy to all `http` services), this tag could be generated automatically out of a `ServiceProfile`

## Open questions

1. What happens if a user defines different `protocol` tag for the same `service` in different `Dataplanes` ?
   * How should Control Plane handle it ?
   * How should Control Plane communicate it to the user ?
