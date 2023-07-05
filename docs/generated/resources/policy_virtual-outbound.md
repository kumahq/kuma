## VirtualOutbound

- `selectors` (required, repeated)

    List of selectors to match dataplanes that this policy applies to

- `conf` (required)

    Child properties:    
    
    - `host` (required)
    
        Host the gotemplate to generate the hostname from the Parameters map    
    
    - `port` (required)
    
        Port the gotemplate to generate the port from the Parameters map    
    
    - `parameters` (required, repeated)
    
        Parameters a mapping between tag keys and template parameter key. This
        must always contain at least `kuma.io/service`

