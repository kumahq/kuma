## MeshGateway

- `selectors` (optional, repeated)

    Selectors is a list of selectors that are used to match builtin
    gateway dataplanes that will receive this MeshGateway configuration.

- `tags` (optional)

    Tags is the set of tags common to all of the gateway's listeners.
    
    This field must not include a `kuma.io/service` tag (the service is always
    defined on the dataplanes).

- `conf` (optional)

    The desired configuration of the MeshGateway.

    Child properties:    
    
    - `listeners` (optional, repeated)
    
        Listeners define logical endpoints that are bound on this MeshGateway's
        address(es).

