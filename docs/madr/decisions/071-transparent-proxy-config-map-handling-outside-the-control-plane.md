# Transparent Proxy ConfigMap Handling Outside the Control Plane

* Status: accepted

## Context and problem statement

In version 2.9, we introduced an experimental feature to configure the transparent proxy using a Kubernetes ConfigMap. This version had a major limitation: the control plane needed access to ConfigMaps in *all* namespaces, which some users didn’t allow.

There are three components outside the control plane that rely on the transparent proxy configuration:

- `kuma-init` - an init container injected alongside the workload container to install the transparent proxy
- `kuma-cni` - our custom CNI plugin that installs the transparent proxy
- `kuma-sidecar` - the data plane proxy injected next to the workload container

The proxy config is built from default values, control plane configuration, annotations, and ConfigMaps. Since ConfigMaps are namespace-scoped, the component that builds the final config must access all of these inputs. The simplest way to do that is in the control plane, which then passes the final config to `kuma-init`, `kuma-cni`, and `kuma-sidecar`. But this requires the control plane to access all workload namespaces, which led to expanding its ClusterRole permissions.

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

We stop using the transparent proxy config fields in the Dataplane object and mark them as deprecated. These fields will no longer be relied on by `kuma-sidecar` in Kubernetes deployments. Instead, the same information will be passed as metadata during the xDS connection, using values from the transparent proxy config passed via a CLI flag.

Both `kuma-init` and `kuma-sidecar` will accept a `--config-dir` CLI flag. This directory will contain one or more configuration files, read in alphabetical order. In case of conflicts, values from the later file override earlier ones.

During sidecar injection, the control plane will:
1. Merge default values from the control plane config and the `kuma-system` ConfigMap
2. Apply any Pod-level annotations
3. Add the resulting values (only those that differ from defaults) as a single annotation: `traffic.kuma.io/transparent-proxy-config`

To avoid unintended overrides, we don’t use environment variables for the full config. Env vars take precedence during parsing and can’t be safely set when the control plane doesn’t have access to workload-level ConfigMaps. Annotations provide better control over precedence and merging.

The annotation will be mounted into each container as a file named `0.yaml` under `/tmp/transparent-proxy`.

If the Pod also has a `traffic.kuma.io/transparent-proxy-configmap-name` annotation, the named ConfigMap will be mounted in the same directory as `1.yaml`. Since files are processed alphabetically, it will override values from the default config.

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
       traffic.kuma.io/transparent-proxy-configmap-name: custom-tproxy-config
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
     name: custom-tproxy-config
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
    traffic.kuma.io/transparent-proxy-configmap-name: custom-tproxy-config
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
    - --config-dir=/tmp/transparent-proxy
    volumeMounts:
    - name: transparent-proxy-default
      mountPath: /tmp/transparent-proxy/0.yaml
      subPath: config.yaml
      readOnly: true
    - name: transparent-proxy-custom
      mountPath: /tmp/transparent-proxy/1.yaml
      subPath: config.yaml
      readOnly: true
  initContainers:
  - name: kuma-init
    command:
    - /usr/bin/kumactl
    - install
    - transparent-proxy
    args:
    - --config-dir=/tmp/transparent-proxy
    volumeMounts:
    - name: transparent-proxy-default
      mountPath: /tmp/transparent-proxy/0.yaml
      subPath: config.yaml
      readOnly: true
    - name: transparent-proxy-custom
      mountPath: /tmp/transparent-proxy/1.yaml
      subPath: config.yaml
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
      name: custom-tproxy-config
...
```

## Downsides

1. Each component now has to parse annotations and handle config merging on its own. They'll also need to mount any custom ConfigMap the user specifies. But we can reuse the same code across components.
2. When using custom ConfigMaps on workloads, it won’t be possible to specify settings needed during sidecar injection, such as `kumaDPUser` or `ebpf`. This isn’t a major limitation because these values rarely change. If needed, they can still be set globally in the main ConfigMap or in the control plane configuration. The need to override them per workload is extremely rare. 
