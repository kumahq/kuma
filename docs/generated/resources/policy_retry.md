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

