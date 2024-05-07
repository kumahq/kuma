# gRPC and TCP probes

- Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/3875

## Context and Problem Statement

Kubernetes support three types of network based probes:
- HTTP Get
- TCP Socket
- gRPC

Kuma transparent proxy captures inbound traffic of pods, including the probe traffic. This leads to unexpected behaviour for Kubernetes probes without special handling: gRPC requests can fail due to lack of mTLS client certificates attached.

We've implemented the [Virtual Probes feature](https://kuma.io/docs/2.6.x/policies/service-health-probes/#virtual-probes) which creates a dedicated listener for HTTP probes. And now, we want to support gRPC probes as well. We'll create a separated MADR for the supporting of TCP probes.

Kubernetes users can specify HTTP and gRPC probes like this:

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
```

When a gRPC probe request is being sent from kubelet, it will:
- include the port as part of the underlying HTTP `Host` header 
- include the `service` as part of gRPC message body

A pod can have multiple containers with each defineing multiple and different probes, the worst case is a pod has all the different types of probes with many probes for each type. 

The key design goal is to solve these two problems:

1. Support the probing capability
2. Differ multiple probes with different types and ports

## Existing design: HTTP Virtual Probes

As per our current design, a pod with multiple probes will be merged and will be handled by a single virtual listener. This listener will forward/redirect probe requests to the corresponding application port without any intermediate modification except path rewriting, etc.

While we provided annotations for users to customize whether they want to enable the Virtual Probe feature and on which port do they want to expose the Virtual Probe listener, the current design actually does not require the user to customize these items and the probes can be translated automatically and transparently, they only need to customize the port when the default probe port (`9000`) in Kuma is conflicting with the application port.

The following example demonstrates how HTTP probes are handled in current implementation.

The original Pod:

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

Pod with sidecar injected:

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

And in the converted `Dataplane` object, we can see the two probes are both handled by the same Virtual Probe listener on port 19000:

```yaml
type: Dataplane
mesh: default
metadata:
  creationTimestamp: null
spec:
  probes:
    endpoints:
      - inboundPath: /metrics
        inboundPort: 8080
        path: /8080/metrics
      - inboundPath: /metrics
        inboundPort: 3001
        path: /3001/metrics
    port: 19000
  ...
```

## Considered Options

**Option 1:** Transform gRPC and TCP probes to HTTP probes, and merge into the existing HTTP based Virtual Probes listener

**Option 2:** Allocate a new and separated port for each of the gRPC probes, and route them back to the application probe handling port accordingly.

### Proposed option: option 1 - Transform gRPC probes to HTTP probes, merge all the HTTP probes by one HTTP listener

Transform user defined gRPC probes into HTTP probes first and then use the existing Virtual Probe HTTP listener to handle incoming probe traffic. This needs n new intermediate layer translating HTTP probes back into gRPC probes (the translator), a reasonable place is put it in the kuma-dp.

Handling incoming probe request from kubelet:

1. If the user defines an HTTP probe, requests from kubelet are forwarded/redirected to application directly with necessary path/port rewriting, etc.

2. If the user defines a gRPC probe, we'll also receive HTTP requests from kubelet, since they are transformed to HTTP probes. We forward these requests to the translator and the translator is responsible for handling the probe request, by detecting the actual status from the application.

The whole workflow will be like the following diagram:

```
  kubelet (HTTP Probes)                                         Application (HTTP Probes)
               ↘                                                ↗ 
                 Pod (HTTP Probes)  --->  Virtual Probe Listener (HTTP routing)
               ↗                                                ↘
  kubelet (gRPC Probes)                                  HTTP to gRPC translator  --->  Application (gRPC Probes)
```

#### Positive Consequences

- Don't need to introduce new ports for gRPC probes

- Better user experience: don't need to introduce new annotations, the user can still specify the port as they are doing now.

- Less virtual listeners on sidecar

- Don't need to introduce new fields on the `Dataplane` resource type

- The translator may be partly reused to support TCP probes

#### Negative Consequences

- Requires an extra component (the translator) in `kuma-dp`, which increases the implementation complexity

- Less intuitiveness, harder for troubleshooting issues

## Decision Drivers

- Provide the best user experience by designing a simple and intuitive way of declaring virtual probes

- Keep the compatibility with the existing HTTP based Virtual Probe and data plane data model

## Decision Outcome

Chosen option: option 1

## Other options

### Option 2 - allocate a new and separated port for each of the gRPC probes, and route them back to the application probe handling port accordingly

The only information we can use to differ from multiple gRPC probes is the port number, while the port number is included in the underlying HTTP2 request as the `Host` header, it will be the same if we want to use a single virtual probe listener to support forwarding traffic for all gRPC probes. So we need to allocate a new virtual probe port for each of them.

To allocate these ports, we'll need a new Pod annotation to support user specifying the range to be used. This annotation, with name `kuma.io/virtual-probes-port-range`, will replace the existing annotation `kuma.io/virtual-probes-port`. The annotation should be optional: when user does not specify, a default range `9000-9020` will be used.

gRPC probe requests are actually HTTP2 requests under the hood, we can use mTLS disabled HTTP listeners to forward requests prefixed with `/grpc.health.v1.Health/` back to application ports. Transparent TCP proxies should not be used as they will make the whole application lose the protection from mTLS. With separated virtual probes listeners, there is no need to introduce extra fields on the `Dataplane` to capture properties of user defined gRPC probes, except the port number. To unify data structure, we can share the parent level property with TCP probes as `tcpProbes`.

The updated `Dataplane` would be like this:

```yaml
type: Dataplane
mesh: default
metadata:
  creationTimestamp: null
spec:
  probes:
    endpoints:
      - inboundPath: /metrics
        inboundPort: 8080
        path: /8080/metrics
      - inboundPath: /metrics
        inboundPort: 3001
        path: /3001/metrics
    port: 19000
  tcpProbes:
  - port: 2379
  ...
```

#### Positive Consequences

- Ease of implementation

- Keeping the compatibility with the existing HTTP based Virtual Probe and data plane data model

#### Negative Consequences

- The user experience of specifying virtual probes ports is more complex
