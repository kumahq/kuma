## TrafficRoute

- `sources` (required, repeated)

    List of selectors to match data plane proxies that are sources of traffic.    
    
    - `match` (optional)
    
        Tags to match, can be used for both source and destinations

- `destinations` (required, repeated)

    List of selectors to match services that are destinations of traffic.
    
    Notice the difference between sources and destinations.
    While the source of traffic is always a data plane proxy within a mesh,
    the destination is a service that could be either within or outside
    of a mesh.    
    
    - `match` (optional)
    
        Tags to match, can be used for both source and destinations

- `conf` (required)

    Configuration for the route.    
    
    - `split` (optional, repeated)
    
        List of destinations with weights assigned to them.
        When used, "destination" is not allowed.    
        
        - `weight` (required)
        
            Weight assigned to that destination.
            Weights are not percentages. For example two destinations with
            weights the same weight "1" will receive both same amount of the traffic.
            0 means that the destination will be ignored.    
        
        - `destination` (required)
        
            Selector to match individual endpoints that comprise that destination.
            
            Notice that an endpoint can be either inside or outside the mesh.
            In the former case an endpoint corresponds to a data plane proxy,
            in the latter case an endpoint is an External Service.    
    
    - `loadBalancer` (optional)
    
        Load balancer configuration for given "split" or "destination"    
        
        - `roundRobin` (optional)    
        
        - `leastRequest` (optional)    
            
            - `choiceCount` (optional)
            
                The number of random healthy hosts from which the host with the fewest
                active requests will be chosen. Defaults to 2 so that we perform
                two-choice selection if the field is not set.    
        
        - `ringHash` (optional)    
            
            - `hashFunction` (optional)
            
                The hash function used to hash hosts onto the ketama ring. The value
                defaults to 'XX_HASH'.    
            
            - `minRingSize` (optional)
            
                Minimum hash ring size.    
            
            - `maxRingSize` (optional)
            
                Maximum hash ring size.    
        
        - `random` (optional)    
        
        - `maglev` (optional)    
    
    - `destination` (optional)
    
        One destination that the traffic will be redirected to.
        When used, "split" is not allowed.    
    
    - `http` (optional, repeated)
    
        Configuration of HTTP traffic. Traffic is matched one by one with the
        order defined in the list. If the request does not match any criteria
        then "split" or "destination" outside of "http" section is executed.    
        
        - `match` (optional)
        
            If request matches against defined criteria then "split" or "destination"
            is executed.    
            
            - `method` (optional)
            
                Method matches method of HTTP request.    
                
                - `prefix` (optional)
                
                    Prefix matches the string against defined prefix.    
                
                - `exact` (optional)
                
                    Exact checks that strings are equal to each other.    
                
                - `regex` (optional)
                
                    Regex checks the string using RE2 syntax.
                    https://github.com/google/re2/wiki/Syntax    
            
            - `path` (optional)
            
                Path matches HTTP path.    
                
                - `prefix` (optional)
                
                    Prefix matches the string against defined prefix.    
                
                - `exact` (optional)
                
                    Exact checks that strings are equal to each other.    
                
                - `regex` (optional)
                
                    Regex checks the string using RE2 syntax.
                    https://github.com/google/re2/wiki/Syntax    
            
            - `headers` (optional)
            
                Headers match HTTP request headers.    
        
        - `modify` (optional)
        
            Modifications to the traffic matched by the match section.    
            
            - `path` (optional)
            
                Path modifications.    
                
                - `rewritePrefix` (optional)
                
                    RewritePrefix rewrites previously matched prefix in match section.    
                
                - `regex` (optional)
                
                    Regex rewrites prefix using regex with substitution.    
                    
                    - `pattern` (required)
                    
                        Pattern of the regex using RE2 syntax.
                        https://github.com/google/re2/wiki/Syntax    
                    
                    - `substitution` (required)
                    
                        Substitution using regex groups. E.g. use \\1 as a first matched
                        group.    
            
            - `host` (optional)
            
                Host modifications.    
                
                - `value` (optional)
                
                    Value replaces the host header with given value.    
                
                - `fromPath` (optional)
                
                    FromPath replaces the host header from path using regex.    
                    
                    - `pattern` (required)
                    
                        Pattern of the regex using RE2 syntax.
                        https://github.com/google/re2/wiki/Syntax    
                    
                    - `substitution` (required)
                    
                        Substitution using regex groups. E.g. use \\1 as a first matched
                        group.    
            
            - `requestHeaders` (optional)
            
                Request headers modifications.    
                
                - `add` (optional, repeated)
                
                    List of add header operations.    
                    
                    - `name` (required)
                    
                        Name of the header.    
                    
                    - `value` (required)
                    
                        Value of the header.    
                    
                    - `append` (optional)
                    
                        If true, it appends the value if there is already a value.
                        Otherwise, value of existing header will be replaced.    
                
                - `remove` (optional, repeated)
                
                    List of remove header operations.    
                    
                    - `name` (required)
                    
                        Name of the header to remove.    
            
            - `responseHeaders` (optional)
            
                Response headers modifications.    
                
                - `add` (optional, repeated)
                
                    List of add header operations.    
                    
                    - `name` (required)
                    
                        Name of the header.    
                    
                    - `value` (required)
                    
                        Value of the header.    
                    
                    - `append` (optional)
                    
                        If true, it appends the value if there is already a value.
                        Otherwise, value of existing header will be replaced.    
                
                - `remove` (optional, repeated)
                
                    List of remove header operations.    
                    
                    - `name` (required)
                    
                        Name of the header to remove.    
        
        - `split` (optional, repeated)
        
            List of destinations with weights assigned to them.
            When used, "destination" is not allowed.    
            
            - `weight` (required)
            
                Weight assigned to that destination.
                Weights are not percentages. For example two destinations with
                weights the same weight "1" will receive both same amount of the traffic.
                0 means that the destination will be ignored.    
            
            - `destination` (required)
            
                Selector to match individual endpoints that comprise that destination.
                
                Notice that an endpoint can be either inside or outside the mesh.
                In the former case an endpoint corresponds to a data plane proxy,
                in the latter case an endpoint is an External Service.    
        
        - `destination` (optional)
        
            One destination that the traffic will be redirected to.
            When used, "split" is not allowed.

