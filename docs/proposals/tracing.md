# Tracing

## Context

There are many tracing technologies in the market. To use tracing in your application you must instrument it with a library, therefore changing underlying technology is expensive (you have to change it in every app).
That's why OpenTracing was created. Instrument your application once and change the backend (tracer) when you want. Around the same time OpenCensus was created which had a same goal.
OpenCensus and OpenTracing have merged into CNCF OpenTelemetry, but the work is still in progress.

Envoy has a [built-in support](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/observability/tracing) for tracing and we want to leverage this support in Kuma.

### Supported backends

Envoy supports following backends: Lightstep, Zipkin,  OpenTracing (with dynamically loaded library), Datadog, OpenCensus (including GCP Stackdriver) and gRPC service for custom implementation.

### Tracing as a Bootstrap Configuration

You can define only one tracing backend for an Envoy and this setting is a part of the bootstrap configuration (the one that CP delivers before the start of DP).
That means that if you want to change it, you have to restart Envoy.

Also if a user deploy Kuma first and then wants to enable tracing, they need to restart Dataplanes.
This is definitely not ideal UX, but it can probably be solved with Envoy's hot restart.

### Application instrumentation

Contrary to common believes, tracing is not drop-in, transparent solution. You still have to instrument your application,
because Envoy has to correlate request that is coming in to Envoy and the App and request that is coming out from the App and Envoy.

What needs to be done from an application perspective is to pass the standard [`x-request-id`](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_conn_man/headers#x-request-id) Envoy header. This way, Envoy can correlate requests and pass tracing headers.
If you want to leverage other backend features like Baggage from OpenTracing you need to use library in your application and pass headers that given backend can understand (for example B3 for Zipkin).  

Additionally, in my opinion the most value from tracing comes from having trace-id in application logs. This means that the application will have to read tracing header anyway and use it in logs. 

The real value from introducing tracing via service mesh is to offload applications from the task of sending traces to the backend of your choice as well as being able to upgrade (or change if you use OpenTracing/OpenCensus) the tracing backend flexibly.

## Requirements

Kuma Dataplane can:
* Deliver traces to the backend of your choice (one chosen backend for the first implementation)
* Generate traces if it did not received any
* Pass traces to the application and other Dataplane

**Out of scope for the first implementation**

* Dynamic restart of Envoys after changing the tracing settings

## Configuration model

Envoy's tracing API is available [here](https://www.envoyproxy.io/docs/envoy/latest/api-v2/config/trace/v2/trace.proto).

Envoy's tracing setting on HTTP Connection manager is available [here](https://www.envoyproxy.io/docs/envoy/latest/api-v2/config/filter/network/http_connection_manager/v2/http_connection_manager.proto#config-filter-network-http-connection-manager-v2-httpconnectionmanager-tracing)

### Zipkin

```yaml
type: Mesh
name: default
tracing:
  defaultBackend: my-zipkin
  backends:
    - name: my-zipkin
      sampling: 10.0 # percentages 0-100
      zipkin:
        url: http://zipkin.local/api/v2/spans
        traceId128bit: false # Generate 128bit traces. Default: false
        apiVersion: httpJson # Pick a version of the API. values: httpJson, httpProto. Default: httpJson
        sharedSpanCotext: true # whether client and server spans share the same span context. Default: true
```

Note: Zipkin can be also supported via OpenCensus. Jaeger is also [compatible](https://www.jaegertracing.io/docs/1.13/features/#backwards-compatibility-with-zipkin) with Zipkin format.

### Datadog

```yaml
type: Mesh
name: default
tracing:
  backends:
    - name: my-datadog
      sampling: 10.0 # percentages 0-100
      datadog:
        address: datadog.address:1234
```

Envoy's Datadog config also requires service name which we could infer from the Dataplane entity.

Note: Datadog can be also supported via OpenTracing and OpenCensus (via Agent)

### Lightstep

To push traces to Lightstep you have to configure Envoy with `access_token_file`.
This means we have to build a system to pass Token stored in CP (probably as a secret) to DP.
I think design for this backend need separate proposal in form of Github issue.

### OpenTracing

```yaml
type: Mesh
name: default
tracing:
  backends:
    - name: my-ot
      sampling: 10.0 # percentages 0-100
      openTracing:
        libraryPath: /usr/local/lib/libinstana_sensor.so
        config: {} # config can be anything as it is specific to the dynamic library 
```

It assumes that every Dataplane in a Mesh will have this library in a given path.
Envoy crashes if there is a problem with loading library (for example file not found).

Jaeger is the most popular tracer of OpenTracing. List of supported tracers in OpenTracing is available [here](https://opentracing.io/docs/supported-tracers/). 

### OpenCensus

```yaml
type: Mesh
name: default
tracing:
  backends:
    - name: my-oc
      sampling: 10.0 # percentages 0-100
      openCensus:
        stackdriver:
          projectId: 1234
          address: # optional
        zipkin:
          url: http://127.0.0.1:9411/api/v2/spans
        agent:
          address: "ipv4:127.0.0.1:345" # https://github.com/grpc/grpc/blob/master/doc/naming.md
        incomingTraceContext: # in what header format trace will be consumed. Default all, Envoy looks for all of them.
          - b3
          - w3c
          - cloud
        outgoingTraceContext: # in what header format trace will be produced (default: b3)
          - b3
```

It seems that you can pick multiple Open Census exporters at once.
Since exporters here are limited to Stackdriver and Zipkin you can send traces to Agent which will send it to all supported backends listed [here](https://opencensus.io/exporters/supported-exporters/go/ocagent/).

### TrafficTrace

Additionally, to enable Tracing, you have to select Dataplanes that tracing will be enabled on.  

```yaml
type: TrafficTrace
mesh: default
name: us
selectors:
  - match:
      zone: us
conf:
  backend: zipkin-us
```

This enables user to segment the destinations of traces.

## Summary

We decided to start with Zipkin backend, and add more backends gradually.
First implementation will require users to restart Envoy to reload configuration, but we want to make contribution
to Envoy and improve it to have dynamic configuration like the rest of the configurations in Kuma.