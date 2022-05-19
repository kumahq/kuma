## kumactl install observability

Install Observability (Metrics, Logging, Tracing) backend in Kubernetes cluster (Prometheus + Grafana + Loki + Jaeger + Zipkin)

### Synopsis

Install Observability (Metrics, Logging, Tracing) backend in Kubernetes cluster (Prometheus + Grafana + Loki + Jaeger + Zipkin) in its own namespace.

```
kumactl install observability [flags]
```

### Options

```
      --components strings          list of components (default [grafana,prometheus,loki,jaeger])
  -h, --help                        help for observability
      --jaeger-address string       the address of jaeger to query (default "http://jaeger-query.mesh-observability")
      --kuma-cp-address string      the address of Kuma CP (default "http://kuma-control-plane.kuma-system:5676")
      --loki-address string         the address of the loki to query (default "http://loki.mesh-observability:3100")
  -m, --mesh string                 mesh to use (default "default")
      --namespace string            namespace to install observability to (default "mesh-observability")
      --prometheus-address string   the address of the prometheus server (default "http://prometheus-server.mesh-observability")
```

### Options inherited from parent commands

```
      --api-timeout duration   the timeout for api calls. It includes connection time, any redirects, and reading the response body. A timeout of zero means no timeout (default 1m0s)
      --config-file string     path to the configuration file to use
      --log-level string       log level: one of off|info|debug (default "off")
      --no-config              if set no config file and config directory will be created
```

### SEE ALSO

* [kumactl install](kumactl_install.md)	 - Install various Kuma components.

