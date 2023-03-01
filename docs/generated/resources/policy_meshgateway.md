## MeshGateway

- `selectors` (optional, repeated)

    Selectors is a list of selectors that are used to match builtin
    gateway dataplanes that will receive this MeshGateway configuration.    
    
    - `match` (optional)
    
        Tags to match, can be used for both source and destinations

- `tags` (optional)

    Tags is the set of tags common to all of the gateway's listeners.
    
    This field must not include a `kuma.io/service` tag (the service is always
    defined on the dataplanes).

- `conf` (optional)

    The desired configuration of the MeshGateway.    
    
    - `listeners` (optional, repeated)
    
        Listeners define logical endpoints that are bound on this MeshGateway's
        address(es).    
        
        - `hostname` (optional)
        
            Hostname specifies the virtual hostname to match for protocol types that
            define this concept. When unspecified, "", or `*`, all hostnames are
            matched. This field can be omitted for protocols that don't require
            hostname based matching.    
        
        - `port` (optional)
        
            Port is the network port. Multiple listeners may use the
            same port, subject to the Listener compatibility rules.    
        
        - `protocol` (optional, enum)
        
            Protocol specifies the network protocol this listener expects to receive.
        
            - `NONE`
        
            - `TCP`
        
            - `HTTP`
        
            - `HTTPS`    
        
        - `tls` (optional)
        
            TLS is the TLS configuration for the Listener. This field
            is required if the Protocol field is "HTTPS" or "TLS" and
            ignored otherwise.    
            
            - `mode` (optional, enum)
            
                Mode defines the TLS behavior for the TLS session initiated
                by the client.
            
                - `NONE`
            
                - `TERMINATE`    
            
            - `certificates` (optional, repeated)
            
                Certificates is an array of datasources that contain TLS
                certificates and private keys.  Each datasource must contain a
                sequence of PEM-encoded objects. The server certificate and private
                key are required, but additional certificates are allowed and will
                be added to the certificate chain.  The server certificate must
                be the first certificate in the datasource.
                
                When multiple certificate datasources are configured, they must have
                different key types. In practice, this means that one datasource
                should contain an RSA key and certificate, and the other an
                ECDSA key and certificate.    
            
            - `options` (optional)
            
                Options should eventually configure how TLS is configured. This
                is where cipher suite and version configuration can be specified,
                client certificates enforced, and so on.    
        
        - `tags` (optional)
        
            Tags specifies a unique combination of tags that routes can use
            to match themselves to this listener.
            
            When matching routes to listeners, the control plane constructs a
            set of matching tags for each listener by forming the union of the
            gateway tags and the listener tags. A route will be attached to the
            listener if all of the route's tags are preset in the matching tags    
        
        - `crossMesh` (optional)
        
            CrossMesh enables traffic to flow to this listener only from other
            meshes.    
        
        - `resources` (optional)
        
            Resources is used to specify listener-specific resource settings.    
            
            - `connectionLimit` (optional)

