## kumactl install gateway kong

Install Kong ingress gateway on Kubernetes

### Synopsis

Install Kong ingress gateway on Kubernetes in its own namespace.

```
kumactl install gateway kong [flags]
```

### Options

```
  -h, --help               help for kong
      --mesh string        mesh to install gateway to (default "kuma-gateway")
      --namespace string   namespace to install gateway to (default "kuma-gateway")
```

### Options inherited from parent commands

```
      --api-timeout duration   the timeout for api calls. It includes connection time, any redirects, and reading the response body. A timeout of zero means no timeout (default 1m0s)
      --config-file string     path to the configuration file to use
      --log-level string       log level: one of off|info|debug (default "off")
      --no-config              if set no config file and config directory will be created
```

### SEE ALSO

* [kumactl install gateway](kumactl_install_gateway.md)	 - Install ingress gateway on Kubernetes

