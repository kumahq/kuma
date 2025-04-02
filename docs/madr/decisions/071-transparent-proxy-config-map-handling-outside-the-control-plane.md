# Transparent Proxy ConfigMap Handling Outside the Control Plane

* Status: accepted

## Context and problem statement

In version 2.9, we introduced an experimental feature to configure the transparent proxy using a Kubernetes ConfigMap. This version had a major limitation: the control plane needed access to ConfigMaps in *all* namespaces, which some users didn’t allow.

There are three components outside the control plane that rely on the transparent proxy configuration:

- `kuma-init` - an init container injected alongside the workload container to install the transparent proxy
- `kuma-cni` - our custom CNI plugin that installs the transparent proxy
- `kuma-dp` - the data plane proxy injected next to the workload container

The proxy config is built from default values, control plane configuration, annotations, and ConfigMaps. Since ConfigMaps are namespace-scoped, the component that builds the final config must access all of these inputs. The simplest way to do that is in the control plane, which then passes the final config to `kuma-init`, `kuma-cni`, and `kuma-dp`. But this requires the control plane to access all workload namespaces, which led to expanding its ClusterRole permissions.

This document proposes removing that requirement for setups that don’t use `kuma-cni`. Instead of having the control plane build the full config, each data plane component (`kuma-init` and `kuma-dp`) will build its own.

## Out of Scope

`kuma-cni` is deployed in the `kube-system` namespace. Without expanding its permissions to access ConfigMaps from other namespaces, it cannot support custom ConfigMaps specified at the workload level. With the proposed changes, where config building moves outside of the control plane, `kuma-cni` will no longer be able to consume workload-specific ConfigMaps at all. This document **_WON'T_** cover the changes required to support such functionality in `kuma-cni`. As a result, the transparent proxy configuration via ConfigMaps will be incomplete and backward incompatible when using `kuma-cni`.

## Problems

### Control plane needs the config to build the Dataplane object

There are two control plane components that require transparent proxy configuration:

- **Sidecar Injector** - injects `kuma-dp` and `kuma-init` containers into Pods that are part of the mesh
- **Pod Reconciler** - creates and updates the `Dataplane` object based on the Pod

The `Dataplane` object includes transparent proxy settings like `ipFamilyMode`, `redirectPortInbound`, and `redirectPortOutbound`. These are used by the control plane to generate the xDS config delivered to `kuma-dp`.

Sidecar injection and Dataplane generation are handled by separate components, with no shared state. This means there is no direct way to pass configuration between them.

### Each component gets config in a different way

- `kuma-init`: gets full config as a CLI argument
- `kuma-dp`: gets config as environment variables (`KUMA_DNS_ENABLED`, `KUMA_DNS_CORE_DNS_PORT`)

This inconsistency makes it hard to track where values are coming from.

## Solution

We stop using the transparent proxy config fields in the Dataplane object and mark them as deprecated. These fields will no longer be relied on by `kuma-dp` in Kubernetes deployments. Instead, the same information will be passed as metadata during the xDS connection, using values from the transparent proxy config passed via a CLI flag.

`kuma-dp` will accept a CLI flag like `--transparent-proxy-config`, just like `kuma-init`. Some fields can still be overridden using env vars.

During sidecar injection, the control plane will:
1. Combine default settings from its config and the `kuma-system` ConfigMap
2. Apply any annotations set on the Pod
3. Compute the differences from defaults
4. Add a single annotation `traffic.kuma.io/transparent-proxy-config` with the final config

Components (`kuma-init`, `kuma-dp`) will read this annotation via a mounted volume using the Downward API.

If there's a `traffic.kuma.io/transparent-proxy-configmap-name` annotation, the named ConfigMap will be mounted into `kuma-init` and `kuma-dp`, and its values will override everything else.

Even though environment variables seem easier, we avoid using them for the full config. They take priority over everything else when parsed, and the control plane can't safely set them if it doesn't have access to the workload's ConfigMaps. Annotations give better control over precedence.

## Downsides

Each component now has to parse annotations and handle config merging on its own. They'll also need to mount any custom ConfigMap the user specifies. But we can reuse the same code across components.
