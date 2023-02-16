## kumactl generate public-key

Generate public key out of signing key

### Synopsis

Generate a public key for validating tokens.

```
kumactl generate public-key [flags]
```

### Examples

```

Extract a public key out of signing key used to issue tokens.

$ kumactl generate signing-key --format=pem > /tmp/key.pem
$ kumactl generate public-key --signing-key-path=/tmp/key.pem

```

### Options

```
  -h, --help                      help for public-key
      --signing-key-path string   path to a file with PEM-encoded private signing key
```

### Options inherited from parent commands

```
      --api-timeout duration   the timeout for api calls. It includes connection time, any redirects, and reading the response body. A timeout of zero means no timeout (default 1m0s)
      --config-file string     path to the configuration file to use
      --log-level string       log level: one of off|info|debug (default "off")
      --no-config              if set no config file and config directory will be created
```

### SEE ALSO

* [kumactl generate](kumactl_generate.md)	 - Generate resources, tokens, etc

