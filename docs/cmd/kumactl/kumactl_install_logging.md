## kumactl install logging

Install Logging backend in Kubernetes cluster (Loki)

### Synopsis

Install Logging backend in Kubernetes cluster (Loki) in a 'kuma-logging' namespace

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
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")
      --no-config            if set no config file and config directory will be created
```

### SEE ALSO

* [kumactl install](kumactl_install.md)	 - Install various Kuma components.

