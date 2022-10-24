## ExternalService

- `networking` (required)    
    
    - `address` (required)
    
        Address of the external service    
    
    - `tls` (optional)    
        
        - `enabled` (optional)
        
            denotes that the external service uses TLS    
        
        - `caCert` (optional)
        
            Data source for the certificate of CA    
        
        - `clientCert` (optional)
        
            Data source for the authentication    
        
        - `clientKey` (optional)
        
            Data source for the authentication    
        
        - `allowRenegotiation` (optional)
        
            If true then TLS session will allow renegotiation.
            It's not recommended to set this to true because of security reasons.
            However, some servers requires this setting, especially when using
            mTLS.    
        
        - `serverName` (optional)
        
            ServerName overrides the default Server Name Indicator set by Kuma.
            The default value is set to "address" specified in "networking".    
    
    - `disableHostDNSEntry` (optional)
    
        If disableHostDNSEntry is set to true then a DNS entry for the external
        service taken from 'networking.address' won't be generated.
        You can still reach this external service using
        external-service-name.mesh:80 where "external-service-name" is taken from
        "kuma.io/service" tag.

- `tags` (required)

    Tags associated with the external service,
    e.g. kuma.io/service=web, kuma.io/protocol, version=1.0.

