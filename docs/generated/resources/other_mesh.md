## Mesh

- `mtls` (optional)

    mTLS settings.
    +optional

    Child properties:    
    
    - `enabledbackend` (required)
    
        Name of the enabled backend    
    
    - `backends` (required, repeated)
    
        List of available Certificate Authority backends

- `tracing` (optional)

    Tracing settings.
    +optional

    Child properties:    
    
    - `defaultbackend` (required)
    
        Name of the default backend    
    
    - `backends` (required, repeated)
    
        List of available tracing backends

- `logging` (optional)

    Logging settings.
    +optional

    Child properties:    
    
    - `defaultbackend` (required)
    
        Name of the default backend    
    
    - `backends` (required, repeated)
    
        List of available logging backends

- `metrics` (optional)

    Configuration for metrics collected and exposed by dataplanes.
    
    Settings defined here become defaults for every dataplane in a given Mesh.
    Additionally, it is also possible to further customize this configuration
    for each dataplane individually using Dataplane resource.
    +optional

    Child properties:    
    
    - `enabledbackend` (optional)
    
        Name of the enabled backend    
    
    - `backends` (optional, repeated)
    
        List of available Metrics backends

- `networking` (optional)

    Networking settings of the mesh

    Child properties:    
    
    - `outbound` (optional)
    
        Outbound settings
    
        Child properties:    
        
        - `passthrough` (optional)
        
            Control the passthrough cluster

- `routing` (optional)

    Routing settings of the mesh

    Child properties:    
    
    - `localityawareloadbalancing` (optional)
    
        Enable the Locality Aware Load Balancing    
    
    - `zoneegress` (optional)
    
        Enable routing traffic to services in other zone or external services
        through ZoneEgress. Default: false

- `constraints` (optional)

    Constraints that applies to the mesh and its entities

    Child properties:    
    
    - `dataplaneproxy` (required)
    
        DataplaneProxyMembership defines a set of requirements for data plane
        proxies to be a member of the mesh.
    
        Child properties:    
        
        - `requirements` (optional, repeated)
        
            Requirements defines a set of requirements that data plane proxies must
            fulfill in order to join the mesh. A data plane proxy must fulfill at
            least one requirement in order to join the mesh. Empty list of allowed
            requirements means that any proxy that is not explicitly denied can join.    
        
        - `restrictions` (optional, repeated)
        
            Restrictions defines a set of restrictions that data plane proxies cannot
            fulfill in order to join the mesh. A data plane proxy cannot fulfill any
            requirement in order to join the mesh.
            Restrictions takes precedence over requirements.
## CertificateAuthorityBackend

- `name` (required)

    Name of the backend

- `type` (required)

    Type of the backend. Has to be one of the loaded plugins (Kuma ships with
    builtin and provided)

- `dpcert` (optional)

    Dataplane certificate settings

    Child properties:    
    
    - `rotation` (optional)
    
        Rotation settings
    
        Child properties:    
        
        - `expiration` (optional)
        
            Time after which generated certificate for Dataplane will expire    
    
    - `requesttimeout` (optional)
    
        Timeout on request to CA for DP certificate generation and retrieval

- `conf` (optional)

    Configuration of the backend

- `mode` (optional)

    Mode defines the behaviour of inbound listeners with regard to traffic
    encryption

    Supported values:

    - `STRICT`

    - `PERMISSIVE`

- `rootchain` (optional)

    Child properties:    
    
    - `requesttimeout` (optional)
    
        Timeout on request for to CA for root certificate chain.
## Networking

- `outbound` (optional)

    Outbound settings

    Child properties:    
    
    - `passthrough` (optional)
    
        Control the passthrough cluster
## Tracing

- `defaultbackend` (required)

    Name of the default backend

- `backends` (required, repeated)

    List of available tracing backends
## TracingBackend

- `name` (required)

    Name of the backend, can be then used in Mesh.tracing.defaultBackend or in
    TrafficTrace

- `sampling` (optional)

    Percentage of traces that will be sent to the backend (range 0.0 - 100.0).
    Empty value defaults to 100.0%

- `type` (required)

    Type of the backend (Kuma ships with 'zipkin')

- `conf` (required)

    Configuration of the backend
## DatadogTracingBackendConfig

- `address` (required)

    Address of datadog collector.

- `port` (required)

    Port of datadog collector
## ZipkinTracingBackendConfig

- `url` (required)

    Address of Zipkin collector.

- `traceid128bit` (optional)

    Generate 128bit traces. Default: false

- `apiversion` (required)

    Version of the API. values: httpJson, httpJsonV1, httpProto. Default:
    httpJson see
    https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/trace/v3/trace.proto#envoy-v3-api-enum-config-trace-v3-zipkinconfig-collectorendpointversion

- `sharedspancontext` (optional)

    Determines whether client and server spans will share the same span
    context. Default: true.
    https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/trace/v3/zipkin.proto#config-trace-v3-zipkinconfig
## Logging

- `defaultbackend` (required)

    Name of the default backend

- `backends` (required, repeated)

    List of available logging backends
## LoggingBackend

- `name` (required)

    Name of the backend, can be then used in Mesh.logging.defaultBackend or in
    TrafficLogging

- `format` (optional)

    Format of access logs. Placeholders available on
    https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log

- `type` (required)

    Type of the backend (Kuma ships with 'tcp' and 'file')

- `conf` (required)

    Configuration of the backend
## FileLoggingBackendConfig

- `path` (required)

    Path to a file that logs will be written to
## TcpLoggingBackendConfig

- `address` (required)

    Address to TCP service that will receive logs
## Routing

- `localityawareloadbalancing` (optional)

    Enable the Locality Aware Load Balancing

- `zoneegress` (optional)

    Enable routing traffic to services in other zone or external services
    through ZoneEgress. Default: false

