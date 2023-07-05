## RateLimit

- `sources` (required, repeated)

    List of selectors to match dataplanes that rate limit will be applied for    
    
    - `match` (optional)
    
        Tags to match, can be used for both source and destinations

- `destinations` (required, repeated)

    List of selectors to match services that need to be rate limited.    
    
    - `match` (optional)
    
        Tags to match, can be used for both source and destinations

- `conf` (required)

    Configuration for RateLimit
    +required    
    
    - `http` (optional)
    
        The HTTP RateLimit configuration
        +optional    
        
        - `requests` (required)
        
            The number of HTTP requests this RateLimiter allows
            +required    
        
        - `interval` (required)
        
            The the interval for which `requests` will be accounted.
            +required    
        
        - `onRateLimit` (optional)
        
            Describes the actions to take on RatelLimiter event
            +optional    
            
            - `status` (optional)
            
                The HTTP status code to be set on a RateLimit event
                +optional    
            
            - `headers` (optional, repeated)
            
                The Headers to be added to the HTTP response on a RateLimit event
                +optional    
                
                - `key` (optional)
                
                    Header name
                    +optional    
                
                - `value` (optional)
                
                    Header value
                    +optional    
                
                - `append` (optional)
                
                    Should the header be appended
                    +optional

