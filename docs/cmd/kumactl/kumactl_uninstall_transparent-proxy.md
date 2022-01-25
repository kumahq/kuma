## kumactl uninstall transparent-proxy

Uninstall Transparent Proxy pre-requisites on the host

### Synopsis

Uninstall Transparent Proxy by restoring the hosts iptables and /etc/resolv.conf

```
kumactl uninstall transparent-proxy [flags]
```

### Options

```
      --dry-run   dry run
  -h, --help      help for transparent-proxy
      --verbose   verbose
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

* [kumactl uninstall](kumactl_uninstall.md)	 - Uninstall various Kuma components.

