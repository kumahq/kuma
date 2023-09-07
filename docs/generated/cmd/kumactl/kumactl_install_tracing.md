## kumactl install tracing

Install Tracing backend in Kubernetes cluster (Jaeger)

### Synopsis

Install Tracing backend in Kubernetes cluster (Jaeger) in its own namespace.

```
kumactl install tracing [flags]
```

### Options

```
<<<<<<< HEAD:docs/generated/cmd/kumactl/kumactl_install_tracing.md
  -h, --help               help for tracing
      --namespace string   namespace to install tracing to (default "kuma-tracing")
=======
  -h, --help          help for meshtcproute
  -m, --mesh string   mesh to use (default "default")
>>>>>>> 5141208c6 (fix(kumactl): add `--mesh` parameter to `inspect <policy>` (#7696)):docs/generated/cmd/kumactl/kumactl_inspect_meshtcproute.md
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

