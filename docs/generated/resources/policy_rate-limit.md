## RateLimit

- `sources` (required, repeated)

    List of selectors to match dataplanes that rate limit will be applied for

- `destinations` (required, repeated)

    List of selectors to match services that need to be rate limited.

- `conf` (required)

    Configuration for RateLimit
    +required

    Child properties:    
    
    - `http` (optional)
    
        The HTTP RateLimit configuration
        +optional
    
        Child properties:    
        
        - `requests` (required)
        
            The number of HTTP requests this RateLimiter allows
            +required    
        
        - `interval` (required)
        
            The the interval for which `requests` will be accounted.
            +required    
        
        - `onratelimit` (optional)
        
            Describes the actions to take on RatelLimiter event
            +optional
        
            Child properties:    
            
            - `status` (optional)
            
                The HTTP status code to be set on a RateLimit event
                +optional    
            
            - `headers` (optional, repeated)
            
                The Headers to be added to the HTTP response on a RateLimit event
                +optional

