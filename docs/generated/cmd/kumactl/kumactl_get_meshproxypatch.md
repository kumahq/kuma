## kumactl get meshproxypatch

Show a single MeshProxyPatch resource

### Synopsis

Show a single MeshProxyPatch resource.

```
kumactl get meshproxypatch NAME [flags]
```

### Options

```
  -h, --help          help for meshproxypatch
  -m, --mesh string   mesh to use (default "default")
```

### Options inherited from parent commands

```
      --api-timeout duration   the timeout for api calls. It includes connection time, any redirects, and reading the response body. A timeout of zero means no timeout (default 1m0s)
      --config-file string     path to the configuration file to use
      --log-format string      log format: one of json|logfmt (default "logfmt")
      --log-level string       log level: one of off|info|debug (default "off")
      --no-config              if set no config file and config directory will be created
  -o, --output string          output format: one of table|yaml|json (default "table")
```

### SEE ALSO

* [kumactl get](kumactl_get.md)	 - Show Kuma resources

