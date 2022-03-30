## kumactl generate zone-ingress-token

Generate Zone Ingress Token

### Synopsis

Generate Zone Ingress Token that is used to prove Zone Ingress identity.

```
kumactl generate zone-ingress-token [flags]
```

### Examples

```

Generate token bound by zone
$ kumactl generate zone-ingress-token --zone zone-1 --valid-for 30d

```

### Options

```
  -h, --help                 help for zone-ingress-token
      --valid-for duration   how long the token will be valid (for example "24h") (default 87600h0m0s)
      --zone string          name of the zone where ingress resides
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

