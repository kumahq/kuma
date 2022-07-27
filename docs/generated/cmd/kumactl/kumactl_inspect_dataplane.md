## kumactl inspect dataplane

Inspect Dataplane

### Synopsis

Inspect Dataplane.

```
kumactl inspect dataplane NAME [flags]
```

### Options

```
  -h, --help          help for dataplane
  -m, --mesh string   mesh to use (default "default")
      --type string   inspection type: one of policies|config-dump|stats|clusters (default "policies")
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

* [kumactl inspect](kumactl_inspect.md)	 - Inspect Kuma resources

