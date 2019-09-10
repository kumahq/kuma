[![][kuma-logo]][kuma-url]

[![CircleCI](https://circleci.com/gh/Kong/kuma.svg?style=svg&circle-token=e3f6c5429ee47ca0eb4bd2542e4b8801a7856373)](https://circleci.com/gh/Kong/kuma)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/Kong/kuma/blob/master/LICENSE)
[![Twitter](https://img.shields.io/twitter/follow/thekonginc.svg?style=social&label=Follow)](https://twitter.com/intent/follow?screen_name=thekonginc)

Kuma is a universal open source control-plane for Service Mesh and Microservices that can run and be operated natively across both Kubernetes and VM environments, in order to be easily adopted by every team in the organization.

Built on top of Envoy, Kuma can instrument any L4/L7 traffic to secure, observe, route and enhance connectivity between any service or database. It can be used natively in Kubernetes via CRDs or via a RESTful API across other environments, and it doesn't require to change our application's code in order to be used.

Built by Envoy contributors at Kong 🦍.

**Need help?** Installing and using Kuma should be as easy as possible. [Contact and chat](https://kuma.io/community) with the community in real-time if you get stuck or need clarifications. We are here to help.

[Installation](https://kuma.io/install) |
[Documentation](https://kuma.io/docs) |
[Slack Chat](https://chat.kuma.io) |
[Community](https://kuma.io/community) |
[Blog](https://konghq.com/blog) |
[Kong](https://konghq.com)

## Summary

- [**Why Kuma?**](#why-kuma)
- [**Features**](#features)
- [**Distributions**](#distributions)
- [**Development**](#development)
- [**Enterprise Demo**](#enterprise-demo)
- [**License**](#license)

## Why Kuma?

If you are building any software architecture, you will inevitably introduce services that will communicate with each ohter ny requests on a network. Every time our services interconnect via a network request, we put the end-user experience at risk. As we all know the connectivity between different services can be slow and unpredictable. It can be insecure, hard to trace, and pose many other problems (e.g., routing, versioning, canary deployments).

With Kuma, our main goal is to reduce the code that has to be written and maintained to build reliable architectures. Therefore, Kuma embraces the sidecar proxy model by leveraging Envoy as its sidecar data-plane technology and by providing a Universal Control Plane that can run on both modern Kubernetes architectures and existing VM-based applications in order to deliver business value today across every team in the organization.

[![][kuma-benefits]][kuma-url]

## Features

* **Universal Control Plane**: Easy to use, distributed, runs anywhere.
* **Lightweight Data Plane**: To process any traffic, powered by Envoy.
* **Automatic**: No code changes required in K8s, flexible on VMs.
* **Multi-Tenancy**: To setup multiple isolated Service Meshes in one cluster and one Control Plane.
* **Network Security**: Automatic mTLS encryption.
* **Traffic Segmentation**: With flexible ACL rules.
* **Traffic Tracing**: Automatic with Zipkin and Jeager integrations.
* **Traffic Metrics**: Automatic with Prometheus/Splunk/ELK integrations.
* **Proxy Configuration Templating**: For advanced users, to configure low-level Envoy configuration.
* **Tagging Selectors**: To apply sophisticated regional, cloud-specific and team-oriented policies.
* **Platform-Agnostic**: Support for K8s, VMs, and bare metal.
* **Powerful APIM Ingress**: Via Kong Gateway integration.

## Distributions

Kuma is a platform-agnostic product that comes in many shapes. You can explore the available installation options from [the official website](https://kuma.io/install).

You can use Kuma for both modern greenfield applications built on containers and Kubernetes, as well as existing applications running on more traditional infrastructure. Kuma can be fully configured via CRDs on Kubernetes, and via a RESTful HTTP API on other environments that can be easily integrated with CI/CD workflows. 

Kuma also provides an easy to use `kumactl` CLI client for every environment.

## Development

Kuma is under active development and prodution-ready.

See [Developer Guide](DEVELOPER.md) for further details.

## Enterprise Demo

If you are implementing Kuma in a mission-critical environment, visit [Request Demo](https://kuma.io/request-demo/) and get in touch with Kong.

## License

```
Copyright 2019-2020 Kong Inc.

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
[kuma-logo]: https://kuma-public-assets.s3.amazonaws.com/kuma-logo.png
[kuma-benefits]: https://kuma-public-assets.s3.amazonaws.com/kuma-benefits.png
