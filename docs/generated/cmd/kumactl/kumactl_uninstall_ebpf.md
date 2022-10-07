## kumactl uninstall ebpf

Uninstall BPF files from the nodes

### Synopsis

Uninstall BPF files from the nodes by removing BPF programs from all the nodes

```
kumactl uninstall ebpf [flags]
```

### Options

```
      --bpffs-path string                 path where bpf programs were installed (default "/sys/fs/bpf")
      --cleanup-image-registry string     image registry for ebpf cleanup job (default "kumahq")
      --cleanup-image-repository string   image repository for ebpf cleanup job (default "kuma-init")
      --cleanup-image-tag string          image tag for ebpf cleanup job (default "unknown")
      --cleanup-job-name string           name of the cleanup job (default "kuma-bpf-cleanup")
  -h, --help                              help for ebpf
      --namespace string                  namespace where job is created (default "kuma-system")
      --remove-only                       cleanup jobs and pods only
      --timeout duration                  timeout for whole process of removing left files (default 2m0s)
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

