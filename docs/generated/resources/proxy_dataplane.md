## Dataplane

- `networking` (optional)

    Networking describes inbound and outbound interfaces of the dataplane.    
    
    - `address` (required)
    
        Public IP on which the dataplane is accessible in the network.    
    
    - `advertisedAddress` (optional)
    
        In some situation, dataplane resides in a private network and not
        reachable via 'address'. advertisedAddress is configured with public
        routable address for such dataplane so that other dataplanes in the mesh
        can connect to it over advertisedAddress and not via address
        Note: Envoy binds to the address not advertisedAddress    
    
    - `gateway` (optional)
    
        Gateway describes configuration of gateway of the dataplane.    
        
        - `tags` (required)
        
            Tags associated with a gateway (e.g., Kong, Contour, etc) this
            dataplane is deployed next to, e.g. service=gateway, env=prod.
            `service` tag is mandatory.    
        
        - `type` (required, enum)
        
            Type of gateway this dataplane manages. The default is a DELEGATED
            gateway, which is an external proxy. The BUILTIN gateway type causes
            the dataplane proxy itself to be configured as a gateway.
        
            - `DELEGATED`
        
            - `BUILTIN`    
    
    - `inbound` (optional, repeated)
    
        Inbound describes a list of inbound interfaces of the dataplane.    
        
        - `port` (required)
        
            Port of the inbound interface that will forward requests to the
            service.    
        
        - `servicePort` (optional)
        
            Port of the service that requests will be forwarded to.    
        
        - `serviceAddress` (optional)
        
            Address of the service that requests will be forwarded to.
            Empty value defaults to '127.0.0.1', since Kuma DP should be deployed
            next to service.    
        
        - `address` (optional)
        
            Address on which inbound listener will be exposed. Defaults to
            networking.address.    
        
        - `tags` (required)
        
            Tags associated with an application this dataplane is deployed next to,
            e.g. kuma.io/service=web, version=1.0.
            `kuma.io/service` tag is mandatory.    
        
        - `health` (optional)
        
            Health is an optional field filled automatically by Kuma Control Plane
            on Kubernetes if Pod has ReadinessProbe configured. If 'health' is
            equal to nil we consider dataplane as healthy. Unhealthy dataplanes
            will be excluded from Endpoints Discovery Service (EDS)    
            
            - `ready` (optional)    
        
        - `serviceProbe` (optional)
        
            ServiceProbe defines parameters for probing service's port    
            
            - `interval` (optional)
            
                Interval between consecutive health checks.    
            
            - `timeout` (optional)
            
                Maximum time to wait for a health check response.    
            
            - `unhealthyThreshold` (optional)
            
                Number of consecutive unhealthy checks before considering a host
                unhealthy.    
            
            - `healthyThreshold` (optional)
            
                Number of consecutive healthy checks before considering a host
                healthy.    
            
            - `tcp` (optional)
            
                Tcp checker tries to establish tcp connection with destination    
    
    - `outbound` (optional, repeated)
    
        Outbound describes a list of outbound interfaces of the dataplane.    
        
        - `address` (optional)
        
            Address on which the service will be available to this dataplane.
            Defaults to 127.0.0.1    
        
        - `port` (required)
        
            Port on which the service will be available to this dataplane.    
        
        - `service` (optional)
        
            DEPRECATED: use networking.outbound[].tags
            Service name.    
        
        - `tags` (optional)
        
            Tags    
    
    - `transparentProxying` (optional)
    
        TransparentProxying describes configuration for transparent proxying.    
        
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
        
        - `port` (optional)
        
            Port on which Envoy Admin API server will be listening

- `metrics` (optional)

    Configuration for metrics that should be collected and exposed by the
    dataplane.
    
    Settings defined here will override their respective defaults
    defined at a Mesh level.    
    
    - `name` (optional)
    
        Name of the backend, can be then used in Mesh.metrics.enabledBackend    
    
    - `type` (optional)
    
        Type of the backend (Kuma ships with 'prometheus')    
    
    - `conf` (optional)
    
        Configuration of the backend

- `probes` (optional)

    Probes describes list of endpoints which will redirect traffic from
    insecure port to localhost path    
    
    - `port` (required)    
    
    - `endpoints` (required, repeated)    
        
        - `inboundPort` (required)    
        
        - `inboundPath` (required)    
        
        - `path` (required)

