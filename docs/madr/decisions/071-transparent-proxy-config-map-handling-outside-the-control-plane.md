# Transparent Proxy ConfigMap Handling Outside the Control Plane

* Status: accepted

## Context and problem statement

In version 2.9, we introduced an experimental feature to configure the transparent proxy using a Kubernetes ConfigMap. This version had a major limitation: the control plane needed access to ConfigMaps in *all* namespaces, which some users didn’t allow.

There are three components outside the control plane that rely on the transparent proxy configuration:

- `kuma-init` - an init container injected alongside the workload container to install the transparent proxy
- `kuma-cni` - our custom CNI plugin that installs the transparent proxy
- `kuma-sidecar` - the data plane proxy injected next to the workload container

The transparent proxy config is built from default values, control plane configuration, annotations, and ConfigMaps. Since ConfigMaps are namespace-scoped, the component that builds the final config must access all of these inputs. The simplest way to do that is in the control plane, which then passes the final config to `kuma-init`, `kuma-cni`, and `kuma-sidecar`. But this requires the control plane to access all workload namespaces, which led to expanding its ClusterRole permissions.

This document proposes removing that requirement for setups that don’t use `kuma-cni`. Instead of having the control plane build the full config, each data plane component (`kuma-init` and `kuma-sidecar`) will build its own.

## Out of Scope

`kuma-cni` is deployed in the `kube-system` namespace. Without expanding its permissions to access ConfigMaps from other namespaces, it cannot support custom ConfigMaps specified at the workload level. With the proposed changes, where config building moves outside of the control plane, `kuma-cni` will no longer be able to consume workload-specific ConfigMaps at all. This document **_WON'T_** cover the changes required to support such functionality in `kuma-cni`. As a result, the transparent proxy configuration via ConfigMaps will be incomplete and backward incompatible when using `kuma-cni`.

## Problems

### Control plane needs the config to build the Dataplane object

There are two control plane components that require transparent proxy configuration:

- **Sidecar Injector** - injects `kuma-sidecar` and `kuma-init` containers into Pods that are part of the mesh
- **Pod Reconciler** - creates and updates the `Dataplane` object based on the Pod

The `Dataplane` object includes transparent proxy settings like `ipFamilyMode`, `redirectPortInbound`, and `redirectPortOutbound`. These are used by the control plane to generate the xDS config delivered to `kuma-sidecar`.

Sidecar injection and Dataplane generation are handled by separate components, with no shared state. This means there is no direct way to pass configuration between them.

### Each component gets config in a different way

- `kuma-init`: gets full config as a CLI argument
- `kuma-sidecar`: gets config as environment variables (`KUMA_DNS_ENABLED`, `KUMA_DNS_CORE_DNS_PORT`)

This inconsistency makes it hard to track where values are coming from.

## Solution

We will stop using the transparent proxy config fields in the Dataplane object and mark them as deprecated. These fields will no longer be used by `kuma-dp`. Instead, the same values will be passed as metadata during the xDS connection, using configuration provided via CLI flags.

During sidecar injection, the control plane will:
1. Merge default values from the control plane config and the ConfigMap in `kuma-system` namespace
2. Apply any Pod-level annotations
3. Add only the values that differ from defaults to a single annotation: `traffic.kuma.io/transparent-proxy-config`

We avoid using environment variables for this config. Env vars take priority when parsed and can't be safely set if the control plane doesn’t have access to the workload’s namespace. Annotations provide better control over merge order and allow overrides per workload.

The resulting config will be mounted into containers at `/tmp/transparent-proxy/default/config.yaml`.

If the Pod includes the `traffic.kuma.io/transparent-proxy-configmap-name` annotation, the specified ConfigMap will be mounted at `/tmp/transparent-proxy/custom`. Since the ConfigMap is required to include a `config.yaml` key, the actual file path available in the container will be `/tmp/transparent-proxy/custom/config.yaml`.

Mounted configuration will then be passed to `kuma-sidecar` and `kuma-init` using appropriate CLI flags.

### Compatibility

Compatibility will be maintained in cases where the control plane is newer than the data plane proxy. Until the feature described in this document becomes Generally Available and enabled by default, the control plane will support both the existing and new ways of configuring and delivering transparent proxy settings. The new method (using `--transparent-proxy-config` and Dataplane Metadata) will take precedence over values set directly in the Dataplane resource.

However, compatibility is not guaranteed when the data plane proxy is newer than the control plane, such as during a control plane downgrade. If a user wants to downgrade to a version that does not support the feature described here, they must first disable it on the control plane, then restart all data plane proxies, and only after that proceed with the downgrade.

### CLI flags

`kuma-dp` will support a new `--transparent-proxy-config` flag. It can be repeated and accepts the following values:

- No value: enables the transparent proxy with default settings
- Comma-separated paths to one or more config files
- A `-` to read YAML from STDIN

Values are merged in the order provided. Later values override earlier ones. Example:

```sh
CONFIG2=$(mktemp)
CONFIG3=$(mktemp)

echo "{ redirect: { inbound: { port: 2222 } }, wait: 2 }" > "$CONFIG2"
echo "{ redirect: { inbound: { port: 3333 } }, waitInterval: 3 }" > "$CONFIG3"

echo "{ redirect: { inbound: { port: 1111 } }, ipFamilyMode: ipv4 }" | kuma-dp run \
  --transparent-proxy-config "$CONFIG2,$CONFIG3" \
  --transparent-proxy-config -
```

The final configuration will be:

```yaml
ipFamilyMode: ipv4
redirect:
  inbound:
    port: 1111
wait: 2
waitInterval: 3
```

In this case, `redirect.inbound.port` appears in all inputs. Since STDIN is last, its value (`1111`) takes precedence.

For convenience, we’ll also support a `--transparent-proxy` flag as an alias for `--transparent-proxy-config`. When used without a value, it enables the transparent proxy with default settings, making the intent clearer.

To ensure consistency, `kumactl install transparent-proxy` will also support the unified `--config` flag, following the same precedence rules. The older `--config-file` flag will be deprecated.

### Example

Assumptions:

1. Kuma was installed with default control plane settings
2. User creates the following Pod:
   ```yaml
   apiVersion: v1
   kind: Pod
   metadata:
     annotations:
       kuma.io/sidecar-injection: enabled
       traffic.kuma.io/exclude-inbound-ports: "7777"
       traffic.kuma.io/transparent-proxy-configmap-name: configmap-with-custom-transparent-proxy-config
   ...
   ```
3. The default transparent proxy ConfigMap was modified:
   ```yaml
   apiVersion: v1
   kind: ConfigMap
   metadata:
     name: kuma-transparent-proxy-config
     namespace: kuma-system
   data:
     config.yaml: |
       redirect:
         outbound:
           excludePorts: [8888]
   ```
4. A custom ConfigMap was created in the Pod’s namespace:
   ```yaml
   apiVersion: v1
   kind: ConfigMap
   metadata:
     name: configmap-with-custom-transparent-proxy-config
     namespace: ...
   data:
     config.yaml: |
       ipFamilyMode: ipv4
   ```

After injection, the resulting Pod should look like:

```yaml
apiVersion: v1
kind: Pod
metadata:
  annotations:
    kuma.io/sidecar-injection: enabled
    traffic.kuma.io/exclude-inbound-ports: "7777"
    traffic.kuma.io/transparent-proxy-configmap-name: configmap-with-custom-transparent-proxy-config
    traffic.kuma.io/transparent-proxy-config: |
      redirect:
        inbound:
          excludePorts: [7777]
        outbound:
          excludePorts: [8888]
spec:
  containers:
  - name: kuma-sidecar
    args:
    - run
    - --transparent-proxy-config=/tmp/transparent-proxy/default/config.yaml
    - --transparent-proxy-config=/tmp/transparent-proxy/custom/config.yaml
    volumeMounts:
    - name: transparent-proxy-default
      mountPath: /tmp/transparent-proxy/default
      readOnly: true
    - name: transparent-proxy-custom
      mountPath: /tmp/transparent-proxy/custom
      readOnly: true
  initContainers:
  - name: kuma-init
    command:
    - /usr/bin/kumactl
    - install
    - transparent-proxy
    args:
    - --config=/tmp/transparent-proxy/default/config.yaml
    - --config=/tmp/transparent-proxy/custom/config.yaml
    volumeMounts:
    - name: transparent-proxy-default
      mountPath: /tmp/transparent-proxy/default
      readOnly: true
    - name: transparent-proxy-custom
      mountPath: /tmp/transparent-proxy/custom
      readOnly: true
  volumes:
  - name: transparent-proxy-default
    downwardAPI:
      items:
      - fieldRef:
          apiVersion: v1
          fieldPath: metadata.annotations['traffic.kuma.io/transparent-proxy-config']
        path: config.yaml
  - name: transparent-proxy-custom
    configMap:
      name: configmap-with-custom-transparent-proxy-config
...
```

## Downsides

1. Each component now has to parse annotations and handle config merging on its own. They'll also need to mount any custom ConfigMap the user specifies. But we can reuse the same code across components.
2. When using custom ConfigMaps on workloads, it won’t be possible to specify settings needed during sidecar injection, such as `kumaDPUser` or `ebpf`. This isn’t a major limitation because these values rarely change. If needed, they can still be set globally in the main ConfigMap or in the control plane configuration. The need to override them per workload is extremely rare. 
