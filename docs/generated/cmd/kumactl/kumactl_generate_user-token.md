## kumactl generate user-token

Generate User Token

### Synopsis

Generate User Token that is used to prove User identity.

```
kumactl generate user-token [flags]
```

### Examples

```

Generate token using a control plane
$ kumactl generate user-token --name john.doe@example.com --group users --valid-for 24h

Generate token using offline signing key
$ kumactl generate user-token --name john.doe@example.com --group users --valid-for 24h --signing-key-path /keys/key.pem --kid 1

```

### Options

```
      --group strings             group of the user
  -h, --help                      help for user-token
      --kid string                ID of the key that is used to issue a token. Required when --signing-key-path is used.
      --name string               name of the user
      --signing-key-path string   path to a file that contains private signing key. When specified, control plane won't be used to issue the token.
      --valid-for duration        how long the token will be valid (for example "24h")
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

