## kumactl generate dataplane-token

Generate Dataplane Token

### Synopsis

Generate Dataplane Token that is used to prove Dataplane identity.

```
kumactl generate dataplane-token [flags]
```

### Examples

```

Generate token bound by name and mesh
$ kumactl generate dataplane-token --mesh demo --name demo-01 --valid-for 24h

Generate token bound by mesh
$ kumactl generate dataplane-token --mesh demo --valid-for 24h

Generate Ingress token
$ kumactl generate dataplane-token --type ingress --valid-for 24h

Generate token bound by tag
$ kumactl generate dataplane-token --mesh demo --tag kuma.io/service=web,web-api --valid-for 24h

```

### Options

```
  -h, --help                 help for dataplane-token
      --name string          name of the Dataplane
      --proxy-type string    type of the Dataplane ("dataplane", "ingress")
      --tag stringToString   required tag values for dataplane (split values by comma to provide multiple values) (default [])
      --valid-for duration   how long the token will be valid (for example "24h") (default 87600h0m0s)
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

* [kumactl generate](kumactl_generate.md)	 - Generate resources, tokens, etc

