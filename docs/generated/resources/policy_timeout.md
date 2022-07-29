## Timeout

- `sources` (required, repeated)

    List of selectors to match dataplanes that are sources of traffic.    
    
    - `match` (optional)
    
        Tags to match, can be used for both source and destinations

- `destinations` (required, repeated)

    List of selectors to match services that are destinations of traffic.    
    
    - `match` (optional)
    
        Tags to match, can be used for both source and destinations

- `conf` (required)    
    
    - `connectTimeout` (optional)
    
        ConnectTimeout defines time to establish connection    
    
    - `tcp` (optional)    
        
        - `idleTimeout` (required)
        
            IdleTimeout is defined as the period in which there are no bytes sent
            or received on either the upstream or downstream connection    
    
    - `http` (optional)    
        
        - `requestTimeout` (optional)
        
            RequestTimeout is a span between the point at which the entire
            downstream request (i.e. end-of-stream) has been processed and when the
            upstream response has been completely processed    
        
        - `idleTimeout` (optional)
        
            IdleTimeout is the time at which a downstream or upstream connection
            will be terminated if there are no active streams    
        
        - `streamIdleTimeout` (optional)
        
            StreamIdleTimeout is the amount of time that the connection manager
            will allow a stream to exist with no upstream or downstream activity    
        
        - `maxStreamDuration` (optional)
        
            MaxStreamDuration is the maximum time that a stream’s lifetime will
            span    
    
    - `grpc` (optional)
    
        Deprecated: set parameters through Http section    
        
        - `streamIdleTimeout` (optional)
        
            StreamIdleTimeout is the amount of time that the connection manager
            will allow a stream to exist with no upstream or downstream activity
            Deprecated: use Http.StreamIdleTimeout instead    
        
        - `maxStreamDuration` (optional)
        
            MaxStreamDuration is the maximum time that a stream’s lifetime will
            span
            Deprecated: use Http.MaxStreamDuration instead

