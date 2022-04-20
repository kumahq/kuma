## kumactl install logging

Install Logging backend in Kubernetes cluster (Loki)

### Synopsis

Install Logging backend in Kubernetes cluster (Loki) in its own namespace.

```
kumactl install logging [flags]
```

### Options

```
  -h, --help               help for logging
      --namespace string   namespace to install logging to (default "kuma-logging")
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

