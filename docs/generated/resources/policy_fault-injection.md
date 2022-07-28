## FaultInjection

- `sources` (required, repeated)

    List of selectors to match dataplanes that are sources of traffic.

- `destinations` (required, repeated)

    List of selectors to match services that are destinations of traffic.

- `conf` (required)

    Configuration of FaultInjection

    Child properties:    
    
    - `delay` (optional)
    
        Delay if specified then response from the destination will be delivered
        with a delay
    
        Child properties:    
        
        - `percentage` (required)
        
            Percentage of requests on which delay will be injected, has to be in
            [0.0 - 100.0] range    
        
        - `value` (required)
        
            The duration during which the response will be delayed    
    
    - `abort` (optional)
    
        Abort if specified makes source side to receive specified httpStatus code
    
        Child properties:    
        
        - `percentage` (required)
        
            Percentage of requests on which abort will be injected, has to be in
            [0.0 - 100.0] range    
        
        - `httpstatus` (required)
        
            HTTP status code which will be returned to source side    
    
    - `responseBandwidth` (optional)
    
        ResponseBandwidth if specified limits the speed of sending response body
    
        Child properties:    
        
        - `percentage` (required)
        
            Percentage of requests on which response bandwidth limit will be
            injected, has to be in [0.0 - 100.0] range    
        
        - `limit` (required)
        
            Limit is represented by value measure in gbps, mbps, kbps or bps, e.g.
            10kbps

