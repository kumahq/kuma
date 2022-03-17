## TrafficRoute

- `sources` (required, repeated)

    List of selectors to match data plane proxies that are sources of traffic.

- `destinations` (required, repeated)

    List of selectors to match services that are destinations of traffic.
    
    Notice the difference between sources and destinations.
    While the source of traffic is always a data plane proxy within a mesh,
    the destination is a service that could be either within or outside
    of a mesh.

- `conf` (required)

    Configuration for the route.

    Child properties:    
    
    - `split` (optional, repeated)
    
        List of destinations with weights assigned to them.
        When used, "destination" is not allowed.    
    
    - `loadBalancer` (optional)
    
        Load balancer configuration for given "split" or "destination"
    
        Child properties:    
        
        - `roundRobin` (optional)
        
            Child properties:    
        
        - `leastRequest` (optional)
        
            Child properties:    
            
            - `choiceCount` (optional)
            
                The number of random healthy hosts from which the host with the fewest
                active requests will be chosen. Defaults to 2 so that we perform
                two-choice selection if the field is not set.    
        
        - `ringHash` (optional)
        
            Child properties:    
            
            - `hashFunction` (optional)
            
                The hash function used to hash hosts onto the ketama ring. The value
                defaults to 'XX_HASH'.    
            
            - `minRingSize` (optional)
            
                Minimum hash ring size.    
            
            - `maxRingSize` (optional)
            
                Maximum hash ring size.    
        
        - `random` (optional)
        
            Child properties:    
        
        - `maglev` (optional)
        
            Child properties:    
    
    - `destination` (optional)
    
        One destination that the traffic will be redirected to.
        When used, "split" is not allowed.    
    
    - `http` (optional, repeated)
    
        Configuration of HTTP traffic. Traffic is matched one by one with the
        order defined in the list. If the request does not match any criteria
        then "split" or "destination" outside of "http" section is executed.

