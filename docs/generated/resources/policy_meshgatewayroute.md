## MeshGatewayRoute

- `selectors` (required, repeated)

    Selectors is used to match this resource to MeshGateway listener.    
    
    - `match` (optional)
    
        Tags to match, can be used for both source and destinations

- `conf` (required)

    Conf specifies the route configuration.    
    
    - `tcp` (optional)    
        
        - `rules` (required, repeated)    
            
            - `backends` (required, repeated)    
                
                - `weight` (required)
                
                    Weight is the proportion of requests this backend will receive
                    when a forwarding rules specifies multiple backends. Traffic
                    weight is computed as "weight/sum(all weights)".
                    
                    A weight of 0 means that the destination will be ignored.    
                
                - `destination` (required)
                
                    Destination is a selector to match the individual endpoints to
                    which the gateway will forward.    
    
    - `http` (optional)    
        
        - `hostnames` (optional, repeated)
        
            Hostnames lists the server names for which this route is valid. The
            hostnames are matched against the TLS Server Name Indication extension
            if this is a TLS session. They are also matched against the HTTP host
            (authority) header in the client's HTTP request.    
        
        - `rules` (required, repeated)
        
            Rules specifies how the gateway should match and process HTTP requests.    
            
            - `matches` (required, repeated)
            
                Matches are checked in order. If any match is successful, the
                rule is selected (OR semantics).    
                
                - `path` (optional)    
                    
                    - `match` (optional, enum)
                    
                        - `EXACT`
                    
                        - `PREFIX`
                    
                        - `REGEX`    
                    
                    - `value` (required)
                    
                        Value is the path to match against. For EXACT and PREFIX match
                        types, it must be a HTTP URI path. For the REGEX match type,
                        it must be a RE2 regular expression.
                        Note that a PREFIX match succeeds only if the prefix is the
                        the entire path or is followed by a /. I.e. a prefix of the
                        path in terms of path elements.    
                
                - `method` (optional, enum)
                
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
                
                - `headers` (optional, repeated)    
                    
                    - `match` (optional, enum)
                    
                        - `EXACT`
                    
                        - `REGEX`
                    
                        - `ABSENT`
                    
                        - `PRESENT`    
                    
                    - `name` (required)
                    
                        Name of the HTTP header containing the value to match.    
                    
                    - `value` (required)
                    
                        Value that the HTTP header value should be matched against.    
                
                - `queryParameters` (optional, repeated)    
                    
                    - `match` (optional, enum)
                    
                        - `EXACT`
                    
                        - `REGEX`    
                    
                    - `name` (required)
                    
                        Name of the query parameter containing the value to match.    
                    
                    - `value` (required)
                    
                        Value that the query parameter value should be matched against.    
            
            - `filters` (optional, repeated)
            
                Filters are request processing steps that are applied to
                matched requests.
                
                If the redirect filter is specified, it must be the only
                filter given.    
                
                - `requestHeader` (optional)    
                    
                    - `set` (optional, repeated)    
                        
                        - `name` (required)    
                        
                        - `value` (required)    
                    
                    - `add` (optional, repeated)    
                        
                        - `name` (required)    
                        
                        - `value` (required)    
                    
                    - `remove` (optional, repeated)    
                
                - `mirror` (optional)    
                    
                    - `backend` (required)
                    
                        Backend denotes the service to which requests will be mirrored. The
                        "weight" field must not be given.    
                        
                        - `weight` (required)
                        
                            Weight is the proportion of requests this backend will receive
                            when a forwarding rules specifies multiple backends. Traffic
                            weight is computed as "weight/sum(all weights)".
                            
                            A weight of 0 means that the destination will be ignored.    
                        
                        - `destination` (required)
                        
                            Destination is a selector to match the individual endpoints to
                            which the gateway will forward.    
                    
                    - `percentage` (required)
                    
                        Percentage specifies the percentage of requests to mirror to
                        the backend (in the range 0.0 - 100.0, inclusive).    
                
                - `redirect` (optional)    
                    
                    - `scheme` (required)
                    
                        The scheme for the redirect URL. Usually "http" or "https".    
                    
                    - `hostname` (required)
                    
                        The hostname to redirect to.    
                    
                    - `port` (optional)
                    
                        The port to redirect to.    
                    
                    - `statusCode` (required)
                    
                        The HTTP response status code. This must be in the range 300 - 308.    
                
                - `rewrite` (optional)    
                    
                    - `replaceFull` (optional)    
                    
                    - `replacePrefixMatch` (optional)
                    
                        Note that rewriting "/prefix" to "/" will do the right thing:
                        - the path "/prefix" is rewritten to "/"
                        - the path "/prefix/rest" is rewritten to "/rest"    
                
                - `responseHeader` (optional)    
                    
                    - `set` (optional, repeated)    
                        
                        - `name` (required)    
                        
                        - `value` (required)    
                    
                    - `add` (optional, repeated)    
                        
                        - `name` (required)    
                        
                        - `value` (required)    
                    
                    - `remove` (optional, repeated)    
                
                - `rewriteHostToBackendHostname` (optional)
                
                    Option to indicate that during forwarding, the host header should be
                    swapped with the hostname of the upstream host chosen by the Envoy's
                    cluster manager.
                    BE AWARE:
                    - it's mutually exclusive with request_header filter which explicitly
                    replaces "host" header    
            
            - `backends` (optional, repeated)
            
                Backends is the set of services to which the gateway will
                forward requests. If a redirect filter is specified, no
                backends are allowed. Otherwise, at least one backend
                must be given.    
                
                - `weight` (required)
                
                    Weight is the proportion of requests this backend will receive
                    when a forwarding rules specifies multiple backends. Traffic
                    weight is computed as "weight/sum(all weights)".
                    
                    A weight of 0 means that the destination will be ignored.    
                
                - `destination` (required)
                
                    Destination is a selector to match the individual endpoints to
                    which the gateway will forward.

