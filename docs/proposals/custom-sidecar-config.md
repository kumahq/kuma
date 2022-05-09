# Custom Container Configuration

## Introduction
Users and field team have requested access to custom configuration of kubernetes sidecar and init containers.
See [#3925](https://github.com/kumahq/kuma/issues/3925).
This intuitively aligns with the ProxyTemplate configuration,
which provides a free-form configuration of Envoy.
The current plan is to introduce a new CRD,
similar to ProxyTemplate,
which can modify sidecar config utilizing [jsonpatch](https://jsonpatch.com).

## Custom Resource Definition
The new CRD will be named ContainerTemplate.
It will allow for customer configuration of
both sidecar and init containers.
It will be namespace scoped.
The webhook will validate that it can only be applied in a namespace where Kuma CP is running.
The spec will contain an array of jsonpatch strings which describe the modifications to be performed.
There is no Universal mode equivalent.

### Example

```yaml
apiVersion: kuma.io/v1alpha1
kind: ContainerTemplate
metadata:
  name: container-template-1
spec:
  sidecarTemplate:
    - op: add
      path: /securityContext/privileged
      value: true
  initTemplate:
    - op: add
      path: /securityContext/runAsNonRoot
      value: true
    - op: remove
      path: /securityContext/runAsUser

```

This will change the `securityContext` section of `kuma-sidecar` container from

```yaml
      securityContext:
        runAsGroup: 5678
        runAsUser: 5678
```

to


```yaml
      securityContext:
        runAsGroup: 5678
        runAsUser: 5678
        priviledged: true
```

and similarly change the `securityContext` section of the `init` container from

```yaml
      securityContext:
        capabilities:
          add:
          - NET_ADMIN
          - NET_RAW
        runAsGroup: 0
        runAsUser: 0
```

to

```yaml
      securityContext:
        capabilities:
          add:
          - NET_ADMIN
          - NET_RAW
        runAsGroup: 0
        runAsNonRoot: true
```

## Workload Matching
A `ContainerTemplate` will be matched to a Pod via an annotation on the workload.
Each annotation may be an ordered list of `ContainerTemplate` names,
which will be applied in the order specified.

We will include a configuration option to Kuma CP
to specify a default list of `ContainerTemplate` patches to apply.

### Example

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: app-ns
  name: app-depl
spec:
  replicas: 1
  selector:
    matchLabels:
      app: app-depl
  template:
    metadata:
      labels:
        app: app-depl
      annotations:
        kuma.io/container-template:
          - containter-template-1
          - containter-template-2
    spec: ...
```

## Multizone

We will not add any special handling for multizone.
To patch containers in multizone, the config will need to be pushed to the appropriate zones.

## Error Modes and Validation

We will validate that the rendered template meets the kubernetes specification.
We will not validate that it is a sane configuration.
It is assumed that anyone using this feature has the expertise to create and debug a container configuration
via interaction with kubernetes directly.

## Alternative Solutions

Other options we considered but have decided against for now.

* Using a selector within the CRD to identify Pods / dataplanes; chose annotation on workload.
* Full container specification templates; chose jsonPatch.
  eg
  ```yaml
  initTemplate: |
    name: kuma-init
    command: {{ .Command }}
    args: {{ .Args }}
    image: {{ .Image }}
    imagePullPolicy: {{ .ImagePullPolicy }}
    resources:
      limits:
        cpu: {{ .Resources.Limits.Cpu }}
        memory: {{ .Resources.Limits.Memory }}
      requests:
        cpu: {{ .Resources.Requests.Cpu }}
        memory: {{ .Resources.Requests.Memory }}
    securityContext:
      capabilities:
        add:
        - NET_ADMIN
        - NET_RAW
      runAsGroup: {{ .SecurityContext.RunAsGroup }}
      runAsNonRoot: true
    terminationMessagePath: {{ .TerminationMessagePath }}
    terminationMessagePolicy: {{ .TerminationMessagePolicy }}
    volumeMounts: {{ .VolumeMounts }}
  ```
* Our own style of "patch"; chose jsonPatch.
  eg
  ```yaml
  sidecarTemplate:
    - securityContext:
        operation: add
        value: |
          privileged: true
  ```
* Applying CRD in lexicographical order; obviated by choice of annotation instead of selector.
* Something more like Istio's solution
  * https://istio.io/latest/docs/setup/additional-setup/sidecar-injection/#customizing-injection
  * https://istio.io/latest/docs/setup/additional-setup/sidecar-injection/#custom-templates-experimental
