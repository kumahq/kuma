## kumactl install metrics

Install Metrics backend in Kubernetes cluster (Prometheus + Grafana)

### Synopsis

Install Metrics backend in Kubernetes cluster (Prometheus + Grafana) in a kuma-metrics namespace

```
kumactl install metrics [flags]
```

### Options

```
  -h, --help                                help for metrics
      --kuma-cp-address string              the address of Kuma CP (default "grpc://kuma-control-plane.kuma-system:5676")
      --kuma-prometheus-sd-image string     image name of Kuma Prometheus SD (default "docker.io/kumahq/kuma-prometheus-sd")
      --kuma-prometheus-sd-version string   version of Kuma Prometheus SD (default "unknown")
      --namespace string                    namespace to install metrics to (default "kuma-metrics")
      --without-grafana                     disable Grafana resources generation
      --without-prometheus                  disable Prometheus resources generation
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

