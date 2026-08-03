# Cross-mesh communication UX

_Issue_: https://github.com/kumahq/kuma/issues/3323

This proposal allows the operators of one `Mesh` to expose services to other `Meshes`
via `MeshGateway`, the same interface for exposing to external traffic.

The proposal addresses UX both for _exposing_ and the _consuming_
meshes and only HTTP/S services. TCP will be addressed in a later proposal.

## Design

In the following examples, the mesh exposing its services will be `expose` and
the mesh consuming another mesh's services will be `consume`.

<details>
<summary>Recap: <code>MeshGateway</code> configuration</summary>

```
type: MeshGateway
mesh: expose
name: edge-gateway
selectors:
- match:
    kuma.io/service: edge-gateway
conf:
  listeners:
  - port: 8080
    protocol: HTTP
    hostname: foo.example.com
    tags:
      port: "8080"
---
type: MeshGatewayRoute
mesh: expose
name: edge-gateway-route
selectors:
  - match:
      kuma.io/service: edge-gateway
conf:
  http:
    rules:
      - matches:
          - path:
              match: PREFIX
              value: /api
        backends:
          - destination:
              kuma.io/service: server
```

</details>

### `expose` UX

The main UX feature will be a mechanism to configure a
`MeshGateway` to permit traffic from services in one
`Mesh` to reach the services behind the `MeshGateway`.
The configuration will make it possible to expose `MeshGatewayRoute`s depending
on cross-mesh traffic's source `Mesh` and source _tags_.

First of all, we require users to explicitly enable cross-mesh traffic on
the `MeshGateway` listener by marking it `crossMesh`. We also require
mTLS on any meshes involved in cross-mesh traffic.
This option will allow us
to configure the corresponding Envoy listener appropriately for mTLS.

```
type: MeshGateway
mesh: expose
name: edge-gateway
selectors:
- match:
    kuma.io/service: edge-gateway
conf:
  listeners:
  - port: 8090
    protocol: HTTP
    hostname: expose.mesh # optional in general
    crossMesh: true
    tags:
      name: mesh-listener
```

The initial implementation will require `protocol: HTTP` when using `crossMesh`.
When using `crossMesh`, communications are already protected with TLS so the use
case for both `crossMesh` and `HTTPS` is limited.
In a later iteration we may examine whether and how `HTTPS` can be supported for
cross-mesh communication.

#### Permissions

We will take advantage of the new `MeshTrafficPermission` to define which
cross-mesh traffic is allowed for a matching `MeshGateway` listener and
`routeName`. It's being discussed in https://github.com/kumahq/kuma/issues/4222

For the first iteration, traffic will be permitted from all services in all
meshes.

#### Inline `MeshGatewayRoute`

We decided against adding cross-mesh access control to the `MeshGatewayRoute`
resource, as another type of `conf.http.rules.matches` entry or
potentially at the top level under `conf.http`.

It reduces our flexibility in the future.

### `consume` UX

In order for `consume` services to communicate with `expose`, we need to give
those services a way to contact the `MeshGateway`. A `MeshGateway` doesn't have
an address or concrete identity beyond the identities of the `kuma-dp` proxies that
serve the gateway.

#### Virtual outbounds

Every `MeshGateway` listener can be given a `hostname`. For cross-mesh, we will
require one be set and use it to automatically generate a virtual outbound for
`Dataplane`s in other `Mesh`es.

The listener from above

```
type: MeshGateway
mesh: expose
name: edge-gateway
selectors:
- match:
    kuma.io/service: edge-gateway
conf:
  listeners:
  - port: 8090
    protocol: HTTPS
    hostname: expose.mesh
    tags:
      name: mesh-listener
```

should give us the DNS hostname `expose.mesh` as well as a virtual outbound
like:

```
address: 240.0.0.1
port: 8090
mesh: expose # new, we need to keep track of the mesh somehow
tags:
  kuma.io/service: edge-gateway
  name: mesh-listener
```

giving access from `consume` to `expose` at `http://expose.mesh:8090/api`.

## Performance considerations

When using cross-mesh `MeshGateway`, the configuration generation for
one `Mesh` becomes dependent on changes in another.
Any change in `expose`'s `MeshGateway` or any
`MeshGateway` `Dataplane`s will cause a rebuild of the config for every
`Dataplane` in `consume`. In an environment with a large number of meshes,
this may mean a significant amount of recomputation.

In the future, we could improve this by including an option
to require other meshes like `consume`
to opt-in to using `expose`'s cross-mesh `MeshGateway` or further limiting
exposure to cross-mesh `MeshGateway`s.
