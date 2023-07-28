## Dataplane

- `networking` (optional)

    Networking describes inbound and outbound interfaces of the data plane
    proxy.    
    
    - `address` (required)
    
        IP on which the data plane proxy is accessible to the control plane and
        other data plane proxies in the same network. This can also be a
        hostname, in which case the control plane will periodically resolve it.    
    
    - `advertisedAddress` (optional)
    
        In some situations, a data plane proxy resides in a private network (e.g.
        Docker) and is not reachable via `address` to other data plane proxies.
        `advertisedAddress` is configured with a routable address for such data
        plane proxy so that other proxies in the mesh can connect to it over
        `advertisedAddress` and not via address.
        
        Envoy still binds to the `address`, not `advertisedAddress`.    
    
    - `gateway` (optional)
    
        Gateway describes a configuration of the gateway of the data plane proxy.    
        
        - `tags` (required)
        
            Tags associated with a gateway of this data plane to, e.g.
            `kuma.io/service=gateway`, `env=prod`. `kuma.io/service` tag is
            mandatory.    
        
        - `type` (required, enum)
        
            Type of gateway this data plane proxy manages.
            There are two types: `DELEGATED` and `BUILTIN`. Defaults to
            `DELEGATED`.
            
            A `DELEGATED` gateway is an independently deployed proxy (e.g., Kong,
            Contour, etc) that receives inbound traffic that is not proxied by
            Kuma, and it sends outbound traffic into the data plane proxy.
            
            The `BUILTIN` gateway type causes the data plane proxy itself to be
            configured as a gateway.
            
            See https://kuma.io/docs/latest/explore/gateway/ for more information.
        
            - `DELEGATED`
        
            - `BUILTIN`    
    
    - `inbound` (optional, repeated)
    
        Inbound describes a list of inbound interfaces of the data plane proxy.
        
        Inbound describes a service implemented by the data plane proxy.
        All incoming traffic to a data plane proxy is going through inbound
        listeners. For every defined Inbound there is a corresponding Envoy
        Listener.    
        
        - `port` (required)
        
            Port of the inbound interface that will forward requests to the
            service.
            
            When transparent proxying is used, it is a port on which the service is
            listening to. When transparent proxying is not used, Envoy will bind to
            this port.    
        
        - `servicePort` (optional)
        
            Port of the service that requests will be forwarded to.
            Defaults to the same value as `port`.    
        
        - `serviceAddress` (optional)
        
            Address of the service that requests will be forwarded to.
            Defaults to 'inbound.address', since Kuma DP should be deployed next
            to the service.    
        
        - `address` (optional)
        
            Address on which inbound listener will be exposed.
            Defaults to `networking.address`.    
        
        - `tags` (required)
        
            Tags associated with an application this data plane proxy is deployed
            next to, e.g. `kuma.io/service=web`, `version=1.0`. You can then
            reference these tags in policies like MeshTrafficPermission.
            `kuma.io/service` tag is mandatory.    
        
        - `health` (optional)
        
            Health describes the status of an inbound.
            If 'health' is nil we consider data plane proxy as healthy.
            Unhealthy data plane proxies are excluded from Endpoints Discovery
            Service (EDS). On Kubernetes, it is filled automatically by the control
            plane if Pod has readiness probe configured. On Universal, it can be
            set by the external health checking system, but the most common way is
            to use service probes.
            
            See https://kuma.io/docs/latest/documentation/health for more
            information.    
            
            - `ready` (optional)
            
                Ready indicates if the data plane proxy is ready to serve the
                traffic.    
        
        - `serviceProbe` (optional)
        
            ServiceProbe defines parameters for probing the service next to
            sidecar. When service probe is defined, Envoy will periodically health
            check the application next to it and report the status to the control
            plane. On Kubernetes, Kuma deployments rely on Kubernetes probes so
            this is not used.
            
            See https://kuma.io/docs/latest/documentation/health for more
            information.    
            
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
    
        Outbound describes a list of services consumed by the data plane proxy.
        For every defined Outbound, there is a corresponding Envoy Listener.    
        
        - `address` (optional)
        
            IP on which the consumed service will be available to this data plane
            proxy. On Kubernetes, it's usually ClusterIP of a Service or PodIP of a
            Headless Service. Defaults to 127.0.0.1    
        
        - `port` (required)
        
            Port on which the consumed service will be available to this data plane
            proxy. When transparent proxying is not used, Envoy will bind to this
            port.    
        
        - `tags` (optional)
        
            Tags of consumed data plane proxies.
            `kuma.io/service` tag is required.
            These tags can then be referenced in `destinations` section of policies
            like TrafficRoute or in `to` section in policies like MeshAccessLog. It
            is recommended to only use `kuma.io/service`. If you need to consume
            specific data plane proxy of a service (for example: `version=v2`) the
            better practice is to use TrafficRoute.    
    
    - `transparentProxying` (optional)
    
        TransparentProxying describes the configuration for transparent proxying.
        It is used by default on Kubernetes.    
        
        - `redirectPortInbound` (optional)
        
            Port on which all inbound traffic is being transparently redirected.    
        
        - `redirectPortOutbound` (optional)
        
            Port on which all outbound traffic is being transparently redirected.    
        
        - `directAccessServices` (optional, repeated)
        
            List of services that will be accessed directly via IP:PORT
            Use `*` to indicate direct access to every service in the Mesh.
            Using `*` to directly access every service is a resource-intensive
            operation, use it only if needed.    
        
        - `redirectPortInboundV6` (optional)
        
            Port on which all IPv6 inbound traffic is being transparently
            redirected.    
        
        - `reachableServices` (optional, repeated)
        
            List of reachable services (represented by the value of
            `kuma.io/service`) via transparent proxying. Setting an explicit list
            can dramatically improve the performance of the mesh. If not specified,
            all services in the mesh are reachable.    
    
    - `admin` (optional)
    
        Admin describes configuration related to Envoy Admin API.
        Due to security, all the Envoy Admin endpoints are exposed only on
        localhost. Additionally, Envoy will expose `/ready` endpoint on
        `networking.address` for health checking systems to be able to check the
        state of Envoy. The rest of the endpoints exposed on `networking.address`
        are always protected by mTLS and only meant to be consumed internally by
        the control plane.    
        
        - `port` (optional)
        
            Port on which Envoy Admin API server will be listening

- `metrics` (optional)

    Configuration for metrics that should be collected and exposed by the
    data plane proxy.
    
    Settings defined here will override their respective defaults
    defined at a Mesh level.    
    
    - `name` (optional)
    
        Name of the backend, can be then used in Mesh.metrics.enabledBackend    
    
    - `type` (optional)
    
        Type of the backend (Kuma ships with 'prometheus')    
    
    - `conf` (optional)
    
        Configuration of the backend

- `probes` (optional)

    Probes describe a list of endpoints that will be exposed without mTLS.
    This is useful to expose the health endpoints of the application so the
    orchestration system (e.g. Kubernetes) can still health check the
    application.
    
    See
    https://kuma.io/docs/latest/policies/service-health-probes/#virtual-probes
    for more information.    
    
    - `port` (required)
    
        Port on which the probe endpoints will be exposed. This cannot overlap
        with any other ports.    
    
    - `endpoints` (required, repeated)
    
        List of endpoints to expose without mTLS.    
        
        - `inboundPort` (required)
        
            Inbound port is a port of the application from which we expose the
            endpoint.    
        
        - `inboundPath` (required)
        
            Inbound path is a path of the application from which we expose the
            endpoint. It is recommended to be as specific as possible.    
        
        - `path` (required)
        
            Path is a path on which we expose inbound path on the probes port.

