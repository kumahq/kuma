## ProxyTemplate

- `selectors` (required, repeated)

    List of Dataplane selectors.    
    
    - `match` (optional)
    
        Tags to match, can be used for both source and destinations

- `conf` (required)

    Configuration for ProxyTemplate    
    
    - `imports` (optional, repeated)
    
        List of imported profiles.
        +optional    
    
    - `resources` (optional, repeated)
    
        List of raw xDS resources.
        +optional    
        
        - `name` (required)
        
            The resource's name, to distinguish it from others of the same type of
            resource.    
        
        - `version` (required)
        
            The resource level version. It allows xDS to track the state of individual
            resources.    
        
        - `resource` (required)
        
            xDS resource.    
    
    - `modifications` (optional, repeated)
    
        List of config modifications    
        
        - `cluster` (optional)
        
            Cluster modification    
            
            - `match` (optional)
            
                Only clusters that match will be modified    
                
                - `origin` (optional)
                
                    Origin of the resource generation. (inbound, outbound, prometheus,
                    transparent, ingress)    
                
                - `name` (required)
                
                    Name of the cluster to match    
            
            - `operation` (required)
            
                Operation to apply on a cluster (add, remove, patch)    
            
            - `value` (optional)
            
                xDS cluster    
        
        - `listener` (optional)
        
            Listener modification    
            
            - `match` (optional)
            
                Only listeners that match will be modified    
                
                - `origin` (optional)
                
                    Origin of the resource generation. (inbound, outbound, prometheus,
                    transparent, ingress)    
                
                - `name` (required)
                
                    Name of the listener to match    
                
                - `tags` (optional)
                
                    Tags available in Listener#Metadata#FilterMetadata[io.kuma.tags]    
            
            - `operation` (required)
            
                Operation to apply on a listener (add, remove, patch)    
            
            - `value` (optional)
            
                xDS listener    
        
        - `networkFilter` (optional)
        
            Network Filter modification    
            
            - `match` (optional)
            
                Only network filters that match will be modified    
                
                - `origin` (optional)
                
                    Origin of the resource generation. (inbound, outbound, prometheus,
                    transparent, ingress)    
                
                - `name` (required)
                
                    Name of the network filter    
                
                - `listenerName` (optional)
                
                    Name of the listener that network filter modifications will be
                    applied to    
                
                - `listenerTags` (optional)
                
                    ListenerTags available in
                    Listener#Metadata#FilterMetadata[io.kuma.tags]    
            
            - `operation` (required)
            
                Operation to apply on network filter (addFirst, addLast, addBefore,
                addAfter, remove, patch)    
            
            - `value` (optional)
            
                xDS network filter    
        
        - `httpFilter` (optional)
        
            HTTP Filter modification    
            
            - `match` (optional)
            
                Only HTTP filters that match will be modified    
                
                - `origin` (optional)
                
                    Origin of the resource generation. (inbound, outbound, prometheus,
                    transparent, ingress)    
                
                - `name` (optional)
                
                    Name of the network filter    
                
                - `listenerName` (optional)
                
                    Name of the listener that http filter modifications will be applied
                    to    
                
                - `listenerTags` (optional)
                
                    ListenerTags available in
                    Listener#Metadata#FilterMetadata[io.kuma.tags]    
            
            - `operation` (required)
            
                Operation to apply on network filter (addFirst, addLast, addBefore,
                addAfter, remove, patch)    
            
            - `value` (optional)
            
                xDS HTTP filter    
        
        - `virtualHost` (optional)
        
            Virtual Host modifications    
            
            - `match` (optional)
            
                Only virtual hosts that match will be modified    
                
                - `origin` (optional)
                
                    Origin of the resource generation. (inbound, outbound, prometheus,
                    transparent, ingress)    
                
                - `name` (required)
                
                    Name of the virtual host to match    
                
                - `routeConfigurationName` (optional)
                
                    Name of the route configuration    
            
            - `operation` (required)
            
                Operation to apply on a virtual hosts (add, remove, patch)    
            
            - `value` (optional)
            
                xDS virtual host
## ProxyTemplateSource

- `name` (optional)

    Name of a configuration source.
    +optional

- `profile` (optional)

    Profile, e.g. `default-proxy`.
    +optional    
    
    - `name` (optional)
    
        Profile name.    
    
    - `params` (optional)
    
        Profile params if any.
        +optional

- `raw` (optional)

    Raw xDS resources.
    +optional    
    
    - `resources` (optional, repeated)
    
        List of raw xDS resources.
        +optional    
        
        - `name` (required)
        
            The resource's name, to distinguish it from others of the same type of
            resource.    
        
        - `version` (required)
        
            The resource level version. It allows xDS to track the state of individual
            resources.    
        
        - `resource` (required)
        
            xDS resource.
## ProxyTemplateProfileSource

- `name` (optional)

    Profile name.

- `params` (optional)

    Profile params if any.
    +optional
## ProxyTemplateRawSource

- `resources` (optional, repeated)

    List of raw xDS resources.
    +optional    
    
    - `name` (required)
    
        The resource's name, to distinguish it from others of the same type of
        resource.    
    
    - `version` (required)
    
        The resource level version. It allows xDS to track the state of individual
        resources.    
    
    - `resource` (required)
    
        xDS resource.
## ProxyTemplateRawResource

- `name` (required)

    The resource's name, to distinguish it from others of the same type of
    resource.

- `version` (required)

    The resource level version. It allows xDS to track the state of individual
    resources.

- `resource` (required)

    xDS resource.

