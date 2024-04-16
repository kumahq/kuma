# TCP and gRPC probes

- Status: 

Technical Story: https://github.com/kumahq/kuma/issues/3875

## Context and Problem Statement

Kubernetes support three types of network based probes:
- HTTP Get
- TCP Socket
- gRPC

Kuma transparent proxy captures inbound traffic of pods, including the probe traffic. This leads to unexpected behaviour for Kubernetes probes without special handling: A TCP socket connection will always succeed to establish as it's redirected and handled by inbound passthrough listeners; HTTP requests can fail due to lack of mTLS client certificates attached, so are gRPC requests.

We've implemented the [Virtual Probes feature](https://kuma.io/docs/2.6.x/policies/service-health-probes/#virtual-probes) which creates a dedicated listener for HTTP probes. And now, we want to support TCP socket and gRPC probes as well.

Kubernetes users can specify HTTP, gRPC and TCP probes like this:

```
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: app
    startupProbe:
      httpGet:
        path: /healthz
        port: 8080
    livenessProbe:
      grpc:
        port: 2379
        service: liveness
    readinessProbe:
      tcpSocket:
        host: some-other-host
        port: 1433
```

The `host` field in the `tcpSocket` probe is only used to resolve the network address when one wants to probe a different host than the pod's IP address and it's optional. The `service` field on gRPC probes is also optional. There is little space for us to carry contextual information on them.

A pod can have multiple containers with each defines multiple and different probes, the worst case is a pod has all 3 different types of probes with many probes for each type.

### HTTP Virtual Probes (existing design)

As per our current design, a pod with multiple probes will be merged and will be handled by a single virtual listener. This listener will forward/redirect probe requests to the corresponding application port without any intermediate transforming except path rewriting, etc. 

While we provided annotations for users to customize whether they want to enable the Virtual Probe feature and on which port do they want to expose the Virtual Probe listener, the current design actually does not require the user to customize these items and the probes can be translated automatically and transparently.

The following example demonstrates how HTTP probes are handled in current implementation. Two probes are both handled by the Virtual Probe listener on port 19000:

[inject.14.input.yaml](https://github.com/kumahq/kuma/blob/master/pkg/plugins/runtime/k8s/webhooks/injector/testdata/inject.14.input.yaml)

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: busybox
  labels:
    run: busybox
  annotations:
    kuma.io/virtual-probes-port: "19000"
spec:
  containers:
    - name: busybox
      image: busybox
      resources: {}
      readinessProbe:
        httpGet:
          path: /metrics
          port: 3001
        initialDelaySeconds: 3
        periodSeconds: 3
      livenessProbe:
        httpGet:
          path: /metrics
          port: 8080
        initialDelaySeconds: 3
        periodSeconds: 3
```

[inject.14.golden.yaml](https://github.com/kumahq/kuma/blob/master/pkg/plugins/runtime/k8s/webhooks/injector/testdata/inject.14.golden.yaml)

```yaml
apiVersion: v1
kind: Pod
metadata:
  annotations:
    kuma.io/virtual-probes: enabled
    kuma.io/virtual-probes-port: "19000"
    ...
  labels:
    run: busybox
  name: busybox
spec:
  containers:
  - name: kuma-sidecar
    image: kuma/kuma-sidecar:latest
    ...
  - image: busybox
    livenessProbe:
      httpGet:
        path: /8080/metrics
        port: 19000
      initialDelaySeconds: 3
      periodSeconds: 3
    name: busybox
    readinessProbe:
      httpGet:
        path: /3001/metrics
        port: 19000
      initialDelaySeconds: 3
      periodSeconds: 3
  initContainers:
  - name: kuma-init
    image: kuma/kuma-init:latest
  ...
```

### TCP and gRPC Virtual Probes (proposed design)

A pod can have multiple containers with each defines different probes, the worst case is a pod has all 3 different types of probes with many probes for each type.

Introduce a new annotation `kuma.io/virtual-probes-port-mapping` to let user specify the ports to use for probes, and deprecate existing annotations.

A user can now specify the Virtual Probe ports like this:

```yaml
apiVersion: v1
kind: Pod
metadata:
  annotations:
    kuma.io/virtual-probes: "enabled"
    kuma.io/virtual-probes-port-mapping: '{"http": 19000, "grpc-2379": 19001, "tcp-1433": 19002}'
spec:
  containers:
  - name: app
    startupProbe:
      httpGet:
        path: /healthz
        port: 8080
    livenessProbe:
      grpc:
        port: 2379
        service: liveness
    readinessProbe:
      tcpSocket:
        port: 1433
```

In which:
- `http` is the port to be used for all the HTTP probes, replacing the existing annotation `kuma.io/virtual-probes-port`
- `grpc-*` is the ports to be used for each of gRPC probes
- `tcp-*` is the ports to be used for each of TCP probes

If a probe is not specified in the `kuma.io/virtual-probes-port-mapping`, we'll generate the port automatically by incrementing from the default port (`9000`), skipping the known and taken ports.

### Virtual Probes Listeners

We generate a TCP proxy listener for each of gRPC and TCP probes, and forward the probe traffic transparently to the application. We also keep the existing HTTP Virtual Probe listener for HTTP probes.

This keep the compatibility with the existing HTTP based Virtual Probe and data plane data model.

### Data Plane Data Model 

Introduce two new fields onto the `Dataplane` resource to store the Virtual Probe ports:

```yaml
  probes:
    endpoints:
    - inboundPath: /metrics
      inboundPort: 8081
      path: /8081/metrics
    port: 19000
  tcpProbes:
  - port: 19001
    applicationPort: 1433
  grpcProbes:
  - port: 19002
    service: abcd
    applicationPort: 2379
```

## Decision Drivers

- Provide the best user experience by designing a simple and intuitive way of declaring probes 
- Keep the compatibility with the existing HTTP based Virtual Probe and data plane data model

## Considered Options

- Allocate a new and separated port for each of the TCP probes or gRPC probes

- Use one port to support all the three types of probes

## Decision Outcome

Chosen option: 

(TODO)

### Positive Consequences

(TODO)

### Negative Consequences

(TODO)

## Other option

### Merging all probes into one listener

(TODO)

### Positive Consequences

(TODO)

### Negative Consequences

(TODO)
