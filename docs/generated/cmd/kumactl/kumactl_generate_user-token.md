## kumactl generate user-token

Generate User Token

### Synopsis

Generate User Token that is used to prove User identity.

```
kumactl generate user-token [flags]
```

### Examples

```

Generate token
$ kumactl generate user-token --name john.doe@example.com --group users --valid-for 24h

```

### Options

```
      --group strings        group of the user
  -h, --help                 help for user-token
      --name string          name of the user
      --valid-for duration   how long the token will be valid (for example "24h")
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

