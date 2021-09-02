## kumactl install metrics

Install Metrics backend in Kubernetes cluster (Prometheus + Grafana)

### Synopsis

Install Metrics backend in Kubernetes cluster (Prometheus + Grafana) in a kuma-metrics namespace

```
kumactl install metrics [flags]
```

### Options

```
  -h, --help                     help for metrics
      --jaeger-address string    the address of jaeger to query (default "http://jaeger-query.kuma-tracing")
      --kuma-cp-address string   the address of Kuma CP (default "http://kuma-control-plane.kuma-system:5676")
      --loki-address string      the address of the loki to query (default "http://loki.kuma-logging:3100")
      --namespace string         namespace to install metrics to (default "kuma-metrics")
      --without-grafana          disable Grafana resources generation
      --without-prometheus       disable Prometheus resources generation
```

### Options inherited from parent commands

```
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")
      --no-config            if set no config file and config directory will be created
```

### SEE ALSO

* [kumactl install](kumactl_install.md)	 - Install various Kuma components.

