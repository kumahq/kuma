## kumactl install dns

Install DNS to Kubernetes

### Synopsis

Install the DNS forwarding to the CoreDNS ConfigMap in the configured Kubernetes Cluster.
This command requires that the KUBECONFIG environment is set

```
kumactl install dns [flags]
```

### Options

```
  -h, --help               help for dns
      --namespace string   namespace to look for Kuma Control Plane service (default "kuma-system")
      --port string        port of the Kuma DNS server (default "5653")
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

