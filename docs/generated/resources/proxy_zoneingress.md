## ZoneIngress

- `zone` (optional)

    Zone field contains Zone name where ingress is serving, field will be
    automatically set by Global Kuma CP

- `networking` (required)

    Networking defines the address and port of the Ingress to listen on.
    Additionally publicly advertised address and port could be specified.    
    
    - `address` (required)
    
        Address on which inbound listener will be exposed    
    
    - `advertisedAddress` (required)
    
        AdvertisedAddress defines IP or DNS name on which ZoneIngress is
        accessible to other Kuma clusters.    
    
    - `port` (required)
    
        Port of the inbound interface that will forward requests to the service.    
    
    - `advertisedPort` (required)
    
        AdvertisedPort defines port on which ZoneIngress is accessible to other
        Kuma clusters.    
    
    - `admin` (optional)
    
        Admin contains configuration related to Envoy Admin API    
        
        - `port` (optional)
        
            Port on which Envoy Admin API server will be listening

- `availableServices` (optional, repeated)

    AvailableService contains tags that represent unique subset of
    endpoints    
    
    - `tags` (optional)
    
        tags of the service    
    
    - `instances` (optional)
    
        number of instances available for given tags    
    
    - `mesh` (optional)
    
        mesh of the instances available for given tags    
    
    - `externalService` (optional)
    
        instance of external service available from the zone

