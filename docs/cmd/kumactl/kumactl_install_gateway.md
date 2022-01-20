## kumactl install gateway

Install ingress gateway on Kubernetes

### Synopsis

Install ingress gateway on Kubernetes in its own namespace.

### Options

```
  -h, --help   help for gateway
```

### Options inherited from parent commands

```
      --api-timeout duration   the timeout for api calls. It includes connection time, any redirects, and reading the response body. A timeout of zero means no timeout (default 1m0s)
      --config-file string     path to the configuration file to use
      --log-level string       log level: one of off|info|debug (default "off")
  -m, --mesh string            mesh to use (default "default")
      --no-config              if set no config file and config directory will be created
```

### SEE ALSO

* [kumactl install](kumactl_install.md)	 - Install various Kuma components.
* [kumactl install gateway kong](kumactl_install_gateway_kong.md)	 - Install Kong ingress gateway on Kubernetes
* [kumactl install gateway kong-enterprise](kumactl_install_gateway_kong-enterprise.md)	 - Install Kong ingress gateway on Kubernetes

