## kumactl install demo

Install Kuma demo on Kubernetes

### Synopsis

Install Kuma demo on Kubernetes in a 'kuma-demo' namespace.

```
kumactl install demo [flags]
```

### Options

```
  -h, --help               help for demo
      --namespace string   Namespace to install demo to (default "kuma-demo")
      --zone string        Zone in which to install demo (default "local")
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

