# Custom Container Configuration

## Introduction
Users and field team have requested access to custom configuration of kubernetes sidecar and init containers.
See [#3925](https://github.com/kumahq/kuma/issues/3925).
This intuitively aligns with the ProxyTemplate configuration,
which provides a free-form configuration of Envoy.
The current plan is to introduce a new CRD,
similar to ProxyTemplate,
which can modify sidecar config utilizing [jsonpatch](https://datatracker.ietf.org/doc/html/rfc6902)

## Custom Resource Definition
The new CRD will be named ContainerPatch.
It will allow for customer configuration of
both sidecar and init containers.
It will be namespace scoped.
The webhook will validate that it can only be applied in a namespace where Kuma CP is running.
The spec will contain an array of jsonpatch strings which describe the modifications to be performed.
There is no Universal mode equivalent,
because there is no sidecar or init container injection in Universal mode.

### Example

```yaml
apiVersion: kuma.io/v1alpha1
kind: ContainerPatch
metadata:
  name: container-patch-1
spec:
  sidecarPatch:
    - op: add
      path: /securityContext/privileged
      value: true
  initPatch:
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
A `ContainerPatch` will be matched to a Pod via an annotation on the workload.
Each annotation may be an ordered list of `ContainerPatch` names,
which will be applied in the order specified.

We will include a configuration option to Kuma CP
to specify a default list of `ContainerPatch` patches to apply.
In KumaCP configuration, there are `sidecarContainer` and `initContainer` sections.
We will add an option under each named `containerPatch`,
which will hold the name of a `ContainerPatch` CRD instance. This will be applied
to any workload which is not annotated with `kuma.io/container-patches`.

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
        kuma.io/container-patches: containter-patch-1,containter-patch-2
    spec: ...
```

## Multizone

We will not add any special handling for multizone.
Different zones are likely to be in different clusters or on disparate platforms,
so it is unlikely they will need the same sets of patches.
To patch containers in multizone, the config will need to be pushed to the appropriate zones.


## Error Modes and Validation

We will validate that the rendered container spec meets the kubernetes specification.
We will not validate that it is a sane configuration.
It is assumed that anyone using this feature has the expertise to create and debug a container configuration
via interaction with kubernetes directly.
If a workload refers to a ContainerPatch which does not exist, the injection will explicitly fail and log the failure.
The rationale behind this is that the user may be adding security features to the workload,
in which case failing open may constitute  a security downgrade.

## Alternative Solutions

Other options we considered but have decided against for now.

* Using a selector within the CRD to identify Pods / dataplanes.
  We chose annotation on workload because it eliminates ambiguity associated with multiple selector matches,
  and is generally easier to understand and configure.
* Full container specification templates
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
  We chose a patch style because it is simpler, less error-prone, and does not require a base template to be acquired by the user.
* Our own style of "patch"
  eg
  ```yaml
  sidecarTemplate:
    - securityContext:
        operation: add
        value: |
          privileged: true
  ```
  We chose jsonpatch because it is a standard and has supporting libraries and documentation.
* Applying CRD in lexicographical order; obviated by choice of annotation instead of selector.
* Something more like Istio's solution.
  These seemed overly complicated compared to simply matching a jsonpatch string to a container.
  * https://istio.io/latest/docs/setup/additional-setup/sidecar-injection/#customizing-injection
  * https://istio.io/latest/docs/setup/additional-setup/sidecar-injection/#custom-templates-experimental
* Apply default patch, followed by annotated patches on the same container. Decided that annotation will have override
  semantics, to make the overall configuration easier to follow.
