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
the `MeshGateway` listener as well as mTLS on any meshes doing cross-mesh
traffic.
This option will allow us to configure the corresponding Envoy listener
with all mesh CAs.

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
    hostname: expose.mesh # optional
    tags:
      name: mesh-listener
    tls:
      mode: TERMINATE
      options:
        crossMesh: true # new option
```

There are two ways we can enable configuring more specific access rules.

#### `MeshGatewayTrafficPermission`

The first is to add a _new resource_ that defines which
cross-mesh traffic is allowed for a matching `MeshGateway` listener and
`routeName`.

```
type: MeshGatewayTrafficPermission
mesh: expose
name: cross-mesh
selector: # these match a listener
- tags:
    kuma.io/service: edge-gateway
    name: mesh-listener
default: deny
rules:
- routeName: cross-mesh
  action: allow
  sources:
  - mesh:
      name: consume # or '*' for any mesh
      tags:
        kuma.io/service: client
```

This concentrates `MeshGateway` permission management into its own resource.
This may be a more generally useful abstraction given potential
other kinds of `sources`. For example, one for filtering based on JWT token:

```
  - routeName: cross-mesh
    action: allow
    sources:
    - user:
        email: abc@xyz.com
```

#### Inline `MeshGatewayRoute`

Another option is to add cross-mesh access control to the `MeshGatewayRoute`
resource, as another type of `conf.http.rules.matches` entry or
potentially at the top level under `conf.http`.

```
type: MeshGatewayRoute
mesh: expose
name: edge-gateway-route
selectors:
  - match:
      kuma.io/service: edge-gateway
      name: mesh-listener
conf:
  http:
    mesh:                    # new entry type
      - name: consume
        tags:
          kuma.io/service: client
    rules:
      - matches:
          # ALL match types must succeed
          - path:
              match: PREFIX
              value: /api
          - mesh:                    # and/or a new entry type here
            - name: consume
              tags:
                kuma.io/service: client
        backends:
          - destination:
              kuma.io/service: server
      - matches:    # ALL matches must succeed
          - path:
              match: PREFIX
              value: /api/v2
          - mesh:
            - name: *
        backends:
          - destination:
              kuma.io/service: server
```

The advantage of this option is probably reduced complexity for less complicated
cases.

### `consume` UX

In order for `consume` services to communicate with `expose`, we need to give
those services a way to contact the `MeshGateway`. A `MeshGateway` doesn't have
an address or concrete identity beyond the identities of the `kuma-dp` proxies that
serve the gateway.

#### Options

There are multiple UX options for making the `MeshGateway` addressable from
other meshes.

##### Virtual outbounds

We can add support for `MeshGateway` listeners to `VirtualOutbounds` to give applications
a way to address a `MeshGateway` listener from another mesh.

The `VirtualOutbound` selects `Dataplane`s and uses them to generate outbounds.
These generated outbounds are added to _all_ `Dataplane` resources.

```
---
type: VirtualOutbound
mesh: consume
name: test
selectors: # this VirtualOutbound will be built using the matched data plane proxies and will be added for all data plane proxies
  - match:
      kuma.io/service: edge-gateway
      name: mesh-listener
conf:
  host: "gateway.{{.gateway.mesh}}.mesh"
    # or host: "{{.gateway.listener.hostname}}"
  port: "{{.gateway.listener.port}}"
  parameters:
    - name: service
      tagKey: "kuma.io/service"
    - name: gateway
      gatewayListener: {}
```

At the moment, the `selectors` are only run for `Dataplane` resources whereas we
need support for `crossMesh`-enabled `MeshGateway` listeners (of all `Mesh`es) as well. In the
implementation, this means keeping track of all `MeshGateway`-serving `Dataplane`s
of _any mesh_ when generating a `VirtualOutbound` of _any mesh_.

We also need to provide the properties of the `MeshGateway` listener to the `host`
template in order to generate the hostname.

This should give us a virtual outbound:

```
address: 240.0.0.1
port: 8090
mesh: expose           # we need to keep track of the mesh somehow
tags:
  kuma.io/service: edge-gateway
  name: mesh-listener
```

along with the hostname `gateway.expose.mesh`, giving access from `consume` to `expose`
at `http://gateway.expose.mesh:8090/api`.

##### Listener hostname

Another option for giving access to the `MeshGateway` would be to require and
use the `listeners[].hostname` option to generate an outbound for every 
dataplane implicitly. In effect, this would be a virtual outbound with the
`host` template equal to `{{ .gateway.listener.hostname }}`.
