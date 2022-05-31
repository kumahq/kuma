## kumactl install tracing

Install Tracing backend in Kubernetes cluster (Jaeger)

### Synopsis

Install Tracing backend in Kubernetes cluster (Jaeger) in its own namespace.

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
      --api-timeout duration   the timeout for api calls. It includes connection time, any redirects, and reading the response body. A timeout of zero means no timeout (default 1m0s)
      --config-file string     path to the configuration file to use
      --log-level string       log level: one of off|info|debug (default "off")
      --no-config              if set no config file and config directory will be created
```

### SEE ALSO

* [kumactl install](kumactl_install.md)	 - Install various Kuma components.

