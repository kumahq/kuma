## Retry

- `sources` (required, repeated)

    List of selectors to match dataplanes that retry policy should be
    configured for    
    
    - `match` (optional)
    
        Tags to match, can be used for both source and destinations

- `destinations` (required, repeated)

    List of selectors to match services that need to be health checked.    
    
    - `match` (optional)
    
        Tags to match, can be used for both source and destinations

- `conf` (required)

    +required    
    
    - `http` (optional)    
        
        - `numRetries` (optional)
        
            +optional    
        
        - `perTryTimeout` (optional)
        
            +optional    
        
        - `backOff` (optional)
        
            +optional    
            
            - `baseInterval` (required)
            
                +required    
            
            - `maxInterval` (optional)
            
                +optional    
        
        - `retriableStatusCodes` (optional, repeated)
        
            +optional    
        
        - `retriableMethods` (optional, repeated, enum)
        
            +optional
        
            - `NONE`
        
            - `CONNECT`
        
            - `DELETE`
        
            - `GET`
        
            - `HEAD`
        
            - `OPTIONS`
        
            - `PATCH`
        
            - `POST`
        
            - `PUT`
        
            - `TRACE`    
        
        - `retryOn` (optional, repeated, enum)
        
            +optional
        
            - `all_5xx`
        
            - `gateway_error`
        
            - `reset`
        
            - `connect_failure`
        
            - `envoy_ratelimited`
        
            - `retriable_4xx`
        
            - `refused_stream`
        
            - `retriable_status_codes`
        
            - `retriable_headers`
        
            - `http3_post_connect_failure`    
    
    - `tcp` (optional)    
        
        - `maxConnectAttempts` (optional)
        
            +optional    
    
    - `grpc` (optional)    
        
        - `retryOn` (optional, repeated, enum)
        
            +optional
        
            - `cancelled`
        
            - `deadline_exceeded`
        
            - `internal`
        
            - `resource_exhausted`
        
            - `unavailable`    
        
        - `numRetries` (optional)
        
            +optional    
        
        - `perTryTimeout` (optional)
        
            +optional    
        
        - `backOff` (optional)
        
            +optional    
            
            - `baseInterval` (required)
            
                +required    
            
            - `maxInterval` (optional)
            
                +optional

