## kumactl install tracing

Install Tracing backend in Kubernetes cluster (Jaeger)

### Synopsis

Install Tracing backend in Kubernetes cluster (Jaeger) in a 'kuma-tracing' namespace

```
kumactl install tracing [flags]
```

### Options

```
  -h, --help               help for tracing
      --namespace string   namespace to install tracing to (default "kuma-tracing")
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

