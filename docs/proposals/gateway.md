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

Right now we require valid inbound interface in Dataplane entity. We can introduce new field which can be defined instead of `inbound`.
```yaml
type: Dataplane
mesh: demo
name: gateway-01
networking:
  gateway:
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

## Workflow with Kong API Gateway

On both K8S and Universal, Kong is deployed into the mesh as any other service.

### Universal

First we deploy Kong and Kuma-DP next to Kong with following DP definition

```yaml
type: Dataplane
mesh: demo
name: kong-01
networking:
  gateway:
    tags:
      service: kong
  outbound:
  - interface: :44044
    service: admin
  - interface: :44043
    service: frontend
```

Next, we configure Kong with services and routes. 

```yaml
services:
- name: admin
  url: http://localhost:44044
  routes:
  - name: admin-route
    paths:
    - /user
- name: frontend
  url: http://localhost:44043
  routes:
  - name: frontend-route
    paths:
    - /frontend
```

Since such dataplane won't expose inbound listener, we pass the traffic to the listener exposed by Kong itself.

### Kubernetes with Kong Ingress Controller

First we annotate [the Deployment](https://raw.githubusercontent.com/Kong/kubernetes-ingress-controller/master/deploy/single/all-in-one-dbless.yaml) of the Kong with annotation `kuma.io/gateway=true` like this
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: ingress-kong
  name: ingress-kong
  namespace: kong
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ingress-kong
  template:
    metadata:
      annotations:
        kuma.io/gateway: enabled
```

Then we configure Kong via K8S Ingress CRD

```yaml
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: demo
spec:
  rules:
  - http:
      paths:
      - path: /admin
        backend:
          serviceName: admin
          servicePort: 80
  - http:
      paths:
      - path: /frontend
        backend:
          serviceName: frontend
          servicePort: 80
```

**Note:** Kong Ingress Controller does load balancing by itself, then it send a request to selected Endpoint with `Host` header of the service.
This breaks transparent proxying since Kuma for now works with L4 and does not recognise `Host` header. Once Kuma works with L7 it can rebalance it with proper endpoint.
There is an option to turn off load balancing by Kong for given service, to do that `Service` to which traffic is directed should be annotated with `ingress.kubernetes.io/service-upstream=true` [annotation](https://github.com/Kong/kubernetes-ingress-controller/blob/master/docs/references/annotations.md#ingresskubernetesioservice-upstream).  

## Summary

We picked option with Dataplane entity, we won't introduce new entity.
Additionally, there should be an option to filter gateways in kumactl and HTTP API.
Right now we want to focus on integrating with current API Gateways, meaning Kuma won't work as Gateway itself.