# gRPC and TCP probes

- Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/3875

## Context and Problem Statement

Kubernetes support three types of network based probes:
- HTTP Get
- TCP Socket
- gRPC

Kuma transparent proxy captures inbound traffic of pods, including the probe traffic. This leads to unexpected behaviour for Kubernetes probes without special handling: gRPC requests can fail due to lack of mTLS client certificates attached.

We've implemented the [Virtual Probes feature](https://kuma.io/docs/2.6.x/policies/service-health-probes/#virtual-probes) which creates a dedicated listener for HTTP probes. And now, we want to support gRPC and TCP probes as well.

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

**Option 1:** Transform gRPC and TCP probes to HTTP probes, use an intermediate Virtual Probe Handler to handle all incoming probe traffic.

**Option 2:** Allocate a new and separated port for each of the gRPC and TCP probes, and route them back to the application probe handling port accordingly.

### Proposed option: option 1 - Transform gRPC and TCP probes to HTTP probes, use an intermediate Virtual Probe Handler to handle all incoming probe traffic.

Transform user defined gRPC and TCP probes into HTTP probes first, create a HTTP based Virtual Probe Handler in `kuma-dp` to handle all incoming probe traffic sent from `kubelet`. This handler translates HTTP probes back into original format (gRPC, TCP or user defined HTTP probes) and eventually be handled by the application.

The Virtual Probe Handler will handle incoming probe requests like this:

1. If the user defines an HTTP probe, requests from kubelet are rewritten back to the user defined paths and forwarded to the application port.

2. If the user defines a gRPC probe, the Virtual Probe Handler will also receive HTTP requests from kubelet, since they are transformed to HTTP probes. The handler will then detect the actual status from the application by sending gRPC health check requests and respond to HTTP requests accordingly.

3. If the user defines a TCP probe, The handler will then detect the actual status from the application by establishing connections to the application port and respond to HTTP requests accordingly.

The whole workflow can be described as the following diagram, the column "Transformed" will be put into pod manifest after the Kuma sidecar/init containers are injected into a pod:

```
  User Defined      Transformed            Virtual Probe Handler                Application Endpoint   


  TCP Probes                                         HTTP to TCP translator  -->   TCP server
               ↘                                    ↗ 
  HTTP Probes -->  Pod (HTTP Probes)  -->  kuma-dp --> HTTP to HTTP rewriter -->  HTTP handler
               ↗                                    ↘
  gRPC Probes                                        HTTP to gRPC translator -->  gRPC healthcheck service
```

By also using this same Virtual Probe Handler in `kuma-dp` to handle user defined HTTP probes, we remove the requirements to generate a Virtual Probe listener within Envoy. As a side effect, it adds a requirement to exclude this port of the "Virtual Probe Handler" from the traffic redirection mechanism.

In this way, we can unify the handling of all user defined probes, and removes the `probes` field from the `Dataplane` object.

#### Positive Consequences

- Don't need to introduce new port for virtual probes

- Better user experience: don't need to introduce new annotations, the user can still specify the port as they are doing now.

- Less virtual listeners on sidecars: the Virtual Probe Handler will be used to handle all three types of probes

- Simplifies the model by removing the `probes` field, and don't need to introduce new fields on the `Dataplane` object,

#### Negative Consequences

- Requires an extra component (the Virtual Probe Handler) in `kuma-dp`, which increases the architecture complexity

- Less intuitiveness, harder for troubleshooting issues

#### Implementation details

The probes support can be implemented in these steps:

1. Create a new component in `kuma-dp` (the Virtual Probe Handler) to handle the incoming HTTP probes by performing the actual probes and returning the result
2. Make the sidecar injector to:
   1. Get the correct port for the Virtual Probe Handler, and generate corresponding command line arguments for `kuma-dp` according to the user defined probes and the Virtual Probe port
   2. Exclude the Virtual Probe port by including it into `kuma-init` command line arguments or by excluding it in `kuma-cni`
3. Remove the existing Virtual Probe listener (generated in `probe_generator`) in Envoy
4. Deprecate the `probes` field in the `Dataplane` type

In step 1, the Virtual Probe Handler is an HTTP server and needs to listen on a port. We can use the existing Virtual Probe port, since it will not be used in Envoy anymore. It defaults to `9000` and will be configurable by `runtime.kubernetes.injector.virtualProbesPort`. The user is not expected to change this port unless it is conflicting with an application port. 

A pod with multiple probes defined will be converted by the injector like this:

Original pod manifest:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: busybox
  labels:
    run: busybox
  annotations:
    traffic.kuma.io/exclude-inbound-ports: "9800"
spec:
  containers:
    - name: busybox
      image: busybox
      readinessProbe:
        httpGet:
          path: /healthz
          port: 6851
        initialDelaySeconds: 3
        periodSeconds: 3
      livenessProbe:
        grpc:
          port: 6852
          service: liveness 
        initialDelaySeconds: 3
        periodSeconds: 3
      startupProbe:
        tcpSocket:
          port: 6853
        initialDelaySeconds: 3
        periodSeconds: 3
```

Injected pod manifest:

```yaml
apiVersion: v1
kind: Pod
metadata:
  annotations:
    ...
    kuma.io/virtual-probes: enabled
    # this will be injected by the injector and may be consunmed by kuma-cni 
    kuma.io/virtual-probes-port: "9000"
    traffic.kuma.io/exclude-inbound-ports: "9800"
  labels:
    run: busybox
  name: busybox
spec:
  initContainers:
  - args:
    - --config-file
    - /tmp/kumactl/config
    - --redirect-outbound-port
    - "15001"
    - --redirect-inbound=true
    - --redirect-inbound-port
    - "15006"
    - --kuma-dp-uid
    - "5678"
    - --exclude-inbound-ports
    # 9000 is the virtual probe port and appended by the injector
    - "9000,9800"
    - --exclude-outbound-ports
    - ""
    - --verbose
    - --ip-family-mode
    - dualstack
    command:
    - /usr/bin/kumactl
    - install
    - transparent-proxy
    ...
  containers:
  - args:
    - run
    - --log-level=info
    - --concurrency=2
    env:
    # this is converted from user defined probes and be injected by the injector 
    - name: KUMA_DATAPLANE_PROBES
      value: "[{\"httpGet\":{\"path\":\"/healthz\",\"port\":6852}},{\"grpc\":{\"port\":3636,\"service\":\"liveness\"}},{\"tcpSocket\":{\"port\":6826}}]"
    - name: INSTANCE_IP
      valueFrom:
        fieldRef:
          apiVersion: v1
          fieldPath: status.podIP
    ...
    image: kuma/kuma-sidecar:latest
    imagePullPolicy: IfNotPresent
    ...
  - image: busybox
    name: busybox
    # these are transformed from user defined probes by the injector
    readinessProbe:
      httpGet:
        path: /6851/healthz
        port: 9000
      initialDelaySeconds: 3
      periodSeconds: 3
    livenessProbe:
      httpGet:
        path: /grpc/6852/liveness
        port: 9000 
      initialDelaySeconds: 3
      periodSeconds: 3
    startupProbe:
      httpGet:
        path: /tcp/6853
        port: 9000
      initialDelaySeconds: 3
      periodSeconds: 3
```


The `Dataplane` type definition will be updated, here is an example:

```yaml
type: Dataplane
name: backend-1
mesh: default
spec:
  networking:
    address: 192.168.0.2
    inbound:
     - port: 8080
       tags:
          kuma.io/service: backend
  # this field will be deprecated
  # probes:
  #  endpoints:
  #    - inboundPath: /healthz
  #      inboundPort: 6851
  #      path: /6851/healthz
  #  port: 9000
```

In the probe Virtual Probe Handler, it will:

- return `503` responses for failed probes, return `200` responses for successful probes
- use the pod IP as the `Host` header in the gRPC request to the application
- use socket options `SO_LINGER` and timeout 0 to close the connection immediately on detecting TCP connections

## Decision Drivers

- Provide the best user experience by designing a simple and intuitive way of declaring virtual probes

- Keep the compatibility with the existing HTTP based Virtual Probe and data plane data model

## Decision Outcome

Chosen option: option 1

## Other options

### Option 2 - allocate a new and separated port for each of the gRPC/TCP probes, and route them back to the application probe handling port accordingly

The only information we can use to differ from multiple gRPC probes is the port number, while the port number is included in the underlying HTTP2 request as the `Host` header, it will be the same if we want to use a single virtual probe listener to support forwarding traffic for all gRPC probes. So we need to allocate a new virtual probe port for each of them. TCP probes share the same scenario, so we also need to allocate a new port for each of them.

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
  # both gRPC and TCP probes are generated as "tcpProbes" in the Dataplane object
  tcpProbes:
  - port: 2379
  - port: 3476
  ...
```

#### Positive Consequences

- No translation between user defined probes and actual probes, so it's intuitive enough for issue troubleshooting

- Keeping the compatibility with the existing HTTP based Virtual Probe and data plane data model

#### Negative Consequences

- The user experience of specifying virtual probes ports is more complex
