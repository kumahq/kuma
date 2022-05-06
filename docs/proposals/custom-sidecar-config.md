# Custom Container Configuration

## Introduction
Users and field team have requested access to custom configuration of kubernetes sidecar and init containers.
See [#3925](https://github.com/kumahq/kuma/issues/3925).
This intuitively aligns with the ProxyTemplate configuration,
which provides a free-form configuration of Envoy.
The current plan is to introduce a new CRD,
similar to ProxyTemplate,
which can accommodate sidecar config.

## Custom Resource Definition
The new CRD will be named ContainerTemplate.
It will allow for customer configuration of
both sidecar and init containers.

This will follow a similar format to ProxyTemplate.
A ContainerTemplate will match pods via namespaces and labels,
and modify the containers associated with those pods.
The spec will contain full templates for “sidecar” and “init” dataplane containers.

### Example

```yaml
apiVersion: kuma.io/v1alpha1
kind: ContainerTemplate
metadata:
  name: container-template-1
spec:
  selectors:
    - namespace: namespace-1
  sidecarTemplate: | …
  initTemplate: | …
```

## Dynamic Configuration
Kuma creates the existing container specifications, in part,
using dynamic data known at container creation time.
Since the full configuration will be specified in the CRD,
we need a method for these dynamic elements to be provided by kuma,
but also potentially overridden by the user.
We will achieve this by providing the user with a full container template
where the dynamic values are specified as template variables.
If the user needs to override the value,
they can replace the variable with their own literal override.
Otherwise, it will be filled in by kuma.

## Error Modes and Validation
We will validate that the rendered template meets the kubernetes specification.
We will not validate that it is a sane configuration.
It is assumed that anyone using this feature has the expertise to create and debug a container configuration
via interaction with kubernetes directly.

If multiple ContainerTemplate policies match the same container,
the most recently applied policy wins.

## Templates
Template specification will be the full container specification,
with variables for Kuma to fill with default values.
Ideally, the user can modify individual values,
and upgrades which affect unrelated values will just fill in those variables
without user intervention.
Larger changes across versions will require user to integrate their changes into a new template.

### Example - Init Container Default

```yaml
apiVersion: kuma.io/v1alpha1
kind: ContainerTemplate
metadata:
  name: container-template-1
spec:
  selectors:
    - namespace: namespace-1
  initTemplate: |
    name: kuma-init
    command: {{ .Command }}
    args: {{ .CommandArgs }}
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
      runAsUser: {{ .SecurityContext.RunAsUser }}
    terminationMessagePath: {{ .TerminationMessagePath }}
    terminationMessagePolicy: {{ .TerminationMessagePolicy }}
    volumeMounts: {{ .VolumeMounts }}
```

### Example - Init Container RunAsNonRoot

```yaml
apiVersion: kuma.io/v1alpha1
kind: ContainerTemplate
metadata:
  name: container-template-1
spec:
  selectors:
    - namespace: namespace-1
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

### Example - Sidecar Container Default

```yaml
apiVersion: kuma.io/v1alpha1
kind: ContainerTemplate
metadata:
  name: container-template-1
spec:
  selectors:
    - namespace: namespace-1
  sidecarTemplate: |
    name: kuma-sidecar
    args: {{ .Args }}
    env: {{ .Env }}
    image: {{ .Image }}
    imagePullPolicy: {{ .ImagePullPolicy }}
    livenessProbe: {{ .LivenessProbe }}
    readinessProbe: {{ .ReadinessProbe }}
    resources:
      limits:
        cpu: {{ .Resources.Limits.Cpu }}
        memory: {{ .Resources.Limits.Memory }}
      requests:
        cpu: {{ .Resources.Requests.Cpu }}
        memory: {{ .Resources.Requests.Memory }}
    securityContext:
      runAsGroup: {{ .SecurityContext.RunAsGroup }}
      runAsUser: {{ .SecurityContext.RunAsUser }}
    terminationMessagePath: {{ .TerminationMessagePath }}
    terminationMessagePolicy: {{ .TerminationMessagePolicy }}
    volumeMounts: {{ .VolumeMounts }}
```

### Example - Sidecar Container Privileged Modification

```yaml
apiVersion: kuma.io/v1alpha1
kind: ContainerTemplate
metadata:
  name: container-template-1
spec:
  selectors:
    - namespace: namespace-1
  sidecarTemplate: |
    name: kuma-sidecar
    args: {{ .Args }}
    env: {{ .Env }}
    image: {{ .Image }}
    imagePullPolicy: {{ .ImagePullPolicy }}
    livenessProbe: {{ .LivenessProbe }}
    readinessProbe: {{ .ReadinessProbe }}
    resources:
      limits:
        cpu: {{ .Resources.Limits.Cpu }}
        memory: {{ .Resources.Limits.Memory }}
      requests:
        cpu: {{ .Resources.Requests.Cpu }}
        memory: {{ .Resources.Requests.Memory }}
    securityContext:
      runAsGroup: {{ .SecurityContext.RunAsGroup }}
      runAsUser: {{ .SecurityContext.RunAsUser }}
      privileged: true
    terminationMessagePath: {{ .TerminationMessagePath }}
    terminationMessagePolicy: {{ .TerminationMessagePolicy }}
    volumeMounts: {{ .VolumeMounts }}
```
