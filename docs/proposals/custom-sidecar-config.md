# Custom Container Configuration

## Introduction
Users and field team have requested access to custom configuration of kubernetes sidecar and init containers. See [#3925](https://github.com/kumahq/kuma/issues/3925). This intuitively aligns with the ProxyTemplate configuration, which provides a free-form configuration of Envoy. The current plan is to introduce a new CRD, similar to ProxyTemplate, which can accommodate sidecar config.

## Custom Resource Definition
The new CRD will be named ContainerTemplate. It will allow for customer configuration of both sidecar and init containers.

This will follow a similar format to ProxyTemplate. A ContainerTemplate will match pods via namespaces and labels, and modify the containers associated with those pods. The spec will contain full templates for “sidecar” and “init” dataplane containers.

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
Kuma creates the existing container specifications, in part, using dynamic data known at container creation time. Since the full configuration will be specified in the CRD, we need a method for these dynamic elements to be provided by kuma, but also potentially overridden by the user. We will achieve this by providing the user with a full container template where the dynamic values are specified as template variables. If the user needs to override the value, they can replace the variable with their own literal override. Otherwise, it will be filled in by kuma.

## Error Modes and Validation
We will validate that the rendered template meets the kubernetes specification. We will not validate that it is a sane configuration. It is assumed that anyone using this feature has the expertise to create and debug a container configuration via interaction with kubernetes directly.

If multiple ContainerConfig policies match the same container, the most recently applied policy wins.

