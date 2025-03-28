# Reworking transparent proxy config with ConfigMaps

## Context and problem statement

In version 2.9, we added an experimental feature to configure the transparent proxy using a ConfigMap in Kubernetes. But the first version had issues. The control plane needed access to ConfigMaps in *all* namespaces, which some users didn't allow.

This document proposes removing that requirement. Instead of the control plane building the full proxy config, each data plane component (`kuma-init`, `kuma-dp`, `kuma-cni`) will build its own config.

## Problems

### Control plane needs the config to build the Dataplane object

The Dataplane object includes some transparent proxy settings, which the control plane uses to generate xDS config. The control plane creates or updates this object when a Pod changes. But it doesn't have a clean way to get the proxy config from the sidecar injector, because they're separate components.

### Each component gets config in a different way

- `kuma-init`: gets full config as a CLI argument
- `kuma-cni`: gets full config as an annotation (`traffic.kuma.io/transparent-proxy-config`)
- `kuma-dp`: gets config as environment variables (`KUMA_DNS_ENABLED`, `KUMA_DNS_CORE_DNS_PORT`)

This inconsistency makes it hard to track where values are coming from.

## Solution

We remove proxy config from the Dataplane object. Instead, the control plane will pass the config as metadata when `kuma-dp` connects.

`kuma-dp` will accept a CLI flag like `--transparent-proxy-config`, just like `kuma-init`. Some fields can still be overridden using env vars.

During sidecar injection, the control plane will:
1. Combine default settings from its config and the `kuma-system` ConfigMap
2. Apply any annotations set on the Pod
3. Compute the differences from defaults
4. Add a single annotation `traffic.kuma.io/transparent-proxy-config` with the final config

Components (`kuma-init`, `kuma-dp`, `kuma-cni`) will read this annotation via a mounted volume using the Downward API.

If there's a `traffic.kuma.io/transparent-proxy-configmap-name` annotation, the named ConfigMap will be mounted into `kuma-init` and `kuma-dp`, and its values will override everything else.

Even though environment variables seem easier, we avoid using them for the full config. They take priority over everything else when parsed, and the control plane can't safely set them if it doesn't have access to the workload's ConfigMaps. Annotations give better control over precedence.

## Downsides

Each component now has to parse annotations and handle config merging on its own. They'll also need to mount any custom ConfigMap the user specifies. But we can reuse the same code across components.
