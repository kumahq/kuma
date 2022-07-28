## TrafficPermission

- `sources` (required, repeated)

    List of selectors to match dataplanes that are sources of traffic.    
    
    - `match` (optional)
    
        Tags to match, can be used for both source and destinations

- `destinations` (required, repeated)

    List of selectors to match services that are destinations of traffic.    
    
    - `match` (optional)
    
        Tags to match, can be used for both source and destinations

