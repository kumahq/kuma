## kumactl install crds

Install Kuma Custom Resource Definitions on Kubernetes

```
kumactl install crds [flags]
```

### Options

```
      --experimental-meshgateway   install experimental built-in MeshGateway support
  -h, --help                       help for crds
      --only-missing               install only resources which are not already present in a cluster
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

