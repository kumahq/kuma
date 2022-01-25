## kumactl inspect dataplanes

Inspect Dataplanes

### Synopsis

Inspect Dataplanes.

```
kumactl inspect dataplanes [flags]
```

### Options

```
      --gateway              filter gateway dataplanes
  -h, --help                 help for dataplanes
      --ingress              filter ingress dataplanes
      --tag stringToString   filter by tag in format of key=value. You can provide many tags (default [])
```

### Options inherited from parent commands

```
      --api-timeout duration   the timeout for api calls. It includes connection time, any redirects, and reading the response body. A timeout of zero means no timeout (default 1m0s)
      --config-file string     path to the configuration file to use
      --log-level string       log level: one of off|info|debug (default "off")
  -m, --mesh string            mesh to use (default "default")
      --no-config              if set no config file and config directory will be created
  -o, --output string          output format: one of table|yaml|json (default "table")
```

### SEE ALSO

* [kumactl inspect](kumactl_inspect.md)	 - Inspect Kuma resources

