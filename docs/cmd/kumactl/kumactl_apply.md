## kumactl apply

Create or modify Kuma resources

### Synopsis

Create or modify Kuma resources.

```
kumactl apply [flags]
```

### Examples

```

Apply a resource from file
$ kumactl apply -f resource.yaml

Apply a resource from stdin
$ echo "
type: Mesh
name: demo
" | kumactl apply -f -

Apply a resource from external URL
$ kumactl apply -f https://example.com/resource.yaml

```

### Options

```
      --dry-run              Resolve variable and prints result out without actual applying
  -f, --file -               Path to file to apply. Pass - to read from stdin
  -h, --help                 help for apply
  -v, --var stringToString   Variable to replace in configuration (default [])
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

* [kumactl](kumactl.md)	 - Management tool for Kuma

