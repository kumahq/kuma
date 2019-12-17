# Ingress Gateway

## Context

When kuma-dp is deployed, both ingress and egress connections are intercepted by Envoy, probably with mTLS applied.
We want to enable users to access the Mesh with Ingress Gateway, like this:
```
--(Traffic from the Internet)--> Ingress Gateway --> Gateway's Envoy --(Traffic encrypted by Kuma)--> App's Envoy --> App 
```
This means that **only** outbound traffic of Ingress Gateway should be intercepted by Envoy.

## Configuration model

### Universal

There are several options how to model this

#### Dataplane entity

Right now we require valid inbound interface in Dataplane entity.
We could change this restriction and introduce convention to not open inbound interface when value is `:0` or `false` 
```yaml
type: Dataplane
mesh: demo
name: gateway-01
networking:
  inbound:
  - interface: :0
    tags:
      service: gateway
  outbound:
  - interface: :33033
    service: database
  - interface: :44044
    service: user
```

#### Gateway entity

We can also introduce brand new entity.

```yaml
type: Gateway
mesh: demo
name: gateway-01
tags:
  service: gateway
networking:
  outbound:
  - interface: :33033
    service: database
  - interface: :44044
    service: user
```

While `Gateway` entity seems like a better option UX-wise it gives us a couple challenges with implementation.

Gateway is a special case of Dataplane, so all features for Dataplane should be available for Gateway.
* Should gateway appear in `get dataplanes`? If not, we need to duplicate `get dataplanes` and `inspect dataplanes` (+ API, + core entities, + logic) to provide statistics.
  If we include it into `get dataplanes` and GET request, should PUT/DELETE requests to Dataplanes operate on Gateways?
* We use Dataplane elements across the CP logic. We would need to convert Gateway into Dataplane on the fly, which brings another question: Can mesh+name be the same for a Gateway as for a Dataplane? 
* It's potentially error prone. When dealing with Dataplanes we would always have to think about Gateways. We can miss including Gateway in some logic.

### Kubernetes

Since we don't specify `Dataplane` objects manually, we can just mark Pod as Gateway with annotation `kuma.io/gateway=true`.
Inbounds traffic to apps in pods marked with this annotation won't be intercepted by Envoy.
Injector will detect this annotation and set empty `-b` argument to [iptables script](https://github.com/istio/cni/blob/master/tools/packaging/common/istio-iptables.sh). 

We could also introduce more specific configuration with ports that should be intercepted `kuma.io/inboundPorts=1234,5678`. 