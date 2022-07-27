## VirtualOutbound

- `selectors` (required, repeated)

    List of selectors to match dataplanes that this policy applies to    
    
    - `match` (optional)
    
        Tags to match, can be used for both source and destinations

- `conf` (required)    
    
    - `host` (required)
    
        Host the gotemplate to generate the hostname from the Parameters map    
    
    - `port` (required)
    
        Port the gotemplate to generate the port from the Parameters map    
    
    - `parameters` (required, repeated)
    
        Parameters a mapping between tag keys and template parameter key. This
        must always contain at least `kuma.io/service`    
        
        - `name` (required)
        
            Name the name of the template parameter (must be alphanumeric).    
        
        - `tagKey` (optional)
        
            TagKey the name of the tag in the Kuma outbound (optional if absent it
            will use Name).

