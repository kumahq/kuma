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

## Existing design: HTTP Virtual Probes

As per our current design, a pod with multiple probes will be merged and will be handled by a single virtual listener. This listener will forward/redirect probe requests to the corresponding application port without any intermediate transforming except path rewriting, etc.

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

**Option 1:** Allocate a new and separated port for each of the TCP probes, and skip these ports when setting up the transparent proxy iptables rules

**Option 2:** Transform TCP probes to HTTP probes, merging all the HTTP based probes by one HTTP listener, split traffic and forward to corresponding probe handling port. 

### Option 2: Transform TCP probes to HTTP probes, merging all the HTTP probes by one HTTP listener (proposed)

Use the existing virtual probe HTTP listener to support both HTTP, gRPC probes. 

For TCP probes, transform them into HTTP probes first and then also use the same HTTP virtual probe HTTP listener to handle incoming probe probe traffic.  This needs an new intermediate layer translating HTTP probes back into TCP probes (the translator), a reasonable place to is put it in the kuma-dp.

So:

1. If the user probe is HTTP based, requests are forwarded/redirected to application directly with necessary path/port rewriting, etc.

2. If the user probe is TCP based, we'll also receive HTTP requests for these probes, since they are transformed to HTTP probes. We forward these requests to the translator and the translator is responsible for detecting the actual probe status from the application.

The whole workflow can be shown in the following diagram:

```
  gRPC Probes                                                              Application
              ↘                                                          ↗
HTTP Probes --> HTTP Probes --->  Virtual Probe Listener ->  HTTP Probes  --->  Application
              ↗                                           ↘
   TCP Probes                                               HTTP to TCP translator  --->  Application
```

#### Positive Consequences

- Don't need to introduce new ports for TCP and gRPC probes

- Better user experience: don't need to introduce new annotations, the user can still specify the port as they are doing now. 

- Less virtual listeners on sidecar

- Don't need to introduce new fields on the `Dataplane` resource type

#### Negative Consequences

- Requires an extra component (the translator) in `kuma-dp`, which increases the implementation complexity

- Less intuitiveness, so harder for troubleshooting issues

## Other options

### Allocate a new and separate port for each of TCP probe ports and skip them when setting up the transparent proxy iptables rules

The TCP probe mechanism is broken by the transparent proxy redirection, when we skip redirecting traffic to these virtual probe ports, the probe traffic will go to the application directly. 

#### Positive Consequences

- Ease of implementation

- Keeping the compatibility with the existing HTTP based Virtual Probe and data plane data model

#### Negative Consequences

- The user experience of specifying virtual probes ports is more complex

## Decision Drivers

- Provide the best user experience by designing a simple and intuitive way of declaring virtual probes

- Keep the compatibility with the existing HTTP based Virtual Probe and data plane data model

## Decision Outcome

Chosen option: (to be chosen)