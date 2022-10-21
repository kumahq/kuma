## kumactl uninstall transparent-proxy

Uninstall Transparent Proxy pre-requisites on the host

### Synopsis

Uninstall Transparent Proxy by restoring the hosts iptables and /etc/resolv.conf or removing leftover ebpf objects

```
kumactl uninstall transparent-proxy [flags]
```

### Options

```
      --dry-run                  dry run
      --ebpf-bpffs-path string   the path of the BPF filesystem (default "/sys/fs/bpf")
      --ebpf-enabled             uninstall transparent proxy with ebpf mode
  -h, --help                     help for transparent-proxy
      --verbose                  verbose
```

### Options inherited from parent commands

```
      --api-timeout duration   the timeout for api calls. It includes connection time, any redirects, and reading the response body. A timeout of zero means no timeout (default 1m0s)
      --config-file string     path to the configuration file to use
      --log-level string       log level: one of off|info|debug (default "off")
      --no-config              if set no config file and config directory will be created
```

### SEE ALSO

* [kumactl uninstall](kumactl_uninstall.md)	 - Uninstall various Kuma components.

