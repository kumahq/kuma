[![][kuma-logo]][kuma-url]

[![CircleCI](https://circleci.com/gh/kumahq/kuma/tree/release-0.7.svg?style=svg&circle-token=e3f6c5429ee47ca0eb4bd2542e4b8801a7856373)](https://circleci.com/gh/kumahq/kuma/tree/release-0.7)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/kumahq/kuma/blob/master/LICENSE)
[![Slack](https://chat.kuma.io/badge.svg)](https://chat.kuma.io/)
[![Twitter](https://img.shields.io/twitter/follow/KumaMesh.svg?style=social&label=Follow)](https://twitter.com/intent/follow?screen_name=KumaMesh)

Kuma is a platform agnostic open-source control plane for Service Mesh and Microservices. It can run and be operated natively across both Kubernetes and VM environments, making it easy to adopt by every team in the organization.

Bundling Envoy as a data-plane, Kuma can instrument any L4/L7 traffic to secure, observe, route and enhance connectivity between any service or database. It can be used natively in Kubernetes via CRDs or via a RESTful API across other environments like VMs and Bare Metal.

While being simple to use for most use-cases, Kuma also provides policies to configure the underlying Envoy data-planes in a more fine-grained manner. By doing so, Kuma can be used by both first-time users of Service Mesh, as well as the most experienced ones.

Built by Envoy contributors at Kong 🦍.

**Need help?** Installing and using Kuma should be as easy as possible. [Contact and chat](https://kuma.io/community) with the community in real-time if you get stuck or need clarifications. We are here to help.

[Installation](https://kuma.io/install) |
[Documentation](https://kuma.io/docs) |
[Community + Slack Chat](https://kuma.io/community) |
[Blog](https://konghq.com/blog)

## Summary

- [**Why Kuma?**](#why-kuma)
- [**Features**](#features)
- [**Distributions**](#distributions)
- [**Development**](#development)
- [**Enterprise Demo**](#enterprise-demo)
- [**License**](#license)

## Why Kuma?

Built with enterprise use-cases in mind, Kuma is a universal service mesh that supports both Kubernetes and VMs deployments across single and multi-zone setups, with turnkey mesh policies to get up and running easily while supporting multi-tenancy and multi-mesh on the same control plane. Kuma is a donated CNCF Sandbox project.

Modern applications will inevitably make requests over a network to communicate to other services, like databases, caches or microservices. But - as we all know - the network is by default unreliable and unsecure, and can introduce significant challenges to any modern environment like security, tracing and routing among the others.

Kuma is a better way to build L4/L7 connectivity among your services and applications (Service Mesh) by reducing the code that application teams have to write, enabling to ship products faster and improve the reliability and security of the overall architecture with minimal effort. 

Kuma embraces the sidecar proxy model by bundling Envoy as its sidecar data-plane technology and by providing a platform-agnostic Control Plane that can run on both modern Kubernetes and existing VM/Bare Metal architectures in order to deliver business value across every team in the organization with one comprehensive solution.

Kuma has been built with universality and scalability in mind. Kubernetes and VMs are both first class citizens, it supports multiple isolated meshes on one control plane and offers global/remote CPs deployments for both single and multi zone setups across different platforms, clouds and data-centers in order to tackle the most complex deployments. With built-in connectivity thanks to the automatic ingress mode and built-in discovery, Kuma abstracts away connectivity across the entire mesh.

[![][kuma-benefits]][kuma-url]

## Features

* **Universal Control Plane**: Easy to use, distributed, runs anywhere on both Kubernetes and VM/Bare Metal.
* **Lightweight Data Plane**: Powered by Envoy to process any L4/L7 traffic, with automatic Envoy bootstrapping.
* **Automatic DP Injection**: No code changes required in K8s. Easy YAML specification for VM and Bare Metal deployments.
* **Multi-Mesh**: To setup multiple isolated Meshes in one cluster and one Control Plane, lowering OPs cost.
* **Single and Multi Zone**: To deploy a service mesh that is cross-platform, cross-cloud and cross-cluster.
* **Automatic Discovery & Ingress**: With built-in service discovery and connectivity across single and multi-zones.
* **Global & Remote CPs**: For scalability across deployments with multiple zones, including hybrid VMs + K8s meshes.
* **mTLS**: Automatic mTLS issuing, identity and encryption with optional support for third-party CA.
* **TLS Rotation**: Automatic certificate rotation for all the data planes, with configurable settings.
* **Traffic Permissions**: To firewall traffic between the services of a Mesh.
* **Traffic Routing**: With dynamic load-balancing for blue/green, canary, versioning and rollback deployments.
* **Fault Injection**: To harden our systems by injecting controlled artificial faults and observe the behavior.
* **Traffic Logs**: To log all the activity to a third-party service, like Splunk or ELK.
* **Traffic Tracing**: To observe the full trace of the service traffic and determine bottlenecks.
* **Traffic Metrics**: For every Envoy dataplane managed by Kuma with native Prometheus/Grafana support.
* **Proxy Configuration Templating**: The easiest way to run and configure Envoy with low-level configuration.
* **Gateway Support**: To support any API Gateway or Ingress, like [Kong Gateway](https://github.com/Kong/kong).
* **Healthchecks**: Both active and passive.
* **GUI**: Out of the box browser GUI to explore all the Service Meshes configured in the system.
* **Tagging Selectors**: To apply sophisticated regional, cloud-specific and team-oriented policies.
* **Platform-Agnostic**: Support for Kubernetes, VMs, and bare metal. Including hybrid deployments.
* **Network Overlay**: Create a configurable Mesh overlay across different Kubernetes clusters and namespaces.

Below you can find a snapshot of the GUI that ships with Kuma:

[![][kuma-gui]][kuma-url]

## Distributions

Kuma is a platform-agnostic product that ships in different distributions. You can explore the available installation options at [the official website](https://kuma.io/install).

You can use Kuma for modern greenfield applications built on containers as well as existing applications running on more traditional infrastructure. Kuma can be fully configured via CRDs (Custom Resource Definitions) on Kubernetes and via a RESTful HTTP API in other environments that can be easily integrated with CI/CD workflows.

Kuma also provides an easy to use `kumactl` CLI client for every environment, and an official GUI that can be accessed by the browser.

## Development

Kuma is under active development and production-ready.

See [Developer Guide](DEVELOPER.md) for further details.

## Enterprise Support

If you are implementing Kuma in a mission-critical environment and require Enterprise support and features, please visit [Enterprise](https://kuma.io/enterprise/).

## License

```
Copyright 2020 the Kuma Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```

[kuma-url]: https://kuma.io/
[kuma-logo]: https://kuma-public-assets.s3.amazonaws.com/kuma-logo-v2.png
[kuma-gui]: https://kuma-public-assets.s3.amazonaws.com/kuma-gui-v2.png
[kuma-benefits]: https://kuma-public-assets.s3.amazonaws.com/kuma-benefits-v2.png
