## Dataplane

- `networking` (optional)

    Networking describes inbound and outbound interfaces of the dataplane.

    Child properties:    
    
    - `address` (required)
    
        Public IP on which the dataplane is accessible in the network.    
    
    - `advertisedaddress` (optional)
    
        In some situation, dataplane resides in a private network and not
        reachable via 'address'. advertisedAddress is configured with public
        routable address for such dataplane so that other dataplanes in the mesh
        can connect to it over advertisedAddress and not via address
        Note: Envoy binds to the address not advertisedAddress    
    
    - `gateway` (optional)
    
        Gateway describes configuration of gateway of the dataplane.
    
        Child properties:    
        
        - `tags` (required)
        
            Tags associated with a gateway (e.g., Kong, Contour, etc) this
            dataplane is deployed next to, e.g. service=gateway, env=prod.
            `service` tag is mandatory.    
        
        - `type` (required)
        
            Type of gateway this dataplane manages. The default is a DELEGATED
            gateway, which is an external proxy. The BUILTIN gateway type causes
            the dataplane proxy itself to be configured as a gateway.
        
            Supported values:
        
            - `DELEGATED`
        
            - `BUILTIN`    
    
    - `inbound` (optional, repeated)
    
        Inbound describes a list of inbound interfaces of the dataplane.    
    
    - `outbound` (optional, repeated)
    
        Outbound describes a list of outbound interfaces of the dataplane.    
    
    - `transparentProxying` (optional)
    
        TransparentProxying describes configuration for transparent proxying.
    
        Child properties:    
        
        - `redirectPortInbound` (optional)
        
            Port on which all inbound traffic is being transparently redirected.    
        
        - `redirectPortOutbound` (optional)
        
            Port on which all outbound traffic is being transparently redirected.    
        
        - `directAccessServices` (optional, repeated)
        
            List of services that will be access directly via IP:PORT    
        
        - `redirectPortInboundV6` (optional)
        
            Port on which all IPv6 inbound traffic is being transparently
            redirected.    
        
        - `reachableServices` (optional, repeated)
        
            List of reachable services (represented by the value of
            kuma.io/service) via transparent proxying. Setting an explicit list can
            dramatically improve the performance of the mesh. If not specified, all
            services in the mesh are reachable.    
    
    - `admin` (optional)
    
        Admin contains configuration related to Envoy Admin API
    
        Child properties:    
        
        - `port` (optional)
        
            Port on which Envoy Admin API server will be listening

- `metrics` (optional)

    Configuration for metrics that should be collected and exposed by the
    dataplane.
    
    Settings defined here will override their respective defaults
    defined at a Mesh level.

    Child properties:    
    
    - `name` (optional)
    
        Name of the backend, can be then used in Mesh.metrics.enabledBackend    
    
    - `type` (optional)
    
        Type of the backend (Kuma ships with 'prometheus')    
    
    - `conf` (optional)
    
        Configuration of the backend

- `probes` (optional)

    Probes describes list of endpoints which will redirect traffic from
    insecure port to localhost path

    Child properties:    
    
    - `port` (required)    
    
    - `endpoints` (required, repeated)

