## HealthCheck

- `sources` (required, repeated)

    List of selectors to match dataplanes that should be configured to do
    health checks.

- `destinations` (required, repeated)

    List of selectors to match services that need to be health checked.

- `conf` (required)

    Configuration for various types of health checking.

    Child properties:    
    
    - `interval` (required)
    
        Interval between consecutive health checks.    
    
    - `timeout` (required)
    
        Maximum time to wait for a health check response.    
    
    - `unhealthyThreshold` (required)
    
        Number of consecutive unhealthy checks before considering a host
        unhealthy.    
    
    - `healthyThreshold` (required)
    
        Number of consecutive healthy checks before considering a host healthy.    
    
    - `initialJitter` (optional)
    
        If specified, Envoy will start health checking after for a random time in
        ms between 0 and initial_jitter. This only applies to the first health
        check.    
    
    - `intervalJitter` (optional)
    
        If specified, during every interval Envoy will add interval_jitter to the
        wait time.    
    
    - `intervalJitterPercent` (optional)
    
        If specified, during every interval Envoy will add interval_ms *
        interval_jitter_percent / 100 to the wait time. If interval_jitter_ms and
        interval_jitter_percent are both set, both of them will be used to
        increase the wait time.    
    
    - `healthyPanicThreshold` (optional)
    
        Allows to configure panic threshold for Envoy cluster. If not specified,
        the default is 50%. To disable panic mode, set to 0%.    
    
    - `failTrafficOnPanic` (optional)
    
        If set to true, Envoy will not consider any hosts when the cluster is in
        'panic mode'. Instead, the cluster will fail all requests as if all hosts
        are unhealthy. This can help avoid potentially overwhelming a failing
        service.    
    
    - `eventLogPath` (optional)
    
        Specifies the path to the file where Envoy can log health check events.
        If empty, no event log will be written.    
    
    - `alwaysLogHealthCheckFailures` (optional)
    
        If set to true, health check failure events will always be logged. If set
        to false, only the initial health check failure event will be logged. The
        default value is false.    
    
    - `noTrafficInterval` (optional)
    
        The "no traffic interval" is a special health check interval that is used
        when a cluster has never had traffic routed to it. This lower interval
        allows cluster information to be kept up to date, without sending a
        potentially large amount of active health checking traffic for no reason.
        Once a cluster has been used for traffic routing, Envoy will shift back
        to using the standard health check interval that is defined. Note that
        this interval takes precedence over any other. The default value for "no
        traffic interval" is 60 seconds.    
    
    - `tcp` (optional)
    
        Child properties:    
        
        - `send` (optional)
        
            Bytes which will be send during the health check to the target    
        
        - `receive` (optional, repeated)
        
            Bytes blocks expected as a response. When checking the response,
            “fuzzy” matching is performed such that each block must be found, and
            in the order specified, but not necessarily contiguous.    
    
    - `http` (optional)
    
        Child properties:    
        
        - `path` (required)
        
            The HTTP path which will be requested during the health check
            (ie. /health)
            +required    
        
        - `requestHeadersToAdd` (optional, repeated)
        
            The list of HTTP headers which should be added to each health check
            request
            +optional    
        
        - `expectedStatuses` (optional, repeated)
        
            List of HTTP response statuses which are considered healthy
            +optional    
    
    - `reuseConnection` (optional)
    
        Reuse health check connection between health checks. Default is true.

