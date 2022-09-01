## kumactl get meshtrafficpermissions

Show MeshTrafficPermission

### Synopsis

Show MeshTrafficPermission entities.

```
kumactl get meshtrafficpermissions [flags]
```

### Options

```
  -h, --help            help for meshtrafficpermissions
  -m, --mesh string     mesh to use (default "default")
      --offset string   the offset that indicates starting element of the resources list to retrieve
      --size int        maximum number of elements to return
```

### Options inherited from parent commands

```
      --api-timeout duration   the timeout for api calls. It includes connection time, any redirects, and reading the response body. A timeout of zero means no timeout (default 1m0s)
      --config-file string     path to the configuration file to use
      --log-level string       log level: one of off|info|debug (default "off")
      --no-config              if set no config file and config directory will be created
  -o, --output string          output format: one of table|yaml|json (default "table")
```

### SEE ALSO

* [kumactl get](kumactl_get.md)	 - Show Kuma resources

