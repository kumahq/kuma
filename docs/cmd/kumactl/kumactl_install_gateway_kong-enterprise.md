## kumactl install gateway kong-enterprise

Install Kong ingress gateway on Kubernetes

### Synopsis

Install Kong ingress gateway on Kubernetes in its own namespace.

```
kumactl install gateway kong-enterprise [flags]
```

### Options

```
  -h, --help                  help for kong-enterprise
      --license-path string   path to license file
      --namespace string      namespace to install gateway to (default "kong-enterprise-gateway")
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

* [kumactl install gateway](kumactl_install_gateway.md)	 - Install ingress gateway on Kubernetes

