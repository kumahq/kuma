## Retry

- `sources` (required, repeated)

    List of selectors to match dataplanes that retry policy should be
    configured for

- `destinations` (required, repeated)

    List of selectors to match services that need to be health checked.

- `conf` (required)

    +required

    Child properties:    
    
    - `http` (optional)
    
        Child properties:    
        
        - `numRetries` (optional)
        
            +optional    
        
        - `perTryTimeout` (optional)
        
            +optional    
        
        - `backOff` (optional)
        
            +optional
        
            Child properties:    
            
            - `baseInterval` (required)
            
                +required    
            
            - `maxInterval` (optional)
            
                +optional    
        
        - `retriableStatusCodes` (optional, repeated)
        
            +optional    
        
        - `retriableMethods` (optional, repeated)
        
            +optional
        
            Supported values:
        
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
    
        Child properties:    
        
        - `maxConnectAttempts` (optional)
        
            +optional    
    
    - `grpc` (optional)
    
        Child properties:    
        
        - `retryOn` (optional, repeated)
        
            +optional
        
            Supported values:
        
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
        
            Child properties:    
            
            - `baseInterval` (required)
            
                +required    
            
            - `maxInterval` (optional)
            
                +optional

