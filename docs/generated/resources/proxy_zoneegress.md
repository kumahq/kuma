## ZoneEgress

- `zone` (optional)

    Zone field contains Zone name where egress is serving, field will be
    automatically set by Global Kuma CP

- `networking` (required)

    Networking defines the address and port of the Egress to listen on.

    Child properties:    
    
    - `address` (required)
    
        Address on which inbound listener will be exposed    
    
    - `port` (required)
    
        Port of the inbound interface that will forward requests to the service.    
    
    - `admin` (optional)
    
        Admin contains configuration related to Envoy Admin API
    
        Child properties:    
        
        - `port` (optional)
        
            Port on which Envoy Admin API server will be listening

